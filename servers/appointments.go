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
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/kiebitz-oss/services"
	"github.com/kiebitz-oss/services/crypto"
	"github.com/kiebitz-oss/services/databases"
	kbForms "github.com/kiebitz-oss/services/forms"
	"github.com/kiebitz-oss/services/jsonrpc"
	"github.com/kiebitz-oss/services/metrics"
	"sort"
	"strings"
	"time"
)

type Appointments struct {
	server        *jsonrpc.JSONRPCServer
	metricsServer *metrics.PrometheusMetricsServer
	db            services.Database
	meter         services.Meter
	settings      *services.AppointmentsSettings
}

func MakeAppointments(settings *services.Settings) (*Appointments, error) {

	Appointments := &Appointments{
		db:       settings.DatabaseObj,
		meter:    settings.MeterObj,
		settings: settings.Appointments,
	}

	methods := map[string]*jsonrpc.Method{
		"confirmProvider": {
			Form:    &kbForms.ConfirmProviderForm,
			Handler: Appointments.confirmProvider,
		},
		"addMediatorPublicKeys": {
			Form:    &kbForms.AddMediatorPublicKeysForm,
			Handler: Appointments.addMediatorPublicKeys,
		},
		"addCodes": {
			Form:    &kbForms.AddCodesForm,
			Handler: Appointments.addCodes,
		},
		"uploadDistances": {
			Form:    &kbForms.UploadDistancesForm,
			Handler: Appointments.uploadDistances,
		},
		"getStats": {
			Form:    &kbForms.GetStatsForm,
			Handler: Appointments.getStats,
		},
		"getKeys": {
			Form:    &kbForms.GetKeysForm,
			Handler: Appointments.getKeys,
		},
		"getAppointmentsByZipCode": {
			Form:    &kbForms.GetAppointmentsByZipCodeForm,
			Handler: Appointments.getAppointmentsByZipCode,
		},
		"getProviderAppointments": {
			Form:    &kbForms.GetProviderAppointmentsForm,
			Handler: Appointments.getProviderAppointments,
		},
		"publishAppointments": {
			Form:    &kbForms.PublishAppointmentsForm,
			Handler: Appointments.publishAppointments,
		},
		"bookAppointment": {
			Form:    &kbForms.BookAppointmentForm,
			Handler: Appointments.bookAppointment,
		},
		"cancelAppointment": {
			Form:    &kbForms.CancelAppointmentForm,
			Handler: Appointments.cancelAppointment,
		},
		"getToken": {
			Form:    &kbForms.GetTokenForm,
			Handler: Appointments.getToken,
		},
		"storeProviderData": {
			Form:    &kbForms.StoreProviderDataForm,
			Handler: Appointments.storeProviderData,
		},
		"checkProviderData": {
			Form:    &kbForms.CheckProviderDataForm,
			Handler: Appointments.checkProviderData,
		},
		"getPendingProviderData": {
			Form:    &kbForms.GetPendingProviderDataForm,
			Handler: Appointments.getPendingProviderData,
		},
		"getVerifiedProviderData": {
			Form:    &kbForms.GetVerifiedProviderDataForm,
			Handler: Appointments.getVerifiedProviderData,
		},
	}

	handler, err := jsonrpc.MethodsHandler(methods)

	if err != nil {
		return nil, err
	}

	if jsonrpcServer, err := jsonrpc.MakeJSONRPCServer(settings.Appointments.RPC, handler, "appointments"); err != nil {
		return nil, err
	} else {
		Appointments.server = jsonrpcServer
		return Appointments, nil
	}
}

func (c *Appointments) Start() error {
	return c.server.Start()
}

func (c *Appointments) Stop() error {
	return c.server.Stop()
}

// Method Handlers

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

// { id, key, providerData, keyData }, keyPair
func (c *Appointments) confirmProvider(context *jsonrpc.Context, params *services.ConfirmProviderSignedParams) *jsonrpc.Response {

	success := false
	transaction, finalize, err := c.transaction(&success)

	if err != nil {
		services.Log.Error(err)
		return context.InternalError()
	}

	defer finalize()

	if resp, _ := c.isMediator(context, []byte(params.JSON), params.Signature, params.PublicKey); resp != nil {
		return resp
	}

	hash := crypto.Hash(params.Data.SignedKeyData.Data.Signing)
	keys := transaction.Map("keys", []byte("providers"))

	providerKey := &services.ActorKey{
		Data:      params.Data.SignedKeyData.JSON,
		Signature: params.Data.SignedKeyData.Signature,
		PublicKey: params.Data.SignedKeyData.PublicKey,
	}

	bd, err := json.Marshal(providerKey)
	if err != nil {
		services.Log.Error(err)
		return context.InternalError()
	}

	if err := keys.Set(hash, bd); err != nil {
		services.Log.Error(err)
		return context.InternalError()
	}

	unverifiedProviderData := transaction.Map("providerData", []byte("unverified"))
	verifiedProviderData := transaction.Map("providerData", []byte("verified"))
	checkedProviderData := transaction.Map("providerData", []byte("checked"))
	publicProviderData := transaction.Map("providerData", []byte("public"))

	oldPd, err := unverifiedProviderData.Get(params.Data.ID)

	if err != nil {
		if err == databases.NotFound {
			// maybe this provider has already been verified before...
			if oldPd, err = verifiedProviderData.Get(params.Data.ID); err != nil {
				if err == databases.NotFound {
					return context.NotFound()
				} else {
					services.Log.Error(err)
					return context.InternalError()
				}
			}
		} else {
			services.Log.Error(err)
			return context.InternalError()
		}
	}

	if err := unverifiedProviderData.Del(params.Data.ID); err != nil {
		if err != databases.NotFound {
			services.Log.Error(err)
			return context.InternalError()
		}
	}

	if err := verifiedProviderData.Set(params.Data.ID, oldPd); err != nil {
		services.Log.Error(err)
		return context.InternalError()
	}

	// we store a copy of the signed data for the provider to check
	if err := checkedProviderData.Set(params.Data.ID, []byte(params.JSON)); err != nil {
		services.Log.Error(err)
		return context.InternalError()
	}

	if params.Data.PublicProviderData != nil {
		signedData := map[string]interface{}{
			"data":      params.Data.PublicProviderData.JSON,
			"signature": params.Data.PublicProviderData.Signature,
			"publicKey": params.Data.PublicProviderData.PublicKey,
		}
		if data, err := json.Marshal(signedData); err != nil {
			services.Log.Error(err)
			return context.InternalError()
		} else if err := publicProviderData.Set(hash, data); err != nil {
			services.Log.Error(err)
			return context.InternalError()
		}
	}

	success = true

	return context.Acknowledge()
}

// { keys }, keyPair
// add the mediator key to the list of keys (only for testing)
func (c *Appointments) addMediatorPublicKeys(context *jsonrpc.Context, params *services.AddMediatorPublicKeysSignedParams) *jsonrpc.Response {
	rootKey := c.settings.Key("root")
	if rootKey == nil {
		services.Log.Error("root key missing")
		return context.InternalError()
	}
	if ok, err := rootKey.Verify(&crypto.SignedData{
		Data:      []byte(params.JSON),
		Signature: params.Signature,
	}); !ok {
		return context.Error(403, "invalid signature", nil)
	} else if err != nil {
		services.Log.Error(err)
		return context.InternalError()
	}
	if expired(params.Data.Timestamp) {
		return context.Error(410, "signature expired", nil)
	}
	hash := crypto.Hash(params.Data.Signing)
	keys := c.db.Map("keys", []byte("mediators"))
	bd, err := json.Marshal(context.Request.Params)
	if err != nil {
		services.Log.Error(err)
		return context.InternalError()
	}
	if err := keys.Set(hash, bd); err != nil {
		services.Log.Error(err)
		return context.InternalError()
	}
	if result, err := keys.Get(hash); err != nil {
		services.Log.Error(err)
		return context.InternalError()
	} else if !bytes.Equal(result, bd) {
		services.Log.Error("does not match")
		return context.InternalError()
	}
	return context.Acknowledge()
}

func (c *Appointments) addCodes(context *jsonrpc.Context, params *services.AddCodesParams) *jsonrpc.Response {
	rootKey := c.settings.Key("root")
	if rootKey == nil {
		services.Log.Error("root key missing")
		return context.InternalError()
	}
	if ok, err := rootKey.Verify(&crypto.SignedData{
		Data:      []byte(params.JSON),
		Signature: params.Signature,
	}); !ok {
		return context.Error(403, "invalid signature", nil)
	} else if err != nil {
		services.Log.Error(err)
		return context.InternalError()
	}
	if expired(params.Data.Timestamp) {
		return context.Error(410, "signature expired", nil)
	}
	codes := c.db.Set("codes", []byte(params.Data.Actor))
	for _, code := range params.Data.Codes {
		if err := codes.Add(code); err != nil {
			services.Log.Error(err)
			return context.InternalError()
		}
	}
	return context.Acknowledge()
}

func (c *Appointments) getDistance(distanceType, from, to string) (float64, error) {

	dst := c.db.Map("distances", []byte(distanceType))
	keyA := fmt.Sprintf("%s:%s", from, to)
	keyB := fmt.Sprintf("%s:%s", to, from)
	value, err := dst.Get([]byte(keyA))

	if err != nil && err != databases.NotFound {
		return 0.0, err
	}

	if value == nil {
		value, err = dst.Get([]byte(keyB))
	}

	if err != nil {
		return 0.0, err
	}

	buf := bytes.NewReader(value)
	var distance float64
	if err := binary.Read(buf, binary.LittleEndian, &distance); err != nil {
		return 0.0, err
	}

	return distance, nil

}

func (c *Appointments) uploadDistances(context *jsonrpc.Context, params *services.UploadDistancesSignedParams) *jsonrpc.Response {
	rootKey := c.settings.Key("root")
	if rootKey == nil {
		services.Log.Error("root key missing")
		return context.InternalError()
	}
	if ok, err := rootKey.Verify(&crypto.SignedData{
		Data:      []byte(params.JSON),
		Signature: params.Signature,
	}); !ok {
		return context.Error(403, "invalid signature", nil)
	} else if err != nil {
		services.Log.Error(err)
		return context.InternalError()
	}
	if expired(params.Data.Timestamp) {
		return context.Error(410, "signature expired", nil)
	}
	dst := c.db.Map("distances", []byte(params.Data.Type))
	for _, distance := range params.Data.Distances {
		neighborsFrom := c.db.SortedSet(fmt.Sprintf("distances::neighbors::%s", params.Data.Type), []byte(distance.From))
		neighborsTo := c.db.SortedSet(fmt.Sprintf("distances::neighbors::%s", params.Data.Type), []byte(distance.To))
		neighborsFrom.Add([]byte(distance.To), int64(distance.Distance))
		neighborsTo.Add([]byte(distance.From), int64(distance.Distance))
		key := fmt.Sprintf("%s:%s", distance.From, distance.To)
		buf := new(bytes.Buffer)
		if err := binary.Write(buf, binary.LittleEndian, distance.Distance); err != nil {
			services.Log.Error(err)
			return context.InternalError()
		}
		if err := dst.Set([]byte(key), buf.Bytes()); err != nil {
			services.Log.Error(err)
			return context.InternalError()
		}
	}

	return context.Acknowledge()
}

// signed requests are valid only 1 minute
func expired(timestamp *time.Time) bool {
	return time.Now().Add(-time.Minute).After(*timestamp)
}

// public endpoints

func toStringMap(data []byte) (map[string]interface{}, error) {
	var v map[string]interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, err
	}
	return v, nil
}

func toInterface(data []byte) (interface{}, error) {
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, err
	}
	return v, nil
}

func findActorKey(keys []*services.ActorKey, publicKey []byte) (*services.ActorKey, error) {
	for _, key := range keys {
		if akd, err := key.KeyData(); err != nil {
			services.Log.Error(err)
			continue
		} else if bytes.Equal(akd.Signing, publicKey) {
			return key, nil
		}
	}
	return nil, nil
}

func (c *Appointments) getListKeys(key string) ([]*services.ActorKey, error) {
	mk, err := c.db.Map("keys", []byte(key)).GetAll()

	if err != nil {
		services.Log.Error(err)
		return nil, err
	}

	actorKeys := []*services.ActorKey{}

	for _, v := range mk {
		var m *services.ActorKey
		if err := json.Unmarshal(v, &m); err != nil {
			services.Log.Error(err)
			continue
		} else {
			actorKeys = append(actorKeys, m)
		}
	}

	return actorKeys, nil

}

func (c *Appointments) getKeysData() (*services.Keys, error) {

	providerDataKey := c.settings.Key("provider")

	// to do: remove once the settings are updated
	if providerDataKey == nil {
		providerDataKey = c.settings.Key("providerData")
	}

	return &services.Keys{
		ProviderData: providerDataKey.PublicKey,
		RootKey:      c.settings.Key("root").PublicKey,
		TokenKey:     c.settings.Key("token").PublicKey,
	}, nil

}

func (c *Appointments) getActorKeys() (*services.KeyLists, error) {
	mediatorKeys, err := c.getListKeys("mediators")

	if err != nil {
		return nil, err
	}

	providerKeys, err := c.getListKeys("providers")

	if err != nil {
		return nil, err
	}

	return &services.KeyLists{
		Providers: providerKeys,
		Mediators: mediatorKeys,
	}, nil
}

// return all public keys present in the system
func (c *Appointments) getKeys(context *jsonrpc.Context, params *services.GetKeysParams) *jsonrpc.Response {

	keys, err := c.getKeysData()

	if err != nil {
		services.Log.Error(err)
		return context.InternalError()
	}

	return context.Result(keys)
}

// data endpoints

// user endpoints

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

// provider-only endpoints

var tws = []services.TimeWindowFunc{
	services.Minute,
	services.QuarterHour,
	services.Hour,
	services.Day,
	services.Week,
	services.Month,
}

/*
- Get all neighbors of the given zip code within the given radius
*/
func (c *Appointments) getAppointmentsByZipCode(context *jsonrpc.Context, params *services.GetAppointmentsByZipCodeParams) *jsonrpc.Response {

	keys, err := c.getActorKeys()

	if err != nil {
		services.Log.Error(err)
		return context.InternalError()
	}

	neighbors := c.db.SortedSet("distances::neighbors::zipCode", []byte(params.ZipCode))
	publicProviderData := c.db.Map("providerData", []byte("public"))

	allNeighbors, err := neighbors.Range(0, -1)
	if err != nil {
		services.Log.Error(err)
		return context.InternalError()
	}

	distances := map[string]int64{}

	for _, neighbor := range allNeighbors {
		distances[string(neighbor.Data)] = neighbor.Score
	}

	providerAppointmentsList := []*services.ProviderAppointments{}

	for _, providerKey := range keys.Providers {
		pkd, err := providerKey.ProviderKeyData()
		if err != nil {
			services.Log.Error(err)
			continue
		}

		if pkd.QueueData.ZipCode != params.ZipCode {
			if distance, ok := distances[pkd.QueueData.ZipCode]; !ok {
				continue
			} else if distance > params.Radius {
				continue
			}
		}

		// the provider "ID" is the hash of the signing key
		hash := crypto.Hash(pkd.Signing)

		pd, err := publicProviderData.Get(hash)

		if err != nil {
			if err != databases.NotFound {
				services.Log.Error(err)
			}
			services.Log.Info("provider data not found")
			continue
		}

		providerData := &services.SignedProviderData{}
		var providerDataMap map[string]interface{}

		if err := json.Unmarshal(pd, &providerDataMap); err != nil {
			services.Log.Error(err)
			continue
		}

		if params, err := kbForms.SignedProviderDataForm.Validate(providerDataMap); err != nil {
			services.Log.Error(err)
			continue
		} else if err := kbForms.SignedProviderDataForm.Coerce(providerData, params); err != nil {
			services.Log.Error(err)
			continue
		}

		providerData.ID = hash

		// appointments are stored in a provider-specific key
		appointmentsByID := c.db.Map("appointmentsByID", hash)
		allDates, err := appointmentsByID.GetAll()

		if err != nil {
			services.Log.Error(err)
			return context.InternalError()
		}

		signedAppointments := make([]*services.SignedAppointment, 0)

		for _, date := range allDates {
			dateKey := append(hash, date...)
			appointmentsByDate := c.db.Map("appointmentsByDate", dateKey)
			allAppointments, err := appointmentsByDate.GetAll()
			if err != nil {
				services.Log.Error(err)
				return context.InternalError()
			}

			for _, appointment := range allAppointments {
				var signedAppointment *services.SignedAppointment
				if err := json.Unmarshal(appointment, &signedAppointment); err != nil {
					services.Log.Error(err)
					continue
				}

				slots := make([]*services.Slot, len(signedAppointment.Bookings))

				for i, booking := range signedAppointment.Bookings {
					slots[i] = &services.Slot{ID: booking.ID}
				}

				// we remove the bookings as the user is not allowed to see them
				signedAppointment.Bookings = nil
				signedAppointment.BookedSlots = slots

				signedAppointments = append(signedAppointments, signedAppointment)
			}
		}

		if len(signedAppointments) == 0 {
			continue
		}

		mediatorKey, err := findActorKey(keys.Mediators, providerKey.PublicKey)

		if err != nil {
			services.Log.Error(err)
			continue
		}

		keyChain := &services.KeyChain{
			Provider: providerKey,
			Mediator: mediatorKey,
		}

		providerAppointments := &services.ProviderAppointments{
			Provider: providerData,
			Offers:   signedAppointments,
			KeyChain: keyChain,
		}

		providerAppointmentsList = append(providerAppointmentsList, providerAppointments)

	}

	return context.Result(providerAppointmentsList)
}

func (c *Appointments) getProviderAppointments(context *jsonrpc.Context, params *services.GetProviderAppointmentsSignedParams) *jsonrpc.Response {

	// make sure this is a valid provider asking for tokens
	resp, providerKey := c.isProvider(context, []byte(params.JSON), params.Signature, params.PublicKey)

	if resp != nil {
		return resp
	}

	if expired(params.Data.Timestamp) {
		return context.Error(410, "signature expired", nil)
	}

	pkd, err := providerKey.ProviderKeyData()

	if err != nil {
		services.Log.Error(err)
		return context.InternalError()
	}

	// the provider "ID" is the hash of the signing key
	hash := crypto.Hash(pkd.Signing)

	// appointments are stored in a provider-specific key
	appointmentsByID := c.db.Map("appointmentsByID", hash)
	allDates, err := appointmentsByID.GetAll()

	if err != nil {
		services.Log.Error(err)
		return context.InternalError()
	}

	signedAppointments := make([]*services.SignedAppointment, 0)

	for _, date := range allDates {
		dateKey := append(hash, date...)
		appointmentsByDate := c.db.Map("appointmentsByDate", dateKey)
		allAppointments, err := appointmentsByDate.GetAll()
		if err != nil {
			services.Log.Error(err)
			return context.InternalError()
		}

		for _, appointment := range allAppointments {
			var signedAppointment *services.SignedAppointment
			if err := json.Unmarshal(appointment, &signedAppointment); err != nil {
				services.Log.Error(err)
				continue
			}
			signedAppointments = append(signedAppointments, signedAppointment)
		}
	}

	return context.Result(signedAppointments)
}

/*
appointmentsByDate::providerID map[ID]date // maps appointments to dates
appoinments::providerID::date map[ID]appointment // contains bookings
*/

func (c *Appointments) publishAppointments(context *jsonrpc.Context, params *services.PublishAppointmentsSignedParams) *jsonrpc.Response {

	success := false
	transaction, finalize, err := c.transaction(&success)

	if err != nil {
		services.Log.Error(err)
		return context.InternalError()
	}

	defer finalize()

	// make sure this is a valid provider asking for tokens
	resp, providerKey := c.isProvider(context, []byte(params.JSON), params.Signature, params.PublicKey)

	if resp != nil {
		return resp
	}

	if expired(params.Data.Timestamp) {
		return context.Error(410, "signature expired", nil)
	}

	pkd, err := providerKey.ProviderKeyData()

	if err != nil {
		services.Log.Error(err)
		return context.InternalError()
	}

	// the provider "ID" is the hash of the signing key
	hash := crypto.Hash(pkd.Signing)
	hexUID := hex.EncodeToString(hash)

	// appointments are stored in a provider-specific key
	appointmentsByID := transaction.Map("appointmentsByID", hash)

	// appointments expire automatically after 120 days
	if err := transaction.Expire("appointments", hash, time.Hour*24*120); err != nil {
		services.Log.Error(err)
		return context.InternalError()
	}

	usedTokens := transaction.Set("bookings", []byte("tokens"))

	var bookedSlots, openSlots int64

	for _, appointment := range params.Data.Offers {

		// check if there's an existing appointment
		if date, err := appointmentsByID.Get(appointment.Data.ID); err == nil {

			if err := appointmentsByID.Del(appointment.Data.ID); err != nil {
				services.Log.Error(err)
				return context.InternalError()
			}

			appointmentsByDate := transaction.Map("appointmentsByDate", append(hash, date...))

			existingAppointment := &services.SignedAppointment{}
			var mapData map[string]interface{}

			if data, err := appointmentsByDate.Get(appointment.Data.ID); err != nil {
				services.Log.Error(err)
				return context.InternalError()
			} else if err := appointmentsByDate.Del(appointment.Data.ID); err != nil {
				services.Log.Error(err)
				return context.InternalError()
			} else if err := json.Unmarshal(data, &mapData); err != nil {
				services.Log.Error(err)
				return context.InternalError()
			} else if params, err := kbForms.AppointmentForm.Validate(mapData); err != nil {
				services.Log.Error(err)
			} else if err := kbForms.AppointmentForm.Coerce(existingAppointment, params); err != nil {
				services.Log.Error(err)
			} else {
				bookings := make([]*services.Booking, 0)
				for _, existingSlotData := range existingAppointment.Data.SlotData {
					found := false
					for _, slotData := range appointment.Data.SlotData {
						if bytes.Equal(slotData.ID, existingSlotData.ID) {
							found = true
							break
						}
					}
					if found {
						// this slot has been preserved, if there's any booking for it we migrate it
						for _, booking := range existingAppointment.Bookings {
							if bytes.Equal(booking.ID, existingSlotData.ID) {
								bookings = append(bookings, booking)
								break
							}
						}
					} else {
						// this slot has been deleted, if there's any booking for it we delete it
						for _, booking := range existingAppointment.Bookings {
							if bytes.Equal(booking.ID, existingSlotData.ID) {
								// we re-enable the associated token
								if err := usedTokens.Del(booking.Token); err != nil {
									services.Log.Error(err)
									return context.InternalError()
								}
								break
							}
						}
					}
				}
				appointment.Bookings = bookings
			}
		}

		date := []byte(appointment.Data.Timestamp.Format("2006-01-02"))
		// the hash is under our control so it's safe to concatenate it directly with the date
		dateKey := append(hash, date...)

		appointmentsByDate := transaction.Map("appointmentsByDate", dateKey)

		// appointments will auto-delete one day after their timestamp
		if err := transaction.Expire("appointmentsByDate", dateKey, appointment.Data.Timestamp.Sub(time.Now())+time.Hour*24); err != nil {
			services.Log.Error(err)
			return context.InternalError()
		}

		// ID map will auto-delete after one year (purely for storage reasons, it does not contain sensitive data)
		if err := transaction.Expire("appointmentsByID", hash, time.Hour*24*365); err != nil {
			services.Log.Error(err)
			return context.InternalError()
		}

		if err := appointmentsByID.Set(appointment.Data.ID, date); err != nil {
			services.Log.Error(err)
			return context.InternalError()
		}

		appointment.UpdatedAt = time.Now()

		if jsonData, err := json.Marshal(appointment); err != nil {
			services.Log.Error(err)
			return context.InternalError()
		} else if err := appointmentsByDate.Set(appointment.Data.ID, jsonData); err != nil {
			services.Log.Error(err)
			return context.InternalError()
		}
	}

	success = true

	if c.meter != nil {

		now := time.Now().UTC().UnixNano()

		addTokenStats := func(tw services.TimeWindow, data map[string]string) error {
			// we add the maximum of the open appointments
			if err := c.meter.AddMax("queues", "open", hexUID, data, tw, openSlots); err != nil {
				return err
			}
			// we add the maximum of the booked appointments
			if err := c.meter.AddMax("queues", "booked", hexUID, data, tw, bookedSlots); err != nil {
				return err
			}
			// we add the info that this provider is active
			if err := c.meter.AddOnce("queues", "active", hexUID, data, tw, 1); err != nil {
				return err
			}
			return nil
		}

		for _, twt := range tws {

			// generate the time window
			tw := twt(now)

			// global statistics
			if err := addTokenStats(tw, map[string]string{}); err != nil {
				services.Log.Error(err)
			}

			// statistics by zip code
			if err := addTokenStats(tw, map[string]string{
				"zipCode": pkd.QueueData.ZipCode,
			}); err != nil {
				services.Log.Error(err)
			}

		}

	}

	return context.Acknowledge()
}

func (c *Appointments) bookAppointment(context *jsonrpc.Context, params *services.BookAppointmentSignedParams) *jsonrpc.Response {

	success := false
	transaction, finalize, err := c.transaction(&success)

	if err != nil {
		services.Log.Error(err)
		return context.InternalError()
	}

	defer finalize()

	var result interface{}

	usedTokens := transaction.Set("bookings", []byte("tokens"))

	notAuthorized := context.Error(401, "not authorized", nil)

	signedData := &crypto.SignedStringData{
		Data:      params.Data.SignedTokenData.JSON,
		Signature: params.Data.SignedTokenData.Signature,
	}

	tokenKey := c.settings.Key("token")

	if ok, err := tokenKey.VerifyString(signedData); err != nil {
		services.Log.Error(err)
		return context.InternalError()
	} else if !ok {
		return context.Error(400, "invalid signature", nil)
	}

	token := params.Data.SignedTokenData.Data.Token

	if ok, err := usedTokens.Has(token); err != nil {
		services.Log.Error()
		return context.InternalError()
	} else if ok {
		return notAuthorized
	}

	// we verify the signature (without veryfing e.g. the provenance of the key)
	if ok, err := crypto.VerifyWithBytes([]byte(params.JSON), params.Signature, params.PublicKey); err != nil {
		services.Log.Errorf("Cannot verify with bytes: %s", err)
		return context.InternalError()
	} else if !ok {
		return context.Error(400, "invalid signature", nil)
	}

	appointmentsByID := c.db.Map("appointmentsByID", params.Data.ProviderID)

	if date, err := appointmentsByID.Get(params.Data.ID); err != nil {
		services.Log.Errorf("Cannot get appointment by ID: %v", err)
		return context.InternalError()
	} else {

		dateKey := append(params.Data.ProviderID, date...)
		appointmentsByDate := c.db.Map("appointmentsByDate", dateKey)

		if appointment, err := appointmentsByDate.Get(params.Data.ID); err != nil {
			services.Log.Errorf("Cannot get appointment by date: %v", err)
			return context.InternalError()
		} else {
			signedAppointment := &services.SignedAppointment{}
			var mapData map[string]interface{}
			if err := json.Unmarshal(appointment, &mapData); err != nil {
				services.Log.Error(err)
				return context.InternalError()
			} else if params, err := kbForms.AppointmentForm.Validate(mapData); err != nil {
				services.Log.Error(err)
				return context.InternalError()
			} else if err := kbForms.AppointmentForm.Coerce(signedAppointment, params); err != nil {
				services.Log.Error(err)
				return context.InternalError()
			}
			// we try to find an open slot

			foundSlot := false
			for _, slotData := range signedAppointment.Data.SlotData {

				found := false

				for _, booking := range signedAppointment.Bookings {
					if bytes.Equal(booking.ID, slotData.ID) {
						found = true
						break
					}
				}

				if found {
					continue
				}

				// this slot is open, we book it!

				booking := &services.Booking{
					PublicKey:     params.PublicKey,
					ID:            slotData.ID,
					Token:         token,
					EncryptedData: params.Data.EncryptedData,
				}

				signedAppointment.Bookings = append(signedAppointment.Bookings, booking)
				foundSlot = true

				result = booking

				break
			}

			if !foundSlot {
				return context.NotFound()
			}

			// we mark the token as used
			if err := usedTokens.Add(token); err != nil {
				services.Log.Error(err)
				return context.InternalError()
			}

			signedAppointment.UpdatedAt = time.Now()

			// we update the appointment
			if jsonData, err := json.Marshal(signedAppointment); err != nil {
				services.Log.Error(err)
				return context.InternalError()
			} else if err := appointmentsByDate.Set(signedAppointment.Data.ID, jsonData); err != nil {
				services.Log.Error(err)
				return context.InternalError()
			}

		}

	}

	if c.meter != nil {

		now := time.Now().UTC().UnixNano()

		for _, twt := range tws {

			// generate the time window
			tw := twt(now)

			// we add the info that a booking was made
			if err := c.meter.Add("queues", "bookings", map[string]string{}, tw, 1); err != nil {
				services.Log.Error(err)
			}

		}

	}

	return context.Result(result)

}

func (c *Appointments) cancelAppointment(context *jsonrpc.Context, params *services.CancelAppointmentSignedParams) *jsonrpc.Response {
	// we verify the signature (without veryfing e.g. the provenance of the key)
	if ok, err := crypto.VerifyWithBytes([]byte(params.JSON), params.Signature, params.PublicKey); err != nil {
		services.Log.Error(err)
		return context.InternalError()
	} else if !ok {
		return context.Error(400, "invalid signature", nil)
	}

	signedData := &crypto.SignedStringData{
		Data:      params.Data.SignedTokenData.JSON,
		Signature: params.Data.SignedTokenData.Signature,
	}

	tokenKey := c.settings.Key("token")

	if ok, err := tokenKey.VerifyString(signedData); err != nil {
		services.Log.Error(err)
		return context.InternalError()
	} else if !ok {
		return context.Error(400, "invalid signature", nil)
	}

	success := false
	transaction, finalize, err := c.transaction(&success)

	if err != nil {
		services.Log.Error(err)
		return context.InternalError()
	}

	defer finalize()
	appointmentsByID := c.db.Map("appointmentsByID", params.Data.ProviderID)

	if date, err := appointmentsByID.Get(params.Data.ID); err != nil {
		services.Log.Errorf("Cannot get appointment by ID: %v", err)
		return context.InternalError()
	} else {

		dateKey := append(params.Data.ProviderID, date...)
		appointmentsByDate := c.db.Map("appointmentsByDate", dateKey)

		if appointment, err := appointmentsByDate.Get(params.Data.ID); err != nil {
			services.Log.Errorf("Cannot get appointment by date: %v", err)
			return context.InternalError()
		} else {
			signedAppointment := &services.SignedAppointment{}
			var mapData map[string]interface{}
			if err := json.Unmarshal(appointment, &mapData); err != nil {
				services.Log.Error(err)
				return context.InternalError()
			} else if params, err := kbForms.AppointmentForm.Validate(mapData); err != nil {
				services.Log.Error(err)
				return context.InternalError()
			} else if err := kbForms.AppointmentForm.Coerce(signedAppointment, params); err != nil {
				services.Log.Error(err)
				return context.InternalError()
			}

			// we try to find an open slot

			newBookings := make([]*services.Booking, 0)

			token := params.Data.SignedTokenData.Data.Token

			found := false
			for _, booking := range signedAppointment.Bookings {
				if bytes.Equal(booking.Token, token) {
					found = true
					continue
				}
				newBookings = append(newBookings, booking)
			}

			if !found {
				return context.NotFound()
			}

			signedAppointment.Bookings = newBookings

			usedTokens := transaction.Set("bookings", []byte("tokens"))

			// we mark the token as unused
			if err := usedTokens.Del(token); err != nil {
				services.Log.Error(err)
				return context.InternalError()
			}

			signedAppointment.UpdatedAt = time.Now()

			// we update the appointment
			if jsonData, err := json.Marshal(signedAppointment); err != nil {
				services.Log.Error(err)
				return context.InternalError()
			} else if err := appointmentsByDate.Set(signedAppointment.Data.ID, jsonData); err != nil {
				services.Log.Error(err)
				return context.InternalError()
			}

		}

	}

	return context.Acknowledge()

}

// get provider data

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

// store provider data

func (c *Appointments) transaction(success *bool) (services.Transaction, func(), error) {
	transaction, err := c.db.Begin()

	if err != nil {
		return nil, nil, err
	}

	finalize := func() {
		if *success {
			if err := transaction.Commit(); err != nil {
				services.Log.Error(err)
			}
		} else {
			if err := transaction.Rollback(); err != nil {
				services.Log.Error(err)
			}
		}
	}

	return transaction, finalize, nil

}

// { id, encryptedData, code }, keyPair
func (c *Appointments) storeProviderData(context *jsonrpc.Context, params *services.StoreProviderDataSignedParams) *jsonrpc.Response {

	// we verify the signature (without veryfing e.g. the provenance of the key)
	if ok, err := crypto.VerifyWithBytes([]byte(params.JSON), params.Signature, params.PublicKey); err != nil {
		services.Log.Error(err)
		return context.InternalError()
	} else if !ok {
		return context.Error(400, "invalid signature", nil)
	}

	success := false
	transaction, finalize, err := c.transaction(&success)

	if err != nil {
		services.Log.Error(err)
		return context.InternalError()
	}

	defer finalize()

	verifiedProviderData := transaction.Map("providerData", []byte("verified"))
	providerData := transaction.Map("providerData", []byte("unverified"))
	codes := transaction.Set("codes", []byte("provider"))
	codeScores := c.db.SortedSet("codeScores", []byte("provider"))

	hash := crypto.Hash(params.PublicKey)

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

	if err := providerData.Set(hash, []byte(params.JSON)); err != nil {
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

	success = true

	return context.Acknowledge()
}

func (c *Appointments) isMediator(context *jsonrpc.Context, data, signature, publicKey []byte) (*jsonrpc.Response, *services.ActorKey) {

	keys, err := c.getActorKeys()

	if err != nil {
		services.Log.Error(err)
		return context.InternalError(), nil
	}

	return c.isOnKeyList(context, data, signature, publicKey, keys.Mediators)
}

func (c *Appointments) isProvider(context *jsonrpc.Context, data, signature, publicKey []byte) (*jsonrpc.Response, *services.ActorKey) {

	keys, err := c.getActorKeys()

	if err != nil {
		services.Log.Error(err)
		return context.InternalError(), nil
	}

	return c.isOnKeyList(context, data, signature, publicKey, keys.Providers)
}

func (c *Appointments) isOnKeyList(context *jsonrpc.Context, data, signature, publicKey []byte, keyList []*services.ActorKey) (*jsonrpc.Response, *services.ActorKey) {

	actorKey, err := findActorKey(keyList, publicKey)

	if err != nil {
		services.Log.Error(err)
		return context.InternalError(), nil
	}

	if actorKey == nil {
		return context.Error(403, "not authorized", nil), nil
	}

	if ok, err := crypto.VerifyWithBytes(data, signature, publicKey); err != nil {
		services.Log.Error(err)
		return context.InternalError(), nil
	} else if !ok {
		return context.Error(401, "invalid signature", nil), nil
	}

	return nil, actorKey

}

// mediator-only endpoint
// { limit }, keyPair
func (c *Appointments) getVerifiedProviderData(context *jsonrpc.Context, params *services.GetVerifiedProviderDataSignedParams) *jsonrpc.Response {

	if resp, _ := c.isMediator(context, []byte(params.JSON), params.Signature, params.PublicKey); resp != nil {
		return resp
	}

	providerData := c.db.Map("providerData", []byte("verified"))

	pd, err := providerData.GetAll()

	if err != nil {
		services.Log.Error(err)
		return context.InternalError()
	}

	pdEntries := []map[string]interface{}{}

	for _, v := range pd {
		var m map[string]interface{}
		if err := json.Unmarshal(v, &m); err != nil {
			services.Log.Error(err)
			continue
		} else {
			pdEntries = append(pdEntries, m)
		}
	}

	return context.Result(pdEntries)

}

// mediator-only endpoint
// { limit }, keyPair
func (c *Appointments) getPendingProviderData(context *jsonrpc.Context, params *services.GetPendingProviderDataSignedParams) *jsonrpc.Response {

	if resp, _ := c.isMediator(context, []byte(params.JSON), params.Signature, params.PublicKey); resp != nil {
		return resp
	}

	providerData := c.db.Map("providerData", []byte("unverified"))

	pd, err := providerData.GetAll()

	if err != nil {
		services.Log.Error(err)
		return context.InternalError()
	}

	pdEntries := []map[string]interface{}{}

	for k, v := range pd {
		var m map[string]interface{}
		if err := json.Unmarshal(v, &m); err != nil {
			services.Log.Error(err)
			continue
		} else {
			m["id"] = []byte(k)
			pdEntries = append(pdEntries, m)
		}
	}

	return context.Result(pdEntries)

}

// mediator-only endpoint

// public endpoint
func (c *Appointments) getStats(context *jsonrpc.Context, params *services.GetStatsParams) *jsonrpc.Response {

	if c.meter == nil {
		return context.InternalError()
	}

	toTime := time.Now().UTC().UnixNano()

	var metrics []*services.Metric
	var err error

	if params.N != nil {
		metrics, err = c.meter.N(params.ID, toTime, *params.N, params.Name, params.Type)
	} else {
		metrics, err = c.meter.Range(params.ID, params.From.UnixNano(), params.To.UnixNano(), params.Name, params.Type)
	}

	if err != nil {
		services.Log.Error(err)
		return context.InternalError()
	}

	values := make([]*services.StatsValue, 0)

addMetric:
	for _, metric := range metrics {
		if params.Metric != "" && metric.Name != params.Metric {
			continue
		}
		if metric.Name[0] == '_' {
			// we skip internal metrics (which start with a '_')
			continue
		}

		if params.Filter != nil {
			for k, v := range params.Filter {
				// if v is nil we only return metrics without a value for the given key
				if v == nil {
					if _, ok := metric.Data[k]; ok {
						continue addMetric
					}
				} else if dv, ok := metric.Data[k]; !ok || dv != v {
					// filter value is missing or does not match
					continue addMetric
				}
			}
		}

		values = append(values, &services.StatsValue{
			From:  time.Unix(metric.TimeWindow.From/1e9, metric.TimeWindow.From%1e9).UTC(),
			To:    time.Unix(metric.TimeWindow.To/1e9, metric.TimeWindow.From%1e9).UTC(),
			Name:  metric.Name,
			Value: metric.Value,
			Data:  metric.Data,
		})
	}

	// we store the statistics
	sortableValues := Values{values: values}
	sort.Sort(sortableValues)

	return context.Result(values)
}

type Values struct {
	values []*services.StatsValue
}

func (f Values) Len() int {
	return len(f.values)
}

func (f Values) Less(i, j int) bool {
	r := (f.values[i].From).Sub(f.values[j].From)
	if r < 0 {
		return true
	}
	// if the from times match we compare the names
	if r == 0 {
		if strings.Compare(f.values[i].Name, f.values[j].Name) < 0 {
			return true
		}
	}
	return false
}

func (f Values) Swap(i, j int) {
	f.values[i], f.values[j] = f.values[j], f.values[i]

}
