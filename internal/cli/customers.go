package cli

import (
	"fmt"
	"strings"

	"github.com/bitnob/bitnob-cli/internal/app"
	"github.com/bitnob/bitnob-cli/internal/customers"
	"github.com/bitnob/bitnob-cli/internal/output"
	"github.com/spf13/cobra"
)

func newCustomersCommand(printer output.Printer, application *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "customers",
		Short: "Manage Bitnob customers",
	}

	cmd.AddCommand(
		newCustomersCountriesCommand(printer, application),
		newCustomersListCommand(printer, application),
		newCustomersGetCommand(printer, application),
		newCustomersCreateCommand(printer, application),
		newCustomersUpdateCommand(printer, application),
	)

	return cmd
}

func newCustomersCountriesCommand(printer output.Printer, application *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "countries",
		Short: "List supported customer countries",
		RunE: func(cmd *cobra.Command, _ []string) error {
			response, err := application.CustomersService.ListCountries(cmd.Context())
			if err != nil {
				return err
			}
			return printer.PrintJSON(response)
		},
	}

	cmd.AddCommand(newCustomersCountryStatesCommand(printer, application))
	return cmd
}

func newCustomersCountryStatesCommand(printer output.Printer, application *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "states <country-code>",
		Short: "List states for a customer country",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			response, err := application.CustomersService.ListStatesByCountry(cmd.Context(), strings.ToUpper(strings.TrimSpace(args[0])))
			if err != nil {
				return err
			}
			return printer.PrintJSON(response)
		},
	}
}

func newCustomersListCommand(printer output.Printer, application *app.App) *cobra.Command {
	var page int
	var limit int
	var blacklist bool
	var blacklistSet bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List customers",
		RunE: func(cmd *cobra.Command, _ []string) error {
			var filter *bool
			if cmd.Flags().Changed("blacklist") {
				blacklistSet = true
			}
			if blacklistSet {
				filter = &blacklist
			}

			response, err := application.CustomersService.List(cmd.Context(), page, limit, filter)
			if err != nil {
				return err
			}
			return printer.PrintJSON(response)
		},
	}

	cmd.Flags().IntVar(&page, "page", 0, "Page number")
	cmd.Flags().IntVar(&limit, "limit", 0, "Number of items to return")
	cmd.Flags().BoolVar(&blacklist, "blacklist", false, "Filter customers by blacklist status")
	return cmd
}

func newCustomersGetCommand(printer output.Printer, application *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "get <id-or-ref-or-email>",
		Short: "Get a customer by ID, reference, or email",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			response, err := application.CustomersService.Get(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			return printer.PrintJSON(response)
		},
	}
}

func newCustomersCreateCommand(printer output.Printer, application *app.App) *cobra.Command {
	var input customers.CreateCustomerInput

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new customer",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if input.CustomerType != "individual" && input.CustomerType != "business" {
				return fmt.Errorf("customer-type must be one of: individual, business")
			}
			if input.Email == "" {
				return fmt.Errorf("email is required")
			}

			response, err := application.CustomersService.Create(cmd.Context(), input)
			if err != nil {
				return err
			}
			return printer.PrintJSON(response)
		},
	}

	cmd.Flags().StringVar(&input.CustomerType, "customer-type", "", "Customer type: individual or business")
	cmd.Flags().StringVar(&input.FirstName, "first-name", "", "Customer first name")
	cmd.Flags().StringVar(&input.LastName, "last-name", "", "Customer last name")
	cmd.Flags().StringVar(&input.Email, "email", "", "Customer email")
	cmd.Flags().StringVar(&input.Phone, "phone", "", "Customer phone")
	cmd.Flags().StringVar(&input.CountryCode, "country-code", "", "Customer country code")
	return cmd
}

func newCustomersUpdateCommand(printer output.Printer, application *app.App) *cobra.Command {
	var input customers.UpdateCustomerInput

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an existing customer",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if input.FirstName == "" || input.LastName == "" || input.Phone == "" || input.Email == "" {
				return fmt.Errorf("first-name, last-name, phone, and email are required")
			}

			response, err := application.CustomersService.Update(cmd.Context(), args[0], input)
			if err != nil {
				return err
			}
			return printer.PrintJSON(response)
		},
	}

	cmd.Flags().StringVar(&input.FirstName, "first-name", "", "Customer first name")
	cmd.Flags().StringVar(&input.LastName, "last-name", "", "Customer last name")
	cmd.Flags().StringVar(&input.Phone, "phone", "", "Customer phone")
	cmd.Flags().StringVar(&input.Email, "email", "", "Customer email")
	return cmd
}
