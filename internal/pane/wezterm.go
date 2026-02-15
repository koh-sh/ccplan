package pane

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

const (
	pollInterval = 500 * time.Millisecond
	maxWaitTime  = 10 * time.Minute
)

// WezTermSpawner spawns commands in a new WezTerm pane.
type WezTermSpawner struct{}

func (w *WezTermSpawner) Available() bool {
	_, err := exec.LookPath("wezterm")
	return err == nil
}

func (w *WezTermSpawner) Name() string {
	return "wezterm"
}

func (w *WezTermSpawner) SpawnAndWait(cmd string, args []string) error {
	fullCmd := append([]string{cmd}, args...)
	splitArgs := append([]string{
		"cli", "split-pane", "--bottom", "--percent", "80", "--",
	}, fullCmd...)

	out, err := exec.Command("wezterm", splitArgs...).Output()
	if err != nil {
		return fmt.Errorf("wezterm split-pane: %w", err)
	}
	paneID := strings.TrimSpace(string(out))

	timeout := time.After(maxWaitTime)
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return fmt.Errorf("timeout waiting for pane %s to close", paneID)
		case <-ticker.C:
			if !w.paneExists(paneID) {
				return nil
			}
		}
	}
}

func (w *WezTermSpawner) paneExists(paneID string) bool {
	out, err := exec.Command("wezterm", "cli", "list", "--format", "json").Output()
	if err != nil {
		return false
	}
	var panes []struct {
		PaneID json.Number `json:"pane_id"`
	}
	if err := json.Unmarshal(out, &panes); err != nil {
		return false
	}
	for _, p := range panes {
		if p.PaneID.String() == paneID {
			return true
		}
	}
	return false
}
