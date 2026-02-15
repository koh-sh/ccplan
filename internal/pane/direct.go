package pane

import (
	"fmt"
	"os"
	"os/exec"
)

// DirectSpawner runs commands directly in the current terminal as a fallback.
type DirectSpawner struct{}

func (d *DirectSpawner) Available() bool {
	return true
}

func (d *DirectSpawner) Name() string {
	return "direct"
}

func (d *DirectSpawner) SpawnAndWait(cmd string, args []string) error {
	c := exec.Command(cmd, args...)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		return fmt.Errorf("direct spawn: %w", err)
	}
	return nil
}
