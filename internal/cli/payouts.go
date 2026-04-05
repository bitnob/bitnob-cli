package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/bitnob/bitnob-cli/internal/app"
	"github.com/bitnob/bitnob-cli/internal/output"
	"github.com/bitnob/bitnob-cli/internal/payouts"
	"github.com/spf13/cobra"
)

func newPayoutsCommand(printer output.Printer, application *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "payouts",
		Short: "Manage payouts quotes, beneficiaries, and limits",
	}

	cmd.AddCommand(
		newPayoutsQuotesCommand(printer, application),
		newPayoutsInitializeCommand(printer, application),
		newPayoutsBeneficiaryLookupCommand(printer, application),
		newPayoutsSimulateDepositCommand(printer, application),
		newPayoutsFetchCommand(printer, application),
		newPayoutsCountryRequirementsCommand(printer, application),
		newPayoutsSupportedCountriesCommand(printer, application),
		newPayoutsLimitsCommand(printer, application),
	)

	return cmd
}

func newPayoutsQuotesCommand(printer output.Printer, application *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "quotes",
		Short: "Manage payout quotes",
	}

	cmd.AddCommand(
		newPayoutsQuotesCreateCommand(printer, application),
		newPayoutsQuotesListCommand(printer, application),
		newPayoutsQuotesGetCommand(printer, application),
		newPayoutsQuotesFinalizeCommand(printer, application),
	)

	return cmd
}

func newPayoutsQuotesCreateCommand(printer output.Printer, application *app.App) *cobra.Command {
	var input payouts.CreateQuoteInput

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a payout quote",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if input.FromAsset == "" || input.ToCurrency == "" || input.Source == "" || input.PaymentReason == "" || input.Reference == "" || input.Country == "" {
				return fmt.Errorf("from-asset, to-currency, source, payment-reason, reference, and country are required")
			}
			if input.Amount == "" && input.SettlementAmount == "" {
				return fmt.Errorf("amount or settlement-amount is required")
			}

			response, err := application.PayoutsService.CreateQuote(cmd.Context(), input)
			if err != nil {
				return err
			}
			return printer.PrintJSON(response)
		},
	}

	cmd.Flags().StringVar(&input.FromAsset, "from-asset", "", "Crypto asset to convert from, for example BTC, USDT, or USDC")
	cmd.Flags().StringVar(&input.ToCurrency, "to-currency", "", "Fiat settlement currency, for example NGN or KES")
	cmd.Flags().StringVar(&input.Source, "source", "", "Funding source: onchain or offchain")
	cmd.Flags().StringVar(&input.Chain, "chain", "", "Blockchain network, required when source is onchain")
	cmd.Flags().StringVar(&input.Amount, "amount", "", "Crypto amount to convert")
	cmd.Flags().StringVar(&input.SettlementAmount, "settlement-amount", "", "Target fiat amount to settle")
	cmd.Flags().StringVar(&input.PaymentReason, "payment-reason", "", "Payment purpose or description")
	cmd.Flags().StringVar(&input.Reference, "reference", "", "Unique quote reference")
	cmd.Flags().StringVar(&input.ClientMetaData, "client-meta-data", "", "JSON string of client metadata")
	cmd.Flags().StringVar(&input.Country, "country", "", "Two-letter destination country code")
	return cmd
}

func newPayoutsQuotesListCommand(printer output.Printer, application *app.App) *cobra.Command {
	var order string
	var page int
	var take int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List payout quotes",
		RunE: func(cmd *cobra.Command, _ []string) error {
			response, err := application.PayoutsService.ListQuotes(cmd.Context(), order, page, take)
			if err != nil {
				return err
			}
			return printer.PrintJSON(response)
		},
	}

	cmd.Flags().StringVar(&order, "order", "", "Sort order: ASC or DESC")
	cmd.Flags().IntVar(&page, "page", 0, "Page number")
	cmd.Flags().IntVar(&take, "take", 0, "Results per page")
	return cmd
}

func newPayoutsQuotesGetCommand(printer output.Printer, application *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "get <quote-id>",
		Short: "Get a payout quote by quote ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			response, err := application.PayoutsService.GetQuoteByQuoteID(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			return printer.PrintJSON(response)
		},
	}
}

func newPayoutsQuotesFinalizeCommand(printer output.Printer, application *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "finalize <quote-id>",
		Short: "Finalize a payout quote",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			response, err := application.PayoutsService.Finalize(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			return printer.PrintJSON(response)
		},
	}
}

func newPayoutsInitializeCommand(printer output.Printer, application *app.App) *cobra.Command {
	var input payouts.InitializeInput
	var beneficiaryJSON string
	var beneficiaryID string
	var beneficiaryCountry string

	cmd := &cobra.Command{
		Use:   "initialize <quote-id>",
		Short: "Initialize a payout from an existing quote",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if input.CustomerID == "" || input.Reference == "" || input.PaymentReason == "" {
				return fmt.Errorf("customer-id, reference, and payment-reason are required")
			}
			if beneficiaryJSON == "" && beneficiaryID == "" {
				return fmt.Errorf("beneficiary or beneficiary-id is required")
			}
			if beneficiaryJSON != "" && beneficiaryID != "" {
				return fmt.Errorf("beneficiary and beneficiary-id are mutually exclusive")
			}

			if beneficiaryID != "" {
				beneficiary, err := application.BeneficiariesService.Get(cmd.Context(), beneficiaryID)
				if err != nil {
					return err
				}
				input.Beneficiary = map[string]any{
					"type":           strings.ToUpper(strings.TrimSpace(beneficiary.Data.Type)),
					"country":        strings.ToUpper(strings.TrimSpace(beneficiaryCountry)),
					"account_name":   beneficiary.Data.Destination.AccountName,
					"account_number": beneficiary.Data.Destination.AccountNumber,
					"bank_code":      beneficiary.Data.Destination.BankCode,
				}
				if input.Beneficiary["country"] == "" {
					return fmt.Errorf("beneficiary-country is required when using beneficiary-id")
				}
			} else {
				beneficiary, err := parseJSONObjectFlag(beneficiaryJSON, "beneficiary")
				if err != nil {
					return err
				}
				input.Beneficiary = beneficiary
			}

			response, err := application.PayoutsService.Initialize(cmd.Context(), args[0], input)
			if err != nil {
				return err
			}
			return printer.PrintJSON(response)
		},
	}

	cmd.Flags().StringVar(&input.CustomerID, "customer-id", "", "Customer ID initiating the payout")
	cmd.Flags().StringVar(&beneficiaryJSON, "beneficiary", "", "Beneficiary JSON object")
	cmd.Flags().StringVar(&beneficiaryID, "beneficiary-id", "", "Saved beneficiary ID to use for this payout")
	cmd.Flags().StringVar(&beneficiaryCountry, "beneficiary-country", "", "Two-letter beneficiary country code when using beneficiary-id")
	cmd.Flags().StringVar(&input.Reference, "reference", "", "Unique payout reference")
	cmd.Flags().StringVar(&input.PaymentReason, "payment-reason", "", "Payment purpose or description")
	cmd.Flags().StringVar(&input.ClientMetaData, "client-meta-data", "", "JSON string of client metadata")
	cmd.Flags().StringVar(&input.CallbackURL, "callback-url", "", "Callback URL for payout updates")
	return cmd
}

func newPayoutsBeneficiaryLookupCommand(printer output.Printer, application *app.App) *cobra.Command {
	var input payouts.BeneficiaryLookupInput

	cmd := &cobra.Command{
		Use:   "beneficiary-lookup",
		Short: "Validate payout beneficiary account details",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if input.Country == "" || input.AccountNumber == "" || input.BankCode == "" || input.Type == "" {
				return fmt.Errorf("country, account-number, bank-code, and type are required")
			}

			response, err := application.PayoutsService.BeneficiaryLookup(cmd.Context(), input)
			if err != nil {
				return err
			}
			return printer.PrintJSON(response)
		},
	}

	cmd.Flags().StringVar(&input.Country, "country", "", "Two-letter country code")
	cmd.Flags().StringVar(&input.AccountNumber, "account-number", "", "Beneficiary account number")
	cmd.Flags().StringVar(&input.BankCode, "bank-code", "", "Beneficiary bank code")
	cmd.Flags().StringVar(&input.Type, "type", "", "Beneficiary destination type, for example bank_account")
	return cmd
}

func newPayoutsSimulateDepositCommand(printer output.Printer, application *app.App) *cobra.Command {
	var input payouts.SimulateDepositInput

	cmd := &cobra.Command{
		Use:   "simulate-deposit",
		Short: "Simulate a payout deposit in sandbox",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if input.QuoteID == "" || input.Amount == "" || input.TxHash == "" {
				return fmt.Errorf("quote-id, amount, and tx-hash are required")
			}

			response, err := application.PayoutsService.SimulateDeposit(cmd.Context(), input)
			if err != nil {
				return err
			}
			return printer.PrintJSON(response)
		},
	}

	cmd.Flags().StringVar(&input.QuoteID, "quote-id", "", "Quote ID to simulate deposit for")
	cmd.Flags().StringVar(&input.Amount, "amount", "", "Crypto amount to simulate")
	cmd.Flags().StringVar(&input.TxHash, "tx-hash", "", "Simulated transaction hash")
	return cmd
}

func newPayoutsFetchCommand(printer output.Printer, application *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fetch",
		Short: "Fetch payout quotes by internal identifiers",
	}

	cmd.AddCommand(
		newPayoutsFetchByIDCommand(printer, application),
		newPayoutsFetchByReferenceCommand(printer, application),
	)

	return cmd
}

func newPayoutsFetchByIDCommand(printer output.Printer, application *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "id <id>",
		Short: "Fetch a payout quote by internal ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			response, err := application.PayoutsService.GetQuoteByID(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			return printer.PrintJSON(response)
		},
	}
}

func newPayoutsFetchByReferenceCommand(printer output.Printer, application *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "reference <reference>",
		Short: "Fetch a payout quote by reference",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			response, err := application.PayoutsService.GetQuoteByReference(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			return printer.PrintJSON(response)
		},
	}
}

func newPayoutsCountryRequirementsCommand(printer output.Printer, application *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "country-requirements <country-code>",
		Short: "Get payout destination requirements for a country",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			response, err := application.PayoutsService.CountryRequirements(cmd.Context(), strings.ToUpper(args[0]))
			if err != nil {
				return err
			}
			return printer.PrintJSON(response)
		},
	}
}

func newPayoutsSupportedCountriesCommand(printer output.Printer, application *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "supported-countries",
		Short: "List payout supported countries",
		RunE: func(cmd *cobra.Command, _ []string) error {
			response, err := application.PayoutsService.SupportedCountries(cmd.Context())
			if err != nil {
				return err
			}
			return printer.PrintJSON(response)
		},
	}
}

func newPayoutsLimitsCommand(printer output.Printer, application *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "limits",
		Short: "List payout transaction limits",
		RunE: func(cmd *cobra.Command, _ []string) error {
			response, err := application.PayoutsService.Limits(cmd.Context())
			if err != nil {
				return err
			}
			return printer.PrintJSON(response)
		},
	}
}

func parseJSONObjectFlag(value string, name string) (map[string]any, error) {
	var parsed map[string]any
	if err := json.Unmarshal([]byte(value), &parsed); err != nil {
		return nil, fmt.Errorf("invalid %s JSON: %w", name, err)
	}
	if len(parsed) == 0 {
		return nil, fmt.Errorf("%s must be a non-empty JSON object", name)
	}
	return parsed, nil
}
