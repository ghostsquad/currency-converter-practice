package config

import (
	pkgerrors "github.com/pkg/errors"
)

type Config struct {
	IOStreams   IOStreams
	Port        int    `env:"PORT" envDefault:"8080"`
	BindAddress string `env:"BIND_ADDRESS"`
}

func Empty(val string) bool {
	return val == ""
}

func ValidateConfig(g Config) error {
	var err error

	return pkgerrors.WithStack(err)
}
