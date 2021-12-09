// Kiebitz - Privacy-Friendly Appointment Scheduling
// Copyright (C) 2021-2021 The Kiebitz Authors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package servers

import (
	"bytes"
	"github.com/kiebitz-oss/services"
	"github.com/kiebitz-oss/services/crypto"
	"time"
)

func findActorKey(keys []*services.ActorKey, publicKey []byte) (*services.ActorKey, error) {
	for _, key := range keys {
		if akd, err := key.KeyData(); err != nil {
			services.Log.Error(err)
			continue
		} else if bytes.Equal(akd.Signing, publicKey) {
			return key, nil
		}
	}
	return nil, nil
}

func isRoot(context services.Context, data, signature []byte, timestamp time.Time, keys []*crypto.Key) services.Response {
	rootKey := services.Key(keys, "root")
	if rootKey == nil {
		services.Log.Error("root key missing")
		return context.InternalError()
	}
	if ok, err := rootKey.Verify(&crypto.SignedData{
		Data:      data,
		Signature: signature,
	}); !ok {
		return context.Error(403, "invalid signature", nil)
	} else if err != nil {
		services.Log.Error(err)
		return context.InternalError()
	}
	if expired(timestamp) {
		return context.Error(410, "signature expired", nil)
	}
	return nil
}

func expired(timestamp time.Time) bool {
	return time.Now().Add(-time.Minute).After(timestamp)
}
