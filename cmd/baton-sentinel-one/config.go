package main

import (
	"context"
	"fmt"

	"github.com/conductorone/baton-sdk/pkg/cli"
	"github.com/spf13/cobra"
)

// config defines the external configuration required for the connector to run.
type config struct {
	cli.BaseConfig `mapstructure:",squash"` // Puts the base config options in the same place as the connector options

	Token         string `mapstructure:"api-token"`
	ManagementUrl string `mapstructure:"management-console-url"`
}

// validateConfig is run after the configuration is loaded, and should return an error if it isn't valid.
func validateConfig(ctx context.Context, cfg *config) error {
	if cfg.Token == "" {
		return fmt.Errorf("api token must be provided")
	}

	if cfg.ManagementUrl == "" {
		return fmt.Errorf("management console url must be provided")
	}

	return nil
}

// cmdFlags sets the cmdFlags required for the connector.
func cmdFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().String("api-token", "", "API token for your management console used to authenticate with SentinelOne API. ($BATON_API_TOKEN)")
	cmd.PersistentFlags().String("management-console-url", "", "Your management console url. ($BATON_MANAGEMENT_CONSOLE_URL)")
}
