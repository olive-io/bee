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
	"context"
	"fmt"
	"os"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/d5/tengo/v2"

	"github.com/olive-io/bee/tengo/builtin/trace/internal"
)

var (
	Importable tengo.Importable = NewTrace()
)

const defaultLevel = internal.LevelPrint

type ImportModule struct {
	tengo.ObjectImpl

	level   internal.Level
	handler *traceHandler
	lg      *internal.Logger
	Attrs   map[string]tengo.Object
	fields  []*traceField
}

func NewTrace() *ImportModule {
	jsonHandler := internal.NewJSONHandler(os.Stdout, &internal.HandlerOptions{
		AddSource: false,
		Level:     defaultLevel,
		ReplaceAttr: func(groups []string, a internal.Attr) internal.Attr {
			switch a.Key {
			case internal.TimeKey, internal.LevelKey:
				return internal.Attr{}
			}
			return a
		},
	})

	handler := newTraceHandler(defaultLevel, jsonHandler)
	lg := internal.New(handler)
	tm := &ImportModule{
		handler: handler,
		lg:      lg,
		fields:  []*traceField{},
	}
	attrs := map[string]tengo.Object{}
	attrs["add_handler"] = &tengo.UserFunction{Name: "add_handler", Value: tm.AddHandler()}
	attrs["add_hook"] = &tengo.UserFunction{Name: "add_hook", Value: tm.AddHook()}
	attrs["set_level"] = &tengo.UserFunction{Name: "set_level", Value: tm.SetLevel()}
	attrs["int"] = &tengo.UserFunction{Name: "int", Value: tm.IntField()}
	attrs["float"] = &tengo.UserFunction{Name: "float", Value: tm.FloatField()}
	attrs["string"] = &tengo.UserFunction{Name: "string", Value: tm.StringField()}
	attrs["bool"] = &tengo.UserFunction{Name: "bool", Value: tm.BoolField()}
	attrs["duration"] = &tengo.UserFunction{Name: "duration", Value: tm.DurationField()}
	attrs["time"] = &tengo.UserFunction{Name: "time", Value: tm.TimeField()}
	attrs["fields"] = &tengo.UserFunction{Name: "fields", Value: tm.Fields()}
	attrs["debug"] = &tengo.UserFunction{Name: "debug", Value: tm.Debug()}
	attrs["info"] = &tengo.UserFunction{Name: "info", Value: tm.Info()}
	attrs["warn"] = &tengo.UserFunction{Name: "warn", Value: tm.Warn()}
	attrs["errorf"] = &tengo.UserFunction{Name: "errorf", Value: tm.Error()}
	attrs["print"] = &tengo.UserFunction{Name: "print", Value: tm.Print()}
	attrs["try"] = &tengo.UserFunction{Name: "try", Value: tm.Try()}
	attrs["assert"] = &tengo.UserFunction{Name: "assert", Value: tm.Assert()}
	tm.Attrs = attrs
	return tm
}

// Import returns an immutable map for the module.
func (m *ImportModule) Import(name string) (interface{}, error) {
	return m.AsImmutableMap(name), nil
}

func (m *ImportModule) Version() string {
	return "v1.0.0"
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

func (m *ImportModule) SetLevel() tengo.CallableFunc {
	return func(args ...tengo.Object) (ret tengo.Object, err error) {
		numArgs := len(args)
		if numArgs != 1 {
			return nil, errors.Wrap(tengo.ErrWrongNumArguments, "missing level")
		}

		levelStr, ok := args[0].(*tengo.String)
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "level",
				Expected: "string",
				Found:    args[1].TypeName(),
			}
		}
		level, ok := parseLevel(levelStr.Value)
		if !ok {
			level = defaultLevel
		}
		m.level = level
		m.handler.setLevel(level)

		return tengo.UndefinedValue, nil
	}
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

		attr := internal.Int64(name.Value, value.Value)
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

		attr := internal.Float64(name.Value, value.Value)
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

		attr := internal.String(name.Value, value.Value)
		return &traceField{Value: attr}, nil
	}
}

func (m *ImportModule) BoolField() tengo.CallableFunc {
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

		value, ok := args[1].(*tengo.Bool)
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "value",
				Expected: "boolean",
				Found:    args[1].TypeName(),
			}
		}

		attr := internal.Bool(name.Value, !value.IsFalsy())
		return &traceField{Value: attr}, nil
	}
}

func (m *ImportModule) DurationField() tengo.CallableFunc {
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

		var t int64
		switch tv := args[1].(type) {
		case *tengo.Int:
			t = tv.Value
		case *tengo.Time:
			t = tv.Value.UnixNano()
		default:
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "value",
				Expected: "int|time",
				Found:    args[1].TypeName(),
			}
		}

		attr := internal.Duration(name.Value, time.Duration(t))
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

		attr := internal.Time(name.Value, value.Value)
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

		fields = append(m.fields, fields...)
		m.fields = fields

		tm := &ImportModule{
			level:   m.level,
			handler: m.handler,
			lg:      internal.New(m.handler),
			Attrs:   m.Attrs,
			fields:  fields,
		}
		return tm, nil
	}
}

func (m *ImportModule) log(level internal.Level, args ...tengo.Object) (ret tengo.Object, err error) {
	numArgs := len(args)
	if numArgs == 0 {
		return nil, tengo.ErrWrongNumArguments
	}

	ctx := context.TODO()
	attrs := make([]any, 0)
	for _, attr := range m.fields {
		attrs = append(attrs, attr.Value)
	}

	if numArgs == 1 {
		s := fmt.Sprintf("%v", args[0].String())
		m.lg.Log(ctx, level, unquote(s), attrs...)
		return tengo.UndefinedValue, nil
	}

	format, ok := args[0].(*tengo.String)
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "format",
			Expected: "string",
			Found:    args[0].TypeName(),
		}
	}
	s, err := tengo.Format(format.Value, args[1:]...)
	if err != nil {
		return nil, err
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
	lg := internal.New(m.handler)
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
		return m.log(internal.LevelDebug, args...)
	}
}

func (m *ImportModule) Info() tengo.CallableFunc {
	return func(args ...tengo.Object) (ret tengo.Object, err error) {
		return m.log(internal.LevelInfo, args...)
	}
}

func (m *ImportModule) Warn() tengo.CallableFunc {
	return func(args ...tengo.Object) (ret tengo.Object, err error) {
		return m.log(internal.LevelWarn, args...)
	}
}

func (m *ImportModule) Error() tengo.CallableFunc {
	return func(args ...tengo.Object) (ret tengo.Object, err error) {
		return m.log(internal.LevelError, args...)
	}
}

func (m *ImportModule) Print() tengo.CallableFunc {
	return func(args ...tengo.Object) (ret tengo.Object, err error) {
		return m.log(internal.LevelPrint, args...)
	}
}

func (m *ImportModule) Try() tengo.CallableFunc {
	return func(args ...tengo.Object) (ret tengo.Object, err error) {
		numArgs := len(args)
		if numArgs == 0 {
			return nil, errors.Wrap(tengo.ErrWrongNumArguments, "args must great than 0")
		}

		tErr, ok := args[0].(*tengo.Error)
		if !ok {
			return tengo.UndefinedValue, nil
		}

		attrs := make([]any, 0)
		s := fmt.Sprintf("%v", tErr.Value.String())
		attrs = append(attrs, internal.String("error", unquote(s)))

		msg := "occurred error"
		if len(args) > 1 {
			value, ok := args[1].(*tengo.String)
			if ok {
				msg = unquote(value.Value)
			}

			for _, arg := range args[1:] {
				if field, ok := arg.(*traceField); ok {
					attrs = append(attrs, field.Value)
				}
			}
		}

		m.lg.Log(context.TODO(), internal.LevelPrint, msg, attrs...)
		os.Exit(1)
		return tengo.UndefinedValue, nil
	}
}

func (m *ImportModule) Assert() tengo.CallableFunc {
	return func(args ...tengo.Object) (ret tengo.Object, err error) {
		numArgs := len(args)
		if numArgs < 2 {
			return nil, errors.Wrap(tengo.ErrWrongNumArguments, "args must grant than 1")
		}

		a := args[0]
		b := args[1]
		if a.String() == b.String() {
			return tengo.UndefinedValue, nil
		}

		text := fmt.Sprintf(`got %v, expected %v`, a.String(), b.String())
		attr := internal.String("error", text)

		attrs := make([]any, 0)
		attrs = append(attrs, attr)

		msg := "assert error"
		if len(args) > 2 {
			value, ok := args[2].(*tengo.String)
			if ok {
				msg = unquote(value.Value)
			}

			for _, arg := range args[2:] {
				if field, ok := arg.(*traceField); ok {
					attrs = append(attrs, field.Value)
				}
			}
		}

		m.lg.Log(context.TODO(), internal.LevelPrint, msg, attrs...)
		os.Exit(1)

		return tengo.UndefinedValue, nil
	}
}

type traceField struct {
	tengo.ObjectImpl
	Value internal.Attr
}

func (tf *traceField) TypeName() string {
	return "traceField"
}

func (tf *traceField) Copy() tengo.Object {
	attr := internal.Attr{Key: tf.Value.Key, Value: tf.Value.Value}
	return &traceField{Value: attr}
}

func (tf *traceField) String() string {
	return tf.Value.Key + "=" + tf.Value.Value.String()
}
