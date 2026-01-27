package main

import (
	cfg "github.com/conductorone/baton-sentinel-one/pkg/config"
	"github.com/conductorone/baton-sdk/pkg/config"
)

func main() {
	config.Generate("sentinelOne", cfg.Config)
}
