package io.dataease.system.dao.auto.mapper;

import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import io.dataease.system.entity.SysRole;
import io.swagger.v3.oas.annotations.media.Schema;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;
import org.apache.ibatis.annotations.Select;
import org.apache.ibatis.annotations.Update;

import java.util.List;

@Mapper
public interface SysRoleMapper extends BaseMapper<SysRole> {

    @Select("SELECT * FROM sys_role WHERE status = 1 ORDER BY create_time DESC")
    List<SysRole> listRoles();

    @Select("SELECT * FROM sys_role WHERE role_id = #{roleId} AND status = 1")
    SysRole getRoleById(@Param("roleId") Long roleId);

    @Select("SELECT COUNT(*) FROM sys_role WHERE role_code = #{roleCode} AND status = 1")
    Integer checkRoleCodeExists(@Param("roleCode") String roleCode);

    @Select("SELECT * FROM sys_role WHERE parent_id = #{parentId} AND status = 1 ORDER BY create_time DESC")
    List<SysRole> listByParentId(@Param("parentId") Long parentId);

    @Update("UPDATE sys_role SET role_name = #{roleName}, role_desc = #{roleDesc}, data_scope = #{dataScope}, update_time = NOW() WHERE role_id = #{roleId}")
    void updateRole(@Param("roleId") Long roleId, @Param("roleName") String roleName, @Param("roleDesc") String roleDesc, @Param("dataScope") String dataScope);

    @Update("UPDATE sys_role SET status = #{status}, update_time = NOW() WHERE role_id = #{roleId}")
    void updateRoleStatus(@Param("roleId") Long roleId, @Param("status") Integer status);

    @Update("UPDATE sys_role SET del_flag = 1, update_time = NOW() WHERE role_id = #{roleId}")
    void deleteRole(@Param("roleId") Long roleId);
}
