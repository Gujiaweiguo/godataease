package io.dataease.system.dao.auto.mapper;

import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import io.dataease.system.entity.DataPermColumn;
import io.swagger.v3.oas.annotations.media.Schema;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;
import org.apache.ibatis.annotations.Select;

import java.util.List;

@Mapper
public interface DataPermColumnMapper extends BaseMapper<DataPermColumn> {

    @Select("SELECT * FROM data_perm_column WHERE dataset_id = #{datasetId}")
    List<DataPermColumn> listByDatasetId(@Param("datasetId") Long datasetId);

    @Select("SELECT * FROM data_perm_column WHERE dataset_id = #{datasetId} AND field_name = #{fieldName}")
    DataPermColumn getByDatasetIdAndField(@Param("datasetId") Long datasetId, @Param("fieldName") String fieldName);

    @Select("SELECT COUNT(*) FROM data_perm_column WHERE dataset_id = #{datasetId} AND field_name = #{fieldName}")
    Integer checkFieldPermExists(@Param("datasetId") Long datasetId, @Param("fieldName") String fieldName);

    @Select("SELECT DISTINCT field_name FROM data_perm_column WHERE dataset_id = #{datasetId} AND status = 1")
    List<String> listAllColumnNamesByDatasetId(@Param("datasetId") Long datasetId);
}
