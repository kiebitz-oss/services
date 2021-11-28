package services

import (
	"encoding/json"
	"time"
)

// ConfirmProvider

type ConfirmProviderSignedParams struct {
	JSON      string                 `json:"json"`
	Data      *ConfirmProviderParams `json:"data"`
	Signature []byte                 `json:"signature"`
	PublicKey []byte                 `json:"publicKey"`
}

// this data is accessible to the provider, nothing "secret" should be
// stored here...
type ConfirmProviderParams struct {
	ID                    []byte              `json:"id"`
	PublicProviderData    *SignedProviderData `json:"publicProviderData"`
	EncryptedProviderData *ECDHEncryptedData  `json:"encryptedProviderData"`
	SignedKeyData         *SignedKeyData      `json:"signedKeyData"`
}

type SignedKeyData struct {
	JSON      string   `json:"json"`
	Data      *KeyData `json:"data"`
	Signature []byte   `json:"signature"`
	PublicKey []byte   `json:"publicKey"`
}

type KeyData struct {
	Signing    []byte             `json:"signing"`
	Encryption []byte             `json:"encryption"`
	QueueData  *ProviderQueueData `json:"queueData"`
}

type ProviderQueueData struct {
	ZipCode    string `json:"zipCode"`
	Accessible bool   `json:"accessible"`
}

// AddMediatorPublicKeys

type AddMediatorPublicKeysSignedParams struct {
	JSON      string                       `json:"json"`
	Data      *AddMediatorPublicKeysParams `json:"data"`
	Signature []byte                       `json:"signature"`
	PublicKey []byte                       `json:"publicKey"`
}

type AddMediatorPublicKeysParams struct {
	Timestamp  *time.Time `json:"timestamp"`
	Encryption []byte     `json:"encryption"`
	Signing    []byte     `json:"signing"`
}

// AddCodes

type AddCodesParams struct {
	JSON      string     `json:"json"`
	Data      *CodesData `json:"data"`
	Signature []byte     `json:"signature"`
	PublicKey []byte     `json:"publicKey"`
}

type CodesData struct {
	Actor     string     `json:"actor"`
	Timestamp *time.Time `json:"timestamp"`
	Codes     [][]byte   `json:"codes"`
}

// UploadDistances

type UploadDistancesSignedParams struct {
	JSON      string                 `json:"json"`
	Data      *UploadDistancesParams `json:"data"`
	Signature []byte                 `json:"signature"`
	PublicKey []byte                 `json:"publicKey"`
}

type UploadDistancesParams struct {
	Timestamp *time.Time `json:"timestamp"`
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
	Encryption []byte     `json:"encryption"`
	Signing    []byte     `json:"signing"`
	Timestamp  *time.Time `json:"timestamp"`
}

type ProviderKeyData struct {
	Encryption []byte             `json:"encryption"`
	Signing    []byte             `json:"signing"`
	QueueData  *ProviderQueueData `json:"queueData"`
	Timestamp  *time.Time         `json:"timestamp,omitempty"`
}

// GetToken

type GetTokenParams struct {
	Hash      []byte `json:"hash"`
	Code      []byte `json:"code"`
	PublicKey []byte `json:"publicKey"`
}

type SignedTokenData struct {
	JSON      string     `json:"json"`
	Data      *TokenData `json:"data"`
	Signature []byte     `json:"signature"`
	PublicKey []byte     `json:"publicKey"`
}

type TokenData struct {
	PublicKey []byte `json:"publicKey"`
	Token     []byte `json:"token"`
	Hash      []byte `json:"hash"`
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
	ID        []byte        `json:"id"`
	JSON      string        `json:"data" coerce:"name:json"`
	Data      *ProviderData `json:"-" coerce:"name:data"`
	Signature []byte        `json:"signature"`
	PublicKey []byte        `json:"publicKey"`
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
	JSON      string                         `json:"json"`
	Data      *GetProviderAppointmentsParams `json:"data"`
	Signature []byte                         `json:"signature"`
	PublicKey []byte                         `json:"publicKey"`
}

type GetProviderAppointmentsParams struct {
	Timestamp *time.Time `json:"timestamp"`
}

// PublishAppointments

type PublishAppointmentsSignedParams struct {
	JSON      string                     `json:"json"`
	Data      *PublishAppointmentsParams `json:"data"`
	Signature []byte                     `json:"signature"`
	PublicKey []byte                     `json:"publicKey"`
}

type PublishAppointmentsParams struct {
	Timestamp *time.Time           `json:"timestamp"`
	Offers    []*SignedAppointment `json:"offers"`
	Reset     bool                 `json:"reset"`
}

type SignedAppointment struct {
	JSON      string       `json:"data" coerce:"name:json"`
	Data      *Appointment `json:"-" coerce:"name:data"`
	Signature []byte       `json:"signature"`
	PublicKey []byte       `json:"publicKey"`
}

type Appointment struct {
	UpdatedAt  time.Time              `json:"updatedAt"`
	Timestamp  time.Time              `json:"timestamp"`
	Duration   int64                  `json:"duration"`
	Properties map[string]interface{} `json:"properties"`
	SlotData   []*Slot                `json:"slotData"`
	ID         []byte                 `json:"id"`
	PublicKey  []byte                 `json:"publicKey"`
}

type Slot struct {
	ID []byte `json:"id"`
}

// GetBookedAppointments

type GetBookedAppointmentsSignedParams struct {
	JSON      string                       `json:"json"`
	Data      *GetBookedAppointmentsParams `json:"data"`
	Signature []byte                       `json:"signature"`
	PublicKey []byte                       `json:"publicKey"`
}

type GetBookedAppointmentsParams struct {
	Timestamp *time.Time `json:"timestamp"`
}

// CancelBooking

type CancelBookingSignedParams struct {
	JSON      string               `json:"json"`
	Data      *CancelBookingParams `json:"data"`
	Signature []byte               `json:"signature"`
	PublicKey []byte               `json:"publicKey"`
}

type CancelBookingParams struct {
	Timestamp *time.Time `json:"timestamp"`
	ID        []byte     `json:"id"`
}

// BookSlot

type BookSlotSignedParams struct {
	JSON      string          `json:"json"`
	Data      *BookSlotParams `json:"data"`
	Signature []byte          `json:"signature"`
	PublicKey []byte          `json:"publicKey"`
}

type BookSlotParams struct {
	ProviderID      []byte             `json:"providerID"`
	ID              []byte             `json:"id"`
	EncryptedData   *ECDHEncryptedData `json:"encryptedData"`
	SignedTokenData *SignedTokenData   `json:"signedTokenData"`
	Timestamp       *time.Time         `json:"timestamp"`
}

type Booking struct {
	ID            []byte             `json:"id"`
	PublicKey     []byte             `json:"publicKey"`
	Token         []byte             `json:"token"`
	EncryptedData *ECDHEncryptedData `json:"encryptedData"`
}

// CancelSlot

type CancelSlotSignedParams struct {
	JSON      string            `json:"json"`
	Data      *CancelSlotParams `json:"data"`
	Signature []byte            `json:"signature"`
	PublicKey []byte            `json:"publicKey"`
}

type CancelSlotParams struct {
	ProviderID      []byte           `json:"providerID"`
	SignedTokenData *SignedTokenData `json:"signedTokenData"`
	ID              []byte           `json:"id"`
}

// CheckProviderData

type CheckProviderDataSignedParams struct {
	JSON      string                   `json:"json"`
	Data      *CheckProviderDataParams `json:"data"`
	Signature []byte                   `json:"signature"`
	PublicKey []byte                   `json:"publicKey"`
}

type CheckProviderDataParams struct {
	Timestamp *time.Time `json:"timestamp"`
}

// StoreProviderData

type StoreProviderDataSignedParams struct {
	JSON      string                   `json:"json"`
	Data      *StoreProviderDataParams `json:"data"`
	Signature []byte                   `json:"signature"`
	PublicKey []byte                   `json:"publicKey"`
}

type StoreProviderDataParams struct {
	EncryptedData *ECDHEncryptedData `json:"encryptedData"`
	Code          []byte             `json:"code"`
}

// GetPendingProviderData

type GetPendingProviderDataSignedParams struct {
	JSON      string                        `json:"json"`
	Data      *GetPendingProviderDataParams `json:"data"`
	Signature []byte                        `json:"signature"`
	PublicKey []byte                        `json:"publicKey"`
}

type GetPendingProviderDataParams struct {
	N int64 `json:"n"`
}

// GetVerifiedProviderData

type GetVerifiedProviderDataSignedParams struct {
	JSON      string                         `json:"json"`
	Data      *GetVerifiedProviderDataParams `json:"data"`
	Signature []byte                         `json:"signature"`
	PublicKey []byte                         `json:"publicKey"`
}

type GetVerifiedProviderDataParams struct {
	N int64 `json:"n"`
}

// GetStats

type GetStatsParams struct {
	ID     string                 `json:"id"`
	Type   string                 `json:"type"`
	Filter map[string]interface{} `json:"filter"`
	Metric string                 `json:"metric"`
	Name   string                 `json:"name"`
	From   *time.Time             `json:"from"`
	To     *time.Time             `json:"to"`
	N      *int64                 `json:"n"`
}

type StatsValue struct {
	Name  string            `json:"name"`
	From  time.Time         `json:"from"`
	To    time.Time         `json:"to"`
	Data  map[string]string `json:"data"`
	Value int64             `json:"value"`
}
