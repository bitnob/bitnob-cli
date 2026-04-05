package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/bitnob/bitnob-cli/internal/app"
	"github.com/bitnob/bitnob-cli/internal/output"
	"github.com/peterh/liner"
	"github.com/spf13/cobra"
)

type shellMode string

const (
	shellModeCommand shellMode = "command"
)

var (
	newLinerState = func() *liner.State {
		return liner.NewLiner()
	}
)

func newShellCommand(printer output.Printer, application *app.App) *cobra.Command {
	var mode string

	cmd := &cobra.Command{
		Use:   "shell",
		Short: "Start an interactive shell with history",
		Long: `Start an interactive shell for running Bitnob CLI commands continuously.

Modes:
  command   Execute commands directly (for example: balances, whoami)`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			selectedMode, err := parseShellMode(mode)
			if err != nil {
				return err
			}
			return runShell(cmd.Context(), printer, application, selectedMode)
		},
	}

	cmd.Flags().StringVar(&mode, "mode", string(shellModeCommand), "Shell mode: command")
	return cmd
}

func parseShellMode(raw string) (shellMode, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case string(shellModeCommand):
		return shellModeCommand, nil
	case "assistant":
		return "", fmt.Errorf("assistant mode is not available yet")
	default:
		return "", fmt.Errorf("unsupported shell mode %q (expected: command)", raw)
	}
}

func runShell(ctx context.Context, printer output.Printer, application *app.App, initialMode shellMode) error {
	line := newLinerState()
	defer line.Close()

	line.SetCtrlCAborts(true)

	historyPath := shellHistoryPath(application.StateDir)
	_ = loadShellHistory(line, historyPath)
	defer func() {
		_ = saveShellHistory(line, historyPath)
	}()

	mode := initialMode

	_ = printer.Println(`shell started. type "exit" to quit.`)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		input, err := line.Prompt(fmt.Sprintf("bitnob[%s]> ", mode))
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			if errors.Is(err, liner.ErrPromptAborted) {
				continue
			}
			return err
		}

		commandLine := strings.TrimSpace(input)
		if commandLine == "" {
			continue
		}

		line.AppendHistory(commandLine)

		if shouldExitShell(commandLine) {
			return nil
		}

		if strings.HasPrefix(strings.ToLower(strings.TrimSpace(commandLine)), "mode ") {
			nextMode, err := parseModeSwitch(commandLine)
			if err != nil {
				_ = printer.Errorln(err.Error())
				continue
			}
			mode = nextMode
			_ = printer.Println(fmt.Sprintf("switched to %s mode", mode))
			continue
		}

		args, err := parseShellInput(commandLine)
		if err != nil {
			_ = printer.Errorln(err.Error())
			continue
		}

		if len(args) == 0 {
			continue
		}
		if args[0] == "shell" {
			_ = printer.Errorln("shell cannot be started from within shell")
			continue
		}

		if err := executeShellArgs(ctx, printer, application, args); err != nil {
			_ = printer.Errorln(err.Error())
		}
	}
}

func executeShellArgs(ctx context.Context, printer output.Printer, application *app.App, args []string) error {
	root := NewRootCommand(ctx, printer, application)
	root.SetArgs(args)
	return root.ExecuteContext(ctx)
}

func shellHistoryPath(stateDir string) string {
	base := strings.TrimSpace(stateDir)
	if base == "" {
		base = "."
	}
	return filepath.Join(base, "shell", "history")
}

func loadShellHistory(line *liner.State, path string) error {
	file, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	defer file.Close()
	_, err = line.ReadHistory(file)
	return err
}

func saveShellHistory(line *liner.State, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = line.WriteHistory(file)
	return err
}

func shouldExitShell(input string) bool {
	switch strings.ToLower(strings.TrimSpace(input)) {
	case "exit", "quit":
		return true
	default:
		return false
	}
}

func parseModeSwitch(input string) (shellMode, error) {
	fields := strings.Fields(strings.TrimSpace(strings.ToLower(input)))
	if len(fields) != 2 || fields[0] != "mode" {
		return "", fmt.Errorf("usage: mode command")
	}
	mode, err := parseShellMode(fields[1])
	if err != nil {
		return "", err
	}
	return mode, nil
}

func parseShellInput(input string) ([]string, error) {
	var args []string
	var current strings.Builder
	inSingle := false
	inDouble := false
	escaped := false

	flush := func() {
		if current.Len() > 0 {
			args = append(args, current.String())
			current.Reset()
		}
	}

	for _, r := range input {
		if escaped {
			current.WriteRune(r)
			escaped = false
			continue
		}

		switch r {
		case '\\':
			escaped = true
		case '\'':
			if !inDouble {
				inSingle = !inSingle
				continue
			}
			current.WriteRune(r)
		case '"':
			if !inSingle {
				inDouble = !inDouble
				continue
			}
			current.WriteRune(r)
		case ' ', '\t':
			if inSingle || inDouble {
				current.WriteRune(r)
			} else {
				flush()
			}
		default:
			current.WriteRune(r)
		}
	}

	if escaped {
		return nil, fmt.Errorf("unterminated escape sequence")
	}
	if inSingle || inDouble {
		return nil, fmt.Errorf("unterminated quoted string")
	}

	flush()
	return args, nil
}
