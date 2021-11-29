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
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"github.com/kiebitz-oss/services"
	"github.com/kiebitz-oss/services/crypto"
	"github.com/kiebitz-oss/services/databases"
	"github.com/kiebitz-oss/services/jsonrpc"
)

func (c *Appointments) priorityToken() (uint64, []byte, error) {
	data := c.db.Value("priorityToken", []byte("primary"))
	if token, err := data.Get(); err != nil && err != databases.NotFound {
		return 0, nil, err
	} else {
		var intToken uint64
		if err == nil {
			intToken = binary.LittleEndian.Uint64(token)
		}
		intToken = intToken + 1
		bs := make([]byte, 8)
		binary.LittleEndian.PutUint64(bs, intToken)

		if err := data.Set(bs, 0); err != nil {
			return 0, nil, err
		}

		h := hmac.New(sha256.New, c.settings.Secret)
		h.Write(bs)

		token := h.Sum(nil)

		return intToken, token[:], nil

	}
}

//{hash, code, publicKey}
// get a token for a given queue
func (c *Appointments) getToken(context *jsonrpc.Context, params *services.GetTokenParams) *jsonrpc.Response {

	codes := c.db.Set("codes", []byte("user"))
	codeScores := c.db.SortedSet("codeScores", []byte("user"))

	tokenKey := c.settings.Key("token")
	if tokenKey == nil {
		services.Log.Error("token key missing")
		return context.InternalError()
	}

	var signedData *crypto.SignedStringData

	if c.settings.UserCodesEnabled {
		notAuthorized := context.Error(401, "not authorized", nil)
		if params.Code == nil {
			return notAuthorized
		}
		if ok, err := codes.Has(params.Code); err != nil {
			services.Log.Error()
			return context.InternalError()
		} else if !ok {
			return notAuthorized
		}
	}

	if _, token, err := c.priorityToken(); err != nil {
		services.Log.Error(err)
		return context.InternalError()
	} else {
		tokenData := &services.TokenData{
			Hash:      params.Hash,
			Token:     token,
			PublicKey: params.PublicKey,
		}

		td, err := json.Marshal(tokenData)

		if err != nil {
			services.Log.Error(err)
			return context.InternalError()
		}

		if signedData, err = tokenKey.SignString(string(td)); err != nil {
			services.Log.Error(err)
			return context.InternalError()
		}
	}

	// if this is a new token we delete the user code
	if c.settings.UserCodesEnabled {
		score, err := codeScores.Score(params.Code)
		if err != nil && err != databases.NotFound {
			services.Log.Error(err)
			return context.InternalError()
		}

		score += 1

		if score > c.settings.UserCodesReuseLimit {
			if err := codes.Del(params.Code); err != nil {
				services.Log.Error(err)
				return context.InternalError()
			}
		} else if err := codeScores.Add(params.Code, score); err != nil {
			services.Log.Error(err)
			return context.InternalError()
		}
	}

	return context.Result(signedData)

}
