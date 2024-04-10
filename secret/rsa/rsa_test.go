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

package rsa

import (
	"encoding/hex"
	"testing"
)

func TestEncode(t *testing.T) {
	source := []byte("hello world")

	out, err := Encode([]byte(source))
	if err != nil {
		t.Fatal(err)
	}

	orig, err := Decode(out)
	if err != nil {
		t.Fatal(err)
	}

	if string(source) != string(orig) {
		t.Log()
	}

	t.Log(hex.EncodeToString(out))
}
