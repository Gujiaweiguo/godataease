package io.dataease.system.dao.auto.mapper;

import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import io.dataease.system.entity.SysPerm;
import io.swagger.v3.oas.annotations.media.Schema;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;
import org.apache.ibatis.annotations.Select;

import java.util.List;

@Mapper
public interface SysPermMapper extends BaseMapper<SysPerm> {

    @Select("SELECT * FROM sys_perm WHERE status = 1 ORDER BY create_time DESC")
    List<SysPerm> listAllPerms();

    @Select("SELECT * FROM sys_perm WHERE perm_id = #{permId} AND status = 1")
    SysPerm getPermById(@Param("permId") Long permId);

    @Select("SELECT * FROM sys_perm WHERE perm_type = #{permType} AND status = 1 ORDER BY create_time DESC")
    List<SysPerm> listByPermType(@Param("permType") String permType);

    @Select("SELECT * FROM sys_perm WHERE perm_code = #{permCode} AND status = 1")
    SysPerm getPermByCode(@Param("permCode") String permCode);

    @Select("SELECT COUNT(*) FROM sys_perm WHERE perm_code = #{permCode} AND status = 1")
    Integer checkPermCodeExists(@Param("permCode") String permCode);
}
