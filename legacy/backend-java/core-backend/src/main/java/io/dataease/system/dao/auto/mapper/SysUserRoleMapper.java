package io.dataease.system.dao.auto.mapper;

import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import io.dataease.system.entity.SysUserRole;
import io.swagger.v3.oas.annotations.media.Schema;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;
import org.apache.ibatis.annotations.Select;

import java.util.List;

@Mapper
public interface SysUserRoleMapper extends BaseMapper<SysUserRole> {

    @Select("SELECT * FROM sys_user_role WHERE user_id = #{userId} AND org_id = #{orgId}")
    List<SysUserRole> listByUserIdAndOrgId(@Param("userId") Long userId, @Param("orgId") Long orgId);

    @Select("SELECT * FROM sys_user_role WHERE role_id = #{roleId} AND org_id = #{orgId}")
    List<SysUserRole> listByRoleIdAndOrgId(@Param("roleId") Long roleId, @Param("orgId") Long orgId);

    @Select("DELETE FROM sys_user_role WHERE user_id = #{userId} AND org_id = #{orgId}")
    void deleteByUserIdAndOrgId(@Param("userId") Long userId, @Param("orgId") Long orgId);

    @Select("DELETE FROM sys_user_role WHERE role_id = #{roleId}")
    void deleteByRoleId(@Param("roleId") Long roleId);
}
