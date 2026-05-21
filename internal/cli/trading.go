package cli

import (
	"context"
	"fmt"
	"math/big"
	"net/url"
	"strings"

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
		newTradingScheduledOrdersCommand(printer, application),
		newTradingTargetOrdersCommand(printer, application),
	)

	return cmd
}

func newTradingQuotesCommand(printer output.Printer, application *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "quotes",
		Short: "Manage trading quotes",
	}

	cmd.AddCommand(
		newTradingQuotesCreateCommand(printer, application),
		newTradingQuotesGetCommand(printer, application),
	)
	return cmd
}

func newTradingQuotesCreateCommand(printer output.Printer, application *app.App) *cobra.Command {
	var input trading.CreateQuoteInput
	var quoteAmount string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a trading quote",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if input.BaseCurrency == "" || input.QuoteCurrency == "" || input.Side == "" {
				return fmt.Errorf("base-currency, quote-currency, and side are required")
			}

			input.Side = strings.ToLower(strings.TrimSpace(input.Side))
			if input.Side != "buy" && input.Side != "sell" {
				return fmt.Errorf("side must be one of: buy, sell")
			}

			hasQuantity := strings.TrimSpace(input.Quantity) != ""
			hasQuoteAmount := strings.TrimSpace(quoteAmount) != ""
			if hasQuantity == hasQuoteAmount {
				return fmt.Errorf("provide exactly one of: --quantity (base amount) or --quote-amount")
			}

			if hasQuoteAmount {
				quantity, err := deriveBaseQuantityFromQuoteAmount(cmd.Context(), application.TradingService, input.BaseCurrency, input.QuoteCurrency, quoteAmount)
				if err != nil {
					return err
				}
				input.Quantity = quantity
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
	cmd.Flags().StringVar(&input.Quantity, "quantity", "", "Base currency quantity (mutually exclusive with --quote-amount)")
	cmd.Flags().StringVar(&quoteAmount, "quote-amount", "", "Quote currency amount, for example 15 USDT for BTC/USDT (mutually exclusive with --quantity)")
	return cmd
}

func deriveBaseQuantityFromQuoteAmount(ctx context.Context, tradingService *trading.Service, baseCurrency string, quoteCurrency string, quoteAmountRaw string) (string, error) {
	quoteAmount, err := parsePositiveDecimal("quote-amount", quoteAmountRaw)
	if err != nil {
		return "", err
	}

	response, err := tradingService.ListPrices(ctx)
	if err != nil {
		return "", fmt.Errorf("resolve quote amount conversion price: %w", err)
	}

	price, err := findPairPrice(baseCurrency, quoteCurrency, response.Data.Prices)
	if err != nil {
		return "", err
	}
	if price.Sign() <= 0 {
		return "", fmt.Errorf("invalid non-positive price for %s/%s", strings.ToUpper(baseCurrency), strings.ToUpper(quoteCurrency))
	}

	quantity := new(big.Rat).Quo(quoteAmount, price)
	if quantity.Sign() <= 0 {
		return "", fmt.Errorf("derived non-positive base quantity for quote amount conversion")
	}

	return trimTrailingZeros(quantity.FloatString(18)), nil
}

func findPairPrice(baseCurrency string, quoteCurrency string, prices []trading.Price) (*big.Rat, error) {
	for _, entry := range prices {
		if strings.EqualFold(entry.BaseCurrency, baseCurrency) && strings.EqualFold(entry.QuoteCurrency, quoteCurrency) {
			return parsePositiveDecimal("price", entry.Price)
		}
	}

	for _, entry := range prices {
		if strings.EqualFold(entry.BaseCurrency, quoteCurrency) && strings.EqualFold(entry.QuoteCurrency, baseCurrency) {
			inverse, err := parsePositiveDecimal("price", entry.Price)
			if err != nil {
				return nil, err
			}
			return new(big.Rat).Inv(inverse), nil
		}
	}

	return nil, fmt.Errorf("no trading price available for pair %s/%s", strings.ToUpper(baseCurrency), strings.ToUpper(quoteCurrency))
}

func parsePositiveDecimal(name string, raw string) (*big.Rat, error) {
	value := strings.TrimSpace(raw)
	r := new(big.Rat)
	if _, ok := r.SetString(value); !ok {
		return nil, fmt.Errorf("%s must be a valid decimal value", name)
	}
	if r.Sign() <= 0 {
		return nil, fmt.Errorf("%s must be greater than zero", name)
	}
	return r, nil
}

func trimTrailingZeros(value string) string {
	value = strings.TrimRight(value, "0")
	value = strings.TrimRight(value, ".")
	if value == "" {
		return "0"
	}
	return value
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
	var status string
	var limit int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List trading orders",
		RunE: func(cmd *cobra.Command, _ []string) error {
			response, err := application.TradingService.ListOrders(cmd.Context(), status, limit)
			if err != nil {
				return err
			}
			return printer.PrintJSON(response)
		},
	}
	cmd.Flags().StringVar(&status, "status", "", "Filter by order status")
	cmd.Flags().IntVar(&limit, "limit", 0, "Maximum orders to return")
	return cmd
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
	cmd := &cobra.Command{
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

	cmd.AddCommand(newTradingPricesGetCommand(printer, application))
	return cmd
}

func newTradingQuotesGetCommand(printer output.Printer, application *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get a trading quote by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			response, err := application.TradingService.GetQuote(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			return printer.PrintJSON(response)
		},
	}
}

func newTradingPricesGetCommand(printer output.Printer, application *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "get <pair>",
		Short: "Get price by pair, for example BTC-USDT",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			response, err := application.TradingService.GetPriceByPair(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			return printer.PrintJSON(response)
		},
	}
}

func newTradingScheduledOrdersCommand(printer output.Printer, application *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "scheduled-orders",
		Short: "Manage trading scheduled orders",
	}

	cmd.AddCommand(
		newTradingScheduledOrdersCreateCommand(printer, application),
		newTradingScheduledOrdersListCommand(printer, application),
		newTradingScheduledOrdersGetCommand(printer, application),
		newTradingScheduledOrdersUpdateCommand(printer, application),
		newTradingScheduledOrdersExecutionsCommand(printer, application),
	)
	return cmd
}

func newTradingScheduledOrdersCreateCommand(printer output.Printer, application *app.App) *cobra.Command {
	var data string
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create scheduled order",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runRawRequest(cmd.Context(), printer, application, "POST", "/api/trading/scheduled-orders", data)
		},
	}
	cmd.Flags().StringVar(&data, "data", "", "JSON request body")
	return cmd
}

func newTradingScheduledOrdersListCommand(printer output.Printer, application *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List scheduled orders",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runRawRequest(cmd.Context(), printer, application, "GET", "/api/trading/scheduled-orders", "")
		},
	}
}

func newTradingScheduledOrdersGetCommand(printer output.Printer, application *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get scheduled order",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRawRequest(cmd.Context(), printer, application, "GET", rawAPIPath("api", "trading", "scheduled-orders", args[0]), "")
		},
	}
}

func newTradingScheduledOrdersUpdateCommand(printer output.Printer, application *app.App) *cobra.Command {
	var data string
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update scheduled order",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRawRequest(cmd.Context(), printer, application, "PUT", rawAPIPath("api", "trading", "scheduled-orders", args[0]), data)
		},
	}
	cmd.Flags().StringVar(&data, "data", "", "JSON request body")
	return cmd
}

func newTradingScheduledOrdersExecutionsCommand(printer output.Printer, application *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "executions <id>",
		Short: "List scheduled-order executions",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRawRequest(cmd.Context(), printer, application, "GET", rawAPIPath("api", "trading", "scheduled-orders", args[0], "executions"), "")
		},
	}
}

func newTradingTargetOrdersCommand(printer output.Printer, application *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "target-orders",
		Short: "Manage trading target orders",
	}
	cmd.AddCommand(
		newTradingTargetOrdersCreateCommand(printer, application),
		newTradingTargetOrdersListCommand(printer, application),
		newTradingTargetOrdersGetCommand(printer, application),
	)
	return cmd
}

func newTradingTargetOrdersCreateCommand(printer output.Printer, application *app.App) *cobra.Command {
	var data string
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create target order",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runRawRequest(cmd.Context(), printer, application, "POST", "/api/trading/target-orders", data)
		},
	}
	cmd.Flags().StringVar(&data, "data", "", "JSON request body")
	return cmd
}

func newTradingTargetOrdersListCommand(printer output.Printer, application *app.App) *cobra.Command {
	var status string
	var limit int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List target orders",
		RunE: func(cmd *cobra.Command, _ []string) error {
			query := url.Values{}
			if status != "" {
				query.Set("status", status)
			}
			if limit > 0 {
				query.Set("limit", fmt.Sprintf("%d", limit))
			}
			path := "/api/trading/target-orders"
			if encoded := query.Encode(); encoded != "" {
				path += "?" + encoded
			}
			return runRawRequest(cmd.Context(), printer, application, "GET", path, "")
		},
	}
	cmd.Flags().StringVar(&status, "status", "", "Filter by target-order status")
	cmd.Flags().IntVar(&limit, "limit", 0, "Maximum target orders to return")
	return cmd
}

func newTradingTargetOrdersGetCommand(printer output.Printer, application *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get target order",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRawRequest(cmd.Context(), printer, application, "GET", rawAPIPath("api", "trading", "target-orders", args[0]), "")
		},
	}
}
