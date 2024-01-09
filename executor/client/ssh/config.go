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

package ssh

import (
	"fmt"

	"go.uber.org/zap"
	"golang.org/x/crypto/ssh"

	"github.com/olive-io/bee/executor/client"
)

type Config struct {
	Network string
	Addr    string

	ClientConfig *ssh.ClientConfig
	lg           *zap.Logger
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
		lg: lg,
	}

	return cfg
}

func (cfg *Config) Validate() error {
	if cfg.lg == nil {
		cfg.lg = zap.NewExample()
	}

	if cfg.ClientConfig == nil {
		return fmt.Errorf("missing ClientConfig")
	}

	if cfg.ClientConfig.Timeout == 0 {
		cfg.ClientConfig.Timeout = client.DefaultDialTimeout
	}

	return nil
}
