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

package module

import "strings"

// Repl the kind of interpreter
type Repl string

const (
	Unknown    Repl = ""
	Tengo      Repl = "tengo"
	Bash       Repl = "bash"
	Powershell Repl = "powershell"
)

var ks = map[Repl][]string{
	Tengo:      []string{".tengo"},
	Bash:       []string{".bash", ".sh"},
	Powershell: []string{".ps", ".bat"},
}

func KnownExt(ext string) (Repl, bool) {
	for kind, exts := range ks {
		for _, item := range exts {
			if strings.HasSuffix(ext, item) {
				return kind, true
			}
		}
	}
	return Unknown, false
}
