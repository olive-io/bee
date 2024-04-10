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

package filepath

import (
	"path/filepath"

	"github.com/d5/tengo/v2"

	"github.com/olive-io/bee/tengo/extra"
)

var (
	Importable tengo.Importable = NewFP()
)

type ImportFP struct {
	Attrs map[string]tengo.Object
}

func NewFP() *ImportFP {
	f := &ImportFP{}
	attrs := map[string]tengo.Object{
		"separator":      &tengo.String{Value: string(filepath.Separator)},
		"list_separator": &tengo.String{Value: string(filepath.ListSeparator)},
		"line_separator": &tengo.String{Value: LineSeparator},
		"clean": &tengo.UserFunction{
			Name:  "clean",
			Value: clean,
		},
		"is_local": &tengo.UserFunction{
			Name:  "is_local",
			Value: isLocal,
		},
		"to_slash": &tengo.UserFunction{
			Name:  "to_slash",
			Value: toSlash,
		},
		"from_slash": &tengo.UserFunction{
			Name:  "from_slash",
			Value: fromSlash,
		},
		"split_list": &tengo.UserFunction{
			Name:  "split_list",
			Value: splitList,
		},
		"split": &tengo.UserFunction{
			Name:  "split",
			Value: split,
		},
		"rel": &tengo.UserFunction{
			Name:  "rel",
			Value: rel,
		},
		"join": &tengo.UserFunction{
			Name:  "join",
			Value: join,
		},
		"abs": &tengo.UserFunction{
			Name:  "abs",
			Value: abs,
		},
		"ext": &tengo.UserFunction{
			Name:  "ext",
			Value: ext,
		},
		"base": &tengo.UserFunction{
			Name:  "base",
			Value: base,
		},
		"dir": &tengo.UserFunction{
			Name:  "dir",
			Value: dir,
		},
		"volume_name": &tengo.UserFunction{
			Name:  "volume_name",
			Value: volumeName,
		},
	}
	f.Attrs = attrs

	return f
}

// Import returns an immutable map for the module.
func (fp *ImportFP) Import(moduleName string) (interface{}, error) {
	return fp.AsImmutableMap(moduleName), nil
}

func (fp *ImportFP) Version() string {
	return "v1.0.0"
}

// AsImmutableMap converts builtin module into an immutable map.
func (fp *ImportFP) AsImmutableMap(name string) *tengo.ImmutableMap {
	attrs := make(map[string]tengo.Object, len(fp.Attrs))
	for k, v := range fp.Attrs {
		attrs[k] = v.Copy()
	}
	attrs["__module_name__"] = &tengo.String{Value: name}
	return &tengo.ImmutableMap{Value: attrs}
}

func clean(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 1 {
		return nil, tengo.ErrWrongNumArguments
	}
	path, ok := tengo.ToString(args[0])
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "path",
			Expected: "string(compatible)",
			Found:    args[0].TypeName(),
		}
	}

	return &tengo.String{Value: filepath.Clean(path)}, nil
}

func isLocal(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 1 {
		return nil, tengo.ErrWrongNumArguments
	}
	path, ok := tengo.ToString(args[0])
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "path",
			Expected: "string",
			Found:    args[0].TypeName(),
		}
	}

	return &extra.BoolP{Value: filepath.IsLocal(path)}, nil
}

func toSlash(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 1 {
		return nil, tengo.ErrWrongNumArguments
	}
	path, ok := tengo.ToString(args[0])
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "path",
			Expected: "string(compatible)",
			Found:    args[0].TypeName(),
		}
	}

	return tengo.FromInterface(filepath.ToSlash(path))
}

func fromSlash(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 1 {
		return nil, tengo.ErrWrongNumArguments
	}
	path, ok := tengo.ToString(args[0])
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "path",
			Expected: "string(compatible)",
			Found:    args[0].TypeName(),
		}
	}

	return tengo.FromInterface(filepath.FromSlash(path))
}

func splitList(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 1 {
		return nil, tengo.ErrWrongNumArguments
	}
	path, ok := tengo.ToString(args[0])
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "path",
			Expected: "string(compatible)",
			Found:    args[0].TypeName(),
		}
	}

	out := &tengo.Array{Value: []tengo.Object{}}
	parts := filepath.SplitList(path)
	for _, part := range parts {
		out.Value = append(out.Value, &tengo.String{Value: part})
	}
	return out, nil
}

func split(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 1 {
		return nil, tengo.ErrWrongNumArguments
	}
	path, ok := tengo.ToString(args[0])
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "path",
			Expected: "string(compatible)",
			Found:    args[0].TypeName(),
		}
	}

	dirText, fileText := filepath.Split(path)
	return tengo.FromInterface(map[string]interface{}{
		"dir":  dirText,
		"file": fileText,
	})
}

func rel(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 2 {
		return nil, tengo.ErrWrongNumArguments
	}
	basePath, ok := tengo.ToString(args[0])
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "basePath",
			Expected: "string(compatible)",
			Found:    args[0].TypeName(),
		}
	}

	target, ok := tengo.ToString(args[1])
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "targetPath",
			Expected: "string(compatible)",
			Found:    args[1].TypeName(),
		}
	}

	related, err := filepath.Rel(basePath, target)
	if err != nil {
		return nil, err
	}

	return tengo.FromInterface(related)
}

func join(args ...tengo.Object) (tengo.Object, error) {
	elems := make([]string, 0)
	for _, arg := range args {
		v, ok := tengo.ToString(arg)
		if ok {
			elems = append(elems, v)
		}
	}

	return tengo.FromInterface(filepath.Join(elems...))
}

func abs(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 1 {
		return nil, tengo.ErrWrongNumArguments
	}
	path, ok := tengo.ToString(args[0])
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "path",
			Expected: "string(compatible)",
			Found:    args[0].TypeName(),
		}
	}

	out, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	return tengo.FromInterface(out)
}

func ext(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 1 {
		return nil, tengo.ErrWrongNumArguments
	}
	path, ok := tengo.ToString(args[0])
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "path",
			Expected: "string(compatible)",
			Found:    args[0].TypeName(),
		}
	}

	return tengo.FromInterface(filepath.Ext(path))
}

func base(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 1 {
		return nil, tengo.ErrWrongNumArguments
	}
	path, ok := tengo.ToString(args[0])
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "path",
			Expected: "string(compatible)",
			Found:    args[0].TypeName(),
		}
	}

	return tengo.FromInterface(filepath.Base(path))
}

func dir(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 1 {
		return nil, tengo.ErrWrongNumArguments
	}
	path, ok := tengo.ToString(args[0])
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "path",
			Expected: "string(compatible)",
			Found:    args[0].TypeName(),
		}
	}

	return tengo.FromInterface(filepath.Dir(path))
}

func volumeName(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 1 {
		return nil, tengo.ErrWrongNumArguments
	}
	path, ok := tengo.ToString(args[0])
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "path",
			Expected: "string(compatible)",
			Found:    args[0].TypeName(),
		}
	}

	return tengo.FromInterface(filepath.VolumeName(path))
}
