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

package secret

import (
	"path"

	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/pebble"
	"go.uber.org/zap"

	"github.com/olive-io/bee/secret/rsa"
)

const (
	defaultPrefix = "_bee"
)

var (
	ErrNotFound    = errors.New("key not found")
	ErrDBOperation = errors.New("failed to operate db")
)

type PasswordManager struct {
	lg *zap.Logger
	db *pebble.DB
}

func NewPasswordManager(lg *zap.Logger, db *pebble.DB) *PasswordManager {
	if lg == nil {
		lg = zap.NewNop()
	}

	pm := &PasswordManager{
		lg: lg,
		db: db,
	}
	return pm
}

func (pm *PasswordManager) GetRawPassword(name string, opts ...OpOption) (string, error) {
	password, err := pm.GetRSAPassword(name, opts...)
	if err != nil {
		return "", err
	}
	rawPasswd, _ := rsa.Decode(password)
	return string(rawPasswd), nil
}

func (pm *PasswordManager) GetRSAPassword(name string, opts ...OpOption) ([]byte, error) {
	op := &OpOptions{}
	for _, opt := range opts {
		opt(op)
	}

	prefix := path.Join(defaultPrefix, "passwd")
	if op.Namespace != "" {
		prefix = path.Join(prefix, op.Namespace)
	}

	keyBuf := path.Join(prefix, name)
	value, closer, err := pm.db.Get([]byte(keyBuf))
	if err != nil {
		return nil, parseErr(err)
	}
	_ = closer.Close()
	return value, nil
}

func (pm *PasswordManager) SetPassword(name, password string, opts ...OpOption) error {
	op := &OpOptions{}
	for _, opt := range opts {
		opt(op)
	}

	rsaPasswd, _ := rsa.Encode([]byte(password))
	prefix := path.Join(defaultPrefix, "passwd")
	if op.Namespace != "" {
		prefix = path.Join(prefix, op.Namespace)
	}

	keyBuf := path.Join(prefix, name)
	wo := pm.writeOptions()

	err := pm.db.Set([]byte(keyBuf), rsaPasswd, wo)
	return parseErr(err)
}

func (pm *PasswordManager) writeOptions() *pebble.WriteOptions {
	wo := &pebble.WriteOptions{Sync: true}
	return wo
}

func parseErr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pebble.ErrNotFound) {
		return ErrNotFound
	}
	return errors.Join(err, ErrDBOperation)
}
