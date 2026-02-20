# DataEase 项目开发指南

## 📌 项目概述

DataEase 是开源的 BI（商业智能）工具，支持通过拖拽方式制作数据可视化图表，支持多种数据源连接。

**版本**: 2.10.19
**许可证**: GPL v3
**官网**: https://dataease.cn

---

## 🛠 技术栈

### 前端
- **框架**: Vue.js 3 + TypeScript
- **构建工具**: Vite 4
- **UI 组件库**: Element Plus
- **图表库**: AntV (G2Plot, L7, S2)、ECharts
- **状态管理**: Pinia
- **路由**: Vue Router 4

### 后端（Go 主线）
- **框架**: Go 1.21+
- **HTTP**: Gin
- **ORM**: GORM
- **缓存**: Redis

### 后端（Java 备份 - 只读）
- **框架**: Spring Boot 3.3.0 (Java 21)
- **ORM**: MyBatis Plus 3.5.6
- **SQL 处理**: Apache Calcite 1.35.24
- **缓存**: Redis + Ehcache
- **文档工具**: Knife4j 4.4.0

### 基础设施
- **数据库**: MySQL 8
- **缓存**: Redis
- **容器**: Docker

---

## 📁 项目结构

```
godataease/
├── apps/                      # 运行时应用
│   ├── backend-go/           # Go 后端（主线）
│   │   └── internal/
│   │       ├── domain/       # 领域模型
│   │       ├── service/      # 业务逻辑
│   │       ├── repository/   # 数据访问
│   │       └── transport/    # HTTP/WebSocket
│   └── frontend/             # Vue 3 前端
│       └── src/
│           ├── api/          # API 接口定义
│           ├── components/   # 通用组件
│           ├── views/        # 页面视图
│           └── store/        # Pinia 状态管理
├── legacy/                    # 历史备份（只读）
│   ├── backend-java/         # Java 后端备份
│   │   └── core-backend/
│   │       └── src/main/java/io/dataease/
│   │           ├── ai/       # AI 功能（SQLBot 集成）
│   │           ├── chart/    # 图表管理
│   │           ├── dataset/  # 数据集管理
│   │           └── ...
│   └── sdk/                  # Java SDK 模块
├── infra/                     # 部署与运维
│   ├── compose/              # Docker Compose 配置
│   ├── scripts/              # 部署脚本
│   └── assets/               # 运维资产
├── openspec/                 # OpenSpec 变更管理
└── docs/                     # 文档
```

---

## 🚀 开发环境设置

### 1. 环境要求
- **Go**: 1.21+
- **Node.js**: 18+
- **MySQL**: 8.0+
- **Redis**: 7.0+

### 2. 启动方式（Docker Compose）

```bash
# 启动 Redis + MySQL + DataEase
cd infra/compose
docker compose up -d

# 访问地址
# 前端: http://localhost:8080
# 后端 API: http://localhost:8080/api
```

### 3. 前端开发模式

```bash
cd apps/frontend

# 安装依赖
npm install

# 启动开发服务器
npm run dev

# 访问: http://localhost:5173
```

### 4. Go 后端开发模式

```bash
cd apps/backend-go

# 安装依赖
go mod tidy

# 运行
make run

# 或直接运行
go run ./cmd/api
```

### 5. Java 后端开发模式（仅参考）

```bash
# 编译后端
mvn -f legacy/pom.xml clean install -DskipTests

# 启动应用（需要配置数据库连接）
cd legacy/backend-java/core-backend
mvn spring-boot:run
```

---

## 📝 开发规范

### 1. 后端（Go）

**目录结构**:
```
apps/backend-go/internal/
├── domain/          # 领域模型
├── service/         # 业务逻辑
├── repository/      # 数据访问
├── transport/       # HTTP/WebSocket
│   ├── handler/    # 请求处理
│   └── middleware/ # 中间件
└── pkg/            # 公共包
```

**示例：添加新 API**
```go
// internal/transport/handler/my_feature_handler.go
func (h *Handler) GetMyFeatureList(c *gin.Context) {
    list, err := h.myFeatureService.List(c.Request.Context())
    if err != nil {
        response.Error(c, err)
        return
    }
    response.Success(c, list)
}
```

### 2. 前端（Vue 3 + TypeScript）

**目录结构**:
```
apps/frontend/src/
├── api/              # API 请求函数
├── components/       # 通用组件
├── views/            # 页面视图
├── router/           # 路由配置
├── store/            # Pinia store
└── utils/            # 工具函数
```

**示例：添加新页面**
```typescript
// src/views/my-feature/index.vue
<template>
  <div class="my-feature">
    <h1>My New Feature</h1>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
</script>

<style scoped>
.my-feature {
  padding: 20px;
}
</style>
```

---

## ➕ 添加新功能步骤

### 步骤 1：Go 后端开发

1. **创建领域模型**
```bash
# 在 domain 下创建
apps/backend-go/internal/domain/myfeature/myfeature.go
```

2. **创建 Repository**
```bash
# 创建数据访问层
apps/backend-go/internal/repository/myfeature_repo.go
```

3. **创建 Service**
```bash
# 创建服务层
apps/backend-go/internal/service/myfeature_service.go
```

4. **创建 Handler**
```bash
# 创建请求处理
apps/backend-go/internal/transport/http/handler/myfeature_handler.go
```

### 步骤 2：前端开发

1. **创建 API 接口**
```typescript
// apps/frontend/src/api/my-feature.ts
import request from '@/config/axios'

export const getMyFeatureList = () => {
  return request({ url: '/my-feature/list', method: 'get' })
}
```

2. **创建页面组件**
```bash
# 在 views 下创建
apps/frontend/src/views/my-feature/
```

3. **添加路由**
```typescript
// 修改 router/index.ts
```

### 步骤 3：测试

```bash
# 前端测试
cd apps/frontend
npm run lint
npm run ts:check

# Go 后端测试
cd apps/backend-go
make test
```

---

## 🔧 常用命令

### 前端
```bash
cd apps/frontend

# 开发模式
npm run dev

# 构建
npm run build:base

# 代码检查
npm run lint
npm run ts:check
```

### Go 后端
```bash
cd apps/backend-go

# 运行
make run

# 构建
make build

# 测试
make test

# Lint
golangci-lint run
```

---

## 📚 更多资源

- **在线文档**: https://dataease.io/docs/
- **社区论坛**: https://bbs.fit2cloud.com/c/de/6
- **GitHub Issues**: https://github.com/dataease/dataease/issues
- **视频介绍**: https://www.bilibili.com/video/BV1Y8dAYLErb/

---

## 💡 开发提示

1. **使用 OpenSpec 管理大型功能**
   - 查看 `openspec/AGENTS.md` 了解如何创建功能提案
   - 参考现有 changes 目录的示例

2. **遵循现有代码风格**
   - 后端：使用 Go 标准规范
   - 前端：使用 Vue 3 Composition API + TypeScript

3. **提交 PR**
   - 保持 PR 小而专注
   - 确保代码通过 lint 和类型检查
   - 添加适当的注释和文档

---

## 🏗 核心模块说明

### Go 后端核心模块

| 模块 | 说明 | 路径 |
|------|------|------|
| **domain** | 领域模型 | `apps/backend-go/internal/domain/` |
| **service** | 业务逻辑 | `apps/backend-go/internal/service/` |
| **repository** | 数据访问 | `apps/backend-go/internal/repository/` |
| **transport** | HTTP/WebSocket | `apps/backend-go/internal/transport/` |

### Java 后端核心模块（只读参考）

| 模块 | 说明 | 路径 |
|------|------|------|
| **ai** | AI 功能集成（SQLBot） | `legacy/backend-java/core-backend/src/main/java/io/dataease/ai/` |
| **chart** | 图表管理 | `legacy/backend-java/core-backend/src/main/java/io/dataease/chart/` |
| **dataset** | 数据集管理 | `legacy/backend-java/core-backend/src/main/java/io/dataease/dataset/` |

### 前端核心页面

| 页面 | 说明 | 路径 |
|------|------|------|
| **仪表板** | 数据可视化仪表板管理 | `apps/frontend/src/views/dashboard/` |
| **数据视图** | 图表视图管理 | `apps/frontend/src/views/chart/` |
| **数据集** | 数据集配置 | `apps/frontend/src/views/dataset/` |
| **数据源** | 数据源配置 | `apps/frontend/src/views/datasource/` |
| **系统管理** | 用户、角色、权限管理 | `apps/frontend/src/views/system/` |

---

## 🔑 重要配置文件

| 文件 | 说明 |
|------|------|
| `apps/backend-go/configs/config.yaml` | Go 后端配置 |
| `apps/frontend/package.json` | 前端依赖配置 |
| `infra/compose/docker-compose.yml` | Docker 容器编排配置 |
| `apps/frontend/vite.config.ts` | Vite 构建配置 |

---

## 📖 开发文档索引

- [贡献指南](./CONTRIBUTING.md)
- [行为准则](./CODE_OF_CONDUCT.md)
- [安全策略](./SECURITY.md)
- [OpenSpec 规范](./openspec/AGENTS.md)
- [Java 后端只读规则](./legacy/README-READONLY.md)
