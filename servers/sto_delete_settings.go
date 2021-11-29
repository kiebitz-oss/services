package servers

import (
	"github.com/kiebitz-oss/services"
	"github.com/kiebitz-oss/services/jsonrpc"
)

func (c *Storage) deleteSettings(context *jsonrpc.Context, params *services.GetSettingsParams) *jsonrpc.Response {
	value := c.db.Value("settings", params.ID)
	if err := value.Del(); err != nil {
		services.Log.Error(err)
		return context.InternalError()
	}
	return context.Acknowledge()
}
