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

package executor

import "reflect"

type Zone int

func (z Zone) String() string {
	s := zoneS[z]
	return s
}

const (
	Processed Zone = iota + 1
	Ok
	Dark
	Changed
	Skipped
	Rescued
	Ignored
)

var zoneS = map[Zone]string{
	Processed: "processed",
	Ok:        "ok",
	Dark:      "dark",
	Changed:   "changed",
	Skipped:   "skipped",
	Rescued:   "rescued",
	Ignored:   "ignored",
}

type AggregateStats struct {
	aggregates map[Zone]map[string]int64
	custom     map[string]map[string]any
}

func NewStats() *AggregateStats {
	as := &AggregateStats{
		aggregates: map[Zone]map[string]int64{},
		custom:     map[string]map[string]any{},
	}
	for zone, _ := range zoneS {
		as.aggregates[zone] = map[string]int64{}
	}
	return as
}

func (as *AggregateStats) Increment(what Zone, host string) {
	stats, ok := as.aggregates[what]
	if !ok {
		as.aggregates[what] = map[string]int64{}
		stats = as.aggregates[what]
	}
	_, ok = stats[host]
	if !ok {
		stats[host] = 0
	}
	stats[host] += 1
}

func (as *AggregateStats) Decrement(what Zone, host string) {
	stats, ok := as.aggregates[what]
	if !ok {
		as.aggregates[what] = map[string]int64{}
		stats = as.aggregates[what]
	}
	_, ok = stats[host]
	if !ok {
		stats[host] = 0
	}
	if stats[host] > 1 {
		stats[host] -= 1
	}
}

func (as *AggregateStats) Summarize(host string) map[string]int64 {
	stat := map[string]int64{}
	for zone, stats := range as.aggregates {
		stats[zone.String()] = stats[host]
	}
	return stat
}

func (as *AggregateStats) GetCustomStats(host, which string) (any, bool) {
	customs, ok := as.custom[host]
	if !ok {
		return nil, false
	}
	value, ok := customs[which]
	return value, ok
}

func (as *AggregateStats) SetCustomStats(which string, what any, host string) {
	if host == "" {
		host = "_run"
	}
	customs, ok := as.custom[host]
	if !ok {
		as.custom[host] = map[string]any{which: what}
		return
	}
	customs[which] = what
}

func (as *AggregateStats) UpdateCustomStats(which string, what any, host string) {
	if host == "" {
		host = "_run"
	}
	customs, ok := as.custom[host]
	if !ok {
		as.SetCustomStats(which, what, host)
		return
	}
	value, ok := customs[which]
	if !ok {
		as.SetCustomStats(which, what, host)
		return
	}

	vt := reflect.ValueOf(value)
	if vt.Type().Name() != reflect.TypeOf(what).Name() {
		return
	}

	wt := reflect.ValueOf(what)
	if vt.Kind() == reflect.Map {
		iter := wt.MapRange()
		for iter.Next() {
			key, iv := iter.Key(), iter.Value()
			vt.SetMapIndex(key, iv)
		}

		return
	}

	if vt.CanInt() {
		i := vt.Int() + wt.Int()
		as.custom[host][which] = i
	} else if vt.CanUint() {
		i := vt.Uint() + wt.Uint()
		as.custom[host][which] = i
	} else if vt.CanFloat() {
		f := vt.Float() + wt.Float()
		as.custom[host][which] = f
	} else if vt.String() != "" {
		s := vt.String() + wt.String()
		as.custom[host][which] = s
	}
}
