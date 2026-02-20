package system

type SettingItem struct {
	Pkey string `json:"pkey" binding:"required"`
	Pval string `json:"pval" binding:"required"`
	Type string `json:"type" binding:"required"`
	Sort int    `json:"sort" binding:"required"`
}

type OnlineMapEditor struct {
	MapType      string `json:"mapType" binding:"required"`
	Key          string `json:"key" binding:"required"`
	SecurityCode string `json:"securityCode"`
}

type SQLBotConfig struct {
	Domain  string `json:"domain"`
	ID      string `json:"id"`
	Enabled bool   `json:"enabled"`
	Valid   bool   `json:"valid"`
}

type ShareBase struct {
	Disable   bool `json:"disable"`
	PERequire bool `json:"peRequire"`
}
