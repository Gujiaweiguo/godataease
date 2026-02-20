package io.dataease.system.dao.auto.mapper;

import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import io.dataease.system.entity.SysRolePerm;
import io.swagger.v3.oas.annotations.media.Schema;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;
import org.apache.ibatis.annotations.Select;

import java.util.List;

@Mapper
public interface SysRolePermMapper extends BaseMapper<SysRolePerm> {

    @Select("SELECT * FROM sys_role_perm WHERE status = 1 ORDER BY create_time DESC")
    List<SysRolePerm> listAll();

    @Select("SELECT * FROM sys_role_perm WHERE role_id = #{roleId}")
    List<java.lang.Object> listByRoleId(@Param("roleId") Long roleId);

    @Select("SELECT * FROM sys_role_perm WHERE user_id = #{userId} AND org_id = #{orgId}")
    List<java.lang.Object> listByUserIdAndOrgId(@Param("userId") Long userId, @Param("orgId") Long orgId);

    @Select("SELECT DISTINCT role_id FROM sys_role_perm WHERE user_id = #{userId} AND org_id = #{orgId}")
    List<Long> listRoleIdsByUserIdAndOrgId(@Param("userId") Long userId, @Param("orgId") Long orgId);
}

