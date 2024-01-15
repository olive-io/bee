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

package module_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"

	"github.com/olive-io/bee/module"
)

const tmp = `name: "my_ping"
long: "a long text for description"
script: "ping.tengo"
authors:
  - John, Jonh@gmail.com
example: |

params:
  - name: ip
    type: string
    short: I
    default: 127.0.0.1
    desc: "target ip address"
    example: "127.0.0.1"
  - name: count
    type: int
    short: C
    default: 3

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
	m, err := module.LoadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, m.Name, "my_ping")
}

func TestModule_Execute(t *testing.T) {
	dir, cancel := initDir(t, "foo")
	defer cancel()
	m, err := module.LoadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	c, err := m.Execute("ip=10.0.0.100")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, c.Name, "my_ping")
	flags := c.Flags()
	ip, _ := flags.GetString("ip")
	assert.Equal(t, ip, "10.0.0.100")

	c, err = m.Execute("")
	if err != nil {
		t.Fatal(err)
	}
	ip, _ = c.Flags().GetString("ip")
	assert.Equal(t, ip, "127.0.0.1")

	c.Flags().VisitAll(func(flag *pflag.Flag) {
		t.Logf("%s=%s\n", flag.Name, flag.Value.String())
	})
}
