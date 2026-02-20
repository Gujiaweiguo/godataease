# Plan: Go Embedded 模块实现

## 任务清单

### EMB-001 实体定义 [完成]
- [x] 创建 CoreEmbedded 实体（映射 core_embedded 表）
- [x] 定义 EmbeddedCreator 请求 DTO
- [x] 定义 EmbeddedEditor 请求 DTO
- [x] 定义 EmbeddedResetRequest 请求 DTO
- [x] 定义 EmbeddedOrigin 请求 DTO
- [x] 定义 EmbeddedGridVO 响应 VO
- [x] 定义分页响应结构

### EMB-002 工具函数 [完成]
- [x] 实现 generateAppId（格式：app_{snowflakeId}）
- [x] 实现 generateAppSecret（随机字符串生成）
- [x] 实现 maskAppSecret（密钥脱敏显示）
- [x] 实现 parseDomains（域名列表解析）
- [x] 实现 isOriginAllowed（域名白名单验证）

### EMB-003 Repository 层 [完成]
- [x] 实现 EmbeddedRepository（Create、Update、Delete、GetByID）
- [x] 实现 Query 方法（支持 keyword 过滤）
- [x] 实现 GetByAppId（按 AppId 查询）
- [x] 实现 ListDistinctDomains（获取去重域名列表）
- [x] 实现分页查询

### EMB-004 Service 层 [完成]
- [x] 实现 CreateEmbedded（含 appId、appSecret 生成）
- [x] 实现 EditEmbedded
- [x] 实现 DeleteEmbedded
- [x] 实现 BatchDeleteEmbedded
- [x] 实现 ResetSecret（重置密钥）
- [x] 实现 QueryGrid（分页查询，含密钥脱敏）
- [x] 实现 GetDomainList
- [x] 实现 InitIframe（Token 验证 + 域名白名单验证）
- [x] 实现 GetTokenArgs
- [x] 实现 GetLimitCount

### EMB-005 Handler 层 [完成]
- [x] 实现 POST /api/embedded/pager/{goPage}/{pageSize}
- [x] 实现 POST /api/embedded/create
- [x] 实现 POST /api/embedded/edit
- [x] 实现 POST /api/embedded/delete/:id
- [x] 实现 POST /api/embedded/batchDelete
- [x] 实现 POST /api/embedded/reset
- [x] 实现 GET /api/embedded/domainList
- [x] 实现 POST /api/embedded/initIframe
- [x] 实现 GET /api/embedded/getTokenArgs
- [x] 实现 GET /api/embedded/limitCount
- [x] 实现统一的响应格式（code、data、msg）
- [x] 实现错误处理（返回 code: 500000）

### EMB-006 集成与验证 [完成]
- [x] 注册路由（RegisterEmbeddedRoutes）
- [x] 在 router.go 中初始化依赖
- [x] 审计日志集成（通过现有中间件）
- [x] API 响应格式与 Java 版本兼容
- [x] OpenSpec 验证通过
- [x] 代码通过编译检查

## 里程碑

- [x] M1: 实体定义完成
- [x] M2: Repository 层完成
- [x] M3: Service 层完成
- [x] M4: Handler 层完成
- [x] M5: 集成验证通过

## 文件清单

| 文件 | 说明 |
|------|------|
| `internal/domain/embedded/embedded.go` | 嵌入式实体定义 |
| `internal/repository/embedded_repo.go` | GORM Repository 层 |
| `internal/service/embedded_service.go` | 业务逻辑层 |
| `internal/transport/http/handler/embedded_handler.go` | HTTP Handler |
| `internal/transport/http/router.go` | 路由注册（需更新） |

## 数据库表结构

### core_embedded
| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGINT | 主键（自增） |
| name | VARCHAR(255) | 应用名称 |
| app_id | VARCHAR(100) | 应用 ID |
| app_secret | VARCHAR(255) | 应用密钥 |
| domain | TEXT | 允许域名列表 |
| secret_length | INT | 密钥长度 |
| create_time | BIGINT | 创建时间 |
| update_by | VARCHAR(100) | 更新人 |
| update_time | BIGINT | 更新时间 |

## 依赖关系

```
implement-go-embedded-module
├── 依赖: implement-go-audit-module ✅
├── 依赖: implement-go-user-module ✅
├── 依赖: implement-go-org-module ✅
├── 依赖: implement-go-permission-module ✅
└── 依赖: refactor-backend-to-go ✅
```

## 技术要点

### JWT Token 格式
- 算法：HMAC-SHA256
- Claims：uid（用户ID）、oid（组织ID）、appId（应用ID）、exp（过期时间）
- 过期时间：24 小时（86400000ms）

### 密钥脱敏规则
- 长度 ≤ 8：显示 `********`
- 长度 > 8：显示前4位 + `****` + 后4位

### 域名验证规则
- 支持逗号、分号、空格分隔多个域名
- 去除尾部斜杠
- 支持完整 URL 或仅主机名匹配
