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
	"github.com/kiebitz-oss/services/databases"
)

func (c *Appointments) revokeMediator(context services.Context, params *services.RevokeMediatorSignedParams) services.Response {

	if resp, _ := c.isMediator(context, []byte(params.JSON), params.Signature, params.PublicKey); resp != nil {
		return resp
	}

	mediatorID := params.Data.MediatorID

	mediatorKeys := c.db.Map("keys", []byte("mediators"))

	if err := mediatorKeys.Del([]byte(mediatorID)); err != nil && err != databases.NotFound {
		services.Log.Error(err)
		return context.InternalError()
	}

	return context.Acknowledge()
}
