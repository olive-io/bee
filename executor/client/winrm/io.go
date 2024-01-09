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

package winrm

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/google/uuid"
	"github.com/masterzen/winrm"
	"go.uber.org/zap"

	"github.com/olive-io/bee/executor/client"
)

type fileWalker struct {
	ctx     context.Context
	lg      *zap.Logger
	cc      *winrm.Client
	toDir   string
	fromDir string
	fn      client.IOTraceFn
}

func (fw *fileWalker) copyFile(fromPath string, fi os.FileInfo, err error) error {
	if err != nil {
		return err
	}

	if !shouldUploadFile(fi) {
		return nil
	}

	hostPath, _ := filepath.Abs(fromPath)
	fromDir, _ := filepath.Abs(fw.fromDir)
	relPath, _ := filepath.Rel(fromDir, hostPath)
	toPath := filepath.Join(fw.toDir, relPath)

	f, err := os.Open(hostPath)
	if err != nil {
		return fmt.Errorf("couldn't read file %s: %v", fromPath, err)
	}

	return doCopy(fw.ctx, fw.lg, fw.cc, f, winPath(toPath), fw.fn)
}

func shouldUploadFile(fi os.FileInfo) bool {
	// Ignore dir entries and OS X special hidden file
	return !fi.IsDir() && ".DS_Store" != fi.Name()
}

func doCopy(ctx context.Context, lg *zap.Logger, cc *winrm.Client, in *os.File, toPath string, fn client.IOTraceFn) error {
	tempFile := tempFileName()
	tempPath := "$env:TEMP\\" + tempFile

	defer func() {
		_ = cleanupContent(ctx, lg, cc, tempPath)
	}()

	var trace *client.IOTrace
	if fn != nil {
		trace = &client.IOTrace{
			Name: filepath.Base(in.Name()),
			Src:  in.Name(),
			Dst:  toPath,
		}
		stat, _ := in.Stat()
		if stat != nil {
			trace.Total = stat.Size()
		}
	}
	err := uploadContent(ctx, lg, cc, 0, "%TEMP%\\"+tempFile, in, trace, fn)
	if err != nil {
		return errors.Errorf("error uploading file to %s", tempPath)
	}

	err = restoreContent(ctx, lg, cc, tempPath, toPath)
	if err != nil {
		return errors.Wrapf(err, "error restoring file from %s to %s", tempPath, toPath)
	}

	return nil
}

func uploadContent(
	ctx context.Context, lg *zap.Logger, cc *winrm.Client,
	maxChunks int, filePath string, reader io.Reader,
	trace *client.IOTrace, fn client.IOTraceFn,
) (err error) {
	done := false
	for !done {
		done, err = uploadChunks(ctx, lg, cc, filePath, maxChunks, reader, trace, fn)
		if err != nil {
			return err
		}
	}

	return nil
}

func uploadChunks(
	ctx context.Context, lg *zap.Logger, cc *winrm.Client,
	filePath string, maxChunks int, reader io.Reader,
	trace *client.IOTrace, fn client.IOTraceFn,
) (bool, error) {
	shell, err := cc.CreateShell()
	if err != nil {
		return false, err
	}
	defer shell.Close()

	// Upload the file in chunks to get around the Windows command line size limit.
	// Base64 encodes each set of three bytes into four bytes. In addition the output
	// is padded to always be a multiple of four.
	//
	//   ceil(n / 3) * 4 = m1 - m2
	//
	//   where:
	//     n  = bytes
	//     m1 = max (8192 character command limit.)
	//     m2 = len(filePath)
	chunkSize := ((8000 - len(filePath)) / 4) * 3
	chunk := make([]byte, chunkSize)

	if maxChunks == 0 {
		maxChunks = 1
	}

	last := time.Now()
	for i := 0; i < maxChunks; i++ {
		n, err := reader.Read(chunk)
		if err != nil && err != io.EOF {
			return false, err
		}

		if fn != nil {
			now := time.Now()
			trace.Chunk += int64(n)
			trace.Speed = int64(float64(n) / (now.Sub(last).Seconds()))
			last = now
			fn(trace)
		}

		if n == 0 {
			return true, nil
		}

		content := base64.StdEncoding.EncodeToString(chunk[:n])
		if err = appendContent(ctx, lg, shell, filePath, content); err != nil {
			return false, err
		}
	}

	return false, nil
}

func restoreContent(ctx context.Context, lg *zap.Logger, cc *winrm.Client, fromPath, toPath string) error {
	shell, err := cc.CreateShell()
	if err != nil {
		return err
	}

	lg.Debug("restore file content", zap.String("from", fromPath), zap.String("to", toPath))

	defer shell.Close()
	script := fmt.Sprintf(`
		$tmp_file_path = [System.IO.Path]::GetFullPath("%s")
		$dest_file_path = [System.IO.Path]::GetFullPath("%s".Trim("'"))
		if (Test-Path $dest_file_path) {
			if (Test-Path -Path $dest_file_path -PathType container) {
				Exit 1
			} else {
				rm $dest_file_path
			}
		}
		else {
			$dest_dir = ([System.IO.Path]::GetDirectoryName($dest_file_path))
			New-Item -ItemType directory -Force -ErrorAction SilentlyContinue -Path $dest_dir | Out-Null
		}

		if (Test-Path $tmp_file_path) {
			$reader = [System.IO.File]::OpenText($tmp_file_path)
			$writer = [System.IO.File]::OpenWrite($dest_file_path)
			try {
				for(;;) {
					$base64_line = $reader.ReadLine()
					if ($base64_line -eq $null) { break }
					$bytes = [System.Convert]::FromBase64String($base64_line)
					$writer.write($bytes, 0, $bytes.Length)
				}
			}
			finally {
				$reader.Close()
				$writer.Close()
			}
		} else {
			echo $null > $dest_file_path
		}
	`, fromPath, toPath)

	cmd, err := shell.ExecuteWithContext(ctx, winrm.Powershell(script))
	if err != nil {
		return err
	}
	defer cmd.Close()

	var wg sync.WaitGroup
	copyFunc := func(w io.Writer, r io.Reader) {
		defer wg.Done()
		_, _ = io.Copy(w, r)
	}

	wg.Add(2)
	go copyFunc(os.Stdout, cmd.Stdout)
	go copyFunc(os.Stderr, cmd.Stderr)

	cmd.Wait()
	wg.Wait()

	if cmd.ExitCode() != 0 {
		return errors.Wrapf(client.ErrRequest, "restore operation returned code=%d", cmd.ExitCode())
	}
	return nil
}

func cleanupContent(ctx context.Context, lg *zap.Logger, cc *winrm.Client, filePath string) error {
	shell, err := cc.CreateShell()
	if err != nil {
		return err
	}

	lg.Debug("cleanup file", zap.String("path", filePath))

	defer shell.Close()
	script := fmt.Sprintf(`
		$tmp_file_path = [System.IO.Path]::GetFullPath("%s")
		if (Test-Path $tmp_file_path) {
			Remove-Item $tmp_file_path -ErrorAction SilentlyContinue
		}
	`, filePath)

	cmd, err := shell.ExecuteWithContext(ctx, winrm.Powershell(script))
	if err != nil {
		return err
	}
	defer cmd.Close()

	var wg sync.WaitGroup
	copyFunc := func(w io.Writer, r io.Reader) {
		defer wg.Done()
		_, _ = io.Copy(w, r)
	}

	wg.Add(2)
	go copyFunc(os.Stdout, cmd.Stdout)
	go copyFunc(os.Stderr, cmd.Stderr)

	cmd.Wait()
	wg.Wait()

	if cmd.ExitCode() != 0 {
		return errors.Wrapf(client.ErrRequest, "cleanup operation returned code=%d", cmd.ExitCode())
	}
	return nil
}

func appendContent(ctx context.Context, lg *zap.Logger, shell *winrm.Shell, filePath, content string) error {
	cmd, err := shell.ExecuteWithContext(ctx, fmt.Sprintf(`echo %s >> %s`, content, filePath))
	if err != nil {
		return err
	}

	lg.Debug("append file", zap.String("path", filePath))

	defer cmd.Close()
	var wg sync.WaitGroup

	copyFunc := func(w io.Writer, r io.Reader) {
		defer wg.Done()
		_, _ = io.Copy(w, r)
	}

	wg.Add(2)
	go copyFunc(os.Stdout, cmd.Stdout)
	go copyFunc(os.Stderr, cmd.Stderr)

	cmd.Wait()
	wg.Wait()

	if cmd.ExitCode() != 0 {
		return errors.Wrapf(client.ErrRequest, "upload operation returned code=%d", cmd.ExitCode())
	}

	return nil
}

func tempFileName() string {
	return fmt.Sprintf("winrmcp-%s.tmp", uuid.New().String())
}

func readContent(ctx context.Context, lg *zap.Logger, cc *winrm.Client, dst string, writer *os.File, buf []byte, fn client.IOTraceFn) error {
	shell, err := cc.CreateShell()
	if err != nil {
		return err
	}

	lg.Debug("read file", zap.String("path", dst))
	defer shell.Close()
	script := fmt.Sprintf(`
$dest_file_path = [System.IO.Path]::GetFullPath("%s".Trim("'"))
Get-Content $dest_file_path`, dst)

	cmd, err := shell.ExecuteWithContext(ctx, winrm.Powershell(script))
	if err != nil {
		return err
	}
	defer cmd.Close()

	var trace *client.IOTrace
	if fn != nil {
		trace = &client.IOTrace{
			Name: filepath.Base(writer.Name()),
			Src:  writer.Name(),
			Dst:  dst,
		}
		stat, _ := writer.Stat()
		if stat != nil {
			trace.Total = stat.Size()
		}
	}

	var wg sync.WaitGroup
	wg.Add(2)

	errBuf := make([]byte, 1024*4)
	go func() {
		defer wg.Done()
		n, _ := cmd.Stderr.Read(errBuf)
		errBuf = errBuf[:n]
	}()

	go func() {
		defer wg.Done()
		written := int64(0)
		last := time.Now()
		for {
			nr, er := cmd.Stdout.Read(buf)
			if nr > 0 {
				chunk := bytes.Replace(buf[0:nr], []byte("\uFEFF"), []byte(""), -1)
				nw, ew := writer.Write(chunk)
				if nw < 0 || nr < nw {
					nw = 0
					if ew == nil {
						ew = client.ErrInvalidWrite
					}
				}

				if fn != nil {
					now := time.Now()
					trace.Chunk += int64(nw)
					trace.Speed = int64(float64(nw) / (now.Sub(last).Seconds()))
					last = now
					fn(trace)
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
			if er != nil {
				if er != io.EOF {
					err = er
				}
				break
			}
		}
	}()

	cmd.Wait()
	wg.Wait()

	if cmd.ExitCode() != 0 {
		return errors.Wrapf(client.ErrRequest, "read file operation returned code=%d, %s", cmd.ExitCode(), string(errBuf))
	}
	return nil
}

func winPath(path string) string {
	if len(path) == 0 {
		return path
	}

	if strings.Contains(path, " ") {
		path = fmt.Sprintf("'%s'", strings.Trim(path, "'\""))
	}

	return strings.Replace(path, "/", "\\", -1)
}
