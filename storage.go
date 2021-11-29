package services

type DeleteSettingsParams struct {
	ID []byte `json:"id"`
}

type GetSettingsParams struct {
	ID []byte `json:"id"`
}

type StoreSettingsParams struct {
	ID   []byte      `json:"id"`
	Data interface{} `json:"data"`
}
