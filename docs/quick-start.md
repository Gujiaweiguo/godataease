# DataEase 开发者快速入门指南

## 快速开始

### 1. 环境要求

- **Java**: JDK 21+
- **Node.js**: 18+
- **Maven**: 3.8+
- **MySQL**: 8.0+
- **Redis**: 7.0+

### 2. 本地开发

```bash
# 克隆项目
git clone https://github.com/your-org/dataease.git
cd dataease

# 启动基础设施
docker-compose up -d

# 编译后端
cd core/core-backend
mvn clean install -DskipTests

# 编译前端
cd core/core-frontend
npm install

# 启动前端开发服务器
npm run dev
# 访问 http://localhost:5173

# 启动后端
cd core/core-backend
mvn spring-boot:run
# API 访问 http://localhost:8100
```

### 3. 项目结构

```
dataease/
├── core/
│   ├── core-backend/    # Spring Boot 后端
│   └── core-frontend/   # Vue 3 + TypeScript 前端
├── sdk/                 # 共享 SDK 模块
│   ├── common/          # 通用工具类
│   ├── api/             # API 定义
│   └── distributed/      # 分布式组件
├── openspec/            # OpenSpec 变更管理
└── docs/               # 文档
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

### 后端（Java + Spring Boot）

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

    @PostMapping("/create")
    public Result<?> create(@RequestBody MyFeatureDTO dto) {
        return Result.success(myFeatureService.create(dto));
    }
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
cd core/core-backend
mvn test -Dtest=MyFeatureTest
```

### 前端测试

```bash
cd core/core-frontend
npm run lint          # 代码检查
npm run ts:check       # 类型检查
```

## 常用命令

### Maven（后端）

```bash
mvn clean install              # 清理并编译
mvn spring-boot:run            # 运行应用
mvn test                      # 运行测试
mvn flyway:migrate            # 数据库迁移
```

### NPM（前端）

```bash
npm install                   # 安装依赖
npm run dev                   # 开发模式
npm run build:distributed        # 构建分布式版本
npm run lint                  # ESLint 检查
npm run ts:check             # TypeScript 类型检查
```

### Docker

```bash
docker-compose up -d            # 启动所有服务
docker-compose logs -f dataease  # 查看应用日志
docker-compose restart dataease  # 重启应用
```

## 故障排查

### 常见问题

1. **后端启动失败**
   ```bash
   # 检查数据库连接
   docker-compose ps
   curl http://localhost:3306
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
