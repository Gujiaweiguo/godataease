# 嵌入式 BI (Embedded BI) API 文档

## 概述

DataEase 支持多维度嵌入，允许将 Dashboard、Screen、模块页面和图表嵌入到第三方系统中。

当前 API 由 Go 主线后端（`apps/backend-go`）提供；Java 后端（`legacy/backend-java`）为只读备份。

## 支持的嵌入类型

| 类型 | 端点 | 说明 |
|------|--------|------|
| Dashboard 编辑器 | `/embedded/iframe/dashboard/{id}/designer` | 嵌入可编辑的仪表板 |
| Dashboard 预览 | `/embedded/iframe/dashboard/{id}/view` | 嵌入已完成仪表板，支持交互 |
| Screen 编辑器 | `/embedded/iframe/screen/{id}/designer` | 嵌入可编辑的数据大屏 |
| Screen 预览 | `/embedded/iframe/screen/{id}/view` | 嵌入已完成数据大屏，支持交互 |
| 模块页面 | `/embedded/iframe/module/{type}/view` | 嵌入数据集、数据源等模块页面 |
| 单个图表 | `/embedded/iframe/chart/{id}/view` | 嵌入单个图表视图 |

## 认证与安全

### Token 管理

#### 1. 生成 Token

```
POST /api/embedded/iframe/tokenInfo
Content-Type: application/json

Request Body:
{
  "token": "your-token",
  "origin": "https://your-domain.com"
}

Response:
{
  "code": "000000",
  "msg": "success",
  "data": {
    "token": "embed-token",
    "allowedOrigins": ["https://your-domain.com", "https://trusted-domain.com"]
  }
}
```

#### 2. Token 刷新

- **自动刷新**：默认每 5 分钟刷新一次
- **手动触发**：通过 postMessage 发送 `REFRESH_TOKEN` 事件
- **有效期**：由后端配置（默认 1 小时）

#### 3. Origin 验证

所有嵌入请求会验证 `Origin` 头是否在允许列表中。

### DIV 嵌入模式

适用于不需要 iframe 的场景，通过 JavaScript SDK 直接集成。

```javascript
import { DataEaseEmbed } from '@dataease/embed-sdk'

const embed = new DataEaseEmbed({
  mode: 'div',
  containerId: 'dataease-container',
  apiUrl: 'https://your-dataease-instance.com'
})

embed.init()
```

## 参数传递

### 父页面 → 嵌入内容

发送参数到嵌入内容：

```javascript
window.parent.postMessage({
  type: 'SET_PARAMS',
  value: {
    filters: { status: 'active' },
    dateRange: { start: '2024-01-01', end: '2024-12-31' },
    customParam: 'value'
  }
}, '*')
```

### 嵌入内容 → 父页面

嵌入内容发送事件到父页面：

| 事件类型 | 说明 | Payload |
|----------|------|---------|
| `INIT_READY` | 嵌入内容已准备好 | `{ ready: true }` |
| `PARAM_UPDATE` | 参数已更新 | `{ params: {...} }` |
| `FILTER_CHANGE` | 筛选条件已改变 | `{ filters: {...} }` |
| `ERROR` | 发生错误 | `{ error: {...} }` |
| `LOADING` | 加载状态改变 | `{ loading: true/false }` |
| `EXPORT_READY` | 导出功能可用 | `{ type: 'image/png' }` |

## 示例代码

### 1. Iframe 嵌入 Dashboard

```html
<iframe
  id="dataease-iframe"
  src="https://your-dataease.com/embedded/iframe/dashboard/123/view?token=xxx"
  width="100%"
  height="600px"
  allow="fullscreen; clipboard-write">
</iframe>

<script>
  const iframe = document.getElementById('dataease-iframe')

  iframe.addEventListener('load', () => {
    iframe.contentWindow.postMessage({
      type: 'SET_PARAMS',
      value: { filters: { status: 'active' } }
    }, '*')
  })

  window.addEventListener('message', (event) => {
    if (event.data.type === 'EXPORT_READY') {
      // 处理导出
      const blob = event.data.blob
      const url = URL.createObjectURL(blob)
      window.open(url)
    }
  })
</script>
```

### 2. DIV 嵌入模块页面

```html
<div id="dataease-container"></div>

<script type="module">
  import { DataEaseEmbed } from '@dataease/embed-sdk'

  const embed = new DataEaseEmbed({
    mode: 'div',
    containerId: 'dataease-container',
    resourceId: 'dataset-123',
    apiUrl: 'https://your-dataease.com'
  })

  embed.init()

  embed.on('PARAM_UPDATE', (params) => {
    console.log('Received parameters:', params)
  })
</script>
```

## 安全建议

1. **使用 HTTPS**：所有嵌入请求应使用 HTTPS
2. **验证 Origin**：确保嵌入域在允许列表中
3. **Token 管理**：
   - 不要在客户端存储敏感 Token
   - 定期刷新 Token
   - 处理 Token 过期情况
4. **CSP 头**：设置适当的 Content-Security-Policy

## 限制

- **Token 有效期**：默认 1 小时，可配置
- **刷新频率**：最短 30 秒，避免频繁请求
- **Origin 列表**：每个 Token 最多 10 个允许域名
- **消息大小**：postMessage 消息不超过 1MB

## 故障排查

| 问题 | 解决方案 |
|------|----------|
| Token 失效 | 刷新 Token，检查 Origin 是否在允许列表 |
| 嵌入失败 | 检查浏览器控制台错误，验证 API 地址 |
| 参数不生效 | 确认事件监听器正确设置，检查消息格式 |
| CORS 错误 | 确保 DataEase 后端正确配置 CORS 头 |
