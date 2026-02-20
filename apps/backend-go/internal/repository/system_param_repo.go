package repository

import (
	"errors"
	"strconv"
	"strings"

	"dataease/backend/internal/domain/system"

	"gorm.io/gorm"
)

const (
	basicPrefix  = "basic."
	mapPrefix    = "map."
	sqlbotPrefix = "sqlbot."
)

type coreSysSetting struct {
	ID   int64  `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Pkey string `gorm:"column:pkey" json:"pkey"`
	Pval string `gorm:"column:pval" json:"pval"`
	Type string `gorm:"column:type" json:"type"`
	Sort int    `gorm:"column:sort" json:"sort"`
}

func (coreSysSetting) TableName() string {
	return "core_sys_setting"
}

type SystemParamRepository struct {
	db *gorm.DB
}

func NewSystemParamRepository(db *gorm.DB) *SystemParamRepository {
	return &SystemParamRepository{db: db}
}

func (r *SystemParamRepository) ListBasicSettings() ([]system.SettingItem, error) {
	list, err := r.listByPrefix(basicPrefix)
	if err != nil {
		return nil, err
	}
	result := make([]system.SettingItem, 0, len(list))
	for _, item := range list {
		result = append(result, system.SettingItem{
			Pkey: item.Pkey,
			Pval: item.Pval,
			Type: item.Type,
			Sort: item.Sort,
		})
	}
	return result, nil
}

func (r *SystemParamRepository) SaveBasicSettings(items []system.SettingItem) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, item := range items {
			pkey := item.Pkey
			if !strings.HasPrefix(pkey, basicPrefix) {
				pkey = basicPrefix + pkey
			}
			tp := strings.TrimSpace(item.Type)
			if tp == "" {
				tp = "text"
			}
			sort := item.Sort
			if sort <= 0 {
				sort = 1
			}
			if err := upsertByPkey(tx, pkey, item.Pval, tp, sort); err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *SystemParamRepository) GetOnlineMap() (*system.OnlineMapEditor, error) {
	mt, err := r.singleVal(mapPrefix + "mapType")
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(mt) == "" {
		mt = "gaode"
	}
	return r.GetOnlineMapByType(mt)
}

func (r *SystemParamRepository) GetOnlineMapByType(mapType string) (*system.OnlineMapEditor, error) {
	mt := strings.TrimSpace(mapType)
	if mt == "" {
		mt = "gaode"
	}
	prefix := mapKeyPrefixByType(mt)

	valMap, err := r.groupVal(prefix)
	if err != nil {
		return nil, err
	}

	return &system.OnlineMapEditor{
		MapType:      mt,
		Key:          valMap[prefix+"key"],
		SecurityCode: valMap[prefix+"securityCode"],
	}, nil
}

func (r *SystemParamRepository) SaveOnlineMap(editor *system.OnlineMapEditor) error {
	if editor == nil {
		return nil
	}
	mt := strings.TrimSpace(editor.MapType)
	if mt == "" {
		mt = "gaode"
	}

	return r.db.Transaction(func(tx *gorm.DB) error {
		fields := map[string]string{
			"mapType":      mt,
			"key":          editor.Key,
			"securityCode": editor.SecurityCode,
		}
		for field, val := range fields {
			prefix := mapPrefix
			if field != "mapType" && mt != "gaode" {
				prefix = mt + "." + mapPrefix
			}
			if err := upsertByPkey(tx, prefix+field, val, "text", 1); err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *SystemParamRepository) GetSQLBotConfig() (*system.SQLBotConfig, error) {
	valMap, err := r.groupVal(sqlbotPrefix)
	if err != nil {
		return nil, err
	}

	domain := valMap[sqlbotPrefix+"domain"]
	id := valMap[sqlbotPrefix+"id"]
	enabled := parseBool(valMap[sqlbotPrefix+"enabled"])
	valid := parseBool(valMap[sqlbotPrefix+"valid"])

	if domain == "" && id == "" && !enabled && !valid {
		return nil, nil
	}

	return &system.SQLBotConfig{
		Domain:  domain,
		ID:      id,
		Enabled: enabled,
		Valid:   valid,
	}, nil
}

func (r *SystemParamRepository) SaveSQLBotConfig(cfg *system.SQLBotConfig) error {
	if cfg == nil {
		return nil
	}
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("pkey LIKE ?", sqlbotPrefix+"%").Delete(&coreSysSetting{}).Error; err != nil {
			return err
		}

		records := []coreSysSetting{
			{Pkey: sqlbotPrefix + "domain", Pval: cfg.Domain, Type: "text", Sort: 0},
			{Pkey: sqlbotPrefix + "id", Pval: cfg.ID, Type: "text", Sort: 0},
			{Pkey: sqlbotPrefix + "enabled", Pval: strconv.FormatBool(cfg.Enabled), Type: "text", Sort: 0},
			{Pkey: sqlbotPrefix + "valid", Pval: strconv.FormatBool(cfg.Valid), Type: "text", Sort: 0},
		}
		for _, rec := range records {
			if err := tx.Create(&rec).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *SystemParamRepository) GetShareBase() (*system.ShareBase, error) {
	disableText, err := r.singleVal("basic.shareDisable")
	if err != nil {
		return nil, err
	}
	peRequireText, err := r.singleVal("basic.sharePeRequire")
	if err != nil {
		return nil, err
	}
	return &system.ShareBase{
		Disable:   parseBool(disableText),
		PERequire: parseBool(peRequireText),
	}, nil
}

func (r *SystemParamRepository) GetRequestTimeOut() (int, error) {
	val, err := r.singleVal("basic.frontTimeOut")
	if err != nil {
		return 0, err
	}
	if strings.TrimSpace(val) == "" {
		return 60, nil
	}
	n, convErr := strconv.Atoi(val)
	if convErr != nil || n <= 0 {
		return 60, nil
	}
	return n, nil
}

func (r *SystemParamRepository) GetDefaultSettings() (map[string]interface{}, error) {
	result := map[string]interface{}{
		"defaultSort": "1",
	}

	defaultSort, err := r.singleVal("basic.defaultSort")
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(defaultSort) != "" {
		result["defaultSort"] = defaultSort
	}

	defaultOpen, err := r.singleVal("basic.defaultOpen")
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(defaultOpen) != "" {
		result["defaultOpen"] = defaultOpen
	}

	return result, nil
}

func (r *SystemParamRepository) GetUI() ([]interface{}, error) {
	return []interface{}{
		map[string]interface{}{"pkey": "community", "pval": true},
		map[string]interface{}{"pkey": "showDemoTips", "pval": false},
		map[string]interface{}{"pkey": "demoTipsContent", "pval": ""},
	}, nil
}

func (r *SystemParamRepository) GetDefaultLogin() (int, error) {
	val, err := r.singleVal("basic.defaultLogin")
	if err != nil {
		return 0, err
	}
	if strings.TrimSpace(val) == "" {
		return 0, nil
	}
	n, convErr := strconv.Atoi(val)
	if convErr != nil {
		return 0, nil
	}
	return n, nil
}

func (r *SystemParamRepository) GetI18nOptions() (map[string]string, error) {
	return map[string]string{}, nil
}

func (r *SystemParamRepository) singleVal(key string) (string, error) {
	var row coreSysSetting
	err := r.db.Where("pkey = ?", key).Take(&row).Error
	if err == nil {
		return row.Pval, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", nil
	}
	return "", err
}

func (r *SystemParamRepository) listByPrefix(prefix string) ([]coreSysSetting, error) {
	var list []coreSysSetting
	err := r.db.Where("pkey LIKE ?", prefix+"%").Order("sort ASC").Find(&list).Error
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (r *SystemParamRepository) groupVal(prefix string) (map[string]string, error) {
	list, err := r.listByPrefix(prefix)
	if err != nil {
		return nil, err
	}
	result := make(map[string]string, len(list))
	for _, item := range list {
		result[item.Pkey] = item.Pval
	}
	return result, nil
}

func mapKeyPrefixByType(mapType string) string {
	if mapType == "gaode" {
		return mapPrefix
	}
	return mapType + "." + mapPrefix
}

func parseBool(v string) bool {
	return strings.EqualFold(strings.TrimSpace(v), "true")
}

func upsertByPkey(tx *gorm.DB, pkey, pval, tp string, sort int) error {
	var exists coreSysSetting
	err := tx.Where("pkey = ?", pkey).Take(&exists).Error
	if err == nil {
		exists.Pval = pval
		exists.Type = tp
		exists.Sort = sort
		return tx.Save(&exists).Error
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return tx.Create(&coreSysSetting{
		Pkey: pkey,
		Pval: pval,
		Type: tp,
		Sort: sort,
	}).Error
}
