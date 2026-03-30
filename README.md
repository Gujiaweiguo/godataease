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
- Go: 1.24+
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

### 本地开发（推荐：本地前后端 + Docker Redis + 共用 MySQL）

```bash
# 克隆项目
git clone https://github.com/Gujiaweiguo/godataease.git
cd godataease

# 1. 启动 Redis 容器（加入 my-net 网络，复用已有的 mysql8 容器）
docker compose -f infra/compose/docker-compose.dev.yml --env-file infra/compose/.env.dev up -d

# 2. 启动 Go 后端
cd apps/backend-go
DATABASE_HOST=127.0.0.1 DATABASE_PORT=3306 DATABASE_NAME=dataease_dev DATABASE_USER=root DATABASE_PASSWORD=Admin168 REDIS_HOST=127.0.0.1 REDIS_PORT=16379 make run-local

# 3. 启动前端开发服务器
cd ../frontend
npm install
npm run dev
```

本地开发默认端口：
- 前端：`http://localhost:5173`
- 后端：`http://localhost:8080`
- Redis（Docker）：`127.0.0.1:16379`
- MySQL（共用 mysql8 容器）：`127.0.0.1:3306`

配置文件：
- 开发环境: `infra/compose/.env.dev`（可手工修改）
- 正式环境: `infra/compose/.env.prod`（可手工修改）

默认登录凭据：
- 用户名: `admin`
- 密码: `admin123`

### 打包构建

```bash
# Go 后端打包
cd apps/backend-go
make build

# 前端构建
cd apps/frontend
npm run build:base
```

### 正式环境部署（全容器化）

所有服务（前后端、MySQL、Redis、Nginx）全部容器化。

```bash
# 构建
./scripts/dev.sh build

# 启动正式环境
./scripts/dev.sh prod-start

# 访问
# 应用地址: http://localhost:8080
# API 文档: http://localhost:8080/doc.html
```

配置文件 `infra/compose/.env.prod` 可手工修改，修改后重启生效。

更多开发指南请参考 [development_guide.md](./development_guide.md)、[AGENTS.md](./AGENTS.md) 和 [开发验证清单 Runbook](./docs/runbook/dev-validation-checklist.md)。


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
