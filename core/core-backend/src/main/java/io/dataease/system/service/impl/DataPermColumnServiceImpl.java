package io.dataease.system.service.impl;

import io.dataease.system.dao.auto.mapper.DataPermColumnMapper;
import io.dataease.system.entity.DataPermColumn;
import io.dataease.system.service.IDataPermColumnService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.List;

@Service
public class DataPermColumnServiceImpl implements IDataPermColumnService {

    @Autowired
    private DataPermColumnMapper mapper;

    @Override
    @Transactional(rollbackFor = Exception.class)
    public Long create(DataPermColumn dataPermColumn) {
        dataPermColumn.setCreateTime(java.time.LocalDateTime.now());
        dataPermColumn.setUpdateTime(java.time.LocalDateTime.now());
        dataPermColumn.setStatus(1);
        mapper.insert(dataPermColumn);
        return dataPermColumn.getId();
    }

    @Override
    @Transactional(rollbackFor = Exception.class)
    public void update(DataPermColumn dataPermColumn) {
        if (dataPermColumn.getId() == null) {
            return;
        }
        dataPermColumn.setUpdateTime(java.time.LocalDateTime.now());
        mapper.updateById(dataPermColumn);
    }

    @Override
    @Transactional(rollbackFor = Exception.class)
    public void delete(Long id) {
        if (id == null) {
            return;
        }
        mapper.deleteById(id);
    }

    @Override
    public DataPermColumn getById(Long id) {
        if (id == null) {
            return null;
        }
        return mapper.selectById(id);
    }

    @Override
    public List<DataPermColumn> listByDatasetId(Long datasetId) {
        if (datasetId == null) {
            return List.of();
        }
        return mapper.listByDatasetId(datasetId);
    }

    @Override
    public void updateStatus(Long id, Integer status) {
        if (id == null) {
            return;
        }
        DataPermColumn entity = new DataPermColumn();
        entity.setId(id);
        entity.setStatus(status != null ? status : 1);
        entity.setUpdateTime(java.time.LocalDateTime.now());
        mapper.updateById(entity);
    }

    @Override
    public DataPermColumn getByDatasetIdAndField(Long datasetId, String fieldName) {
        if (datasetId == null || fieldName == null) {
            return null;
        }
        return mapper.getByDatasetIdAndField(datasetId, fieldName);
    }

    @Override
    public Integer checkFieldPermExists(Long datasetId, String fieldName) {
        if (datasetId == null || fieldName == null) {
            return 0;
        }
        Integer count = mapper.checkFieldPermExists(datasetId, fieldName);
        return count != null ? count : 0;
    }

    @Override
    public List<DataPermColumn> listByDatasetIdAndStatus(Long datasetId, Integer status) {
        if (datasetId == null) {
            return List.of();
        }
        if (status == null) {
            return mapper.listByDatasetId(datasetId);
        }
        List<DataPermColumn> allColumns = mapper.listByDatasetId(datasetId);
        return allColumns.stream()
            .filter(col -> col.getStatus().equals(status))
            .collect(java.util.stream.Collectors.toList());
    }

    @Override
    public List<String> listAllColumnNames(Long datasetId) {
        if (datasetId == null) {
            return List.of();
        }
        return mapper.listAllColumnNamesByDatasetId(datasetId);
    }
}
