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
  - name: "sub"
    long: "a long text for description"
    script: "sub.tengo"
    authors:
      - John, Jonh@gmail.com
    examples: |

    params:
      - name: ip
        type: string
        short: I
        default: 127.0.0.1
        desc: "target ip address"
        example: "127.0.0.1"
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
	c, err := m.Execute("my_ping.sub", "ip=10.0.0.100")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, c.Name, "sub")
	ip, _ := c.Flags().GetString("ip")
	assert.Equal(t, ip, "10.0.0.100")

	c, err = m.Execute("ip=127.0.0.1")
	if err != nil {
		t.Fatal(err)
	}
	ip, _ = c.Flags().GetString("ip")
	assert.Equal(t, ip, "127.0.0.1")

	c.Flags().VisitAll(func(flag *pflag.Flag) {
		t.Logf("%s=%s\n", flag.Name, flag.Value.String())
	})
}
