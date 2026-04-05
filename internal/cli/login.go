package cli

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/bitnob/bitnob-cli/internal/app"
	"github.com/bitnob/bitnob-cli/internal/output"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var promptLine = func(label string) (string, error) {
	fmt.Fprintf(os.Stdout, "%s: ", label)
	reader := bufio.NewReader(os.Stdin)
	value, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(value), nil
}

var promptSecret = func(label string) (string, error) {
	fmt.Fprintf(os.Stdout, "%s: ", label)
	value, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(os.Stdout)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(value)), nil
}

func newLoginCommand(printer output.Printer, application *app.App) *cobra.Command {
	var profileName string
	var clientID string
	var secretKey string
	var interactive bool

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate with Bitnob",
		Long: `Authenticate with Bitnob.

Credentials are verified with the whoami endpoint before being saved.
If --profile is not provided, the whoami environment becomes the profile name.`,
		Example: `  bitnob login --interactive
  BITNOB_CLIENT_ID=your-client-id BITNOB_SECRET_KEY=your-secret-key bitnob login
  bitnob login -i --client-id your-client-id
  bitnob login --profile sandbox --interactive
  bitnob login --profile live --interactive`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if strings.TrimSpace(clientID) == "" {
				clientID = strings.TrimSpace(os.Getenv("BITNOB_CLIENT_ID"))
			}
			if strings.TrimSpace(secretKey) == "" {
				secretKey = strings.TrimSpace(os.Getenv("BITNOB_SECRET_KEY"))
			}

			if interactive {
				var err error
				if strings.TrimSpace(clientID) == "" {
					clientID, err = promptLine("Client ID")
					if err != nil {
						return err
					}
				}
				if strings.TrimSpace(secretKey) == "" {
					secretKey, err = promptSecret("Secret Key")
					if err != nil {
						return err
					}
				}
			}

			if clientID == "" || secretKey == "" {
				return errors.New("client id and secret key are required; use --interactive, set BITNOB_CLIENT_ID/BITNOB_SECRET_KEY, or pass both flags")
			}

			result, err := application.CredentialsService.Login(cmd.Context(), profileName, clientID, secretKey)
			if err != nil {
				return err
			}

			return printer.PrintJSON(map[string]any{
				"status":  "ok",
				"message": "Credentials verified and stored for profile",
				"profile": result.Profile,
				"whoami":  result.WhoAmI,
			})
		},
	}

	cmd.Flags().StringVar(&profileName, "profile", "", "Profile to store credentials under; defaults to the environment returned by whoami")
	cmd.Flags().StringVar(&clientID, "client-id", "", "Bitnob client ID to store for the selected profile (or set BITNOB_CLIENT_ID)")
	cmd.Flags().StringVar(&secretKey, "secret-key", "", "Bitnob secret key to store for the selected profile (or set BITNOB_SECRET_KEY; interactive/env preferred to avoid shell history leaks)")
	cmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "Prompt for missing credentials")

	return cmd
}
