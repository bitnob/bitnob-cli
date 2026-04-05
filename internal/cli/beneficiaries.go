package cli

import (
	"fmt"

	"github.com/bitnob/bitnob-cli/internal/app"
	"github.com/bitnob/bitnob-cli/internal/beneficiaries"
	"github.com/bitnob/bitnob-cli/internal/output"
	"github.com/spf13/cobra"
)

func newBeneficiariesCommand(printer output.Printer, application *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "beneficiaries",
		Short: "Manage Bitnob beneficiaries",
	}

	cmd.AddCommand(
		newBeneficiariesListCommand(printer, application),
		newBeneficiariesGetCommand(printer, application),
		newBeneficiariesCreateCommand(printer, application),
		newBeneficiariesUpdateCommand(printer, application),
		newBeneficiariesDeleteCommand(printer, application),
	)

	return cmd
}

func newBeneficiariesListCommand(printer output.Printer, application *app.App) *cobra.Command {
	var limit int
	var offset int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List beneficiaries",
		RunE: func(cmd *cobra.Command, _ []string) error {
			response, err := application.BeneficiariesService.List(cmd.Context(), limit, offset)
			if err != nil {
				return err
			}
			return printer.PrintJSON(response)
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 0, "Number of beneficiaries to return")
	cmd.Flags().IntVar(&offset, "offset", 0, "Starting offset for beneficiaries")
	return cmd
}

func newBeneficiariesGetCommand(printer output.Printer, application *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get a beneficiary by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			response, err := application.BeneficiariesService.Get(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			return printer.PrintJSON(response)
		},
	}
}

func newBeneficiariesCreateCommand(printer output.Printer, application *app.App) *cobra.Command {
	var input beneficiaries.CreateInput

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new beneficiary",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if input.CompanyID == "" || input.Type == "" || input.Name == "" {
				return fmt.Errorf("company-id, type, and name are required")
			}
			if input.Destination.AccountName == "" || input.Destination.AccountNumber == "" || input.Destination.BankCode == "" {
				return fmt.Errorf("account-name, account-number, and bank-code are required")
			}

			response, err := application.BeneficiariesService.Create(cmd.Context(), input)
			if err != nil {
				return err
			}
			return printer.PrintJSON(response)
		},
	}

	cmd.Flags().StringVar(&input.CompanyID, "company-id", "", "Company ID that owns the beneficiary")
	cmd.Flags().StringVar(&input.Type, "type", "", "Beneficiary type, for example bank")
	cmd.Flags().StringVar(&input.Name, "name", "", "Beneficiary display name")
	cmd.Flags().StringVar(&input.Destination.AccountName, "account-name", "", "Bank account name")
	cmd.Flags().StringVar(&input.Destination.AccountNumber, "account-number", "", "Bank account number")
	cmd.Flags().StringVar(&input.Destination.BankCode, "bank-code", "", "Bank code")
	cmd.Flags().StringVar(&input.CreatedBy, "created-by", "", "Creator ID for auditing")
	return cmd
}

func newBeneficiariesUpdateCommand(printer output.Printer, application *app.App) *cobra.Command {
	var input beneficiaries.UpdateInput
	var blacklist bool
	var blacklistSet bool

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an existing beneficiary",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if input.Name == "" {
				return fmt.Errorf("name is required")
			}
			if input.Destination.AccountName == "" || input.Destination.AccountNumber == "" || input.Destination.BankCode == "" {
				return fmt.Errorf("account-name, account-number, and bank-code are required")
			}
			if cmd.Flags().Changed("blacklisted") {
				blacklistSet = true
			}
			if blacklistSet {
				input.IsBlacklisted = &blacklist
			}

			response, err := application.BeneficiariesService.Update(cmd.Context(), args[0], input)
			if err != nil {
				return err
			}
			return printer.PrintJSON(response)
		},
	}

	cmd.Flags().StringVar(&input.Name, "name", "", "Beneficiary display name")
	cmd.Flags().StringVar(&input.Destination.AccountName, "account-name", "", "Bank account name")
	cmd.Flags().StringVar(&input.Destination.AccountNumber, "account-number", "", "Bank account number")
	cmd.Flags().StringVar(&input.Destination.BankCode, "bank-code", "", "Bank code")
	cmd.Flags().BoolVar(&blacklist, "blacklisted", false, "Set beneficiary blacklisted status")
	return cmd
}

func newBeneficiariesDeleteCommand(printer output.Printer, application *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a beneficiary",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			response, err := application.BeneficiariesService.Delete(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			return printer.PrintJSON(response)
		},
	}
}
