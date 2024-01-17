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

package trace

import (
	"context"
	"log/slog"
	"os"

	"github.com/cockroachdb/errors"
	"github.com/d5/tengo/v2"
)

var (
	Importable tengo.Importable = NewTrace()
)

const defaultLevel = slog.LevelDebug

type ImportModule struct {
	tengo.ObjectImpl

	level   slog.Level
	handler *traceHandler
	lg      *slog.Logger
	Attrs   map[string]tengo.Object
	fields  []*traceField
}

func NewTrace() *ImportModule {
	jsonHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: false,
		Level:     defaultLevel,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			switch a.Key {
			case slog.TimeKey, slog.LevelKey:
				return slog.Attr{}
			}
			return a
		},
	})

	handler := newTraceHandler(defaultLevel, jsonHandler)
	lg := slog.New(handler)
	tm := &ImportModule{
		handler: handler,
		lg:      lg,
		fields:  []*traceField{},
	}
	attrs := map[string]tengo.Object{}
	attrs["add_handler"] = &tengo.UserFunction{Name: "add_handler", Value: tm.AddHandler()}
	attrs["add_hook"] = &tengo.UserFunction{Name: "add_hook", Value: tm.AddHook()}
	attrs["int"] = &tengo.UserFunction{Name: "int", Value: tm.IntField()}
	attrs["float"] = &tengo.UserFunction{Name: "float", Value: tm.FloatField()}
	attrs["string"] = &tengo.UserFunction{Name: "string", Value: tm.StringField()}
	attrs["time"] = &tengo.UserFunction{Name: "time", Value: tm.TimeField()}
	attrs["fields"] = &tengo.UserFunction{Name: "fields", Value: tm.Fields()}
	attrs["debug"] = &tengo.UserFunction{Name: "debug", Value: tm.Debug()}
	attrs["info"] = &tengo.UserFunction{Name: "info", Value: tm.Info()}
	attrs["warn"] = &tengo.UserFunction{Name: "warn", Value: tm.Warn()}
	attrs["error"] = &tengo.UserFunction{Name: "error", Value: tm.Error()}
	tm.Attrs = attrs
	return tm
}

// Import returns an immutable map for the module.
func (m *ImportModule) Import(name string) (interface{}, error) {
	return m.AsImmutableMap(name), nil
}

// AsImmutableMap converts builtin module into an immutable map.
func (m *ImportModule) AsImmutableMap(name string) *tengo.ImmutableMap {
	attrs := make(map[string]tengo.Object, len(m.Attrs))
	for k, v := range m.Attrs {
		attrs[k] = v.Copy()
	}
	attrs["__module_name__"] = &tengo.String{Value: name}
	return &tengo.ImmutableMap{Value: attrs}
}

func (m *ImportModule) IntField() tengo.CallableFunc {
	return func(args ...tengo.Object) (ret tengo.Object, err error) {
		numArgs := len(args)
		if numArgs != 2 {
			return nil, errors.Wrap(tengo.ErrWrongNumArguments, "length != 2")
		}

		name, ok := args[0].(*tengo.String)
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "name",
				Expected: "string",
				Found:    args[0].TypeName(),
			}
		}

		value, ok := args[1].(*tengo.Int)
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "value",
				Expected: "string",
				Found:    args[1].TypeName(),
			}
		}

		attr := slog.Int64(name.Value, value.Value)
		return &traceField{Value: attr}, nil
	}
}

func (m *ImportModule) FloatField() tengo.CallableFunc {
	return func(args ...tengo.Object) (ret tengo.Object, err error) {
		numArgs := len(args)
		if numArgs != 2 {
			return nil, errors.Wrap(tengo.ErrWrongNumArguments, "length != 2")
		}

		name, ok := args[0].(*tengo.String)
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "name",
				Expected: "string",
				Found:    args[0].TypeName(),
			}
		}

		value, ok := args[1].(*tengo.Float)
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "value",
				Expected: "float",
				Found:    args[1].TypeName(),
			}
		}

		attr := slog.Float64(name.Value, value.Value)
		return &traceField{Value: attr}, nil
	}
}

func (m *ImportModule) StringField() tengo.CallableFunc {
	return func(args ...tengo.Object) (ret tengo.Object, err error) {
		numArgs := len(args)
		if numArgs != 2 {
			return nil, errors.Wrap(tengo.ErrWrongNumArguments, "length != 2")
		}

		name, ok := args[0].(*tengo.String)
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "name",
				Expected: "string",
				Found:    args[0].TypeName(),
			}
		}

		value, ok := args[1].(*tengo.String)
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "value",
				Expected: "string",
				Found:    args[1].TypeName(),
			}
		}

		attr := slog.String(name.Value, value.Value)
		return &traceField{Value: attr}, nil
	}
}

func (m *ImportModule) TimeField() tengo.CallableFunc {
	return func(args ...tengo.Object) (ret tengo.Object, err error) {
		numArgs := len(args)
		if numArgs != 2 {
			return nil, errors.Wrap(tengo.ErrWrongNumArguments, "length != 2")
		}

		name, ok := args[0].(*tengo.String)
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "name",
				Expected: "string",
				Found:    args[0].TypeName(),
			}
		}

		value, ok := args[1].(*tengo.Time)
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "value",
				Expected: "time",
				Found:    args[1].TypeName(),
			}
		}

		attr := slog.Time(name.Value, value.Value)
		return &traceField{Value: attr}, nil
	}
}

func (m *ImportModule) Fields() tengo.CallableFunc {
	return func(args ...tengo.Object) (tengo.Object, error) {
		fields := make([]*traceField, 0)

		for _, arg := range args {
			if field, ok := arg.(*traceField); ok {
				fields = append(fields, field)
			}
		}

		tm := &ImportModule{
			handler: m.handler,
			lg:      slog.New(m.handler),
			Attrs:   m.Attrs,
			fields:  fields,
		}
		return tm, nil
	}
}

func (m *ImportModule) log(level slog.Level, args ...tengo.Object) (ret tengo.Object, err error) {
	numArgs := len(args)
	if numArgs == 0 {
		return nil, tengo.ErrWrongNumArguments
	}

	ctx := context.TODO()
	format, ok := args[0].(*tengo.String)
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "format",
			Expected: "string",
			Found:    args[0].TypeName(),
		}
	}
	if numArgs == 1 {
		m.lg.Log(ctx, level, format.Value)
		return tengo.UndefinedValue, nil
	}
	s, err := tengo.Format(format.Value, args[1:]...)
	if err != nil {
		return nil, err
	}

	attrs := make([]any, 0)
	for _, attr := range m.fields {
		attrs = append(attrs, attr.Value)
	}
	m.lg.Log(ctx, level, s, attrs...)
	return tengo.UndefinedValue, nil
}

// TypeName returns the name of the type.
func (m *ImportModule) TypeName() string {
	return "trace-module"
}

func (m *ImportModule) String() string {
	return "<trace-module>"
}

// Copy returns a copy of the type.
func (m *ImportModule) Copy() tengo.Object {
	lg := slog.New(m.handler)
	return &ImportModule{
		handler: m.handler,
		lg:      lg,
		Attrs:   m.Attrs,
		fields:  m.fields,
	}
}

func (m *ImportModule) IndexGet(arg tengo.Object) (tengo.Object, error) {
	name, ok := arg.(*tengo.String)
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "name",
			Expected: "string",
			Found:    arg.TypeName(),
		}
	}

	m.level, ok = parseLevel(name.Value)
	if !ok {
		return nil, errors.Wrapf(tengo.ErrNotImplemented, "called %s", name.Value)
	}
	return m, nil
}

// Equals returns true if the value of the type is equal to the value of
// another object.
func (m *ImportModule) Equals(_ tengo.Object) bool {
	return false
}

func (m *ImportModule) Call(args ...tengo.Object) (tengo.Object, error) {
	return m.log(m.level, args...)
}

func (m *ImportModule) CanCall() bool { return true }

func (m *ImportModule) Debug() tengo.CallableFunc {
	return func(args ...tengo.Object) (ret tengo.Object, err error) {
		return m.log(slog.LevelDebug, args...)
	}
}

func (m *ImportModule) Info() tengo.CallableFunc {
	return func(args ...tengo.Object) (ret tengo.Object, err error) {
		return m.log(slog.LevelInfo, args...)
	}
}

func (m *ImportModule) Warn() tengo.CallableFunc {
	return func(args ...tengo.Object) (ret tengo.Object, err error) {
		return m.log(slog.LevelWarn, args...)
	}
}

func (m *ImportModule) Error() tengo.CallableFunc {
	return func(args ...tengo.Object) (ret tengo.Object, err error) {
		return m.log(slog.LevelError, args...)
	}
}

type traceField struct {
	tengo.ObjectImpl
	Value slog.Attr
}

func (tf *traceField) TypeName() string {
	return "traceField"
}

func (tf *traceField) Copy() tengo.Object {
	attr := slog.Attr{Key: tf.Value.Key, Value: tf.Value.Value}
	return &traceField{Value: attr}
}

func (tf *traceField) String() string {
	return tf.Value.Key + "=" + tf.Value.Value.String()
}
