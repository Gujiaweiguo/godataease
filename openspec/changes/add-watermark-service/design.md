## Context

水印功能是 DataEase 数据安全体系的一部分，用于在仪表板和可视化资源上显示用户标识水印，防止敏感数据泄露。前端已有 `watermark.ts` API 封装，后端需要提供对应的 API 实现。

当前状态：
- 前端 API 封装已就绪
- 后端代码已实现但未提交
- 数据库表 `visualization_watermark` 需要存在

## Goals / Non-Goals

**Goals:**
- 提供 `/watermark/find` 和 `/watermark/save` API
- 支持全局水印配置的查询和保存
- 使用 upsert 模式确保只有一个默认水印记录

**Non-Goals:**
- 面板级别的水印定制（前端已实现）
- 水印渲染逻辑（前端负责）
- 多租户水印隔离

## Decisions

1. **单一记录模式**：使用固定的 `id = "default"` 确保只有一个全局水印配置
2. **JSON 配置**：水印设置以 JSON 字符串存储，保持灵活性
3. **Upsert 实现**：使用 GORM 的 `OnConflict` 实现插入或更新

## Risks / Trade-offs

| 风险 | 缓解措施 |
|------|----------|
| JSON 格式变更 | 前端兼容处理，后端仅存储 |
| 无权限控制 | 当前依赖 API 层的通用认证中间件 |
