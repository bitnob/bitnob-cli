package cli

import (
	"github.com/bitnob/bitnob-cli/internal/app"
	"github.com/bitnob/bitnob-cli/internal/output"
	"github.com/spf13/cobra"
)

func newWhoAmICommand(printer output.Printer, application *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "whoami",
		Short: "Show the active CLI identity",
		RunE: func(cmd *cobra.Command, _ []string) error {
			identity, err := application.IdentityService.WhoAmI(cmd.Context())
			if err != nil {
				return err
			}

			return printer.PrintJSON(identity)
		},
	}
	return cmd
}
