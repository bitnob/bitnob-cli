package cli

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/bitnob/bitnob-cli/internal/output"
)

func TestLoginUsesEnvironmentCredentials(t *testing.T) {
	application := newTestApp(t)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	printer := output.New(stdout, stderr)
	ctx := context.Background()

	t.Setenv("BITNOB_CLIENT_ID", "env_client_123")
	t.Setenv("BITNOB_SECRET_KEY", "env_secret_123")

	if err := runWithPrinter(ctx, printer, application, []string{"login"}); err != nil {
		t.Fatalf("login returned error: %v", err)
	}

	stdout.Reset()
	if err := runWithPrinter(ctx, printer, application, []string{"whoami"}); err != nil {
		t.Fatalf("whoami returned error after env login: %v", err)
	}

	if got := stdout.String(); !strings.Contains(got, `"client_id": "env_client_123"`) {
		t.Fatalf("unexpected whoami output after env login: %q", got)
	}
}
