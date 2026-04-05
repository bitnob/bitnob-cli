package cli

import (
	"github.com/bitnob/bitnob-cli/internal/app"
	"github.com/bitnob/bitnob-cli/internal/output"
	"github.com/spf13/cobra"
)

func newVersionCommand(printer output.Printer, application *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show CLI version",
		RunE: func(_ *cobra.Command, _ []string) error {
			return printer.PrintJSON(application.Version)
		},
	}
}
