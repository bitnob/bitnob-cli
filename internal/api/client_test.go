package api

import "testing"

func TestSignWithEmptyBody(t *testing.T) {
	clientID := "2b9e8b4b-d8dc-42e1-82c7-3afe8a9ecd1c"
	secretKey := "secret_de5b82d253d4ddfbcb34ced93e6d17c7288c598cda0a3696514f73f168110cef"
	timestamp := "1711020000"
	nonce := "00112233445566778899aabbccddeeff"

	got := sign(clientID, timestamp, nonce, "", secretKey)
	want := "0be2223c6785af8b608fffe8de180d07c299b1f8862fd541abc790d90c89eb90"

	if got != want {
		t.Fatalf("unexpected signature\nwant: %s\ngot:  %s", want, got)
	}
}

func TestSignWithJSONBody(t *testing.T) {
	clientID := "client_test"
	secretKey := "secret_test"
	timestamp := "1711020000"
	nonce := "abcdef0123456789abcdef0123456789"
	body := `{"email":"john.doe@example.com","first_name":"John"}`

	got := sign(clientID, timestamp, nonce, body, secretKey)
	want := "46ec3f628da36f918b9c73eb4b0039190e5f636873ceab75ff0a342df715b37d"

	if got != want {
		t.Fatalf("unexpected signature\nwant: %s\ngot:  %s", want, got)
	}
}
