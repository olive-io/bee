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

import "github.com/olive-io/bee/module"

var Hooks = map[string]*CommandHook{
	"bee.builtin.copy":  copyHook,
	"bee.builtin.fetch": fetchHook,
}

type CommandHook struct {
	PreRun  module.RunE
	Run     module.RunE
	PostRun module.RunE
}
