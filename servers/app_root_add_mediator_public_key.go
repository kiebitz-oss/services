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
	"github.com/kiebitz-oss/services"
	"github.com/kiebitz-oss/services/crypto"
)

// { keys }, keyPair
// add the mediator key to the list of keys (only for testing)
func (c *Appointments) addMediatorPublicKeys(context services.Context, params *services.AddMediatorPublicKeysSignedParams) services.Response {

	if resp := c.isRoot(context, &services.SignedParams{
		JSON:      params.JSON,
		Signature: params.Signature,
		PublicKey: params.PublicKey,
		Timestamp: params.Data.Timestamp,
	}); resp != nil {
		return resp
	}

	mediatorKey := &services.ActorKey{
		Data:      params.Data.SignedKeyData.JSON,
		Signature: params.Data.SignedKeyData.Signature,
		PublicKey: params.Data.SignedKeyData.PublicKey,
	}

	hash := crypto.Hash(params.Data.SignedKeyData.Data.Signing)

	keys := c.backend.Keys("mediators")

	if err := keys.Set(hash, mediatorKey); err != nil {
		services.Log.Error(err)
		return context.InternalError()
	}

	return context.Acknowledge()
}
