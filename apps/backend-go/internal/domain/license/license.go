package license

type LicenseRequest struct {
	License string `json:"license"`
}

type LicenseInfo struct {
	Corporation string `json:"corporation"`
	Expired     string `json:"expired"`
	Count       int64  `json:"count"`
	Version     string `json:"version"`
	Edition     string `json:"edition"`
	SerialNo    string `json:"serialNo"`
	Remark      string `json:"remark"`
	ISV         string `json:"isv"`
}

type ValidateResult struct {
	Status  string       `json:"status"`
	Message string       `json:"message,omitempty"`
	License *LicenseInfo `json:"license,omitempty"`
}

type ExpiryWarning struct {
	IsExpiringSoon bool   `json:"isExpiringSoon"`
	DaysRemaining  int    `json:"daysRemaining"`
	ExpiredDate    string `json:"expiredDate"`
	WarningLevel   string `json:"warningLevel"` // "none", "info", "warning", "critical"
	Message        string `json:"message"`
}
