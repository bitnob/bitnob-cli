package cli

import (
	"github.com/bitnob/bitnob-cli/internal/app"
	"github.com/bitnob/bitnob-cli/internal/output"
	"github.com/spf13/cobra"
)

func newLogoutCommand(printer output.Printer, application *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Clear credentials for the active profile",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := application.CredentialsService.Logout(cmd.Context()); err != nil {
				return err
			}

			return printer.PrintJSON(map[string]string{
				"status":  "ok",
				"message": "Credentials cleared for active profile",
			})
		},
	}
}
