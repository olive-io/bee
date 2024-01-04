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

package rpctype

import (
	"github.com/cockroachdb/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/olive-io/bee/client"
)

func ToGRPCErr(err error) *status.Status {
	if errors.Is(err, client.ErrTimeout) {
		return status.New(codes.Canceled, errors.Unwrap(err).Error())
	}
	if errors.Is(err, client.ErrNotExists) {
		return status.New(codes.NotFound, errors.Unwrap(err).Error())
	}

	return status.New(codes.Unknown, err.Error())
}

func ParseGRPCErr(err error) error {
	if vs, ok := status.FromError(err); ok {
		switch vs.Code() {
		case codes.NotFound:
			return errors.Wrap(client.ErrNotExists, vs.Message())
		default:
			return errors.Wrap(client.ErrRequest, vs.Message())
		}
	}

	return err
}
