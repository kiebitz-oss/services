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
	"encoding/json"
	"github.com/kiebitz-oss/services"
	"github.com/kiebitz-oss/services/crypto"
	"github.com/kiebitz-oss/services/databases"
	"github.com/kiebitz-oss/services/jsonrpc"
)

// { id, encryptedData, code }, keyPair
func (c *Appointments) checkProviderData(context *jsonrpc.Context, params *services.CheckProviderDataSignedParams) *jsonrpc.Response {

	// make sure this is a valid provider
	resp, _ := c.isProvider(context, []byte(params.JSON), params.Signature, params.PublicKey)

	if resp != nil {
		return resp
	}

	if expired(params.Data.Timestamp) {
		return context.Error(410, "signature expired", nil)
	}

	hash := crypto.Hash(params.PublicKey)
	verifiedProviderData := c.db.Map("providerData", []byte("checked"))

	if data, err := verifiedProviderData.Get(hash); err != nil {
		if err == databases.NotFound {
			return context.NotFound()
		}
		services.Log.Error(err)
		return context.InternalError()
	} else {
		var m map[string]interface{}
		if err := json.Unmarshal(data, &m); err != nil {
			services.Log.Error(err)
			return context.InternalError()
		} else {
			return context.Result(m)
		}
	}

	return context.Acknowledge()
}
