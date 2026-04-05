package cli

import (
	"sort"

	"github.com/bitnob/bitnob-cli/internal/app"
	"github.com/bitnob/bitnob-cli/internal/output"
	"github.com/bitnob/bitnob-cli/internal/profile"
	"github.com/bitnob/bitnob-cli/internal/transactions"
	"github.com/spf13/cobra"
)

type doctorCheck struct {
	Name   string         `json:"name"`
	Status string         `json:"status"`
	Detail string         `json:"detail,omitempty"`
	Data   map[string]any `json:"data,omitempty"`
}

type doctorReport struct {
	Status        string                    `json:"status"`
	ConfigPath    string                    `json:"config_path"`
	ActiveProfile string                    `json:"active_profile,omitempty"`
	Checks        []doctorCheck             `json:"checks"`
	WhoAmI        map[string]any            `json:"whoami,omitempty"`
	Summary       map[string]int            `json:"summary"`
	Profiles      map[string]map[string]any `json:"profiles,omitempty"`
}

func newDoctorCommand(printer output.Printer, application *app.App) *cobra.Command {
	var includeProfiles bool

	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Run configuration, auth, and API health checks",
		RunE: func(cmd *cobra.Command, _ []string) error {
			report := doctorReport{
				Status:     "ok",
				ConfigPath: application.ConfigStore.Path(),
				Summary:    map[string]int{"pass": 0, "fail": 0, "warn": 0},
			}

			cfg, err := application.ConfigStore.Load(cmd.Context())
			if err != nil {
				report.Status = "fail"
				report.Checks = append(report.Checks, doctorCheck{
					Name:   "config.load",
					Status: "fail",
					Detail: err.Error(),
				})
				report.Summary["fail"]++
				return printer.PrintJSON(report)
			}

			report.ActiveProfile = cfg.ActiveProfile
			report.Checks = append(report.Checks, doctorCheck{
				Name:   "config.load",
				Status: "pass",
				Detail: "config loaded successfully",
			})
			report.Summary["pass"]++

			active, err := profile.Active(cfg)
			if err != nil {
				report.Status = "fail"
				report.Checks = append(report.Checks, doctorCheck{
					Name:   "profile.active",
					Status: "fail",
					Detail: err.Error(),
				})
				report.Summary["fail"]++
				return printer.PrintJSON(report)
			}

			profileCheck := doctorCheck{
				Name:   "profile.active",
				Status: "pass",
				Detail: "active profile loaded",
				Data: map[string]any{
					"name":        active.Name,
					"auth_method": active.AuthMethod,
				},
			}
			if !cfg.Profiles[active.Name].CredentialsConfigured {
				profileCheck.Status = "warn"
				profileCheck.Detail = "profile is active but credentials are not marked as configured"
				report.Summary["warn"]++
				report.Status = "degraded"
			} else {
				report.Summary["pass"]++
			}
			report.Checks = append(report.Checks, profileCheck)

			if includeProfiles {
				report.Profiles = make(map[string]map[string]any, len(cfg.Profiles))
				names := make([]string, 0, len(cfg.Profiles))
				for name := range cfg.Profiles {
					names = append(names, name)
				}
				sort.Strings(names)
				for _, name := range names {
					p := cfg.Profiles[name]
					report.Profiles[name] = map[string]any{
						"auth_method":            p.AuthMethod,
						"credentials_configured": p.CredentialsConfigured,
						"is_active":              name == cfg.ActiveProfile,
					}
				}
			}

			whoami, err := application.IdentityService.WhoAmI(cmd.Context())
			if err != nil {
				report.Checks = append(report.Checks, doctorCheck{
					Name:   "auth.whoami",
					Status: "fail",
					Detail: err.Error(),
				})
				report.Summary["fail"]++
				report.Status = "fail"
				return printer.PrintJSON(report)
			}
			report.Checks = append(report.Checks, doctorCheck{
				Name:   "auth.whoami",
				Status: "pass",
				Detail: "credentials verified against live whoami endpoint",
				Data: map[string]any{
					"client_id":         whoami.ClientID,
					"environment":       whoami.Environment,
					"active_company_id": whoami.ActiveCompanyID,
					"active":            whoami.Active,
				},
			})
			report.Summary["pass"]++
			report.WhoAmI = map[string]any{
				"client_id":           whoami.ClientID,
				"client_name":         whoami.ClientName,
				"environment":         whoami.Environment,
				"active_company_id":   whoami.ActiveCompanyID,
				"permissions_count":   len(whoami.Permissions),
				"requests_per_minute": whoami.RateLimit.RequestsPerMinute,
				"requests_per_hour":   whoami.RateLimit.RequestsPerHour,
			}

			addProbe := func(name string, fn func() error) {
				err := fn()
				if err != nil {
					report.Checks = append(report.Checks, doctorCheck{
						Name:   name,
						Status: "warn",
						Detail: err.Error(),
					})
					report.Summary["warn"]++
					if report.Status == "ok" {
						report.Status = "degraded"
					}
					return
				}
				report.Checks = append(report.Checks, doctorCheck{
					Name:   name,
					Status: "pass",
					Detail: "probe succeeded",
				})
				report.Summary["pass"]++
			}

			addProbe("balances.list", func() error {
				_, err := application.BalancesService.Get(cmd.Context())
				return err
			})
			addProbe("transactions.list", func() error {
				_, err := application.TransactionsService.List(cmd.Context(), transactions.ListFilters{Limit: 1})
				return err
			})
			addProbe("trading.prices", func() error {
				_, err := application.TradingService.ListPrices(cmd.Context())
				return err
			})
			addProbe("payouts.limits", func() error {
				_, err := application.PayoutsService.Limits(cmd.Context())
				return err
			})

			return printer.PrintJSON(report)
		},
	}

	cmd.Flags().BoolVar(&includeProfiles, "profiles", false, "Include configured profiles in the report")
	return cmd
}
