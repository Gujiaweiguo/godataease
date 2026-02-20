package org

import "time"

// 状态常量
const (
	StatusEnabled  = 1
	StatusDisabled = 0
)

// 删除标记常量
const (
	DelFlagNormal  = 0
	DelFlagDeleted = 1
)

// 顶级组织父ID
const (
	RootParentID int64 = 0
)

// SysOrg 组织实体 - 映射 sys_org 表
type SysOrg struct {
	OrgID      int64      `gorm:"column:org_id;primaryKey;autoIncrement" json:"orgId"`
	OrgName    string     `gorm:"column:org_name;size:100;not null" json:"orgName"`
	OrgDesc    *string    `gorm:"column:org_desc;size:500" json:"orgDesc"`
	ParentID   int64      `gorm:"column:parent_id;default:0" json:"parentId"`
	Level      int        `gorm:"column:level;default:1" json:"level"`
	DeptID     *int64     `gorm:"column:dept_id" json:"deptId"`
	Status     int        `gorm:"column:status;default:1" json:"status"`
	DelFlag    int        `gorm:"column:del_flag;default:0" json:"delFlag"`
	CreateBy   *string    `gorm:"column:create_by;size:100" json:"createBy"`
	CreateTime time.Time  `gorm:"column:create_time;autoCreateTime" json:"createTime"`
	UpdateBy   *string    `gorm:"column:update_by;size:100" json:"updateBy"`
	UpdateTime *time.Time `gorm:"column:update_time" json:"updateTime"`
}

func (SysOrg) TableName() string {
	return "sys_org"
}

// OrgCreateRequest 创建组织请求
type OrgCreateRequest struct {
	OrgName  string  `json:"orgName" binding:"required"`
	OrgDesc  *string `json:"orgDesc"`
	ParentID *int64  `json:"parentId"`
}

// OrgUpdateRequest 更新组织请求
type OrgUpdateRequest struct {
	OrgID   int64   `json:"orgId" binding:"required"`
	OrgName string  `json:"orgName"`
	OrgDesc *string `json:"orgDesc"`
}

// OrgStatusRequest 更新组织状态请求
type OrgStatusRequest struct {
	OrgID  int64 `json:"orgId" binding:"required"`
	Status int   `json:"status" binding:"required"`
}

// OrgTreeNode 组织树节点
type OrgTreeNode struct {
	OrgID    int64          `json:"orgId"`
	OrgName  string         `json:"orgName"`
	OrgDesc  *string        `json:"orgDesc"`
	ParentID int64          `json:"parentId"`
	Level    int            `json:"level"`
	Status   int            `json:"status"`
	Children []*OrgTreeNode `json:"children,omitempty"`
}

// ToTreeNode 将 SysOrg 转换为 OrgTreeNode
func (o *SysOrg) ToTreeNode() *OrgTreeNode {
	return &OrgTreeNode{
		OrgID:    o.OrgID,
		OrgName:  o.OrgName,
		OrgDesc:  o.OrgDesc,
		ParentID: o.ParentID,
		Level:    o.Level,
		Status:   o.Status,
		Children: []*OrgTreeNode{},
	}
}
