package io.dataease.system.service.impl;

import com.baomidou.mybatisplus.core.conditions.query.QueryWrapper;
import io.dataease.system.dao.auto.mapper.DataPermRowMapper;
import io.dataease.system.entity.DataPermRow;
import io.dataease.system.service.IDataPermRowService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.time.LocalDateTime;
import java.util.List;

@Service
public class DataPermRowServiceImpl implements IDataPermRowService {

    @Autowired
    private DataPermRowMapper dataPermRowMapper;

    @Override
    public Long create(DataPermRow dataPermRow) {
        dataPermRow.setCreateTime(LocalDateTime.now());
        dataPermRow.setUpdateTime(LocalDateTime.now());
        dataPermRowMapper.insert(dataPermRow);
        return dataPermRow.getId();
    }

    @Override
    public void update(DataPermRow dataPermRow) {
        dataPermRow.setUpdateTime(LocalDateTime.now());
        dataPermRowMapper.updateById(dataPermRow);
    }

    @Override
    public void delete(Long id) {
        dataPermRowMapper.deleteById(id);
    }

    @Override
    public DataPermRow getById(Long id) {
        return dataPermRowMapper.selectById(id);
    }

    @Override
    public List<DataPermRow> listByDatasetId(Long datasetId) {
        QueryWrapper<DataPermRow> queryWrapper = new QueryWrapper<>();
        queryWrapper.eq("dataset_id", datasetId);
        queryWrapper.eq("status", 1);
        queryWrapper.orderByDesc("create_time");
        return dataPermRowMapper.listByDatasetId(datasetId);
    }

    @Override
    public void updateStatus(Long id, Integer status) {
        DataPermRow dataPermRow = new DataPermRow();
        dataPermRow.setId(id);
        dataPermRow.setStatus(status);
        dataPermRow.setUpdateTime(LocalDateTime.now());
        dataPermRowMapper.updateById(dataPermRow);
    }
}
