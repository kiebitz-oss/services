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

package services

import (
	"encoding/json"
	"github.com/kiebitz-oss/services/crypto"
	"time"
)

// Generic signed params for checking

type SignedParams struct {
	JSON      string
	Signature []byte
	PublicKey []byte
	Timestamp time.Time
	ExtraData interface{}
}

// ConfirmProvider

type ConfirmProviderSignedParams struct {
	JSON      string                 `json:"data" coerce:"name:json"`
	Data      *ConfirmProviderParams `json:"-" coerce:"name:data"`
	Signature []byte                 `json:"signature"`
	PublicKey []byte                 `json:"publicKey"`
}

// this data is accessible to the provider, nothing "secret" should be
// stored here...
type ConfirmProviderParams struct {
	Timestamp             time.Time              `json:"timestamp"`
	PublicProviderData    *SignedProviderData    `json:"publicProviderData"`
	EncryptedProviderData *EncryptedProviderData `json:"encryptedProviderData"`
	SignedKeyData         *SignedProviderKeyData `json:"signedKeyData"`
}

type EncryptedProviderData struct {
	JSON      string                    `json:"data" coerce:"name:json"`
	Data      *crypto.ECDHEncryptedData `json:"-" coerce:"name:data"`
	Signature []byte                    `json:"signature"`
	PublicKey []byte                    `json:"publicKey"`
}

type SignedProviderKeyData struct {
	JSON      string           `json:"data" coerce:"name:json"`
	Data      *ProviderKeyData `json:"-" coerce:"name:data"`
	Signature []byte           `json:"signature"`
	PublicKey []byte           `json:"publicKey"`
}

func (k *ProviderKeyData) Sign(key *crypto.Key) (*SignedProviderKeyData, error) {
	if data, err := json.Marshal(k); err != nil {
		return nil, err
	} else if signedData, err := key.Sign(data); err != nil {
		return nil, err
	} else {
		return &SignedProviderKeyData{
			JSON:      string(data),
			Signature: signedData.Signature,
			PublicKey: signedData.PublicKey,
			Data:      k,
		}, nil
	}
}

type ProviderKeyData struct {
	Signing    []byte             `json:"signing"`
	Encryption []byte             `json:"encryption"`
	QueueData  *ProviderQueueData `json:"queueData"`
}

type ProviderQueueData struct {
	ZipCode    string `json:"zipCode"`
	Accessible bool   `json:"accessible"`
}

// ResetDB

type ResetDBSignedParams struct {
	JSON      string         `json:"data" coerce:"name:json"`
	Data      *ResetDBParams `json:"-" coerce:"name:data"`
	Signature []byte         `json:"signature"`
	PublicKey []byte         `json:"publicKey"`
}

type ResetDBParams struct {
	Timestamp time.Time `json:"timestamp"`
}

// AddMediatorPublicKeys

type AddMediatorPublicKeysSignedParams struct {
	JSON      string                       `json:"data" coerce:"name:json"`
	Data      *AddMediatorPublicKeysParams `json:"-" coerce:"name:data"`
	Signature []byte                       `json:"signature"`
	PublicKey []byte                       `json:"publicKey"`
}

type AddMediatorPublicKeysParams struct {
	Timestamp     time.Time              `json:"timestamp"`
	SignedKeyData *SignedMediatorKeyData `json:"signedKeyData"`
}

type SignedMediatorKeyData struct {
	JSON      string           `json:"data" coerce:"name:json"`
	Data      *MediatorKeyData `json:"-" coerce:"name:data"`
	Signature []byte           `json:"signature"`
	PublicKey []byte           `json:"publicKey"`
}

func (k *MediatorKeyData) Sign(key *crypto.Key) (*SignedMediatorKeyData, error) {
	if data, err := json.Marshal(k); err != nil {
		return nil, err
	} else if signedData, err := key.Sign(data); err != nil {
		return nil, err
	} else {
		return &SignedMediatorKeyData{
			JSON:      string(data),
			Signature: signedData.Signature,
			PublicKey: signedData.PublicKey,
			Data:      k,
		}, nil
	}
}

type MediatorKeyData struct {
	Signing    []byte `json:"signing"`
	Encryption []byte `json:"encryption"`
}

// AddCodes

type AddCodesParams struct {
	JSON      string     `json:"data" coerce:"name:json"`
	Data      *CodesData `json:"-" coerce:"name:data"`
	Signature []byte     `json:"signature"`
	PublicKey []byte     `json:"publicKey"`
}

type CodesData struct {
	Actor     string    `json:"actor"`
	Timestamp time.Time `json:"timestamp"`
	Codes     [][]byte  `json:"codes"`
}

// UploadDistances

type UploadDistancesSignedParams struct {
	JSON      string                 `json:"data" coerce:"name:json"`
	Data      *UploadDistancesParams `json:"-" coerce:"name:data"`
	Signature []byte                 `json:"signature"`
	PublicKey []byte                 `json:"publicKey"`
}

type UploadDistancesParams struct {
	Timestamp time.Time  `json:"timestamp"`
	Type      string     `json:"type"`
	Distances []Distance `json:"distances"`
}

type Distance struct {
	From     string  `json:"from"`
	To       string  `json:"to"`
	Distance float64 `json:"distance"`
}

// Keys

type GetKeysParams struct {
}

type Keys struct {
	ProviderData []byte `json:"providerData"`
	RootKey      []byte `json:"rootKey"`
	TokenKey     []byte `json:"tokenKey"`
}

type KeyLists struct {
	Providers []*ActorKey `json:"providers"`
	Mediators []*ActorKey `json:"mediators"`
}

type ActorKey struct {
	Data      string        `json:"data"`
	Signature []byte        `json:"signature"`
	PublicKey []byte        `json:"publicKey"`
	data      *ActorKeyData `json:"-"`
}

func (a *ActorKey) KeyData() (*ActorKeyData, error) {
	var akd *ActorKeyData
	if a.data != nil {
		return a.data, nil
	}
	if err := json.Unmarshal([]byte(a.Data), &akd); err != nil {
		return nil, err
	}
	a.data = akd
	return akd, nil
}

func (a *ActorKey) ProviderKeyData() (*ProviderKeyData, error) {
	var pkd *ProviderKeyData
	if err := json.Unmarshal([]byte(a.Data), &pkd); err != nil {
		return nil, err
	}
	if pkd.QueueData == nil {
		pkd.QueueData = &ProviderQueueData{}
	}
	return pkd, nil
}

type ActorKeyData struct {
	Encryption []byte `json:"encryption"`
	Signing    []byte `json:"signing"`
}

// GetToken

type GetTokenParams struct {
	Hash      []byte `json:"hash"`
	Code      []byte `json:"code"`
	PublicKey []byte `json:"publicKey"`
}

type SignedTokenData struct {
	JSON      string     `json:"data" coerce:"name:json"`
	Data      *TokenData `json:"-" coerce:"name:data"`
	Signature []byte     `json:"signature"`
	PublicKey []byte     `json:"publicKey"`
}

func (k *TokenData) Sign(key *crypto.Key) (*SignedTokenData, error) {
	if data, err := json.Marshal(k); err != nil {
		return nil, err
	} else if signedData, err := key.Sign(data); err != nil {
		return nil, err
	} else {
		return &SignedTokenData{
			JSON:      string(data),
			Signature: signedData.Signature,
			PublicKey: signedData.PublicKey,
			Data:      k,
		}, nil
	}
}

type PriorityToken struct {
	N int64 `json:"n"`
}

func (p *PriorityToken) Marshal() ([]byte, error) {
	if data, err := json.Marshal(p); err != nil {
		return nil, err
	} else {
		return data, nil
	}
}

type TokenData struct {
	JSON      string         `json:"data" coerce:"name:json"`
	Data      *PriorityToken `json:"-" coerce:"name:data"`
	PublicKey []byte         `json:"publicKey"`
	Token     []byte         `json:"token"`
	Hash      []byte         `json:"hash"`
}

// GetAppointmentsByZipCode

type GetAppointmentsByZipCodeParams struct {
	ZipCode string `json:"zipCode"`
	Radius  int64  `json:"radius"`
}

type KeyChain struct {
	Provider *ActorKey `json:"provider"`
	Mediator *ActorKey `json:"mediator"`
}

type ProviderAppointments struct {
	Provider *SignedProviderData  `json:"provider"`
	Offers   []*SignedAppointment `json:"offers"`
	Booked   [][]byte             `json:"booked"`
	KeyChain *KeyChain            `json:"keyChain"`
}

type SignedProviderData struct {
	JSON      string        `json:"data" coerce:"name:json"`
	Data      *ProviderData `json:"-" coerce:"name:data"`
	Signature []byte        `json:"signature"`
	PublicKey []byte        `json:"publicKey"`
	ID        []byte        `json:"id"`
}

func (k *ProviderData) Sign(key *crypto.Key) (*SignedProviderData, error) {
	if data, err := json.Marshal(k); err != nil {
		return nil, err
	} else if signedData, err := key.Sign(data); err != nil {
		return nil, err
	} else {
		return &SignedProviderData{
			JSON:      string(data),
			Signature: signedData.Signature,
			PublicKey: signedData.PublicKey,
			Data:      k,
		}, nil
	}
}

type ProviderData struct {
	Name        string `json:"name"`
	Street      string `json:"street"`
	City        string `json:"city"`
	ZipCode     string `json:"zipCode"`
	Description string `json:"description"`
}

// GetProviderAppointments

type GetProviderAppointmentsSignedParams struct {
	JSON      string                         `json:"data" coerce:"name:json"`
	Data      *GetProviderAppointmentsParams `json:"-" coerce:"name:data"`
	Signature []byte                         `json:"signature"`
	PublicKey []byte                         `json:"publicKey"`
}

type GetProviderAppointmentsParams struct {
	Timestamp    time.Time  `json:"timestamp"`
	From         time.Time  `json:"from"`
	To           time.Time  `json:"to"`
	UpdatedSince *time.Time `json:"updatedSince"`
}

// PublishAppointments

type PublishAppointmentsSignedParams struct {
	JSON      string                     `json:"data" coerce:"name:json"`
	Data      *PublishAppointmentsParams `json:"-" coerce:"name:data"`
	Signature []byte                     `json:"signature"`
	PublicKey []byte                     `json:"publicKey"`
}

type PublishAppointmentsParams struct {
	Timestamp time.Time            `json:"timestamp"`
	Offers    []*SignedAppointment `json:"offers"`
}

type SignedAppointment struct {
	UpdatedAt   time.Time    `json:"updatedAt"`
	Bookings    []*Booking   `json:"bookings"`    // only for providers
	BookedSlots []*Slot      `json:"bookedSlots"` // for users
	JSON        string       `json:"data" coerce:"name:json"`
	Data        *Appointment `json:"-" coerce:"name:data"`
	Signature   []byte       `json:"signature"`
	PublicKey   []byte       `json:"publicKey"`
}

func MakeAppointment(timestamp time.Time, slots, duration int64) (*Appointment, error) {
	if id, err := crypto.RandomBytes(32); err != nil {
		return nil, err
	} else {
		slotData := make([]*Slot, slots)
		for i, _ := range slotData {
			sd := &Slot{}
			if id, err := crypto.RandomBytes(32); err != nil {
				return nil, err
			} else {
				sd.ID = id
			}
			slotData[i] = sd
		}
		return &Appointment{
			Timestamp: timestamp,
			Duration:  duration,
			ID:        id,
			SlotData:  slotData,
		}, nil
	}
}

type Appointment struct {
	Timestamp  time.Time              `json:"timestamp"`
	Duration   int64                  `json:"duration"`
	Properties map[string]interface{} `json:"properties"`
	SlotData   []*Slot                `json:"slotData"`
	ID         []byte                 `json:"id"`
	PublicKey  []byte                 `json:"publicKey"`
}

func (k *Appointment) Sign(key *crypto.Key) (*SignedAppointment, error) {
	if data, err := json.Marshal(k); err != nil {
		return nil, err
	} else if signedData, err := key.Sign(data); err != nil {
		return nil, err
	} else {
		return &SignedAppointment{
			JSON:      string(data),
			Signature: signedData.Signature,
			PublicKey: signedData.PublicKey,
			Data:      k,
		}, nil
	}
}

type Slot struct {
	ID []byte `json:"id"`
}

// BookAppointment

type BookAppointmentSignedParams struct {
	JSON      string                 `json:"data" coerce:"name:json"`
	Data      *BookAppointmentParams `json:"-" coerce:"name:data"`
	Signature []byte                 `json:"signature"`
	PublicKey []byte                 `json:"publicKey"`
}

type BookAppointmentParams struct {
	ProviderID      []byte                    `json:"providerID"`
	ID              []byte                    `json:"id"`
	EncryptedData   *crypto.ECDHEncryptedData `json:"encryptedData"`
	SignedTokenData *SignedTokenData          `json:"signedTokenData"`
	Timestamp       time.Time                 `json:"timestamp"`
}

type Booking struct {
	ID            []byte                    `json:"id"`
	PublicKey     []byte                    `json:"publicKey"`
	Token         []byte                    `json:"token"`
	EncryptedData *crypto.ECDHEncryptedData `json:"encryptedData"`
}

// GetAppointment

type GetAppointmentSignedParams struct {
	JSON      string                `json:"data" coerce:"name:json"`
	Data      *GetAppointmentParams `json:"-" coerce:"name:data"`
	Signature []byte                `json:"signature"`
	PublicKey []byte                `json:"publicKey"`
}

type GetAppointmentParams struct {
	Timestamp       time.Time        `json:"timestamp"`
	ProviderID      []byte           `json:"providerID"`
	SignedTokenData *SignedTokenData `json:"signedTokenData"`
	ID              []byte           `json:"id"`
}

// CancelAppointment

type CancelAppointmentSignedParams struct {
	JSON      string                   `json:"data" coerce:"name:json"`
	Data      *CancelAppointmentParams `json:"-" coerce:"name:data"`
	Signature []byte                   `json:"signature"`
	PublicKey []byte                   `json:"publicKey"`
}

type CancelAppointmentParams struct {
	ProviderID      []byte           `json:"providerID"`
	SignedTokenData *SignedTokenData `json:"signedTokenData"`
	ID              []byte           `json:"id"`
}

// CheckProviderData

type CheckProviderDataSignedParams struct {
	JSON      string                   `json:"data" coerce:"name:json"`
	Data      *CheckProviderDataParams `json:"-" coerce:"name:data"`
	Signature []byte                   `json:"signature"`
	PublicKey []byte                   `json:"publicKey"`
}

type CheckProviderDataParams struct {
	Timestamp time.Time `json:"timestamp"`
}

// StoreProviderData

type StoreProviderDataSignedParams struct {
	JSON      string                   `json:"data" coerce:"name:json"`
	Data      *StoreProviderDataParams `json:"-" coerce:"name:data"`
	Signature []byte                   `json:"signature"`
	PublicKey []byte                   `json:"publicKey"`
}

type StoreProviderDataParams struct {
	Timestamp     time.Time                 `json:"timestamp"`
	EncryptedData *crypto.ECDHEncryptedData `json:"encryptedData"`
	Code          []byte                    `json:"code"`
}

type RawProviderData struct {
	ID            []byte                    `json:"id,omitempty"`
	EncryptedData *crypto.ECDHEncryptedData `json:"encryptedData"`
}

// GetPendingProviderData

type GetPendingProviderDataSignedParams struct {
	JSON      string                        `json:"data" coerce:"name:json"`
	Data      *GetPendingProviderDataParams `json:"-" coerce:"name:data"`
	Signature []byte                        `json:"signature"`
	PublicKey []byte                        `json:"publicKey"`
}

type GetPendingProviderDataParams struct {
	Timestamp time.Time `json:"timestamp"`
	Limit     int64     `json:"limit"`
}

// GetVerifiedProviderData

type GetVerifiedProviderDataSignedParams struct {
	JSON      string                         `json:"data" coerce:"name:json"`
	Data      *GetVerifiedProviderDataParams `json:"-" coerce:"name:data"`
	Signature []byte                         `json:"signature"`
	PublicKey []byte                         `json:"publicKey"`
}

type GetVerifiedProviderDataParams struct {
	Timestamp time.Time `json:"timestamp"`
	Limit     int64     `json:"limit"`
}

// GetStats

type GetStatsParams struct {
	ID     string                 `json:"id"`
	Type   string                 `json:"type"`
	Filter map[string]interface{} `json:"filter"`
	Metric string                 `json:"metric"`
	Name   string                 `json:"name"`
	From   *time.Time             `json:"from"` // optional
	To     *time.Time             `json:"to"`   // optional
	N      *int64                 `json:"n"`    // optional
}

type StatsValue struct {
	Name  string            `json:"name"`
	From  time.Time         `json:"from"`
	To    time.Time         `json:"to"`
	Data  map[string]string `json:"data"`
	Value int64             `json:"value"`
}
