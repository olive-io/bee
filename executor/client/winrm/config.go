/*
   Copyright 2024 The bee Authors

   This library is free software; you can redistribute it and/or
   modify it under the terms of the GNU Lesser General Public
   License as published by the Free Software Foundation; either
   version 2.1 of the License, or (at your option) any later version.

   This library is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
   Lesser General Public License for more details.

   You should have received a copy of the GNU Lesser General Public
   License along with this library;
*/

package winrm

import (
	"fmt"

	"github.com/olive-io/winrm"
	"go.uber.org/zap"

	"github.com/olive-io/bee/executor/client"
)

const (
	DefaultWinRMUser = "Administrator"
	DefaultWinRMPort = 5985
)

type Config struct {
	winrm.Endpoint

	Username string
	Password string

	Logger *zap.Logger
}

func NewConfig(lg *zap.Logger, host, username, password string) *Config {
	if lg == nil {
		lg = zap.NewNop()
	}
	cfg := &Config{
		Endpoint: winrm.Endpoint{
			Host:     host,
			Port:     DefaultWinRMPort,
			Insecure: true,
			Timeout:  client.DefaultDialTimeout,
		},
		Username: username,
		Password: password,
		Logger:   lg,
	}
	return cfg
}

func (cfg *Config) Validate() error {
	if cfg.Host == "" {
		return fmt.Errorf("missing host")
	}
	if cfg.Port == 0 {
		cfg.Port = DefaultWinRMPort
	}
	if !cfg.HTTPS {
		cfg.Insecure = true
	}

	if cfg.Username == "" {
		return fmt.Errorf("missing username")
	}
	if cfg.Password == "" {
		return fmt.Errorf("missing password")
	}

	if cfg.Timeout == 0 {
		cfg.Timeout = client.DefaultDialTimeout
	}

	return nil
}
