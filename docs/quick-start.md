# DataEase 开发者快速入门指南

## 快速开始

### 1. 环境要求

- **Go**: 1.21+
- **Node.js**: 18+
- **MySQL**: 8.0+
- **Redis**: 7.0+

### 2. 本地开发

```bash
# 克隆项目
git clone https://github.com/Gujiaweiguo/godataease.git
cd godataease

# 编译后端（Go 主线）
cd apps/backend-go
make build

# 编译前端
cd ../frontend
npm install

# 启动前端开发服务器
npm run dev
# 访问 http://localhost:5173

# 启动后端（需要配置数据库连接）
cd ../backend-go
make run
# API 访问 http://localhost:8080

# Java 后端仅为只读备份（应急场景）
# 参考 legacy/README-READONLY.md
```

### 3. 容器部署（Docker Compose）

```bash
# 在项目根目录启动
docker compose -f infra/compose/docker-compose.yml up -d --build
```

如需自定义数据库信息，请在 `infra/compose` 下创建 `.env`（可从 `.env.example` 复制）：

```env
SERVER_PORT=8080
DB_HOST=mysql
DB_PORT=3306
DB_NAME=dataease10
DB_USER=root
DB_PASSWORD=your-password
MYSQL_PORT=3306
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_EXTERNAL_PORT=6379
```

服务启动后访问：`http://localhost:8080`

停止服务：

```bash
docker compose down
```

### 4. 项目结构

```
godataease/
├── apps/
│   ├── backend-go/      # Go 后端主线
│   └── frontend/        # Vue 3 + TypeScript 前端
├── legacy/
│   ├── backend-java/    # Java 后端只读备份
│   └── sdk/             # Java SDK 模块
├── infra/
│   ├── compose/         # Docker Compose 配置
│   └── scripts/         # 运维脚本
├── openspec/            # OpenSpec 变更管理
└── docs/                # 文档
```

## 新功能使用指南

### 嵌入式 BI

参考文档：[docs/api/embedded-bi.md](./embedded-bi.md)

**快速开始**：

1. 获取嵌入 Token
2. 创建 iframe 或 DIV 容器
3. 设置 postMessage 通信
4. 处理参数传递和事件

### 权限系统

参考文档：[docs/api/permission-system.md](./permission-system.md)

**快速开始**：

1. 创建组织结构
2. 创建角色并分配权限
3. 为用户分配角色和组织
4. 配置数据权限（行级/列级）

## 开发规范

### 后端（Go + Gin）

```go
func (h *Handler) GetMyFeatureList(c *gin.Context) {
    list, err := h.myFeatureService.List(c.Request.Context())
    if err != nil {
        response.Error(c, err)
        return
    }
    response.Success(c, list)
}
```

### 前端（Vue 3 + TypeScript）

```vue
<template>
  <div class="my-feature">
    <el-button @click="handleAdd">添加</el-button>
    <el-table :data="list">
      <el-table-column prop="name" label="名称" />
    </el-table>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getMyFeatureList, createMyFeature } from '@/api/my-feature'

const list = ref([])

const loadList = async () => {
  const { data } = await getMyFeatureList()
  list.value = data
}

const handleAdd = async () => {
  await createMyFeature({ name: '新功能' })
  loadList()
}

onMounted(() => {
  loadList()
})
</script>

<style scoped>
.my-feature {
  padding: 20px;
}
</style>
```

## 测试

### 后端测试

```bash
cd apps/backend-go
make test
```

### 前端测试

```bash
cd apps/frontend
npm run lint          # 代码检查
npm run ts:check       # 类型检查
```

## 常用命令

### Go（后端主线）

```bash
cd apps/backend-go
make build                    # 构建
make run                      # 运行
make test                     # 测试
golangci-lint run            # 代码检查
```

### Legacy Java（仅应急）

```bash
# 仅在应急/对照场景使用
mvn -f legacy/pom.xml clean install -DskipTests
```

### NPM（前端）

```bash
npm install                   # 安装依赖
npm run dev                   # 开发模式
npm run build:base            # 构建前端
npm run lint                  # ESLint 检查
npm run ts:check             # TypeScript 类型检查
```



## 故障排查

### 常见问题

1. **后端启动失败**
   ```bash
   # 检查数据库连接
   # 确认 MySQL 和 Redis 已启动
   # 检查 application.yml 中的数据库配置
   ```

2. **前端编译错误**
   ```bash
   # 清理 node_modules 并重新安装
   rm -rf node_modules package-lock.json
   npm install
   ```

3. **数据库迁移失败**
   ```bash
   # 检查 Flyway 版本历史
   mvn flyway:info
   ```

4. **权限问题**
   ```bash
   # 检查 Token 是否有效
   # 检查用户角色和权限分配
   ```

## 参考文档

- [API 文档](./api/)
- [OpenSpec 规范](../openspec/AGENTS.md)
- [开发指南](../development_guide.md)
- [贡献指南](../CONTRIBUTING.md)
