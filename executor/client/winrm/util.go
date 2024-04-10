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

package winrm

import (
	"context"
	"encoding/xml"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/olive-io/winrm"

	"github.com/olive-io/bee/executor/client"
)

type pslist struct {
	Objects []psobject `xml:"Object"`
}

type psobject struct {
	Properties []psproperty `xml:"Property"`
	Value      string       `xml:",innerxml"`
}

type psproperty struct {
	Name  string `xml:"Name,attr"`
	Value string `xml:",innerxml"`
}

// fileInfo is an artificial type designed to satisfy os.FileInfo.
type fileInfo struct {
	name  string
	size  int64
	mode  os.FileMode
	mtime time.Time
	sys   interface{}
}

// Name returns the base name of the file.
func (fi *fileInfo) Name() string { return fi.name }

// Size returns the length in bytes for regular files; system-dependent for others.
func (fi *fileInfo) Size() int64 { return fi.size }

// Mode returns file mode bits.
func (fi *fileInfo) Mode() os.FileMode { return fi.mode }

// ModTime returns the last modification time of the file.
func (fi *fileInfo) ModTime() time.Time { return fi.mtime }

// IsDir returns true if the file is a directory.
func (fi *fileInfo) IsDir() bool { return fi.Mode().IsDir() }

func (fi *fileInfo) Sys() interface{} { return fi.sys }

func fetchRemoteDir(ctx context.Context, cc *winrm.Client, remotePath string) ([]os.FileInfo, error) {
	script := fmt.Sprintf("Get-ChildItem %s", remotePath)
	stdout, _, _, err := cc.RunPSWithContext(ctx, "powershell -Command \""+script+" | ConvertTo-Xml -NoTypeInformation -As String\"")
	if err != nil {
		return nil, errors.Wrap(client.ErrRequest, err.Error())
	}

	if stdout != "" {
		doc := pslist{}
		err := xml.Unmarshal([]byte(stdout), &doc)
		if err != nil {
			return nil, errors.Wrapf(client.ErrRequest, "couldn't parse results: %v", err)
		}

		return convertFileItems(doc.Objects), nil
	}

	return []os.FileInfo{}, nil
}

func convertFileItems(objects []psobject) []os.FileInfo {
	items := make([]os.FileInfo, 0)

	for _, object := range objects {
		stat := &fileInfo{}
		for _, property := range object.Properties {
			switch property.Name {
			case "Name":
				stat.name = property.Value
			case "Mode":
				if property.Value[0] == 'd' {
					stat.mode = os.ModeDir
				} else {
					stat.mode = os.ModeAppend
				}
			//case "FullName":
			//	items[i].Path = property.Value
			case "Length":
				if n, err := strconv.ParseInt(property.Value, 10, 64); err == nil {
					stat.size = n
				}
			case "LastWriteTime":
				stat.mtime, _ = time.Parse("2006/1/02 15:04:05", property.Value)
			}
		}
		stat.sys = struct{}{}
		items = append(items, stat)
	}

	return items
}
