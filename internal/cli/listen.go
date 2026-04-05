package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bitnob/bitnob-cli/internal/app"
	"github.com/bitnob/bitnob-cli/internal/output"
	"github.com/bitnob/bitnob-cli/internal/webhook"
	"github.com/spf13/cobra"
)

var runWebhookServer = func(ctx context.Context, server *http.Server) error {
	errCh := make(chan error, 1)
	go func() {
		err := server.ListenAndServe()
		if err == http.ErrServerClosed {
			err = nil
		}
		errCh <- err
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
		return <-errCh
	case err := <-errCh:
		return err
	}
}

func newListenCommand(printer output.Printer, application *app.App) *cobra.Command {
	var addr string
	var path string
	var secret string
	var forwardTo string

	cmd := &cobra.Command{
		Use:   "listen",
		Short: "Listen for Bitnob webhooks locally",
		Long: `Start a local HTTP endpoint for Bitnob webhooks.

The listener verifies the x-bitnob-signature header with HMAC-SHA512 using your webhook secret.
If --forward-to is provided, verified events are forwarded to that local URL.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if secret == "" {
				secret = os.Getenv("BITNOB_WEBHOOK_SECRET")
			}
			if secret == "" {
				return fmt.Errorf("webhook secret is required; pass --secret or set BITNOB_WEBHOOK_SECRET")
			}
			if path == "" || path[0] != '/' {
				return fmt.Errorf("path must start with /")
			}

			logger := func(entry webhook.EventLog) error {
				return json.NewEncoder(printer.Stdout).Encode(entry)
			}

			mux := http.NewServeMux()
			mux.Handle(path, application.WebhookService.Handler(webhook.Config{
				Secret:    secret,
				ForwardTo: forwardTo,
			}, logger))

			server := &http.Server{
				Addr:    addr,
				Handler: mux,
			}

			if err := printer.PrintJSON(map[string]any{
				"status":     "listening",
				"address":    addr,
				"path":       path,
				"forward_to": forwardTo,
			}); err != nil {
				return err
			}

			ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
			defer stop()
			return runWebhookServer(ctx, server)
		},
	}

	cmd.Flags().StringVar(&addr, "addr", "127.0.0.1:8080", "Address for the local webhook listener")
	cmd.Flags().StringVar(&path, "path", "/webhook", "Path for incoming Bitnob webhook requests")
	cmd.Flags().StringVar(&secret, "secret", "", "Webhook secret used to verify x-bitnob-signature")
	cmd.Flags().StringVar(&forwardTo, "forward-to", "", "Optional local URL to forward verified events to")
	return cmd
}
