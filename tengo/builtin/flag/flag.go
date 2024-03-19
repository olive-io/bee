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

package flag

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/d5/tengo/v2"
	"github.com/olive-io/bee/tengo/extra"
	"github.com/spf13/pflag"
)

var (
	ErrUndefined = errors.New("flag not defined")
	ErrFlagType  = errors.New("flag invalid type")

	Importable tengo.Importable = NewFlag()
)

type ImportFlag struct {
	fs    *pflag.FlagSet
	Attrs map[string]tengo.Object
}

func NewFlag() *ImportFlag {
	fs := pflag.NewFlagSet("", pflag.ContinueOnError)
	fs.ParseErrorsWhitelist.UnknownFlags = true
	f := &ImportFlag{
		fs: fs,
	}
	attrs := map[string]tengo.Object{}
	attrs["int"] = &tengo.UserFunction{Name: "int", Value: f.Int()}
	attrs["int_array"] = &tengo.UserFunction{Name: "int_array", Value: f.IntArray()}
	attrs["float"] = &tengo.UserFunction{Name: "float", Value: f.Float()}
	attrs["float_array"] = &tengo.UserFunction{Name: "float_array", Value: f.FloatArray()}
	attrs["string"] = &tengo.UserFunction{Name: "string", Value: f.String()}
	attrs["string_array"] = &tengo.UserFunction{Name: "string_array", Value: f.StringArray()}
	attrs["bool"] = &tengo.UserFunction{Name: "bool", Value: f.Bool()}
	attrs["parse"] = &tengo.UserFunction{Name: "parse", Value: f.Parse()}
	attrs["get_int"] = &tengo.UserFunction{Name: "get_int", Value: f.GetInt()}
	attrs["get_int_array"] = &tengo.UserFunction{Name: "get_int_array", Value: f.GetIntArray()}
	attrs["get_float"] = &tengo.UserFunction{Name: "get_float", Value: f.GetFloat()}
	attrs["get_float_array"] = &tengo.UserFunction{Name: "get_float_array", Value: f.GetFloatArray()}
	attrs["get_string"] = &tengo.UserFunction{Name: "get_string", Value: f.GetString()}
	attrs["get_string_array"] = &tengo.UserFunction{Name: "get_string_array", Value: f.GetStringArray()}
	attrs["get_bool"] = &tengo.UserFunction{Name: "get_bool", Value: f.GetBool()}
	f.Attrs = attrs

	return f
}

// Import returns an immutable map for the module.
func (f *ImportFlag) Import(moduleName string) (interface{}, error) {
	return f.AsImmutableMap(moduleName), nil
}

func (f *ImportFlag) Version() string {
	return "v1.0.0"
}

// AsImmutableMap converts builtin module into an immutable map.
func (f *ImportFlag) AsImmutableMap(name string) *tengo.ImmutableMap {
	attrs := make(map[string]tengo.Object, len(f.Attrs))
	for k, v := range f.Attrs {
		attrs[k] = v.Copy()
	}
	attrs["__module_name__"] = &tengo.String{Value: name}
	return &tengo.ImmutableMap{Value: attrs}
}

func (f *ImportFlag) Parse() tengo.CallableFunc {
	return func(args ...tengo.Object) (tengo.Object, error) {
		if len(args) != 0 {
			return nil, tengo.ErrWrongNumArguments
		}
		err := f.fs.Parse(os.Args)
		if err != nil {
			return nil, err
		}
		return tengo.UndefinedValue, nil
	}
}

func (f *ImportFlag) Int() tengo.CallableFunc {
	return func(args ...tengo.Object) (tengo.Object, error) {
		numArgs := len(args)
		if numArgs != 3 {
			return nil, errors.Wrap(tengo.ErrWrongNumArguments, "length != 3")
		}

		name, ok := args[0].(*tengo.String)
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "name",
				Expected: "string",
				Found:    args[0].TypeName(),
			}
		}

		value, ok := args[1].Copy().(*tengo.Int)
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "value",
				Expected: "int",
				Found:    args[1].TypeName(),
			}
		}

		usage, ok := args[2].(*tengo.String)
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "usage",
				Expected: "string",
				Found:    args[2].TypeName(),
			}
		}

		f.fs.Int64Var(&value.Value, name.Value, value.Value, usage.Value)
		return value, nil
	}
}

func (f *ImportFlag) IntArray() tengo.CallableFunc {
	return func(args ...tengo.Object) (tengo.Object, error) {
		numArgs := len(args)
		if numArgs != 3 {
			return nil, errors.Wrap(tengo.ErrWrongNumArguments, "length != 3")
		}

		name, ok := args[0].(*tengo.String)
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "name",
				Expected: "string",
				Found:    args[0].TypeName(),
			}
		}

		value, ok := args[1].Copy().(*tengo.Array)
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "value",
				Expected: "[]int",
				Found:    args[1].TypeName(),
			}
		}

		usage, ok := args[2].(*tengo.String)
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "usage",
				Expected: "string",
				Found:    args[2].TypeName(),
			}
		}

		arr := make([]string, 0)
		for _, item := range value.Value {
			s, _ := item.(*tengo.String)
			if s != nil {
				arr = append(arr, s.Value)
			}
		}

		f.fs.Var(&IntSlice{Value: value}, name.Value, usage.Value)
		return value, nil
	}
}

func (f *ImportFlag) Float() tengo.CallableFunc {
	return func(args ...tengo.Object) (tengo.Object, error) {
		numArgs := len(args)
		if numArgs != 3 {
			return nil, errors.Wrap(tengo.ErrWrongNumArguments, "length != 3")
		}

		name, ok := args[0].(*tengo.String)
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "name",
				Expected: "string",
				Found:    args[0].TypeName(),
			}
		}

		value, ok := args[1].Copy().(*tengo.Float)
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "value",
				Expected: "int",
				Found:    args[1].TypeName(),
			}
		}

		usage, ok := args[2].(*tengo.String)
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "usage",
				Expected: "string",
				Found:    args[2].TypeName(),
			}
		}

		f.fs.Float64Var(&value.Value, name.Value, value.Value, usage.Value)
		return value, nil
	}
}

func (f *ImportFlag) FloatArray() tengo.CallableFunc {
	return func(args ...tengo.Object) (tengo.Object, error) {
		numArgs := len(args)
		if numArgs != 3 {
			return nil, errors.Wrap(tengo.ErrWrongNumArguments, "length != 3")
		}

		name, ok := args[0].(*tengo.String)
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "name",
				Expected: "string",
				Found:    args[0].TypeName(),
			}
		}

		value, ok := args[1].Copy().(*tengo.Array)
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "value",
				Expected: "[]float",
				Found:    args[1].TypeName(),
			}
		}

		usage, ok := args[2].(*tengo.String)
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "usage",
				Expected: "string",
				Found:    args[2].TypeName(),
			}
		}

		arr := make([]string, 0)
		for _, item := range value.Value {
			s, _ := item.(*tengo.String)
			if s != nil {
				arr = append(arr, s.Value)
			}
		}

		f.fs.Var(&FloatSlice{Value: value}, name.Value, usage.Value)
		return value, nil
	}
}

func (f *ImportFlag) String() tengo.CallableFunc {
	return func(args ...tengo.Object) (tengo.Object, error) {
		numArgs := len(args)
		if numArgs != 3 {
			return nil, errors.Wrap(tengo.ErrWrongNumArguments, "length != 3")
		}

		name, ok := args[0].(*tengo.String)
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "name",
				Expected: "string",
				Found:    args[0].TypeName(),
			}
		}

		value, ok := args[1].Copy().(*tengo.String)
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "value",
				Expected: "string",
				Found:    args[1].TypeName(),
			}
		}

		usage, ok := args[2].(*tengo.String)
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "usage",
				Expected: "string",
				Found:    args[2].TypeName(),
			}
		}

		f.fs.StringVar(&value.Value, name.Value, value.Value, usage.Value)
		return value, nil
	}
}

func (f *ImportFlag) StringArray() tengo.CallableFunc {
	return func(args ...tengo.Object) (tengo.Object, error) {
		numArgs := len(args)
		if numArgs != 3 {
			return nil, errors.Wrap(tengo.ErrWrongNumArguments, "length != 3")
		}

		name, ok := args[0].(*tengo.String)
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "name",
				Expected: "string",
				Found:    args[0].TypeName(),
			}
		}

		value, ok := args[1].Copy().(*tengo.Array)
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "value",
				Expected: "[]string",
				Found:    args[1].TypeName(),
			}
		}

		usage, ok := args[2].(*tengo.String)
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "usage",
				Expected: "string",
				Found:    args[2].TypeName(),
			}
		}

		arr := make([]string, 0)
		for _, item := range value.Value {
			s, _ := item.(*tengo.String)
			if s != nil {
				arr = append(arr, s.Value)
			}
		}

		f.fs.Var(&StringSlice{Value: value}, name.Value, usage.Value)
		return value, nil
	}
}

func (f *ImportFlag) Bool() tengo.CallableFunc {
	return func(args ...tengo.Object) (tengo.Object, error) {
		numArgs := len(args)
		if numArgs != 3 {
			return nil, errors.Wrap(tengo.ErrWrongNumArguments, "length != 3")
		}

		name, ok := args[0].(*tengo.String)
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "name",
				Expected: "string",
				Found:    args[0].TypeName(),
			}
		}

		value, ok := args[1].Copy().(*tengo.Bool)
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "value",
				Expected: "bool",
				Found:    args[1].TypeName(),
			}
		}

		usage, ok := args[2].(*tengo.String)
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "usage",
				Expected: "string",
				Found:    args[2].TypeName(),
			}
		}

		p := &extra.BoolP{Value: !value.IsFalsy()}
		f.fs.Var(p, name.Value, usage.Value)
		return p, nil
	}
}

func (f *ImportFlag) GetInt() tengo.CallableFunc {
	return func(args ...tengo.Object) (tengo.Object, error) {
		numArgs := len(args)
		if numArgs == 0 {
			return nil, tengo.ErrWrongNumArguments
		}

		name, ok := args[0].(*tengo.String)
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "name",
				Expected: "int",
				Found:    args[0].TypeName(),
			}
		}

		n, err := f.fs.GetInt64(name.Value)
		if err != nil {
			return nil, err
		}
		return &tengo.Int{Value: n}, nil
	}
}

func (f *ImportFlag) GetIntArray() tengo.CallableFunc {
	return func(args ...tengo.Object) (tengo.Object, error) {
		numArgs := len(args)
		if numArgs == 0 {
			return nil, tengo.ErrWrongNumArguments
		}

		name, ok := args[0].(*tengo.String)
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "name",
				Expected: "string",
				Found:    args[0].TypeName(),
			}
		}

		pf := f.fs.Lookup(name.Value)
		if pf == nil {
			return nil, errors.Wrapf(ErrUndefined, "'%s'", name.Value)
		}
		p, ok := pf.Value.(*IntSlice)
		if !ok {
			return nil, errors.Wrapf(ErrFlagType, "actual %s", pf.Value.Type())
		}

		return p.Value, nil
	}
}

func (f *ImportFlag) GetFloat() tengo.CallableFunc {
	return func(args ...tengo.Object) (tengo.Object, error) {
		numArgs := len(args)
		if numArgs == 0 {
			return nil, tengo.ErrWrongNumArguments
		}

		name, ok := args[0].(*tengo.String)
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "name",
				Expected: "float",
				Found:    args[0].TypeName(),
			}
		}

		n, err := f.fs.GetFloat64(name.Value)
		if err != nil {
			return nil, err
		}
		return &tengo.Float{Value: n}, nil
	}
}

func (f *ImportFlag) GetFloatArray() tengo.CallableFunc {
	return func(args ...tengo.Object) (tengo.Object, error) {
		numArgs := len(args)
		if numArgs == 0 {
			return nil, tengo.ErrWrongNumArguments
		}

		name, ok := args[0].(*tengo.String)
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "name",
				Expected: "string",
				Found:    args[0].TypeName(),
			}
		}

		pf := f.fs.Lookup(name.Value)
		if pf == nil {
			return nil, errors.Wrapf(ErrUndefined, "'%s'", name.Value)
		}
		p, ok := pf.Value.(*FloatSlice)
		if !ok {
			return nil, errors.Wrapf(ErrFlagType, "actual %s", pf.Value.Type())
		}

		return p.Value, nil
	}
}

func (f *ImportFlag) GetString() tengo.CallableFunc {
	return func(args ...tengo.Object) (tengo.Object, error) {
		numArgs := len(args)
		if numArgs == 0 {
			return nil, tengo.ErrWrongNumArguments
		}

		name, ok := args[0].(*tengo.String)
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "name",
				Expected: "string",
				Found:    args[0].TypeName(),
			}
		}

		s, err := f.fs.GetString(name.Value)
		if err != nil {
			return nil, err
		}
		return &tengo.String{Value: s}, nil
	}
}

func (f *ImportFlag) GetStringArray() tengo.CallableFunc {
	return func(args ...tengo.Object) (tengo.Object, error) {
		numArgs := len(args)
		if numArgs == 0 {
			return nil, tengo.ErrWrongNumArguments
		}

		name, ok := args[0].(*tengo.String)
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "name",
				Expected: "string",
				Found:    args[0].TypeName(),
			}
		}

		pf := f.fs.Lookup(name.Value)
		if pf == nil {
			return nil, errors.Wrapf(ErrUndefined, "'%s'", name.Value)
		}
		p, ok := pf.Value.(*StringSlice)
		if !ok {
			return nil, errors.Wrapf(ErrFlagType, "actual %s", pf.Value.Type())
		}

		return p.Value, nil
	}
}

func (f *ImportFlag) GetBool() tengo.CallableFunc {
	return func(args ...tengo.Object) (tengo.Object, error) {
		numArgs := len(args)
		if numArgs == 0 {
			return nil, tengo.ErrWrongNumArguments
		}

		name, ok := args[0].(*tengo.String)
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "name",
				Expected: "string",
				Found:    args[0].TypeName(),
			}
		}

		pf := f.fs.Lookup(name.Value)
		if pf == nil {
			return nil, errors.Wrapf(ErrUndefined, "'%s'", name.Value)
		}
		p, ok := pf.Value.(*extra.BoolP)
		if !ok {
			return nil, errors.Wrapf(ErrFlagType, "actual %s", pf.Value.Type())
		}
		return p, nil
	}
}

type IntSlice struct {
	Value *tengo.Array
}

func (s *IntSlice) String() string {
	arr := make([]string, 0)
	for _, item := range s.Value.Value {
		val, _ := item.(*tengo.Int)
		if val != nil {
			arr = append(arr, fmt.Sprintf("%d", val.Value))
		}
	}
	return strings.Join(arr, ",")
}

func (s *IntSlice) Set(text string) error {
	parts := strings.Split(text, ",")
	s.Value.Value = s.Value.Value[:len(parts)]
	for i, item := range parts {
		n, _ := strconv.ParseInt(strings.TrimSpace(item), 10, 64)
		s.Value.Value[i] = &tengo.Int{Value: n}
	}
	return nil
}

func (s *IntSlice) Type() string {
	return "[]int"
}

type FloatSlice struct {
	Value *tengo.Array
}

func (s *FloatSlice) String() string {
	arr := make([]string, 0)
	for _, item := range s.Value.Value {
		val, _ := item.(*tengo.Float)
		if val != nil {
			arr = append(arr, fmt.Sprintf("%f", val.Value))
		}
	}
	return strings.Join(arr, ",")
}

func (s *FloatSlice) Set(text string) error {
	parts := strings.Split(text, ",")
	s.Value.Value = s.Value.Value[:len(parts)]
	for i, item := range parts {
		n, _ := strconv.ParseFloat(strings.TrimSpace(item), 64)
		s.Value.Value[i] = &tengo.Float{Value: n}
	}
	return nil
}

func (s *FloatSlice) Type() string {
	return "[]float"
}

type StringSlice struct {
	Value *tengo.Array
}

func (s *StringSlice) String() string {
	arr := make([]string, 0)
	for _, item := range s.Value.Value {
		val, _ := item.(*tengo.String)
		if val != nil {
			arr = append(arr, val.Value)
		}
	}
	return strings.Join(arr, ",")
}

func (s *StringSlice) Set(text string) error {
	parts := strings.Split(text, ",")
	s.Value.Value = s.Value.Value[:len(parts)]
	for i, item := range parts {
		item = strings.TrimSpace(item)
		s.Value.Value[i] = &tengo.String{Value: item}
	}
	return nil
}

func (s *StringSlice) Type() string {
	return "[]string"
}
