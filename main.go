package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/caseymrm/menuet"
)

type keychainCredentials struct {
	ClaudeAiOauth struct {
		AccessToken string `json:"accessToken"`
	} `json:"claudeAiOauth"`
}

type usageWindow struct {
	Utilization float64 `json:"utilization"`
	ResetsAt    string  `json:"resets_at"`
}

type usageResponse struct {
	FiveHour usageWindow `json:"five_hour"`
	SevenDay usageWindow `json:"seven_day"`
}

var (
	mu          sync.RWMutex
	currentData *usageResponse
	lastError   string
)

func readAccessToken() (string, error) {
	cmd := exec.Command("security", "find-generic-password", "-s", "Claude Code-credentials", "-w")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("keychain read failed: %w", err)
	}

	raw := strings.TrimSpace(string(out))

	var creds keychainCredentials
	if err := json.Unmarshal([]byte(raw), &creds); err != nil {
		return "", fmt.Errorf("credential parse failed: %w", err)
	}

	if creds.ClaudeAiOauth.AccessToken == "" {
		return "", fmt.Errorf("no access token found in credentials")
	}

	return creds.ClaudeAiOauth.AccessToken, nil
}

func fetchUsage(token string) (*usageResponse, error) {
	req, err := http.NewRequest("GET", "https://api.anthropic.com/api/oauth/usage", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("anthropic-beta", "oauth-2025-04-20")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "claude-code/2.0.31")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned %d: %s", resp.StatusCode, string(body))
	}

	var usage usageResponse
	if err := json.NewDecoder(resp.Body).Decode(&usage); err != nil {
		return nil, fmt.Errorf("response parse failed: %w", err)
	}

	return &usage, nil
}

func refresh() {
	token, err := readAccessToken()
	if err != nil {
		mu.Lock()
		lastError = err.Error()
		currentData = nil
		mu.Unlock()
		menuet.App().SetMenuState(&menuet.MenuState{Title: "CC: ?"})
		return
	}

	usage, err := fetchUsage(token)
	if err != nil {
		mu.Lock()
		lastError = err.Error()
		currentData = nil
		mu.Unlock()
		menuet.App().SetMenuState(&menuet.MenuState{Title: "CC: ?"})
		return
	}

	mu.Lock()
	currentData = usage
	lastError = ""
	mu.Unlock()

	remaining := 100 - usage.FiveHour.Utilization
	title := formatTitle(remaining)
	menuet.App().SetMenuState(&menuet.MenuState{Title: title})
}

func formatTitle(remaining float64) string {
	pct := fmt.Sprintf("%.0f%%", remaining)
	switch {
	case remaining < 20:
		return "CC: 🔴 " + pct
	case remaining < 50:
		return "CC: ⚠ " + pct
	default:
		return "CC: " + pct
	}
}

func formatResetTime(resetStr string) string {
	t, err := time.Parse(time.RFC3339, resetStr)
	if err != nil {
		return "?"
	}
	local := t.Local()
	return local.Format("15:04")
}

func formatResetTimeWithDay(resetStr string) string {
	t, err := time.Parse(time.RFC3339, resetStr)
	if err != nil {
		return "?"
	}
	local := t.Local()
	return local.Format("Mon 15:04")
}

func menuItems() []menuet.MenuItem {
	mu.RLock()
	data := currentData
	errMsg := lastError
	mu.RUnlock()

	var items []menuet.MenuItem

	if errMsg != "" {
		items = append(items, menuet.MenuItem{
			Text: "Error: " + errMsg,
		})
	} else if data != nil {
		fiveRemaining := 100 - data.FiveHour.Utilization
		sevenRemaining := 100 - data.SevenDay.Utilization

		items = append(items,
			menuet.MenuItem{
				Text: fmt.Sprintf("5h: %.0f%% remaining (resets %s)", fiveRemaining, formatResetTime(data.FiveHour.ResetsAt)),
			},
			menuet.MenuItem{
				Text: fmt.Sprintf("7d: %.0f%% remaining (resets %s)", sevenRemaining, formatResetTimeWithDay(data.SevenDay.ResetsAt)),
			},
		)
	} else {
		items = append(items, menuet.MenuItem{
			Text: "Loading...",
		})
	}

	items = append(items,
		menuet.MenuItem{Type: menuet.Separator},
		menuet.MenuItem{
			Text: "Refresh",
			Clicked: func() {
				go refresh()
			},
		},
		menuet.MenuItem{
			Text: "Quit",
			Clicked: func() {
				os.Exit(0)
			},
		},
	)

	return items
}

func pollLoop() {
	refresh()
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		refresh()
	}
}

func main() {
	go pollLoop()

	app := menuet.App()
	app.SetMenuState(&menuet.MenuState{Title: "CC: ..."})
	app.Children = menuItems
	app.Label = "com.github.jakubenglicky.claude-code-bar"
	app.RunApplication()
}
