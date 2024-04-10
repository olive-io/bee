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

package hook

import (
	"github.com/olive-io/bee/executor/client"
	"github.com/olive-io/bee/module"
)

var fetchHook = &CommandHook{
	PreRun:  fetchPreRun,
	PostRun: fetchPostRun,
}

var fetchPreRun module.RunE = func(ctx *module.RunContext, options ...client.ExecOption) ([]byte, error) {
	return []byte(""), nil
}

var fetchPostRun module.RunE = func(ctx *module.RunContext, options ...client.ExecOption) ([]byte, error) {
	out := []byte("")

	fs := ctx.Cmd.Flags()
	src, err := fs.GetString("src")
	if err != nil {
		return nil, err
	}
	dst, err := fs.GetString("dst")
	if err != nil {
		return nil, err
	}

	err = ctx.Conn.Get(ctx, src, dst)
	if err != nil {
		return nil, err
	}

	return out, nil
}
