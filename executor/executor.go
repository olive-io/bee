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

package executor

import (
	"errors"
	"sync"

	"go.uber.org/multierr"
	"go.uber.org/zap"

	"github.com/olive-io/bee/executor/client"
	inv "github.com/olive-io/bee/inventory"
	"github.com/olive-io/bee/secret"
	"github.com/olive-io/bee/vars"
)

var (
	ErrHostNotExists = errors.New("host not exists")
	ErrInvalidClient = errors.New("invalid client kind")
)

type Executor struct {
	lg *zap.Logger

	inventory *inv.Manager
	passwords *secret.PasswordManager

	cmu     sync.RWMutex
	clients map[string]client.IClient
}

func NewExecutor(lg *zap.Logger, inventory *inv.Manager, passwords *secret.PasswordManager) *Executor {
	executor := &Executor{
		lg:        lg,
		inventory: inventory,
		passwords: passwords,
		clients:   map[string]client.IClient{},
	}
	return executor
}

// LoadSources builds the given source client.IClient, if the client.IClient already built, do nothing
func (e *Executor) LoadSources(sources ...string) error {
	var errs []error
	for _, source := range sources {
		_, err := e.GetClient(source)
		if err != nil {
			errs = append(errs, err)
		}
	}
	return multierr.Combine(errs...)
}

func (e *Executor) GetClient(name string) (client.IClient, error) {
	e.cmu.RLock()
	cc, ok := e.clients[name]
	e.cmu.RUnlock()

	if !ok {
		var err error
		cc, err = e.newClient(name)
		if err != nil {
			return nil, err
		}
		e.cmu.Lock()
		e.clients[name] = cc
		e.cmu.Unlock()
	}

	return cc, nil
}

func (e *Executor) newClient(name string) (client.IClient, error) {
	host, ok := e.inventory.FindHost(name)
	if !ok {
		return nil, ErrHostNotExists
	}

	kind := host.Vars[vars.BeeConnectVars]
	if kind == "" {
		kind = client.SSHClient
	} else {
		//TODO: todo when bee_connection=smart|local
	}

	var cc client.IClient
	var err error
	switch kind {
	case client.SSHClient:
		cc, err = e.buildSSHClient(host)
	case client.WinRMClient:
		cc, err = e.buildWinRMClient(host)
	case client.GRPCClient:
		cc, err = e.buildGRPCClient(host)
	default:
		return nil, ErrInvalidClient
	}

	return cc, err
}

func (e *Executor) RemoveClient(name string) (bool, error) {
	e.cmu.RLock()
	conn, ok := e.clients[name]
	e.cmu.RUnlock()
	if !ok {
		return false, nil
	}

	err := conn.Close()
	e.cmu.Lock()
	defer e.cmu.Unlock()
	delete(e.clients, name)
	return ok, err
}

func (e *Executor) Cleanup() error {
	var errs []error

	e.cmu.Lock()
	for _, cc := range e.clients {
		if err := cc.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	e.cmu.Unlock()
	return multierr.Combine(errs...)
}
