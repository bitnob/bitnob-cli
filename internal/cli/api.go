package cli

import (
	"encoding/json"
	"fmt"

	"github.com/bitnob/bitnob-cli/internal/app"
	"github.com/bitnob/bitnob-cli/internal/output"
	"github.com/spf13/cobra"
)

func newAPICommand(printer output.Printer, application *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "api",
		Short: "Make signed Bitnob API requests",
	}

	cmd.AddCommand(
		newAPIRequestCommand(printer, application, "get"),
		newAPIRequestCommand(printer, application, "post"),
		newAPIRequestCommand(printer, application, "put"),
		newAPIRequestCommand(printer, application, "patch"),
		newAPIRequestCommand(printer, application, "delete"),
	)

	return cmd
}

func newAPIRequestCommand(printer output.Printer, application *app.App, method string) *cobra.Command {
	var data string

	cmd := &cobra.Command{
		Use:   method + " <path>",
		Short: "Make a signed " + method + " request",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var body []byte
			if data != "" {
				if !json.Valid([]byte(data)) {
					return fmt.Errorf("request body must be valid JSON")
				}
				body = []byte(data)
			}

			response, err := application.APIService.Do(cmd.Context(), method, args[0], body)
			if err != nil {
				return err
			}

			_, err = printer.Stdout.Write(append(response, '\n'))
			return err
		},
	}

	if method == "post" || method == "put" || method == "patch" {
		cmd.Flags().StringVar(&data, "data", "", "JSON request body")
	}

	return cmd
}
