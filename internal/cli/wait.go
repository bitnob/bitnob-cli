package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/bitnob/bitnob-cli/internal/app"
	"github.com/bitnob/bitnob-cli/internal/output"
	"github.com/spf13/cobra"
)

func newWaitCommand(printer output.Printer, application *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "wait",
		Short: "Poll until a Bitnob resource reaches a target state",
	}

	cmd.AddCommand(
		newWaitCardActiveCommand(printer, application),
		newWaitLightningPaidCommand(printer, application),
		newWaitTransactionSettledCommand(printer, application),
	)

	return cmd
}

func newWaitCardActiveCommand(printer output.Printer, application *app.App) *cobra.Command {
	var timeout time.Duration
	var interval time.Duration
	var watch bool

	cmd := &cobra.Command{
		Use:   "card-active <card-id>",
		Short: "Wait until a card becomes active",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := waitFor(cmd, timeout, interval, watch, printer, func() (any, bool, string, error) {
				response, err := application.CardsService.Get(cmd.Context(), args[0])
				if err != nil {
					return nil, false, "", err
				}
				status := strings.ToLower(strings.TrimSpace(response.Data.Card.Status))
				return response, status == "active", status, nil
			})
			if err != nil {
				return err
			}
			return printer.PrintJSON(result)
		},
	}

	addWaitFlags(cmd, &timeout, &interval, &watch)
	return cmd
}

func newWaitLightningPaidCommand(printer output.Printer, application *app.App) *cobra.Command {
	var timeout time.Duration
	var interval time.Duration
	var watch bool

	cmd := &cobra.Command{
		Use:   "lightning-paid <invoice-id>",
		Short: "Wait until a Lightning invoice is paid",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := waitFor(cmd, timeout, interval, watch, printer, func() (any, bool, string, error) {
				response, err := application.LightningService.GetInvoice(cmd.Context(), args[0])
				if err != nil {
					return nil, false, "", err
				}
				status := strings.ToLower(strings.TrimSpace(response.Data.Status))
				return response, status == "paid", status, nil
			})
			if err != nil {
				return err
			}
			return printer.PrintJSON(result)
		},
	}

	addWaitFlags(cmd, &timeout, &interval, &watch)
	return cmd
}

func newWaitTransactionSettledCommand(printer output.Printer, application *app.App) *cobra.Command {
	var timeout time.Duration
	var interval time.Duration
	var watch bool

	cmd := &cobra.Command{
		Use:   "transaction-settled <id-or-reference>",
		Short: "Wait until a transaction reaches SETTLED state",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := waitFor(cmd, timeout, interval, watch, printer, func() (any, bool, string, error) {
				response, err := application.TransactionsService.Get(cmd.Context(), args[0])
				if err != nil {
					return nil, false, "", err
				}
				state := strings.ToUpper(strings.TrimSpace(response.Data.State))
				return response, state == "SETTLED", state, nil
			})
			if err != nil {
				return err
			}
			return printer.PrintJSON(result)
		},
	}

	addWaitFlags(cmd, &timeout, &interval, &watch)
	return cmd
}

func addWaitFlags(cmd *cobra.Command, timeout *time.Duration, interval *time.Duration, watch *bool) {
	cmd.Flags().DurationVar(timeout, "timeout", 2*time.Minute, "Maximum time to wait")
	cmd.Flags().DurationVar(interval, "interval", 5*time.Second, "Polling interval")
	cmd.Flags().BoolVar(watch, "watch", false, "Emit status snapshots while polling")
}

func waitFor(cmd *cobra.Command, timeout time.Duration, interval time.Duration, watch bool, printer output.Printer, poll func() (any, bool, string, error)) (any, error) {
	if timeout <= 0 {
		return nil, fmt.Errorf("timeout must be greater than zero")
	}
	if interval <= 0 {
		return nil, fmt.Errorf("interval must be greater than zero")
	}

	deadline := time.Now().Add(timeout)
	attempt := 0

	for {
		attempt++
		response, done, state, err := poll()
		if err != nil {
			return nil, err
		}

		if watch {
			if err := printer.PrintJSON(map[string]any{
				"status":    "waiting",
				"attempt":   attempt,
				"state":     state,
				"remaining": maxDurationString(time.Until(deadline)),
			}); err != nil {
				return nil, err
			}
		}

		if done {
			return response, nil
		}

		if time.Now().Add(interval).After(deadline) {
			return nil, fmt.Errorf("timed out after %s waiting for state %q", timeout, state)
		}

		select {
		case <-cmd.Context().Done():
			return nil, cmd.Context().Err()
		case <-time.After(interval):
		}
	}
}

func maxDurationString(d time.Duration) string {
	if d < 0 {
		return "0s"
	}
	return d.Round(time.Second).String()
}
