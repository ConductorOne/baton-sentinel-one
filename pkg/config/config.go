package config

import (
	"github.com/conductorone/baton-sdk/pkg/field"
)

var (
	APIToken = field.StringField(
		"api-token",
		field.WithDisplayName("API token"),
		field.WithDescription("API token for your management console used to authenticate with SentinelOne API."),
		field.WithPlaceholder("Your SentinelOne API token"),
		field.WithRequired(true),
		field.WithIsSecret(true),
	)

	ManagementConsoleURL = field.StringField(
		"management-console-url",
		field.WithDisplayName("SentinelOne base URL"),
		field.WithDescription("Your management console URL."),
		field.WithPlaceholder("Your SentinelOne base URL"),
		field.WithRequired(true),
	)
)

//go:generate go run ./gen
var Config = field.NewConfiguration(
	[]field.SchemaField{
		APIToken,
		ManagementConsoleURL,
	},
	field.WithConnectorDisplayName("SentinelOne"),
	field.WithHelpUrl("/docs/baton/sentinelone"),
	field.WithIconUrl("/static/app-icons/sentinelone.svg"),
)
