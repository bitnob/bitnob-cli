package cli

import (
	"strings"

	"github.com/bitnob/bitnob-cli/internal/app"
	"github.com/bitnob/bitnob-cli/internal/output"
	"github.com/spf13/cobra"
)

func newBalancesCommand(printer output.Printer, application *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "balances",
		Short: "Fetch company balances",
		RunE: func(cmd *cobra.Command, _ []string) error {
			response, err := application.BalancesService.Get(cmd.Context())
			if err != nil {
				return err
			}

			return printer.PrintJSON(response)
		},
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "Fetch company balances",
		RunE: func(cmd *cobra.Command, _ []string) error {
			response, err := application.BalancesService.Get(cmd.Context())
			if err != nil {
				return err
			}

			return printer.PrintJSON(response)
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "get <currency>",
		Short: "Fetch balance for a specific currency",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			response, err := application.BalancesService.GetByCurrency(cmd.Context(), strings.ToUpper(args[0]))
			if err != nil {
				return err
			}

			return printer.PrintJSON(response)
		},
	})

	return cmd
}
