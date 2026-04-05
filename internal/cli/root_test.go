package cli

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bitnob/bitnob-cli/internal/api"
	"github.com/bitnob/bitnob-cli/internal/app"
	"github.com/bitnob/bitnob-cli/internal/output"
	"github.com/bitnob/bitnob-cli/internal/platform"
	"github.com/bitnob/bitnob-cli/internal/version"
)

func TestRunVersion(t *testing.T) {
	application := newTestApp(t)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	err := runWithPrinter(context.Background(), output.New(stdout, stderr), application, []string{"version"})
	if err != nil {
		t.Fatalf("runWithPrinter returned error: %v", err)
	}

	if got := stdout.String(); !strings.Contains(got, `"Version": "test"`) {
		t.Fatalf("unexpected stdout: %q", got)
	}
}

func TestRootLandingPage(t *testing.T) {
	application := newTestApp(t)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	err := runWithPrinter(context.Background(), output.New(stdout, stderr), application, nil)
	if err != nil {
		t.Fatalf("runWithPrinter returned error: %v", err)
	}

	got := stdout.String()
	if !strings.Contains(got, "Bitnob CLI") {
		t.Fatalf("landing page missing title: %q", got)
	}
	if !strings.Contains(got, "bitnob login --interactive") {
		t.Fatalf("landing page missing quick start: %q", got)
	}
}

func TestLoginWhoAmIAndLogout(t *testing.T) {
	application := newTestApp(t)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	printer := output.New(stdout, stderr)
	ctx := context.Background()

	if err := runWithPrinter(ctx, printer, application, []string{"login", "--client-id", "client_test_1234", "--secret-key", "secret_test_12345678"}); err != nil {
		t.Fatalf("login returned error: %v", err)
	}

	stdout.Reset()
	if err := runWithPrinter(ctx, printer, application, []string{"config", "get", "profile.active"}); err != nil {
		t.Fatalf("config get returned error after login: %v", err)
	}

	if got := stdout.String(); !strings.Contains(got, `"value": "sandbox"`) {
		t.Fatalf("unexpected active profile after auto login: %q", got)
	}

	stdout.Reset()
	if err := runWithPrinter(ctx, printer, application, []string{"whoami"}); err != nil {
		t.Fatalf("whoami returned error: %v", err)
	}

	if got := stdout.String(); !strings.Contains(got, `"client_id": "client_test_1234"`) {
		t.Fatalf("unexpected whoami output: %q", got)
	}

	stdout.Reset()
	if err := runWithPrinter(ctx, printer, application, []string{"logout"}); err != nil {
		t.Fatalf("logout returned error: %v", err)
	}
}

func TestSwitchProfile(t *testing.T) {
	application := newTestApp(t)
	ctx := context.Background()

	if err := application.ConfigStore.Save(ctx, applicationConfigWithExtraProfile()); err != nil {
		t.Fatalf("seed config: %v", err)
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	err := runWithPrinter(ctx, output.New(stdout, stderr), application, []string{"switch", "live"})
	if err != nil {
		t.Fatalf("switch returned error: %v", err)
	}

	if got := stdout.String(); !strings.Contains(got, `"profile": "live"`) {
		t.Fatalf("unexpected stdout: %q", got)
	}
}

func TestHelp(t *testing.T) {
	application := newTestApp(t)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	err := runWithPrinter(context.Background(), output.New(stdout, stderr), application, []string{"help"})
	if err != nil {
		t.Fatalf("help returned error: %v", err)
	}

	if got := stdout.String(); !strings.Contains(got, "Available Commands:") {
		t.Fatalf("unexpected help output: %q", got)
	}
}

func TestDoctor(t *testing.T) {
	application := newTestApp(t)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	printer := output.New(stdout, stderr)
	ctx := context.Background()

	if err := runWithPrinter(ctx, printer, application, []string{"login", "--client-id", "client_test_1234", "--secret-key", "secret_test_12345678"}); err != nil {
		t.Fatalf("login returned error: %v", err)
	}

	stdout.Reset()
	if err := runWithPrinter(ctx, printer, application, []string{"doctor"}); err != nil {
		t.Fatalf("doctor returned error: %v", err)
	}

	got := stdout.String()
	if !strings.Contains(got, `"status": "ok"`) {
		t.Fatalf("unexpected doctor output: %q", got)
	}
	if !strings.Contains(got, `"name": "auth.whoami"`) {
		t.Fatalf("doctor output missing whoami check: %q", got)
	}
	if !strings.Contains(got, `"name": "payouts.limits"`) {
		t.Fatalf("doctor output missing payouts probe: %q", got)
	}
}

func TestCompletionInstructions(t *testing.T) {
	application := newTestApp(t)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	originalDetectShell := detectShell
	t.Cleanup(func() {
		detectShell = originalDetectShell
	})
	detectShell = func() string { return "zsh" }

	err := runWithPrinter(context.Background(), output.New(stdout, stderr), application, []string{"completion"})
	if err != nil {
		t.Fatalf("completion returned error: %v", err)
	}

	if got := stdout.String(); !strings.Contains(got, "Detected shell: zsh") {
		t.Fatalf("unexpected completion instructions: %q", got)
	}
}

func TestCompletionScript(t *testing.T) {
	application := newTestApp(t)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	err := runWithPrinter(context.Background(), output.New(stdout, stderr), application, []string{"completion", "--shell", "bash"})
	if err != nil {
		t.Fatalf("completion script returned error: %v", err)
	}

	if got := stdout.String(); !strings.Contains(got, "__start_bitnob") {
		t.Fatalf("unexpected bash completion script: %q", got)
	}
}

func TestConfigListGetAndSet(t *testing.T) {
	application := newTestApp(t)
	ctx := context.Background()

	if err := application.ConfigStore.Save(ctx, applicationConfigWithExtraProfile()); err != nil {
		t.Fatalf("seed config: %v", err)
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	printer := output.New(stdout, stderr)

	if err := runWithPrinter(ctx, printer, application, []string{"config", "get", "profile.active"}); err != nil {
		t.Fatalf("config get returned error: %v", err)
	}
	if got := stdout.String(); !strings.Contains(got, `"value": "default"`) {
		t.Fatalf("unexpected config get output: %q", got)
	}

	stdout.Reset()
	if err := runWithPrinter(ctx, printer, application, []string{"config", "set", "profile.active", "live"}); err != nil {
		t.Fatalf("config set active returned error: %v", err)
	}
	if got := stdout.String(); !strings.Contains(got, `"value": "live"`) {
		t.Fatalf("unexpected config set output: %q", got)
	}

	stdout.Reset()
	if err := runWithPrinter(ctx, printer, application, []string{"config", "list"}); err != nil {
		t.Fatalf("config list returned error: %v", err)
	}
	if got := stdout.String(); !strings.Contains(got, `"active_profile": "live"`) {
		t.Fatalf("unexpected config list output: %q", got)
	}

	stdout.Reset()
	if err := runWithPrinter(ctx, printer, application, []string{"config", "unset", "profile.active"}); err != nil {
		t.Fatalf("config unset returned error: %v", err)
	}
	if got := stdout.String(); !strings.Contains(got, `"value": "default"`) {
		t.Fatalf("unexpected config unset output: %q", got)
	}
}

func TestBalancesJSON(t *testing.T) {
	application := newTestApp(t)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	printer := output.New(stdout, stderr)
	ctx := context.Background()

	if err := runWithPrinter(ctx, printer, application, []string{"login", "--client-id", "client_test_1234", "--secret-key", "secret_test_12345678"}); err != nil {
		t.Fatalf("login returned error: %v", err)
	}

	stdout.Reset()
	if err := runWithPrinter(ctx, printer, application, []string{"balances"}); err != nil {
		t.Fatalf("balances returned error: %v", err)
	}

	if got := stdout.String(); !strings.Contains(got, `"company_id": "company_test"`) {
		t.Fatalf("unexpected balances output: %q", got)
	}
}

func TestBalanceByCurrency(t *testing.T) {
	application := newTestApp(t)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	printer := output.New(stdout, stderr)
	ctx := context.Background()

	if err := runWithPrinter(ctx, printer, application, []string{"login", "--client-id", "client_test_1234", "--secret-key", "secret_test_12345678"}); err != nil {
		t.Fatalf("login returned error: %v", err)
	}

	stdout.Reset()
	if err := runWithPrinter(ctx, printer, application, []string{"balances", "get", "usdt"}); err != nil {
		t.Fatalf("balances get returned error: %v", err)
	}

	if got := stdout.String(); !strings.Contains(got, `"currency": "USDT"`) {
		t.Fatalf("unexpected balances get output: %q", got)
	}
}

func TestWalletsList(t *testing.T) {
	application := newTestApp(t)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	printer := output.New(stdout, stderr)
	ctx := context.Background()

	if err := runWithPrinter(ctx, printer, application, []string{"login", "--client-id", "client_test_1234", "--secret-key", "secret_test_12345678"}); err != nil {
		t.Fatalf("login returned error: %v", err)
	}

	stdout.Reset()
	if err := runWithPrinter(ctx, printer, application, []string{"wallets"}); err != nil {
		t.Fatalf("wallets returned error: %v", err)
	}

	if got := stdout.String(); !strings.Contains(got, `"accounts"`) {
		t.Fatalf("unexpected wallets output: %q", got)
	}
}

func TestLightningInvoiceCreate(t *testing.T) {
	application := newTestApp(t)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	printer := output.New(stdout, stderr)
	ctx := context.Background()

	if err := runWithPrinter(ctx, printer, application, []string{"login", "--client-id", "client_test_1234", "--secret-key", "secret_test_12345678"}); err != nil {
		t.Fatalf("login returned error: %v", err)
	}

	stdout.Reset()
	err := runWithPrinter(ctx, printer, application, []string{"lightning", "invoices", "create", "--amount", "50000", "--description", "Payment for order", "--reference", "order_12345"})
	if err != nil {
		t.Fatalf("lightning invoices create returned error: %v", err)
	}

	if got := stdout.String(); !strings.Contains(got, `"payment_hash": "hash_create_123"`) {
		t.Fatalf("unexpected lightning invoice create output: %q", got)
	}
}

func TestLightningDecode(t *testing.T) {
	application := newTestApp(t)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	printer := output.New(stdout, stderr)
	ctx := context.Background()

	if err := runWithPrinter(ctx, printer, application, []string{"login", "--client-id", "client_test_1234", "--secret-key", "secret_test_12345678"}); err != nil {
		t.Fatalf("login returned error: %v", err)
	}

	stdout.Reset()
	err := runWithPrinter(ctx, printer, application, []string{"lightning", "decode", "--request", "lnbc500u1test"})
	if err != nil {
		t.Fatalf("lightning decode returned error: %v", err)
	}

	if got := stdout.String(); !strings.Contains(got, `"destination": "02e880b95c737bbe110c5ef4f5e70d0d4d04253503e42c4de2760d519f9cb9a81e"`) {
		t.Fatalf("unexpected lightning decode output: %q", got)
	}
}

func TestLightningPaymentSend(t *testing.T) {
	application := newTestApp(t)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	printer := output.New(stdout, stderr)
	ctx := context.Background()

	if err := runWithPrinter(ctx, printer, application, []string{"login", "--client-id", "client_test_1234", "--secret-key", "secret_test_12345678"}); err != nil {
		t.Fatalf("login returned error: %v", err)
	}

	stdout.Reset()
	err := runWithPrinter(ctx, printer, application, []string{"lightning", "payments", "send", "--request", "lnbc1m1test", "--reference", "LT_PT25000900130010051104509"})
	if err != nil {
		t.Fatalf("lightning payments send returned error: %v", err)
	}

	if got := stdout.String(); !strings.Contains(got, `"status": "paid"`) {
		t.Fatalf("unexpected lightning payments send output: %q", got)
	}
}

func TestLightningInvoiceVerify(t *testing.T) {
	application := newTestApp(t)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	printer := output.New(stdout, stderr)
	ctx := context.Background()

	if err := runWithPrinter(ctx, printer, application, []string{"login", "--client-id", "client_test_1234", "--secret-key", "secret_test_12345678"}); err != nil {
		t.Fatalf("login returned error: %v", err)
	}

	stdout.Reset()
	err := runWithPrinter(ctx, printer, application, []string{"lightning", "invoices", "verify", "--payment-hash", "hash_verify_123"})
	if err != nil {
		t.Fatalf("lightning invoices verify returned error: %v", err)
	}

	if got := stdout.String(); !strings.Contains(got, `"is_paid": true`) {
		t.Fatalf("unexpected lightning invoices verify output: %q", got)
	}
}

func TestAddressesList(t *testing.T) {
	application := newTestApp(t)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	printer := output.New(stdout, stderr)
	ctx := context.Background()

	if err := runWithPrinter(ctx, printer, application, []string{"login", "--client-id", "client_test_1234", "--secret-key", "secret_test_12345678"}); err != nil {
		t.Fatalf("login returned error: %v", err)
	}

	stdout.Reset()
	err := runWithPrinter(ctx, printer, application, []string{"addresses", "list", "--page", "1", "--limit", "20"})
	if err != nil {
		t.Fatalf("addresses list returned error: %v", err)
	}

	if got := stdout.String(); !strings.Contains(got, `"addresses"`) {
		t.Fatalf("unexpected addresses list output: %q", got)
	}
}

func TestAddressesCreate(t *testing.T) {
	application := newTestApp(t)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	printer := output.New(stdout, stderr)
	ctx := context.Background()

	if err := runWithPrinter(ctx, printer, application, []string{"login", "--client-id", "client_test_1234", "--secret-key", "secret_test_12345678"}); err != nil {
		t.Fatalf("login returned error: %v", err)
	}

	stdout.Reset()
	err := runWithPrinter(ctx, printer, application, []string{"addresses", "create", "--chain", "bitcoin", "--label", "home", "--reference", "wallet_ref_123"})
	if err != nil {
		t.Fatalf("addresses create returned error: %v", err)
	}

	if got := stdout.String(); !strings.Contains(got, `"address": "0x3752c3f2a2301AaBE0Ab13e2B4308f940Fc2ecCA"`) {
		t.Fatalf("unexpected addresses create output: %q", got)
	}
}

func TestAddressesValidate(t *testing.T) {
	application := newTestApp(t)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	printer := output.New(stdout, stderr)
	ctx := context.Background()

	if err := runWithPrinter(ctx, printer, application, []string{"login", "--client-id", "client_test_1234", "--secret-key", "secret_test_12345678"}); err != nil {
		t.Fatalf("login returned error: %v", err)
	}

	stdout.Reset()
	err := runWithPrinter(ctx, printer, application, []string{"addresses", "validate", "--address", "36FJHVQDuEqRCz8Qcx3ZCkAcHnmC6sgPZjvUVgTUtvHY", "--chain", "solana"})
	if err != nil {
		t.Fatalf("addresses validate returned error: %v", err)
	}

	if got := stdout.String(); !strings.Contains(got, `"valid": true`) {
		t.Fatalf("unexpected addresses validate output: %q", got)
	}
}

func TestAddressesSupportedChains(t *testing.T) {
	application := newTestApp(t)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	printer := output.New(stdout, stderr)
	ctx := context.Background()

	if err := runWithPrinter(ctx, printer, application, []string{"login", "--client-id", "client_test_1234", "--secret-key", "secret_test_12345678"}); err != nil {
		t.Fatalf("login returned error: %v", err)
	}

	stdout.Reset()
	err := runWithPrinter(ctx, printer, application, []string{"addresses", "supported-chains"})
	if err != nil {
		t.Fatalf("addresses supported-chains returned error: %v", err)
	}

	if got := stdout.String(); !strings.Contains(got, `"chain": "arbitrum"`) {
		t.Fatalf("unexpected addresses supported-chains output: %q", got)
	}
}

func TestTransactionsList(t *testing.T) {
	application := newTestApp(t)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	printer := output.New(stdout, stderr)
	ctx := context.Background()

	if err := runWithPrinter(ctx, printer, application, []string{"login", "--client-id", "client_test_1234", "--secret-key", "secret_test_12345678"}); err != nil {
		t.Fatalf("login returned error: %v", err)
	}

	stdout.Reset()
	err := runWithPrinter(ctx, printer, application, []string{"transactions", "list", "--page", "1", "--limit", "10", "--status", "success", "--type", "credit"})
	if err != nil {
		t.Fatalf("transactions list returned error: %v", err)
	}

	if got := stdout.String(); !strings.Contains(got, `"transactions"`) {
		t.Fatalf("unexpected transactions list output: %q", got)
	}
}

func TestTransactionsGet(t *testing.T) {
	application := newTestApp(t)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	printer := output.New(stdout, stderr)
	ctx := context.Background()

	if err := runWithPrinter(ctx, printer, application, []string{"login", "--client-id", "client_test_1234", "--secret-key", "secret_test_12345678"}); err != nil {
		t.Fatalf("login returned error: %v", err)
	}

	stdout.Reset()
	err := runWithPrinter(ctx, printer, application, []string{"transactions", "get", "txn_123"})
	if err != nil {
		t.Fatalf("transactions get returned error: %v", err)
	}

	if got := stdout.String(); !strings.Contains(got, `"transaction_id": "txn_123"`) {
		t.Fatalf("unexpected transactions get output: %q", got)
	}
}

func TestTransactionsListWithCursor(t *testing.T) {
	application := newTestApp(t)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	printer := output.New(stdout, stderr)
	ctx := context.Background()

	if err := runWithPrinter(ctx, printer, application, []string{"login", "--client-id", "client_test_1234", "--secret-key", "secret_test_12345678"}); err != nil {
		t.Fatalf("login returned error: %v", err)
	}

	stdout.Reset()
	err := runWithPrinter(ctx, printer, application, []string{"transactions", "list", "--cursor", "cursor_123"})
	if err != nil {
		t.Fatalf("transactions list with cursor returned error: %v", err)
	}

	if got := stdout.String(); !strings.Contains(got, `"next_cursor": "cursor_123"`) {
		t.Fatalf("unexpected transactions list with cursor output: %q", got)
	}
}

func TestTradingQuoteCreate(t *testing.T) {
	application := newTestApp(t)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	printer := output.New(stdout, stderr)
	ctx := context.Background()

	if err := runWithPrinter(ctx, printer, application, []string{"login", "--client-id", "client_test_1234", "--secret-key", "secret_test_12345678"}); err != nil {
		t.Fatalf("login returned error: %v", err)
	}

	stdout.Reset()
	err := runWithPrinter(ctx, printer, application, []string{"trading", "quotes", "create", "--base-currency", "BTC", "--quote-currency", "USDT", "--side", "buy", "--quantity", "0.0001"})
	if err != nil {
		t.Fatalf("trading quotes create returned error: %v", err)
	}

	if got := stdout.String(); !strings.Contains(got, `"id": "quote_123"`) {
		t.Fatalf("unexpected trading quote output: %q", got)
	}
}

func TestTradingOrderCreate(t *testing.T) {
	application := newTestApp(t)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	printer := output.New(stdout, stderr)
	ctx := context.Background()

	if err := runWithPrinter(ctx, printer, application, []string{"login", "--client-id", "client_test_1234", "--secret-key", "secret_test_12345678"}); err != nil {
		t.Fatalf("login returned error: %v", err)
	}

	stdout.Reset()
	err := runWithPrinter(ctx, printer, application, []string{"trading", "orders", "create", "--base-currency", "BTC", "--quote-currency", "USDT", "--side", "buy", "--quantity", "0.0001", "--price", "90822.540125", "--quote-id", "quote_123", "--reference", "324567890"})
	if err != nil {
		t.Fatalf("trading orders create returned error: %v", err)
	}

	if got := stdout.String(); !strings.Contains(got, `"id": "order_123"`) {
		t.Fatalf("unexpected trading order create output: %q", got)
	}
}

func TestTradingOrdersList(t *testing.T) {
	application := newTestApp(t)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	printer := output.New(stdout, stderr)
	ctx := context.Background()

	if err := runWithPrinter(ctx, printer, application, []string{"login", "--client-id", "client_test_1234", "--secret-key", "secret_test_12345678"}); err != nil {
		t.Fatalf("login returned error: %v", err)
	}

	stdout.Reset()
	err := runWithPrinter(ctx, printer, application, []string{"trading", "orders", "list"})
	if err != nil {
		t.Fatalf("trading orders list returned error: %v", err)
	}

	if got := stdout.String(); !strings.Contains(got, `"orders"`) {
		t.Fatalf("unexpected trading orders list output: %q", got)
	}
}

func TestTradingOrderGet(t *testing.T) {
	application := newTestApp(t)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	printer := output.New(stdout, stderr)
	ctx := context.Background()

	if err := runWithPrinter(ctx, printer, application, []string{"login", "--client-id", "client_test_1234", "--secret-key", "secret_test_12345678"}); err != nil {
		t.Fatalf("login returned error: %v", err)
	}

	stdout.Reset()
	err := runWithPrinter(ctx, printer, application, []string{"trading", "orders", "get", "order_123"})
	if err != nil {
		t.Fatalf("trading orders get returned error: %v", err)
	}

	if got := stdout.String(); !strings.Contains(got, `"id": "order_123"`) {
		t.Fatalf("unexpected trading order get output: %q", got)
	}
}

func TestTradingPrices(t *testing.T) {
	application := newTestApp(t)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	printer := output.New(stdout, stderr)
	ctx := context.Background()

	if err := runWithPrinter(ctx, printer, application, []string{"login", "--client-id", "client_test_1234", "--secret-key", "secret_test_12345678"}); err != nil {
		t.Fatalf("login returned error: %v", err)
	}

	stdout.Reset()
	err := runWithPrinter(ctx, printer, application, []string{"trading", "prices"})
	if err != nil {
		t.Fatalf("trading prices returned error: %v", err)
	}

	if got := stdout.String(); !strings.Contains(got, `"base_currency": "BTC"`) {
		t.Fatalf("unexpected trading prices output: %q", got)
	}
}

func TestTransfersCreate(t *testing.T) {
	application := newTestApp(t)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	printer := output.New(stdout, stderr)
	ctx := context.Background()

	if err := runWithPrinter(ctx, printer, application, []string{"login", "--client-id", "client_test_1234", "--secret-key", "secret_test_12345678"}); err != nil {
		t.Fatalf("login returned error: %v", err)
	}

	stdout.Reset()
	err := runWithPrinter(ctx, printer, application, []string{
		"transfers", "create",
		"--to-address", "0xdaBe4B0Ca57dfBF13763D5f190A2d30B94f1Bf59",
		"--amount", "2000000",
		"--currency", "USDT",
		"--chain", "bsc",
		"--reference", "testing_bsc_013",
		"--description", "Testing on USDT to binance",
	})
	if err != nil {
		t.Fatalf("transfers create returned error: %v", err)
	}

	if got := stdout.String(); !strings.Contains(got, `"transaction_id": "transfer_123"`) {
		t.Fatalf("unexpected transfers create output: %q", got)
	}
}

func TestWaitCardActive(t *testing.T) {
	application := newTestApp(t)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	printer := output.New(stdout, stderr)
	ctx := context.Background()

	if err := runWithPrinter(ctx, printer, application, []string{"login", "--client-id", "client_test_1234", "--secret-key", "secret_test_12345678"}); err != nil {
		t.Fatalf("login returned error: %v", err)
	}

	stdout.Reset()
	err := runWithPrinter(ctx, printer, application, []string{"wait", "card-active", "card_123", "--timeout", "5s", "--interval", "100ms"})
	if err != nil {
		t.Fatalf("wait card-active returned error: %v", err)
	}

	if got := stdout.String(); !strings.Contains(got, `"status": "active"`) {
		t.Fatalf("unexpected wait card-active output: %q", got)
	}
}

func TestPayoutsQuoteCreate(t *testing.T) {
	application := newTestApp(t)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	printer := output.New(stdout, stderr)
	ctx := context.Background()

	if err := runWithPrinter(ctx, printer, application, []string{"login", "--client-id", "client_test_1234", "--secret-key", "secret_test_12345678"}); err != nil {
		t.Fatalf("login returned error: %v", err)
	}

	stdout.Reset()
	err := runWithPrinter(ctx, printer, application, []string{
		"payouts", "quotes", "create",
		"--from-asset", "BTC",
		"--to-currency", "NGN",
		"--source", "onchain",
		"--chain", "BITCOIN",
		"--amount", "0.0015",
		"--payment-reason", "Vendor payout",
		"--reference", "offramp-quote-001",
		"--client-meta-data", `{"invoice_id":"INV-1001"}`,
		"--country", "NG",
	})
	if err != nil {
		t.Fatalf("payouts quote create returned error: %v", err)
	}

	if got := stdout.String(); !strings.Contains(got, `"quote_id": "QT_120732"`) {
		t.Fatalf("unexpected payouts quote create output: %q", got)
	}
}

func TestPayoutsInitialize(t *testing.T) {
	application := newTestApp(t)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	printer := output.New(stdout, stderr)
	ctx := context.Background()

	if err := runWithPrinter(ctx, printer, application, []string{"login", "--client-id", "client_test_1234", "--secret-key", "secret_test_12345678"}); err != nil {
		t.Fatalf("login returned error: %v", err)
	}

	stdout.Reset()
	err := runWithPrinter(ctx, printer, application, []string{
		"payouts", "initialize", "QT_6158",
		"--customer-id", "customer_123",
		"--beneficiary", `{"account_number":"0123456789","bank_code":"058","account_name":"Ada Okafor"}`,
		"--reference", "offramp-init-001",
		"--payment-reason", "Supplier settlement",
		"--client-meta-data", `{"batch":"march-payouts"}`,
		"--callback-url", "https://example.com/webhooks/payouts",
	})
	if err != nil {
		t.Fatalf("payouts initialize returned error: %v", err)
	}

	if got := stdout.String(); !strings.Contains(got, `"status": "initiated"`) {
		t.Fatalf("unexpected payouts initialize output: %q", got)
	}
}

func TestPayoutsInitializeWithBeneficiaryID(t *testing.T) {
	application := newTestApp(t)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	printer := output.New(stdout, stderr)
	ctx := context.Background()

	if err := runWithPrinter(ctx, printer, application, []string{"login", "--client-id", "client_test_1234", "--secret-key", "secret_test_12345678"}); err != nil {
		t.Fatalf("login returned error: %v", err)
	}

	stdout.Reset()
	err := runWithPrinter(ctx, printer, application, []string{
		"payouts", "initialize", "QT_6158",
		"--customer-id", "customer_123",
		"--beneficiary-id", "beneficiary_123",
		"--beneficiary-country", "NG",
		"--reference", "offramp-init-001",
		"--payment-reason", "Supplier settlement",
	})
	if err != nil {
		t.Fatalf("payouts initialize with beneficiary-id returned error: %v", err)
	}

	if got := stdout.String(); !strings.Contains(got, `"status": "initiated"`) {
		t.Fatalf("unexpected payouts initialize with beneficiary-id output: %q", got)
	}
}

func TestPayoutsBeneficiaryLookup(t *testing.T) {
	application := newTestApp(t)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	printer := output.New(stdout, stderr)
	ctx := context.Background()

	if err := runWithPrinter(ctx, printer, application, []string{"login", "--client-id", "client_test_1234", "--secret-key", "secret_test_12345678"}); err != nil {
		t.Fatalf("login returned error: %v", err)
	}

	stdout.Reset()
	err := runWithPrinter(ctx, printer, application, []string{
		"payouts", "beneficiary-lookup",
		"--country", "NG",
		"--account-number", "0123456789",
		"--bank-code", "058",
		"--type", "bank_account",
	})
	if err != nil {
		t.Fatalf("payouts beneficiary-lookup returned error: %v", err)
	}

	if got := stdout.String(); !strings.Contains(got, `"account_name": "ADA OKAFOR"`) {
		t.Fatalf("unexpected payouts beneficiary-lookup output: %q", got)
	}
}

func TestPayoutsLimits(t *testing.T) {
	application := newTestApp(t)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	printer := output.New(stdout, stderr)
	ctx := context.Background()

	if err := runWithPrinter(ctx, printer, application, []string{"login", "--client-id", "client_test_1234", "--secret-key", "secret_test_12345678"}); err != nil {
		t.Fatalf("login returned error: %v", err)
	}

	stdout.Reset()
	err := runWithPrinter(ctx, printer, application, []string{"payouts", "limits"})
	if err != nil {
		t.Fatalf("payouts limits returned error: %v", err)
	}

	if got := stdout.String(); !strings.Contains(got, `"country": "NG"`) {
		t.Fatalf("unexpected payouts limits output: %q", got)
	}
}

func TestAPIGet(t *testing.T) {
	application := newTestApp(t)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	printer := output.New(stdout, stderr)
	ctx := context.Background()

	if err := runWithPrinter(ctx, printer, application, []string{"login", "--client-id", "client_test_1234", "--secret-key", "secret_test_12345678"}); err != nil {
		t.Fatalf("login returned error: %v", err)
	}

	stdout.Reset()
	if err := runWithPrinter(ctx, printer, application, []string{"api", "get", "/api/balances"}); err != nil {
		t.Fatalf("api get returned error: %v", err)
	}

	if got := stdout.String(); !strings.Contains(got, `"company_id": "company_test"`) {
		t.Fatalf("unexpected api get output: %q", got)
	}
}

func TestAPIPost(t *testing.T) {
	application := newTestApp(t)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	printer := output.New(stdout, stderr)
	ctx := context.Background()

	if err := runWithPrinter(ctx, printer, application, []string{"login", "--client-id", "client_test_1234", "--secret-key", "secret_test_12345678"}); err != nil {
		t.Fatalf("login returned error: %v", err)
	}

	stdout.Reset()
	if err := runWithPrinter(ctx, printer, application, []string{"api", "post", "/api/customers", "--data", `{"email":"john@example.com"}`}); err != nil {
		t.Fatalf("api post returned error: %v", err)
	}

	if got := stdout.String(); !strings.Contains(got, `"success": true`) {
		t.Fatalf("unexpected api post output: %q", got)
	}
}

func TestCustomersCreate(t *testing.T) {
	application := newTestApp(t)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	printer := output.New(stdout, stderr)
	ctx := context.Background()

	if err := runWithPrinter(ctx, printer, application, []string{"login", "--client-id", "client_test_1234", "--secret-key", "secret_test_12345678"}); err != nil {
		t.Fatalf("login returned error: %v", err)
	}

	stdout.Reset()
	err := runWithPrinter(ctx, printer, application, []string{"customers", "create", "--customer-type", "individual", "--email", "john@example.com", "--first-name", "John", "--last-name", "Doe"})
	if err != nil {
		t.Fatalf("customers create returned error: %v", err)
	}

	if got := stdout.String(); !strings.Contains(got, `"email": "john@example.com"`) {
		t.Fatalf("unexpected customers create output: %q", got)
	}
}

func TestCustomersList(t *testing.T) {
	application := newTestApp(t)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	printer := output.New(stdout, stderr)
	ctx := context.Background()

	if err := runWithPrinter(ctx, printer, application, []string{"login", "--client-id", "client_test_1234", "--secret-key", "secret_test_12345678"}); err != nil {
		t.Fatalf("login returned error: %v", err)
	}

	stdout.Reset()
	err := runWithPrinter(ctx, printer, application, []string{"customers", "list", "--page", "1", "--limit", "20"})
	if err != nil {
		t.Fatalf("customers list returned error: %v", err)
	}

	if got := stdout.String(); !strings.Contains(got, `"customers"`) {
		t.Fatalf("unexpected customers list output: %q", got)
	}
}

func TestCustomersGet(t *testing.T) {
	application := newTestApp(t)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	printer := output.New(stdout, stderr)
	ctx := context.Background()

	if err := runWithPrinter(ctx, printer, application, []string{"login", "--client-id", "client_test_1234", "--secret-key", "secret_test_12345678"}); err != nil {
		t.Fatalf("login returned error: %v", err)
	}

	stdout.Reset()
	err := runWithPrinter(ctx, printer, application, []string{"customers", "get", "customer_123"})
	if err != nil {
		t.Fatalf("customers get returned error: %v", err)
	}

	if got := stdout.String(); !strings.Contains(got, `"id": "customer_123"`) {
		t.Fatalf("unexpected customers get output: %q", got)
	}
}

func TestCustomersUpdate(t *testing.T) {
	application := newTestApp(t)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	printer := output.New(stdout, stderr)
	ctx := context.Background()

	if err := runWithPrinter(ctx, printer, application, []string{"login", "--client-id", "client_test_1234", "--secret-key", "secret_test_12345678"}); err != nil {
		t.Fatalf("login returned error: %v", err)
	}

	stdout.Reset()
	err := runWithPrinter(ctx, printer, application, []string{"customers", "update", "customer_123", "--first-name", "John", "--last-name", "Doe", "--phone", "08123456789", "--email", "john@example.com"})
	if err != nil {
		t.Fatalf("customers update returned error: %v", err)
	}

	if got := stdout.String(); !strings.Contains(got, `"first_name": "John"`) {
		t.Fatalf("unexpected customers update output: %q", got)
	}
}

func TestBeneficiariesCreate(t *testing.T) {
	application := newTestApp(t)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	printer := output.New(stdout, stderr)
	ctx := context.Background()

	if err := runWithPrinter(ctx, printer, application, []string{"login", "--client-id", "client_test_1234", "--secret-key", "secret_test_12345678"}); err != nil {
		t.Fatalf("login returned error: %v", err)
	}

	stdout.Reset()
	err := runWithPrinter(ctx, printer, application, []string{
		"beneficiaries", "create",
		"--company-id", "company_test",
		"--type", "bank",
		"--name", "Bank Beneficiary",
		"--account-name", "John Doe",
		"--account-number", "1234567890",
		"--bank-code", "058",
	})
	if err != nil {
		t.Fatalf("beneficiaries create returned error: %v", err)
	}

	if got := stdout.String(); !strings.Contains(got, `"id": "beneficiary_123"`) {
		t.Fatalf("unexpected beneficiaries create output: %q", got)
	}
}

func TestBeneficiariesList(t *testing.T) {
	application := newTestApp(t)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	printer := output.New(stdout, stderr)
	ctx := context.Background()

	if err := runWithPrinter(ctx, printer, application, []string{"login", "--client-id", "client_test_1234", "--secret-key", "secret_test_12345678"}); err != nil {
		t.Fatalf("login returned error: %v", err)
	}

	stdout.Reset()
	err := runWithPrinter(ctx, printer, application, []string{"beneficiaries", "list", "--limit", "10", "--offset", "0"})
	if err != nil {
		t.Fatalf("beneficiaries list returned error: %v", err)
	}

	if got := stdout.String(); !strings.Contains(got, `"beneficiaries"`) {
		t.Fatalf("unexpected beneficiaries list output: %q", got)
	}
}

func TestBeneficiariesGet(t *testing.T) {
	application := newTestApp(t)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	printer := output.New(stdout, stderr)
	ctx := context.Background()

	if err := runWithPrinter(ctx, printer, application, []string{"login", "--client-id", "client_test_1234", "--secret-key", "secret_test_12345678"}); err != nil {
		t.Fatalf("login returned error: %v", err)
	}

	stdout.Reset()
	err := runWithPrinter(ctx, printer, application, []string{"beneficiaries", "get", "beneficiary_123"})
	if err != nil {
		t.Fatalf("beneficiaries get returned error: %v", err)
	}

	if got := stdout.String(); !strings.Contains(got, `"name": "Bank Beneficiary"`) {
		t.Fatalf("unexpected beneficiaries get output: %q", got)
	}
}

func TestBeneficiariesUpdate(t *testing.T) {
	application := newTestApp(t)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	printer := output.New(stdout, stderr)
	ctx := context.Background()

	if err := runWithPrinter(ctx, printer, application, []string{"login", "--client-id", "client_test_1234", "--secret-key", "secret_test_12345678"}); err != nil {
		t.Fatalf("login returned error: %v", err)
	}

	stdout.Reset()
	err := runWithPrinter(ctx, printer, application, []string{
		"beneficiaries", "update", "beneficiary_123",
		"--name", "Access Bank GH",
		"--account-name", "Champ Lan",
		"--account-number", "1234567890",
		"--bank-code", "011",
		"--blacklisted=true",
	})
	if err != nil {
		t.Fatalf("beneficiaries update returned error: %v", err)
	}

	if got := stdout.String(); !strings.Contains(got, `"is_blacklisted": true`) {
		t.Fatalf("unexpected beneficiaries update output: %q", got)
	}
}

func TestBeneficiariesDelete(t *testing.T) {
	application := newTestApp(t)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	printer := output.New(stdout, stderr)
	ctx := context.Background()

	if err := runWithPrinter(ctx, printer, application, []string{"login", "--client-id", "client_test_1234", "--secret-key", "secret_test_12345678"}); err != nil {
		t.Fatalf("login returned error: %v", err)
	}

	stdout.Reset()
	err := runWithPrinter(ctx, printer, application, []string{"beneficiaries", "delete", "beneficiary_123"})
	if err != nil {
		t.Fatalf("beneficiaries delete returned error: %v", err)
	}

	if got := stdout.String(); !strings.Contains(got, `"success": true`) {
		t.Fatalf("unexpected beneficiaries delete output: %q", got)
	}
}

func TestCardsCreate(t *testing.T) {
	application := newTestApp(t)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	printer := output.New(stdout, stderr)
	ctx := context.Background()

	if err := runWithPrinter(ctx, printer, application, []string{"login", "--client-id", "client_test_1234", "--secret-key", "secret_test_12345678"}); err != nil {
		t.Fatalf("login returned error: %v", err)
	}

	stdout.Reset()
	err := runWithPrinter(ctx, printer, application, []string{
		"cards", "create",
		"--customer-id", "customer_123",
		"--card-type", "virtual",
		"--card-brand", "visa",
		"--currency", "USD",
		"--billing-line1", "123 Main Street",
		"--billing-city", "Lagos",
		"--billing-state", "Lagos",
		"--billing-postal-code", "100001",
		"--billing-country", "NG",
		"--single-transaction", "1000.00",
		"--daily", "5000.00",
		"--metadata", `{"department":"Engineering"}`,
	})
	if err != nil {
		t.Fatalf("cards create returned error: %v", err)
	}

	if got := stdout.String(); !strings.Contains(got, `"id": "card_123"`) {
		t.Fatalf("unexpected cards create output: %q", got)
	}
}

func TestCardsDetails(t *testing.T) {
	application := newTestApp(t)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	printer := output.New(stdout, stderr)
	ctx := context.Background()

	if err := runWithPrinter(ctx, printer, application, []string{"login", "--client-id", "client_test_1234", "--secret-key", "secret_test_12345678"}); err != nil {
		t.Fatalf("login returned error: %v", err)
	}

	stdout.Reset()
	err := runWithPrinter(ctx, printer, application, []string{"cards", "details", "card_123"})
	if err != nil {
		t.Fatalf("cards details returned error: %v", err)
	}

	if got := stdout.String(); !strings.Contains(got, `"card_number": "4532123456784532"`) {
		t.Fatalf("unexpected cards details output: %q", got)
	}
}

func TestCardsFund(t *testing.T) {
	application := newTestApp(t)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	printer := output.New(stdout, stderr)
	ctx := context.Background()

	if err := runWithPrinter(ctx, printer, application, []string{"login", "--client-id", "client_test_1234", "--secret-key", "secret_test_12345678"}); err != nil {
		t.Fatalf("login returned error: %v", err)
	}

	stdout.Reset()
	err := runWithPrinter(ctx, printer, application, []string{"cards", "fund", "card_123", "--amount", "500.00", "--currency", "USD", "--reference", "fund-001", "--description", "Monthly allowance"})
	if err != nil {
		t.Fatalf("cards fund returned error: %v", err)
	}

	if got := stdout.String(); !strings.Contains(got, `"operation": "fund"`) {
		t.Fatalf("unexpected cards fund output: %q", got)
	}
}

func TestCardsSpendingLimits(t *testing.T) {
	application := newTestApp(t)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	printer := output.New(stdout, stderr)
	ctx := context.Background()

	if err := runWithPrinter(ctx, printer, application, []string{"login", "--client-id", "client_test_1234", "--secret-key", "secret_test_12345678"}); err != nil {
		t.Fatalf("login returned error: %v", err)
	}

	stdout.Reset()
	err := runWithPrinter(ctx, printer, application, []string{
		"cards", "spending-limits", "card_123",
		"--single-transaction", "2000.00",
		"--daily", "10000.00",
		"--allowed-categories", "travel,software",
		"--blocked-merchants", "merchant_123",
	})
	if err != nil {
		t.Fatalf("cards spending-limits returned error: %v", err)
	}

	if got := stdout.String(); !strings.Contains(got, `"allowed_categories": [`) {
		t.Fatalf("unexpected cards spending-limits output: %q", got)
	}
}

func TestLoginToNamedProfileSetsActiveProfile(t *testing.T) {
	application := newTestApp(t)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	printer := output.New(stdout, stderr)
	ctx := context.Background()

	if err := runWithPrinter(ctx, printer, application, []string{"login", "--profile", "live", "--client-id", "live_client", "--secret-key", "live_secret"}); err != nil {
		t.Fatalf("login returned error: %v", err)
	}

	stdout.Reset()
	if err := runWithPrinter(ctx, printer, application, []string{"config", "get", "profile.active"}); err != nil {
		t.Fatalf("config get returned error: %v", err)
	}

	if got := stdout.String(); !strings.Contains(got, `"value": "live"`) {
		t.Fatalf("unexpected active profile after login: %q", got)
	}
}

func TestLoginDoesNotPersistCredentialsWhenWhoAmIFails(t *testing.T) {
	base := t.TempDir()
	application := newTestAppWithTransport(t, base, roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.Path == "/api/whoami" && req.Method == http.MethodGet {
			return &http.Response{
				StatusCode: http.StatusUnauthorized,
				Status:     "401 Unauthorized",
				Body:       io.NopCloser(strings.NewReader(`{"error":"invalid credentials"}`)),
				Header:     make(http.Header),
				Request:    req,
			}, nil
		}

		return &http.Response{
			StatusCode: http.StatusNotFound,
			Status:     "404 Not Found",
			Body:       io.NopCloser(strings.NewReader(`{"error":"not found"}`)),
			Header:     make(http.Header),
			Request:    req,
		}, nil
	}))
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	printer := output.New(stdout, stderr)
	ctx := context.Background()

	err := runWithPrinter(ctx, printer, application, []string{"login", "--client-id", "bad_client", "--secret-key", "bad_secret"})
	if err == nil {
		t.Fatal("expected login to fail")
	}

	stdout.Reset()
	if err := runWithPrinter(ctx, printer, application, []string{"config", "get", "profile.active"}); err != nil {
		t.Fatalf("config get returned error: %v", err)
	}

	if got := stdout.String(); !strings.Contains(got, `"value": "default"`) {
		t.Fatalf("unexpected active profile after failed login: %q", got)
	}
}

func TestInteractiveLoginPromptsForMissingValues(t *testing.T) {
	application := newTestApp(t)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	printer := output.New(stdout, stderr)
	ctx := context.Background()

	originalPromptLine := promptLine
	originalPromptSecret := promptSecret
	t.Cleanup(func() {
		promptLine = originalPromptLine
		promptSecret = originalPromptSecret
	})

	promptLine = func(label string) (string, error) {
		if label != "Client ID" {
			t.Fatalf("unexpected prompt label: %s", label)
		}
		return "interactive_client", nil
	}
	promptSecret = func(label string) (string, error) {
		if label != "Secret Key" {
			t.Fatalf("unexpected prompt label: %s", label)
		}
		return "interactive_secret", nil
	}

	if err := runWithPrinter(ctx, printer, application, []string{"login", "--interactive"}); err != nil {
		t.Fatalf("interactive login returned error: %v", err)
	}

	stdout.Reset()
	if err := runWithPrinter(ctx, printer, application, []string{"whoami"}); err != nil {
		t.Fatalf("whoami returned error after interactive login: %v", err)
	}

	if got := stdout.String(); !strings.Contains(got, `"client_id": "interactive_client"`) {
		t.Fatalf("unexpected whoami output after interactive login: %q", got)
	}
}

func TestInteractiveLoginCanUseProvidedClientID(t *testing.T) {
	application := newTestApp(t)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	printer := output.New(stdout, stderr)
	ctx := context.Background()

	originalPromptLine := promptLine
	originalPromptSecret := promptSecret
	t.Cleanup(func() {
		promptLine = originalPromptLine
		promptSecret = originalPromptSecret
	})

	promptLine = func(label string) (string, error) {
		t.Fatalf("client id prompt should not be called when --client-id is provided")
		return "", nil
	}
	promptSecret = func(label string) (string, error) {
		if label != "Secret Key" {
			t.Fatalf("unexpected prompt label: %s", label)
		}
		return "interactive_secret", nil
	}

	if err := runWithPrinter(ctx, printer, application, []string{"login", "--interactive", "--client-id", "provided_client"}); err != nil {
		t.Fatalf("interactive login with provided client id returned error: %v", err)
	}

	stdout.Reset()
	if err := runWithPrinter(ctx, printer, application, []string{"whoami"}); err != nil {
		t.Fatalf("whoami returned error after partial interactive login: %v", err)
	}

	if got := stdout.String(); !strings.Contains(got, `"client_id": "provided_client"`) {
		t.Fatalf("unexpected whoami output after partial interactive login: %q", got)
	}
}

func TestConfigEdit(t *testing.T) {
	application := newTestApp(t)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	t.Setenv("EDITOR", "")

	originalRunEditor := runEditor
	t.Cleanup(func() {
		runEditor = originalRunEditor
	})

	called := false
	runEditor = func(editor string, path string) error {
		called = true
		if editor != "vi" {
			t.Fatalf("unexpected editor: %s", editor)
		}
		if path != application.ConfigStore.Path() {
			t.Fatalf("unexpected path: %s", path)
		}
		return nil
	}

	err := runWithPrinter(context.Background(), output.New(stdout, stderr), application, []string{"config", "edit"})
	if err != nil {
		t.Fatalf("config edit returned error: %v", err)
	}
	if !called {
		t.Fatalf("expected editor to be invoked")
	}
	if got := stdout.String(); !strings.Contains(got, `"message": "Configuration file opened"`) {
		t.Fatalf("unexpected config edit output: %q", got)
	}
}

func TestListenStartup(t *testing.T) {
	application := newTestApp(t)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	originalRunWebhookServer := runWebhookServer
	t.Cleanup(func() {
		runWebhookServer = originalRunWebhookServer
	})
	runWebhookServer = func(ctx context.Context, server *http.Server) error {
		return nil
	}

	err := runWithPrinter(context.Background(), output.New(stdout, stderr), application, []string{"listen", "--secret", "webhook_secret_test"})
	if err != nil {
		t.Fatalf("listen returned error: %v", err)
	}

	if got := stdout.String(); !strings.Contains(got, `"status": "listening"`) {
		t.Fatalf("unexpected listen output: %q", got)
	}
}

func newTestApp(t *testing.T) *app.App {
	t.Helper()

	base := t.TempDir()
	return newTestAppWithTransport(t, base, roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.Header.Get("x-auth-client") == "" || req.Header.Get("x-auth-signature") == "" {
			return &http.Response{
				StatusCode: http.StatusUnauthorized,
				Status:     "401 Unauthorized",
				Body:       io.NopCloser(strings.NewReader(`{"error":"missing auth"}`)),
				Header:     make(http.Header),
				Request:    req,
			}, nil
		}
		if req.URL.Path == "/api/whoami" && req.Method == http.MethodGet {
			clientID := req.Header.Get("x-auth-client")
			return &http.Response{
				StatusCode: http.StatusOK,
				Status:     "200 OK",
				Body: io.NopCloser(strings.NewReader(`{
					"authenticated": true,
					"auth_method": "hmac",
					"client_id": "` + clientID + `",
					"client_name": "Bitnob CLI Test",
					"active_company_id": "company_test",
					"environment": "sandbox",
					"permissions": ["balances.read", "customers.write"],
					"active": true,
					"metadata": {},
					"rate_limit": {
						"requests_per_minute": 60,
						"requests_per_hour": 1000,
						"burst_size": 10
					},
					"timestamp": "2026-03-23T10:00:00Z"
				}`)),
				Header:  make(http.Header),
				Request: req,
			}, nil
		}
		if req.URL.Path == "/api/customers" && req.Method == http.MethodPost {
			body, err := io.ReadAll(req.Body)
			if err != nil {
				return nil, err
			}
			if strings.Contains(string(body), `"firstName"`) || strings.Contains(string(body), `"lastName"`) || strings.Contains(string(body), `"countryCode"`) {
				return &http.Response{
					StatusCode: http.StatusBadRequest,
					Status:     "400 Bad Request",
					Body:       io.NopCloser(strings.NewReader(`{"error":"expected snake_case customer payload"}`)),
					Header:     make(http.Header),
					Request:    req,
				}, nil
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Status:     "200 OK",
				Body: io.NopCloser(strings.NewReader(`{
					"success": true,
					"message": "Customer created successfully",
					"data": {
						"id": "customer_123",
						"email": "john@example.com",
						"metadata": "{}",
						"is_active": true,
						"customer_type": "individual",
						"first_name": "John",
						"last_name": "Doe",
						"phone": "",
						"country_code": "",
						"blacklist": false,
						"company_id": "company_test",
						"created_by": "client_test_1234",
						"created_at": "2025-06-19T12:30:00Z",
						"updated_at": "2025-06-19T13:45:00Z"
					},
					"metadata": {
						"request_id": "customer_req_test"
					},
					"timestamp": "2026-03-21T19:03:58.39826262Z"
				}`)),
				Header:  make(http.Header),
				Request: req,
			}, nil
		}
		if req.URL.Path == "/api/customers" && req.Method == http.MethodGet {
			return &http.Response{
				StatusCode: http.StatusOK,
				Status:     "200 OK",
				Body: io.NopCloser(strings.NewReader(`{
					"status": true,
					"message": "successfully fetched all customers",
					"data": {
						"customers": [
							{
								"id": "customer_123",
								"created_at": "2025-09-04T14:40:56.621Z",
								"updated_at": "2025-09-04T14:40:56.621Z",
								"first_name": "John",
								"last_name": "Doe",
								"email": "john@example.com",
								"phone": "08123456789",
								"country_code": "+234",
								"blacklist": false
							}
						],
						"meta": {
							"page": 1,
							"take": 20,
							"item_count": 1,
							"page_count": 1,
							"has_previous_page": false,
							"has_next_page": false
						}
					}
				}`)),
				Header:  make(http.Header),
				Request: req,
			}, nil
		}
		if req.URL.Path == "/api/customers/customer_123" && req.Method == http.MethodGet {
			return &http.Response{
				StatusCode: http.StatusOK,
				Status:     "200 OK",
				Body: io.NopCloser(strings.NewReader(`{
					"status": true,
					"message": "successfully fetched customer",
					"data": {
						"id": "customer_123",
						"created_at": "2025-09-04T14:40:56.621Z",
						"updated_at": "2025-09-04T14:40:56.621Z",
						"first_name": "John",
						"last_name": "Doe",
						"email": "john@example.com",
						"phone": "08123456789",
						"country_code": "+234",
						"blacklist": false
					}
				}`)),
				Header:  make(http.Header),
				Request: req,
			}, nil
		}
		if req.URL.Path == "/api/customers/customer_123" && req.Method == http.MethodPut {
			body, err := io.ReadAll(req.Body)
			if err != nil {
				return nil, err
			}
			if strings.Contains(string(body), `"firstName"`) || strings.Contains(string(body), `"lastName"`) {
				return &http.Response{
					StatusCode: http.StatusBadRequest,
					Status:     "400 Bad Request",
					Body:       io.NopCloser(strings.NewReader(`{"error":"expected snake_case customer payload"}`)),
					Header:     make(http.Header),
					Request:    req,
				}, nil
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Status:     "200 OK",
				Body: io.NopCloser(strings.NewReader(`{
					"status": true,
					"message": "Customer updated successfully",
					"data": {
						"id": "customer_123",
						"created_at": "2025-09-04T14:40:56.621Z",
						"updated_at": "2025-09-04T14:40:56.621Z",
						"customer_type": "individual",
						"first_name": "John",
						"last_name": "Doe",
						"email": "john@example.com",
						"phone": "08123456789",
						"country_code": "+234",
						"blacklist": false
					}
				}`)),
				Header:  make(http.Header),
				Request: req,
			}, nil
		}
		if req.URL.Path == "/api/beneficiaries" && req.Method == http.MethodPost {
			body, err := io.ReadAll(req.Body)
			if err != nil {
				return nil, err
			}
			if !strings.Contains(string(body), `"company_id":"company_test"`) || !strings.Contains(string(body), `"destination":"{\"account_name\":\"John Doe\",\"account_number\":\"1234567890\",\"bank_code\":\"058\"}"`) {
				return &http.Response{
					StatusCode: http.StatusBadRequest,
					Status:     "400 Bad Request",
					Body:       io.NopCloser(strings.NewReader(`{"error":"unexpected beneficiary payload"}`)),
					Header:     make(http.Header),
					Request:    req,
				}, nil
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Status:     "200 OK",
				Body: io.NopCloser(strings.NewReader(`{
					"success": true,
					"message": "Beneficiary created successfully",
					"data": {
						"id": "beneficiary_123",
						"company_id": "company_test",
						"type": "bank",
						"name": "Bank Beneficiary",
						"destination": {
							"account_name": "John Doe",
							"account_number": "1234567890",
							"bank_code": "058"
						},
						"is_blacklisted": false,
						"is_active": true,
						"created_by": "client_test_1234",
						"created_at": "2026-03-22T10:00:00Z",
						"updated_at": "2026-03-22T10:00:00Z"
					},
					"metadata": {
						"request_id": "beneficiary_req_test"
					},
					"timestamp": "2026-03-22T10:00:01Z"
				}`)),
				Header:  make(http.Header),
				Request: req,
			}, nil
		}
		if req.URL.Path == "/api/beneficiaries" && req.Method == http.MethodGet {
			return &http.Response{
				StatusCode: http.StatusOK,
				Status:     "200 OK",
				Body: io.NopCloser(strings.NewReader(`{
					"success": true,
					"message": "Beneficiaries retrieved successfully",
					"data": {
						"beneficiaries": [
							{
								"id": "beneficiary_123",
								"company_id": "company_test",
								"type": "bank",
								"name": "Bank Beneficiary",
								"destination": {
									"account_name": "John Doe",
									"account_number": "1234567890",
									"bank_code": "058"
								},
								"is_blacklisted": false,
								"is_active": true,
								"created_by": "client_test_1234",
								"created_at": "2026-03-22T10:00:00Z",
								"updated_at": "2026-03-22T10:00:00Z"
							}
						],
						"total": 1,
						"limit": 10,
						"offset": 0
					},
					"metadata": {
						"request_id": "beneficiary_req_test"
					},
					"timestamp": "2026-03-22T10:00:01Z"
				}`)),
				Header:  make(http.Header),
				Request: req,
			}, nil
		}
		if req.URL.Path == "/api/beneficiaries/beneficiary_123" && req.Method == http.MethodGet {
			return &http.Response{
				StatusCode: http.StatusOK,
				Status:     "200 OK",
				Body: io.NopCloser(strings.NewReader(`{
					"success": true,
					"message": "Beneficiary retrieved successfully",
					"data": {
						"id": "beneficiary_123",
						"company_id": "company_test",
						"type": "bank",
						"name": "Bank Beneficiary",
						"destination": {
							"account_name": "John Doe",
							"account_number": "1234567890",
							"bank_code": "058"
						},
						"is_blacklisted": false,
						"is_active": true,
						"created_by": "client_test_1234",
						"created_at": "2026-03-22T10:00:00Z",
						"updated_at": "2026-03-22T10:00:00Z"
					},
					"metadata": {
						"request_id": "beneficiary_req_test"
					},
					"timestamp": "2026-03-22T10:00:01Z"
				}`)),
				Header:  make(http.Header),
				Request: req,
			}, nil
		}
		if req.URL.Path == "/api/beneficiaries/beneficiary_123" && req.Method == http.MethodPut {
			body, err := io.ReadAll(req.Body)
			if err != nil {
				return nil, err
			}
			if !strings.Contains(string(body), `"name":"Access Bank GH"`) ||
				!strings.Contains(string(body), `"is_blacklisted":true`) ||
				!strings.Contains(string(body), `"destination":"{\"account_name\":\"Champ Lan\",\"account_number\":\"1234567890\",\"bank_code\":\"011\"}"`) {
				return &http.Response{
					StatusCode: http.StatusBadRequest,
					Status:     "400 Bad Request",
					Body:       io.NopCloser(strings.NewReader(`{"error":"unexpected beneficiary payload"}`)),
					Header:     make(http.Header),
					Request:    req,
				}, nil
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Status:     "200 OK",
				Body: io.NopCloser(strings.NewReader(`{
					"success": true,
					"message": "Beneficiary updated successfully",
					"data": {
						"id": "beneficiary_123",
						"company_id": "company_test",
						"type": "bank",
						"name": "Access Bank GH",
						"destination": {
							"account_name": "Champ Lan",
							"account_number": "1234567890",
							"bank_code": "011"
						},
						"is_blacklisted": true,
						"is_active": true,
						"created_by": "client_test_1234",
						"created_at": "2026-03-22T10:00:00Z",
						"updated_at": "2026-03-22T10:05:00Z"
					},
					"metadata": {
						"request_id": "beneficiary_req_test"
					},
					"timestamp": "2026-03-22T10:05:01Z"
				}`)),
				Header:  make(http.Header),
				Request: req,
			}, nil
		}
		if req.URL.Path == "/api/beneficiaries/beneficiary_123" && req.Method == http.MethodDelete {
			return &http.Response{
				StatusCode: http.StatusOK,
				Status:     "200 OK",
				Body: io.NopCloser(strings.NewReader(`{
					"success": true,
					"message": "Beneficiary deleted successfully",
					"data": {
						"success": true
					},
					"metadata": {
						"request_id": "beneficiary_req_test"
					},
					"timestamp": "2026-03-22T10:06:00Z"
				}`)),
				Header:  make(http.Header),
				Request: req,
			}, nil
		}
		if req.URL.Path == "/api/cards" && req.Method == http.MethodPost {
			body, err := io.ReadAll(req.Body)
			if err != nil {
				return nil, err
			}
			if !strings.Contains(string(body), `"customer_id":"customer_123"`) ||
				!strings.Contains(string(body), `"name":"John Doe"`) ||
				!strings.Contains(string(body), `"billing_address":{"line1":"123 Main Street","city":"Lagos","state":"Lagos","postal_code":"100001","country":"NG"}`) ||
				!strings.Contains(string(body), `"spending_limits":{"single_transaction":"1000.00","daily":"5000.00"}`) ||
				!strings.Contains(string(body), `"metadata":{"department":"Engineering"}`) {
				return &http.Response{
					StatusCode: http.StatusBadRequest,
					Status:     "400 Bad Request",
					Body:       io.NopCloser(strings.NewReader(`{"error":"unexpected card payload"}`)),
					Header:     make(http.Header),
					Request:    req,
				}, nil
			}
			return &http.Response{
				StatusCode: http.StatusCreated,
				Status:     "201 Created",
				Body: io.NopCloser(strings.NewReader(`{
					"status": "success",
					"message": "Card created successfully",
					"data": {
						"card": {
							"id": "card_123",
							"company_id": "company_test",
							"customer_id": "customer_123",
							"name": "John Doe",
							"card_type": "virtual",
							"card_brand": "visa",
							"last_four": "4532",
							"currency": "USD",
							"status": "active",
							"balance": "0.00",
							"spending_limits": {
								"single_transaction": "1000.00",
								"daily": "5000.00"
							},
							"billing_address": {
								"line1": "123 Main Street",
								"city": "Lagos",
								"state": "Lagos",
								"postal_code": "100001",
								"country": "NG"
							},
							"metadata": {
								"department": "Engineering"
							},
							"created_at": "2024-01-15T10:30:00Z",
							"updated_at": "2024-01-15T10:30:00Z"
						}
					}
				}`)),
				Header:  make(http.Header),
				Request: req,
			}, nil
		}
		if req.URL.Path == "/api/cards" && req.Method == http.MethodGet {
			return &http.Response{
				StatusCode: http.StatusOK,
				Status:     "200 OK",
				Body: io.NopCloser(strings.NewReader(`{
					"status": "success",
					"message": "Cards retrieved successfully",
					"data": {
						"cards": [
							{
								"id": "card_123",
								"customer_id": "customer_123",
								"name": "John Doe",
								"card_type": "virtual",
								"card_brand": "visa",
								"last_four": "4532",
								"currency": "USD",
								"status": "active",
								"balance": "500.00",
								"created_at": "2024-01-15T10:30:00Z"
							}
						],
						"total": 1,
						"limit": 20,
						"offset": 0
					}
				}`)),
				Header:  make(http.Header),
				Request: req,
			}, nil
		}
		if req.URL.Path == "/api/cards/card_123" && req.Method == http.MethodGet {
			return &http.Response{
				StatusCode: http.StatusOK,
				Status:     "200 OK",
				Body: io.NopCloser(strings.NewReader(`{
					"status": "success",
					"message": "Card retrieved successfully",
					"data": {
						"card": {
							"id": "card_123",
							"company_id": "company_test",
							"customer_id": "customer_123",
							"name": "John Doe",
							"card_type": "virtual",
							"card_brand": "visa",
							"last_four": "4532",
							"currency": "USD",
							"status": "active",
							"balance": "500.00",
							"current_spending": {
								"daily_spent": "150.00",
								"weekly_spent": "450.00",
								"monthly_spent": "1200.00"
							},
							"billing_address": {
								"line1": "123 Main Street",
								"city": "Lagos",
								"state": "Lagos",
								"postal_code": "100001",
								"country": "NG"
							},
							"created_at": "2024-01-15T10:30:00Z",
							"updated_at": "2024-01-20T14:25:00Z"
						}
					}
				}`)),
				Header:  make(http.Header),
				Request: req,
			}, nil
		}
		if req.URL.Path == "/api/cards/card_123/details" && req.Method == http.MethodGet {
			return &http.Response{
				StatusCode: http.StatusOK,
				Status:     "200 OK",
				Body: io.NopCloser(strings.NewReader(`{
					"status": "success",
					"message": "Card details retrieved successfully",
					"data": {
						"card_id": "card_123",
						"card_number": "4532123456784532",
						"cvv": "123",
						"expiry_month": "12",
						"expiry_year": "2027",
						"cardholder_name": "JOHN DOE",
						"billing_address": {
							"line1": "123 Main Street",
							"line2": "Suite 100",
							"city": "Lagos",
							"state": "Lagos",
							"postal_code": "100001",
							"country": "NG"
						}
					}
				}`)),
				Header:  make(http.Header),
				Request: req,
			}, nil
		}
		if req.URL.Path == "/api/cards/card_123/fund" && req.Method == http.MethodPost {
			body, err := io.ReadAll(req.Body)
			if err != nil {
				return nil, err
			}
			if !strings.Contains(string(body), `"amount":500000000`) || !strings.Contains(string(body), `"reference":"fund-001"`) {
				return &http.Response{
					StatusCode: http.StatusBadRequest,
					Status:     "400 Bad Request",
					Body:       io.NopCloser(strings.NewReader(`{"error":"unexpected card fund payload"}`)),
					Header:     make(http.Header),
					Request:    req,
				}, nil
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Status:     "200 OK",
				Body: io.NopCloser(strings.NewReader(`{
					"status": "success",
					"message": "Card funded successfully",
					"data": {
						"card_id": "card_123",
						"operation": "fund",
						"amount": "500.00",
						"currency": "USD",
						"new_balance": "1000.00",
						"transaction_id": "txn_card_fund_123",
						"reference": "fund-001",
						"created_at": "2024-01-15T11:00:00Z"
					}
				}`)),
				Header:  make(http.Header),
				Request: req,
			}, nil
		}
		if req.URL.Path == "/api/cards/card_123/spending-limits" && req.Method == http.MethodPut {
			body, err := io.ReadAll(req.Body)
			if err != nil {
				return nil, err
			}
			if !strings.Contains(string(body), `"allowed_categories":["travel","software"]`) || !strings.Contains(string(body), `"blocked_merchants":["merchant_123"]`) {
				return &http.Response{
					StatusCode: http.StatusBadRequest,
					Status:     "400 Bad Request",
					Body:       io.NopCloser(strings.NewReader(`{"error":"unexpected card spending limits payload"}`)),
					Header:     make(http.Header),
					Request:    req,
				}, nil
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Status:     "200 OK",
				Body: io.NopCloser(strings.NewReader(`{
					"status": "success",
					"message": "Spending limits updated successfully",
					"data": {
						"card_id": "card_123",
						"operation": "update_spending_limits",
						"spending_limits": {
							"single_transaction": "2000.00",
							"daily": "10000.00",
							"allowed_categories": ["travel", "software"],
							"blocked_merchants": ["merchant_123"]
						},
						"updated_at": "2024-01-15T15:00:00Z"
					}
				}`)),
				Header:  make(http.Header),
				Request: req,
			}, nil
		}
		if req.URL.Path == "/api/customers/customer_123/cards" && req.Method == http.MethodGet {
			return &http.Response{
				StatusCode: http.StatusOK,
				Status:     "200 OK",
				Body: io.NopCloser(strings.NewReader(`{
					"status": "success",
					"message": "Customer cards retrieved successfully",
					"data": {
						"cards": [
							{
								"id": "card_123",
								"customer_id": "customer_123",
								"name": "John Doe",
								"card_type": "virtual",
								"card_brand": "visa",
								"last_four": "4532",
								"currency": "USD",
								"status": "active",
								"balance": "500.00",
								"created_at": "2024-01-15T10:30:00Z"
							}
						],
						"total": 1
					}
				}`)),
				Header:  make(http.Header),
				Request: req,
			}, nil
		}
		if req.URL.Path == "/api/wallets" && req.Method == http.MethodGet {
			return &http.Response{
				StatusCode: http.StatusOK,
				Status:     "200 OK",
				Body: io.NopCloser(strings.NewReader(`{
					"success": true,
					"message": "Wallets retrieved successfully",
					"data": {
						"company_id": "company_test",
						"accounts": [
							{
								"account_id": "acct_1",
								"account_number": "1234567890",
								"currency": "USDT",
								"ledger_balance": "16047521",
								"available_balance": "16047521",
								"ledger_balance_formatted": "16.047521 USDT",
								"available_balance_formatted": "16.047521 USDT",
								"created_at": "2026-01-03T11:35:07.767513+00:00"
							}
						]
					},
					"metadata": {
						"request_id": "wallet_req_test"
					},
					"timestamp": "2026-03-29T10:00:00Z"
				}`)),
				Header:  make(http.Header),
				Request: req,
			}, nil
		}
		if req.URL.Path == "/api/balances/USDT" && req.Method == http.MethodGet {
			return &http.Response{
				StatusCode: http.StatusOK,
				Status:     "200 OK",
				Body: io.NopCloser(strings.NewReader(`{
					"success": true,
					"message": "Account balance retrieved successfully",
					"data": {
						"account_number": "1234567890",
						"currency": "USDT",
						"ledger_balance": "16047521",
						"available_balance": "16047521",
						"ledger_balance_formatted": "16.047521 USDT",
						"available_balance_formatted": "16.047521 USDT",
						"as_of": "2026-01-03T11:35:07.767513+00:00",
						"company_id": "company_test"
					},
					"metadata": {
						"request_id": "balance_req_test"
					},
					"timestamp": "2026-03-29T10:00:01Z"
				}`)),
				Header:  make(http.Header),
				Request: req,
			}, nil
		}
		if req.URL.Path == "/api/lightning/invoices" && req.Method == http.MethodPost {
			return &http.Response{
				StatusCode: http.StatusCreated,
				Status:     "201 Created",
				Body: io.NopCloser(strings.NewReader(`{
					"success": true,
					"message": "Lightning invoice created successfully",
					"data": {
						"id": "invoice_123",
						"payment_hash": "hash_create_123",
						"request": "lnbc500u1test",
						"amount_sat": 50000,
						"description": "Payment for order",
						"status": "pending",
						"expires_at": "2026-03-22T14:55:19Z",
						"created_at": "2026-03-22T13:55:19Z"
					},
					"metadata": {
						"request_id": "lightning_req_test"
					},
					"timestamp": "2026-03-22T13:55:19.860296234Z"
				}`)),
				Header:  make(http.Header),
				Request: req,
			}, nil
		}
		if req.URL.Path == "/api/lightning/decode" && req.Method == http.MethodPost {
			return &http.Response{
				StatusCode: http.StatusOK,
				Status:     "200 OK",
				Body: io.NopCloser(strings.NewReader(`{
					"success": true,
					"message": "Payment request decoded successfully",
					"data": {
						"payment_hash": "hash_decode_123",
						"destination": "02e880b95c737bbe110c5ef4f5e70d0d4d04253503e42c4de2760d519f9cb9a81e",
						"amount_sat": 50000,
						"amount_msat": 50000000,
						"description": "Payment for order",
						"expires_at": 1774191318
					},
					"metadata": {
						"request_id": "lightning_req_test"
					},
					"timestamp": "2026-03-22T13:57:08.638205573Z"
				}`)),
				Header:  make(http.Header),
				Request: req,
			}, nil
		}
		if req.URL.Path == "/api/lightning/payments/initiate" && req.Method == http.MethodPost {
			return &http.Response{
				StatusCode: http.StatusOK,
				Status:     "200 OK",
				Body: io.NopCloser(strings.NewReader(`{
					"success": true,
					"message": "Payment initiated successfully",
					"data": {
						"payment_hash": "hash_initiate_123",
						"destination": "02e880b95c737bbe110c5ef4f5e70d0d4d04253503e42c4de2760d519f9cb9a81e",
						"amount_sat": "1",
						"fee_sat": "20",
						"total_sat": "21",
						"description": "Payment for order",
						"expires_at": "1774192722",
						"is_expired": false,
						"route_found": true
					},
					"metadata": {
						"request_id": "lightning_req_test"
					},
					"timestamp": "2026-03-23T10:22:33.327842816Z"
				}`)),
				Header:  make(http.Header),
				Request: req,
			}, nil
		}
		if req.URL.Path == "/api/lightning/payments/send" && req.Method == http.MethodPost {
			return &http.Response{
				StatusCode: http.StatusCreated,
				Status:     "201 Created",
				Body: io.NopCloser(strings.NewReader(`{
					"success": true,
					"message": "Lightning payment sent successfully",
					"data": {
						"id": "payment_123",
						"payment_hash": "hash_send_123",
						"preimage": "preimage_123",
						"amount_sat": 1,
						"fee_sat": 20,
						"status": "paid",
						"created_at": "2026-03-22T14:19:12Z",
						"reference": "LT_PT25000900130010051104509",
						"transaction_id": "payment_123"
					},
					"metadata": {
						"request_id": "lightning_req_test"
					},
					"timestamp": "2026-03-22T14:19:12.848877395Z"
				}`)),
				Header:  make(http.Header),
				Request: req,
			}, nil
		}
		if req.URL.Path == "/api/lightning/invoices/invoice_123" && req.Method == http.MethodGet {
			return &http.Response{
				StatusCode: http.StatusOK,
				Status:     "200 OK",
				Body: io.NopCloser(strings.NewReader(`{
					"status": true,
					"message": "Lightning invoice retrieved successfully",
					"data": {
						"id": "invoice_123",
						"payment_hash": "hash_get_123",
						"request": "lnbc500u1test",
						"preimage": "",
						"amount_sat": 50000,
						"amount_msat": 50000000,
						"fee_sat": 0,
						"fee_msat": 0,
						"description": "Payment for order",
						"status": "pending",
						"direction": "incoming",
						"company_id": "company_test",
						"customer_id": "cust_abc123",
						"reference": "order_12345",
						"expires_at": 1742745600,
						"created_at": "2026-03-22T12:00:00Z",
						"updated_at": "2026-03-22T12:00:00Z"
					}
				}`)),
				Header:  make(http.Header),
				Request: req,
			}, nil
		}
		if req.URL.Path == "/api/lightning/invoices" && req.Method == http.MethodGet {
			return &http.Response{
				StatusCode: http.StatusOK,
				Status:     "200 OK",
				Body: io.NopCloser(strings.NewReader(`{
					"success": true,
					"message": "Lightning invoices retrieved successfully",
					"data": {
						"invoices": [
							{
								"id": "invoice_123",
								"payment_hash": "hash_list_123",
								"request": "lnbc500u1test",
								"amount_sat": 50000,
								"amount_msat": 50000000,
								"description": "Payment for order",
								"status": "pending",
								"direction": "incoming",
								"company_id": "company_test",
								"customer_id": "cust_abc123",
								"reference": "order_12345",
								"expires_at": 1742745600,
								"created_at": "2026-03-22T12:00:00Z"
							}
						],
						"total_count": 1,
						"page": 1,
						"page_size": 20
					},
					"metadata": {
						"request_id": "lightning_req_test"
					},
					"timestamp": "2026-03-22T14:21:40.134469256Z"
				}`)),
				Header:  make(http.Header),
				Request: req,
			}, nil
		}
		if req.URL.Path == "/api/lightning/invoices/verify" && req.Method == http.MethodPost {
			return &http.Response{
				StatusCode: http.StatusOK,
				Status:     "200 OK",
				Body: io.NopCloser(strings.NewReader(`{
					"status": true,
					"message": "Invoice payment verification completed",
					"data": {
						"id": "invoice_123",
						"payment_hash": "hash_verify_123",
						"preimage": "preimage_verify_123",
						"amount_sat": 50000,
						"status": "paid",
						"is_paid": true,
						"ledger_credited": true,
						"settled_at": "2026-03-22T12:05:00Z"
					}
				}`)),
				Header:  make(http.Header),
				Request: req,
			}, nil
		}
		if req.URL.Path == "/api/addresses" && req.Method == http.MethodGet {
			return &http.Response{
				StatusCode: http.StatusOK,
				Status:     "200 OK",
				Body: io.NopCloser(strings.NewReader(`{
					"success": true,
					"message": "Addresses retrieved successfully",
					"data": {
						"addresses": [
							{
								"id": "address_123",
								"address": "0x4C48BcdFA71deD181dE0441BAF72b2c8A599e6Cc",
								"company_id": "company_test",
								"customer_id": "customer_test",
								"label": "Primary Deposit Address",
								"chain": "arbitrum",
								"status": "active",
								"created_at": {
									"seconds": 1765378529,
									"nanos": 599608000
								}
							}
						],
						"page": 1,
						"per_page": 20,
						"total": 1,
						"total_pages": 1
					},
					"metadata": {
						"request_id": "address_req_test"
					},
					"timestamp": "2025-12-10T14:58:19.078048348Z"
				}`)),
				Header:  make(http.Header),
				Request: req,
			}, nil
		}
		if req.URL.Path == "/api/addresses/address_123" && req.Method == http.MethodGet {
			return &http.Response{
				StatusCode: http.StatusOK,
				Status:     "200 OK",
				Body: io.NopCloser(strings.NewReader(`{
					"success": true,
					"message": "Address retrieved successfully",
					"data": {
						"id": "address_123",
						"address": "tb1qsc0n8pnru9773a0kf5kke45cf3ne3ht9xsrpkq",
						"company_id": "company_test",
						"chain": "bitcoin",
						"status": "active",
						"created_at": {
							"seconds": 1767612670,
							"nanos": 911063000
						}
					},
					"metadata": {
						"request_id": "address_req_test"
					},
					"timestamp": "2026-03-21T22:17:44.339188493Z"
				}`)),
				Header:  make(http.Header),
				Request: req,
			}, nil
		}
		if req.URL.Path == "/api/addresses" && req.Method == http.MethodPost {
			return &http.Response{
				StatusCode: http.StatusOK,
				Status:     "200 OK",
				Body: io.NopCloser(strings.NewReader(`{
					"success": true,
					"message": "Address generated successfully",
					"data": {
						"id": "address_generated_123",
						"chain": "ethereum",
						"company_id": "company_test",
						"address": "0x3752c3f2a2301AaBE0Ab13e2B4308f940Fc2ecCA",
						"status": "active",
						"label": "Primary Deposit Address",
						"reference": "wallet_ref_123"
					},
					"metadata": {
						"request_id": "address_req_test"
					},
					"timestamp": "2026-03-21T22:38:02.262962198Z"
				}`)),
				Header:  make(http.Header),
				Request: req,
			}, nil
		}
		if req.URL.Path == "/api/addresses/validate" && req.Method == http.MethodPost {
			return &http.Response{
				StatusCode: http.StatusOK,
				Status:     "200 OK",
				Body: io.NopCloser(strings.NewReader(`{
					"success": true,
					"message": "Address validation completed",
					"data": {
						"valid": true,
						"address": "36FJHVQDuEqRCz8Qcx3ZCkAcHnmC6sgPZjvUVgTUtvHY",
						"chain": "solana"
					},
					"metadata": {
						"request_id": "address_req_test"
					},
					"timestamp": "2025-12-12T12:36:48.525315945Z"
				}`)),
				Header:  make(http.Header),
				Request: req,
			}, nil
		}
		if req.URL.Path == "/api/stablecoins/supported-chains" && req.Method == http.MethodGet {
			return &http.Response{
				StatusCode: http.StatusOK,
				Status:     "200 OK",
				Body: io.NopCloser(strings.NewReader(`{
					"success": true,
					"message": "Supported chains retrieved successfully",
					"data": {
						"chains": [
							{
								"chain": "arbitrum",
								"nativeToken": {"symbol": "ETH", "decimals": 18},
								"stablecoins": [
									{"symbol": "USDT", "decimals": 6},
									{"symbol": "USDC", "decimals": 6}
								]
							}
						]
					}
				}`)),
				Header:  make(http.Header),
				Request: req,
			}, nil
		}
		if req.URL.Path == "/api/trading/prices" && req.Method == http.MethodGet {
			return &http.Response{
				StatusCode: http.StatusOK,
				Status:     "200 OK",
				Body: io.NopCloser(strings.NewReader(`{
					"success": true,
					"message": "Prices retrieved successfully",
					"data": {
						"prices": [
							{
								"base_currency": "USDC",
								"quote_currency": "USDT",
								"price": "1.000450",
								"as_of": "2026-03-26T22:33:37.533303479Z"
							},
							{
								"base_currency": "BTC",
								"quote_currency": "USDT",
								"price": "68789.950",
								"as_of": "2026-03-26T22:33:42.918132025Z"
							}
						]
					},
					"metadata": {
						"request_id": "trading_req_test"
					},
					"timestamp": "2026-03-26T22:33:43.712587189Z"
				}`)),
				Header:  make(http.Header),
				Request: req,
			}, nil
		}
		if req.URL.Path == "/api/trading/quotes" && req.Method == http.MethodPost {
			body, err := io.ReadAll(req.Body)
			if err != nil {
				return nil, err
			}
			if !strings.Contains(string(body), `"base_currency":"BTC"`) || !strings.Contains(string(body), `"quote_currency":"USDT"`) || !strings.Contains(string(body), `"side":"buy"`) || !strings.Contains(string(body), `"quantity":"0.0001"`) {
				return &http.Response{
					StatusCode: http.StatusBadRequest,
					Status:     "400 Bad Request",
					Body:       io.NopCloser(strings.NewReader(`{"error":"unexpected trading quote payload"}`)),
					Header:     make(http.Header),
					Request:    req,
				}, nil
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Status:     "200 OK",
				Body: io.NopCloser(strings.NewReader(`{
					"success": true,
					"message": "Quote retrieved successfully",
					"data": {
						"quote": {
							"id": "quote_123",
							"base_currency": "BTC",
							"quote_currency": "USDT",
							"side": "BUY",
							"quantity": "0.0001",
							"price": "69034.33817175",
							"expires_at": "2026-03-26T21:14:45.357876312Z",
							"created_at": "2026-03-26T21:14:15.357876782Z",
							"original_quantity": "0.0001",
							"consumed_quantity": "0",
							"remaining_quantity": "0.0001",
							"is_exhausted": false,
							"exchange": {
								"send_quantity": "6.91",
								"send_currency": "USDT",
								"receive_quantity": "0.00010000",
								"receive_currency": "BTC"
							}
						}
					},
					"metadata": {
						"request_id": "trading_req_test"
					},
					"timestamp": "2026-03-26T21:14:15.38207726Z"
				}`)),
				Header:  make(http.Header),
				Request: req,
			}, nil
		}
		if req.URL.Path == "/api/trading/orders" && req.Method == http.MethodPost {
			body, err := io.ReadAll(req.Body)
			if err != nil {
				return nil, err
			}
			if !strings.Contains(string(body), `"quote_id":"quote_123"`) || !strings.Contains(string(body), `"reference":"324567890"`) {
				return &http.Response{
					StatusCode: http.StatusBadRequest,
					Status:     "400 Bad Request",
					Body:       io.NopCloser(strings.NewReader(`{"error":"unexpected trading order payload"}`)),
					Header:     make(http.Header),
					Request:    req,
				}, nil
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Status:     "200 OK",
				Body: io.NopCloser(strings.NewReader(`{
					"success": true,
					"message": "Order submitted successfully",
					"data": {
						"order": {
							"id": "order_123",
							"base_currency": "BTC",
							"quote_currency": "USDT",
							"side": "BUY",
							"order_type": "limit",
							"quantity": "0.0001",
							"price": "90822.540125",
							"status": "filled",
							"created_at": "2026-03-18T12:38:37.514379641Z",
							"updated_at": "2026-03-18T12:38:37.514379641Z",
							"filled_quantity": "0.0001",
							"remaining_quantity": "0",
							"quote_id": "quote_123"
						},
						"fills": [
							{
								"id": "fill_123",
								"order_id": "order_123",
								"quantity": "0.0001",
								"created_at": "2026-03-18T12:38:37.631040870Z"
							}
						],
						"quote_consumption": {
							"quote_id": "quote_123",
							"consumed_quantity": "0.0001",
							"remaining_quantity": "0"
						}
					},
					"metadata": {
						"request_id": "trading_req_test"
					},
					"timestamp": "2026-03-18T12:38:37.636126503Z"
				}`)),
				Header:  make(http.Header),
				Request: req,
			}, nil
		}
		if req.URL.Path == "/api/trading/orders" && req.Method == http.MethodGet {
			return &http.Response{
				StatusCode: http.StatusOK,
				Status:     "200 OK",
				Body: io.NopCloser(strings.NewReader(`{
					"success": true,
					"message": "Orders retrieved successfully",
					"data": {
						"orders": [
							{
								"id": "order_123",
								"base_currency": "BTC",
								"quote_currency": "USDT",
								"side": "SELL",
								"order_type": "limit",
								"quantity": "0.00010000",
								"price": "90217.44112500",
								"status": "filled",
								"exchange": "internal",
								"created_at": "2025-12-13T08:35:57.635498000Z",
								"updated_at": "2025-12-13T08:36:00.521536000Z",
								"filled_quantity": "0.00010000",
								"remaining_quantity": "0"
							}
						]
					},
					"metadata": {
						"request_id": "trading_req_test"
					},
					"timestamp": "2025-12-13T10:23:28.71684422Z"
				}`)),
				Header:  make(http.Header),
				Request: req,
			}, nil
		}
		if req.URL.Path == "/api/trading/orders/order_123" && req.Method == http.MethodGet {
			return &http.Response{
				StatusCode: http.StatusOK,
				Status:     "200 OK",
				Body: io.NopCloser(strings.NewReader(`{
					"success": true,
					"message": "Order retrieved successfully",
					"data": {
						"order": {
							"id": "order_123",
							"base_currency": "BTC",
							"quote_currency": "USDT",
							"side": "BUY",
							"order_type": "limit",
							"quantity": "0.00010000",
							"price": "90822.540125",
							"status": "filled",
							"exchange": "internal",
							"created_at": "2025-12-13T11:35:03.121902000Z",
							"updated_at": "2025-12-13T11:35:05.927393000Z",
							"filled_quantity": "0.00010000",
							"remaining_quantity": "0"
						}
					},
					"metadata": {
						"request_id": "trading_req_test"
					},
					"timestamp": "2025-12-13T11:36:28.587373651Z"
				}`)),
				Header:  make(http.Header),
				Request: req,
			}, nil
		}
		if req.URL.Path == "/api/wallets/transfers" && req.Method == http.MethodPost {
			body, err := io.ReadAll(req.Body)
			if err != nil {
				return nil, err
			}
			if !strings.Contains(string(body), `"to_address":"0xdaBe4B0Ca57dfBF13763D5f190A2d30B94f1Bf59"`) ||
				!strings.Contains(string(body), `"amount":"2000000"`) ||
				!strings.Contains(string(body), `"currency":"USDT"`) ||
				!strings.Contains(string(body), `"chain":"bsc"`) ||
				!strings.Contains(string(body), `"reference":"testing_bsc_013"`) {
				return &http.Response{
					StatusCode: http.StatusBadRequest,
					Status:     "400 Bad Request",
					Body:       io.NopCloser(strings.NewReader(`{"error":"unexpected transfer payload"}`)),
					Header:     make(http.Header),
					Request:    req,
				}, nil
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Status:     "200 OK",
				Body: io.NopCloser(strings.NewReader(`{
					"success": true,
					"message": "Transfer initiated successfully",
					"data": {
						"transaction_id": "transfer_123",
						"status": "pending",
						"address": "0xdaBe4B0Ca57dfBF13763D5f190A2d30B94f1Bf59",
						"amount": "2000000",
						"currency": "USDT",
						"chain": "bsc",
						"network": "mainnet",
						"reference": "testing_bsc_013",
						"created_at": "2025-12-12T13:35:43Z",
						"description": "Testing on USDT to binance"
					},
					"metadata": {
						"request_id": "wallet_req_test"
					},
					"timestamp": "2025-12-12T13:35:43.77369979Z"
				}`)),
				Header:  make(http.Header),
				Request: req,
			}, nil
		}
		if req.URL.Path == "/api/payouts/quotes" && req.Method == http.MethodPost {
			body, err := io.ReadAll(req.Body)
			if err != nil {
				return nil, err
			}
			if !strings.Contains(string(body), `"from_asset":"BTC"`) ||
				!strings.Contains(string(body), `"to_currency":"NGN"`) ||
				!strings.Contains(string(body), `"source":"onchain"`) ||
				!strings.Contains(string(body), `"chain":"BITCOIN"`) ||
				!strings.Contains(string(body), `"amount":"0.0015"`) ||
				!strings.Contains(string(body), `"payment_reason":"Vendor payout"`) ||
				!strings.Contains(string(body), `"reference":"offramp-quote-001"`) ||
				!strings.Contains(string(body), `"country":"NG"`) {
				return &http.Response{
					StatusCode: http.StatusBadRequest,
					Status:     "400 Bad Request",
					Body:       io.NopCloser(strings.NewReader(`{"error":"unexpected payouts quote payload"}`)),
					Header:     make(http.Header),
					Request:    req,
				}, nil
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Status:     "200 OK",
				Body: io.NopCloser(strings.NewReader(`{
					"status": "success",
					"message": "Payouts quote generated successfully",
					"data": {
						"quote": {
							"id": "5ddbcd32-3647-4d0a-b1fb-e4769bb9bf04",
							"status": "quote",
							"settlement_currency": "NGN",
							"quote_id": "QT_120732",
							"settlement_amount": 32664,
							"btc_rate": 112240.9,
							"exchange_rate": 1633.2,
							"expiry_timestamp": 1757373541,
							"amount": "0.0015",
							"sat_amount": 17819,
							"expires_in_text": "This quote expires in 15 minutes",
							"quote_text": "NGN 32,664 will be settled for this transaction"
						}
					}
				}`)),
				Header:  make(http.Header),
				Request: req,
			}, nil
		}
		if req.URL.Path == "/api/payouts/quotes/QT_6158/initialize" && req.Method == http.MethodPost {
			body, err := io.ReadAll(req.Body)
			if err != nil {
				return nil, err
			}
			hasRawBeneficiary := strings.Contains(string(body), `"beneficiary":{"account_name":"Ada Okafor","account_number":"0123456789","bank_code":"058"}`)
			hasSavedBeneficiary := strings.Contains(string(body), `"beneficiary":{"account_name":"John Doe","account_number":"1234567890","bank_code":"058","country":"NG","type":"BANK"}`) ||
				strings.Contains(string(body), `"beneficiary":{"country":"NG","type":"BANK","account_name":"John Doe","account_number":"1234567890","bank_code":"058"}`)
			if !strings.Contains(string(body), `"customer_id":"customer_123"`) ||
				!strings.Contains(string(body), `"reference":"offramp-init-001"`) ||
				!strings.Contains(string(body), `"payment_reason":"Supplier settlement"`) ||
				(strings.Contains(string(body), `"callback_url"`) && !strings.Contains(string(body), `"callback_url":"https://example.com/webhooks/payouts"`)) ||
				(!hasRawBeneficiary && !hasSavedBeneficiary) {
				return &http.Response{
					StatusCode: http.StatusBadRequest,
					Status:     "400 Bad Request",
					Body:       io.NopCloser(strings.NewReader(`{"error":"unexpected payouts initialize payload"}`)),
					Header:     make(http.Header),
					Request:    req,
				}, nil
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Status:     "200 OK",
				Body: io.NopCloser(strings.NewReader(`{
					"status": "success",
					"message": "Payouts quote initialized successfully",
					"data": {
						"fees": 0,
						"id": "4c281868-7b0d-4c4f-bcd7-bb3bbc9248bc",
						"address": "TNgGpCkphycU4pLg2LSJGUdb1mdwgMP8xs",
						"chain": "trc20",
						"status": "initiated",
						"payment_eta": "3-5 minutes",
						"reference": "offramp-init-001",
						"from_asset": "USDT",
						"quote_id": "QT_6158",
						"payment_reason": "Supplier settlement",
						"settlement_currency": "NGN",
						"exchange_rate": 1636.21,
						"expiry_timestamp": 1736952410,
						"amount": "0.0015",
						"btc_amount": 0.00201816,
						"sat_amount": 201816,
						"expires_in_text": "This invoice expires in 15 minutes",
						"beneficiary_details": {
							"id": "41b1e004-ce45-4591-bebe-6a2debcf05fd",
							"status": "success",
							"country": "NG",
							"currency": "NGN",
							"created_at": "2025-01-15T14:30:48.773Z",
							"reference": "QT_6158_24ccd39d2345",
							"updated_at": "2025-01-15T14:30:48.773Z",
							"destination": {
								"type": "BANK",
								"bank_code": "000014",
								"account_name": "ADA OKAFOR",
								"account_number": "0123456789"
							}
						},
						"settlement_amount": 327242
					}
				}`)),
				Header:  make(http.Header),
				Request: req,
			}, nil
		}
		if req.URL.Path == "/api/payouts/beneficiary-lookup" && req.Method == http.MethodPost {
			body, err := io.ReadAll(req.Body)
			if err != nil {
				return nil, err
			}
			if !strings.Contains(string(body), `"country":"NG"`) ||
				!strings.Contains(string(body), `"account_number":"0123456789"`) ||
				!strings.Contains(string(body), `"bank_code":"058"`) ||
				!strings.Contains(string(body), `"type":"bank_account"`) {
				return &http.Response{
					StatusCode: http.StatusBadRequest,
					Status:     "400 Bad Request",
					Body:       io.NopCloser(strings.NewReader(`{"error":"unexpected payouts beneficiary lookup payload"}`)),
					Header:     make(http.Header),
					Request:    req,
				}, nil
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Status:     "200 OK",
				Body: io.NopCloser(strings.NewReader(`{
					"status": "success",
					"message": "Beneficiary account verified successfully",
					"data": {
						"account_name": "ADA OKAFOR",
						"account_number": "0123456789",
						"bank_code": "058"
					}
				}`)),
				Header:  make(http.Header),
				Request: req,
			}, nil
		}
		if req.URL.Path == "/api/payouts/limits" && req.Method == http.MethodGet {
			return &http.Response{
				StatusCode: http.StatusOK,
				Status:     "200 OK",
				Body: io.NopCloser(strings.NewReader(`{
					"status": "success",
					"message": "Payouts limits retrieved successfully",
					"data": [
						{
							"lower_limit": "1000",
							"higher_limit": "2000000",
							"currency": "NGN",
							"country": "NG",
							"rate": "1635.81",
							"usd_lower_limit": "0.61",
							"usd_higher_limit": "1222.64"
						}
					]
				}`)),
				Header:  make(http.Header),
				Request: req,
			}, nil
		}
		if req.URL.Path == "/api/transactions" && req.Method == http.MethodGet {
			return &http.Response{
				StatusCode: http.StatusOK,
				Status:     "200 OK",
				Body: io.NopCloser(strings.NewReader(`{
					"success": true,
					"message": "Transactions retrieved successfully",
					"data": {
						"transactions": [
							{
								"transaction_id": "txn_123",
								"account_number": "7882897427",
								"currency": "USDT",
								"type": "WITHDRAWAL_INITIATED",
								"state": "PENDING",
								"amount": "-9",
								"fee": "0",
								"reference": "DEBIT:b6c75784-2526-49b8-982d-ce75760eb290",
								"idempotency_key": "clerk_hold_b6c75784-2526-49b8-982d-ce75760eb290",
								"trade_id": "b6c75784-2526-49b8-982d-ce75760eb290",
								"metadata": {
									"side": "debit",
									"order_side": "Buy"
								},
								"created_at": "2025-12-13T08:34:58.731319+00:00",
								"updated_at": "2025-12-13T08:34:58.731319+00:00",
								"value_date": "2025-12-13",
								"amount_formatted": "-0.000009 USDT",
								"fee_formatted": "0.0 USDT",
								"side": "Debit"
							}
						],
						"next_cursor": "cursor_123",
						"previous_cursor": "",
						"total_count": 1,
						"has_more": false
					},
					"metadata": {
						"request_id": "transaction_req_test"
					},
					"timestamp": "2025-12-13T10:08:00.861267822Z"
				}`)),
				Header:  make(http.Header),
				Request: req,
			}, nil
		}
		if req.URL.Path == "/api/transactions/txn_123" && req.Method == http.MethodGet {
			return &http.Response{
				StatusCode: http.StatusOK,
				Status:     "200 OK",
				Body: io.NopCloser(strings.NewReader(`{
					"success": true,
					"message": "Transaction retrieved successfully",
					"data": {
						"transaction_id": "txn_123",
						"account_number": "7882897427",
						"currency": "USDT",
						"type": "WITHDRAWAL_INITIATED",
						"state": "PENDING",
						"amount": "-10000000",
						"fee": "1200000",
						"reference": "SND_SBLNC_a30e9c646aeb",
						"idempotency_key": "clerk_send_d850e3be-5dd5-4b84-944c-fb337118a488",
						"trade_id": null,
						"metadata": {
							"side": "debit",
							"channel": "onchain",
							"chain": "tron"
						},
						"created_at": "2025-09-04T12:51:16.620Z",
						"updated_at": "2025-09-04T12:51:16.620Z",
						"value_date": "2025-09-04",
						"amount_formatted": "-10.0 USDT",
						"fee_formatted": "1.2 USDT",
						"side": "Debit"
					},
					"metadata": {
						"request_id": "transaction_req_test"
					},
					"timestamp": "2025-12-13T10:08:00.861267822Z"
				}`)),
				Header:  make(http.Header),
				Request: req,
			}, nil
		}
		if req.URL.Path != "/api/balances" {
			return &http.Response{
				StatusCode: http.StatusNotFound,
				Status:     "404 Not Found",
				Body:       io.NopCloser(strings.NewReader(`{"error":"not found"}`)),
				Header:     make(http.Header),
				Request:    req,
			}, nil
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Body: io.NopCloser(strings.NewReader(`{
				"success": true,
				"message": "Company balances retrieved successfully",
				"data": {
					"company_id": "company_test",
					"accounts": [
						{
							"account_id": "acct_1",
							"account_number": "1234567890",
							"currency": "USDT",
							"ledger_balance": "16047521",
							"available_balance": "16047521",
							"ledger_balance_formatted": "16.047521 USDT",
							"available_balance_formatted": "16.047521 USDT",
							"created_at": "2026-01-03T11:35:07.767513+00:00"
						}
					]
				},
				"metadata": {
					"request_id": "balance_req_test"
				},
				"timestamp": "2026-03-21T11:23:30.326138723Z"
			}`)),
			Header:  make(http.Header),
			Request: req,
		}, nil
	}))
}

func newTestAppWithTransport(t *testing.T, base string, transport roundTripFunc) *app.App {
	t.Helper()

	application, err := app.New(context.Background(), app.Options{
		Version: version.Info{
			Version: "test",
			Commit:  "abc123",
			Date:    "2026-03-21",
		},
		ConfigPath:  filepath.Join(base, "config.json"),
		StateDir:    filepath.Join(base, "state"),
		SecretStore: platform.NewFileSecretStore(filepath.Join(base, "state", "secrets")),
		APIClient: api.NewClientWithOptions(api.Options{
			HTTPClient: &http.Client{Transport: transport},
		}),
	})
	if err != nil {
		t.Fatalf("app.New returned error: %v", err)
	}

	return application
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
