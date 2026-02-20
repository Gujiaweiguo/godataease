package io.dataease.system.service;

import io.dataease.system.entity.DataPermColumn;

import java.util.List;

public interface IDataPermColumnService {

    Long create(DataPermColumn dataPermColumn);

    void update(DataPermColumn dataPermColumn);

    void delete(Long id);

    DataPermColumn getById(Long id);

    List<DataPermColumn> listByDatasetId(Long datasetId);

    void updateStatus(Long id, Integer status);

    DataPermColumn getByDatasetIdAndField(Long datasetId, String fieldName);

    Integer checkFieldPermExists(Long datasetId, String fieldName);

    List<DataPermColumn> listByDatasetIdAndStatus(Long datasetId, Integer status);

    List<String> listAllColumnNames(Long datasetId);
}
