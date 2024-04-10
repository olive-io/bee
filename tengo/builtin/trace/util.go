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

package trace

import (
	"strings"

	"github.com/olive-io/bee/tengo/builtin/trace/internal"
)

func parseLevel(text string) (internal.Level, bool) {
	var level internal.Level
	switch text {
	case "debug":
		level = internal.LevelDebug
	case "info":
		level = internal.LevelInfo
	case "warn":
		level = internal.LevelWarn
	case "error":
		level = internal.LevelError
	default:
		return 0, false
	}
	return level, true
}

func unquote(s string) string {
	return strings.Trim(s, `"`)
}
