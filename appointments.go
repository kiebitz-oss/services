package services

import (
	"encoding/json"
	"time"
)

// ConfirmProvider

type ConfirmProviderParams struct {
	JSON      string               `json:"json"`
	Data      *ConfirmProviderData `json:"data"`
	Signature []byte               `json:"signature"`
	PublicKey []byte               `json:"publicKey"`
}

// this data is accessible to the provider, nothing "secret" should be
// stored here...
type ConfirmProviderData struct {
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

type AddMediatorPublicKeysParams struct {
	JSON      string                     `json:"json"`
	Data      *AddMediatorPublicKeysData `json:"data"`
	Signature []byte                     `json:"signature"`
	PublicKey []byte                     `json:"publicKey"`
}

type AddMediatorPublicKeysData struct {
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

type UploadDistancesParams struct {
	JSON      string         `json:"json"`
	Data      *DistancesData `json:"data"`
	Signature []byte         `json:"signature"`
	PublicKey []byte         `json:"publicKey"`
}

type DistancesData struct {
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

type GetProviderAppointmentsParams struct {
	JSON      string                       `json:"json"`
	Data      *GetProviderAppointmentsData `json:"data"`
	Signature []byte                       `json:"signature"`
	PublicKey []byte                       `json:"publicKey"`
}

type GetProviderAppointmentsData struct {
	Timestamp *time.Time `json:"timestamp"`
}

// PublishAppointments

type PublishAppointmentsParams struct {
	JSON      string                   `json:"json"`
	Data      *PublishAppointmentsData `json:"data"`
	Signature []byte                   `json:"signature"`
	PublicKey []byte                   `json:"publicKey"`
}

type PublishAppointmentsData struct {
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

type GetBookedAppointmentsParams struct {
	JSON      string                     `json:"json"`
	Data      *GetBookedAppointmentsData `json:"data"`
	Signature []byte                     `json:"signature"`
	PublicKey []byte                     `json:"publicKey"`
}

type GetBookedAppointmentsData struct {
	Timestamp *time.Time `json:"timestamp"`
}

// CancelBooking

type CancelBookingParams struct {
	JSON      string             `json:"json"`
	Data      *CancelBookingData `json:"data"`
	Signature []byte             `json:"signature"`
	PublicKey []byte             `json:"publicKey"`
}

type CancelBookingData struct {
	Timestamp *time.Time `json:"timestamp"`
	ID        []byte     `json:"id"`
}

// BookSlot

type BookSlotParams struct {
	JSON      string        `json:"json"`
	Data      *BookSlotData `json:"data"`
	Signature []byte        `json:"signature"`
	PublicKey []byte        `json:"publicKey"`
}

type BookSlotData struct {
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

type CancelSlotParams struct {
	JSON      string          `json:"json"`
	Data      *CancelSlotData `json:"data"`
	Signature []byte          `json:"signature"`
	PublicKey []byte          `json:"publicKey"`
}

type CancelSlotData struct {
	ProviderID      []byte           `json:"providerID"`
	SignedTokenData *SignedTokenData `json:"signedTokenData"`
	ID              []byte           `json:"id"`
}

// CheckProviderData

type CheckProviderDataParams struct {
	JSON      string                 `json:"json"`
	Data      *CheckProviderDataData `json:"data"`
	Signature []byte                 `json:"signature"`
	PublicKey []byte                 `json:"publicKey"`
}

type CheckProviderDataData struct {
	Timestamp *time.Time `json:"timestamp"`
}

// StoreProviderData

type StoreProviderDataParams struct {
	JSON      string                 `json:"json"`
	Data      *StoreProviderDataData `json:"data"`
	Signature []byte                 `json:"signature"`
	PublicKey []byte                 `json:"publicKey"`
}

type StoreProviderDataData struct {
	EncryptedData *ECDHEncryptedData `json:"encryptedData"`
	Code          []byte             `json:"code"`
}

// GetPendingProviderData

type GetPendingProviderDataParams struct {
	JSON      string                      `json:"json"`
	Data      *GetPendingProviderDataData `json:"data"`
	Signature []byte                      `json:"signature"`
	PublicKey []byte                      `json:"publicKey"`
}

type GetPendingProviderDataData struct {
	N int64 `json:"n"`
}

// GetVerifiedProviderData

type GetVerifiedProviderDataParams struct {
	JSON      string                       `json:"json"`
	Data      *GetVerifiedProviderDataData `json:"data"`
	Signature []byte                       `json:"signature"`
	PublicKey []byte                       `json:"publicKey"`
}

type GetVerifiedProviderDataData struct {
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
