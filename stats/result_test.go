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

package stats

import (
	"encoding/json"
	"testing"
)

const result1 = `{
	"host": "127.0.0.1",
	"stdout": {
		"changed": true,
		"a": 1,
		"msg": "this is test message",
		"arr": [1, 2, 3],
		"sub": {
			"b": "a"
		}
	}
}`

func TestTaskStdout(t *testing.T) {
	r := &TaskResult{}
	err := json.Unmarshal([]byte(result1), r)
	if err != nil {
		t.Fatal(err)
	}
}
