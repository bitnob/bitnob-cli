package cli

import (
	"fmt"

	"github.com/bitnob/bitnob-cli/internal/app"
	"github.com/bitnob/bitnob-cli/internal/output"
	"github.com/bitnob/bitnob-cli/internal/trading"
	"github.com/spf13/cobra"
)

func newTradingCommand(printer output.Printer, application *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "trading",
		Short: "Manage trading quotes, orders, and prices",
	}

	cmd.AddCommand(
		newTradingQuotesCommand(printer, application),
		newTradingOrdersCommand(printer, application),
		newTradingPricesCommand(printer, application),
	)

	return cmd
}

func newTradingQuotesCommand(printer output.Printer, application *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "quotes",
		Short: "Manage trading quotes",
	}

	cmd.AddCommand(newTradingQuotesCreateCommand(printer, application))
	return cmd
}

func newTradingQuotesCreateCommand(printer output.Printer, application *app.App) *cobra.Command {
	var input trading.CreateQuoteInput

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a trading quote",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if input.BaseCurrency == "" || input.QuoteCurrency == "" || input.Side == "" || input.Quantity == "" {
				return fmt.Errorf("base-currency, quote-currency, side, and quantity are required")
			}
			if input.Side != "buy" && input.Side != "sell" {
				return fmt.Errorf("side must be one of: buy, sell")
			}

			response, err := application.TradingService.CreateQuote(cmd.Context(), input)
			if err != nil {
				return err
			}
			return printer.PrintJSON(response)
		},
	}

	cmd.Flags().StringVar(&input.BaseCurrency, "base-currency", "", "Base currency of the pair, for example BTC")
	cmd.Flags().StringVar(&input.QuoteCurrency, "quote-currency", "", "Quote currency of the pair, for example USDT")
	cmd.Flags().StringVar(&input.Side, "side", "", "Trade side: buy or sell")
	cmd.Flags().StringVar(&input.Quantity, "quantity", "", "Base currency quantity")
	return cmd
}

func newTradingOrdersCommand(printer output.Printer, application *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "orders",
		Short: "Manage trading orders",
	}

	cmd.AddCommand(
		newTradingOrdersCreateCommand(printer, application),
		newTradingOrdersListCommand(printer, application),
		newTradingOrdersGetCommand(printer, application),
	)
	return cmd
}

func newTradingOrdersCreateCommand(printer output.Printer, application *app.App) *cobra.Command {
	var input trading.CreateOrderInput

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a trading order",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if input.BaseCurrency == "" || input.QuoteCurrency == "" || input.Side == "" || input.Quantity == "" || input.Price == "" || input.QuoteID == "" {
				return fmt.Errorf("base-currency, quote-currency, side, quantity, price, and quote-id are required")
			}
			if input.Side != "buy" && input.Side != "sell" {
				return fmt.Errorf("side must be one of: buy, sell")
			}

			response, err := application.TradingService.CreateOrder(cmd.Context(), input)
			if err != nil {
				return err
			}
			return printer.PrintJSON(response)
		},
	}

	cmd.Flags().StringVar(&input.BaseCurrency, "base-currency", "", "Base currency of the pair")
	cmd.Flags().StringVar(&input.QuoteCurrency, "quote-currency", "", "Quote currency of the pair")
	cmd.Flags().StringVar(&input.Side, "side", "", "Order side: buy or sell")
	cmd.Flags().StringVar(&input.Quantity, "quantity", "", "Base currency quantity")
	cmd.Flags().StringVar(&input.Price, "price", "", "Quote price per unit")
	cmd.Flags().StringVar(&input.QuoteID, "quote-id", "", "Quote ID to execute against")
	cmd.Flags().StringVar(&input.Reference, "reference", "", "Reference for the order")
	return cmd
}

func newTradingOrdersListCommand(printer output.Printer, application *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List trading orders",
		RunE: func(cmd *cobra.Command, _ []string) error {
			response, err := application.TradingService.ListOrders(cmd.Context())
			if err != nil {
				return err
			}
			return printer.PrintJSON(response)
		},
	}
}

func newTradingOrdersGetCommand(printer output.Printer, application *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get a trading order by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			response, err := application.TradingService.GetOrder(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			return printer.PrintJSON(response)
		},
	}
}

func newTradingPricesCommand(printer output.Printer, application *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "prices",
		Short: "List current trading prices",
		RunE: func(cmd *cobra.Command, _ []string) error {
			response, err := application.TradingService.ListPrices(cmd.Context())
			if err != nil {
				return err
			}
			return printer.PrintJSON(response)
		},
	}
}
