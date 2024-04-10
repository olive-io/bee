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

package ssh

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func newClient(t *testing.T) *Client {
	passwd, _ := os.LookupEnv("TEST_SSH_PASSWORD")
	cfg := NewAuthConfig(zap.NewExample(), "192.168.2.32:22", "root", passwd)

	c, err := NewClient(*cfg)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestNewClient(t *testing.T) {
	c := newClient(t)
	if err := c.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestClient_Put_Get(t *testing.T) {
	c := newClient(t)
	defer c.Close()

	ctx := context.Background()
	tf := filepath.Join(os.TempDir(), "test.txt")
	err := os.WriteFile(tf, []byte("hello world1"), os.ModePerm)
	if !assert.NoError(t, err) {
		return
	}

	err = c.Put(ctx, tf, "/tmp/test.txt")
	if !assert.NoError(t, err) {
		return
	}

	local := filepath.Join(os.TempDir(), "test1.txt")
	err = c.Get(ctx, "/tmp/test.txt", local)
	if !assert.NoError(t, err) {
		return
	}

	data, _ := os.ReadFile(local)
	if !assert.Equal(t, []byte("hello world1"), data) {
		return
	}
}

func TestClient_Execute(t *testing.T) {
	c := newClient(t)
	defer c.Close()

	ctx := context.Background()
	cmd, err := c.Execute(ctx, "tengo /root/ping.tengo --data=ping")
	if !assert.NoError(t, err) {
		return
	}
	defer cmd.Close()

	data, err := cmd.CombinedOutput()
	if !assert.NoError(t, err) {
		return
	}

	t.Logf("%v", string(data))
}
