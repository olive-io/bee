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

package exec

import (
	"fmt"
	"os/exec"

	"github.com/d5/tengo/v2"
	"github.com/olive-io/bee/tengo/extra"
)

var (
	Importable tengo.Importable = NewExec()
)

type ImportExec struct {
	Attrs map[string]tengo.Object
}

func NewExec() *ImportExec {
	f := &ImportExec{}
	attrs := map[string]tengo.Object{}
	f.Attrs = attrs
	attrs["command"] = &tengo.UserFunction{Name: "command", Value: osExec}

	return f
}

// Import returns an immutable map for the module.
func (ie *ImportExec) Import(moduleName string) (interface{}, error) {
	return ie.AsImmutableMap(moduleName), nil
}

func (ie *ImportExec) Version() string {
	return "v1.0.0"
}

// AsImmutableMap converts builtin module into an immutable map.
func (ie *ImportExec) AsImmutableMap(name string) *tengo.ImmutableMap {
	attrs := make(map[string]tengo.Object, len(ie.Attrs))
	for k, v := range ie.Attrs {
		attrs[k] = v.Copy()
	}
	attrs["__module_name__"] = &tengo.String{Value: name}
	return &tengo.ImmutableMap{Value: attrs}
}

func osExec(args ...tengo.Object) (tengo.Object, error) {
	if len(args) == 0 {
		return nil, tengo.ErrWrongNumArguments
	}
	name, ok := tengo.ToString(args[0])
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "first",
			Expected: "string(compatible)",
			Found:    args[0].TypeName(),
		}
	}
	var execArgs []string
	for idx, arg := range args[1:] {
		execArg, ok := tengo.ToString(arg)
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     fmt.Sprintf("args[%d]", idx),
				Expected: "string(compatible)",
				Found:    args[1+idx].TypeName(),
			}
		}
		execArgs = append(execArgs, execArg)
	}
	return makeOSExecCommand(exec.Command(name, execArgs...)), nil
}

func makeOSExecCommand(cmd *exec.Cmd) *tengo.ImmutableMap {
	return &tengo.ImmutableMap{
		Value: map[string]tengo.Object{
			// combined_output() => bytes/error
			"combined_output": &tengo.UserFunction{
				Name:  "combined_output",
				Value: extra.FuncBYTEE(cmd.CombinedOutput),
			},
			// output() => bytes/error
			"output": &tengo.UserFunction{
				Name:  "output",
				Value: extra.FuncBYTEE(cmd.Output),
			}, //
			// run() => error
			"run": &tengo.UserFunction{
				Name:  "run",
				Value: extra.FuncARE(cmd.Run),
			}, //
			// start() => error
			"start": &tengo.UserFunction{
				Name:  "start",
				Value: extra.FuncARE(cmd.Start),
			}, //
			// wait() => error
			"wait": &tengo.UserFunction{
				Name:  "wait",
				Value: extra.FuncARE(cmd.Wait),
			}, //
			// set_path(path string)
			"set_path": &tengo.UserFunction{
				Name: "set_path",
				Value: func(args ...tengo.Object) (tengo.Object, error) {
					if len(args) != 1 {
						return nil, tengo.ErrWrongNumArguments
					}
					s1, ok := tengo.ToString(args[0])
					if !ok {
						return nil, tengo.ErrInvalidArgumentType{
							Name:     "first",
							Expected: "string(compatible)",
							Found:    args[0].TypeName(),
						}
					}
					cmd.Path = s1
					return tengo.UndefinedValue, nil
				},
			},
			// set_dir(dir string)
			"set_dir": &tengo.UserFunction{
				Name: "set_dir",
				Value: func(args ...tengo.Object) (tengo.Object, error) {
					if len(args) != 1 {
						return nil, tengo.ErrWrongNumArguments
					}
					s1, ok := tengo.ToString(args[0])
					if !ok {
						return nil, tengo.ErrInvalidArgumentType{
							Name:     "first",
							Expected: "string(compatible)",
							Found:    args[0].TypeName(),
						}
					}
					cmd.Dir = s1
					return tengo.UndefinedValue, nil
				},
			},
			// set_env(env array(string))
			"set_env": &tengo.UserFunction{
				Name: "set_env",
				Value: func(args ...tengo.Object) (tengo.Object, error) {
					if len(args) != 1 {
						return nil, tengo.ErrWrongNumArguments
					}

					var env []string
					var err error
					switch arg0 := args[0].(type) {
					case *tengo.Array:
						env, err = stringArray(arg0.Value, "first")
						if err != nil {
							return nil, err
						}
					case *tengo.ImmutableArray:
						env, err = stringArray(arg0.Value, "first")
						if err != nil {
							return nil, err
						}
					default:
						return nil, tengo.ErrInvalidArgumentType{
							Name:     "first",
							Expected: "array",
							Found:    arg0.TypeName(),
						}
					}
					cmd.Env = env
					return tengo.UndefinedValue, nil
				},
			},
		},
	}
}

func stringArray(arr []tengo.Object, argName string) ([]string, error) {
	var sarr []string
	for idx, elem := range arr {
		str, ok := elem.(*tengo.String)
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     fmt.Sprintf("%s[%d]", argName, idx),
				Expected: "string",
				Found:    elem.TypeName(),
			}
		}
		sarr = append(sarr, str.Value)
	}
	return sarr, nil
}
