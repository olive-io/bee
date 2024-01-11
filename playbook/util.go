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

package playbook

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"
)

type YamlKV map[string]interface{}

func (kv YamlKV) GetRV(name string) (reflect.Value, bool) {
	value, ok := kv[name]
	if !ok {
		return reflect.Value{}, false
	}
	return reflect.ValueOf(value), ok
}

func (kv YamlKV) Apply(name string, to any) (bool, error) {
	vv, ok := kv.GetRV(name)
	if !ok {
		return false, nil
	}

	_, err := reflectApply(vv, to)
	return true, err
}

func (kv YamlKV) ApplyArray(name string, to any) (bool, error) {
	vv, ok := kv.GetRV(name)
	if !ok {
		return false, nil
	}

	_, err := reflectApplyArray(vv, to)
	return true, err
}

func (kv YamlKV) ApplyMap(name string, to any) (bool, error) {
	vv, ok := kv.GetRV(name)
	if !ok {
		return false, nil
	}

	_, err := reflectApplyMap(vv, to)
	return true, err
}

func reflectApply(vv reflect.Value, to any) (ok bool, err error) {
	defer func() {
		if e := recover(); e != nil {
			_, file, line, _ := runtime.Caller(3)
			err = fmt.Errorf("%v at %s:%d", e, file, line)
		}
	}()
	tv := reflect.ValueOf(to)
	if tv.Kind() != reflect.Pointer {
		return
	}
	tv = tv.Elem()
	if !tv.CanSet() {
		return
	}

	var rv reflect.Value
	if vv.Kind() == tv.Kind() {
		rv = vv
	} else if vv.CanConvert(tv.Type()) {
		rv = vv.Convert(tv.Type())
	} else {
		return
	}
	tv.Set(rv)
	ok = true
	return
}

func reflectApplyArray(vv reflect.Value, to any) (ok bool, err error) {
	defer func() {
		if e := recover(); e != nil {
			_, file, line, _ := runtime.Caller(3)
			err = fmt.Errorf("%v at %s:%d", e, file, line)
		}
	}()
	tv := reflect.ValueOf(to)
	if tv.Kind() != reflect.Pointer {
		return
	}
	tv = tv.Elem()
	if tv.Kind() != reflect.Slice {
		return
	}
	if !tv.CanSet() {
		return
	}

	var rv reflect.Value
	if vv.Kind() == reflect.Slice {
		rs := reflect.MakeSlice(tv.Type(), 0, vv.Cap())
		for i := 0; i < vv.Len(); i++ {
			value := vv.Index(i)
			vi := reflect.New(reflect.TypeOf(vv.Type())).Interface()
			var ok1 bool
			ok1, err = reflectApply(value, &vi)
			if err != nil {
				return
			}
			if ok1 {
				rs = reflect.Append(rs, reflect.ValueOf(vi))
			}
		}
		rv = rs
	} else {
		tp := tv.Type().Elem()
		if vv.Kind() == tp.Kind() {
			rs := reflect.MakeSlice(tv.Type(), 0, 1)
			rv = reflect.Append(rs, vv)
		} else if vv.CanConvert(tp) {
			rs := reflect.MakeSlice(tp, 0, 1)
			rv = reflect.Append(rs, vv.Convert(tp))
		} else {
			return
		}
	}
	tv.Set(rv)
	ok = true
	return
}

func reflectApplyMap(vv reflect.Value, to any) (ok bool, err error) {
	defer func() {
		if e := recover(); e != nil {
			_, file, line, _ := runtime.Caller(3)
			err = fmt.Errorf("%v at %s:%d", e, file, line)
		}
	}()
	if vv.Kind() != reflect.Map {
		return
	}

	tv := reflect.ValueOf(to)
	if tv.Kind() != reflect.Pointer {
		return
	}
	tv = tv.Elem()
	if tv.Kind() != reflect.Map {
		return
	}
	if !tv.CanSet() {
		return
	}
	tkp := tv.Type().Key()
	tvp := tv.Type().Elem()

	iter := vv.MapRange()
	rv := reflect.MakeMapWithSize(tv.Type(), vv.Len())
	for iter.Next() {
		key, value := iter.Key(), iter.Value()
		vi := reflect.New(reflect.TypeOf(tvp)).Interface()
		var ok1 bool
		ok1, err = reflectApply(value, &vi)
		if key.CanConvert(tkp) && ok1 {
			rv.SetMapIndex(key.Convert(tkp), reflect.ValueOf(vi))
		}
	}
	tv.Set(rv)
	ok = true
	return
}

func parseTaskArgs(s string) map[string]string {
	if strings.TrimSpace(s) == "" {
		return map[string]string{}
	}
	parts := strings.Split(s, " ")
	args := make(map[string]string)
	for _, part := range parts {
		be, af, ok := strings.Cut(part, "=")
		if !ok {
			continue
		}
		args[be] = af
	}
	return args
}
