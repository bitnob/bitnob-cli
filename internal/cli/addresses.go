package cli

import (
	"fmt"

	"github.com/bitnob/bitnob-cli/internal/addresses"
	"github.com/bitnob/bitnob-cli/internal/app"
	"github.com/bitnob/bitnob-cli/internal/output"
	"github.com/spf13/cobra"
)

func newAddressesCommand(printer output.Printer, application *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "addresses",
		Short: "Manage blockchain addresses",
	}

	cmd.AddCommand(
		newAddressesListCommand(printer, application),
		newAddressesGetCommand(printer, application),
		newAddressesCreateCommand(printer, application),
		newAddressesValidateCommand(printer, application),
		newAddressesSupportedChainsCommand(printer, application),
	)

	return cmd
}

func newAddressesListCommand(printer output.Printer, application *app.App) *cobra.Command {
	var page int
	var limit int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List addresses",
		RunE: func(cmd *cobra.Command, _ []string) error {
			response, err := application.AddressesService.List(cmd.Context(), page, limit)
			if err != nil {
				return err
			}
			return printer.PrintJSON(response)
		},
	}

	cmd.Flags().IntVar(&page, "page", 0, "Page number")
	cmd.Flags().IntVar(&limit, "limit", 0, "Number of addresses to return")
	return cmd
}

func newAddressesGetCommand(printer output.Printer, application *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "get <id-or-address>",
		Short: "Get an address by ID or address string",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			response, err := application.AddressesService.Get(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			return printer.PrintJSON(response)
		},
	}
}

func newAddressesCreateCommand(printer output.Printer, application *app.App) *cobra.Command {
	var input addresses.CreateInput

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Generate an address",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if input.Chain == "" {
				return fmt.Errorf("chain is required")
			}

			response, err := application.AddressesService.Create(cmd.Context(), input)
			if err != nil {
				return err
			}
			return printer.PrintJSON(response)
		},
	}

	cmd.Flags().StringVar(&input.Chain, "chain", "", "Blockchain network for the address")
	cmd.Flags().StringVar(&input.Label, "label", "", "Address label")
	cmd.Flags().StringVar(&input.Reference, "reference", "", "Reference string for the address")
	return cmd
}

func newAddressesValidateCommand(printer output.Printer, application *app.App) *cobra.Command {
	var input addresses.ValidateInput

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate an address for a chain",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if input.Address == "" || input.Chain == "" {
				return fmt.Errorf("address and chain are required")
			}

			response, err := application.AddressesService.Validate(cmd.Context(), input)
			if err != nil {
				return err
			}
			return printer.PrintJSON(response)
		},
	}

	cmd.Flags().StringVar(&input.Address, "address", "", "Blockchain address to validate")
	cmd.Flags().StringVar(&input.Chain, "chain", "", "Blockchain network for validation")
	return cmd
}

func newAddressesSupportedChainsCommand(printer output.Printer, application *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "supported-chains",
		Short: "List supported stablecoin chains",
		RunE: func(cmd *cobra.Command, _ []string) error {
			response, err := application.AddressesService.SupportedChains(cmd.Context())
			if err != nil {
				return err
			}
			return printer.PrintJSON(response)
		},
	}
}
