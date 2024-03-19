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
