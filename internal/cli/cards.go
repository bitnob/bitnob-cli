package cli

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/bitnob/bitnob-cli/internal/app"
	"github.com/bitnob/bitnob-cli/internal/cards"
	"github.com/bitnob/bitnob-cli/internal/output"
	"github.com/spf13/cobra"
)

func newCardsCommand(printer output.Printer, application *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cards",
		Short: "Manage virtual and physical cards",
	}

	cmd.AddCommand(
		newCardsCreateCommand(printer, application),
		newCardsGetCommand(printer, application),
		newCardsDetailsCommand(printer, application),
		newCardsListCommand(printer, application),
		newCardsFundCommand(printer, application),
		newCardsWithdrawCommand(printer, application),
		newCardsFreezeCommand(printer, application),
		newCardsUnfreezeCommand(printer, application),
		newCardsTerminateCommand(printer, application),
		newCardsSpendingLimitsCommand(printer, application),
		newCardsCustomerCommand(printer, application),
	)

	return cmd
}

func newCardsCreateCommand(printer output.Printer, application *app.App) *cobra.Command {
	var input cards.CreateInput
	var metadata string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a card",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if input.CustomerID == "" || input.CardType == "" || input.CardBrand == "" || input.Currency == "" {
				return fmt.Errorf("customer-id, card-type, card-brand, and currency are required")
			}
			if input.BillingAddress.Line1 == "" || input.BillingAddress.City == "" || input.BillingAddress.State == "" || input.BillingAddress.PostalCode == "" || input.BillingAddress.Country == "" {
				return fmt.Errorf("billing-address line1, city, state, postal-code, and country are required")
			}
			if input.Name == "" {
				customer, err := application.CustomersService.Get(cmd.Context(), input.CustomerID)
				if err != nil {
					return fmt.Errorf("derive cardholder name from customer: %w", err)
				}
				input.Name = strings.TrimSpace(strings.Join([]string{customer.Data.FirstName, customer.Data.LastName}, " "))
				if input.Name == "" {
					return fmt.Errorf("customer %q has no first_name or last_name; pass --name explicitly", input.CustomerID)
				}
			}
			if metadata != "" {
				if err := json.Unmarshal([]byte(metadata), &input.Metadata); err != nil {
					return fmt.Errorf("parse metadata: %w", err)
				}
			}

			response, err := application.CardsService.Create(cmd.Context(), input)
			if err != nil {
				return err
			}
			return printer.PrintJSON(response)
		},
	}

	cmd.Flags().StringVar(&input.CustomerID, "customer-id", "", "Customer ID to issue the card for")
	cmd.Flags().StringVar(&input.Name, "name", "", "Cardholder name; defaults from the customer record when omitted")
	cmd.Flags().StringVar(&input.CardType, "card-type", "", "Card type, for example virtual or physical")
	cmd.Flags().StringVar(&input.CardBrand, "card-brand", "", "Card brand, for example visa or mastercard")
	cmd.Flags().StringVar(&input.Currency, "currency", "", "Card currency")
	cmd.Flags().StringVar(&input.BillingAddress.Line1, "billing-line1", "", "Billing address line 1")
	cmd.Flags().StringVar(&input.BillingAddress.Line2, "billing-line2", "", "Billing address line 2")
	cmd.Flags().StringVar(&input.BillingAddress.City, "billing-city", "", "Billing address city")
	cmd.Flags().StringVar(&input.BillingAddress.State, "billing-state", "", "Billing address state")
	cmd.Flags().StringVar(&input.BillingAddress.PostalCode, "billing-postal-code", "", "Billing address postal code")
	cmd.Flags().StringVar(&input.BillingAddress.Country, "billing-country", "", "Billing address country code")
	cmd.Flags().StringVar(&input.SpendingLimits.SingleTransaction, "single-transaction", "", "Single transaction spending limit")
	cmd.Flags().StringVar(&input.SpendingLimits.Daily, "daily", "", "Daily spending limit")
	cmd.Flags().StringVar(&input.SpendingLimits.Weekly, "weekly", "", "Weekly spending limit")
	cmd.Flags().StringVar(&input.SpendingLimits.Monthly, "monthly", "", "Monthly spending limit")
	cmd.Flags().StringVar(&metadata, "metadata", "", "Metadata JSON object")
	return cmd
}

func newCardsGetCommand(printer output.Printer, application *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "get <card-id>",
		Short: "Get a card by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			response, err := application.CardsService.Get(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			return printer.PrintJSON(response)
		},
	}
}

func newCardsDetailsCommand(printer output.Printer, application *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "details <card-id>",
		Short: "Get sensitive card details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			response, err := application.CardsService.GetDetails(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			return printer.PrintJSON(response)
		},
	}
}

func newCardsListCommand(printer output.Printer, application *app.App) *cobra.Command {
	var filters cards.ListFilters

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List cards",
		RunE: func(cmd *cobra.Command, _ []string) error {
			response, err := application.CardsService.List(cmd.Context(), filters)
			if err != nil {
				return err
			}
			return printer.PrintJSON(response)
		},
	}

	cmd.Flags().StringVar(&filters.Status, "status", "", "Filter by card status")
	cmd.Flags().StringVar(&filters.CardType, "card-type", "", "Filter by card type")
	cmd.Flags().StringVar(&filters.CustomerID, "customer-id", "", "Filter by customer ID")
	cmd.Flags().IntVar(&filters.Limit, "limit", 0, "Maximum number of cards to return")
	cmd.Flags().IntVar(&filters.Offset, "offset", 0, "Number of cards to skip")
	return cmd
}

func newCardsFundCommand(printer output.Printer, application *app.App) *cobra.Command {
	return newCardsBalanceOperationCommand(printer, application, "fund")
}

func newCardsWithdrawCommand(printer output.Printer, application *app.App) *cobra.Command {
	return newCardsBalanceOperationCommand(printer, application, "withdraw")
}

func newCardsBalanceOperationCommand(printer output.Printer, application *app.App, operation string) *cobra.Command {
	var input cards.FundWithdrawInput
	var amount string

	cmd := &cobra.Command{
		Use:   operation + " <card-id>",
		Short: stringsTitle(operation) + " a card",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if amount == "" || input.Currency == "" || input.Reference == "" {
				return fmt.Errorf("amount, currency, and reference are required")
			}
			parsedAmount, err := parseCardAmountToMicroUnits(amount)
			if err != nil {
				return fmt.Errorf("parse amount: %w", err)
			}
			input.Amount = parsedAmount

			var (
				response cards.BalanceOperationResponse
			)
			if operation == "fund" {
				response, err = application.CardsService.Fund(cmd.Context(), args[0], input)
			} else {
				response, err = application.CardsService.Withdraw(cmd.Context(), args[0], input)
			}
			if err != nil {
				return err
			}
			return printer.PrintJSON(response)
		},
	}

	cmd.Flags().StringVar(&amount, "amount", "", "Amount for the operation in decimal units, for example 15.00")
	cmd.Flags().StringVar(&input.Currency, "currency", "", "Currency for the operation")
	cmd.Flags().StringVar(&input.Description, "description", "", "Description for the operation")
	cmd.Flags().StringVar(&input.Reference, "reference", "", "Reference for idempotency and reconciliation")
	return cmd
}

func newCardsFreezeCommand(printer output.Printer, application *app.App) *cobra.Command {
	return newCardsReasonCommand(printer, application, "freeze")
}

func newCardsTerminateCommand(printer output.Printer, application *app.App) *cobra.Command {
	return newCardsReasonCommand(printer, application, "terminate")
}

func newCardsReasonCommand(printer output.Printer, application *app.App, operation string) *cobra.Command {
	var input cards.ReasonInput

	cmd := &cobra.Command{
		Use:   operation + " <card-id>",
		Short: stringsTitle(operation) + " a card",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if operation == "freeze" {
				response, err := application.CardsService.Freeze(cmd.Context(), args[0], input)
				if err != nil {
					return err
				}
				return printer.PrintJSON(response)
			}

			response, err := application.CardsService.Terminate(cmd.Context(), args[0], input)
			if err != nil {
				return err
			}
			return printer.PrintJSON(response)
		},
	}

	cmd.Flags().StringVar(&input.Reason, "reason", "", "Reason for the operation")
	return cmd
}

func newCardsUnfreezeCommand(printer output.Printer, application *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "unfreeze <card-id>",
		Short: "Unfreeze a card",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			response, err := application.CardsService.Unfreeze(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			return printer.PrintJSON(response)
		},
	}
}

func newCardsSpendingLimitsCommand(printer output.Printer, application *app.App) *cobra.Command {
	var input cards.SpendingLimits
	var allowedCategories []string
	var blockedCategories []string
	var allowedMerchants []string
	var blockedMerchants []string

	cmd := &cobra.Command{
		Use:   "spending-limits <card-id>",
		Short: "Update spending limits for a card",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			input.AllowedCategories = allowedCategories
			input.BlockedCategories = blockedCategories
			input.AllowedMerchants = allowedMerchants
			input.BlockedMerchants = blockedMerchants

			response, err := application.CardsService.UpdateSpendingLimits(cmd.Context(), args[0], input)
			if err != nil {
				return err
			}
			return printer.PrintJSON(response)
		},
	}

	cmd.Flags().StringVar(&input.SingleTransaction, "single-transaction", "", "Single transaction spending limit")
	cmd.Flags().StringVar(&input.Daily, "daily", "", "Daily spending limit")
	cmd.Flags().StringVar(&input.Weekly, "weekly", "", "Weekly spending limit")
	cmd.Flags().StringVar(&input.Monthly, "monthly", "", "Monthly spending limit")
	cmd.Flags().StringSliceVar(&allowedCategories, "allowed-categories", nil, "Allowed merchant categories")
	cmd.Flags().StringSliceVar(&blockedCategories, "blocked-categories", nil, "Blocked merchant categories")
	cmd.Flags().StringSliceVar(&allowedMerchants, "allowed-merchants", nil, "Allowed merchant IDs")
	cmd.Flags().StringSliceVar(&blockedMerchants, "blocked-merchants", nil, "Blocked merchant IDs")
	return cmd
}

func newCardsCustomerCommand(printer output.Printer, application *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "customer <customer-id>",
		Short: "List cards for a customer",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			response, err := application.CardsService.GetCustomerCards(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			return printer.PrintJSON(response)
		},
	}
}

func stringsTitle(value string) string {
	if value == "" {
		return value
	}
	return strings.ToUpper(value[:1]) + value[1:]
}

func parseCardAmountToMicroUnits(value string) (int64, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, fmt.Errorf("amount is empty")
	}
	if strings.HasPrefix(value, "-") {
		return 0, fmt.Errorf("amount must be positive")
	}

	parts := strings.Split(value, ".")
	if len(parts) > 2 {
		return 0, fmt.Errorf("invalid decimal amount %q", value)
	}

	wholePart := parts[0]
	if wholePart == "" {
		wholePart = "0"
	}

	fractional := ""
	if len(parts) == 2 {
		fractional = parts[1]
	}
	if len(fractional) > 6 {
		return 0, fmt.Errorf("amount supports at most 6 decimal places")
	}
	fractional += strings.Repeat("0", 6-len(fractional))

	wholeUnits, err := strconv.ParseInt(wholePart, 10, 64)
	if err != nil {
		return 0, err
	}
	fractionalUnits, err := strconv.ParseInt(fractional, 10, 64)
	if err != nil {
		return 0, err
	}

	return wholeUnits*1_000_000 + fractionalUnits, nil
}
