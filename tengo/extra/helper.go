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

package extra

import (
	"github.com/d5/tengo/v2"
)

// FuncARE transform a function of 'func() error' signature into CallableFunc
// type.
func FuncARE(fn func() error) tengo.CallableFunc {
	return func(args ...tengo.Object) (ret tengo.Object, err error) {
		if len(args) != 0 {
			return nil, tengo.ErrWrongNumArguments
		}
		return wrapError(fn()), nil
	}
}

// FuncBYTEE transform a function of 'func() ([]byte, error)' signature into
// CallableFunc type.
func FuncBYTEE(fn func() ([]byte, error)) tengo.CallableFunc {
	return func(args ...tengo.Object) (ret tengo.Object, err error) {
		if len(args) != 0 {
			return nil, tengo.ErrWrongNumArguments
		}
		res, err := fn()
		if err != nil {
			if len(res) == 0 {
				res = []byte(err.Error())
			}
			return wrapResError(err, res), nil
		}
		if len(res) > tengo.MaxBytesLen {
			return nil, tengo.ErrBytesLimit
		}
		return &tengo.Bytes{Value: res}, nil
	}
}

func wrapError(err error) tengo.Object {
	if err == nil {
		return tengo.TrueValue
	}
	return &tengo.Error{Value: &tengo.String{Value: err.Error()}}
}

func wrapResError(err error, output []byte) tengo.Object {
	if err == nil {
		return tengo.TrueValue
	}
	return &tengo.Error{Value: &tengo.Bytes{Value: output}}
}
