# Java 后端只读治理规则

## 适用范围

`legacy/backend-java/` 目录（原 `core/core-backend/`）。

同时适用于 `legacy/sdk/`（Java SDK 模块）。

## 应急命令（仅限只读场景）

在 `/opt/code/godataease` 执行：

- 验证 Java 聚合：`mvn -f legacy/pom.xml -N validate`
- 构建 Java 模块：`mvn -f legacy/pom.xml clean install -DskipTests`

在 `/opt/code/godataease/legacy/backend-java` 执行：

- 启动后端：`mvn spring-boot:run -pl core-backend`
- Standalone 启动：`mvn spring-boot:run -pl core-backend -Dspring-boot.run.profiles=standalone`
- 打包：`mvn clean package -pl core-backend -Pstandalone -DskipTests`
- 数据库迁移：`mvn flyway:migrate -pl core-backend`

测试（默认跳过，需显式开启）：

- 全量测试：`mvn test -pl core-backend -DskipTests=false`
- 单测类：`mvn test -pl core-backend -Dtest=PermissionManageTest -DskipTests=false`
- 单测方法：`mvn test -pl core-backend -Dtest=PermissionManageTest#testMethodName -DskipTests=false`

在 `/opt/code/godataease/legacy/sdk` 执行：

- 构建 SDK：`mvn clean install -DskipTests`
- 模块：`common`, `api`, `distributed`, `extensions`

## 治理策略

### 状态：长期只读备份

Java 后端代码已进入**只读备份**状态，不再作为主线开发目录。

### 允许的改动类型

以下三类改动**允许**进入 `legacy/backend-java/`：

1. **安全补丁 (Security Patch)**
   - 修复已披露的安全漏洞
   - 必须附带 CVE 编号或安全报告引用
   - 必须经过安全团队评审

2. **应急修复 (Emergency Fix)**
   - 生产环境阻塞性问题的紧急修复
   - 必须在 24 小时内提交后续 Go 版本迁移计划
   - 必须经过技术负责人审批

3. **迁移对照 (Migration Reference)**
   - 为 Go 版本实现提供参考的注释或文档
   - 不改变代码行为
   - 仅限注释和文档改动

### 禁止的改动类型

- ❌ 新功能开发
- ❌ 重构改动
- ❌ 依赖升级（非安全必需）
- ❌ 性能优化
- ❌ 代码风格改动

## 评审门禁

### CODEOWNERS 配置

建议在 `.github/CODEOWNERS` 中添加：

```
/legacy/backend-java/ @security-team @tech-lead
```

### PR 审批要求

- 所有针对 `legacy/backend-java/` 的 PR 必须标注改动类型
- 安全补丁需要安全团队审批
- 应急修复需要技术负责人审批
- 迁移对照需要至少一位评审者确认

## 例外流程

如需突破只读限制：

1. 创建 RFC 文档说明理由
2. 获得技术委员会批准
3. 在 PR 中引用批准记录

## 过期策略

- 每季度审查一次是否需要保留 Java 备份
- 当 Go 后端功能完整度达到 100% 且稳定运行 6 个月后，可考虑归档删除

---

**生效日期**: 2026-02-19
**文档版本**: v1.0
