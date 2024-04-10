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

import (
	"bytes"

	"github.com/cockroachdb/errors"
)

var (
	ErrConflict = errors.New("runtime conflict")
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

func beautify(stdout []byte) []byte {
	return bytes.TrimSuffix(stdout, []byte("\n"))
}
