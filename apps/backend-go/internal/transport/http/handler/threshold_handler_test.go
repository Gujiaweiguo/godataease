package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sort"
	"strconv"
	"strings"
	"testing"

	"dataease/backend/internal/domain/auto"
	thresholddomain "dataease/backend/internal/domain/threshold"
	"dataease/backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type thresholdHandlerRepo struct {
	records map[int64]*auto.XpackThresholdInfo
	nextID  int64
}

func newThresholdHandlerRepo() *thresholdHandlerRepo {
	return &thresholdHandlerRepo{records: make(map[int64]*auto.XpackThresholdInfo), nextID: 1}
}

func (r *thresholdHandlerRepo) Create(ctx context.Context, info *auto.XpackThresholdInfo) error {
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

func (r *thresholdHandlerRepo) Update(ctx context.Context, info *auto.XpackThresholdInfo) error {
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

func (r *thresholdHandlerRepo) GetByID(ctx context.Context, id int64) (*auto.XpackThresholdInfo, error) {
	_ = ctx
	row, ok := r.records[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	cloned := *row
	return &cloned, nil
}

func (r *thresholdHandlerRepo) DeleteByIDs(ctx context.Context, ids []int64) error {
	_ = ctx
	for _, id := range ids {
		delete(r.records, id)
	}
	return nil
}

func (r *thresholdHandlerRepo) DeleteByChartID(ctx context.Context, chartID int64) error {
	_ = ctx
	for id, row := range r.records {
		if row.ChartID == chartID {
			delete(r.records, id)
		}
	}
	return nil
}

func (r *thresholdHandlerRepo) UpdateEnable(ctx context.Context, id int64, enable bool) error {
	_ = ctx
	row, ok := r.records[id]
	if !ok {
		return gorm.ErrRecordNotFound
	}
	row.Enable = enable
	return nil
}

func (r *thresholdHandlerRepo) UpdateRecipients(ctx context.Context, ids []int64, users, roles, emails, larkGroups, larksuiteGroups, webhooks string) error {
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

func (r *thresholdHandlerRepo) Pager(ctx context.Context, req *thresholddomain.GridRequest, goPage, pageSize int) ([]*thresholddomain.GridVO, int64, error) {
	_ = ctx
	if goPage < 1 {
		goPage = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	rows := make([]*auto.XpackThresholdInfo, 0, len(r.records))
	for _, row := range r.records {
		cloned := *row
		rows = append(rows, &cloned)
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].CreateTime > rows[j].CreateTime })

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
		result = append(result, &thresholddomain.GridVO{ID: row.ID, Name: row.Name, ChartID: row.ChartID, ResourceID: row.ResourceID, ResourceType: row.ResourceType, ChartType: row.ChartType, Enable: row.Enable, Status: row.Status, Creator: row.Creator, CreateName: row.CreatorName, CreateTime: row.CreateTime})
	}
	total := int64(len(result))
	start := (goPage - 1) * pageSize
	if start >= len(result) {
		return []*thresholddomain.GridVO{}, total, nil
	}
	end := min(start+pageSize, len(result))
	return result[start:end], total, nil
}

func (r *thresholdHandlerRepo) ExistsByChartID(ctx context.Context, chartID int64) (bool, error) {
	_ = ctx
	for _, row := range r.records {
		if row.ChartID == chartID {
			return true, nil
		}
	}
	return false, nil
}

func (r *thresholdHandlerRepo) InstancePager(ctx context.Context, req *thresholddomain.InstanceRequest, goPage, pageSize int) ([]*thresholddomain.InstanceVO, int64, error) {
	_ = ctx
	_ = req
	_ = goPage
	_ = pageSize
	return []*thresholddomain.InstanceVO{}, 0, nil
}

type thresholdResponseEnvelope struct {
	Code string          `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

func TestThresholdHandler_Save_BadRequest(t *testing.T) {
	engine, _, _ := newThresholdHandlerTestEnv(t)
	resp := performThresholdRequest(t, engine, http.MethodPost, "/threshold/save", []byte{})
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "10001", resp.Body.Code)
	assert.NotEmpty(t, resp.Body.Msg)
}

func TestThresholdHandler_Pager_Success(t *testing.T) {
	engine, svc, _ := newThresholdHandlerTestEnv(t)
	_, err := svc.Create(context.Background(), thresholdHandlerCreateRequest("sales threshold", 701, 801), 9, "pager-user", 1)
	require.NoError(t, err)
	_, err = svc.Create(context.Background(), thresholdHandlerCreateRequest("ops threshold", 702, 802), 9, "pager-user", 1)
	require.NoError(t, err)

	body, err := json.Marshal(map[string]any{"keyword": "sales", "resourceTable": "core"})
	require.NoError(t, err)
	resp := performThresholdRequest(t, engine, http.MethodPost, "/threshold/pager/1/10", body)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "000000", resp.Body.Code)

	var page struct {
		List    []thresholddomain.GridVO `json:"list"`
		Total   int64                    `json:"total"`
		Current int                      `json:"current"`
		Size    int                      `json:"size"`
	}
	require.NoError(t, json.Unmarshal(resp.Body.Data, &page))
	assert.Equal(t, int64(1), page.Total)
	require.Len(t, page.List, 1)
	assert.Equal(t, "sales threshold", page.List[0].Name)
	assert.Equal(t, 1, page.Current)
	assert.Equal(t, 10, page.Size)
}

func TestThresholdHandler_FormInfo_PathParams(t *testing.T) {
	engine, svc, _ := newThresholdHandlerTestEnv(t)
	created, err := svc.Create(context.Background(), thresholdHandlerCreateRequest("form threshold", 703, 803), 5, "form-user", 1)
	require.NoError(t, err)

	resp := performThresholdRequest(t, engine, http.MethodGet, "/threshold/formInfo/"+jsonNumber(created.ID)+"/core", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "000000", resp.Body.Code)

	var form thresholddomain.CreateRequest
	require.NoError(t, json.Unmarshal(resp.Body.Data, &form))
	assert.Equal(t, created.ID, form.ID)
	assert.Equal(t, "form threshold", form.Name)
	assert.Equal(t, "core", form.ResourceTable)
}

func TestThresholdHandler_Delete_ResourceTable(t *testing.T) {
	engine, svc, repo := newThresholdHandlerTestEnv(t)
	created, err := svc.Create(context.Background(), thresholdHandlerCreateRequest("delete threshold", 704, 804), 7, "delete-user", 1)
	require.NoError(t, err)

	body, err := json.Marshal([]int64{created.ID})
	require.NoError(t, err)
	resp := performThresholdRequest(t, engine, http.MethodPost, "/threshold/delete/snapshot", body)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "000000", resp.Body.Code)
	_, getErr := repo.GetByID(context.Background(), created.ID)
	require.NoError(t, getErr)

	resp = performThresholdRequest(t, engine, http.MethodPost, "/threshold/delete/core", body)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "000000", resp.Body.Code)
	_, getErr = repo.GetByID(context.Background(), created.ID)
	assert.ErrorIs(t, getErr, gorm.ErrRecordNotFound)
}

func TestThresholdHandler_AnyThreshold_BoolResponse(t *testing.T) {
	engine, svc, _ := newThresholdHandlerTestEnv(t)
	created, err := svc.Create(context.Background(), thresholdHandlerCreateRequest("bool threshold", 705, 805), 8, "bool-user", 1)
	require.NoError(t, err)

	for _, tc := range []struct {
		name     string
		path     string
		expected bool
	}{
		{name: "true", path: "/threshold/anyThreshold/" + jsonNumber(created.ChartID) + "/core", expected: true},
		{name: "false", path: "/threshold/anyThreshold/999999/core", expected: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resp := performThresholdRequest(t, engine, http.MethodGet, tc.path, nil)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
			assert.Equal(t, "000000", resp.Body.Code)

			var result bool
			require.NoError(t, json.Unmarshal(resp.Body.Data, &result))
			assert.Equal(t, tc.expected, result)
		})
	}
}

type thresholdHTTPResult struct {
	StatusCode int
	Body       thresholdResponseEnvelope
}

func newThresholdHandlerTestEnv(t *testing.T) (*gin.Engine, *service.ThresholdService, *thresholdHandlerRepo) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	repo := newThresholdHandlerRepo()
	svc := service.NewThresholdService(repo)
	h := NewThresholdHandler(svc)
	engine := gin.New()
	RegisterThresholdRoutes(engine, h, nil, nil)
	return engine, svc, repo
}

func performThresholdRequest(t *testing.T, engine *gin.Engine, method, path string, body []byte) thresholdHTTPResult {
	t.Helper()
	reader := bytes.NewReader(body)
	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	var resp thresholdResponseEnvelope
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	return thresholdHTTPResult{StatusCode: w.Code, Body: resp}
}

func thresholdHandlerCreateRequest(name string, chartID, resourceID int64) *thresholddomain.CreateRequest {
	enabled := true
	rateType := 1
	msgType := 0
	repeatSend := true
	showFieldValue := true
	return &thresholddomain.CreateRequest{
		BaseReciDTO: thresholddomain.BaseReciDTO{
			ReciFlagList:       []int{1},
			UIDList:            []string{"user-1"},
			RIDList:            []string{"role-1"},
			EmailList:          []string{"user@example.com"},
			LarkGroupList:      []string{"group-1"},
			LarksuiteGroupList: []string{"suite-1"},
			WebhookList:        []string{"https://example.com/hook"},
		},
		Name:           name,
		Enable:         &enabled,
		RateType:       &rateType,
		RateValue:      "5m",
		ResourceID:     resourceID,
		ResourceType:   "panel",
		ChartID:        chartID,
		ChartType:      "bar",
		ThresholdRules: `{"logic":"and"}`,
		MsgType:        &msgType,
		MsgTitle:       "threshold title",
		MsgContent:     "threshold content",
		RepeatSend:     &repeatSend,
		ShowFieldValue: &showFieldValue,
		ResourceTable:  "core",
	}
}

func jsonNumber(v int64) string {
	return strconv.FormatInt(v, 10)
}
