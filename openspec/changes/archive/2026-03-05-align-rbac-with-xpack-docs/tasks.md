## 1. Baseline and Contract Freeze

- [x] 1.1 冻结用户/角色/组织/权限四模块的 API 契约基线与文档映射清单
- [x] 1.2 补齐导入、重置密码、角色成员管理相关接口的契约测试骨架

## 2. User Management Parity

- [x] 2.1 实现用户 Excel 模板导入（含模板校验、10MB 限制）
- [x] 2.2 实现导入“部分成功 + 错误报告”输出
- [x] 2.3 实现管理员重置密码流程并补齐审计日志

## 3. Role Management Parity

- [x] 3.1 实现角色成员管理接口（组织用户/外部用户添加、移除）
- [x] 3.2 实现唯一角色安全策略并落地后端强约束
- [x] 3.3 实现自定义角色继承约束校验（不可越权）

## 4. Permission Configuration Parity

- [x] 4.1 实现"按用户/按资源"双视角接口一致性校验
- [x] 4.2 实现资源分组权限继承在新增资源上的自动生效
- [x] 4.3 补齐前端权限配置页与后端能力对齐联调
## 5. Organization Behavior Alignment

- [x] 5.1 明确并实现组织删除资源处置策略（与文档一致）
- [x] 5.2 增加子组织拦截与资源处置审计验证
## 6. Verification and Rollout

- [x] 6.1 执行后端 `make test` 与关键集成测试
- [x] 6.2 执行前端 `npm run lint`、`npm run ts:check` 与关键页面回归
- [x] 6.3 完成灰度发布演练与回滚开关验证
