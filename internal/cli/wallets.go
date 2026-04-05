package cli

import (
	"github.com/bitnob/bitnob-cli/internal/app"
	"github.com/bitnob/bitnob-cli/internal/output"
	"github.com/spf13/cobra"
)

func newWalletsCommand(printer output.Printer, application *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "wallets",
		Short: "List company wallets",
		RunE: func(cmd *cobra.Command, _ []string) error {
			response, err := application.WalletsService.List(cmd.Context())
			if err != nil {
				return err
			}

			return printer.PrintJSON(response)
		},
	}
}
