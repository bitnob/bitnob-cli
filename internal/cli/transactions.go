package cli

import (
	"github.com/bitnob/bitnob-cli/internal/app"
	"github.com/bitnob/bitnob-cli/internal/output"
	"github.com/bitnob/bitnob-cli/internal/transactions"
	"github.com/spf13/cobra"
)

func newTransactionsCommand(printer output.Printer, application *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transactions",
		Short: "Retrieve transactions",
	}

	cmd.AddCommand(
		newTransactionsListCommand(printer, application),
		newTransactionsGetCommand(printer, application),
	)

	return cmd
}

func newTransactionsListCommand(printer output.Printer, application *app.App) *cobra.Command {
	var filters transactions.ListFilters

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List transactions with optional filters",
		RunE: func(cmd *cobra.Command, _ []string) error {
			response, err := application.TransactionsService.List(cmd.Context(), filters)
			if err != nil {
				return err
			}
			return printer.PrintJSON(response)
		},
	}

	cmd.Flags().IntVar(&filters.Page, "page", 0, "Page number")
	cmd.Flags().IntVar(&filters.Limit, "limit", 0, "Number of transactions to return")
	cmd.Flags().StringVar(&filters.Cursor, "cursor", "", "Cursor for the next page of transactions")
	cmd.Flags().StringVar(&filters.Reference, "reference", "", "Filter by transaction reference")
	cmd.Flags().StringVar(&filters.Hash, "hash", "", "Filter by transaction hash")
	cmd.Flags().StringVar(&filters.Action, "action", "", "Filter by transaction action")
	cmd.Flags().StringVar(&filters.Channel, "channel", "", "Filter by transaction channel")
	cmd.Flags().StringVar(&filters.Type, "type", "", "Filter by transaction type")
	cmd.Flags().StringVar(&filters.Status, "status", "", "Filter by transaction status")
	cmd.Flags().StringVar(&filters.Currency, "currency", "", "Filter by transaction currency")
	cmd.Flags().StringVar(&filters.BankAccountID, "bank-account-id", "", "Filter by bank account ID")
	cmd.Flags().StringVar(&filters.CustomerReference, "customer-reference", "", "Filter by customer reference")
	cmd.Flags().StringVar(&filters.WalletID, "wallet-id", "", "Filter by wallet ID")
	cmd.Flags().StringVar(&filters.Address, "address", "", "Filter by blockchain address")
	return cmd
}

func newTransactionsGetCommand(printer output.Printer, application *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "get <id-or-reference>",
		Short: "Get a transaction by ID or reference",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			response, err := application.TransactionsService.Get(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			return printer.PrintJSON(response)
		},
	}
}
