package pane

// AutoDetect returns the best available PaneSpawner for the current environment.
func AutoDetect() PaneSpawner {
	spawners := []PaneSpawner{
		&WezTermSpawner{},
		// &TmuxSpawner{}, // Phase 4
	}
	for _, s := range spawners {
		if s.Available() {
			return s
		}
	}
	return &DirectSpawner{}
}

// ByName returns a PaneSpawner by name, falling back to AutoDetect if not found.
func ByName(name string) PaneSpawner {
	switch name {
	case "wezterm":
		return &WezTermSpawner{}
	case "tmux":
		return &DirectSpawner{} // Phase 4: return &TmuxSpawner{}
	case "auto", "":
		return AutoDetect()
	default:
		return AutoDetect()
	}
}
