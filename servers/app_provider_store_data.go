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
	"github.com/kiebitz-oss/services/databases"
)

// { id, encryptedData, code }, keyPair
func (c *Appointments) storeProviderData(context services.Context, params *services.StoreProviderDataSignedParams) services.Response {

	// we verify the signature (without veryfing e.g. the provenance of the key)
	// this is important as we use the public key as an identifier for the provider
	// data so we need to make sure the caller is actually in possession of the key
	if ok, err := crypto.VerifyWithBytes([]byte(params.JSON), params.Signature, params.PublicKey); err != nil {
		services.Log.Error(err)
		return context.InternalError()
	} else if !ok {
		return context.Error(400, "invalid signature", nil)
	}

	// to do: add one-time use check

	hash := crypto.Hash(params.PublicKey)

	lock, err := c.db.Lock("storeProviderData_" + string(hash[:]))
	if err != nil {
		services.Log.Error(err)
		return context.InternalError()
	}

	defer lock.Release()

	verifiedProviderData := c.backend.VerifiedProviderData()
	providerData := c.backend.UnverifiedProviderData()

	codes := c.db.Set("codes", []byte("provider"))
	codeScores := c.db.SortedSet("codeScores", []byte("provider"))

	existingData := false
	if result, err := verifiedProviderData.Get(hash); err != nil {
		if err != databases.NotFound {
			services.Log.Error(err)
			return context.InternalError()
		}
	} else if result != nil {
		existingData = true
	}

	if (!existingData) && c.settings.ProviderCodesEnabled {
		notAuthorized := context.Error(401, "not authorized", nil)
		if params.Data.Code == nil {
			return notAuthorized
		}
		if ok, err := codes.Has(params.Data.Code); err != nil {
			services.Log.Error()
			return context.InternalError()
		} else if !ok {
			return notAuthorized
		}
	}

	if err := providerData.Set(hash, &services.RawProviderData{
		EncryptedData: params.Data.EncryptedData,
	}); err != nil {
		services.Log.Error(err)
		return context.InternalError()
	}

	// we delete the provider code
	if c.settings.ProviderCodesEnabled {
		score, err := codeScores.Score(params.Data.Code)
		if err != nil && err != databases.NotFound {
			services.Log.Error(err)
			return context.InternalError()
		}

		score += 1

		if score > c.settings.ProviderCodesReuseLimit {
			if err := codes.Del(params.Data.Code); err != nil {
				services.Log.Error(err)
				return context.InternalError()
			}
		} else if err := codeScores.Add(params.Data.Code, score); err != nil {
			services.Log.Error(err)
			return context.InternalError()
		}
	}

	return context.Acknowledge()
}
