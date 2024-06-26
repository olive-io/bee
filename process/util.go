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

package process

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"

	json "github.com/json-iterator/go"
	"github.com/muyo/sno"
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

func parseTaskArgs(s string) map[string]any {
	if strings.TrimSpace(s) == "" {
		return map[string]any{}
	}
	parts := strings.Split(s, " ")
	args := make(map[string]any)
	for _, part := range parts {
		be, af, ok := strings.Cut(part, "=")
		if !ok {
			continue
		}
		args[be] = af
	}
	return args
}

func EncodeScriptTask(task *Task) (props map[string]any, headers map[string]any) {
	props = map[string]any{}
	headers = map[string]any{}

	token, _ := json.Marshal(task)
	props["token"] = string(token)
	headers["hosts"] = strings.Join(task.Hosts, ",")
	headers["id"] = task.Id
	headers["name"] = task.Name
	headers["desc"] = task.Desc

	return
}

func DecodeScriptTask(props, headers map[string]any) *Task {
	var task *Task
	if v, ok := props["token"]; ok {
		vv, _ := v.(string)
		_ = json.Unmarshal([]byte(vv), &task)
	}

	return task
}

func EncodeServiceTask(service *Service) (props map[string]any, headers map[string]any) {
	props = map[string]any{}
	headers = map[string]any{}

	token, _ := json.Marshal(service)
	props["token"] = string(token)
	headers["hosts"] = strings.Join(service.Hosts, ",")
	headers["id"] = service.Id
	headers["name"] = service.Name
	headers["desc"] = service.Desc

	return
}

func DecodeServiceTask(props, headers map[string]any) *Service {
	var s *Service
	if v, ok := props["token"]; ok {
		vv, _ := v.(string)
		_ = json.Unmarshal([]byte(vv), &s)
	}

	return s
}

func newSnoId() string {
	return string(sno.New(0).Bytes())
}
