package io.dataease.system.dao.auto.mapper;

import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import io.dataease.system.entity.SysRoleMenu;
import io.swagger.v3.oas.annotations.media.Schema;
import java.util.List;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;
import org.apache.ibatis.annotations.Insert;
import org.apache.ibatis.annotations.Delete;
import org.apache.ibatis.annotations.Select;

@Mapper
public interface SysRoleMenuMapper extends BaseMapper<SysRoleMenu> {

    @Select("SELECT * FROM sys_role_menu WHERE role_id = #{roleId}")
    List<Long> listMenuIdsByRoleId(@Param("roleId") Long roleId);

    @Select("SELECT * FROM sys_role_menu WHERE role_id = #{roleId}")
    List<java.lang.Object> listRoleMenusByRoleId(@Param("roleId") Long roleId);

    @Insert("INSERT INTO sys_role_menu (role_id, menu_id, create_by, create_time) VALUES (#{roleId}, #{menuId}, #{createBy}, NOW())")
    void insertRoleMenu(@Param("roleId") Long roleId, @Param("menuId") Long menuId, @Param("createBy") String createBy);

    @Delete("DELETE FROM sys_role_menu WHERE role_id = #{roleId}")
    void deleteRoleMenusByRoleId(@Param("roleId") Long roleId);
}
