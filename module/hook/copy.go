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
	"path"
	"strings"

	"github.com/google/uuid"

	"github.com/olive-io/bee/executor/client"
	"github.com/olive-io/bee/module"
	"github.com/olive-io/bee/vars"
)

var copyHook = &CommandHook{
	PreRun: copyPreRun,
}

var copyPreRun module.RunE = func(ctx *module.RunContext, options ...client.ExecOption) ([]byte, error) {
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

	home := ctx.Variables.GetDefault(vars.BeeHome, ".bee")
	goos := ctx.Variables.GetDefault(vars.BeePlatformVars, "linux")

	dst = path.Join(home, "tmp", uuid.New().String()+path.Ext(dst))
	if goos == "windows" {
		dst = strings.ReplaceAll(dst, "/", "\\")
	}

	err = ctx.Conn.Put(ctx, src, dst)
	if err != nil {
		return nil, err
	}

	ctx.Variables.Set(module.PrefixFlag+"src", dst)

	return out, nil
}
