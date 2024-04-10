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
