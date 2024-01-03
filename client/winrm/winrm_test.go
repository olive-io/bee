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

package winrm

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/olive-io/bee/client"
)

func newwinrm(t *testing.T) *WinRM {
	passwd, _ := os.LookupEnv("TEST_WinRM_PASSWORD")
	cfg := NewConfig(zap.NewExample(), "192.168.2.164", "Administrator", passwd)

	c, err := NewWinRM(*cfg)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestNewWinRM(t *testing.T) {
	wr := newwinrm(t)
	wr.Close()
}

func TestClient_Put_Get(t *testing.T) {
	c := newwinrm(t)
	defer c.Close()

	ctx := context.Background()
	tf := filepath.Join(os.TempDir(), "test.txt")
	err := os.WriteFile(tf, []byte(`hello world`), os.ModePerm)
	if !assert.NoError(t, err) {
		return
	}

	err = c.Put(ctx, tf, "C:\\howlink\\test.txt")
	if !assert.NoError(t, err) {
		return
	}

	local := filepath.Join(os.TempDir(), "test1.txt")
	err = c.Get(ctx, "C:\\howlink\\test.txt", local)
	if !assert.NoError(t, err) {
		return
	}

	data, _ := os.ReadFile(local)
	t.Logf("Output: %s", string(data))
}

func TestWinRM_Execute(t *testing.T) {
	wr := newwinrm(t)
	wr.Close()

	ctx := context.Background()
	cmd, err := wr.Execute(ctx, "cmd", client.ExecWithArgs([]string{"/C", "ipconfig", "/all"}))
	if !assert.NoError(t, err) {
		return
	}

	reader, err := cmd.StdoutPipe()
	if !assert.NoError(t, err) {
		return
	}

	err = cmd.Start()
	if !assert.NoError(t, err) {
		return
	}

	err = cmd.Wait()
	if !assert.NoError(t, err) {
		return
	}

	data := make([]byte, 1024*32)
	n, _ := reader.Read(data)
	data = data[:n]
	t.Logf("Output: %v", string(data))
}
