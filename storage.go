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
