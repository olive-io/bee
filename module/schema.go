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

package module

import (
	"fmt"
	"strconv"
	"time"
)

type Schema struct {
	Name        string       `yaml:"name"`
	Type        string       `yaml:"type"`
	Short       string       `yaml:"short"`
	Description string       `yaml:"description"`
	Default     string       `yaml:"default"`
	Example     string       `yaml:"example"`
	Value       *SchemaValue `yaml:"-"`
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
