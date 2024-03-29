// Copyright 2024 The bee Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package grpc

import (
	"fmt"
	"time"

	"github.com/olive-io/bee/executor/client"
	"go.uber.org/zap"
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
