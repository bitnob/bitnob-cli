package cli

import "testing"

func TestRawAPIPathEscapesSegments(t *testing.T) {
	t.Parallel()

	got := rawAPIPath("api", "cards", "card/123?x=1#frag", "transactions", "txn 1")
	want := "/api/cards/card%2F123%3Fx=1%23frag/transactions/txn%201"
	if got != want {
		t.Fatalf("rawAPIPath() = %q, want %q", got, want)
	}
}
