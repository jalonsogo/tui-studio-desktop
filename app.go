package main

import (
	"context"
	"fmt"
	"os"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App holds the application context and exposes IPC methods to the frontend.
// All exported methods on *App are automatically bound as window.go.main.App.*()
// in the webview (Promise-returning).
type App struct {
	ctx context.Context
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// NativeSaveDialog opens a native save-file dialog and writes content to the
// chosen path. Returns the chosen path on success, or "" if the user cancelled.
func (a *App) NativeSaveDialog(defaultName string, content string) (string, error) {
	path, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		DefaultFilename: defaultName,
		Title:           "Save TUI File",
		Filters: []runtime.FileFilter{
			{DisplayName: "TUI Studio Files (*.tui)", Pattern: "*.tui"},
			{DisplayName: "All Files (*.*)", Pattern: "*.*"},
		},
	})
	if err != nil {
		return "", fmt.Errorf("save dialog: %w", err)
	}
	if path == "" {
		return "", nil // user cancelled
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("write file: %w", err)
	}
	return path, nil
}

// NativeOpenDialog opens a native open-file dialog and returns the file's text
// content. Returns ("", nil) if the user cancelled.
func (a *App) NativeOpenDialog() (string, error) {
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Open TUI File",
		Filters: []runtime.FileFilter{
			{DisplayName: "TUI Studio Files (*.tui)", Pattern: "*.tui"},
			{DisplayName: "All Files (*.*)", Pattern: "*.*"},
		},
	})
	if err != nil {
		return "", fmt.Errorf("open dialog: %w", err)
	}
	if path == "" {
		return "", nil // user cancelled
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read file: %w", err)
	}
	return string(data), nil
}

// NativePickDirectory opens a native folder-picker dialog and returns the
// chosen directory path. Returns "" if the user cancelled.
func (a *App) NativePickDirectory() (string, error) {
	path, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Download Folder",
	})
	if err != nil {
		return "", fmt.Errorf("directory dialog: %w", err)
	}
	return path, nil
}

// NativeWriteFile writes content to an absolute path (used by the directory
// handle bridge for export operations).
func (a *App) NativeWriteFile(path string, content string) error {
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("write file %q: %w", path, err)
	}
	return nil
}

// NativeReadFile reads and returns the text content of a file at path.
func (a *App) NativeReadFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read file %q: %w", path, err)
	}
	return string(data), nil
}
