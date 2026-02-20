package io.dataease.system.service;

import io.dataease.system.entity.DataPermRow;

import java.util.List;

public interface IDataPermRowService {

    Long create(DataPermRow dataPermRow);

    void update(DataPermRow dataPermRow);

    void delete(Long id);

    DataPermRow getById(Long id);

    List<DataPermRow> listByDatasetId(Long datasetId);

    void updateStatus(Long id, Integer status);
}
