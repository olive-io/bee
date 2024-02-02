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

package trace

import (
	"bytes"
	"crypto/tls"
	"net/http"
	urlpkg "net/url"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/d5/tengo/v2"
	"github.com/olive-io/bee/tengo/slog"
)

const (
	defaultTimeout = time.Second * 5
)

type hook struct {
	url    *urlpkg.URL
	client *http.Client
}

func newHook(url string) (*hook, error) {
	URL, err := urlpkg.Parse(url)
	if err != nil {
		return nil, err
	}

	tr := &http.Transport{}
	if URL.Scheme == "https" {
		tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	client := &http.Client{
		Transport: tr,
		Timeout:   defaultTimeout,
	}

	hk := &hook{
		url:    URL,
		client: client,
	}

	return hk, nil
}

func (h *hook) Write(data []byte) (n int, err error) {
	body := bytes.NewBuffer(data)
	req, err := http.NewRequest(http.MethodPost, h.url.String(), body)
	if err != nil {
		return 0, err
	}

	req.Header.Set("Content-Type", "application/json")

	rsp, err := h.client.Do(req)
	if err != nil {
		return 0, err
	}

	defer rsp.Body.Close()
	return len(data), nil
}

func (m *ImportModule) AddHook() tengo.CallableFunc {
	return func(args ...tengo.Object) (ret tengo.Object, err error) {
		numArgs := len(args)
		if numArgs == 0 {
			return nil, errors.Wrap(tengo.ErrWrongNumArguments, "must greater than 0")
		}

		url, ok := args[0].(*tengo.String)
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "url",
				Expected: "string",
				Found:    args[1].TypeName(),
			}
		}

		hk, err := newHook(url.Value)
		if err != nil {
			return nil, err
		}

		attrs := make([]slog.Attr, 0)
		if len(args) > 1 {
			for _, arg := range args[1:] {
				if attr, ok := arg.(*traceField); ok {
					attrs = append(attrs, attr.Value)
				}
			}
		}

		options := &slog.HandlerOptions{
			Level: m.level,
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				return a
			},
		}

		handler := slog.NewJSONHandler(hk, options)
		m.handler.AddHandler(handler.WithAttrs(attrs))

		return tengo.UndefinedValue, nil
	}
}
