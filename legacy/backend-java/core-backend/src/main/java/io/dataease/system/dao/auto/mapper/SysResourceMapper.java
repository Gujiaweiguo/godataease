package io.dataease.system.dao.auto.mapper;

import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import io.dataease.system.entity.SysResource;
import io.swagger.v3.oas.annotations.media.Schema;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;
import org.apache.ibatis.annotations.Select;

import java.util.List;

@Mapper
public interface SysResourceMapper extends BaseMapper<SysResource> {

    @Select("SELECT * FROM sys_resource ORDER BY create_time DESC")
    List<SysResource> listResources();

    @Select("SELECT * FROM sys_resource WHERE resource_id = #{resourceId} AND status = 1")
    SysResource getResourceById(@Param("resourceId") Long resourceId);

    @Select("SELECT * FROM sys_resource WHERE parent_id = #{parentId} AND status = 1 ORDER BY create_time DESC")
    List<SysResource> listByParentId(@Param("parentId") Long parentId);

    @Select("SELECT * FROM sys_resource WHERE resource_type = #{resourceType} AND status = 1 ORDER BY create_time DESC")
    List<SysResource> listByResourceType(@Param("resourceType") String resourceType);

    @Select("SELECT COUNT(*) FROM sys_resource WHERE resource_name = #{resourceName} AND status = 1")
    Integer checkResourceNameExists(@Param("resourceName") String resourceName);
}
