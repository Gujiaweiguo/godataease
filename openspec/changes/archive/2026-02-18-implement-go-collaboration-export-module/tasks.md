# Plan: Go Collaboration + Export 模块实现

## 任务清单

### PHASE-0 边界与契约
- [x] 盘点 Java 侧 template/share/export 核心端点并冻结首批 API
- [x] 定义分享鉴权与导出权限的安全基线

### TEMPLATE-001 Template 模块
- [x] 新建 template domain/repository/service/handler
- [x] 实现模板管理核心 API（列表、创建、更新、删除）
- [x] 实现模板市场首批查询能力

### SHARE-001 Share 模块
- [x] 新建 share domain/repository/service/handler （已在 add-go-remaining-platform-modules 完成）
- [x] 实现资源分享核心 API
- [x] 实现分享票据校验与失效策略

### EXPORT-001 Export 模块
- [x] 完善 `internal/domain/export/export.go` （已在 add-go-remaining-platform-modules 完成）
- [x] 新建 export repository/service/handler
- [x] 实现导出任务创建、状态查询、下载接口

### INT-001 集成与验证
- [x] 在 `internal/transport/http/router.go` 注册 template/share/export 路由
- [x] 对齐 Java 响应与错误码契约
- [x] 增加安全审计与关键路径测试
- [x] 运行 OpenSpec 严格校验

## 里程碑

- [x] M1: 契约冻结完成
- [x] M2: Template 完成
- [x] M3: Share 完成
- [x] M4: Export 完成
- [x] M5: 集成验证通过
