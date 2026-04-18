package cli

import (
	"github.com/bitnob/bitnob-cli/internal/app"
	"github.com/bitnob/bitnob-cli/internal/output"
	"github.com/spf13/cobra"
)

func newWithdrawalsCommand(printer output.Printer, application *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "withdrawals",
		Short: "Initiate withdrawals",
	}

	cmd.AddCommand(newWithdrawalsCreateCommand(printer, application))
	return cmd
}

func newWithdrawalsCreateCommand(printer output.Printer, application *app.App) *cobra.Command {
	var data string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Initiate a withdrawal",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runRawRequest(cmd.Context(), printer, application, "POST", "/api/withdrawals", data)
		},
		Args: cobra.NoArgs,
	}
	cmd.Flags().StringVar(&data, "data", "", "JSON request body")
	return cmd
}
