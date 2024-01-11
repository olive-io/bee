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

package manager_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/olive-io/bee/module/manager"
)

func initData(t *testing.T, dir string) (string, func()) {
	temp := filepath.Join(os.TempDir(), dir)
	err := os.MkdirAll(temp, os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
	_ = os.MkdirAll(filepath.Join(temp, "repl"), 0o644)
	cancel := func() {
		_ = os.RemoveAll(temp)
	}
	return temp, cancel
}

func Test_ModuleManager(t *testing.T) {
	dir, cancel := initData(t, ".bee")
	defer cancel()

	mg, err := manager.NewModuleManager(zap.NewExample(), dir)
	if err != nil {
		t.Fatal(err)
	}

	m, ok := mg.Find("ping")
	assert.Equal(t, true, ok)
	assert.Equal(t, "bee.builtin.ping", m.Name)
}
