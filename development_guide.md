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

### 后端
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
dataease/
├── core/                      # 核心模块
│   ├── core-backend/          # Spring Boot 后端
│   │   └── src/main/java/io/dataease/
│   │       ├── ai/           # AI 功能（SQLBot 集成）
│   │       ├── chart/        # 图表管理
│   │       ├── dataset/      # 数据集管理
│   │       ├── datasource/   # 数据源管理
│   │       ├── embedded/     # 嵌入式 BI 功能
│   │       ├── engine/       # 查询引擎（Calcite）
│   │       ├── home/         # 首页/工作台
│   │       ├── license/      # 许可证管理
│   │       └── ...
│   └── core-frontend/       # Vue 3 前端
│       └── src/
│           ├── api/          # API 接口定义
│           ├── components/   # 通用组件
│           ├── pages/        # 页面组件
│           ├── router/       # 路由配置
│           ├── store/        # Pinia 状态管理
│           ├── views/        # 页面视图
│           └── ...
├── sdk/                      # SDK 模块
│   ├── api/                 # API 接口定义
│   ├── common/              # 通用工具类
│   └── extensions/          # 扩展模块
├── openspec/                 # OpenSpec 变更管理
└── docs/                     # 文档
```

---

## 🚀 开发环境设置

### 1. 环境要求
- **Java**: 21+
- **Node.js**: 18+
- **Maven**: 3.8+
- **MySQL**: 8.0+
- **Redis**: 7.0+

### 2. 启动方式（Docker Compose）

```bash
# 启动 Redis + MySQL + DataEase
docker-compose up -d

# 访问地址
# 前端: http://localhost:8100
# 后端 API: http://localhost:8100/api
# API 文档: http://localhost:8100/doc.html
```

### 3. 前端开发模式

```bash
cd core/core-frontend

# 安装依赖
npm install

# 启动开发服务器
npm run dev

# 访问: http://localhost:5173
```

### 4. 后端开发模式

```bash
# 编译后端
mvn clean install -DskipTests

# 启动应用（需要配置数据库连接）
cd core/core-backend
mvn spring-boot:run
```

---

## 📝 开发规范

### 1. 后端（Java + Spring Boot）

**目录结构**:
```
io.dataease.{module}
├── controller/      # 控制器层
├── service/         # 服务层
├── mapper/         # 数据访问层
├── entity/         # 实体类
├── dto/            # 数据传输对象
└── {module}/       # 功能包
```

**示例：添加新 API**
```java
@RestController
@RequestMapping("/api/my-feature")
public class MyFeatureController {

    @Autowired
    private MyFeatureService myFeatureService;

    @GetMapping("/list")
    public Result<?> list() {
        return Result.success(myFeatureService.list());
    }
}
```

### 2. 前端（Vue 3 + TypeScript）

**目录结构**:
```
src/
├── api/              # API 请求函数
├── components/       # 通用组件
├── pages/            # 页面组件
├── router/           # 路由配置
├── store/            # Pinia store
└── views/            # 页面视图
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

**添加路由**:
```typescript
// src/router/index.ts
{
  path: '/my-feature',
  name: 'MyFeature',
  component: () => import('@/views/my-feature/index.vue')
}
```

---

## ➕ 添加新功能步骤

### 步骤 1：后端开发

1. **创建实体类**
```bash
# 在对应模块下创建 entity
core/core-backend/src/main/java/io/dataease/{module}/entity/MyFeature.java
```

2. **创建 Mapper**
```bash
# 使用 MyBatis Generator 生成
cd core/core-backend
java MybatisPlusGenerator
```

3. **创建 Service**
```bash
# 创建服务层
core/core-backend/src/main/java/io/dataease/{module}/service/MyFeatureService.java
```

4. **创建 Controller**
```bash
# 创建控制器
core/core-backend/src/main/java/io/dataease/{module}/controller/MyFeatureController.java
```

### 步骤 2：前端开发

1. **创建 API 接口**
```typescript
// core/core-frontend/src/api/my-feature.ts
import request from '@/config/axios'

export const getMyFeatureList = () => {
  return request({ url: '/my-feature/list', method: 'get' })
}
```

2. **创建页面组件**
```bash
# 在 views 下创建
core/core-frontend/src/views/my-feature/
```

3. **添加路由**
```typescript
// 修改 router/index.ts
```

4. **添加菜单**（如需在菜单中显示）
- 在数据库 `sys_menu` 表中添加菜单项
- 或在前端配置菜单

### 步骤 3：数据库迁移

1. **创建 Flyway 脚本**
```bash
# 在 core/core-backend/src/main/resources/db/migration 下创建
V{version}__your_feature_name.sql
```

2. **执行迁移**
```bash
mvn flyway:migrate
```

### 步骤 4：测试

```bash
# 前端测试
cd core/core-frontend
npm run lint
npm run ts:check

# 后端测试
mvn test
```

---

## 🔧 常用命令

### 前端
```bash
cd core/core-frontend

# 开发模式
npm run dev

# 构建（前端）
npm run build:base

# 代码检查
npm run lint
npm run ts:check
```

### 后端
```bash
# 编译
mvn clean install -DskipTests

# 运行
mvn spring-boot:run

# 测试
mvn test
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
   - 参考 `openspec/changes/add-multi-embed/` 的示例

2. **遵循现有代码风格**
   - 后端：使用 Spring Boot + MyBatis Plus 规范
   - 前端：使用 Vue 3 Composition API + TypeScript

3. **提交 PR**
   - 保持 PR 小而专注
   - 确保代码通过 lint 和类型检查
   - 添加适当的注释和文档

---

## 🏗 核心模块说明

### 后端核心模块

| 模块 | 说明 | 路径 |
|------|------|------|
| **ai** | AI 功能集成（SQLBot） | `core/core-backend/src/main/java/io/dataease/ai/` |
| **chart** | 图表管理（创建、编辑、删除） | `core/core-backend/src/main/java/io/dataease/chart/` |
| **dataset** | 数据集管理（数据源、字段、SQL） | `core/core-backend/src/main/java/io/dataease/dataset/` |
| **datasource** | 数据源管理（连接、测试、类型） | `core/core-backend/src/main/java/io/dataease/datasource/` |
| **embedded** | 嵌入式 BI 功能 | `core/core-backend/src/main/java/io/dataease/embedded/` |
| **engine** | 查询引擎（Calcite SQL 解析） | `core/core-backend/src/main/java/io/dataease/engine/` |
| **home** | 首页/工作台（仪表板、数据视图） | `core/core-backend/src/main/java/io/dataease/home/` |
| **license** | 许可证管理（验证、授权） | `core/core-backend/src/main/java/io/dataease/license/` |

### 前端核心页面

| 页面 | 说明 | 路径 |
|------|------|------|
| **仪表板** | 数据可视化仪表板管理 | `core/core-frontend/src/views/dashboard/` |
| **数据视图** | 图表视图管理 | `core/core-frontend/src/views/chart/` |
| **数据集** | 数据集配置 | `core/core-frontend/src/views/dataset/` |
| **数据源** | 数据源配置 | `core/core-frontend/src/views/datasource/` |
| **系统管理** | 用户、角色、权限管理 | `core/core-frontend/src/views/system/` |

---

## 🔑 重要配置文件

| 文件 | 说明 |
|------|------|
| `pom.xml` | Maven 项目配置（根目录） |
| `core/core-backend/pom.xml` | 后端依赖配置 |
| `core/core-frontend/package.json` | 前端依赖配置 |
| `docker-compose.yml` | Docker 容器编排配置 |
| `core/core-backend/src/main/resources/application.yml` | Spring Boot 应用配置 |
| `core/core-frontend/vite.config.ts` | Vite 构建配置 |

---

## 📖 开发文档索引

- [贡献指南](./CONTRIBUTING.md)
- [行为准则](./CODE_OF_CONDUCT.md)
- [安全策略](./SECURITY.md)
- [OpenSpec 规范](./openspec/AGENTS.md)
- [嵌入式 BI 规范](./openspec/changes/add-multi-embed/proposal.md)
