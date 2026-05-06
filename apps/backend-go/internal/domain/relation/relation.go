package relation

type RelationDTO struct {
	ID          int64          `json:"id"`
	Name        string         `json:"name"`
	Auths       string         `json:"auths"`
	Type        string         `json:"type"`
	Creator     string         `json:"creator"`
	UpdateTime  int64          `json:"updateTime"`
	SubRelation []*RelationDTO `json:"subRelation"`
}

type RelationResponse struct {
	ID           int64          `json:"id"`
	BusiFlag     string         `json:"busiFlag"`
	RelationList []*RelationDTO `json:"relationList"`
}

type CheckPermissionResponse struct {
	ID        int64 `json:"id"`
	Editable  bool  `json:"editable"`
	Creatable bool  `json:"creatable"`
}
