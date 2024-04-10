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

package manager_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/olive-io/bee/module/manager"
)

func Test_ModuleManager(t *testing.T) {
	mg, err := manager.NewModuleManager(zap.NewExample(), "../../build/")
	if err != nil {
		t.Fatal(err)
	}

	m, ok := mg.Find("ping")
	assert.Equal(t, true, ok)
	assert.Equal(t, "bee.builtin.ping", m.Name)
}
