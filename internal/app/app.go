package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bitnob/bitnob-cli/internal/addresses"
	"github.com/bitnob/bitnob-cli/internal/api"
	"github.com/bitnob/bitnob-cli/internal/auth"
	"github.com/bitnob/bitnob-cli/internal/balances"
	"github.com/bitnob/bitnob-cli/internal/beneficiaries"
	"github.com/bitnob/bitnob-cli/internal/cards"
	"github.com/bitnob/bitnob-cli/internal/config"
	"github.com/bitnob/bitnob-cli/internal/customers"
	"github.com/bitnob/bitnob-cli/internal/genericapi"
	"github.com/bitnob/bitnob-cli/internal/lightning"
	"github.com/bitnob/bitnob-cli/internal/output"
	"github.com/bitnob/bitnob-cli/internal/payouts"
	"github.com/bitnob/bitnob-cli/internal/platform"
	"github.com/bitnob/bitnob-cli/internal/profile"
	"github.com/bitnob/bitnob-cli/internal/trading"
	"github.com/bitnob/bitnob-cli/internal/transactions"
	"github.com/bitnob/bitnob-cli/internal/transfers"
	"github.com/bitnob/bitnob-cli/internal/version"
	"github.com/bitnob/bitnob-cli/internal/wallets"
	"github.com/bitnob/bitnob-cli/internal/webhook"
)

type Options struct {
	Version     version.Info
	ConfigPath  string
	StateDir    string
	APIClient   *api.Client
	SecretStore *platform.SecretStore
}

type App struct {
	Version              version.Info
	StateDir             string
	StartupWarnings      []string
	ConfigStore          *config.Store
	ConfigService        *config.Service
	CredentialsService   *auth.CredentialsService
	AddressesService     *addresses.Service
	BalancesService      *balances.Service
	BeneficiariesService *beneficiaries.Service
	CardsService         *cards.Service
	CustomersService     *customers.Service
	LightningService     *lightning.Service
	TransactionsService  *transactions.Service
	WalletsService       *wallets.Service
	APIService           *genericapi.Service
	IdentityService      *auth.IdentityService
	ProfileService       *profile.Service
	TradingService       *trading.Service
	TransfersService     *transfers.Service
	WebhookService       *webhook.Service
	Output               output.Printer
	PayoutsService       *payouts.Service
}

func New(ctx context.Context, opts Options) (*App, error) {
	configPath := opts.ConfigPath
	if configPath == "" {
		configPath = config.DefaultPath()
	}

	stateDir := opts.StateDir
	if stateDir == "" {
		stateDir = filepath.Dir(configPath)
	}

	configStore := config.NewStore(configPath)
	secretStore := opts.SecretStore
	var startupWarnings []string
	if secretStore == nil {
		backendMode := strings.ToLower(strings.TrimSpace(os.Getenv("BITNOB_SECRET_BACKEND")))
		if backendMode == "" {
			backendMode = "auto"
		}

		fallbackDir := filepath.Join(stateDir, "secrets")

		switch backendMode {
		case "keyring":
			secretStore = platform.NewSecretStore("bitnob-cli")
		case "file":
			secretStore = platform.NewFileSecretStore(fallbackDir)
			startupWarnings = append(startupWarnings, "warning: using file secret store; credentials are not stored in OS keyring")
		case "auto":
			autoStore, warning, err := platform.NewAutoSecretStore("bitnob-cli", fallbackDir)
			if err != nil {
				return nil, err
			}
			secretStore = autoStore
			if warning != "" {
				startupWarnings = append(startupWarnings, warning)
			}
		default:
			return nil, fmt.Errorf("unsupported BITNOB_SECRET_BACKEND %q (expected: auto, keyring, file)", backendMode)
		}
	}
	authStore := auth.NewStore(secretStore)
	apiClient := opts.APIClient
	if apiClient == nil {
		apiClient = api.NewClient()
	}

	_, startupWarning, err := configStore.LoadOrRecover(ctx)
	if err != nil {
		return nil, err
	}

	if startupWarning != "" {
		startupWarnings = append(startupWarnings, startupWarning)
	}

	return &App{
		Version:              opts.Version,
		StateDir:             stateDir,
		StartupWarnings:      startupWarnings,
		ConfigStore:          configStore,
		ConfigService:        config.NewService(configStore),
		CredentialsService:   auth.NewCredentialsService(configStore, authStore, apiClient),
		AddressesService:     addresses.NewService(configStore, authStore, apiClient),
		BalancesService:      balances.NewService(configStore, authStore, apiClient),
		BeneficiariesService: beneficiaries.NewService(configStore, authStore, apiClient),
		CardsService:         cards.NewService(configStore, authStore, apiClient),
		CustomersService:     customers.NewService(configStore, authStore, apiClient),
		LightningService:     lightning.NewService(configStore, authStore, apiClient),
		TransactionsService:  transactions.NewService(configStore, authStore, apiClient),
		WalletsService:       wallets.NewService(configStore, authStore, apiClient),
		APIService:           genericapi.NewService(configStore, authStore, apiClient),
		IdentityService:      auth.NewIdentityService(configStore, authStore, apiClient),
		ProfileService:       profile.NewService(configStore),
		TradingService:       trading.NewService(configStore, authStore, apiClient),
		TransfersService:     transfers.NewService(configStore, authStore, apiClient),
		WebhookService:       webhook.NewService(apiClient.HTTPClient),
		PayoutsService:       payouts.NewService(configStore, authStore, apiClient),
		Output:               output.New(os.Stdout, os.Stderr),
	}, nil
}
