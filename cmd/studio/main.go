package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/charmbracelet/huh"
)

type action struct {
	label string
	key   string
	desc  string
	cmd   []string // if set, run directly instead of make
}

var actions = []action{
	{"Dev", "dev", "Start Vite dev server + Wails window (hot reload)", nil},
	{"Build", "build", "Compile the .app bundle", nil},
	{"Package", "package", "Compile the .app bundle and wrap in a .dmg", nil},
	{"Clean", "clean", "Remove build output", nil},
	{"Downloads", "downloads", "Show release download counts", []string{"sh", "-c", "gh api repos/jalonsogo/tui-studio-desktop/releases | jq '.[].assets[] | {name, download_count}'"}},
}

func main() {
	var choice string

	opts := make([]huh.Option[string], len(actions))
	for i, a := range actions {
		opts[i] = huh.NewOption(fmt.Sprintf("%-10s %s", a.label, a.desc), a.key)
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("TUIStudio — what do you want to do?").
				Options(opts...).
				Value(&choice),
		),
	)

	if err := form.Run(); err != nil {
		os.Exit(0) // user cancelled
	}

	var selected action
	for _, a := range actions {
		if a.key == choice {
			selected = a
			break
		}
	}

	var cmd *exec.Cmd
	if selected.cmd != nil {
		fmt.Printf("\n→ %s\n\n", selected.desc)
		cmd = exec.Command(selected.cmd[0], selected.cmd[1:]...)
	} else {
		fmt.Printf("\n→ make %s\n\n", choice)
		cmd = exec.Command("make", choice)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		os.Exit(1)
	}
}
