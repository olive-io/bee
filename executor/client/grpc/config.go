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

package grpc

import (
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/olive-io/bee/executor/client"
)

type Config struct {
	Address string
	Timeout time.Duration

	lg *zap.Logger
}

func NewConfig(lg *zap.Logger, addr string) *Config {
	if lg == nil {
		lg = zap.NewNop()
	}
	cfg := &Config{
		Address: addr,
		Timeout: client.DefaultDialTimeout,
		lg:      lg,
	}

	return cfg
}

func (cfg *Config) Validate() error {
	if cfg.Address == "" {
		return fmt.Errorf("missing address")
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = client.DefaultDialTimeout
	}
	if cfg.lg == nil {
		cfg.lg = zap.NewNop()
	}

	return nil
}
