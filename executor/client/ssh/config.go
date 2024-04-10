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

package ssh

import (
	"fmt"
	"strings"

	"go.uber.org/zap"
	"golang.org/x/crypto/ssh"

	"github.com/olive-io/bee/executor/client"
)

const (
	DefaultUser = "root"
	DefaultPort = 22
)

type Config struct {
	Network string
	Addr    string

	ClientConfig *ssh.ClientConfig
	Logger       *zap.Logger
}

func NewAuthConfig(lg *zap.Logger, host, user, password string) *Config {
	cfg := &Config{
		Network: "tcp",
		Addr:    host,
		ClientConfig: &ssh.ClientConfig{
			User:            user,
			Auth:            []ssh.AuthMethod{ssh.Password(password)},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			Timeout:         client.DefaultDialTimeout,
		},
		Logger: lg,
	}

	return cfg
}

func (cfg *Config) Validate() error {
	if cfg.Logger == nil {
		cfg.Logger = zap.NewExample()
	}

	if cfg.Addr == "" {
		return fmt.Errorf("missing address")
	}
	if !strings.Contains(cfg.Addr, ":") {
		cfg.Addr = fmt.Sprintf("%s:%d", cfg.Addr, DefaultPort)
	}

	if cfg.ClientConfig == nil {
		return fmt.Errorf("missing ClientConfig")
	}
	if cfg.ClientConfig.User == "" {
		cfg.ClientConfig.User = DefaultUser
	}

	if cfg.ClientConfig.Timeout == 0 {
		cfg.ClientConfig.Timeout = client.DefaultDialTimeout
	}

	return nil
}
