package cli

import (
	"github.com/bitnob/bitnob-cli/internal/app"
	"github.com/bitnob/bitnob-cli/internal/output"
	"github.com/spf13/cobra"
)

func newSwitchCommand(printer output.Printer, application *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "switch <profile>",
		Short: "Switch the active profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			selected, err := application.ProfileService.Switch(cmd.Context(), args[0])
			if err != nil {
				return err
			}

			return printer.PrintJSON(map[string]string{
				"status":      "ok",
				"message":     "Active profile switched",
				"profile":     selected.Name,
				"auth_method": selected.AuthMethod,
			})
		},
	}
}
