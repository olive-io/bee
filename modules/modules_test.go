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

package modules

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

const tmp = `name: "ping"
long: "a long text for description"
script: "ping.tengo"
authors:
  - John, Jonh@gmail.com
example: |

params:
  - name: ip
    type: string
    short: I
    desc: "target ip address"
    example: "127.0.0.1"

returns:
  - name: "a"
    type: "string"
    example: ""

commands:
  - name: "sub-command"
    long: "a long text for description"
    script: "sub.tengo"
    authors:
      - John, Jonh@gmail.com
      -
    examples: |

    params:
    returns:`

func initDir(t *testing.T, dir string) (string, func()) {
	root, err := os.MkdirTemp(os.TempDir(), dir)
	if err != nil {
		t.Fatal(err)
	}
	_ = os.WriteFile(filepath.Join(root, "bee.yml"), []byte(tmp), os.ModePerm)
	_ = os.WriteFile(filepath.Join(root, "ping.tengo"), []byte("#!/usr/bin/tengo"), os.ModePerm)
	_ = os.WriteFile(filepath.Join(root, "sub.tengo"), []byte("#!/usr/bin/tengo"), os.ModePerm)

	sub := filepath.Join(root, "ping")
	_ = os.Mkdir(sub, os.ModePerm)
	_ = os.WriteFile(filepath.Join(sub, "bee.yml"), []byte(tmp), os.ModePerm)
	_ = os.WriteFile(filepath.Join(sub, "ping.tengo"), []byte("#!/usr/bin/tengo"), os.ModePerm)
	_ = os.WriteFile(filepath.Join(sub, "sub.tengo"), []byte("#!/usr/bin/tengo"), os.ModePerm)
	return root, func() {
		_ = os.RemoveAll(root)
	}
}

func TestLoadDir(t *testing.T) {
	dir, cancel := initDir(t, "foo")
	defer cancel()
	m, err := LoadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, m.Name, "ping")
}
