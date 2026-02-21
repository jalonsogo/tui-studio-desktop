package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

type action struct {
	label      string
	key        string
	desc       string
	cmd        []string // if set, run directly instead of make
	useSpinner bool
	render     func([]byte) // if set, called with raw output instead of printing
}

var actions = []action{
	{"Dev", "dev", "Start Vite dev server + Wails window (hot reload)", nil, false, nil},
	{"Build", "build", "Compile the .app bundle", nil, true, nil},
	{"Package", "package", "Compile the .app bundle and wrap in a .dmg", nil, true, nil},
	{"Clean", "clean", "Remove build output", nil, false, nil},
	{"Downloads", "downloads", "Show release download counts",
		[]string{"sh", "-c", "gh api repos/jalonsogo/tui-studio-desktop/releases | jq '[.[] | {tag: .tag_name, published: .published_at[:10], assets: [.assets[] | {name, download_count}]}]'"},
		false, renderDownloads},
}

// — downloads renderer —

type releaseAsset struct {
	Name          string `json:"name"`
	DownloadCount int    `json:"download_count"`
}

type releaseInfo struct {
	Tag       string         `json:"tag"`
	Published string         `json:"published"`
	Assets    []releaseAsset `json:"assets"`
}

func renderDownloads(raw []byte) {
	var releases []releaseInfo
	if err := json.Unmarshal(raw, &releases); err != nil {
		fmt.Print(string(raw))
		return
	}

	tagStyle    := lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true)
	dateStyle   := lipgloss.NewStyle().Foreground(lipgloss.Color("243"))
	nameStyle   := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	countStyle  := lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Bold(true)
	ruleStyle   := lipgloss.NewStyle().Foreground(lipgloss.Color("238"))
	labelStyle  := lipgloss.NewStyle().Foreground(lipgloss.Color("243"))
	headerStyle := lipgloss.NewStyle().Bold(true)

	const width = 54
	fmt.Println()
	fmt.Println("  " + headerStyle.Render("Release Downloads"))
	fmt.Println("  " + ruleStyle.Render(strings.Repeat("─", width)))

	total := 0
	for _, r := range releases {
		for _, a := range r.Assets {
			total += a.DownloadCount
			fmt.Printf("  %-18s  %s  %-20s  %s\n",
				tagStyle.Render(r.Tag),
				dateStyle.Render(r.Published),
				nameStyle.Render(a.Name),
				countStyle.Render(fmt.Sprintf("%d", a.DownloadCount)),
			)
		}
	}

	fmt.Println("  " + ruleStyle.Render(strings.Repeat("─", width)))
	fmt.Printf("  %-44s%s\n",
		labelStyle.Render("Total"),
		countStyle.Render(fmt.Sprintf("%d", total)),
	)
	fmt.Println()
}

// — spinner TUI —

type doneMsg struct {
	output []byte
	err    error
}

type spinnerModel struct {
	spinner spinner.Model
	label   string
	output  string
	cmdErr  error
}

func (m spinnerModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m spinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case doneMsg:
		m.output = string(msg.output)
		m.cmdErr = msg.err
		return m, tea.Quit
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m spinnerModel) View() string {
	return fmt.Sprintf("\n  %s %s\n", m.spinner.View(), m.label)
}

func runWithSpinner(label string, cmdArgs []string) ([]byte, error) {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	p := tea.NewProgram(spinnerModel{spinner: s, label: label})

	go func() {
		cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
		out, err := cmd.CombinedOutput()
		p.Send(doneMsg{output: out, err: err})
	}()

	final, err := p.Run()
	if err != nil {
		return nil, err
	}

	fm := final.(spinnerModel)
	return []byte(fm.output), fm.cmdErr
}

// — main —

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

	cmdArgs := selected.cmd
	if cmdArgs == nil {
		cmdArgs = []string{"make", selected.key}
	}

	if selected.useSpinner || selected.cmd != nil {
		fmt.Printf("\n→ %s\n", selected.desc)
		out, err := runWithSpinner(selected.desc+"...", cmdArgs)
		if err != nil {
			fmt.Print(string(out))
			os.Exit(1)
		}
		if selected.render != nil {
			selected.render(out)
		} else {
			fmt.Println()
			fmt.Print(string(out))
		}
		return
	}

	fmt.Printf("\n→ make %s\n\n", choice)
	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		os.Exit(1)
	}
}
