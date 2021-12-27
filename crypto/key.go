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

type Key struct {
	Name      string                 `json:"name"`
	Type      string                 `json:"type"`
	Format    string                 `json:"format"`
	Params    map[string]interface{} `json:"params"`
	PublicKey []byte                 `json:"publicKey"`
	Purposes  []string               `json:"purposes"`
	// only defined for local signing operations
	PrivateKey []byte `json:"privateKey,omitempty"`
}

func (k *Key) Encrypt(data []byte, recipient *Key) (*ECDHEncryptedData, error) {
	if privateKey, err := LoadPrivateKey(k.PrivateKey); err != nil {
		return nil, err
	} else if publicKey, err := LoadPublicKey(recipient.PublicKey); err != nil {
		return nil, err
	} else {
		key := DeriveKey(publicKey, privateKey)
		if encryptedData, err := Encrypt(data, key); err != nil {
			return nil, err
		} else {
			return &ECDHEncryptedData{
				IV:        encryptedData.IV,
				Data:      encryptedData.Data,
				PublicKey: k.PublicKey,
			}, nil
		}
	}
}

func (k *Key) SignString(data string) (*SignedStringData, error) {
	if signature, err := k.Sign([]byte(data)); err != nil {
		return nil, err
	} else {
		return &SignedStringData{
			Data:      string(signature.Data),
			Signature: signature.Signature,
			PublicKey: signature.PublicKey,
		}, nil
	}
}

func (k *Key) Sign(data []byte) (*SignedData, error) {
	if privateKey, err := LoadPrivateKey(k.PrivateKey); err != nil {
		return nil, err
	} else if signature, err := Sign(data, privateKey); err != nil {
		return nil, err
	} else {
		return &SignedData{
			Data:      data,
			Signature: signature.Serialize(),
			PublicKey: k.PublicKey,
		}, nil
	}
}

func (k *Key) Verify(data *SignedData) (bool, error) {
	if publicKey, err := LoadPublicKey(k.PublicKey); err != nil {
		return false, err
	} else {
		return Verify(data.Data, data.Signature, publicKey)
	}
}

func (k *Key) VerifyString(data *SignedStringData) (bool, error) {
	if publicKey, err := LoadPublicKey(k.PublicKey); err != nil {
		return false, err
	} else {
		return Verify([]byte(data.Data), data.Signature, publicKey)
	}
}
