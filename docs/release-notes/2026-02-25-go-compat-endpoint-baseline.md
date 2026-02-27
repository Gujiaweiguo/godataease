# Go 迁移后兼容端点验收基线（2026-02-25）

## 目的

用于在每次发版后快速确认：

- 登录与首页可用
- 角色权限管理关键接口不再出现 `404/500`
- 仪表板资源树接口可用

## 关键端点基线

### 角色权限相关

- `POST /de2api/role/byCurOrg` → `200` + `code=000000`
- `GET /de2api/auth/menuPermission` → `200` + `code=000000`
- `GET /de2api/auth/busiPermission` → `200` + `code=000000`
- `POST /de2api/role/permission/save` → `200` + `code=000000`
- `POST /de2api/system/role/permission/save` → `200` + `code=000000`

### 仪表板资源树

- `POST /de2api/dataVisualization/tree` → `200` + `code=000000`

### 说明

- `xpackComponent/content/*` 的 `501` 为许可证相关预期行为，不计入本基线失败。

## API 快速回归命令

在仓库根目录 `/opt/code/godataease` 执行：

### 推荐：脚本一键检查

```bash
./infra/scripts/check-compat-baseline.sh
```

可选参数：

```bash
BASE_URL=http://localhost:8080 ROLE_ID=1 ./infra/scripts/check-compat-baseline.sh
```

脚本通过标准：全部检查项 `PASS`，进程退出码为 `0`。

### 手工命令（排障用）

```bash
curl -i -s -X POST http://localhost:8080/de2api/role/byCurOrg -H "Content-Type: application/json" -d '{}'
curl -i -s http://localhost:8080/de2api/auth/menuPermission
curl -i -s http://localhost:8080/de2api/auth/busiPermission
curl -i -s -X POST http://localhost:8080/de2api/role/permission/save -H "Content-Type: application/json" -d '{"roleId":1,"permIds":[]}'
curl -i -s -X POST http://localhost:8080/de2api/system/role/permission/save -H "Content-Type: application/json" -d '{"roleId":1,"permIds":[]}'
curl -i -s -X POST http://localhost:8080/de2api/dataVisualization/tree -H "Content-Type: application/json" -d '{"busiFlag":"dashboard-dataV"}'
```

通过标准：上述请求均返回 HTTP `200`，且响应体 `code` 为 `000000`。

## 浏览器回归基线

建议每次发版后执行一次浏览器链路回归：

1. `http://localhost:8080/login` 登录（`admin / Dataease@123`）
2. 进入角色管理页，确认无 `404/500`
3. 进入仪表板页，确认资源树加载成功

通过标准：

- 登录成功并进入工作台/仪表板页面
- 角色权限页请求 `role/byCurOrg`、`auth/menuPermission`、`auth/busiPermission`、`role/permission/save` 全部成功
- 仪表板树请求 `dataVisualization/tree` 成功

## 常见失败与处理

- `Error 1146 ... sys_role_perm doesn't exist`
  - 原因：数据库缺权限关系表
  - 处理：执行本次迁移 SQL 后复测

## 数据库迁移执行（本次新增）

在仓库根目录 `/opt/code/godataease` 执行：

```bash
./infra/scripts/apply-compat-migrations.sh
```

或手工执行：

```bash
docker exec -i mysql8 mysql -uroot -pAdmin168 -D dataease_dev < apps/backend-go/migrations/mysql/20260225_create_sys_role_perm.sql
docker exec -i mysql8 mysql -uroot -pAdmin168 -D dataease_dev < apps/backend-go/migrations/mysql/20260225_seed_sys_perm_baseline.sql
```

执行后建议立即跑：

```bash
./infra/scripts/check-compat-baseline.sh
```

- `404 page not found`（兼容端点）
  - 原因：兼容路由未注册或 `/de2api` 映射偏差
  - 处理：检查 `router` 中兼容注册与 `NoRoute` 映射逻辑
