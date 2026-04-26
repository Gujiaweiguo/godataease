# 后端安全防护栏（Backend Security Guardrails）

本文档记录 DataEase Go 后端已经建立的安全防护措施，目标是让后续开发直接复用现有模式，避免把已经修掉的安全问题重新引入。

## 适用范围

- 主线后端：`apps/backend-go`
- 需要了解后端约束的前端调用方：`apps/frontend`

说明：本文档面向仓库内部开发与维护，不替代根目录 `SECURITY.md` 的漏洞披露说明。

## 1. 出站 HTTP / SSRF 防护

### 已防护面

- 市场模板抓取：`VisualizationService.fetchMarketTemplate`
- 远程 Excel 下载：`ExcelService.downloadFile`

### 关键实现位置

- `apps/backend-go/internal/service/ssrf_safe_client.go`
  - `marketTemplateHTTPClient`
  - `marketTemplateURLValidator`
  - `isBlockedIP`
  - `marketTemplateMaxResponseBytes`

### 当前策略

- 仅允许 `http` / `https`
- DNS 解析后拒绝：
  - loopback
  - link-local
  - private / unspecified（用于 SSRF 场景）
- redirect 最多 3 跳，且每一跳都重新校验
- 响应体必须设置大小上限（当前 10MB）

### 复用规则

- 新增任何“用户可控 URL -> 服务端抓取”的功能时，**禁止直接使用裸 `http.Client` / `http.Get`**。
- 必须复用 `ssrf_safe_client.go` 的校验逻辑，或保持同等强度：
  - URL 解析
  - DNS 解析后地址过滤
  - redirect 复验
  - 响应体大小限制
- `ExcelService.downloadFile` 当前不是直接复用整个 `marketTemplateHTTPClient`，而是复用它的 `Transport` 和 `CheckRedirect`；后续新增下载逻辑时也应至少复用这两部分，而不是只复制 timeout。

## 2. 数据源连通性探测边界

### 已防护面

- `apps/backend-go/internal/service/datasource_service.go: pingTCP`
- `apps/backend-go/internal/service/datasource_service.go: isBlockedPingIP`

### 当前策略

- **阻止**：
  - loopback
  - link-local
  - unspecified
  - multicast
- **允许**：
  - RFC1918 私网地址

### 为什么和 SSRF 规则不同

- SSRF 场景（模板抓取、远程文件下载）面对的是**用户提供的任意 URL**，需要拦住内网目标。
- 数据源探测面对的是 **DataEase 连接数据库** 这一核心能力，很多真实部署依赖内网数据库，**不能直接拦掉 RFC1918**。

### 复用规则

- 新增 datasource 类型时，如果实现了“先 ping / 再连”的探测逻辑，必须保持和 `pingTCP` 一致的边界。
- 不要直接复用 `isBlockedIP` 去拦掉内网数据库。

## 3. 本地文件路径约束

### 已防护面

- 审计导出下载：`AuditHandler.DownloadExportFile`
- 字体文件下载：`FontHandler.Download`
- 静态资源 Base64 读取：`StaticHandler.FindResourceAsBase64`

### 当前策略

- 任何由用户输入参与构造的本地路径，都至少应包含：
  1. `filepath.Clean()`
  2. `..` 拒绝
  3. 与受控基目录的前缀校验，确保最终路径不逃逸

说明：不是所有场景都逐字使用同一套写法。比如审计导出下载使用的是 `filepath.Abs()`（内部已包含 clean 语义）+ 受控目录前缀校验；核心原则是不允许最终路径逃逸到允许目录之外。

### 额外约束

- `AuditHandler.DownloadExportFile` 只允许下载：
  - `os.TempDir()` 下
  - 名字符合 `audit_logs_*.csv|json`
- `FontHandler.Download` 允许公开下载，但必须经过严格路径校验
- `StaticHandler.FindResourceAsBase64` 先从输入路径中提取 basename，再做 `filepath.Clean()` / `..` 拒绝；对非法路径返回空字符串，不泄露目录结构

## 4. 路由鉴权边界

### 已建立边界

- `static` 管理接口：走受保护 `/api` 组，需要 JWT
- `font` 管理接口：走受保护 `/api` 组，需要 JWT
- `font download`：**保留公开**

### 为什么 `font download` 不能直接加鉴权

- 前端通过 `@font-face` / 资源 URL 直接拉字体文件
- 浏览器字体请求通常不会自动附带 `Authorization` / `X-DE-TOKEN` header
- 后端鉴权中间件当前只识别 header token，因此把下载路由也强制鉴权会造成字体加载回归

### 复用规则

- 对“管理类接口”优先加 JWT
- 对“浏览器静态资源直连”场景，先判断浏览器真实请求行为，再决定是否允许公开访问

## 5. 仍待跟进的低优先级项

- `font download` 增加更严格的扩展名白名单
- `font delete` 增加纵深路径校验（当前 `FileTransName` 来自数据库，若记录被污染，仍可能逃逸到字体目录之外）
- `static upload` 的 `fileId` 增加路径约束
- `audit service` 把硬编码 `/tmp` 统一成 `os.TempDir()`
- `FindResourceAsBase64` 增加文件大小/数量限制

## 6. 开发检查清单

当你新增以下能力时，提交前至少自查一次：

- [ ] 是否存在用户可控 URL 被服务端主动抓取？
- [ ] 是否复用了 SSRF-safe client / URL validator？
- [ ] 是否限制了响应体大小？
- [ ] 是否存在用户输入参与本地文件路径拼接？
- [ ] 是否做了 `filepath.Clean` + `..` 拒绝 + 基目录前缀校验？
- [ ] 是否给管理接口加了 JWT 保护？
- [ ] 如果是浏览器静态资源直连，是否确认过不会因为鉴权方式导致回归？
- [ ] 如果是 datasource 探测逻辑，是否沿用了 `pingTCP` 的边界，而不是误拦 RFC1918？

## 变更记录

| 日期 | 变更内容 |
|------|----------|
| 2026-04-26 | 初始版本，覆盖 SSRF、防止任意文件读取、datasource ping 边界与路由鉴权约定 |
