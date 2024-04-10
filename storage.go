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

package bee

import (
	"github.com/cockroachdb/pebble"
	"go.uber.org/zap"
)

const (
	DefaultCacheSize = 1024 * 1024 * 10
)

var defaultSplit pebble.Split = func(a []byte) int {
	return 1
}

func openDB(lg *zap.Logger, dir string) (*pebble.DB, error) {
	if lg == nil {
		lg = zap.NewNop()
	}

	cache := pebble.NewCache(DefaultCacheSize)
	comparer := pebble.DefaultComparer
	comparer.Split = defaultSplit
	lopts := zap.Fields(zap.String("embed-db", "pebble"))
	plg := lg.WithOptions(lopts).Sugar()
	options := &pebble.Options{
		Cache:    cache,
		Comparer: comparer,
		Logger:   plg,
	}
	db, err := pebble.Open(dir, options)
	if err != nil {
		return nil, err
	}
	return db, nil
}
