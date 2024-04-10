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

type BoolP struct {
	tengo.ObjectImpl

	Value bool
}

func (b *BoolP) TypeName() string {
	return "bool"
}

func (b *BoolP) Copy() tengo.Object {
	return &BoolP{Value: b.Value}
}

func (b *BoolP) String() string {
	if b.Value {
		return "true"
	}
	return "false"
}

func (b *BoolP) Set(text string) error {
	if text == "true" || text == "1" {
		b.Value = true
	} else {
		b.Value = false
	}
	return nil
}

func (b *BoolP) Type() string {
	return "bool"
}
