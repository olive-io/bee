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
