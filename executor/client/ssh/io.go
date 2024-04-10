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

package ssh

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/sftp"

	"github.com/olive-io/bee/executor/client"
)

func get(
	ctx context.Context,
	ftp *sftp.Client,
	src string,
	dst string,
	buf []byte,
	fn client.IOTraceFn,
) (written int64, err error) {

	reader, err := ftp.Open(src)
	if err != nil {
		return 0, err
	}
	defer reader.Close()

	writer, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer writer.Close()

	var trace *client.IOTrace
	if fn != nil {
		trace = &client.IOTrace{
			Name: filepath.Base(reader.Name()),
			Src:  src,
			Dst:  dst,
		}
		if stat, _ := reader.Stat(); stat != nil {
			trace.Total = stat.Size()
		}
	}

	return fcopy(ctx, reader, writer, trace, buf, fn)
}

func put(
	ctx context.Context,
	ftp *sftp.Client,
	src string,
	dst string,
	buf []byte,
	fn client.IOTraceFn,
) (written int64, err error) {

	reader, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer reader.Close()

	writer, err := ftp.Create(dst)
	if err != nil {
		return 0, err
	}
	defer writer.Close()

	if stat, _ := reader.Stat(); stat != nil {
		_ = ftp.Chmod(dst, stat.Mode())
	}

	var trace *client.IOTrace
	if fn != nil {
		trace = &client.IOTrace{
			Name: reader.Name(),
			Src:  src,
			Dst:  dst,
		}
		if stat, _ := reader.Stat(); stat != nil {
			trace.Total = stat.Size()
		}
	}

	return fcopy(ctx, reader, writer, trace, buf, fn)
}

func fcopy(
	ctx context.Context,
	reader io.Reader,
	writer io.Writer,
	trace *client.IOTrace,
	buf []byte,
	fn client.IOTraceFn,
) (written int64, err error) {
	if buf == nil {
		buf = make([]byte, 32*1024)
	}

	last := time.Now()
	sub := int64(0)
	for {
		select {
		case <-ctx.Done():
			err = client.ErrTimeout
			return
		default:
		}

		nr, er := reader.Read(buf)
		if nr > 0 {
			nw, ew := writer.Write(buf[0:nr])
			if nw < 0 || nr < nw {
				nw = 0
				if ew == nil {
					ew = client.ErrInvalidWrite
				}
			}
			written += int64(nw)
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = io.ErrShortWrite
				break
			}
		}
		if fn != nil {
			now := time.Now()
			trace.Chunk = written
			trace.Speed = int64(float64(written-sub) / (now.Sub(last).Seconds()))
			last = now
			sub = written
			fn(trace)
		}
		if er != nil {
			if er != io.EOF {
				err = er
			}
			break
		}
	}
	return written, err
}
