# Frontend E2E Test Account Configuration

## Overview

This document describes the E2E test account configuration for DataEase frontend system smoke tests.

## Test Account

- **Username**: `admin`
- **User ID**: 1
- **Password Source of Truth**: `E2E_PASSWORD`（GitHub Actions secret / 本地环境变量）

> 说明：为避免口径漂移，不在此文档中维护固定默认密码；请始终以运行环境注入的 `E2E_PASSWORD` 为准。

## Required Permissions

### Datasource Management (SYS-SMK-006)

To run the datasource creation button test (`SYS-SMK-006`), the test account needs the following permission:

- **Permission**: `datasource:manage`
- **Permission ID**: 8
- **Permission Type**: `data`

## Permission Assignment

The admin user (user_id=1) is assigned to the admin role (role_id=1) by default. To enable datasource management:

### SQL Command (Docker)

```bash
docker exec mysql8 mysql -uroot -pAdmin168 dataease_dev -e "
INSERT INTO sys_role_perm (role_id, perm_id, create_time, update_time)
VALUES (1, 8, NOW(), NOW())
ON DUPLICATE KEY UPDATE update_time = NOW();
"
```

### Verification

After adding the permission, verify:

```bash
docker exec mysql8 mysql -uroot -pAdmin168 dataease_dev -e "
SELECT rp.role_id, r.role_name, p.perm_key, p.perm_name
FROM sys_role_perm rp
JOIN sys_role r ON rp.role_id = r.role_id
JOIN sys_perm p ON rp.perm_id = p.perm_id
WHERE rp.role_id = 1 AND p.perm_key = 'datasource:manage';
"
```

Expected output:
```
role_id | role_name | perm_key           | perm_name
1      | admin     | datasource:manage | datasource:manage
```

## CI/CD Configuration

### GitHub Secrets

Configure the following secrets in your GitHub repository:

- `E2E_BASE_URL`: Test environment URL (e.g., `http://localhost:18080`)
- `E2E_USERNAME`: Test account username (default: `admin`)
- `E2E_PASSWORD`: Test account password

### Local Testing

For local testing with Docker Compose:

```bash
# Start services
docker compose -f infra/compose/docker-compose.yml up -d

# Add permission (one-time setup)
docker exec mysql8 mysql -uroot -pAdmin168 dataease_dev -e "
INSERT INTO sys_role_perm (role_id, perm_id, create_time, update_time)
VALUES (1, 8, NOW(), NOW())
ON DUPLICATE KEY UPDATE update_time = NOW();
"

# Run tests
cd apps/frontend
E2E_BASE_URL=http://localhost:18080 \
E2E_USERNAME=admin \
E2E_PASSWORD='<your-e2e-password>' \
npm run e2e:system-smoke
```

## Test Results

After proper configuration:

- `SYS-SMK-004`: ✅ Login success
- `SYS-SMK-005`: ✅ Datasource list navigation
- `SYS-SMK-006`: ✅ Datasource create button visibility

## Troubleshooting

### Permission Not Granted

If `SYS-SMK-006` is skipped:

```bash
# Check if permission exists
docker exec mysql8 mysql -uroot -pAdmin168 dataease_dev -e "
SELECT * FROM sys_perm WHERE perm_key = 'datasource:manage';
"

# Check role-permission mapping
docker exec mysql8 mysql -uroot -pAdmin168 dataease_dev -e "
SELECT * FROM sys_role_perm WHERE role_id = 1 AND perm_id = 8;
"
```

### Wrong Password

If login fails:

```bash
# Check admin user
docker exec mysql8 mysql -uroot -pAdmin168 dataease_dev -e "
SELECT user_id, username, enabled FROM sys_user WHERE username = 'admin';
"

# Confirm E2E_PASSWORD is set in current shell/session
echo "$E2E_PASSWORD" | wc -c
```

## Maintenance

### Adding New Permissions

For future E2E tests requiring additional permissions:

1. Find the permission key in `sys_perm` table
2. Add role-permission mapping in `sys_role_perm` table
3. Update this documentation

### Password Rotation

When rotating test account passwords:

1. Update GitHub secret `E2E_PASSWORD`
2. Update local `.env` files
3. Update this documentation

## Related Documentation

- [Frontend E2E System Smoke Matrix](./frontend-e2e-system-smoke-matrix.md)
- [Frontend Quality Baseline](./frontend-quality-baseline.md)
