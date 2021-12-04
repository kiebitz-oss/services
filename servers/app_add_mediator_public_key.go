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
	"encoding/json"
	"github.com/kiebitz-oss/services"
	"github.com/kiebitz-oss/services/crypto"
)

// { keys }, keyPair
// add the mediator key to the list of keys (only for testing)
func (c *Appointments) addMediatorPublicKeys(context services.Context, params *services.AddMediatorPublicKeysSignedParams) services.Response {
	if response := c.isRoot(context, []byte(params.JSON), params.Signature, params.Data.Timestamp); response != nil {
		return response
	}
	hash := crypto.Hash(params.Data.Signing)
	keys := c.db.Map("keys", []byte("mediators"))
	bd, err := json.Marshal(params)
	if err != nil {
		services.Log.Error(err)
		return context.InternalError()
	}
	if err := keys.Set(hash, bd); err != nil {
		services.Log.Error(err)
		return context.InternalError()
	}
	if result, err := keys.Get(hash); err != nil {
		services.Log.Error(err)
		return context.InternalError()
	} else if !bytes.Equal(result, bd) {
		services.Log.Error("does not match")
		return context.InternalError()
	}
	return context.Acknowledge()
}
