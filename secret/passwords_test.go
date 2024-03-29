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

package secret_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/olive-io/bee/secret"
	testdb "github.com/olive-io/bee/test/db"
)

func newPM(t *testing.T) (*secret.PasswordManager, func()) {
	dir, err := os.MkdirTemp("", "*foo")
	if err != nil {
		t.Fatal(err)
	}
	lg := zap.NewExample()
	db, err := testdb.NewDB(lg, dir)
	if err != nil {
		t.Fatal(err)
	}
	pm := secret.NewPasswordManager(lg, db)
	if err != nil {
		t.Fatal(err)
	}
	cancel := func() {
		_ = db.Close()
		_ = os.RemoveAll(dir)
	}
	return pm, cancel
}

func Test_PasswordManager(t *testing.T) {
	pm, cancel := newPM(t)
	defer cancel()

	err := pm.SetPassword("web1", "password", secret.WithNamespace("ssh"))
	if err != nil {
		t.Fatal(err)
	}

	value, err := pm.GetRawPassword("web1", secret.WithNamespace("ssh"))
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "password", value)
}
