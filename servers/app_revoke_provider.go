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
	//"encoding/json"
	"github.com/kiebitz-oss/services"
	//"github.com/kiebitz-oss/services/crypto"
	"github.com/kiebitz-oss/services/databases"
)

// { id, key, providerData, keyData }, keyPair
func (c *Appointments) revokeProvider(context services.Context, params *services.RevokeProviderSignedParams) services.Response {

	if resp, _ := c.isMediator(context, []byte(params.JSON), params.Signature, params.PublicKey); resp != nil {
		return resp
	}

	providerID := params.Data.ProviderID

	providerDataKeys := []services.Map {
		c.db.Map("providerData", []byte("unverified")),
		c.db.Map("providerData", []byte("verified")),
		c.db.Map("providerData", []byte("checked")),
		c.db.Map("providerData", []byte("public")),
	}

	for _, providerData := range providerDataKeys {
		err := providerData.Del([]byte(providerID))

		if err!= nil && err != databases.NotFound {
			services.Log.Error(err)
			return context.InternalError()
		}

	}

	appById := c.db.Map("appointmentsByID", params.Data.ProviderID)
	allDates, err := appById.GetAll()

	if err != nil && err != databases.NotFound {
		services.Log.Error(err)
		return context.InternalError()

	} else if err == nil {

		for _, date := range allDates{
			if err := c.db.Map("appointmentsByDate", append(providerID, date...)).Remove(); err != nil {
				services.Log.Error(err)
				return context.InternalError()
			}
		}

		if err := appById.Remove(); err != nil {
			services.Log.Error(err)
			return context.InternalError()
		}

	}

	return context.Acknowledge()
}
