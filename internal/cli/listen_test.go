package cli

import (
	"bytes"
	"context"
	"net/http"
	"testing"

	"github.com/bitnob/bitnob-cli/internal/output"
)

func TestListenDefaultAddrIsLocalhost(t *testing.T) {
	t.Parallel()

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

	cmd := newListenCommand(output.New(stdout, stderr), application)
	addrFlag := cmd.Flag("addr")
	if addrFlag == nil {
		t.Fatal("expected addr flag to be defined")
	}
	if addrFlag.DefValue != "127.0.0.1:8080" {
		t.Fatalf("unexpected default addr: %q", addrFlag.DefValue)
	}
}
