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
