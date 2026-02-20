package handler

import (
	"encoding/json"
	"errors"
	"io"
	"strconv"
	"strings"

	"dataease/backend/internal/domain/chart"
	"dataease/backend/internal/domain/dataset"
	"dataease/backend/internal/domain/datasource"
	"dataease/backend/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

func RegisterCompatibilityBridgeRoutes(r gin.IRouter, user *UserHandler, org *OrgHandler, datasourceHandler *DatasourceHandler, datasetHandler *DatasetHandler, chartHandler *ChartHandler) {
	_ = user
	_ = org
	getCurrentUserID := func(c *gin.Context) int64 {
		if uid, exists := c.Get("user_id"); exists {
			switch v := uid.(type) {
			case int64:
				return v
			case uint64:
				return int64(v)
			case int:
				return int64(v)
			case float64:
				return int64(v)
			}
		}
		return 1
	}

	getCurrentUsername := func(c *gin.Context) string {
		if username, exists := c.Get("username"); exists {
			if s, ok := username.(string); ok {
				return s
			}
		}
		return "admin"
	}

	if datasourceHandler != nil {
		datasourceGroup := r.Group("/datasource")
		{
			datasourceGroup.POST("/list", datasourceHandler.List)
			datasourceGroup.POST("/tree", func(c *gin.Context) {
				var req datasource.ListRequest
				if err := c.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
					response.Error(c, "500000", "Invalid request: "+err.Error())
					return
				}

				list, err := datasourceHandler.service.Tree(&req)
				if err != nil {
					response.Error(c, "500000", "Failed: "+err.Error())
					return
				}
				response.Success(c, list)
			})
			datasourceGroup.POST("/validate", datasourceHandler.Validate)
			datasourceGroup.GET("/validate/:id", func(c *gin.Context) {
				id, err := strconv.ParseInt(c.Param("id"), 10, 64)
				if err != nil {
					response.Error(c, "500000", "Invalid datasource ID")
					return
				}
				result, err := datasourceHandler.service.ValidateByID(id)
				if err != nil {
					response.Error(c, "500000", "Failed: "+err.Error())
					return
				}
				response.Success(c, result)
			})
			datasourceGroup.POST("/types", func(c *gin.Context) {
				response.Success(c, []map[string]string{
					{"type": "MySQL", "name": "MySQL"},
					{"type": "PostgreSQL", "name": "PostgreSQL"},
					{"type": "SQLServer", "name": "SQL Server"},
					{"type": "Oracle", "name": "Oracle"},
					{"type": "Excel", "name": "Excel"},
				})
			})
			datasourceGroup.POST("/getTables", func(c *gin.Context) {
				req, ok := parseTableRequest(c)
				if !ok {
					return
				}
				result, err := datasourceHandler.service.GetTables(req)
				if err != nil {
					response.Error(c, "500000", "Failed: "+err.Error())
					return
				}
				response.Success(c, result)
			})
			datasourceGroup.POST("/getTableStatus", func(c *gin.Context) {
				req, ok := parseTableRequest(c)
				if !ok {
					return
				}
				result, err := datasourceHandler.service.GetTableStatus(req)
				if err != nil {
					response.Error(c, "500000", "Failed: "+err.Error())
					return
				}
				response.Success(c, result)
			})
			datasourceGroup.POST("/getSchema", func(c *gin.Context) {
				result, err := datasourceHandler.service.GetSchema()
				if err != nil {
					response.Error(c, "500000", "Failed: "+err.Error())
					return
				}
				response.Success(c, result)
			})
			datasourceGroup.POST("/getTableField", func(c *gin.Context) {
				req, ok := parseTableRequest(c)
				if !ok {
					return
				}
				result, err := datasourceHandler.service.GetTableField(req)
				if err != nil {
					response.Error(c, "500000", "Failed: "+err.Error())
					return
				}
				response.Success(c, result)
			})
			datasourceGroup.POST("/previewData", func(c *gin.Context) {
				req, ok := parseTableRequest(c)
				if !ok {
					return
				}
				result, err := datasourceHandler.service.PreviewData(req)
				if err != nil {
					response.Error(c, "500000", "Failed: "+err.Error())
					return
				}
				response.Success(c, result)
			})
			datasourceGroup.GET("/get/:id", func(c *gin.Context) {
				id, err := strconv.ParseInt(c.Param("id"), 10, 64)
				if err != nil {
					response.Error(c, "500000", "Invalid datasource ID")
					return
				}
				result, err := datasourceHandler.service.GetByID(id)
				if err != nil {
					response.Error(c, "500000", "Failed: "+err.Error())
					return
				}
				response.Success(c, result)
			})
			datasourceGroup.GET("/hidePw/:id", func(c *gin.Context) {
				id, err := strconv.ParseInt(c.Param("id"), 10, 64)
				if err != nil {
					response.Error(c, "500000", "Invalid datasource ID")
					return
				}
				result, err := datasourceHandler.service.GetByID(id)
				if err != nil {
					response.Error(c, "500000", "Failed: "+err.Error())
					return
				}
				response.Success(c, result)
			})
			datasourceGroup.GET("/getSimpleDs/:id", func(c *gin.Context) {
				id, err := strconv.ParseInt(c.Param("id"), 10, 64)
				if err != nil {
					response.Error(c, "500000", "Invalid datasource ID")
					return
				}
				result, err := datasourceHandler.service.GetByID(id)
				if err != nil {
					response.Error(c, "500000", "Failed: "+err.Error())
					return
				}
				response.Success(c, gin.H{"id": result.ID, "name": result.Name, "type": result.Type})
			})
			datasourceGroup.GET("/showFinishPage", func(c *gin.Context) {
				userID := getCurrentUserID(c)
				result, err := datasourceHandler.service.ShowFinishPage(userID)
				if err != nil {
					response.Error(c, "500000", "Failed: "+err.Error())
					return
				}
				response.Success(c, result)
			})
			datasourceGroup.POST("/setShowFinishPage", func(c *gin.Context) {
				userID := getCurrentUserID(c)
				if err := datasourceHandler.service.SetShowFinishPage(userID); err != nil {
					response.Error(c, "500000", "Failed: "+err.Error())
					return
				}
				response.Success(c, nil)
			})
			datasourceGroup.POST("/latestUse", func(c *gin.Context) {
				username := getCurrentUsername(c)
				result, err := datasourceHandler.service.LatestTypes(username)
				if err != nil {
					response.Error(c, "500000", "Failed: "+err.Error())
					return
				}
				response.Success(c, result)
			})
			datasourceGroup.POST("/save", func(c *gin.Context) {
				req, ok := parseDatasourceWriteRequest(c, true)
				if !ok {
					return
				}
				result, err := datasourceHandler.service.Save(req)
				if err != nil {
					response.Error(c, "500000", "Failed: "+err.Error())
					return
				}
				response.Success(c, result)
			})
			datasourceGroup.POST("/update", func(c *gin.Context) {
				req, ok := parseDatasourceWriteRequest(c, true)
				if !ok {
					return
				}
				result, err := datasourceHandler.service.Update(req)
				if err != nil {
					response.Error(c, "500000", "Failed: "+err.Error())
					return
				}
				response.Success(c, result)
			})
			datasourceGroup.POST("/move", func(c *gin.Context) {
				var body map[string]interface{}
				if err := c.ShouldBindJSON(&body); err != nil {
					response.Error(c, "500000", "Invalid request: "+err.Error())
					return
				}
				id, ok := parseInt64Value(body["id"])
				if !ok || id <= 0 {
					response.Error(c, "500000", "Invalid datasource ID")
					return
				}
				pid, ok := parseInt64Value(body["pid"])
				if !ok {
					pid = 0
				}
				result, err := datasourceHandler.service.Move(id, pid)
				if err != nil {
					response.Error(c, "500000", "Failed: "+err.Error())
					return
				}
				response.Success(c, result)
			})
			datasourceGroup.POST("/reName", func(c *gin.Context) {
				var body map[string]interface{}
				if err := c.ShouldBindJSON(&body); err != nil {
					response.Error(c, "500000", "Invalid request: "+err.Error())
					return
				}
				id, ok := parseInt64Value(body["id"])
				if !ok || id <= 0 {
					response.Error(c, "500000", "Invalid datasource ID")
					return
				}
				name, _ := body["name"].(string)
				result, err := datasourceHandler.service.Rename(id, name)
				if err != nil {
					response.Error(c, "500000", "Failed: "+err.Error())
					return
				}
				response.Success(c, result)
			})
			datasourceGroup.POST("/createFolder", func(c *gin.Context) {
				var body map[string]interface{}
				if err := c.ShouldBindJSON(&body); err != nil {
					response.Error(c, "500000", "Invalid request: "+err.Error())
					return
				}
				name, _ := body["name"].(string)
				pid, ok := parseInt64Value(body["pid"])
				if !ok {
					pid = 0
				}
				result, err := datasourceHandler.service.CreateFolder(name, pid)
				if err != nil {
					response.Error(c, "500000", "Failed: "+err.Error())
					return
				}
				response.Success(c, result)
			})
			datasourceGroup.POST("/checkRepeat", func(c *gin.Context) {
				req, ok := parseDatasourceWriteRequest(c, false)
				if !ok {
					return
				}
				result, err := datasourceHandler.service.CheckRepeat(req)
				if err != nil {
					response.Error(c, "500000", "Failed: "+err.Error())
					return
				}
				response.Success(c, result)
			})
			datasourceGroup.POST("/checkApiDatasource", func(c *gin.Context) {
				response.Error(c, "501000", "Not implemented: checkApiDatasource")
			})
			datasourceGroup.POST("/loadRemoteFile", func(c *gin.Context) {
				response.Error(c, "501000", "Not implemented: loadRemoteFile")
			})
			datasourceGroup.POST("/syncApiTable", func(c *gin.Context) {
				response.Error(c, "501000", "Not implemented: syncApiTable")
			})
			datasourceGroup.POST("/syncApiDs", func(c *gin.Context) {
				response.Error(c, "501000", "Not implemented: syncApiDs")
			})
			datasourceGroup.POST("/uploadFile", func(c *gin.Context) {
				response.Error(c, "501000", "Not implemented: uploadFile")
			})
			datasourceGroup.POST("/listSyncRecord/:dsId/:page/:limit", func(c *gin.Context) {
				response.Error(c, "501000", "Not implemented: listSyncRecord")
			})
			datasourceGroup.GET("/delete/:id", func(c *gin.Context) {
				id, err := strconv.ParseInt(c.Param("id"), 10, 64)
				if err != nil {
					response.Error(c, "500000", "Invalid datasource ID")
					return
				}
				if err = datasourceHandler.service.Delete(id); err != nil {
					response.Error(c, "500000", "Failed: "+err.Error())
					return
				}
				response.Success(c, nil)
			})
			datasourceGroup.POST("/perDelete/:id", func(c *gin.Context) {
				id, err := strconv.ParseInt(c.Param("id"), 10, 64)
				if err != nil {
					response.Error(c, "500000", "Invalid datasource ID")
					return
				}
				result, err := datasourceHandler.service.PerDelete(id)
				if err != nil {
					response.Error(c, "500000", "Failed: "+err.Error())
					return
				}
				response.Success(c, result)
			})
		}
	}

	if datasetHandler != nil {
		datasetTreeGroup := r.Group("/datasetTree")
		{
			datasetTreeGroup.POST("/tree", datasetHandler.Tree)
			datasetTreeGroup.POST("/get/:id", func(c *gin.Context) {
				id, err := strconv.ParseInt(c.Param("id"), 10, 64)
				if err != nil {
					response.Error(c, "500000", "Invalid dataset ID")
					return
				}
				result, err := buildDatasetDetail(datasetHandler, id)
				if err != nil {
					response.Error(c, "500000", "Failed: "+err.Error())
					return
				}
				response.Success(c, result)
			})
			datasetTreeGroup.POST("/details/:id", func(c *gin.Context) {
				id, err := strconv.ParseInt(c.Param("id"), 10, 64)
				if err != nil {
					response.Error(c, "500000", "Invalid dataset ID")
					return
				}
				result, err := buildDatasetDetail(datasetHandler, id)
				if err != nil {
					response.Error(c, "500000", "Failed: "+err.Error())
					return
				}
				response.Success(c, result)
			})
			datasetTreeGroup.POST("/dsDetails", func(c *gin.Context) {
				ids, ok := parseDatasetIDs(c)
				if !ok {
					return
				}
				result := make([]gin.H, 0, len(ids))
				for _, id := range ids {
					detail, err := buildDatasetDetail(datasetHandler, id)
					if err != nil {
						continue
					}
					result = append(result, detail)
				}
				response.Success(c, result)
			})
			datasetTreeGroup.POST("/detailWithPerm", func(c *gin.Context) {
				ids, ok := parseDatasetIDs(c)
				if !ok {
					return
				}
				result := make([]gin.H, 0, len(ids))
				for _, id := range ids {
					detail, err := buildDatasetDetail(datasetHandler, id)
					if err != nil {
						continue
					}
					result = append(result, detail)
				}
				response.Success(c, result)
			})
			datasetTreeGroup.POST("/getSqlParams", func(c *gin.Context) {
				ids, ok := parseDatasetIDs(c)
				if !ok {
					return
				}
				result, err := datasetHandler.service.GetSQLParams(ids)
				if err != nil {
					response.Error(c, "500000", "Failed: "+err.Error())
					return
				}
				response.Success(c, result)
			})
			datasetTreeGroup.POST("/save", func(c *gin.Context) {
				req, ok := parseDatasetWriteRequest(c, true)
				if !ok {
					return
				}
				result, err := datasetHandler.service.Save(req)
				if err != nil {
					response.Error(c, "500000", "Failed: "+err.Error())
					return
				}
				response.Success(c, result)
			})
			datasetTreeGroup.POST("/create", func(c *gin.Context) {
				req, ok := parseDatasetWriteRequest(c, true)
				if !ok {
					return
				}
				result, err := datasetHandler.service.Create(req)
				if err != nil {
					response.Error(c, "500000", "Failed: "+err.Error())
					return
				}
				response.Success(c, result)
			})
			datasetTreeGroup.POST("/rename", func(c *gin.Context) {
				req, ok := parseDatasetWriteRequest(c, true)
				if !ok {
					return
				}
				if req.ID <= 0 {
					response.Error(c, "500000", "Invalid dataset ID")
					return
				}
				result, err := datasetHandler.service.Rename(req.ID, req.Name)
				if err != nil {
					response.Error(c, "500000", "Failed: "+err.Error())
					return
				}
				response.Success(c, result)
			})
			datasetTreeGroup.POST("/move", func(c *gin.Context) {
				req, ok := parseDatasetWriteRequest(c, false)
				if !ok {
					return
				}
				if req.ID <= 0 {
					response.Error(c, "500000", "Invalid dataset ID")
					return
				}
				pid := int64(0)
				if req.PID != nil {
					pid = *req.PID
				}
				result, err := datasetHandler.service.Move(req.ID, pid)
				if err != nil {
					response.Error(c, "500000", "Failed: "+err.Error())
					return
				}
				response.Success(c, result)
			})
			datasetTreeGroup.POST("/delete/:id", func(c *gin.Context) {
				id, err := strconv.ParseInt(c.Param("id"), 10, 64)
				if err != nil {
					response.Error(c, "500000", "Invalid dataset ID")
					return
				}
				if err = datasetHandler.service.Delete(id); err != nil {
					response.Error(c, "500000", "Failed: "+err.Error())
					return
				}
				response.Success(c, nil)
			})
			datasetTreeGroup.POST("/perDelete/:id", func(c *gin.Context) {
				id, err := strconv.ParseInt(c.Param("id"), 10, 64)
				if err != nil {
					response.Error(c, "500000", "Invalid dataset ID")
					return
				}
				result, err := datasetHandler.service.PerDelete(id)
				if err != nil {
					response.Error(c, "500000", "Failed: "+err.Error())
					return
				}
				response.Success(c, result)
			})
			datasetTreeGroup.GET("/barInfo/:id", func(c *gin.Context) {
				response.Error(c, "501000", "Not implemented: barInfo")
			})
			datasetTreeGroup.POST("/exportDataset", func(c *gin.Context) {
				response.Error(c, "501000", "Not implemented: exportDataset")
			})
		}

		datasetDataGroup := r.Group("/datasetData")
		{
			datasetDataGroup.POST("/tableField", datasetHandler.Fields)
			datasetDataGroup.POST("/previewData", datasetHandler.Preview)
			datasetDataGroup.POST("/getDatasetTotal", func(c *gin.Context) {
				var body map[string]interface{}
				if err := c.ShouldBindJSON(&body); err != nil {
					response.Error(c, "500000", "Invalid request: "+err.Error())
					return
				}
				id, ok := parseInt64Value(body["id"])
				if !ok {
					response.Success(c, int64(0))
					return
				}
				preview, err := datasetHandler.service.Preview(&dataset.PreviewRequest{DatasetGroupID: id, Limit: 1})
				if err != nil {
					response.Success(c, int64(0))
					return
				}
				response.Success(c, preview.Total)
			})
			datasetDataGroup.POST("/previewSql", func(c *gin.Context) {
				var req dataset.SQLPreviewRequest
				if err := c.ShouldBindJSON(&req); err != nil {
					response.Error(c, "500000", "Invalid request: "+err.Error())
					return
				}
				result, err := datasetHandler.service.PreviewSQL(&req)
				if err != nil {
					response.Error(c, "500000", "Failed: "+err.Error())
					return
				}
				response.Success(c, result)
			})
			datasetDataGroup.POST("/enumValueObj", func(c *gin.Context) {
				req, ok := parseEnumValueRequest(c)
				if !ok {
					return
				}
				result, err := datasetHandler.service.GetFieldEnumObj(req)
				if err != nil {
					response.Error(c, "500000", "Failed: "+err.Error())
					return
				}
				response.Success(c, result)
			})
			datasetDataGroup.POST("/enumValueDs", func(c *gin.Context) {
				fieldID, ok := parseEnumFieldID(c)
				if !ok {
					return
				}
				result, err := datasetHandler.service.GetFieldEnumDs(fieldID)
				if err != nil {
					response.Error(c, "500000", "Failed: "+err.Error())
					return
				}
				response.Success(c, result)
			})
			datasetDataGroup.POST("/enumValue", func(c *gin.Context) {
				req, ok := parseMultFieldValuesRequest(c)
				if !ok {
					return
				}
				result, err := datasetHandler.service.GetFieldEnum(req)
				if err != nil {
					response.Error(c, "500000", "Failed: "+err.Error())
					return
				}
				response.Success(c, result)
			})
		}
	}

	if chartHandler != nil {
		chartDataGroup := r.Group("/chartData")
		{
			chartDataGroup.POST("/getData", chartHandler.Data)
			chartDataGroup.POST("/getFieldData/:fieldId/:fieldType", func(c *gin.Context) {
				fieldID, err := strconv.ParseInt(c.Param("fieldId"), 10, 64)
				if err != nil {
					response.Error(c, "500000", "Invalid field ID")
					return
				}
				if datasetHandler == nil {
					response.Success(c, []string{})
					return
				}
				result, err := datasetHandler.service.GetFieldEnum(&dataset.MultFieldValuesRequest{FieldIDs: []int64{fieldID}, ResultMode: 1})
				if err != nil {
					response.Error(c, "500000", "Failed: "+err.Error())
					return
				}
				response.Success(c, result)
			})
			chartDataGroup.POST("/getDrillFieldData/:fieldId", func(c *gin.Context) {
				fieldID, err := strconv.ParseInt(c.Param("fieldId"), 10, 64)
				if err != nil {
					response.Error(c, "500000", "Invalid field ID")
					return
				}
				if datasetHandler == nil {
					response.Success(c, []string{})
					return
				}
				result, err := datasetHandler.service.GetFieldEnumDs(fieldID)
				if err != nil {
					response.Error(c, "500000", "Failed: "+err.Error())
					return
				}
				response.Success(c, result)
			})
			chartDataGroup.POST("/innerExportDetails", func(c *gin.Context) {
				response.Error(c, "501000", "Not implemented: innerExportDetails")
			})
			chartDataGroup.POST("/innerExportDataSetDetails", func(c *gin.Context) {
				response.Error(c, "501000", "Not implemented: innerExportDataSetDetails")
			})
		}

		chartGroup := r.Group("/chart")
		{
			chartGroup.POST("/getData", chartHandler.Data)
			chartGroup.POST("/getChart/:id", func(c *gin.Context) {
				idStr := c.Param("id")
				id, err := strconv.ParseInt(idStr, 10, 64)
				if err != nil {
					response.Error(c, "500000", "Invalid chart ID")
					return
				}

				result, err := chartHandler.service.Query(&chart.ChartQueryRequest{ID: id})
				if err != nil {
					response.Error(c, "500000", "Failed: "+err.Error())
					return
				}
				response.Success(c, result)
			})
			chartGroup.POST("/getDetail/:id", func(c *gin.Context) {
				idStr := c.Param("id")
				id, err := strconv.ParseInt(idStr, 10, 64)
				if err != nil {
					response.Error(c, "500000", "Invalid chart ID")
					return
				}

				result, err := chartHandler.service.Query(&chart.ChartQueryRequest{ID: id})
				if err != nil {
					response.Error(c, "500000", "Failed: "+err.Error())
					return
				}
				response.Success(c, result)
			})
			chartGroup.GET("/checkSameDataSet/:viewIdSource/:viewIdTarget", func(c *gin.Context) {
				sourceID, err := strconv.ParseInt(c.Param("viewIdSource"), 10, 64)
				if err != nil {
					response.Error(c, "500000", "Invalid source chart ID")
					return
				}
				targetID, err := strconv.ParseInt(c.Param("viewIdTarget"), 10, 64)
				if err != nil {
					response.Error(c, "500000", "Invalid target chart ID")
					return
				}

				source, err := chartHandler.service.Query(&chart.ChartQueryRequest{ID: sourceID})
				if err != nil {
					response.Error(c, "500000", "Failed: "+err.Error())
					return
				}
				target, err := chartHandler.service.Query(&chart.ChartQueryRequest{ID: targetID})
				if err != nil {
					response.Error(c, "500000", "Failed: "+err.Error())
					return
				}

				same := false
				if source.TableID != nil && target.TableID != nil && *source.TableID == *target.TableID {
					same = true
				}
				response.Success(c, same)
			})
			chartGroup.POST("/save", func(c *gin.Context) {
				var body map[string]interface{}
				if err := c.ShouldBindJSON(&body); err != nil {
					response.Error(c, "500000", "Invalid request: "+err.Error())
					return
				}
				result, err := chartHandler.service.SaveFromMap(body)
				if err != nil {
					response.Error(c, "500000", "Failed: "+err.Error())
					return
				}
				response.Success(c, result)
			})
			chartGroup.POST("/listByDQ/:id/:chartId", func(c *gin.Context) {
				datasetGroupID, err := strconv.ParseInt(c.Param("id"), 10, 64)
				if err != nil {
					response.Error(c, "500000", "Invalid dataset ID")
					return
				}
				chartID, err := strconv.ParseInt(c.Param("chartId"), 10, 64)
				if err != nil {
					response.Error(c, "500000", "Invalid chart ID")
					return
				}
				result, err := chartHandler.service.ListByDQ(datasetGroupID, chartID)
				if err != nil {
					response.Error(c, "500000", "Failed: "+err.Error())
					return
				}
				response.Success(c, result)
			})
			chartGroup.POST("/copyField/:id/:chartId", func(c *gin.Context) {
				id, err := strconv.ParseInt(c.Param("id"), 10, 64)
				if err != nil {
					response.Error(c, "500000", "Invalid field ID")
					return
				}
				chartID, err := strconv.ParseInt(c.Param("chartId"), 10, 64)
				if err != nil {
					response.Error(c, "500000", "Invalid chart ID")
					return
				}
				if err = chartHandler.service.CopyField(id, chartID); err != nil {
					response.Error(c, "500000", "Failed: "+err.Error())
					return
				}
				response.Success(c, nil)
			})
			chartGroup.POST("/deleteField/:id", func(c *gin.Context) {
				id, err := strconv.ParseInt(c.Param("id"), 10, 64)
				if err != nil {
					response.Error(c, "500000", "Invalid field ID")
					return
				}
				if err = chartHandler.service.DeleteField(id); err != nil {
					response.Error(c, "500000", "Failed: "+err.Error())
					return
				}
				response.Success(c, nil)
			})
			chartGroup.POST("/deleteFieldByChart/:chartId", func(c *gin.Context) {
				chartID, err := strconv.ParseInt(c.Param("chartId"), 10, 64)
				if err != nil {
					response.Error(c, "500000", "Invalid chart ID")
					return
				}
				if err = chartHandler.service.DeleteFieldByChart(chartID); err != nil {
					response.Error(c, "500000", "Failed: "+err.Error())
					return
				}
				response.Success(c, nil)
			})
		}
	}

	if user != nil {
		userGroup := r.Group("/user")
		{
			userGroup.POST("/list", user.ListUsers)
			userGroup.POST("/create", user.CreateUser)
			userGroup.POST("/edit", user.UpdateUser)
			userGroup.POST("/update", user.UpdateUser)
			userGroup.POST("/delete/:id", user.DeleteUser)
			userGroup.GET("/options", user.GetUserOptions)
			userGroup.GET("/org/option", user.GetUserOptions)
			userGroup.POST("/byCurOrg", user.ListUsers)
		}
	}

	if org != nil {
		orgGroup := r.Group("/org")
		{
			orgGroup.POST("/create", org.CreateOrg)
			orgGroup.POST("/update", org.UpdateOrg)
			orgGroup.POST("/delete/:orgId", org.DeleteOrg)
			orgGroup.GET("/list", org.ListOrgs)
			orgGroup.GET("/info/:orgId", org.GetOrgByID)
			orgGroup.GET("/tree", org.GetOrgTree)
			orgGroup.GET("/checkName", org.CheckOrgName)
			orgGroup.POST("/updateStatus", org.UpdateOrgStatus)
			orgGroup.GET("/children/:parentId", org.GetChildOrgs)
			orgGroup.POST("/mounted", func(c *gin.Context) {
				orgs, err := org.orgService.ListOrgs()
				if err != nil {
					response.Error(c, "500000", "Failed: "+err.Error())
					return
				}
				response.Success(c, orgs)
			})
		}
	}
}

func parseTableRequest(c *gin.Context) (*datasource.TableRequest, bool) {
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil && !errors.Is(err, io.EOF) {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return nil, false
	}

	request := &datasource.TableRequest{}
	if id, ok := parseInt64Value(body["datasourceId"]); ok {
		request.DatasourceID = id
	}
	if tableName, ok := body["tableName"].(string); ok {
		request.TableName = tableName
	}
	if limit, ok := parseInt64Value(body["limit"]); ok {
		request.Limit = int(limit)
	}

	return request, true
}

func parseDatasourceWriteRequest(c *gin.Context, requireName bool) (*datasource.WriteRequest, bool) {
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil && !errors.Is(err, io.EOF) {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return nil, false
	}

	req := &datasource.WriteRequest{}
	if id, ok := parseInt64Value(body["id"]); ok {
		req.ID = id
	}

	if pidVal, exists := body["pid"]; exists {
		if pid, ok := parseInt64Value(pidVal); ok {
			req.PID = &pid
		} else {
			pid := int64(0)
			req.PID = &pid
		}
	}

	if name, ok := body["name"].(string); ok {
		req.Name = name
	}
	if desc, ok := body["description"].(string); ok {
		req.Description = &desc
	} else if desc, ok := body["desc"].(string); ok {
		req.Description = &desc
	}
	if dsType, ok := body["type"].(string); ok {
		req.Type = dsType
	}
	if nodeType, ok := body["nodeType"].(string); ok {
		req.NodeType = nodeType
	}
	if editType, ok := body["editType"].(string); ok {
		req.EditType = &editType
	} else if editTypeNum, ok := parseInt64Value(body["editType"]); ok {
		tmp := strconv.FormatInt(editTypeNum, 10)
		req.EditType = &tmp
	}

	if cfg, ok := body["configuration"].(string); ok {
		req.Configuration = &cfg
	} else if cfg, ok := body["configuration"].(map[string]interface{}); ok {
		b, err := json.Marshal(cfg)
		if err != nil {
			response.Error(c, "500000", "Invalid configuration")
			return nil, false
		}
		tmp := string(b)
		req.Configuration = &tmp
	} else if cfg, ok := body["configuration"].([]interface{}); ok {
		b, err := json.Marshal(cfg)
		if err != nil {
			response.Error(c, "500000", "Invalid configuration")
			return nil, false
		}
		tmp := string(b)
		req.Configuration = &tmp
	}

	if enable, ok := body["enableDataFill"].(bool); ok {
		req.EnableDataFill = &enable
	}

	if requireName && strings.TrimSpace(req.Name) == "" {
		response.Error(c, "500000", "datasource name is required")
		return nil, false
	}

	return req, true
}

func parseDatasetWriteRequest(c *gin.Context, requireName bool) (*dataset.WriteRequest, bool) {
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil && !errors.Is(err, io.EOF) {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return nil, false
	}

	req := &dataset.WriteRequest{}
	if id, ok := parseInt64Value(body["id"]); ok {
		req.ID = id
	}

	if pidVal, exists := body["pid"]; exists {
		if pid, ok := parseInt64Value(pidVal); ok {
			req.PID = &pid
		} else {
			pid := int64(0)
			req.PID = &pid
		}
	}

	if name, ok := body["name"].(string); ok {
		req.Name = name
	}
	if nodeType, ok := body["nodeType"].(string); ok {
		req.NodeType = nodeType
	}
	if dsType, ok := body["type"].(string); ok {
		tmp := dsType
		req.Type = &tmp
	}
	if isCross, ok := body["isCross"].(bool); ok {
		req.IsCross = &isCross
	}

	if requireName && strings.TrimSpace(req.Name) == "" {
		response.Error(c, "500000", "dataset name is required")
		return nil, false
	}

	return req, true
}

func buildDatasetDetail(h *DatasetHandler, datasetGroupID int64) (gin.H, error) {
	fields, err := h.service.Fields(&dataset.FieldsRequest{DatasetGroupID: datasetGroupID})
	if err != nil {
		return nil, err
	}

	previewData := make([]map[string]interface{}, 0)
	total := int64(0)
	preview, err := h.service.Preview(&dataset.PreviewRequest{DatasetGroupID: datasetGroupID, Limit: 100})
	if err == nil {
		previewData = preview.Rows
		total = preview.Total
	}

	return gin.H{
		"id":        datasetGroupID,
		"allFields": fields,
		"data": gin.H{
			"fields": fields,
			"data":   previewData,
		},
		"total":   total,
		"union":   []gin.H{},
		"isCross": false,
	}, nil
}

func parseDatasetIDs(c *gin.Context) ([]int64, bool) {
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil && !errors.Is(err, io.EOF) {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return nil, false
	}

	ids := make([]int64, 0)
	if id, ok := parseInt64Value(body["id"]); ok {
		ids = append(ids, id)
	}
	if id, ok := parseInt64Value(body["datasetGroupId"]); ok {
		ids = append(ids, id)
	}
	if arr, ok := body["ids"].([]interface{}); ok {
		for _, item := range arr {
			if id, ok := parseInt64Value(item); ok {
				ids = append(ids, id)
			}
		}
	}
	if len(ids) == 0 {
		return []int64{}, true
	}

	uniq := make(map[int64]struct{}, len(ids))
	result := make([]int64, 0, len(ids))
	for _, id := range ids {
		if _, ok := uniq[id]; ok {
			continue
		}
		uniq[id] = struct{}{}
		result = append(result, id)
	}

	return result, true
}

func parseEnumValueRequest(c *gin.Context) (*dataset.EnumValueRequest, bool) {
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil && !errors.Is(err, io.EOF) {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return nil, false
	}

	req := &dataset.EnumValueRequest{}
	if v, ok := parseInt64Value(body["queryId"]); ok {
		req.QueryID = v
	}
	if v, ok := parseInt64Value(body["displayId"]); ok {
		req.DisplayID = v
	}
	if v, ok := parseInt64Value(body["sortId"]); ok {
		req.SortID = v
	}
	if sort, ok := body["sort"].(string); ok {
		req.Sort = sort
	}
	if searchText, ok := body["searchText"].(string); ok {
		req.SearchText = searchText
	}
	if resultMode, ok := parseInt64Value(body["resultMode"]); ok {
		req.ResultMode = int(resultMode)
	}
	req.Filter = parseEnumFilters(body["filter"])

	if req.QueryID <= 0 {
		response.Success(c, []gin.H{})
		return nil, false
	}

	return req, true
}

func parseMultFieldValuesRequest(c *gin.Context) (*dataset.MultFieldValuesRequest, bool) {
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil && !errors.Is(err, io.EOF) {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return nil, false
	}

	req := &dataset.MultFieldValuesRequest{
		FieldIDs:   make([]int64, 0),
		ResultMode: 0,
	}
	if list, ok := body["fieldIds"].([]interface{}); ok {
		for _, item := range list {
			if id, idOK := parseInt64Value(item); idOK {
				req.FieldIDs = append(req.FieldIDs, id)
			}
		}
	}
	if resultMode, ok := parseInt64Value(body["resultMode"]); ok {
		req.ResultMode = int(resultMode)
	}
	req.Filter = parseEnumFilters(body["filter"])

	uniq := make(map[int64]struct{}, len(req.FieldIDs))
	uniqueIDs := make([]int64, 0, len(req.FieldIDs))
	for _, id := range req.FieldIDs {
		if id <= 0 {
			continue
		}
		if _, ok := uniq[id]; ok {
			continue
		}
		uniq[id] = struct{}{}
		uniqueIDs = append(uniqueIDs, id)
	}
	req.FieldIDs = uniqueIDs

	return req, true
}

func parseEnumFieldID(c *gin.Context) (int64, bool) {
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil && !errors.Is(err, io.EOF) {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return 0, false
	}

	field, ok := body["field"].(map[string]interface{})
	if !ok {
		return 0, true
	}
	id, _ := parseInt64Value(field["id"])
	return id, true
}

func parseEnumFilters(v interface{}) []dataset.EnumFilter {
	items, ok := v.([]interface{})
	if !ok || len(items) == 0 {
		return []dataset.EnumFilter{}
	}

	filters := make([]dataset.EnumFilter, 0, len(items))
	for _, item := range items {
		obj, objOK := item.(map[string]interface{})
		if !objOK {
			continue
		}
		filter := dataset.EnumFilter{Value: make([]interface{}, 0)}
		if fieldID, exists := obj["fieldId"]; exists {
			switch val := fieldID.(type) {
			case string:
				filter.FieldID = strings.TrimSpace(val)
			default:
				if parsed, parsedOK := parseInt64Value(val); parsedOK {
					filter.FieldID = strconv.FormatInt(parsed, 10)
				}
			}
		}
		if op, ok := obj["operator"].(string); ok {
			filter.Operator = strings.TrimSpace(op)
		}
		if values, ok := obj["value"].([]interface{}); ok {
			filter.Value = values
		}
		filters = append(filters, filter)
	}
	return filters
}

func parseInt64Value(v interface{}) (int64, bool) {
	switch n := v.(type) {
	case float64:
		return int64(n), true
	case int64:
		return n, true
	case int:
		return int64(n), true
	case string:
		parsed, err := strconv.ParseInt(n, 10, 64)
		if err != nil {
			return 0, false
		}
		return parsed, true
	default:
		return 0, false
	}
}
