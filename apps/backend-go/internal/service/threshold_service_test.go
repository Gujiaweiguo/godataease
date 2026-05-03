package service

import (
	"context"
	"sort"
	"strings"
	"testing"

	"dataease/backend/internal/domain/auto"
	thresholddomain "dataease/backend/internal/domain/threshold"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type fakeThresholdRepo struct {
	records   map[int64]*auto.XpackThresholdInfo
	nextID    int64
	instances []auto.XpackThresholdInstance
}

func newFakeThresholdRepo() *fakeThresholdRepo {
	return &fakeThresholdRepo{
		records: make(map[int64]*auto.XpackThresholdInfo),
		nextID:  1,
	}
}

func (r *fakeThresholdRepo) Create(ctx context.Context, info *auto.XpackThresholdInfo) error {
	_ = ctx
	if info == nil {
		return gorm.ErrInvalidData
	}
	cloned := *info
	if cloned.ID <= 0 {
		cloned.ID = r.nextID
		r.nextID++
	}
	r.records[cloned.ID] = &cloned
	info.ID = cloned.ID
	return nil
}

func (r *fakeThresholdRepo) Update(ctx context.Context, info *auto.XpackThresholdInfo) error {
	_ = ctx
	if info == nil || info.ID <= 0 {
		return gorm.ErrInvalidData
	}
	if _, ok := r.records[info.ID]; !ok {
		return gorm.ErrRecordNotFound
	}
	cloned := *info
	r.records[info.ID] = &cloned
	return nil
}

func (r *fakeThresholdRepo) GetByID(ctx context.Context, id int64) (*auto.XpackThresholdInfo, error) {
	_ = ctx
	row, ok := r.records[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	cloned := *row
	return &cloned, nil
}

func (r *fakeThresholdRepo) DeleteByIDs(ctx context.Context, ids []int64) error {
	_ = ctx
	for _, id := range ids {
		delete(r.records, id)
	}
	return nil
}

func (r *fakeThresholdRepo) DeleteByChartID(ctx context.Context, chartID int64) error {
	_ = ctx
	for id, row := range r.records {
		if row.ChartID == chartID {
			delete(r.records, id)
		}
	}
	return nil
}

func (r *fakeThresholdRepo) UpdateEnable(ctx context.Context, id int64, enable bool) error {
	_ = ctx
	row, ok := r.records[id]
	if !ok {
		return gorm.ErrRecordNotFound
	}
	row.Enable = enable
	return nil
}

func (r *fakeThresholdRepo) UpdateRecipients(ctx context.Context, ids []int64, users, roles, emails, larkGroups, larksuiteGroups, webhooks string) error {
	_ = ctx
	for _, id := range ids {
		row, ok := r.records[id]
		if !ok {
			continue
		}
		row.ReciUsers = users
		row.ReciRoles = roles
		row.ReciEmails = emails
		row.ReciLarkGroups = larkGroups
		row.ReciLarksuiteGroups = larksuiteGroups
		row.ReciWebhooks = webhooks
	}
	return nil
}

func (r *fakeThresholdRepo) Pager(ctx context.Context, req *thresholddomain.GridRequest, goPage, pageSize int) ([]*thresholddomain.GridVO, int64, error) {
	_ = ctx
	goPage, pageSize = normalizeThresholdPage(goPage, pageSize)
	rows := r.listRecords()
	result := make([]*thresholddomain.GridVO, 0, len(rows))
	for _, row := range rows {
		if req != nil {
			if keyword := strings.TrimSpace(req.Keyword); keyword != "" && !strings.Contains(row.Name, keyword) {
				continue
			}
			if req.ChartID != nil && row.ChartID != *req.ChartID {
				continue
			}
		}
		result = append(result, &thresholddomain.GridVO{
			ID:           row.ID,
			Name:         row.Name,
			ResourceID:   row.ResourceID,
			ResourceType: row.ResourceType,
			ChartID:      row.ChartID,
			ChartType:    row.ChartType,
			Status:       row.Status,
			Enable:       row.Enable,
			Creator:      row.Creator,
			CreateName:   row.CreatorName,
			CreateTime:   row.CreateTime,
		})
	}
	total := int64(len(result))
	start := (goPage - 1) * pageSize
	if start >= len(result) {
		return []*thresholddomain.GridVO{}, total, nil
	}
	end := min(start+pageSize, len(result))
	return result[start:end], total, nil
}

func (r *fakeThresholdRepo) ExistsByChartID(ctx context.Context, chartID int64) (bool, error) {
	_ = ctx
	for _, row := range r.records {
		if row.ChartID == chartID {
			return true, nil
		}
	}
	return false, nil
}

func (r *fakeThresholdRepo) InstancePager(ctx context.Context, req *thresholddomain.InstanceRequest, goPage, pageSize int) ([]*thresholddomain.InstanceVO, int64, error) {
	_ = ctx
	goPage, pageSize = normalizeThresholdPage(goPage, pageSize)
	filtered := make([]*thresholddomain.InstanceVO, 0, len(r.instances))
	for _, row := range r.instances {
		if req != nil {
			if req.ThresholdID != nil && row.TaskID != *req.ThresholdID {
				continue
			}
			if keyword := strings.TrimSpace(req.Keyword); keyword != "" && !strings.Contains(row.Content, keyword) && !strings.Contains(row.Msg, keyword) {
				if info, ok := r.records[row.TaskID]; !ok || !strings.Contains(info.Name, keyword) {
					continue
				}
			}
		}
		name := ""
		if info, ok := r.records[row.TaskID]; ok {
			name = info.Name
		}
		filtered = append(filtered, &thresholddomain.InstanceVO{
			ID:       row.ID,
			TaskID:   row.TaskID,
			Name:     name,
			ExecTime: row.ExecTime,
			Status:   row.Status,
			Content:  row.Content,
			Msg:      row.Msg,
		})
	}
	sort.Slice(filtered, func(i, j int) bool { return filtered[i].ExecTime > filtered[j].ExecTime })
	total := int64(len(filtered))
	start := (goPage - 1) * pageSize
	if start >= len(filtered) {
		return []*thresholddomain.InstanceVO{}, total, nil
	}
	end := min(start+pageSize, len(filtered))
	return filtered[start:end], total, nil
}

func (r *fakeThresholdRepo) listRecords() []*auto.XpackThresholdInfo {
	rows := make([]*auto.XpackThresholdInfo, 0, len(r.records))
	for _, row := range r.records {
		cloned := *row
		rows = append(rows, &cloned)
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].CreateTime > rows[j].CreateTime })
	return rows
}

func normalizeThresholdPage(goPage, pageSize int) (int, int) {
	if goPage < 1 {
		goPage = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	return goPage, pageSize
}

func newThresholdServiceForTest() (*ThresholdService, *fakeThresholdRepo) {
	repo := newFakeThresholdRepo()
	return NewThresholdService(repo), repo
}

func sampleThresholdRequest() *thresholddomain.CreateRequest {
	return &thresholddomain.CreateRequest{
		BaseReciDTO: thresholddomain.BaseReciDTO{
			ReciFlagList:       []int{1, 2},
			UIDList:            []string{"u1", "u2"},
			RIDList:            []string{"r1"},
			EmailList:          []string{"a@test.com"},
			LarkGroupList:      []string{"lg1"},
			LarksuiteGroupList: []string{"lsg1"},
			WebhookList:        []string{"https://hook"},
		},
		Name:           "CPU 告警",
		RateValue:      "5m",
		ResourceID:     11,
		ResourceType:   "panel",
		ChartID:        22,
		ChartType:      "bar",
		ThresholdRules: `{"logic":"and"}`,
		MsgTitle:       "title",
		MsgContent:     "content",
		ShowFieldValue: thresholdBoolPtr(true),
		ResourceTable:  "core",
	}
}

func TestThresholdService_Create_Success(t *testing.T) {
	svc, _ := newThresholdServiceForTest()

	created, err := svc.Create(context.Background(), sampleThresholdRequest(), 9, "tester", 99)
	require.NoError(t, err)
	assert.NotZero(t, created.ID)
	assert.Equal(t, "CPU 告警", created.Name)
	assert.True(t, created.Enable)
	assert.EqualValues(t, 1, created.RateType)
	assert.EqualValues(t, 0, created.MsgType)
	assert.True(t, created.RepeatSend)
	assert.Equal(t, int64(9), created.Creator)
	assert.Equal(t, "tester", created.CreatorName)
	assert.Equal(t, int64(99), created.Oid)
	assert.NotZero(t, created.CreateTime)
	assert.Equal(t, `["u1","u2"]`, created.ReciUsers)
	assert.Equal(t, `[1,2]`, created.Recisetting)
}

func TestThresholdService_Create_ValidationMissingName(t *testing.T) {
	svc, _ := newThresholdServiceForTest()
	req := sampleThresholdRequest()
	req.Name = "  "

	created, err := svc.Create(context.Background(), req, 1, "tester", 1)
	assert.Nil(t, created)
	assert.ErrorIs(t, err, gorm.ErrInvalidData)
}

func TestThresholdService_Create_ValidationMissingChartId(t *testing.T) {
	svc, _ := newThresholdServiceForTest()
	req := sampleThresholdRequest()
	req.ChartID = 0

	created, err := svc.Create(context.Background(), req, 1, "tester", 1)
	assert.Nil(t, created)
	assert.ErrorIs(t, err, gorm.ErrInvalidData)
}

func TestThresholdService_Edit_Success(t *testing.T) {
	svc, _ := newThresholdServiceForTest()
	created, err := svc.Create(context.Background(), sampleThresholdRequest(), 1, "tester", 2)
	require.NoError(t, err)

	req := sampleThresholdRequest()
	req.ID = created.ID
	req.Name = "Memory 告警"
	disabled := false
	req.Enable = &disabled
	updated, err := svc.Edit(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, "Memory 告警", updated.Name)
	assert.False(t, updated.Enable)
	assert.Equal(t, created.Creator, updated.Creator)
}

func TestThresholdService_Edit_NotFound(t *testing.T) {
	svc, _ := newThresholdServiceForTest()
	req := sampleThresholdRequest()
	req.ID = 404

	updated, err := svc.Edit(context.Background(), req)
	assert.Nil(t, updated)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestThresholdService_FormInfo_Success(t *testing.T) {
	svc, _ := newThresholdServiceForTest()
	created, err := svc.Create(context.Background(), sampleThresholdRequest(), 1, "tester", 1)
	require.NoError(t, err)

	form, err := svc.FormInfo(context.Background(), created.ID, "core")
	require.NoError(t, err)
	assert.Equal(t, created.ID, form.ID)
	assert.Equal(t, sampleThresholdRequest().UIDList, form.UIDList)
	assert.Equal(t, sampleThresholdRequest().RIDList, form.RIDList)
	assert.Equal(t, sampleThresholdRequest().EmailList, form.EmailList)
	assert.Equal(t, sampleThresholdRequest().LarkGroupList, form.LarkGroupList)
	assert.Equal(t, sampleThresholdRequest().LarksuiteGroupList, form.LarksuiteGroupList)
	assert.Equal(t, sampleThresholdRequest().WebhookList, form.WebhookList)
	assert.Equal(t, sampleThresholdRequest().ReciFlagList, form.ReciFlagList)
	assert.NotNil(t, form.ShowFieldValue)
	assert.False(t, *form.ShowFieldValue)
	assert.Equal(t, "core", form.ResourceTable)
}

func TestThresholdService_FormInfo_Snapshot(t *testing.T) {
	svc, _ := newThresholdServiceForTest()

	form, err := svc.FormInfo(context.Background(), 1, "snapshot")
	require.NoError(t, err)
	assert.Equal(t, &thresholddomain.CreateRequest{}, form)
}

func TestThresholdService_SwitchEnable(t *testing.T) {
	svc, _ := newThresholdServiceForTest()
	created, err := svc.Create(context.Background(), sampleThresholdRequest(), 1, "tester", 1)
	require.NoError(t, err)

	disabled := false
	err = svc.SwitchEnable(context.Background(), &thresholddomain.SwitchRequest{ID: created.ID, Enable: &disabled, ResourceTable: "core"})
	require.NoError(t, err)

	form, err := svc.FormInfo(context.Background(), created.ID, "core")
	require.NoError(t, err)
	assert.NotNil(t, form.Enable)
	assert.False(t, *form.Enable)
}

func TestThresholdService_Delete(t *testing.T) {
	svc, repo := newThresholdServiceForTest()
	created, err := svc.Create(context.Background(), sampleThresholdRequest(), 1, "tester", 1)
	require.NoError(t, err)

	err = svc.Delete(context.Background(), []int64{created.ID}, "core")
	require.NoError(t, err)
	_, getErr := repo.GetByID(context.Background(), created.ID)
	assert.ErrorIs(t, getErr, gorm.ErrRecordNotFound)
}

func TestThresholdService_DeleteWithChart(t *testing.T) {
	svc, repo := newThresholdServiceForTest()
	created, err := svc.Create(context.Background(), sampleThresholdRequest(), 1, "tester", 1)
	require.NoError(t, err)

	err = svc.DeleteWithChart(context.Background(), created.ChartID, "core")
	require.NoError(t, err)
	_, getErr := repo.GetByID(context.Background(), created.ID)
	assert.ErrorIs(t, getErr, gorm.ErrRecordNotFound)
}

func TestThresholdService_BatchReci(t *testing.T) {
	svc, repo := newThresholdServiceForTest()
	one, err := svc.Create(context.Background(), sampleThresholdRequest(), 1, "tester", 1)
	require.NoError(t, err)
	twoReq := sampleThresholdRequest()
	twoReq.Name = "another"
	twoReq.ChartID = 23
	two, err := svc.Create(context.Background(), twoReq, 1, "tester", 1)
	require.NoError(t, err)

	err = svc.BatchReci(context.Background(), &thresholddomain.BatchReciRequest{
		BaseReciDTO: thresholddomain.BaseReciDTO{
			UIDList:            []string{"new-user"},
			RIDList:            []string{"new-role"},
			EmailList:          []string{"new@test.com"},
			LarkGroupList:      []string{"lg2"},
			LarksuiteGroupList: []string{"lsg2"},
			WebhookList:        []string{"https://new-hook"},
		},
		IDList: []int64{one.ID, two.ID},
	})
	require.NoError(t, err)
	assert.Equal(t, `["new-user"]`, repo.records[one.ID].ReciUsers)
	assert.Equal(t, `["new-user"]`, repo.records[two.ID].ReciUsers)
}

func TestThresholdService_Pager(t *testing.T) {
	svc, _ := newThresholdServiceForTest()
	for i, name := range []string{"one", "two", "three"} {
		req := sampleThresholdRequest()
		req.Name = name
		req.ChartID = int64(100 + i)
		_, err := svc.Create(context.Background(), req, 1, "tester", 1)
		require.NoError(t, err)
	}

	page, err := svc.Pager(context.Background(), &thresholddomain.GridRequest{Keyword: "t"}, 1, 2)
	require.NoError(t, err)
	assert.Equal(t, int64(2), page.Total)
	rows, ok := page.List.([]*thresholddomain.GridVO)
	require.True(t, ok)
	assert.Len(t, rows, 2)
}

func TestThresholdService_AnyThreshold(t *testing.T) {
	svc, _ := newThresholdServiceForTest()
	created, err := svc.Create(context.Background(), sampleThresholdRequest(), 1, "tester", 1)
	require.NoError(t, err)

	exists, err := svc.AnyThreshold(context.Background(), created.ChartID, "core")
	require.NoError(t, err)
	assert.True(t, exists)

	exists, err = svc.AnyThreshold(context.Background(), 9999, "core")
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestThresholdService_InstancePager(t *testing.T) {
	svc, repo := newThresholdServiceForTest()
	created, err := svc.Create(context.Background(), sampleThresholdRequest(), 1, "tester", 1)
	require.NoError(t, err)
	repo.instances = []auto.XpackThresholdInstance{
		{ID: 1, TaskID: created.ID, ExecTime: 100, Status: true, Content: "hit", Msg: ""},
		{ID: 2, TaskID: created.ID, ExecTime: 200, Status: false, Content: "miss", Msg: "failed"},
	}

	page, err := svc.InstancePager(context.Background(), &thresholddomain.InstanceRequest{Keyword: "fail", ThresholdID: &created.ID}, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(1), page.Total)
	rows, ok := page.List.([]*thresholddomain.InstanceVO)
	require.True(t, ok)
	assert.Len(t, rows, 1)
	assert.Equal(t, int64(2), rows[0].ID)
	assert.Equal(t, "CPU 告警", rows[0].Name)
}

func TestThresholdService_Preview_NotImplemented(t *testing.T) {
	svc, _ := newThresholdServiceForTest()
	content, err := svc.Preview(context.Background(), &thresholddomain.PreviewRequest{})
	assert.Empty(t, content)
	assert.EqualError(t, err, "preview requires chart data access - not yet implemented")
}

func TestThresholdService_RepoNotReady(t *testing.T) {
	svc := NewThresholdService(nil)
	_, err := svc.Create(context.Background(), sampleThresholdRequest(), 1, "tester", 1)
	assert.ErrorIs(t, err, errThresholdRepoNotReady)
}

func TestThresholdService_Edit_InvalidRequest(t *testing.T) {
	svc, _ := newThresholdServiceForTest()
	updated, err := svc.Edit(context.Background(), nil)
	assert.Nil(t, updated)
	assert.ErrorIs(t, err, gorm.ErrInvalidData)
}

func TestThresholdService_BatchReci_EmptyIDs(t *testing.T) {
	svc, _ := newThresholdServiceForTest()
	err := svc.BatchReci(context.Background(), &thresholddomain.BatchReciRequest{})
	assert.NoError(t, err)
}

func TestThresholdService_SnapshotShortcuts(t *testing.T) {
	svc := NewThresholdService(nil)
	err := svc.SwitchEnable(context.Background(), &thresholddomain.SwitchRequest{ResourceTable: "snapshot"})
	assert.NoError(t, err)
	err = svc.Delete(context.Background(), []int64{1}, "snapshot")
	assert.NoError(t, err)
	err = svc.DeleteWithChart(context.Background(), 1, "snapshot")
	assert.NoError(t, err)
	exists, existsErr := svc.AnyThreshold(context.Background(), 1, "snapshot")
	assert.NoError(t, existsErr)
	assert.False(t, exists)
}
