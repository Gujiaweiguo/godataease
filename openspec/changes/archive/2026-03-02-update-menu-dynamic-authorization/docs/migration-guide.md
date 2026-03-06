# Migration and Rollback Guide

## Overview
本文档提供菜单动态授权功能的迁移与回滚指南。

---

## Pre-Migration Checklist

- [ ] 备份 `sys_role_menu` 表（如已存在）
- [ ] 确认数据库连接配置正确
- [ ] 准备回滚脚本
- [ ] 通知相关团队成员

---

## Migration Steps

### Step 1: Execute Database Migration
```bash
# 连接到数据库
mysql -h ${DB_HOST} -u ${DB_USER} -p${DB_PASSWORD} ${DB_NAME}

# 执行迁移脚本
source migrations/mysql/20260222_create_sys_role_menu.sql
source migrations/mysql/20260222_seed_admin_menu_auth.sql
```

### Step 2: Verify Migration
```sql
-- 检查表结构
DESC sys_role_menu;

-- 检查索引
SHOW INDEX FROM sys_role_menu;

-- 检查唯一约束
SELECT * FROM information_schema.TABLE_CONSTRAINTS 
WHERE TABLE_NAME = 'sys_role_menu' AND CONSTRAINT_TYPE = 'UNIQUE';

-- 检查初始数据
SELECT * FROM sys_role_menu WHERE role_id = 1 LIMIT 10;
```

### Step 3: Verify Application Startup
```bash
# 启动应用
./dataease-backend

# 检查日志中的迁移完成信息
# 期望输出: "Database migration completed" with table count
```

### Step 4: Verify Menu API
```bash
# 使用管理员账号登录获取 token
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"your_password"}'

# 查询菜单树
curl -X GET http://localhost:8080/api/menu/query \
  -H "Authorization: Bearer ${TOKEN}"

# 期望返回: 菜单树结构
```

---

## Rollback Procedure

### Emergency Rollback
如果迁移导致问题，立即执行回滚：

```bash
# 停止应用
kill $(pgrep dataease-backend)

# 执行回滚脚本
mysql -h ${DB_HOST} -u ${DB_USER} -p${DB_PASSWORD} ${DB_NAME} < migrations/mysql/20260222_rollback_sys_role_menu.sql

# 恢复备份（如有）
# mysql -h ${DB_HOST} -u ${DB_USER} -p${DB_PASSWORD} ${DB_NAME} < backup_sys_role_menu.sql

# 重启应用
./dataease-backend
```

### Rollback Verification
```sql
-- 确认表已删除
SELECT COUNT(*) FROM information_schema.TABLES 
WHERE TABLE_SCHEMA = '${DB_NAME}' AND TABLE_NAME = 'sys_role_menu';
-- 期望结果: 0

-- 确认应用日志无错误
```

---

## Performance Expectations

### Migration Time
- 表创建: < 1 秒
- 索引创建: < 2 秒
- 初始数据填充: < 5 秒（取决于菜单数量）
- 总预计时间: < 10 秒

### Rollback Time
- 表删除: < 1 秒
- 总预计时间: < 2 秒

---

## Rollback Time Metrics Template

```
Migration Start Time: ________
Migration End Time: ________
Total Duration: ________

Rollback Start Time: ________
Rollback End Time: ________
Total Duration: ________

Data Restoration Verified: [ ] Yes [ ] No
```

---

## Staging Environment Execution Record

**Environment**: staging  
**Executor**: ________  
**Date**: ________  

### 9.1 Execute migration in staging environment
- Command executed: _______________________
- Logs: [link to logs]
- Result: [ ] Success [ ] Failure
  - If failure, reason: _______________________

### 9.2 Verify bootstrap admin menu mappings
- Method: [ ] SQL [ ] API [ ] UI
- Query/Command used: _______________________
- Evidence: [screenshot/query result link]
- Result: [ ] Correct [ ] Incorrect

### 9.3 Execute rollback drill
- Rollback command: _______________________
- Restore verification method: _______________________
- Evidence: [link to evidence]
- Result: [ ] Success [ ] Failure

### 9.4 Rollback time metrics
- Rollback start time: _______________________
- Rollback end time: _______________________
- Total duration: _______________________
- Record document: [link to detailed record]

---

**Document Version**: 1.0  
**Last Updated**: 2026-03-06  
**Author**: Gujiaweiguo
