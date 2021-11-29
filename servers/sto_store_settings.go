package servers

import (
	"encoding/json"
	"github.com/kiebitz-oss/services"
	"github.com/kiebitz-oss/services/jsonrpc"
	"time"
)

// store the settings in the database by ID
func (c *Storage) storeSettings(context *jsonrpc.Context, params *services.StoreSettingsParams) *jsonrpc.Response {
	value := c.db.Value("settings", params.ID)
	if dv, err := json.Marshal(params.Data); err != nil {
		services.Log.Error(err)
		return context.InternalError()
	} else if err := value.Set(dv, time.Duration(c.settings.SettingsTTLDays*24)*time.Hour); err != nil {
		return context.InternalError()
	}
	return context.Acknowledge()
}
