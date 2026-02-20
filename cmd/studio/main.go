package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/charmbracelet/huh"
)

type action struct {
	label string
	make  string
	desc  string
}

var actions = []action{
	{"Dev", "dev", "Start Vite dev server + Wails window (hot reload)"},
	{"Build", "build", "Compile the .app bundle"},
	{"Package", "package", "Compile the .app bundle and wrap in a .dmg"},
	{"Clean", "clean", "Remove build output"},
}

func main() {
	var choice string

	opts := make([]huh.Option[string], len(actions))
	for i, a := range actions {
		opts[i] = huh.NewOption(fmt.Sprintf("%-10s %s", a.label, a.desc), a.make)
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

	fmt.Printf("\n→ make %s\n\n", choice)

	cmd := exec.Command("make", choice)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		os.Exit(1)
	}
}
