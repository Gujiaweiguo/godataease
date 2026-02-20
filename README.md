<p align="center"><a href="https://dataease.cn"><img src="https://dataease.oss-cn-hangzhou.aliyuncs.com/img/dataease-logo.png" alt="DataEase" width="300" /></a></p>
<h3 align="center">人人可用的开源 BI 工具</h3>
<p align="center">
  <a href="https://www.gnu.org/licenses/gpl-3.0.html"><img src="https://img.shields.io/github/license/dataease/dataease?color=%231890FF" alt="License: GPL v3"></a>
  <a href="https://app.codacy.com/gh/dataease/dataease?utm_source=github.com&utm_medium=referral&utm_content=dataease/dataease&utm_campaign=Badge_Grade_Dashboard"><img src="https://app.codacy.com/project/badge/Grade/da67574fd82b473992781d1386b937ef" alt="Codacy"></a>
  <a href="https://github.com/Gujiaweiguo/godataease"><img src="https://img.shields.io/github/stars/Gujiaweiguo/godataease?color=%231890FF&style=flat-square" alt="GitHub Stars"></a>
  <a href="https://github.com/Gujiaweiguo/godataease/releases"><img src="https://img.shields.io/github/v/release/Gujiaweiguo/godataease" alt="GitHub release"></a>
</p>
<p align="center">
  <a href="/README.md"><img alt="中文(简体)" src="https://img.shields.io/badge/中文(简体)-d9d9d9"></a>
  <a href="/docs/README.en.md"><img alt="English" src="https://img.shields.io/badge/English-d9d9d9"></a>
  <a href="/docs/README.zh-Hant.md"><img alt="中文(繁體)" src="https://img.shields.io/badge/中文(繁體)-d9d9d9"></a>
  <a href="/docs/README.ja.md"><img alt="日本語" src="https://img.shields.io/badge/日本語-d9d9d9"></a>
  <a href="/docs/README.pt-br.md"><img alt="Português (Brasil)" src="https://img.shields.io/badge/Português (Brasil)-d9d9d9"></a>
  <a href="/docs/README.ar.md"><img alt="العربية" src="https://img.shields.io/badge/العربية-d9d9d9"></a>
  <a href="/docs/README.de.md"><img alt="Deutsch" src="https://img.shields.io/badge/Deutsch-d9d9d9"></a>
  <a href="/docs/README.es.md"><img alt="Español" src="https://img.shields.io/badge/Español-d9d9d9"></a>
  <a href="/docs/README.fr.md"><img alt="français" src="https://img.shields.io/badge/français-d9d9d9"></a>
  <a href="/docs/README.ko.md"><img alt="한국어" src="https://img.shields.io/badge/한국어-d9d9d9"></a>
  <a href="/docs/README.id.md"><img alt="Bahasa Indonesia" src="https://img.shields.io/badge/Bahasa Indonesia-d9d9d9"></a>
  <a href="/docs/README.tr.md"><img alt="Türkçe" src="https://img.shields.io/badge/Türkçe-d9d9d9"></a>
</p>
<p align="center">
  <a href="https://trendshift.io/repositories/1563" target="_blank"><img src="https://trendshift.io/api/badge/repositories/1563" alt="dataease%2Fdataease | Trendshift" style="width: 250px; height: 55px;" width="250" height="55"/></a>
</p>

------------------------------

## 什么是 DataEase？

DataEase 是开源的 BI 工具，帮助用户快速分析数据并洞察业务趋势，从而实现业务的改进与优化。DataEase 支持丰富的数据源连接，能够通过拖拉拽方式快速制作图表，并可以方便的与他人分享。

**DataEase 的优势：**

-   开源开放：零门槛，线上快速获取和安装，按月迭代；
-   简单易用：极易上手，通过鼠标点击和拖拽即可完成分析；
-   全场景支持：多平台安装和多样化嵌入支持；
-   安全分享：支持多种数据分享方式，确保数据安全；
-   AI 加持：无缝集成 [SQLBot](https://github.com/dataease/SQLBot) 实现智能问数。

**DataEase 支持的数据源：**

-   OLTP 数据库： MySQL、Oracle、SQL Server、PostgreSQL、MariaDB、Db2、TiDB、MongoDB-BI 等；
-   OLAP 数据库： ClickHouse、Apache Doris、Apache Impala、StarRocks 等；
-   数据仓库/数据湖： Amazon RedShift 等；
-   数据文件： Excel、CSV 等；
-   API 数据源。

如果您需要向团队介绍 DataEase，可以使用这个 [官方 PPT 材料](https://fit2cloud.com/dataease/download/introduce-dataease_202511.pdf)，或者购买由华东师大和 DataEase 联合出品的图书： [《数据可视化分析与实践》](https://item.jd.com/10207058297099.html)。



## 快速开始（源码安装）

### 环境要求
- Go: 1.21+
- Node.js: 18+
- MySQL: 8.0+
- Redis: 7.0+

### 目录结构

```
godataease/
├── apps/                    # 运行时应用
│   ├── backend-go/         # Go 后端（主线）
│   └── frontend/           # Vue 3 前端
├── legacy/                  # 历史备份（只读）
│   ├── backend-java/       # Java 后端备份
│   └── sdk/                # Java SDK 模块
├── infra/                   # 部署与运维
│   ├── compose/            # Docker Compose 配置
│   ├── assets/             # 运维资产（地图等）
│   └── scripts/            # 部署脚本
├── docs/                    # 文档
└── openspec/               # OpenSpec 规范
```

### 本地开发

```bash
# 克隆项目
git clone https://github.com/Gujiaweiguo/godataease.git
cd godataease

# 编译 Go 后端
cd apps/backend-go
make build

# 编译前端
cd ../frontend
npm install
npm run dev  # 访问 http://localhost:5173

# 启动 Go 后端（需要配置数据库）
cd ../backend-go
make run  # API 访问 http://localhost:8080
```

### 打包构建

```bash
# Go 后端打包
cd apps/backend-go
make build

# 前端构建
cd apps/frontend
npm run build:base
```

### 容器部署（Docker Compose）

使用 Docker Compose 部署完整的开发环境（MySQL + Redis + DataEase）。

#### 部署步骤

1. 构建 Go 后端

```bash
cd apps/backend-go
make build
```

2. 构建前端资源

```bash
cd apps/frontend
npm install
npm run build:base
```

3. 创建外部网络（供多系统共用）

```bash
docker network create my-net
```

4. 启动所有服务

```bash
cd ../../
docker compose -f infra/compose/docker-compose.yml up -d --build
```

服务包括：
- **mysql8**: MySQL 8.0 数据库（端口 3306）
- **redis**: Redis 7.0 缓存（端口 6379）
- **dataease-app**: DataEase 应用（端口 8080）

5. 自定义配置（可选）

在项目根目录创建 `.env`：

```env
DB_HOST=mysql8
DB_PORT=3306
DB_NAME=dataease10
DB_USER=root
DB_PASSWORD=Admin168
```

6. 访问服务

- 应用地址: http://localhost:8080
- API 文档: http://localhost:8080/doc.html

7. 查看日志

```bash
# 查看所有服务日志
docker compose logs -f

# 查看特定服务日志
docker compose logs -f dataease-app
docker compose logs -f mysql8
docker compose logs -f redis
```

8. 停止服务

```bash
# 停止并删除容器
docker compose down

# 停止并删除容器和数据卷（⚠️ 会清除数据）
docker compose down -v
```

#### 复用现有容器

如果系统中已有 MySQL 和 Redis 容器，可以只构建 dataease-app 服务：

```bash
# 修改 docker-compose.yml，注释掉 mysql8 和 redis 服务
docker compose up -d --build dataease-app
```

确保 application-standalone.yml 配置正确指向现有容器：
- MySQL: `mysql8:3306`
- Redis: `redis:6379`

更多开发指南请参考 [development_guide.md](./development_guide.md) 和 [AGENTS.md](./AGENTS.md)。


## UI 展示

<table style="border-collapse: collapse; border: 1px solid black;">
  <tr>
    <td style="padding: 5px;background-color:#fff;"><img src= "/docs/assets/ui/workbench.png" alt="DataEase 工作台"   /></td>
    <td style="padding: 5px;background-color:#fff;"><img src= "/docs/assets/ui/dashboard.png" alt="DataEase 仪表板"   /></td>
  </tr>

  <tr>
    <td style="padding: 5px;background-color:#fff;"><img src= "/docs/assets/ui/datasource.png" alt="DataEase 数据源"   /></td>
    <td style="padding: 5px;background-color:#fff;"><img src= "/docs/assets/ui/template.png" alt="DataEase 模板中心"   /></td>
  </tr>
</table>

## 技术栈

-   前端：[Vue.js](https://vuejs.org/)、[Element](https://element.eleme.cn/)
-   图库：[AntV](https://antv.vision/zh)
-   后端（主线）：[Go](https://go.dev/) + [Gin](https://gin-gonic.com/)
-   后端（历史只读备份）：[Spring Boot](https://spring.io/projects/spring-boot)
-   数据库：[MySQL](https://www.mysql.com/)
-   数据处理：[Apache Calcite](https://github.com/apache/calcite/)、[Apache SeaTunnel](https://github.com/apache/seatunnel)
-   基础设施：[Docker](https://www.docker.com/)

## 飞致云的其他明星项目

- [1Panel](https://github.com/1panel-dev/1panel/) - 现代化、开源的 Linux 服务器运维管理面板
- [MaxKB](https://github.com/1panel-dev/MaxKB/) - 基于 LLM 大语言模型的开源知识库问答系统
- [JumpServer](https://github.com/jumpserver/jumpserver/) - 广受欢迎的开源堡垒机
- [Cordys CRM](https://github.com/1Panel-dev/CordysCRM) - 新一代的开源 AI CRM 系统
- [Halo](https://github.com/halo-dev/halo/) - 强大易用的开源建站工具
- [MeterSphere](https://github.com/metersphere/metersphere/) - 新一代的开源持续测试工具

## License

Copyright (c) 2014-2026 [FIT2CLOUD 飞致云](https://fit2cloud.com/), All rights reserved.

Licensed under The GNU General Public License version 3 (GPLv3)  (the "License"); you may not use this file except in compliance with the License. You may obtain a copy of the License at

<https://www.gnu.org/licenses/gpl-3.0.html>

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.
