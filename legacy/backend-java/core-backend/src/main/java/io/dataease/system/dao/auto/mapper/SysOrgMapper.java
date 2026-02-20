package io.dataease.system.dao.auto.mapper;

import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import io.dataease.system.entity.SysOrg;
import io.swagger.v3.oas.annotations.media.Schema;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;
import org.apache.ibatis.annotations.Select;
import org.apache.ibatis.annotations.Update;

import java.util.List;

@Mapper
public interface SysOrgMapper extends BaseMapper<SysOrg> {

    @Select("SELECT * FROM sys_org WHERE status = 1 AND del_flag = 0 ORDER BY create_time DESC")
    List<SysOrg> listOrgs();

    @Select("SELECT * FROM sys_org WHERE org_id = #{orgId} AND status = 1 AND del_flag = 0")
    SysOrg getOrgById(@Param("orgId") Long orgId);

    @Select("SELECT COUNT(*) FROM sys_org WHERE org_name = #{orgName} AND del_flag = 0")
    Integer checkOrgNameExists(@Param("orgName") String orgName);

    @Select("SELECT * FROM sys_org WHERE parent_id = #{parentId} AND status = 1 AND del_flag = 0 ORDER BY create_time DESC")
    List<SysOrg> listByParentId(@Param("parentId") Long parentId);

    @Update("UPDATE sys_org SET org_name = #{orgName}, org_desc = #{orgDesc}, update_time = NOW() WHERE org_id = #{orgId}")
    void updateOrg(@Param("orgId") Long orgId, @Param("orgName") String orgName, @Param("orgDesc") String orgDesc);

    @Update("UPDATE sys_org SET status = #{status}, update_time = NOW() WHERE org_id = #{orgId}")
    void updateOrgStatus(@Param("orgId") Long orgId, @Param("status") Integer status);

    @Update("UPDATE sys_org SET del_flag = 1, update_time = NOW() WHERE org_id = #{orgId}")
    void deleteOrg(@Param("orgId") Long orgId);

    @Select("SELECT COUNT(*) FROM sys_org WHERE parent_id = #{orgId} AND status = 1 AND del_flag = 0")
    Integer countChildOrgs(@Param("orgId") Long orgId);
}
