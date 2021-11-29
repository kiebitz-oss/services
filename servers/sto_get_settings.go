package servers

import (
	"encoding/json"
	"github.com/kiebitz-oss/services"
	"github.com/kiebitz-oss/services/databases"
	"github.com/kiebitz-oss/services/jsonrpc"
)

func toInterface(data []byte) (interface{}, error) {
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, err
	}
	return v, nil
}

func (c *Storage) getSettings(context *jsonrpc.Context, params *services.GetSettingsParams) *jsonrpc.Response {
	value := c.db.Value("settings", params.ID)
	if data, err := value.Get(); err != nil {
		if err == databases.NotFound {
			return context.NotFound()
		} else {
			services.Log.Error(err)
			return context.InternalError()
		}
	} else if i, err := toInterface(data); err != nil {
		services.Log.Error(err)
		return context.InternalError()
	} else {
		return context.Result(i)
	}
}
