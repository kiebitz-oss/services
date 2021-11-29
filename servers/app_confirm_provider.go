package servers

import (
	"encoding/json"
	"github.com/kiebitz-oss/services"
	"github.com/kiebitz-oss/services/crypto"
	"github.com/kiebitz-oss/services/databases"
	"github.com/kiebitz-oss/services/jsonrpc"
)

// { id, key, providerData, keyData }, keyPair
func (c *Appointments) confirmProvider(context *jsonrpc.Context, params *services.ConfirmProviderSignedParams) *jsonrpc.Response {

	if resp, _ := c.isMediator(context, []byte(params.JSON), params.Signature, params.PublicKey); resp != nil {
		return resp
	}

	hash := crypto.Hash(params.Data.SignedKeyData.Data.Signing)

	/*
		lock, err := c.db.Lock("bookAppointment_" + string(hash[:]))
		if err != nil {
			services.Log.Error(err)
			return context.InternalError()
		}

		defer lock.Release()
	*/

	keys := c.db.Map("keys", []byte("providers"))

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

	unverifiedProviderData := c.db.Map("providerData", []byte("unverified"))
	verifiedProviderData := c.db.Map("providerData", []byte("verified"))
	checkedProviderData := c.db.Map("providerData", []byte("checked"))
	publicProviderData := c.db.Map("providerData", []byte("public"))

	oldPd, err := unverifiedProviderData.Get(hash)

	if err != nil {
		if err == databases.NotFound {
			// maybe this provider has already been verified before...
			if oldPd, err = verifiedProviderData.Get(hash); err != nil {
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

	if err := unverifiedProviderData.Del(hash); err != nil {
		if err != databases.NotFound {
			services.Log.Error(err)
			return context.InternalError()
		}
	}

	if err := verifiedProviderData.Set(hash, oldPd); err != nil {
		services.Log.Error(err)
		return context.InternalError()
	}

	// we store a copy of the signed data for the provider to check
	if err := checkedProviderData.Set(hash, []byte(params.JSON)); err != nil {
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

	return context.Acknowledge()
}
