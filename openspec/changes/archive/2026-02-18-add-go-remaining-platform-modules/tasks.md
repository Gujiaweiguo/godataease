# Plan: Go 剩余平台模块迁移

## 任务清单

### PHASE-0 范围冻结
- [x] 确认剩余待迁移模块清单与优先级
- [x] 输出 Java -> Go 迁移矩阵（模块、前缀、状态、责任人）
- [x] 冻结统一响应与错误码兼容规则（沿用 compatibility-rules.md）

### OPS-001 消息与导出
- [x] 实现 `msg-center` domain/repository/service/handler（已存在）
- [x] 实现 `exportCenter` domain/repository/service/handler
- [x] 完成消息已读幂等与导出任务状态流转

### OPS-002 分享与票据
- [x] 实现 `share` domain/repository/service/handler
- [x] 实现 `ticket` domain/repository/service/handler
- [x] 完成分享鉴权、票据校验与失效策略

### OPS-003 引擎与驱动
- [x] 实现 `engine` domain/repository/service/handler
- [x] 实现 `datasourceDriver` domain/repository/service/handler
- [x] 对齐前端依赖接口（引擎能力查询、驱动管理）

### OPS-004 地图与静态资源能力
- [x] 实现 `geometry` 与 `customGeo` 相关核心接口
- [x] 实现 `staticResource`、`store`、`typeface` 核心接口
- [x] 完成文件与资源访问安全边界校验

### INT-001 集成验证
- [x] 在 `internal/transport/http/router.go` 注册所有新增模块路由
- [x] 增加关键路径测试（模块主流程 + 错误路径）
- [x] 运行 OpenSpec 严格校验

## 里程碑

- [x] M1: 范围与矩阵冻结
- [x] M2: 消息/导出完成
- [x] M3: 分享/票据完成
- [x] M4: 引擎/驱动与地图/静态资源完成
- [x] M5: 集成验证通过
