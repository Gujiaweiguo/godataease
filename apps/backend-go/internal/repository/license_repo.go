package repository

import (
	"strconv"

	"dataease/backend/internal/domain/license"

	"gorm.io/gorm"
)

const licensePrefix = "license."

type LicenseRepository struct {
	db *gorm.DB
}

func NewLicenseRepository(db *gorm.DB) *LicenseRepository {
	return &LicenseRepository{db: db}
}

func (r *LicenseRepository) Load() (*license.ValidateResult, string, error) {
	var list []coreSysSetting
	if err := r.db.Where("pkey LIKE ?", licensePrefix+"%").Order("sort ASC").Find(&list).Error; err != nil {
		return nil, "", err
	}
	if len(list) == 0 {
		return nil, "", nil
	}

	kv := make(map[string]string, len(list))
	for _, item := range list {
		kv[item.Pkey] = item.Pval
	}

	status := kv[licensePrefix+"status"]
	message := kv[licensePrefix+"message"]
	raw := kv[licensePrefix+"raw"]

	count := int64(0)
	if c, err := strconv.ParseInt(kv[licensePrefix+"count"], 10, 64); err == nil {
		count = c
	}

	lic := &license.LicenseInfo{
		Corporation: kv[licensePrefix+"corporation"],
		Expired:     kv[licensePrefix+"expired"],
		Count:       count,
		Version:     kv[licensePrefix+"version"],
		Edition:     kv[licensePrefix+"edition"],
		SerialNo:    kv[licensePrefix+"serialNo"],
		Remark:      kv[licensePrefix+"remark"],
		ISV:         kv[licensePrefix+"isv"],
	}

	if lic.Corporation == "" && lic.Expired == "" && lic.Count == 0 && lic.Version == "" && lic.Edition == "" && lic.SerialNo == "" && lic.Remark == "" && lic.ISV == "" {
		lic = nil
	}

	return &license.ValidateResult{
		Status:  status,
		Message: message,
		License: lic,
	}, raw, nil
}

func (r *LicenseRepository) Save(result *license.ValidateResult, raw string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("pkey LIKE ?", licensePrefix+"%").Delete(&coreSysSetting{}).Error; err != nil {
			return err
		}

		sort := 1
		records := []coreSysSetting{
			{Pkey: licensePrefix + "status", Pval: result.Status, Type: "text", Sort: sort},
		}
		sort++
		if result.Message != "" {
			records = append(records, coreSysSetting{Pkey: licensePrefix + "message", Pval: result.Message, Type: "text", Sort: sort})
			sort++
		}

		if result.License != nil {
			records = append(records,
				coreSysSetting{Pkey: licensePrefix + "corporation", Pval: result.License.Corporation, Type: "text", Sort: sort},
			)
			sort++
			records = append(records, coreSysSetting{Pkey: licensePrefix + "expired", Pval: result.License.Expired, Type: "text", Sort: sort})
			sort++
			records = append(records, coreSysSetting{Pkey: licensePrefix + "count", Pval: strconv.FormatInt(result.License.Count, 10), Type: "text", Sort: sort})
			sort++
			records = append(records, coreSysSetting{Pkey: licensePrefix + "version", Pval: result.License.Version, Type: "text", Sort: sort})
			sort++
			records = append(records, coreSysSetting{Pkey: licensePrefix + "edition", Pval: result.License.Edition, Type: "text", Sort: sort})
			sort++
			records = append(records, coreSysSetting{Pkey: licensePrefix + "serialNo", Pval: result.License.SerialNo, Type: "text", Sort: sort})
			sort++
			records = append(records, coreSysSetting{Pkey: licensePrefix + "remark", Pval: result.License.Remark, Type: "text", Sort: sort})
			sort++
			records = append(records, coreSysSetting{Pkey: licensePrefix + "isv", Pval: result.License.ISV, Type: "text", Sort: sort})
			sort++
		}

		records = append(records, coreSysSetting{Pkey: licensePrefix + "raw", Pval: raw, Type: "text", Sort: sort})

		for _, rec := range records {
			if err := tx.Create(&rec).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *LicenseRepository) Clear() error {
	return r.db.Where("pkey LIKE ?", licensePrefix+"%").Delete(&coreSysSetting{}).Error
}
