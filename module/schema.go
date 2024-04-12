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

package module

import (
	"fmt"
	"strconv"
	"time"
)

type Schema struct {
	Name    string       `json:"name,omitempty" yaml:"name,omitempty"`
	Type    string       `json:"type,omitempty" yaml:"type,omitempty"`
	Short   string       `json:"short,omitempty" yaml:"short,omitempty"`
	Desc    string       `json:"desc,omitempty" yaml:"desc,omitempty"`
	Default string       `json:"default,omitempty" yaml:"default,omitempty"`
	Example string       `json:"example,omitempty" yaml:"example,omitempty"`
	Value   *SchemaValue `json:"-" yaml:"-"`
}

func (s *Schema) InitValue() *SchemaValue {
	if s.Type == "" {
		s.Type = "string"
	}
	sv := &SchemaValue{}
	switch s.Type {
	case "string":
		sv.StringP = new(string)
	case "int", "int32", "int64":
		sv.IntP = new(int64)
	case "uint", "uint32", "uint64":
		sv.UintP = new(uint64)
	case "float", "float32", "float64":
		sv.FloatP = new(float64)
	case "duration":
		sv.DurationP = new(time.Duration)
	}

	s.Value = sv
	return sv
}

type SchemaValue struct {
	IntP      *int64
	UintP     *uint64
	StringP   *string
	FloatP    *float64
	DurationP *time.Duration
}

func (sv *SchemaValue) String() string {
	if sv.IntP != nil {
		return fmt.Sprintf("%d", *sv.IntP)
	}
	if sv.UintP != nil {
		return fmt.Sprintf("%d", &sv.UintP)
	}
	if sv.StringP != nil {
		return *sv.StringP
	}
	if sv.FloatP != nil {
		return fmt.Sprintf("%f", *sv.FloatP)
	}
	if sv.DurationP != nil {
		return sv.DurationP.String()
	}
	return ""
}

func (sv *SchemaValue) Set(text string) error {
	if sv.IntP != nil {
		i, err := strconv.ParseInt(text, 10, 64)
		if err != nil {
			return err
		}
		sv.IntP = &i
		return nil
	}
	if sv.UintP != nil {
		i, err := strconv.ParseUint(text, 10, 64)
		if err != nil {
			return err
		}
		sv.UintP = &i
		return nil
	}
	if sv.StringP != nil {
		sv.StringP = &text
		return nil
	}
	if sv.FloatP != nil {
		f, err := strconv.ParseFloat(text, 64)
		if err != nil {
			return err
		}
		sv.FloatP = &f
		return nil
	}
	if sv.DurationP != nil {
		d, err := time.ParseDuration(text)
		if err != nil {
			return err
		}
		sv.DurationP = &d
		return nil
	}
	return nil
}

func (sv *SchemaValue) Type() string {
	if sv.IntP != nil {
		return "int64"
	}
	if sv.UintP != nil {
		return "uint64"
	}
	if sv.StringP != nil {
		return "string"
	}
	if sv.FloatP != nil {
		return "float64"
	}
	if sv.DurationP != nil {
		return "duration"
	}
	return "string"
}
