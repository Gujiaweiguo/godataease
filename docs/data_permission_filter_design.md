# 数据权限过滤功能详细设计文档

## 1. 功能概述

### 1.1 目标
实现基于组织的自动数据权限过滤功能，使系统能够根据用户的组织归属自动过滤其可见的数据，无需为每个用户手动配置权限规则。

### 1.2 背景
当前DataEase系统已具备完善的行级权限系统，支持：
- 基于用户、角色的显式权限配置
- 树形结构的过滤条件定义
- 系统变量支持（如${sysParams.userId}）
- 白名单机制

但缺少自动化的组织级数据过滤，管理员需要为每个用户或角色手动配置数据访问规则。

### 1.3 解决方案
在现有权限系统基础上，增加**自动组织过滤**功能：
- 为每个数据集配置组织字段映射
- 系统根据用户的组织ID自动生成过滤条件
- 自动过滤与手动权限规则叠加生效

## 2. 系统架构

### 2.1 现有系统分析

#### 2.1.1 权限查询流程
```
用户请求数据
  ↓
DatasetDataManage.previewDataWithLimit()
  ↓
PermissionManage.getRowPermissionsTree()
  ↓
获取用户、角色、组织的显式权限规则
  ↓
WhereTree2Str.transFilterTrees()
  ↓
将权限树转换为WHERE子句
  ↓
执行查询并返回过滤后的数据
```

#### 2.1.2 核心类说明

| 类名 | 职责 |
|------|------|
| `PermissionManage` | 权限管理核心类，负责获取和合并各种权限规则 |
| `WhereTree2Str` | 将权限树转换为SQL WHERE子句 |
| `DataSetRowPermissionsTreeDTO` | 行权限规则的数据传输对象 |
| `DatasetRowPermissionsTreeObj` | 权限树结构，包含逻辑操作符（AND/OR）和条件项 |
| `DatasetRowPermissionsTreeItem` | 单个过滤条件项，包含字段、操作符、值 |

#### 2.1.3 现有系统支持的功能
1. **用户级权限**: 为特定用户配置数据过滤规则
2. **角色级权限**: 为角色配置数据过滤规则，所有拥有该角色的用户自动继承
3. **系统变量**: 支持动态替换如 `${sysParams.userId}`、`${sysParams.userName}` 等系统变量
4. **白名单机制**: 可以排除某些用户/角色不受权限规则限制
5. **导出数据权限**: 支持为导出操作单独配置权限

### 2.2 新增功能架构

#### 2.2.1 自动组织过滤流程
```
用户请求数据
  ↓
DatasetDataManage.previewDataWithLimit()
  ↓
PermissionManage.getRowPermissionsTree()
  ↓
获取显式权限规则（用户、角色、组织）
  ↓
【新增】获取自动组织过滤器
  ↓
  - 检查数据集是否启用自动过滤
  ↓
  - 获取用户的组织ID
  ↓
  - 根据配置的组织字段生成过滤条件
  ↓
WhereTree2Str.transFilterTrees()
  ↓
合并显式规则和自动过滤器（AND逻辑）
  ↓
执行查询并返回过滤后的数据
```

#### 2.2.2 核心概念

##### 概念1：自动组织过滤器 (AutoOrgFilter)
- **定义**: 由系统根据用户组织ID自动生成的数据过滤条件
- **生成规则**:
  ```
  数据字段名 = 用户所属组织ID
  ```
- **特点**:
  - 无需管理员手动配置
  - 基于用户的组织归属动态生成
  - 与手动权限规则叠加生效（AND逻辑）

##### 概念2：数据集组织字段映射 (DatasetOrgFieldMapping)
- **定义**: 配置数据集中哪个字段对应组织ID
- **存储方式**: 在数据集配置表中新增字段
- **配置格式**: JSON
  ```json
  {
    "orgFieldId": 12345,  // 组织字段ID
    "enableAutoFilter": true,  // 是否启用自动过滤
    "orgFilterType": "current_org"  // 过滤类型：当前组织/当前及子组织
  }
  ```

##### 概念3：过滤类型 (FilterType)
- **current_org**: 只查看当前组织的数据
- **current_and_sub_orgs**: 查看当前组织及所有子组织的数据

### 2.3 数据库设计

#### 2.3.1 修改现有表

**core_dataset_group 表新增字段**:
```sql
ALTER TABLE core_dataset_group ADD COLUMN org_permission_config TEXT COMMENT '组织权限配置(JSON)';

-- 示例数据
-- {
--   "orgFieldId": 12345,
--   "enableAutoFilter": true,
--   "orgFilterType": "current_org"
-- }
```

#### 2.3.2 数据结构说明

| 字段名 | 类型 | 说明 |
|--------|------|------|
| org_permission_config | TEXT | 组织权限配置JSON数据 |

**org_permission_config JSON结构**:
```typescript
interface OrgPermissionConfig {
  // 组织字段ID（数据集中的哪个字段代表组织）
  orgFieldId: number;

  // 是否启用自动组织过滤
  enableAutoFilter: boolean;

  // 过滤类型
  // "current_org": 只看当前组织
  // "current_and_sub_orgs": 看当前及子组织
  orgFilterType: 'current_org' | 'current_and_sub_orgs';
}
```

## 3. 详细实现设计

### 3.1 核心类设计

#### 3.1.1 OrgPermissionConfig (DTO)
```java
package io.dataease.dataset.dto;

import lombok.Data;
import io.swagger.v3.oas.annotations.media.Schema;

@Data
@Schema(description = "组织权限配置")
public class OrgPermissionConfig {

    @Schema(description = "组织字段ID")
    private Long orgFieldId;

    @Schema(description = "是否启用自动过滤")
    private Boolean enableAutoFilter;

    @Schema(description = "过滤类型")
    private String orgFilterType;
}
```

#### 3.1.2 AutoOrgFilterService (新增服务类)
```java
package io.dataease.dataset.manage;

import io.dataease.api.permissions.dataset.dto.DataSetRowPermissionsTreeDTO;
import io.dataease.dataset.dto.OrgPermissionConfig;
import io.dataease.auth.bo.TokenUserBO;

/**
 * 自动组织过滤服务
 * 负责根据用户组织信息生成自动过滤器
 */
@Component
public class AutoOrgFilterService {

    @Resource
    private DatasetGroupManage datasetGroupManage;

    @Resource
    private DatasetTableFieldManage datasetTableFieldManage;

    @Resource
    private SysOrganizationManage sysOrganizationManage;

    /**
     * 获取自动组织过滤器
     *
     * @param datasetGroupId 数据集ID
     * @param userId 用户ID
     * @return 自动组织过滤树，如果不需要过滤返回null
     */
    public DataSetRowPermissionsTreeDTO getAutoOrgFilter(Long datasetGroupId, Long userId) {
        // 实现见 3.2
    }

    /**
     * 检查数据集是否启用自动组织过滤
     *
     * @param datasetGroupId 数据集ID
     * @return 是否启用
     */
    public boolean isAutoFilterEnabled(Long datasetGroupId) {
        // 实现见 3.2
    }

    /**
     * 生成组织过滤树
     *
     * @param orgPermissionConfig 组织权限配置
     * @param userOrgId 用户组织ID
     * @return 权限树
     */
    private DataSetRowPermissionsTreeDTO generateOrgFilterTree(
        OrgPermissionConfig orgPermissionConfig,
        Long userOrgId
    ) {
        // 实现见 3.2
    }

    /**
     * 获取用户可访问的组织ID列表
     *
     * @param userOrgId 用户组织ID
     * @param filterType 过滤类型
     * @return 组织ID列表
     */
    private List<Long> getAccessibleOrgIds(Long userOrgId, String filterType) {
        // 实现见 3.2
    }
}
```

### 3.2 核心方法实现

#### 3.2.1 getAutoOrgFilter - 获取自动组织过滤器

```java
public DataSetRowPermissionsTreeDTO getAutoOrgFilter(Long datasetGroupId, Long userId) {
    // 1. 获取数据集的组织权限配置
    OrgPermissionConfig config = getOrgPermissionConfig(datasetGroupId);
    if (config == null || !config.getEnableAutoFilter()) {
        return null;  // 未启用自动过滤
    }

    // 2. 获取用户的组织信息
    TokenUserBO user = AuthUtils.getUser();
    if (user == null || user.getOrganizationId() == null) {
        return null;  // 用户无组织信息
    }
    Long userOrgId = user.getOrganizationId();

    // 3. 生成组织过滤树
    return generateOrgFilterTree(config, userOrgId);
}

private OrgPermissionConfig getOrgPermissionConfig(Long datasetGroupId) {
    // 从数据库获取数据集配置
    DatasetGroupInfoDTO datasetGroup = datasetGroupManage.getDatasetGroupInfoDTO(datasetGroupId, null);
    if (datasetGroup == null || datasetGroup.getOrgPermissionConfig() == null) {
        return null;
    }

    // 解析JSON配置
    try {
        return JsonUtil.parseObject(
            datasetGroup.getOrgPermissionConfig(),
            OrgPermissionConfig.class
        );
    } catch (Exception e) {
        logger.error("Parse org permission config failed", e);
        return null;
    }
}
```

#### 3.2.2 generateOrgFilterTree - 生成组织过滤树

```java
private DataSetRowPermissionsTreeDTO generateOrgFilterTree(
    OrgPermissionConfig config,
    Long userOrgId
) {
    // 1. 获取组织字段
    DatasetTableFieldDTO orgField = datasetTableFieldManage.selectById(config.getOrgFieldId());
    if (orgField == null) {
        logger.warn("Org field not found: {}", config.getOrgFieldId());
        return null;
    }

    // 2. 获取用户可访问的组织ID列表
    List<Long> accessibleOrgIds = getAccessibleOrgIds(userOrgId, config.getOrgFilterType());
    if (CollectionUtils.isEmpty(accessibleOrgIds)) {
        return null;
    }

    // 3. 构建权限树
    DataSetRowPermissionsTreeDTO filter = new DataSetRowPermissionsTreeDTO();
    filter.setDatasetId(null);  // 自动过滤不绑定特定数据集
    filter.setAuthTargetType("auto_org");  // 标记为自动组织过滤
    filter.setEnable(true);

    // 4. 构建树结构
    DatasetRowPermissionsTreeObj tree = new DatasetRowPermissionsTreeObj();
    tree.setLogic("and");  // 单条件，逻辑无影响

    // 5. 创建过滤条件项
    DatasetRowPermissionsTreeItem item = new DatasetRowPermissionsTreeItem();
    item.setType("item");
    item.setFieldId(orgField.getId());
    item.setField(orgField);
    item.setFilterType("enum");  // 使用IN操作符
    item.setTerm("in");
    item.setEnumValue(accessibleOrgIds.stream().map(String::valueOf).collect(Collectors.toList()));

    tree.setItems(Collections.singletonList(item));
    filter.setTree(tree);

    return filter;
}
```

#### 3.2.3 getAccessibleOrgIds - 获取可访问组织ID

```java
private List<Long> getAccessibleOrgIds(Long userOrgId, String filterType) {
    if ("current_org".equals(filterType)) {
        // 只查看当前组织
        return Collections.singletonList(userOrgId);
    } else if ("current_and_sub_orgs".equals(filterType)) {
        // 查看当前及所有子组织
        List<SysOrganization> allOrgs = sysOrganizationManage.getAllOrganizations();
        return getSubOrgIds(allOrgs, userOrgId);
    }
    return Collections.singletonList(userOrgId);
}

/**
 * 递归获取所有子组织ID
 */
private List<Long> getSubOrgIds(List<SysOrganization> allOrgs, Long parentOrgId) {
    List<Long> result = new ArrayList<>();
    result.add(parentOrgId);

    for (SysOrganization org : allOrgs) {
        if (parentOrgId.equals(org.getParentId())) {
            result.addAll(getSubOrgIds(allOrgs, org.getId()));
        }
    }

    return result;
}
```

### 3.3 PermissionManage 集成

修改 `PermissionManage.getRowPermissionsTree()` 方法：

```java
public List<DataSetRowPermissionsTreeDTO> getRowPermissionsTree(Long datasetId, Long user) {
    // 1. 获取显式权限规则（现有逻辑）
    List<DataSetRowPermissionsTreeDTO> records = rowPermissionsTree(datasetId, user);

    // 2. 【新增】获取自动组织过滤器
    DataSetRowPermissionsTreeDTO autoOrgFilter = autoOrgFilterService.getAutoOrgFilter(datasetId, user);

    // 3. 合并权限规则
    if (autoOrgFilter != null) {
        if (CollectionUtils.isEmpty(records)) {
            records = new ArrayList<>();
        }
        records.add(autoOrgFilter);
    }

    // 4. 构建权限树中的field，如果field不存在，置为null
    if (ObjectUtils.isNotEmpty(datasetId)) {
        for (DataSetRowPermissionsTreeDTO record : records) {
            getField(record.getTree());
        }
    }
    return records;
}
```

**关键点**：
- 自动过滤器与显式规则在同一列表中
- WhereTree2Str.transFilterTrees() 会将它们用 OR 连接
- 这样用户必须同时满足：自动过滤规则 AND 显式权限规则

### 3.4 WhereTree2Str 修改

**无需修改！** 因为：
- 自动过滤器生成的结构与显式权限规则完全一致
- 都是 `DataSetRowPermissionsTreeDTO` 对象
- WhereTree2Str 不区分来源，统一处理

### 3.5 API设计

#### 3.5.1 配置组织权限

**接口**: `POST /api/dataset/{id}/org-permission/config`

**请求体**:
```json
{
  "orgFieldId": 12345,
  "enableAutoFilter": true,
  "orgFilterType": "current_org"
}
```

**响应**:
```json
{
  "code": 200,
  "msg": "success",
  "data": null
}
```

#### 3.5.2 获取组织权限配置

**接口**: `GET /api/dataset/{id}/org-permission/config`

**响应**:
```json
{
  "code": 200,
  "msg": "success",
  "data": {
    "orgFieldId": 12345,
    "enableAutoFilter": true,
    "orgFilterType": "current_org"
  }
}
```

#### 3.5.3 启用/禁用自动过滤

**接口**: `PUT /api/dataset/{id}/org-permission/enable`

**请求体**:
```json
{
  "enableAutoFilter": true
}
```

### 3.6 前端设计

#### 3.6.1 数据集配置页面

在数据集编辑/配置页面新增"组织权限"标签页：

```vue
<template>
  <el-tabs v-model="activeTab">
    <el-tab-pane label="字段配置" name="fields">...</el-tab-pane>
    <el-tab-pane label="行权限" name="row-permissions">...</el-tab-pane>
    <el-tab-pane label="列权限" name="column-permissions">...</el-tab-pane>
    <!-- 新增 -->
    <el-tab-pane label="组织权限" name="org-permissions">
      <OrgPermissionConfig :datasetId="datasetId" />
    </el-tab-pane>
  </el-tabs>
</template>

<script setup lang="ts">
import OrgPermissionConfig from './components/OrgPermissionConfig.vue'
</script>
```

#### 3.6.2 OrgPermissionConfig 组件

```vue
<template>
  <div class="org-permission-config">
    <el-form :model="config" label-width="120px">
      <el-form-item label="启用自动过滤">
        <el-switch v-model="config.enableAutoFilter" />
      </el-form-item>

      <template v-if="config.enableAutoFilter">
        <el-form-item label="组织字段" required>
          <el-select v-model="config.orgFieldId" placeholder="选择组织字段">
            <el-option
              v-for="field in orgFields"
              :key="field.id"
              :label="field.name"
              :value="field.id"
            />
          </el-select>
        </el-form-item>

        <el-form-item label="过滤类型">
          <el-radio-group v-model="config.orgFilterType">
            <el-radio label="current_org">只查看当前组织</el-radio>
            <el-radio label="current_and_sub_orgs">查看当前及子组织</el-radio>
          </el-radio-group>
        </el-form-item>
      </template>

      <el-form-item>
        <el-button type="primary" @click="save">保存</el-button>
      </el-form-item>
    </el-form>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getOrgPermissionConfig, saveOrgPermissionConfig } from '@/api/dataset'

const props = defineProps<{ datasetId: number }>()

const config = ref({
  orgFieldId: null,
  enableAutoFilter: false,
  orgFilterType: 'current_org'
})

const orgFields = ref([])

onMounted(async () => {
  // 加载配置
  const result = await getOrgPermissionConfig(props.datasetId)
  if (result) {
    config.value = result
  }

  // 加载字段列表（过滤出适合作为组织字段的字段）
  orgFields.value = await getSuitableFields(props.datasetId)
})

const save = async () => {
  await saveOrgPermissionConfig(props.datasetId, config.value)
  ElMessage.success('保存成功')
}
</script>
```

## 4. 使用场景

### 4.1 场景1：多租户SaaS平台

**需求**:
- 不同租户（组织）的数据完全隔离
- 每个租户用户只能看到自己租户的数据

**配置**:
1. 在数据集"用户订单"表中，配置`org_id`字段为组织字段
2. 启用自动过滤，过滤类型为"current_org"

**效果**:
- 租户A的用户（org_id=1）只能看到`org_id=1`的订单
- 租户B的用户（org_id=2）只能看到`org_id=2`的订单

### 4.2 场景2：多级组织管理

**需求**:
- 总部可以看到所有分支的数据
- 分支可以看到本分支及下属部门的数据
- 部门只能看到本部门的数据

**配置**:
1. 在数据集"员工信息"表中，配置`dept_id`字段为组织字段
2. 启用自动过滤，过滤类型为"current_and_sub_orgs"

**效果**:
- 总部员工（org_id=1）可以看到所有组织的数据（递归包含子组织）
- 华东分公司（org_id=2，父组织为1）可以看到华东及下属的数据
- 杭州办事处（org_id=3，父组织为2）只能看到杭州的数据

### 4.3 场景3：混合权限控制

**需求**:
- 用户自动看到本组织数据
- 部分用户可以额外查看其他部门的数据

**配置**:
1. 配置自动组织过滤（按场景1）
2. 为特殊用户配置额外的行权限规则

**效果**:
- 普通用户：自动过滤 + 无额外规则 = 只看本组织
- 特殊用户：自动过滤 + 额外规则 = 本组织 + 指定组织（OR逻辑）

## 5. 实现步骤

### 阶段1：数据库和基础模型（1天）
- [ ] 修改core_dataset_group表，添加org_permission_config字段
- [ ] 创建OrgPermissionConfig DTO类
- [ ] 编写Flyway迁移脚本

### 阶段2：核心服务实现（2天）
- [ ] 创建AutoOrgFilterService服务类
- [ ] 实现getAutoOrgFilter方法
- [ ] 实现generateOrgFilterTree方法
- [ ] 实现getAccessibleOrgIds方法
- [ ] 单元测试

### 阶段3：权限管理集成（1天）
- [ ] 修改PermissionManage.getRowPermissionsTree()
- [ ] 注入AutoOrgFilterService依赖
- [ ] 集成测试

### 阶段4：API开发（1天）
- [ ] DatasetGroupManage添加配置方法
- [ ] DatasetGroupServer添加API接口
- [ ] API文档更新

### 阶段5：前端开发（2天）
- [ ] 创建OrgPermissionConfig组件
- [ ] 集成到数据集配置页面
- [ ] 字段选择器实现
- [ ] 前端测试

### 阶段6：联调和测试（2天）
- [ ] 端到端测试
- [ ] 性能测试
- [ ] 边界条件测试
- [ ] 文档更新

**总计**: 约9个工作日

## 6. 测试计划

### 6.1 单元测试

```java
@SpringBootTest
class AutoOrgFilterServiceTest {

    @Resource
    private AutoOrgFilterService autoOrgFilterService;

    @Test
    void testGetAutoOrgFilter_WhenNotEnabled_ReturnsNull() {
        // 测试：未启用自动过滤时返回null
    }

    @Test
    void testGetAutoOrgFilter_WhenEnabled_ReturnsFilter() {
        // 测试：启用自动过滤时返回正确的过滤器
    }

    @Test
    void testGetAccessibleOrgIds_CurrentOrg() {
        // 测试：current_org类型只返回当前组织
    }

    @Test
    void testGetAccessibleOrgIds_CurrentAndSubOrgs() {
        // 测试：current_and_sub_orgs类型递归返回子组织
    }

    @Test
    void testGenerateOrgFilterTree_FieldNotFound() {
        // 测试：配置的字段不存在时的处理
    }
}
```

### 6.2 集成测试

```java
@SpringBootTest
class PermissionManageIntegrationTest {

    @Resource
    private PermissionManage permissionManage;

    @Test
    void testGetRowPermissionsTree_WithAutoOrgFilter() {
        // 测试：自动过滤器与显式规则正确合并
    }

    @Test
    void testGetData_WithOrgFilter() {
        // 测试：查询数据时正确应用组织过滤
    }
}
```

### 6.3 端到端测试

1. **配置流程测试**
   - 创建数据集
   - 配置组织字段
   - 启用自动过滤
   - 保存配置

2. **权限过滤测试**
   - 用户A（org_id=1）查询数据，验证只看到org_id=1的数据
   - 用户B（org_id=2）查询数据，验证只看到org_id=2的数据
   - 系统管理员查询数据，验证可以看到所有数据

3. **混合权限测试**
   - 配置自动过滤
   - 为用户添加额外行权限规则
   - 验证用户可以看到：本组织数据 OR 额外组织数据

4. **边界条件测试**
   - 用户无组织信息
   - 组织字段配置错误
   - 组织树循环依赖
   - 禁用自动过滤后数据正常显示

## 7. 性能优化

### 7.1 缓存策略

```java
@Service
public class AutoOrgFilterService {

    @Cacheable(value = "orgPermissionConfig", key = "#datasetGroupId")
    public OrgPermissionConfig getOrgPermissionConfig(Long datasetGroupId) {
        // 缓存组织权限配置
    }

    @Cacheable(value = "allOrganizations", unless = "#result == null")
    public List<SysOrganization> getAllOrganizations() {
        // 缓存所有组织信息
    }

    @Cacheable(value = "subOrgIds", key = "#orgId + '_' + #filterType")
    public List<Long> getSubOrgIds(Long orgId, String filterType) {
        // 缓存子组织ID列表
    }
}
```

### 7.2 批量查询优化

- 一次加载所有组织到内存，避免递归查询数据库
- 使用Map存储组织关系，提高查找效率

```java
private List<Long> getSubOrgIds(Long parentOrgId, String filterType) {
    if (!"current_and_sub_orgs".equals(filterType)) {
        return Collections.singletonList(parentOrgId);
    }

    // 从缓存或内存加载所有组织
    List<SysOrganization> allOrgs = getAllOrganizations();

    // 构建父子关系Map
    Map<Long, List<Long>> childrenMap = new HashMap<>();
    for (SysOrganization org : allOrgs) {
        childrenMap.computeIfAbsent(org.getParentId(), k -> new ArrayList<>())
                  .add(org.getId());
    }

    // BFS遍历获取所有子组织
    List<Long> result = new ArrayList<>();
    Queue<Long> queue = new LinkedList<>();
    queue.add(parentOrgId);

    while (!queue.isEmpty()) {
        Long current = queue.poll();
        result.add(current);

        List<Long> children = childrenMap.get(current);
        if (children != null) {
            queue.addAll(children);
        }
    }

    return result;
}
```

## 8. 安全考虑

### 8.1 权限验证

- 只有数据集所有者或管理员可以配置组织权限
- 自动过滤器对系统管理员（sysAdmin）不生效
- 配置信息加密存储（可选）

### 8.2 SQL注入防护

- 使用参数化查询，不拼接SQL
- 组织ID值进行严格校验

### 8.3 数据泄露防护

- 确保自动过滤与其他权限规则AND组合
- 验证权限规则不会被绕过

## 9. 扩展性考虑

### 9.1 支持更复杂的过滤类型

未来可以支持：
- **自定义SQL**: 允许管理员编写复杂的过滤逻辑
- **字段映射表**: 支持多字段联合过滤
- **继承规则**: 支持权限规则的继承和覆盖

### 9.2 支持多维度过滤

当前只支持组织维度，未来可扩展：
- **部门过滤**: 基于部门ID过滤
- **地域过滤**: 基于地区字段过滤
- **自定义标签过滤**: 基于业务标签过滤

### 9.3 审计日志

记录自动过滤的使用情况：
- 哪些数据集启用了自动过滤
- 每次查询应用的过滤规则
- 异常访问尝试

## 10. 注意事项

### 10.1 兼容性

- 现有数据集默认不启用自动过滤
- 已有的显式权限规则不受影响
- 迁移到新版本无需额外操作

### 10.2 性能影响

- 启用自动过滤会增加查询生成的时间（约5-10ms）
- 通过缓存可以显著降低影响
- 建议在高并发场景进行压力测试

### 10.3 用户体验

- 在数据集配置页面明确显示自动过滤状态
- 提供配置预览功能
- 在查询结果中显示应用的过滤规则（调试模式）

## 11. 参考资料

- Apache Calcite SQL解析: https://calcite.apache.org/
- DataEase权限系统文档
- MyBatis Plus文档: https://baomidou.com/
- Spring Boot缓存文档: https://docs.spring.io/

---

**文档版本**: v1.0
**创建日期**: 2025-01-24
**作者**: Sisyphus
**审核人**: 待定
