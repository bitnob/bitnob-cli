package cli

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bitnob/bitnob-cli/internal/output"
)

func TestParseShellInput(t *testing.T) {
	t.Parallel()

	args, err := parseShellInput(`api post /api/customers --data "{\"email\":\"john@example.com\"}"`)
	if err != nil {
		t.Fatalf("parseShellInput returned error: %v", err)
	}
	if len(args) != 5 {
		t.Fatalf("unexpected args length: %d", len(args))
	}
	if args[4] != `{"email":"john@example.com"}` {
		t.Fatalf("unexpected json arg: %q", args[4])
	}
}

func TestParseShellInput_UnterminatedQuote(t *testing.T) {
	t.Parallel()

	_, err := parseShellInput(`balances "oops`)
	if err == nil {
		t.Fatal("expected parse error")
	}
}

func TestParseModeSwitch(t *testing.T) {
	t.Parallel()

	mode, err := parseModeSwitch("mode command")
	if err != nil {
		t.Fatalf("expected mode switch to parse: %v", err)
	}
	if mode != shellModeCommand {
		t.Fatalf("unexpected mode: %s", mode)
	}
}

func TestParseShellModeRejectsAssistant(t *testing.T) {
	t.Parallel()

	_, err := parseShellMode("assistant")
	if err == nil {
		t.Fatal("expected assistant mode rejection")
	}
}

func TestParseModeSwitchRejectsAssistant(t *testing.T) {
	t.Parallel()

	_, err := parseModeSwitch("mode assistant")
	if err == nil {
		t.Fatal("expected assistant mode rejection")
	}
}

func TestShellHistoryPath(t *testing.T) {
	t.Parallel()

	got := shellHistoryPath("/tmp/bitnob-state")
	want := filepath.Join("/tmp/bitnob-state", "shell", "history")
	if got != want {
		t.Fatalf("unexpected history path: got=%q want=%q", got, want)
	}
}

func TestExecuteShellArgsVersion(t *testing.T) {
	t.Parallel()

	application := newTestApp(t)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	printer := output.New(stdout, stderr)

	if err := executeShellArgs(context.Background(), printer, application, []string{"version"}); err != nil {
		t.Fatalf("executeShellArgs returned error: %v", err)
	}

	if got := stdout.String(); !strings.Contains(got, `"Version": "test"`) {
		t.Fatalf("unexpected output: %q", got)
	}
}
