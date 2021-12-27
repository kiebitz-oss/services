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

package servers

import (
	"github.com/kiebitz-oss/services"
)

// mediator-only endpoint
// { limit }, keyPair
func (c *Appointments) getVerifiedProviderData(context services.Context, params *services.GetPendingProviderDataSignedParams) services.Response {

	resp, _ := c.isMediator(context, &services.SignedParams{
		JSON:      params.JSON,
		Signature: params.Signature,
		PublicKey: params.PublicKey,
		Timestamp: params.Data.Timestamp,
	})

	if resp != nil {
		return resp
	}

	verifiedProviderData := c.backend.VerifiedProviderData()

	providerDataMap, err := verifiedProviderData.GetAll()

	if err != nil {
		services.Log.Error(err)
		return context.InternalError()
	}

	pdEntries := []*services.RawProviderData{}

	for id, pd := range providerDataMap {
		pd.ID = []byte(id)
		pdEntries = append(pdEntries, pd)
	}

	return context.Result(pdEntries)

}
