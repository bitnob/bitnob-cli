package cli

import (
	"github.com/bitnob/bitnob-cli/internal/app"
	"github.com/bitnob/bitnob-cli/internal/output"
	"github.com/spf13/cobra"
)

func newBulkTransfersCommand(printer output.Printer, application *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bulk-transfers",
		Short: "Manage bulk transfer batches, schedules, and executions",
	}

	cmd.AddCommand(
		newBulkTransfersBatchesCommand(printer, application),
		newBulkTransfersSchedulesCommand(printer, application),
		newBulkTransfersExecutionsCommand(printer, application),
	)

	return cmd
}

func newBulkTransfersBatchesCommand(printer output.Printer, application *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "batches",
		Short: "Manage bulk transfer batches",
	}

	cmd.AddCommand(
		newBulkTransfersCreateBatchCommand(printer, application),
		newBulkTransfersUploadURLsCommand(printer, application),
		newBulkTransfersPreviewCommand(printer, application),
		newBulkTransfersListBatchesCommand(printer, application),
		newBulkTransfersGetBatchCommand(printer, application),
		newBulkTransfersConfirmBatchCommand(printer, application),
		newBulkTransfersRetryBatchCommand(printer, application),
	)

	return cmd
}

func newBulkTransfersCreateBatchCommand(printer output.Printer, application *app.App) *cobra.Command {
	var data string
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create draft batch",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runRawRequest(cmd.Context(), printer, application, "POST", "/api/bulk-transfers", data)
		},
	}
	cmd.Flags().StringVar(&data, "data", "", "JSON request body")
	return cmd
}

func newBulkTransfersUploadURLsCommand(printer output.Printer, application *app.App) *cobra.Command {
	var data string
	cmd := &cobra.Command{
		Use:   "upload-urls",
		Short: "Get upload URL (S3 presigned)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runRawRequest(cmd.Context(), printer, application, "POST", "/api/bulk-transfers/upload-urls", data)
		},
	}
	cmd.Flags().StringVar(&data, "data", "", "JSON request body")
	return cmd
}

func newBulkTransfersPreviewCommand(printer output.Printer, application *app.App) *cobra.Command {
	var data string
	cmd := &cobra.Command{
		Use:   "preview",
		Short: "Preview uploaded file",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runRawRequest(cmd.Context(), printer, application, "POST", "/api/bulk-transfers/previews", data)
		},
	}
	cmd.Flags().StringVar(&data, "data", "", "JSON request body")
	return cmd
}

func newBulkTransfersListBatchesCommand(printer output.Printer, application *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List batches",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runRawRequest(cmd.Context(), printer, application, "GET", "/api/bulk-transfers", "")
		},
	}
}

func newBulkTransfersGetBatchCommand(printer output.Printer, application *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "get <batch-id>",
		Short: "Get batch by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRawRequest(cmd.Context(), printer, application, "GET", rawAPIPath("api", "bulk-transfers", args[0]), "")
		},
	}
}

func newBulkTransfersConfirmBatchCommand(printer output.Printer, application *app.App) *cobra.Command {
	var data string
	cmd := &cobra.Command{
		Use:   "confirm <batch-id>",
		Short: "Confirm batch",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRawRequest(cmd.Context(), printer, application, "POST", rawAPIPath("api", "bulk-transfers", args[0], "confirm"), data)
		},
	}
	cmd.Flags().StringVar(&data, "data", "", "JSON request body")
	return cmd
}

func newBulkTransfersRetryBatchCommand(printer output.Printer, application *app.App) *cobra.Command {
	var data string
	cmd := &cobra.Command{
		Use:   "retry <batch-id>",
		Short: "Retry failed items in batch",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRawRequest(cmd.Context(), printer, application, "POST", rawAPIPath("api", "bulk-transfers", args[0], "retry"), data)
		},
	}
	cmd.Flags().StringVar(&data, "data", "", "JSON request body")
	return cmd
}

func newBulkTransfersSchedulesCommand(printer output.Printer, application *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "schedules",
		Short: "Manage bulk transfer schedules",
	}

	cmd.AddCommand(
		newBulkTransfersSchedulesCreateCommand(printer, application),
		newBulkTransfersSchedulesListCommand(printer, application),
		newBulkTransfersSchedulesUpdateCommand(printer, application),
		newBulkTransfersSchedulesCancelCommand(printer, application),
		newBulkTransfersSchedulesExecutionsCommand(printer, application),
	)

	return cmd
}

func newBulkTransfersSchedulesCreateCommand(printer output.Printer, application *app.App) *cobra.Command {
	var data string
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create schedule",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runRawRequest(cmd.Context(), printer, application, "POST", "/api/bulk-transfers/schedules", data)
		},
	}
	cmd.Flags().StringVar(&data, "data", "", "JSON request body")
	return cmd
}

func newBulkTransfersSchedulesListCommand(printer output.Printer, application *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List schedules",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runRawRequest(cmd.Context(), printer, application, "GET", "/api/bulk-transfers/schedules", "")
		},
	}
}

func newBulkTransfersSchedulesUpdateCommand(printer output.Printer, application *app.App) *cobra.Command {
	var data string
	cmd := &cobra.Command{
		Use:   "update <schedule-id>",
		Short: "Update schedule",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRawRequest(cmd.Context(), printer, application, "PUT", rawAPIPath("api", "bulk-transfers", "schedules", args[0]), data)
		},
	}
	cmd.Flags().StringVar(&data, "data", "", "JSON request body")
	return cmd
}

func newBulkTransfersSchedulesCancelCommand(printer output.Printer, application *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "cancel <schedule-id>",
		Short: "Cancel schedule",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRawRequest(cmd.Context(), printer, application, "DELETE", rawAPIPath("api", "bulk-transfers", "schedules", args[0]), "")
		},
	}
}

func newBulkTransfersSchedulesExecutionsCommand(printer output.Printer, application *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "executions <schedule-id>",
		Short: "List schedule executions",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRawRequest(cmd.Context(), printer, application, "GET", rawAPIPath("api", "bulk-transfers", "schedules", args[0], "executions"), "")
		},
	}
}

func newBulkTransfersExecutionsCommand(printer output.Printer, application *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "executions",
		Short: "Manage bulk transfer executions",
	}
	cmd.AddCommand(newBulkTransfersExecutionsRetryCommand(printer, application))
	return cmd
}

func newBulkTransfersExecutionsRetryCommand(printer output.Printer, application *app.App) *cobra.Command {
	var data string
	cmd := &cobra.Command{
		Use:   "retry <execution-id>",
		Short: "Retry failed execution",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRawRequest(cmd.Context(), printer, application, "POST", rawAPIPath("api", "bulk-transfers", "executions", args[0], "retry"), data)
		},
	}
	cmd.Flags().StringVar(&data, "data", "", "JSON request body")
	return cmd
}
