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

package executor

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/cockroachdb/errors"
	cwr "github.com/olive-io/winrm"
	"go.uber.org/zap"
	cssh "golang.org/x/crypto/ssh"

	"github.com/olive-io/bee/executor/client"
	"github.com/olive-io/bee/executor/client/grpc"
	"github.com/olive-io/bee/executor/client/ssh"
	"github.com/olive-io/bee/executor/client/winrm"
	"github.com/olive-io/bee/parser"
	"github.com/olive-io/bee/secret"
	"github.com/olive-io/bee/vars"
)

func (e *Executor) buildSSHClient(host *parser.Host) (*ssh.Client, error) {
	lg := e.lg
	ch, name := host.Name, host.Name
	variables := host.Vars
	if val, ok := variables[vars.BeeHostVars]; ok {
		ch = val
	}
	ch, _, _ = strings.Cut(ch, ":")

	port := ssh.DefaultPort
	if val, ok := variables[vars.BeePortVars]; ok {
		if i, _ := strconv.ParseInt(val, 10, 64); i > 0 {
			port = int(i)
		}
	}
	addr := fmt.Sprintf("%s:%d", ch, port)
	user := ssh.DefaultUser
	if val, ok := variables[vars.BeeUserVars]; ok {
		user = val
	}

	lfields := []zap.Field{
		zap.String("client", "ssh"),
		zap.String("name", name),
		zap.String("addr", addr),
		zap.String("user", user)}

	authMethods := make([]cssh.AuthMethod, 0)

	passwd, _ := e.passwords.GetRawPassword(name, secret.WithNamespace("ssh"))
	if passwd != "" {
		authMethods = append(authMethods, cssh.Password(passwd))
	} else if v, ok := variables[vars.BeeSSHPasswdVars]; ok {
		authMethods = append(authMethods, cssh.Password(v))
	}

	privateKey := ""
	var err error
	if v, ok := variables[vars.BeeSSHPrivateKeyVars]; ok {
		privateKey = v
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, errors.Wrap(err, "load home dir")
		}
		privateKey = filepath.Join(home, ".ssh", "id_rsa")
	}

	data, _ := os.ReadFile(privateKey)
	if data != nil {
		var signer cssh.Signer
		if v := variables[vars.BeeSSHPassphraseVars]; v == "" {
			signer, err = cssh.ParsePrivateKey(data)
		} else {
			signer, err = cssh.ParsePrivateKeyWithPassphrase(data, []byte(v))
		}
		if err != nil {
			return nil, errors.Wrap(err, "parse ssh private key")
		}
		authMethods = append(authMethods, cssh.PublicKeys(signer))
	}

	ccfg := &cssh.ClientConfig{
		Config:          cssh.Config{},
		User:            user,
		Auth:            authMethods,
		HostKeyCallback: cssh.InsecureIgnoreHostKey(),
		Timeout:         client.DefaultDialTimeout,
	}

	scfg := ssh.Config{
		Network:      "tcp",
		Addr:         addr,
		ClientConfig: ccfg,
		Logger:       lg,
	}

	if err = scfg.Validate(); err != nil {
		return nil, err
	}

	lg.Debug("create new bee connection", lfields...)

	var sc *ssh.Client
	sc, err = ssh.NewClient(scfg)
	if err != nil {
		return nil, err
	}
	return sc, nil
}

func (e *Executor) buildWinRMClient(host *parser.Host) (*winrm.WinRM, error) {
	lg := e.lg

	ch, name := host.Name, host.Name
	variables := host.Vars
	if val, ok := variables[vars.BeeHostVars]; ok {
		ch = val
	}
	ch, _, _ = strings.Cut(ch, ":")

	port := winrm.DefaultWinRMPort
	if val, ok := variables[vars.BeePortVars]; ok {
		if i, _ := strconv.ParseInt(val, 10, 64); i > 0 {
			port = int(i)
		}
	}
	addr := fmt.Sprintf("%s:%d", ch, port)

	user := winrm.DefaultWinRMUser
	if val, ok := variables[vars.BeeUserVars]; ok {
		user = val
	}

	lfields := []zap.Field{
		zap.String("client", "winrm"),
		zap.String("name", name),
		zap.String("addr", addr),
		zap.String("user", user)}

	passwd, err := e.passwords.GetRawPassword(name, secret.WithNamespace("ssh"))
	if v, ok := variables[vars.BeeWMPasswdVars]; ok {
		passwd = v
	}

	endpoint := cwr.Endpoint{
		Host:     ch,
		Port:     port,
		Insecure: true,
		Timeout:  client.DefaultDialTimeout,
	}
	wcfg := winrm.Config{
		Endpoint: endpoint,
		Username: user,
		Password: passwd,
		Logger:   lg,
	}
	if err = wcfg.Validate(); err != nil {
		return nil, err
	}

	lg.Debug("create new bee connection", lfields...)

	var wc *winrm.WinRM
	wc, err = winrm.NewWinRM(wcfg)
	if err != nil {
		return nil, err
	}
	return wc, nil
}

func (e *Executor) buildGRPCClient(host *parser.Host) (*grpc.Client, error) {
	lg := e.lg

	ch, name := host.Name, host.Name
	variables := host.Vars
	if val, ok := variables[vars.BeeHostVars]; ok {
		ch = val
	}
	ch, _, _ = strings.Cut(ch, ":")

	port := grpc.DefaultGRPCPort
	if val, ok := variables[vars.BeePortVars]; ok {
		if i, _ := strconv.ParseInt(val, 10, 64); i > 0 {
			port = int(i)
		}
	}
	addr := fmt.Sprintf("%s:%d", ch, port)

	lfields := []zap.Field{
		zap.String("client", "grpc"),
		zap.String("name", name),
		zap.String("addr", addr)}

	var err error
	gcfg := grpc.Config{
		Address: addr,
		Timeout: client.DefaultDialTimeout,
	}
	if err = gcfg.Validate(); err != nil {
		return nil, err
	}

	lg.Debug("create new bee connection", lfields...)

	var cc *grpc.Client
	cc, err = grpc.NewClient(gcfg)
	if err != nil {
		return nil, err
	}
	return cc, nil
}
