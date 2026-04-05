package cli

import (
	"fmt"

	"github.com/bitnob/bitnob-cli/internal/app"
	"github.com/bitnob/bitnob-cli/internal/lightning"
	"github.com/bitnob/bitnob-cli/internal/output"
	"github.com/spf13/cobra"
)

func newLightningCommand(printer output.Printer, application *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lightning",
		Short: "Manage Lightning invoices and payments",
	}

	cmd.AddCommand(
		newLightningInvoicesCommand(printer, application),
		newLightningDecodeCommand(printer, application),
		newLightningPaymentsCommand(printer, application),
	)

	return cmd
}

func newLightningInvoicesCommand(printer output.Printer, application *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "invoices",
		Short: "Manage Lightning invoices",
	}

	cmd.AddCommand(
		newLightningInvoicesCreateCommand(printer, application),
		newLightningInvoicesGetCommand(printer, application),
		newLightningInvoicesListCommand(printer, application),
		newLightningInvoicesVerifyCommand(printer, application),
	)

	return cmd
}

func newLightningInvoicesCreateCommand(printer output.Printer, application *app.App) *cobra.Command {
	var input lightning.CreateInvoiceInput

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a Lightning invoice",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if input.Amount <= 0 {
				return fmt.Errorf("amount is required and must be greater than zero")
			}
			if input.Description == "" || input.Reference == "" {
				return fmt.Errorf("description and reference are required")
			}

			response, err := application.LightningService.CreateInvoice(cmd.Context(), input)
			if err != nil {
				return err
			}
			return printer.PrintJSON(response)
		},
	}

	cmd.Flags().StringVar(&input.CustomerID, "customer-id", "", "Customer identifier for the invoice")
	cmd.Flags().Int64Var(&input.Amount, "amount", 0, "Invoice amount in satoshis")
	cmd.Flags().StringVar(&input.Description, "description", "", "Invoice description")
	cmd.Flags().Int64Var(&input.Expiry, "expiry", 0, "Invoice expiry in seconds")
	cmd.Flags().StringVar(&input.Reference, "reference", "", "Idempotency reference for the invoice")
	return cmd
}

func newLightningInvoicesGetCommand(printer output.Printer, application *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get a Lightning invoice by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			response, err := application.LightningService.GetInvoice(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			return printer.PrintJSON(response)
		},
	}
}

func newLightningInvoicesListCommand(printer output.Printer, application *app.App) *cobra.Command {
	var page int
	var pageSize int
	var status string
	var direction string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Lightning invoices",
		RunE: func(cmd *cobra.Command, _ []string) error {
			response, err := application.LightningService.ListInvoices(cmd.Context(), page, pageSize, status, direction)
			if err != nil {
				return err
			}
			return printer.PrintJSON(response)
		},
	}

	cmd.Flags().IntVar(&page, "page", 0, "Page number")
	cmd.Flags().IntVar(&pageSize, "page-size", 0, "Number of invoices per page")
	cmd.Flags().StringVar(&status, "status", "", "Invoice status filter")
	cmd.Flags().StringVar(&direction, "direction", "", "Invoice direction filter")
	return cmd
}

func newLightningInvoicesVerifyCommand(printer output.Printer, application *app.App) *cobra.Command {
	var input lightning.VerifyInvoiceInput

	cmd := &cobra.Command{
		Use:   "verify",
		Short: "Verify a Lightning invoice payment by payment hash",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if input.PaymentHash == "" {
				return fmt.Errorf("payment-hash is required")
			}

			response, err := application.LightningService.VerifyInvoice(cmd.Context(), input)
			if err != nil {
				return err
			}
			return printer.PrintJSON(response)
		},
	}

	cmd.Flags().StringVar(&input.PaymentHash, "payment-hash", "", "Lightning payment hash to verify")
	return cmd
}

func newLightningDecodeCommand(printer output.Printer, application *app.App) *cobra.Command {
	var input lightning.DecodeInput

	cmd := &cobra.Command{
		Use:   "decode",
		Short: "Decode a Lightning payment request",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if input.Request == "" {
				return fmt.Errorf("request is required")
			}

			response, err := application.LightningService.Decode(cmd.Context(), input)
			if err != nil {
				return err
			}
			return printer.PrintJSON(response)
		},
	}

	cmd.Flags().StringVar(&input.Request, "request", "", "BOLT11 payment request")
	return cmd
}

func newLightningPaymentsCommand(printer output.Printer, application *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "payments",
		Short: "Initiate and send Lightning payments",
	}

	cmd.AddCommand(
		newLightningPaymentsInitiateCommand(printer, application),
		newLightningPaymentsSendCommand(printer, application),
	)

	return cmd
}

func newLightningPaymentsInitiateCommand(printer output.Printer, application *app.App) *cobra.Command {
	var input lightning.InitiatePaymentInput

	cmd := &cobra.Command{
		Use:   "initiate",
		Short: "Initiate a Lightning payment preview",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if input.Request == "" {
				return fmt.Errorf("request is required")
			}

			response, err := application.LightningService.InitiatePayment(cmd.Context(), input)
			if err != nil {
				return err
			}
			return printer.PrintJSON(response)
		},
	}

	cmd.Flags().StringVar(&input.Request, "request", "", "BOLT11 payment request")
	return cmd
}

func newLightningPaymentsSendCommand(printer output.Printer, application *app.App) *cobra.Command {
	var input lightning.SendPaymentInput

	cmd := &cobra.Command{
		Use:   "send",
		Short: "Send a Lightning payment",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if input.Request == "" || input.Reference == "" {
				return fmt.Errorf("request and reference are required")
			}

			response, err := application.LightningService.SendPayment(cmd.Context(), input)
			if err != nil {
				return err
			}
			return printer.PrintJSON(response)
		},
	}

	cmd.Flags().StringVar(&input.CustomerID, "customer-id", "", "Customer identifier for the payment")
	cmd.Flags().StringVar(&input.Request, "request", "", "BOLT11 payment request")
	cmd.Flags().StringVar(&input.Reference, "reference", "", "Idempotency reference for the payment")
	cmd.Flags().Int64Var(&input.MaxFee, "max-fee", 0, "Maximum routing fee in satoshis")
	return cmd
}
