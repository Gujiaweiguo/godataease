package io.dataease.dataset.manage;

import io.dataease.api.permissions.dataset.dto.DataSetRowPermissionsTreeDTO;
import io.dataease.dataset.dao.auto.entity.CoreDatasetGroup;
import io.dataease.dataset.dao.auto.mapper.CoreDatasetGroupMapper;
import io.dataease.dataset.dto.OrgPermissionConfig;
import io.dataease.extensions.datasource.dto.DatasetTableFieldDTO;
import io.dataease.utils.JsonUtil;
import jakarta.annotation.Resource;
import org.apache.commons.collections4.CollectionUtils;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Component;

import java.util.*;

@Component
public class AutoOrgFilterService {

    private static final Logger logger = LoggerFactory.getLogger(AutoOrgFilterService.class);

    @Resource
    private CoreDatasetGroupMapper coreDatasetGroupMapper;

    @Resource
    private DatasetTableFieldManage datasetTableFieldManage;

    public DataSetRowPermissionsTreeDTO getAutoOrgFilter(Long datasetGroupId, Long userId) {
        // Temporarily disable auto-filter check to avoid org_permission_config table issues
        return null;
    }

    public boolean isAutoFilterEnabled(Long datasetGroupId) {
        // Temporarily disable auto-filter check to avoid org_permission_config table issues
        return false;
    }

    private OrgPermissionConfig getOrgPermissionConfig(Long datasetGroupId) {
        // Temporarily disable to avoid org_permission_config table issues
        return null;
    }

    public void saveOrgPermissionConfig(Long datasetGroupId, OrgPermissionConfig config) {
        // Temporarily disable to avoid org_permission_config table issues
        CoreDatasetGroup datasetGroup = coreDatasetGroupMapper.selectById(datasetGroupId);
        if (datasetGroup == null) {
            throw new RuntimeException("Dataset not found: " + datasetGroupId);
        }

        coreDatasetGroupMapper.updateById(datasetGroup);
    }

    private DataSetRowPermissionsTreeDTO generateOrgFilterTree(
        OrgPermissionConfig config,
        Long userOrgId,
        Long datasetGroupId
    ) {
        return null;
    }

    private List<Long> getAccessibleOrgIds(Long userOrgId, String filterType) {
        return Collections.singletonList(userOrgId);
    }

    public OrgPermissionConfig getOrgPermissionConfigForApi(Long datasetGroupId) {
        return getOrgPermissionConfig(datasetGroupId);
    }
}
