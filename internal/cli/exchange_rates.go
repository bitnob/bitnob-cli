package cli

import (
	"net/url"

	"github.com/bitnob/bitnob-cli/internal/app"
	"github.com/bitnob/bitnob-cli/internal/output"
	"github.com/spf13/cobra"
)

func newExchangeRatesCommand(printer output.Printer, application *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "exchange-rates",
		Short: "Get exchange rates and conversions",
	}

	cmd.AddCommand(
		newExchangeRatesGetCommand(printer, application),
		newExchangeRatesConvertCommand(printer, application),
	)
	return cmd
}

func newExchangeRatesGetCommand(printer output.Printer, application *app.App) *cobra.Command {
	var from string
	var to string

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get single-pair exchange rate",
		RunE: func(cmd *cobra.Command, _ []string) error {
			query := url.Values{}
			if from != "" {
				query.Set("from", from)
			}
			if to != "" {
				query.Set("to", to)
			}
			path := "/api/exchange-rates"
			if encoded := query.Encode(); encoded != "" {
				path += "?" + encoded
			}
			return runRawRequest(cmd.Context(), printer, application, "GET", path, "")
		},
	}
	cmd.Flags().StringVar(&from, "from", "", "Base asset/currency")
	cmd.Flags().StringVar(&to, "to", "", "Quote/target currency")
	return cmd
}

func newExchangeRatesConvertCommand(printer output.Printer, application *app.App) *cobra.Command {
	var from string
	var to string
	var amount string

	cmd := &cobra.Command{
		Use:   "convert",
		Short: "Convert amount between currencies",
		RunE: func(cmd *cobra.Command, _ []string) error {
			query := url.Values{}
			if from != "" {
				query.Set("from", from)
			}
			if to != "" {
				query.Set("to", to)
			}
			if amount != "" {
				query.Set("amount", amount)
			}
			path := "/api/exchange-rates/convert"
			if encoded := query.Encode(); encoded != "" {
				path += "?" + encoded
			}
			return runRawRequest(cmd.Context(), printer, application, "GET", path, "")
		},
	}
	cmd.Flags().StringVar(&from, "from", "", "Base asset/currency")
	cmd.Flags().StringVar(&to, "to", "", "Quote/target currency")
	cmd.Flags().StringVar(&amount, "amount", "", "Amount to convert")
	return cmd
}
