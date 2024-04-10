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

	e.cmu.Lock()
	defer e.cmu.Unlock()
	delete(e.clients, name)
	err := conn.Close()
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
