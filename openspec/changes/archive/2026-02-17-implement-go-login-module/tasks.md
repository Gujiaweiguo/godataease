# Plan: Go Login 模块实现

## 任务清单

### AUTH-001 实体定义
- [x] 定义 PwdLoginDTO 请求结构
- [x] 定义 TokenVO 响应结构
- [x] 定义登录配置结构

### AUTH-002 Service 层
- [x] 实现 JWT Token 生成（HMAC-SHA256）
- [x] 实现 LocalLogin 方法
- [x] 实现密码验证

### AUTH-003 Handler 层
- [x] 实现 POST /login/localLogin
- [x] 实现 GET /logout
- [x] 实现统一的响应格式

### AUTH-004 集成与验证
- [x] 注册路由
- [x] OpenSpec 验证通过
- [x] 代码通过编译检查

## 里程碑

- [x] M1: 实体定义完成
- [x] M2: Service 层完成
- [x] M3: Handler 层完成
- [x] M4: 集成验证通过

## 技术要点

### JWT Token 格式
- 算法：HMAC-SHA256
- Claims：uid（用户ID）、oid（组织ID）
- 密钥：MD5(password)

### 登录验证
- 仅支持 admin 账号
- 密码从环境变量 ADMIN_PASSWORD 读取（默认 dataease）
- 简化版：不做 RSA 加解密

### 响应格式
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "exp": 0
}
```
