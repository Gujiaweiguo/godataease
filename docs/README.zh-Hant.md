<p align="center"><a href="https://dataease.io"><img src="https://dataease.oss-cn-hangzhou.aliyuncs.com/img/dataease-logo.png" alt="DataEase" width="300" /></a></p>
<h3 align="center">簡單易用的開源 BI 工具</h3>
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

------------------------------

## 什麼是 DataEase？

DataEase 是開源的 BI 工具，幫助用戶快速分析數據並洞察業務趨勢，從而實現業務的改進與優化。DataEase 支持豐富的數據源連接，能夠通過拖拉拽方式快速製作圖表，並可以方便地與他人分享。

**DataEase 的優勢：**

-   開源開放：零門檻，線上快速獲取和安裝，按月迭代；
-   簡單易用：極易上手，通過鼠標點擊和拖拉即可完成分析；
-   全場景支持：多平台安裝和多樣化嵌入支持；
-   安全分享：支持多種數據分享方式，確保數據安全。

**DataEase 支持的數據源：**

-   OLTP 數據庫： MySQL、Oracle、SQL Server、PostgreSQL、MariaDB、Db2、TiDB、MongoDB-BI 等；
-   OLAP 數據庫： ClickHouse、Apache Doris、Apache Impala、StarRocks 等；
-   數據倉庫/數據湖： Amazon RedShift 等；
-   數據文件： Excel、CSV 等；
-   API 數據源。

如果您需要向團隊介紹 DataEase，可以使用這個 [官方 PPT 材料](https://fit2cloud.com/dataease/download/introduce-dataease_202411.pdf)。

## 快速開始（原始碼安裝）

### 環境需求
- Java: JDK 21+
- Node.js: 18+
- Maven: 3.8+
- MySQL: 8.0+
- Redis: 7.0+

### 本機開發

```bash
# 取得專案
git clone https://github.com/Gujiaweiguo/godataease.git
cd dataease

# 編譯後端
cd core/core-backend
mvn clean install -DskipTests

# 編譯前端
cd ../core-frontend
npm install
npm run dev  # 造訪 http://localhost:5173

# 啟動後端（需先設定資料庫）
cd ../core-backend
mvn spring-boot:run  # API 造訪 http://localhost:8100
```

### 打包建置

```bash
# 後端打包
mvn clean package -DskipTests

# 前端建置
cd core/core-frontend
npm run build:base
```

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

## 技術棧

-   前端：[Vue.js](https://vuejs.org/)、[Element](https://element.eleme.cn/)
-   圖庫：[AntV](https://antv.vision/zh)
-   後端：[Spring Boot](https://spring.io/projects/spring-boot)
-   資料庫：[MySQL](https://www.mysql.com/)
-   資料處理：[Apache Calcite](https://github.com/apache/calcite/)、[Apache SeaTunnel](https://github.com/apache/seatunnel)
-   基礎設施：[Docker](https://www.docker.com/)

## 飛致雲的其他明星專案

- [1Panel](https://github.com/1panel-dev/1panel/) - 現代化、開源的 Linux 伺服器運維管理面板
- [MaxKB](https://github.com/1panel-dev/MaxKB/) - 基於 LLM 大語言模型的開源知識庫問答系統
- [JumpServer](https://github.com/jumpserver/jumpserver/) - 廣受歡迎的開源堡壘機
- [Halo](https://github.com/halo-dev/halo/) - 強大易用的開源建站工具
- [MeterSphere](https://github.com/metersphere/metersphere/) - 新一代的開源持續測試工具

## License

Copyright (c) 2014-2026 [FIT2CLOUD 飞致云](https://fit2cloud.com/), All rights reserved.

Licensed under The GNU General Public License version 3 (GPLv3)  (the "License"); you may not use this file except in compliance with the License. You may obtain a copy of the License at

<https://www.gnu.org/licenses/gpl-3.0.html>

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.
