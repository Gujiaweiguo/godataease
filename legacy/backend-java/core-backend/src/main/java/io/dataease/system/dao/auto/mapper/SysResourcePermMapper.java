package io.dataease.system.dao.auto.mapper;

import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import io.dataease.system.entity.SysResourcePerm;
import io.swagger.v3.oas.annotations.media.Schema;
import java.util.List;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;
import org.apache.ibatis.annotations.Insert;
import org.apache.ibatis.annotations.Delete;
import org.apache.ibatis.annotations.Select;

@Mapper
public interface SysResourcePermMapper extends BaseMapper<SysResourcePerm> {

    @Select("SELECT * FROM sys_resource_perm WHERE status = 1 ORDER BY create_time DESC")
    List<SysResourcePerm> listAll();

    @Select("SELECT * FROM sys_resource_perm WHERE resource_id = #{resourceId}")
    List<SysResourcePerm> listByResourceId(@Param("resourceId") Long resourceId);

    @Insert("INSERT INTO sys_resource_perm (resource_id, perm_id, create_by, create_time) VALUES (#{resourceId}, #{permId}, #{createBy}, NOW())")
    void insertResourcePerm(@Param("resourceId") Long resourceId, @Param("permId") Long permId, @Param("createBy") String createBy);

    @Delete("DELETE FROM sys_resource_perm WHERE resource_id = #{resourceId} AND perm_id = #{permId}")
    void deleteResourcePerm(@Param("resourceId") Long resourceId, @Param("permId") Long permId);

    @Select("SELECT COUNT(*) FROM sys_resource_perm WHERE resource_id = #{resourceId}")
    Integer countResourcePerms(@Param("resourceId") Long resourceId);

    @Select("SELECT COUNT(*) FROM sys_resource_perm WHERE resource_id = #{resourceId} AND perm_id = #{permId}")
    Integer checkResourcePermExists(@Param("resourceId") Long resourceId, @Param("permId") Long permId);
}
