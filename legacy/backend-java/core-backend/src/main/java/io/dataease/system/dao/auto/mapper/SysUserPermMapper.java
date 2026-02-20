package io.dataease.system.dao.auto.mapper;

import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import io.dataease.system.entity.SysUserPerm;
import io.swagger.v3.oas.annotations.media.Schema;
import org.apache.ibatis.annotations.Delete;
import org.apache.ibatis.annotations.Insert;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;
import org.apache.ibatis.annotations.Select;

import java.util.List;

@Mapper
public interface SysUserPermMapper extends BaseMapper<SysUserPerm> {

    @Select("SELECT * FROM sys_user_perm WHERE status = 1 AND del_flag = 0 ORDER BY create_time DESC")
    List<SysUserPerm> listUserPerms(@Param("userId") Long userId);

    @Select("SELECT * FROM sys_user_perm WHERE user_id = #{userId} AND org_id = #{orgId} AND status = 1 AND del_flag = 0 ORDER BY create_time DESC")
    List<SysUserPerm> listUserOrgPerms(@Param("userId") Long userId, @Param("orgId") Long orgId);

    @Select("SELECT COUNT(*) FROM sys_user_perm WHERE user_id = #{userId} AND perm_id = #{permId} AND org_id = #{orgId} AND status = 1 AND del_flag = 0")
    Integer checkUserPermExists(@Param("userId") Long userId, @Param("orgId") Long orgId, @Param("permId") Long permId);

    @Select("SELECT * FROM sys_user_perm WHERE perm_id = #{permId} AND org_id = #{orgId} AND status = 1 AND del_flag = 0")
    List<SysUserPerm> listUsersByPermAndOrg(@Param("permId") Long permId, @Param("orgId") Long orgId);

    @Insert("INSERT INTO sys_user_perm (user_id, org_id, perm_id, create_by, create_time) VALUES (#{userId}, #{orgId}, #{permId}, #{createBy}, NOW())")
    void insertUserPerm(@Param("userId") Long userId, @Param("orgId") Long orgId, @Param("permId") Long permId, @Param("createBy") String createBy);

    @Delete("DELETE FROM sys_user_perm WHERE user_id = #{userId} AND org_id = #{orgId} AND perm_id = #{permId}")
    void deleteUserPerm(@Param("userId") Long userId, @Param("orgId") Long orgId, @Param("permId") Long permId);
}
