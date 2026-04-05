package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bitnob/bitnob-cli/internal/output"
	"github.com/spf13/cobra"
)

var detectShell = func() string {
	shell := os.Getenv("SHELL")
	if shell == "" {
		return ""
	}
	return filepath.Base(shell)
}

func newCompletionCommand(printer output.Printer) *cobra.Command {
	var shell string

	cmd := &cobra.Command{
		Use:   "completion",
		Short: "Set up shell autocompletion",
		Long: `Set up shell autocompletion for Bitnob CLI.

If you pass --shell, the command prints the raw completion script.
If you omit --shell, the CLI tries to detect your shell and prints setup instructions.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			resolvedShell := shell
			if resolvedShell == "" {
				resolvedShell = detectShell()
				if resolvedShell == "" {
					return fmt.Errorf("could not detect your shell; use --shell bash or --shell zsh")
				}
				return printer.Println(completionInstructions(resolvedShell))
			}

			switch resolvedShell {
			case "bash":
				return cmd.Root().GenBashCompletion(printer.Stdout)
			case "zsh":
				return cmd.Root().GenZshCompletion(printer.Stdout)
			default:
				return fmt.Errorf("unsupported shell %q; supported shells are bash and zsh", resolvedShell)
			}
		},
	}

	cmd.Flags().StringVar(&shell, "shell", "", "The shell for which autocompletion will be generated")
	return cmd
}

func completionInstructions(shell string) string {
	switch shell {
	case "bash":
		return strings.TrimSpace(`Detected shell: bash

To enable completion for the current session:
  source <(bitnob completion --shell bash)

To enable completion permanently:
  mkdir -p ~/.local/share/bash-completion/completions
  bitnob completion --shell bash > ~/.local/share/bash-completion/completions/bitnob`)
	case "zsh":
		return strings.TrimSpace(`Detected shell: zsh

To enable completion for the current session:
  source <(bitnob completion --shell zsh)

To enable completion permanently:
  mkdir -p ~/.zsh/completions
  bitnob completion --shell zsh > ~/.zsh/completions/_bitnob

Make sure ~/.zsh/completions is included in your fpath before compinit runs.`)
	default:
		return fmt.Sprintf("Unsupported detected shell %q; use --shell bash or --shell zsh", shell)
	}
}
