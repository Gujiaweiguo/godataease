package engine

type Engine struct {
	ID            int64   `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name          string  `gorm:"column:name" json:"name"`
	Type          string  `gorm:"column:type" json:"type"`
	Configuration *string `gorm:"column:configuration" json:"configuration"`
	CreateBy      *string `gorm:"column:create_by" json:"createBy"`
	CreateTime    *int64  `gorm:"column:create_time" json:"createTime"`
}

func (Engine) TableName() string {
	return "core_engine"
}

type EngineDTO struct {
	ID            int64   `json:"id"`
	Name          string  `json:"name"`
	Type          string  `json:"type"`
	Configuration *string `json:"configuration"`
}

type ValidateRequest struct {
	ID            *int64  `json:"id"`
	Type          *string `json:"type"`
	Configuration *string `json:"configuration"`
}

type ValidateResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}
