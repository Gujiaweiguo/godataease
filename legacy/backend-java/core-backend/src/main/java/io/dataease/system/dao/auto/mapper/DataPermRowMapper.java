package io.dataease.system.dao.auto.mapper;

import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import io.dataease.system.entity.DataPermRow;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;

import java.util.List;

@Mapper
public interface DataPermRowMapper extends BaseMapper<DataPermRow> {

    List<DataPermRow> listByDatasetId(@Param("datasetId") Long datasetId);

    DataPermRow getByDatasetIdAndTarget(@Param("datasetId") Long datasetId,
                                      @Param("authTargetType") String authTargetType,
                                      @Param("authTargetId") Long authTargetId);

    void updateStatus(@Param("id") Long id, @Param("status") Integer status);
}
