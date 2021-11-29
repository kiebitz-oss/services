package servers

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/kiebitz-oss/services"
	"github.com/kiebitz-oss/services/crypto"
	"github.com/kiebitz-oss/services/jsonrpc"
)

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
