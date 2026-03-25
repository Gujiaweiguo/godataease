## 1. Legacy Resource Contract Freeze and Classification Baseline

- [x] 1.1 冻结历史 datasource / dataset / dashboard / screen 的稳定资源身份规则、父级归属规则与组织归属规则
- [x] 1.2 明确历史资源补纳管分类口径：可自动纳管、需跳过并记录、需人工修复三类边界
- [x] 1.3 形成历史资源存量基线与异常样本清单，覆盖资源类型、父级缺失、跨组织漂移和孤儿资源场景

## 2. Backend Resource Registration and Inheritance Backfill

- [x] 2.1 实现历史资源到 `sys_resource` / `sys_resource_perm` 的幂等补注册流程，避免重复记录与重复关系
- [x] 2.2 实现历史资源基于父级/分组治理边界的继承权限补算，并为跳过资源输出可追踪审计记录
- [x] 2.3 对齐资源权限查询、用户视角查询与运行时鉴权消费同一补纳管结果，避免历史资源继续走旧语义分叉

## 3. Bounded Rollout and Permission-Center Alignment

- [x] 3.1 支持按资源类型、组织或批次执行补纳管，明确最小回滚边界与重复执行策略
- [x] 3.2 对齐统一权限中心中的历史资源展示与选择行为，保证补纳管后的历史资源可在资源视角中稳定呈现
- [x] 3.3 形成异常资源处置清单与治理结论，明确哪些资源已纳管、哪些资源被跳过、哪些需要后续独立 change 处理

## 4. Verification Gate

- [x] 4.1 执行历史资源补注册、继承补算、重复执行幂等与跳过审计的后端测试及 MySQL 集成验证
- [x] 4.2 执行统一权限中心按用户 / 按资源视角的一致性验证，确认历史资源与新资源共享同一有效授权结果
- [x] 4.3 执行 datasource / dataset / dashboard / screen 的运行时回归验证，形成“历史资源已进入统一治理模型”的验收记录
