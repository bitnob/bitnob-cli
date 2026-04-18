package cli

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/bitnob/bitnob-cli/internal/app"
	"github.com/bitnob/bitnob-cli/internal/output"
)

func runRawRequest(ctx context.Context, printer output.Printer, application *app.App, method string, path string, data string) error {
	var body []byte
	if data != "" {
		if !json.Valid([]byte(data)) {
			return fmt.Errorf("request body must be valid JSON")
		}
		body = []byte(data)
	}

	response, err := application.APIService.Do(ctx, method, path, body)
	if err != nil {
		return err
	}
	_, err = printer.Stdout.Write(append(response, '\n'))
	return err
}
