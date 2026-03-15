package service

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"dataease/backend/internal/domain/role"
	"dataease/backend/internal/repository"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var roleMockDriverSeq uint64

type roleMockStore struct {
	mu         sync.Mutex
	nextID     int64
	roles      map[int64]*role.SysRole
	failCreate error
	failUpdate error
	failDelete error
	failGet    error
	failQuery  error
}

func newRoleMockStore() *roleMockStore {
	return &roleMockStore{
		nextID: 1,
		roles:  make(map[int64]*role.SysRole),
	}
}

func (s *roleMockStore) add(rle *role.SysRole) int64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := s.nextID
	s.nextID++
	cp := cloneRole(rle)
	cp.RoleID = id
	s.roles[id] = cp
	return id
}

func (s *roleMockStore) get(id int64) (*role.SysRole, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	rle, ok := s.roles[id]
	if !ok {
		return nil, false
	}
	return cloneRole(rle), true
}

func cloneRole(rle *role.SysRole) *role.SysRole {
	if rle == nil {
		return nil
	}
	cp := *rle
	if rle.RoleDesc != nil {
		v := *rle.RoleDesc
		cp.RoleDesc = &v
	}
	if rle.ParentID != nil {
		v := *rle.ParentID
		cp.ParentID = &v
	}
	if rle.Level != nil {
		v := *rle.Level
		cp.Level = &v
	}
	if rle.DataScope != nil {
		v := *rle.DataScope
		cp.DataScope = &v
	}
	if rle.CreateBy != nil {
		v := *rle.CreateBy
		cp.CreateBy = &v
	}
	if rle.CreateTime != nil {
		v := *rle.CreateTime
		cp.CreateTime = &v
	}
	if rle.UpdateBy != nil {
		v := *rle.UpdateBy
		cp.UpdateBy = &v
	}
	if rle.UpdateTime != nil {
		v := *rle.UpdateTime
		cp.UpdateTime = &v
	}
	return &cp
}

type roleMockDriver struct {
	store *roleMockStore
}

func (d *roleMockDriver) Open(_ string) (driver.Conn, error) {
	return &roleMockConn{store: d.store}, nil
}

type roleMockConn struct {
	store *roleMockStore
}

func (c *roleMockConn) Prepare(query string) (driver.Stmt, error) {
	return &roleMockStmt{conn: c, query: query}, nil
}

func (c *roleMockConn) Close() error { return nil }

func (c *roleMockConn) Begin() (driver.Tx, error) { return &roleMockTx{}, nil }

func (c *roleMockConn) BeginTx(_ context.Context, _ driver.TxOptions) (driver.Tx, error) {
	return &roleMockTx{}, nil
}

func (c *roleMockConn) ExecContext(_ context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	q := strings.ToLower(query)

	c.store.mu.Lock()
	defer c.store.mu.Unlock()

	switch {
	case strings.HasPrefix(q, "insert into `sys_role`"):
		if c.store.failCreate != nil {
			return nil, c.store.failCreate
		}
		cols := parseInsertColumns(query)
		rle := &role.SysRole{}
		for idx, col := range cols {
			if idx >= len(args) {
				break
			}
			setRoleColumn(rle, col, args[idx].Value)
		}
		rle.RoleID = c.store.nextID
		c.store.nextID++
		c.store.roles[rle.RoleID] = cloneRole(rle)
		return roleMockResult{lastInsertID: rle.RoleID, rowsAffected: 1}, nil

	case strings.HasPrefix(q, "update `sys_role` set"):
		if c.store.failUpdate != nil {
			return nil, c.store.failUpdate
		}
		setCols := parseUpdateSetColumns(query)
		if len(args) == 0 {
			return roleMockResult{rowsAffected: 0}, nil
		}
		id := mockToInt64(args[len(args)-1].Value)
		rle, ok := c.store.roles[id]
		if !ok {
			return roleMockResult{rowsAffected: 0}, nil
		}
		for idx, col := range setCols {
			if idx >= len(args)-1 {
				break
			}
			setRoleColumn(rle, col, args[idx].Value)
		}
		return roleMockResult{rowsAffected: 1}, nil

	case strings.HasPrefix(q, "delete from `sys_role`"):
		if c.store.failDelete != nil {
			return nil, c.store.failDelete
		}
		if len(args) == 0 {
			return roleMockResult{rowsAffected: 0}, nil
		}
		id := mockToInt64(args[0].Value)
		if _, ok := c.store.roles[id]; !ok {
			return roleMockResult{rowsAffected: 0}, nil
		}
		delete(c.store.roles, id)
		return roleMockResult{rowsAffected: 1}, nil
	}

	return roleMockResult{rowsAffected: 0}, nil
}

func (c *roleMockConn) QueryContext(_ context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	q := strings.ToLower(query)

	c.store.mu.Lock()
	defer c.store.mu.Unlock()

	if strings.Contains(q, "from `sys_role`") && strings.Contains(q, "where role_id = ?") {
		if c.store.failGet != nil {
			return nil, c.store.failGet
		}
		if len(args) < 1 {
			return newRoleMockRows(nil), nil
		}
		id := mockToInt64(args[0].Value)
		rle, ok := c.store.roles[id]
		if !ok {
			return newRoleMockRows(nil), nil
		}
		if strings.Contains(q, "and status = ?") {
			if len(args) < 2 {
				return newRoleMockRows(nil), nil
			}
			status := int(mockToInt64(args[1].Value))
			if rle.Status != status {
				return newRoleMockRows(nil), nil
			}
		}
		return newRoleMockRows([]*role.SysRole{cloneRole(rle)}), nil
	}

	if strings.Contains(q, "from `sys_role`") {
		if c.store.failQuery != nil {
			return nil, c.store.failQuery
		}
		roles := make([]*role.SysRole, 0, len(c.store.roles))
		keyword := ""
		for _, arg := range args {
			val := mockToString(arg.Value)
			if strings.Contains(val, "%") {
				keyword = strings.Trim(val, "%")
				break
			}
		}

		var statusFilter *int
		if strings.Contains(q, "status = ?") && len(args) > 0 {
			status := int(mockToInt64(args[len(args)-1].Value))
			statusFilter = &status
		}

		hasIDFilter := strings.Contains(q, "role_id in") || strings.Contains(q, "`role_id` in")
		idFilter := make(map[int64]bool)
		if hasIDFilter {
			limit := len(args)
			if statusFilter != nil && limit > 0 {
				limit--
			}
			for i := 0; i < limit; i++ {
				idFilter[mockToInt64(args[i].Value)] = true
			}
		}
		for _, rle := range c.store.roles {
			if statusFilter != nil && rle.Status != *statusFilter {
				continue
			}
			if hasIDFilter && !idFilter[rle.RoleID] {
				continue
			}
			if keyword != "" &&
				!strings.Contains(strings.ToLower(rle.RoleName), strings.ToLower(keyword)) &&
				!strings.Contains(strings.ToLower(rle.RoleCode), strings.ToLower(keyword)) {
				continue
			}
			roles = append(roles, cloneRole(rle))
		}
		sort.Slice(roles, func(i, j int) bool {
			return roles[i].RoleID > roles[j].RoleID
		})
		return newRoleMockRows(roles), nil
	}

	return newRoleMockRows(nil), nil
}

type roleMockStmt struct {
	conn  *roleMockConn
	query string
}

func (s *roleMockStmt) Close() error { return nil }

func (s *roleMockStmt) NumInput() int { return -1 }

func (s *roleMockStmt) Exec(args []driver.Value) (driver.Result, error) {
	named := make([]driver.NamedValue, 0, len(args))
	for i := range args {
		named = append(named, driver.NamedValue{Ordinal: i + 1, Value: args[i]})
	}
	return s.conn.ExecContext(context.Background(), s.query, named)
}

func (s *roleMockStmt) Query(args []driver.Value) (driver.Rows, error) {
	named := make([]driver.NamedValue, 0, len(args))
	for i := range args {
		named = append(named, driver.NamedValue{Ordinal: i + 1, Value: args[i]})
	}
	return s.conn.QueryContext(context.Background(), s.query, named)
}

type roleMockTx struct{}

func (t *roleMockTx) Commit() error { return nil }

func (t *roleMockTx) Rollback() error { return nil }

type roleMockResult struct {
	lastInsertID int64
	rowsAffected int64
}

func (r roleMockResult) LastInsertId() (int64, error) { return r.lastInsertID, nil }

func (r roleMockResult) RowsAffected() (int64, error) { return r.rowsAffected, nil }

type roleMockRows struct {
	idx  int
	data [][]driver.Value
}

func newRoleMockRows(roles []*role.SysRole) driver.Rows {
	rows := &roleMockRows{idx: 0, data: make([][]driver.Value, 0, len(roles))}
	for _, rle := range roles {
		rows.data = append(rows.data, []driver.Value{
			rle.RoleID,
			rle.RoleName,
			rle.RoleCode,
			ptrStringValue(rle.RoleDesc),
			ptrInt64Value(rle.ParentID),
			ptrIntValue(rle.Level),
			ptrStringValue(rle.DataScope),
			rle.Status,
			ptrStringValue(rle.CreateBy),
			ptrTimeValue(rle.CreateTime),
			ptrStringValue(rle.UpdateBy),
			ptrTimeValue(rle.UpdateTime),
		})
	}
	return rows
}

func (r *roleMockRows) Columns() []string {
	return []string{
		"role_id",
		"role_name",
		"role_code",
		"role_desc",
		"parent_id",
		"level",
		"data_scope",
		"status",
		"create_by",
		"create_time",
		"update_by",
		"update_time",
	}
}

func (r *roleMockRows) Close() error { return nil }

func (r *roleMockRows) Next(dest []driver.Value) error {
	if r.idx >= len(r.data) {
		return io.EOF
	}
	copy(dest, r.data[r.idx])
	r.idx++
	return nil
}

func parseInsertColumns(query string) []string {
	start := strings.Index(query, "(")
	valuesIdx := strings.Index(strings.ToLower(query), " values")
	if start == -1 || valuesIdx == -1 || valuesIdx <= start {
		return nil
	}
	segment := query[start+1 : valuesIdx-1]
	parts := strings.Split(segment, ",")
	cols := make([]string, 0, len(parts))
	for _, p := range parts {
		cols = append(cols, strings.Trim(strings.TrimSpace(p), "`"))
	}
	return cols
}

func parseUpdateSetColumns(query string) []string {
	lower := strings.ToLower(query)
	setIdx := strings.Index(lower, " set ")
	whereIdx := strings.Index(lower, " where ")
	if setIdx == -1 || whereIdx == -1 || whereIdx <= setIdx {
		return nil
	}
	segment := query[setIdx+5 : whereIdx]
	parts := strings.Split(segment, ",")
	cols := make([]string, 0, len(parts))
	for _, p := range parts {
		left := strings.SplitN(p, "=", 2)[0]
		cols = append(cols, strings.Trim(strings.TrimSpace(left), "`"))
	}
	return cols
}

func setRoleColumn(rle *role.SysRole, col string, value any) {
	switch col {
	case "role_id":
		rle.RoleID = mockToInt64(value)
	case "role_name":
		rle.RoleName = mockToString(value)
	case "role_code":
		rle.RoleCode = mockToString(value)
	case "role_desc":
		rle.RoleDesc = mockToStringPtr(value)
	case "parent_id":
		rle.ParentID = mockToInt64Ptr(value)
	case "level":
		rle.Level = mockToIntPtr(value)
	case "data_scope":
		rle.DataScope = mockToStringPtr(value)
	case "status":
		rle.Status = int(mockToInt64(value))
	case "create_by":
		rle.CreateBy = mockToStringPtr(value)
	case "create_time":
		rle.CreateTime = mockToTimePtr(value)
	case "update_by":
		rle.UpdateBy = mockToStringPtr(value)
	case "update_time":
		rle.UpdateTime = mockToTimePtr(value)
	}
}

func mockToString(v any) string {
	switch x := v.(type) {
	case nil:
		return ""
	case string:
		return x
	case []byte:
		return string(x)
	default:
		return fmt.Sprintf("%v", x)
	}
}

func mockToStringPtr(v any) *string {
	if v == nil {
		return nil
	}
	s := mockToString(v)
	return &s
}

func mockToInt64(v any) int64 {
	switch x := v.(type) {
	case int64:
		return x
	case int:
		return int64(x)
	case int32:
		return int64(x)
	case uint64:
		return int64(x)
	case []byte:
		var out int64
		_, _ = fmt.Sscan(string(x), &out)
		return out
	default:
		return 0
	}
}

func mockToInt64Ptr(v any) *int64 {
	if v == nil {
		return nil
	}
	n := mockToInt64(v)
	return &n
}

func mockToIntPtr(v any) *int {
	if v == nil {
		return nil
	}
	n := int(mockToInt64(v))
	return &n
}

func mockToTimePtr(v any) *time.Time {
	t, ok := v.(time.Time)
	if !ok {
		return nil
	}
	return &t
}

func ptrStringValue(v *string) any {
	if v == nil {
		return nil
	}
	return *v
}

func ptrInt64Value(v *int64) any {
	if v == nil {
		return nil
	}
	return *v
}

func ptrIntValue(v *int) any {
	if v == nil {
		return nil
	}
	return *v
}

func ptrTimeValue(v *time.Time) any {
	if v == nil {
		return nil
	}
	return *v
}

func setupRoleService(t *testing.T) (*RoleService, *roleMockStore, func()) {
	t.Helper()
	store := newRoleMockStore()
	driverName := fmt.Sprintf("role_mock_%d", atomic.AddUint64(&roleMockDriverSeq, 1))
	sql.Register(driverName, &roleMockDriver{store: store})

	sqlDB, err := sql.Open(driverName, "")
	if err != nil {
		t.Fatalf("open mock sql db failed: %v", err)
	}

	gdb, err := gorm.Open(mysql.New(mysql.Config{Conn: sqlDB, SkipInitializeWithVersion: true}), &gorm.Config{DisableAutomaticPing: true})
	if err != nil {
		_ = sqlDB.Close()
		t.Fatalf("open gorm db failed: %v", err)
	}

	repo := repository.NewRoleRepository(gdb)
	userRepo := repository.NewUserRepository(gdb)
	userRoleRepo := repository.NewUserRoleRepository(gdb)
	return NewRoleService(repo, userRepo, userRoleRepo), store, func() {
		_ = sqlDB.Close()
	}
}

func TestRoleCreateRole_SuccessAndAutoRoleCode(t *testing.T) {
	svc, store, cleanup := setupRoleService(t)
	defer cleanup()

	desc := "管理员角色"
	id, err := svc.CreateRole(&role.RoleCreator{Name: "管理员", Desc: &desc}, "tester")
	if err != nil {
		t.Fatalf("CreateRole failed: %v", err)
	}
	if id == 0 {
		t.Fatal("expected non-zero role id")
	}

	created, ok := store.get(id)
	if !ok {
		t.Fatalf("created role %d not found in store", id)
	}
	if !strings.HasPrefix(created.RoleCode, "role_") {
		t.Fatalf("expected generated role code with prefix role_, got %q", created.RoleCode)
	}
	if created.DataScope == nil || *created.DataScope != role.DataScopeSelf {
		t.Fatalf("expected dataScope self, got %+v", created.DataScope)
	}
	if created.CreateBy == nil || *created.CreateBy != "tester" {
		t.Fatalf("expected createBy tester, got %+v", created.CreateBy)
	}
	if created.Status != role.StatusEnabled {
		t.Fatalf("expected enabled status, got %d", created.Status)
	}
}

func TestRoleCreateRole_Fail(t *testing.T) {
	svc, store, cleanup := setupRoleService(t)
	defer cleanup()

	store.failCreate = errors.New("create failed")
	_, err := svc.CreateRole(&role.RoleCreator{Name: "普通用户"}, "tester")
	if err == nil || !strings.Contains(err.Error(), "failed to create role") {
		t.Fatalf("expected create error, got %v", err)
	}
}

func TestRoleEditRole_Success(t *testing.T) {
	svc, store, cleanup := setupRoleService(t)
	defer cleanup()

	initialDesc := "初始描述"
	id := store.add(&role.SysRole{RoleName: "旧名称", RoleCode: "old_code", RoleDesc: &initialDesc, Status: role.StatusEnabled})

	newDesc := "新描述"
	err := svc.EditRole(&role.RoleEditor{ID: id, Name: "新名称", Desc: &newDesc}, "editor")
	if err != nil {
		t.Fatalf("EditRole failed: %v", err)
	}

	updated, ok := store.get(id)
	if !ok {
		t.Fatalf("updated role %d not found", id)
	}
	if updated.RoleName != "新名称" {
		t.Fatalf("expected role name updated, got %q", updated.RoleName)
	}
	if updated.RoleDesc == nil || *updated.RoleDesc != newDesc {
		t.Fatalf("expected role desc updated, got %+v", updated.RoleDesc)
	}
	if updated.UpdateBy == nil || *updated.UpdateBy != "editor" {
		t.Fatalf("expected updateBy editor, got %+v", updated.UpdateBy)
	}
	if updated.UpdateTime == nil {
		t.Fatal("expected updateTime to be set")
	}
}

func TestRoleEditRole_NotFound(t *testing.T) {
	svc, _, cleanup := setupRoleService(t)
	defer cleanup()

	err := svc.EditRole(&role.RoleEditor{ID: 999, Name: "不存在角色"}, "editor")
	if err == nil || !strings.Contains(err.Error(), "role not found") {
		t.Fatalf("expected role not found error, got %v", err)
	}
}

func TestRoleEditRole_UpdateFail(t *testing.T) {
	svc, store, cleanup := setupRoleService(t)
	defer cleanup()

	id := store.add(&role.SysRole{RoleName: "可更新角色", RoleCode: "updatable", Status: role.StatusEnabled})
	store.failUpdate = errors.New("update failed")

	err := svc.EditRole(&role.RoleEditor{ID: id, Name: "更新后名称"}, "editor")
	if err == nil || !strings.Contains(err.Error(), "failed to update role") {
		t.Fatalf("expected update error, got %v", err)
	}
}

func TestRoleDeleteRole_Success(t *testing.T) {
	svc, store, cleanup := setupRoleService(t)
	defer cleanup()

	id := store.add(&role.SysRole{RoleName: "待删除角色", RoleCode: "to_delete", Status: role.StatusEnabled})
	err := svc.DeleteRole(id)
	if err != nil {
		t.Fatalf("DeleteRole failed: %v", err)
	}

	if _, ok := store.get(id); ok {
		t.Fatalf("role %d should be deleted", id)
	}
}

func TestRoleDeleteRole_Fail(t *testing.T) {
	svc, store, cleanup := setupRoleService(t)
	defer cleanup()

	id := store.add(&role.SysRole{RoleName: "待删除失败角色", RoleCode: "to_delete_fail", Status: role.StatusEnabled})
	store.failDelete = errors.New("delete failed")
	err := svc.DeleteRole(id)
	if err == nil || !strings.Contains(err.Error(), "failed to delete role") {
		t.Fatalf("expected delete error, got %v", err)
	}
}

func TestRoleGetRoleByID_SuccessAndVOConvert(t *testing.T) {
	svc, store, cleanup := setupRoleService(t)
	defer cleanup()

	desc := "测试角色"
	parentID := int64(100)
	level := 2
	scope := role.DataScopeDeptAndChild
	id := store.add(&role.SysRole{
		RoleName:  "数据分析师",
		RoleCode:  "analyst",
		RoleDesc:  &desc,
		ParentID:  &parentID,
		Level:     &level,
		DataScope: &scope,
		Status:    role.StatusEnabled,
	})

	vo, err := svc.GetRoleByID(id)
	if err != nil {
		t.Fatalf("GetRoleByID failed: %v", err)
	}
	if vo.ID != id || vo.Name != "数据分析师" || vo.Code != "analyst" {
		t.Fatalf("unexpected vo basic fields: %+v", vo)
	}
	if vo.Desc == nil || *vo.Desc != desc {
		t.Fatalf("unexpected vo desc: %+v", vo.Desc)
	}
	if vo.ParentID == nil || *vo.ParentID != parentID {
		t.Fatalf("unexpected vo parentID: %+v", vo.ParentID)
	}
	if vo.Level == nil || *vo.Level != level {
		t.Fatalf("unexpected vo level: %+v", vo.Level)
	}
	if vo.DataScope == nil || *vo.DataScope != scope {
		t.Fatalf("unexpected vo dataScope: %+v", vo.DataScope)
	}
}

func TestRoleGetRoleByID_Fail(t *testing.T) {
	svc, _, cleanup := setupRoleService(t)
	defer cleanup()

	vo, err := svc.GetRoleByID(999)
	if err == nil || !strings.Contains(err.Error(), "role not found") {
		t.Fatalf("expected role not found error, got %v", err)
	}
	if vo != nil {
		t.Fatalf("expected nil vo when error, got %+v", vo)
	}
}

func TestRoleQueryRoles_KeywordSearch(t *testing.T) {
	svc, store, cleanup := setupRoleService(t)
	defer cleanup()

	zeroParent := int64(0)
	oneParent := int64(1)
	store.add(&role.SysRole{RoleName: "管理员", RoleCode: "admin", Status: role.StatusEnabled, ParentID: &zeroParent})
	store.add(&role.SysRole{RoleName: "访客", RoleCode: "visitor", Status: role.StatusEnabled, ParentID: &oneParent})
	store.add(&role.SysRole{RoleName: "数据分析", RoleCode: "analyst", Status: role.StatusDisabled})

	keyword := "管理"
	list, err := svc.QueryRoles(&role.RoleQueryRequest{Keyword: &keyword})
	if err != nil {
		t.Fatalf("QueryRoles keyword failed: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 role by keyword, got %d", len(list))
	}
	if list[0].Name != "管理员" {
		t.Fatalf("expected 管理员, got %s", list[0].Name)
	}
	if !list[0].Root {
		t.Fatal("expected root role when parent_id == 0")
	}
}

func TestRoleQueryRoles_EmptyResult(t *testing.T) {
	svc, store, cleanup := setupRoleService(t)
	defer cleanup()

	store.add(&role.SysRole{RoleName: "管理员", RoleCode: "admin", Status: role.StatusEnabled})
	keyword := "不存在"
	list, err := svc.QueryRoles(&role.RoleQueryRequest{Keyword: &keyword})
	if err != nil {
		t.Fatalf("QueryRoles empty failed: %v", err)
	}
	if len(list) != 0 {
		t.Fatalf("expected empty result, got %d", len(list))
	}
}

func TestRoleQueryRoles_FullList(t *testing.T) {
	svc, store, cleanup := setupRoleService(t)
	defer cleanup()

	parent := int64(3)
	store.add(&role.SysRole{RoleName: "管理员", RoleCode: "admin", Status: role.StatusEnabled})
	store.add(&role.SysRole{RoleName: "普通用户", RoleCode: "user", Status: role.StatusEnabled, ParentID: &parent})

	list, err := svc.QueryRoles(&role.RoleQueryRequest{})
	if err != nil {
		t.Fatalf("QueryRoles full list failed: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 roles, got %d", len(list))
	}

	rootCount := 0
	nonRootCount := 0
	for _, item := range list {
		if item.Root {
			rootCount++
		} else {
			nonRootCount++
		}
	}
	if rootCount != 1 || nonRootCount != 1 {
		t.Fatalf("unexpected root/non-root split: root=%d nonRoot=%d", rootCount, nonRootCount)
	}
}

func TestRoleQueryRoles_Fail(t *testing.T) {
	svc, store, cleanup := setupRoleService(t)
	defer cleanup()

	store.failQuery = errors.New("query failed")
	_, err := svc.QueryRoles(&role.RoleQueryRequest{})
	if err == nil || !strings.Contains(err.Error(), "failed to query roles") {
		t.Fatalf("expected query error, got %v", err)
	}
}

func TestRoleStrPtr(t *testing.T) {
	p := strPtr("hello")
	if p == nil {
		t.Fatal("expected non-nil pointer")
	}
	if *p != "hello" {
		t.Fatalf("expected hello, got %q", *p)
	}
}

// TestUnmountUser_LastRoleProtection 测试唯一角色安全策略
func TestUnmountUser_LastRoleProtection(t *testing.T) {
	svc, _, cleanup := setupRoleService(t)
	defer cleanup()

	// 用户只有一个角色时应该拒绝移除
	err := svc.UnmountUser(&role.UnmountUserRequest{Uid: 1, Rid: 1})
	if err == nil {
		t.Fatal("expected error when removing user's last role")
	}
	if !strings.Contains(err.Error(), "last role") {
		t.Fatalf("expected last role error, got: %v", err)
	}
}

// TestUnmountUser_MultipleRoles 测试用户有多个角色时可以移除
// 注意：此测试需要完整的 sys_user_role mock 支持
// 当前 mock 仅支持 sys_role 表，因此跳过此测试
// 在集成测试中会覆盖此场景
func TestUnmountUser_MultipleRoles(t *testing.T) {
	t.Skip("requires sys_user_role mock support - covered by integration tests")
}

// TestCreateRoleWithInheritance_Success 测试带继承的角色创建
func TestCreateRoleWithInheritance_Success(t *testing.T) {
	svc, store, cleanup := setupRoleService(t)
	defer cleanup()

	// 创建父角色（内置角色）
	parentID := store.add(&role.SysRole{RoleName: "Admin", RoleCode: "admin", Status: role.StatusEnabled, ParentID: ptrInt64(0)})

	// 创建继承自父角色的自定义角色
	childID, err := svc.CreateRoleWithInheritance(&role.RoleCreator{Name: "Custom Admin"}, &parentID, "tester")
	if err != nil {
		t.Fatalf("CreateRoleWithInheritance failed: %v", err)
	}
	if childID == 0 {
		t.Fatal("expected non-zero child role id")
	}

	// 验证子角色的父ID设置正确
	child, ok := store.get(childID)
	if !ok {
		t.Fatalf("child role %d not found", childID)
	}
	if child.ParentID == nil || *child.ParentID != parentID {
		t.Fatalf("expected parentID %d, got %+v", parentID, child.ParentID)
	}
}

// TestCreateRoleWithInheritance_InvalidParent 测试继承无效父角色
func TestCreateRoleWithInheritance_InvalidParent(t *testing.T) {
	svc, _, cleanup := setupRoleService(t)
	defer cleanup()

	// 尝试继承不存在的角色
	invalidParentID := int64(9999)
	_, err := svc.CreateRoleWithInheritance(&role.RoleCreator{Name: "Invalid Child"}, &invalidParentID, "tester")
	if err == nil {
		t.Fatal("expected error when inheriting from non-existent parent")
	}
	if !strings.Contains(err.Error(), "parent role not found") {
		t.Fatalf("expected parent not found error, got: %v", err)
	}
}

// TestValidatePermissionInheritance_NoParent 测试无父角色的权限验证
func TestValidatePermissionInheritance_NoParent(t *testing.T) {
	svc, store, cleanup := setupRoleService(t)
	defer cleanup()

	// 创建无父角色的角色
	roleID := store.add(&role.SysRole{RoleName: "Standalone", RoleCode: "standalone", Status: role.StatusEnabled})

	// 验证权限继承应该直接通过
	err := svc.ValidatePermissionInheritance(roleID, []int64{1, 2, 3})
	if err != nil {
		t.Fatalf("ValidatePermissionInheritance failed for role without parent: %v", err)
	}
}

func TestValidatePermissionInheritance_WithParent(t *testing.T) {
	svc, store, cleanup := setupRoleService(t)
	defer cleanup()

	parentID := store.add(&role.SysRole{RoleName: "Root", RoleCode: "root", Status: role.StatusEnabled, ParentID: ptrInt64(0)})
	childID := store.add(&role.SysRole{RoleName: "Child", RoleCode: "child", Status: role.StatusEnabled, ParentID: &parentID})

	err := svc.ValidatePermissionInheritance(childID, []int64{1, 2})
	if err != nil {
		t.Fatalf("ValidatePermissionInheritance failed for role with parent: %v", err)
	}
}

func TestValidatePermissionInheritance_RoleNotFound(t *testing.T) {
	svc, _, cleanup := setupRoleService(t)
	defer cleanup()

	err := svc.ValidatePermissionInheritance(99999, []int64{1})
	if err == nil {
		t.Fatal("expected error when role not found")
	}
	if !strings.Contains(err.Error(), "role not found") {
		t.Fatalf("expected role not found error, got: %v", err)
	}
}

func TestCreateRoleWithInheritance_InvalidGrandparent(t *testing.T) {
	svc, store, cleanup := setupRoleService(t)
	defer cleanup()

	missingGrandparentID := int64(99998)
	parentID := store.add(&role.SysRole{RoleName: "Parent", RoleCode: "parent", Status: role.StatusEnabled, ParentID: &missingGrandparentID})

	_, err := svc.CreateRoleWithInheritance(&role.RoleCreator{Name: "ChildWithBadGrandparent"}, &parentID, "tester")
	if err == nil {
		t.Fatal("expected error when inheriting with invalid grandparent")
	}
	if !strings.Contains(err.Error(), "grandparent role inheritance invalid") {
		t.Fatalf("expected invalid grandparent error, got: %v", err)
	}
}

func TestMountUsers_UserRoleRepoNil(t *testing.T) {
	svc, _, cleanup := setupRoleService(t)
	defer cleanup()

	svc.userRoleRepo = nil
	err := svc.MountUsers(&role.MountUserRequest{Rid: 1, OrgId: 1, Uids: []int64{1}})
	if err == nil {
		t.Fatal("expected error when userRoleRepo is nil")
	}
	if !strings.Contains(err.Error(), "userRoleRepo not initialized") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMountExternalUser_UserRoleRepoNil(t *testing.T) {
	svc, _, cleanup := setupRoleService(t)
	defer cleanup()

	svc.userRoleRepo = nil
	err := svc.MountExternalUser(&role.MountExternalUserRequest{Rid: 1, Uid: 1}, 1)
	if err == nil {
		t.Fatal("expected error when userRoleRepo is nil")
	}
	if !strings.Contains(err.Error(), "userRoleRepo not initialized") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func ptrInt64(v int64) *int64 {
	return &v
}
