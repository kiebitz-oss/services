// Kiebitz - Privacy-Friendly Appointment Scheduling
// Copyright (C) 2021-2021 The Kiebitz Authors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version. Additional terms
// as defined in section 7 of the license (e.g. regarding attribution)
// are specified at https://kiebitz.eu/en/docs/open-source/additional-terms.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

// This code was extracted from the Kodex CE project
// (https://github.com/kiprotect/kodex), the author has
// the copyright to the code so no attribution is necessary.

package encryptFs

import (
	"encoding/json"
	"fmt"
	"github.com/kiebitz-oss/services/crypto"
	"io"
	"io/fs"
)

type EncryptedFS struct {
	envfs fs.FS
	key   []byte
}

type EncryptedFile struct {
	file   fs.File
	key    []byte
	buffer []byte
	isEOF  bool
}

func New(fs fs.FS, key []byte) fs.FS {
	return &EncryptedFS{
		envfs: fs,
		key:   key,
	}
}

func (e *EncryptedFS) Open(name string) (fs.File, error) {
	if file, err := e.envfs.Open(name); err != nil {
		return nil, err
	} else {
		return &EncryptedFile{
			file:   file,
			key:    e.key,
			buffer: []byte{},
			isEOF:  false,
		}, nil
	}
}

func (e *EncryptedFS) ReadDir(name string) ([]fs.DirEntry, error) {
	return fs.ReadDir(e.envfs, name)
}

func (e *EncryptedFile) Close() error {
	return e.file.Close()
}

func (e *EncryptedFile) Stat() (fs.FileInfo, error) {
	return e.file.Stat()
}

// this is not efficient and should only be used for relatively small files
func (e *EncryptedFile) Read(p []byte) (int, error) {
	if len(e.buffer) == 0 && !e.isEOF {

		var data []byte
		var buffer = make([]byte, 32)

		for {
			size, err := e.file.Read(buffer)
			if err == io.EOF {
				break
			} else if err != nil {
				return 0, err
			}
			data = append(data, buffer[:size]...)
		}

		var encData crypto.EncryptedData
		if err := json.Unmarshal(data, &encData); err != nil || encData.IV == nil || encData.Data == nil {
			e.buffer = data
		} else {
			decData, err := crypto.Decrypt(&encData, e.key)
			if err != nil {
				return 0, fmt.Errorf("decrypt failed")
			}
			e.buffer = decData
		}

	}
	if e.isEOF {
		return 0, io.EOF
	}

	size := copy(p, e.buffer)
	e.buffer = e.buffer[size:]
	if len(e.buffer) == 0 {
		e.isEOF = true
	}

	return size, nil

}
