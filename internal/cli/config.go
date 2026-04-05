package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/bitnob/bitnob-cli/internal/app"
	"github.com/bitnob/bitnob-cli/internal/config"
	"github.com/bitnob/bitnob-cli/internal/output"
	"github.com/spf13/cobra"
)

var runEditor = func(editor string, path string) error {
	cmd := exec.Command(editor, path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

type restoreResult struct {
	ConfigPath       string
	RestoredFrom     string
	PreviousBackedUp string
}

var runConfigRestore = func(configPath string, backupPath string) (restoreResult, error) {
	backupData, err := os.ReadFile(backupPath)
	if err != nil {
		return restoreResult{}, err
	}

	var cfg config.Config
	if err := json.Unmarshal(backupData, &cfg); err != nil {
		return restoreResult{}, fmt.Errorf("backup file is not a valid config JSON: %w", err)
	}

	absBackupPath, err := filepath.Abs(backupPath)
	if err != nil {
		absBackupPath = backupPath
	}

	result := restoreResult{
		ConfigPath:   configPath,
		RestoredFrom: absBackupPath,
	}

	if _, err := os.Stat(configPath); err == nil {
		preRestorePath := fmt.Sprintf("%s.pre-restore.%s", configPath, time.Now().UTC().Format("20060102T150405Z"))
		if err := os.Rename(configPath, preRestorePath); err != nil {
			return restoreResult{}, fmt.Errorf("backup current config before restore: %w", err)
		}
		result.PreviousBackedUp = preRestorePath

		store := config.NewStore(configPath)
		if err := store.Save(context.Background(), cfg); err != nil {
			_ = os.Rename(preRestorePath, configPath)
			return restoreResult{}, fmt.Errorf("write restored config: %w", err)
		}
		return result, nil
	} else if !os.IsNotExist(err) {
		return restoreResult{}, err
	}

	store := config.NewStore(configPath)
	if err := store.Save(context.Background(), cfg); err != nil {
		return restoreResult{}, fmt.Errorf("write restored config: %w", err)
	}

	return result, nil
}

func newConfigCommand(printer output.Printer, application *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Inspect and manage non-secret CLI configuration",
	}

	cmd.AddCommand(
		newConfigEditCommand(printer, application),
		newConfigListCommand(printer, application),
		newConfigGetCommand(printer, application),
		newConfigSetCommand(printer, application),
		newConfigUnsetCommand(printer, application),
		newConfigRestoreCommand(printer, application),
	)

	return cmd
}

func newConfigEditCommand(printer output.Printer, application *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "edit",
		Short: "Open the configuration file in an editor",
		RunE: func(cmd *cobra.Command, _ []string) error {
			editor := os.Getenv("EDITOR")
			if editor == "" {
				editor = "vi"
			}

			if err := runEditor(editor, application.ConfigStore.Path()); err != nil {
				return err
			}

			return printer.PrintJSON(map[string]string{
				"status":  "ok",
				"message": "Configuration file opened",
				"path":    application.ConfigStore.Path(),
				"editor":  editor,
			})
		},
	}
}

func newConfigListCommand(printer output.Printer, application *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List current CLI configuration",
		RunE: func(cmd *cobra.Command, _ []string) error {
			summary, err := application.ConfigService.Summary(cmd.Context())
			if err != nil {
				return err
			}

			return printer.PrintJSON(summary)
		},
	}
	return cmd
}

func newConfigGetCommand(printer output.Printer, application *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <key>",
		Short: "Get a configuration value",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			value, err := application.ConfigService.Get(cmd.Context(), args[0])
			if err != nil {
				return err
			}

			return printer.PrintJSON(value)
		},
	}
	return cmd
}

func newConfigSetCommand(printer output.Printer, application *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a configuration value",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			value, err := application.ConfigService.Set(cmd.Context(), args[0], args[1])
			if err != nil {
				return err
			}

			return printer.PrintJSON(value)
		},
	}
	return cmd
}

func newConfigUnsetCommand(printer output.Printer, application *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unset <key>",
		Short: "Unset a configuration value",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			value, err := application.ConfigService.Unset(cmd.Context(), args[0])
			if err != nil {
				return err
			}

			return printer.PrintJSON(value)
		},
	}
	return cmd
}

func newConfigRestoreCommand(printer output.Printer, application *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "restore <backup-file>",
		Short: "Restore configuration from a backup file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := runConfigRestore(application.ConfigStore.Path(), args[0])
			if err != nil {
				return err
			}

			return printer.PrintJSON(map[string]string{
				"status":               "ok",
				"message":              "Configuration restored from backup",
				"path":                 result.ConfigPath,
				"restored_from":        result.RestoredFrom,
				"previous_backup_path": result.PreviousBackedUp,
			})
		},
	}
}
