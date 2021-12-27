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

package crypto

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	pbkdf2 "golang.org/x/crypto/pbkdf2"
	"os"
)

const rounds = 100_000
const base64Salt = "tlsfpYaKiH/WZUnWkoeE2g=="
const envPassName = "KIEBITZ_PASSPHRASE"

func BuildKeyFromEnv() ([]byte, error) {
	salt, _ := base64.StdEncoding.DecodeString(base64Salt)
	passphrase := os.Getenv(envPassName)
	if passphrase == "" {
		return nil, fmt.Errorf("no passphrase in environment")
	}
	return pbkdf2.Key([]byte(passphrase), salt, rounds, 32, sha256.New), nil
}
