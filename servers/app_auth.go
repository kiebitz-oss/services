package servers

import (
	"bytes"
	"github.com/kiebitz-oss/services"
	"github.com/kiebitz-oss/services/crypto"
	"time"
)

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

func (c *Storage) isRoot(context services.Context, params *services.SignedParams) services.Response {
	return isRoot(context, []byte(params.JSON), params.Signature, params.Timestamp, c.settings.Keys)
}

func isRoot(context services.Context, data, signature []byte, timestamp *time.Time, keys []*crypto.Key) services.Response {
	rootKey := services.Key(keys, "root")
	if rootKey == nil {
		services.Log.Error("root key missing")
		return context.InternalError()
	}
	if ok, err := rootKey.Verify(&crypto.SignedData{
		Data:      data,
		Signature: signature,
	}); !ok {
		return context.Error(403, "invalid signature", nil)
	} else if err != nil {
		services.Log.Error(err)
		return context.InternalError()
	}
	if expired(timestamp) {
		return context.Error(410, "signature expired", nil)
	}
	return nil
}
