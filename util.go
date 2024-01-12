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

package bee

import (
	"context"

	"github.com/cockroachdb/errors"
)

var (
	ErrConflict = errors.New("runtime conflict")
)

const (
	HomeKey = "beeHome"
	OsKey   = "goos"
)

func checkRepl(goos string, r Repl) (repl string, err error) {
	repl = string(r)
	if (goos == "windows" && r == Bash) ||
		(goos == "linux" && r == Powershell) {
		err = errors.Wrapf(ErrConflict, "exec %s in %s", r, goos)
	}
	if goos == "windows" {
		repl += ".exe"
	}
	return
}

func CtxValueDefault(ctx context.Context, key, defaultVal string) string {
	value := ctx.Value(key)
	vv, _ := value.(string)
	if vv == "" {
		vv = defaultVal
	}
	return vv
}
