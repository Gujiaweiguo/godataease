# Plan: Go Platform Ops 模块实现

## 任务清单

### PHASE-0 契约与优先级
- [x] 盘点 Java 侧 system/license/msgCenter 核心端点并冻结首批 API
- [x] 明确参数、许可证、消息能力的权限模型与审计策略

### SYS-001 系统参数模块
- [x] 新建 system parameter domain/repository/service/handler
- [x] 实现系统参数查询与更新核心接口
- [x] 实现参数变更审计记录

### LIC-001 许可证模块
- [x] 新建 license domain/repository/service/handler
- [x] 实现许可证信息查询与有效性校验接口
- [x] 实现许可证到期预警基础能力

### MSG-001 消息中心模块
- [x] 新建 msgcenter domain/repository/service/handler
- [x] 实现消息列表查询与已读状态更新接口
- [x] 实现消息状态更新幂等控制

### INT-001 集成与验证
- [x] 在 `internal/transport/http/router.go` 注册 platform-ops 路由
- [x] 对齐 Java 响应与错误码契约
- [x] 增加关键路径测试与异常回归用例
- [x] 运行 OpenSpec 严格校验

## 里程碑

- [x] M1: 契约冻结完成
- [x] M2: 系统参数模块完成
- [x] M3: 许可证模块完成
- [x] M4: 消息中心模块完成
- [x] M5: 集成验证通过
