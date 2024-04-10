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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAggregateStats(t *testing.T) {
	as := NewStats()

	as.SetCustomStats("xx", map[string]string{"a": "1"}, "host1")
	m := map[string]string{"b": "3", "a": "1"}
	as.UpdateCustomStats("xx", m, "host1")

	value, ok := as.GetCustomStats("host1", "xx")
	assert.Equal(t, true, ok)
	sm, ok := value.(map[string]string)
	assert.Equal(t, true, ok)
	assert.Equal(t, "1", sm["a"])

	as.UpdateCustomStats("i", 2, "host1")
	as.UpdateCustomStats("i", 3, "host1")
	value, _ = as.GetCustomStats("host1", "i")
	assert.Equal(t, int64(5), value)

	as.UpdateCustomStats("f", 2.1, "host1")
	as.UpdateCustomStats("f", 3.2, "host1")
	value, _ = as.GetCustomStats("host1", "f")
	assert.Equal(t, int64(5), int64(value.(float64)))

	as.UpdateCustomStats("s", "hello", "host1")
	as.UpdateCustomStats("s", " ", "host1")
	as.UpdateCustomStats("s", "world", "host1")

	value, _ = as.GetCustomStats("host1", "s")
	assert.Equal(t, "hello world", value)
}
