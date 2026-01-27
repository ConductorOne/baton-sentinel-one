package config

import (
	"github.com/conductorone/baton-sdk/pkg/field"
)

var (
	APIToken = field.StringField(
		"api-token",
		field.WithDescription("API token for your management console used to authenticate with SentinelOne API."),
		field.WithRequired(true),
		field.WithIsSecret(true),
	)

	ManagementConsoleURL = field.StringField(
		"management-console-url",
		field.WithDescription("Your management console URL."),
		field.WithRequired(true),
	)
)

//go:generate go run ./gen
var Config = field.NewConfiguration([]field.SchemaField{
	APIToken,
	ManagementConsoleURL,
})
