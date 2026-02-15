package pane

import "fmt"

// TmuxSpawner is a placeholder for future tmux support.
type TmuxSpawner struct{}

func (t *TmuxSpawner) Available() bool {
	return false
}

func (t *TmuxSpawner) Name() string {
	return "tmux"
}

func (t *TmuxSpawner) SpawnAndWait(cmd string, args []string) error {
	return fmt.Errorf("tmux spawner not yet implemented")
}
