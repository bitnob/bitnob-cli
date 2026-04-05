package cli

import (
	"fmt"

	"github.com/bitnob/bitnob-cli/internal/app"
	"github.com/bitnob/bitnob-cli/internal/output"
	"github.com/bitnob/bitnob-cli/internal/transfers"
	"github.com/spf13/cobra"
)

func newTransfersCommand(printer output.Printer, application *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transfers",
		Short: "Create wallet transfers",
	}

	cmd.AddCommand(newTransfersCreateCommand(printer, application))
	return cmd
}

func newTransfersCreateCommand(printer output.Printer, application *app.App) *cobra.Command {
	var input transfers.CreateInput

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a wallet transfer",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if input.ToAddress == "" || input.Amount == "" || input.Currency == "" || input.Chain == "" {
				return fmt.Errorf("to-address, amount, currency, and chain are required")
			}

			response, err := application.TransfersService.Create(cmd.Context(), input)
			if err != nil {
				return err
			}
			return printer.PrintJSON(response)
		},
	}

	cmd.Flags().StringVar(&input.ToAddress, "to-address", "", "Recipient blockchain address")
	cmd.Flags().StringVar(&input.Amount, "amount", "", "Amount in smallest units")
	cmd.Flags().StringVar(&input.Currency, "currency", "", "Transfer currency, for example USDT, USDC, or BTC")
	cmd.Flags().StringVar(&input.Chain, "chain", "", "Transfer chain, for example tron, bsc, polygon, solana, ethereum, or bitcoin")
	cmd.Flags().StringVar(&input.Reference, "reference", "", "Reference for idempotency and reconciliation")
	cmd.Flags().StringVar(&input.Description, "description", "", "Description for the transfer")
	return cmd
}
