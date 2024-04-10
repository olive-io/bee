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

package rpctype

import (
	"io"
	"os"

	"github.com/cockroachdb/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/olive-io/bee/executor/client"
)

func ToGRPCErr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, io.EOF) {
		return err
	}
	if errors.Is(err, os.ErrNotExist) {
		return status.New(codes.NotFound, errors.Unwrap(err).Error()).Err()
	}
	if errors.Is(err, client.ErrTimeout) {
		return status.New(codes.Canceled, errors.Unwrap(err).Error()).Err()
	}
	if errors.Is(err, client.ErrNotExists) {
		return status.New(codes.NotFound, errors.Unwrap(err).Error()).Err()
	}

	return status.New(codes.Unknown, err.Error()).Err()
}

func ParseGRPCErr(err error) error {
	if err == nil {
		return nil
	}

	if vs, ok := status.FromError(err); ok {
		switch vs.Code() {
		case codes.OK:
			return nil
		case codes.NotFound:
			return errors.Wrap(client.ErrNotExists, vs.Message())
		default:
			return errors.Wrap(client.ErrRequest, vs.Message())
		}
	}

	return err
}
