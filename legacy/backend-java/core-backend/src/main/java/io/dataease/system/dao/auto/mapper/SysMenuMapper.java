package io.dataease.system.dao.auto.mapper;

import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import io.dataease.system.entity.SysMenu;
import io.swagger.v3.oas.annotations.media.Schema;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;
import org.apache.ibatis.annotations.Select;

import java.util.List;

@Mapper
public interface SysMenuMapper extends BaseMapper<SysMenu> {

    @Select("SELECT * FROM sys_menu WHERE status = 1 ORDER BY order_num ASC")
    List<SysMenu> listAllMenus();

    @Select("SELECT * FROM sys_menu WHERE status = 1 AND parent_id = #{parentId} ORDER BY order_num ASC")
    List<SysMenu> listByParentId(@Param("parentId") Long parentId);

    @Select("SELECT * FROM sys_menu WHERE menu_id = #{menuId} AND status = 1")
    SysMenu getMenuById(@Param("menuId") Long menuId);
}
