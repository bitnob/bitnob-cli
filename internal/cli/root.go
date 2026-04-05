package cli

import (
	"context"
	"os"
	"strings"

	"github.com/bitnob/bitnob-cli/internal/app"
	"github.com/bitnob/bitnob-cli/internal/output"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

const (
	colorReset = "\033[0m"
	colorGreen = "\033[38;5;42m"
)

var isTerminal = term.IsTerminal

func Run(ctx context.Context, application *app.App, args []string) error {
	printer := output.New(os.Stdout, os.Stderr)
	return runWithPrinter(ctx, printer, application, args)
}

func runWithPrinter(ctx context.Context, printer output.Printer, application *app.App, args []string) error {
	if len(args) == 0 {
		return printer.Println(renderLandingPage(colorEnabled(printer.Stdout)))
	}

	root := NewRootCommand(ctx, printer, application)
	root.SetArgs(args)
	return root.ExecuteContext(ctx)
}

func NewRootCommand(ctx context.Context, printer output.Printer, application *app.App) *cobra.Command {
	root := &cobra.Command{
		Use:               "bitnob",
		Short:             "Bitnob CLI",
		SilenceUsage:      true,
		SilenceErrors:     true,
		CompletionOptions: cobra.CompletionOptions{DisableDefaultCmd: true},
	}

	root.SetContext(ctx)
	root.SetOut(printer.Stdout)
	root.SetErr(printer.Stderr)

	root.AddCommand(
		newAPICommand(printer, application),
		newAddressesCommand(printer, application),
		newBalancesCommand(printer, application),
		newBeneficiariesCommand(printer, application),
		newCardsCommand(printer, application),
		newCompletionCommand(printer),
		newConfigCommand(printer, application),
		newCustomersCommand(printer, application),
		newDoctorCommand(printer, application),
		newLightningCommand(printer, application),
		newListenCommand(printer, application),
		newLoginCommand(printer, application),
		newLogoutCommand(printer, application),
		newPayoutsCommand(printer, application),
		newSwitchCommand(printer, application),
		newTradingCommand(printer, application),
		newTransfersCommand(printer, application),
		newTransactionsCommand(printer, application),
		newWaitCommand(printer, application),
		newWalletsCommand(printer, application),
		newWhoAmICommand(printer, application),
		newVersionCommand(printer, application),
	)

	return root
}

func renderLandingPage(useColor bool) string {
	banner := strings.Join([]string{
		"██████╗ ██╗████████╗███╗   ██╗ ██████╗ ██████╗ ",
		"██╔══██╗██║╚══██╔══╝████╗  ██║██╔═══██╗██╔══██╗",
		"██████╔╝██║   ██║   ██╔██╗ ██║██║   ██║██████╔╝",
		"██╔══██╗██║   ██║   ██║╚██╗██║██║   ██║██╔══██╗",
		"██████╔╝██║   ██║   ██║ ╚████║╚██████╔╝██████╔╝",
		"╚═════╝ ╚═╝   ╚═╝   ╚═╝  ╚═══╝ ╚═════╝ ╚═════╝ ",
	}, "\n")

	if useColor {
		banner = colorGreen + banner + colorReset
	}

	lines := []string{
		banner,
		"",
		"Bitnob CLI",
		"",
		"Quick start:",
		"  bitnob login --interactive",
		"  bitnob whoami",
		"  bitnob balances",
		"  bitnob help",
	}

	return strings.Join(lines, "\n")
}

func colorEnabled(w any) bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}

	file, ok := w.(*os.File)
	if !ok {
		return false
	}

	return isTerminal(int(file.Fd()))
}
