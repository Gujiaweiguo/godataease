package io.dataease.system.dao.auto.mapper;

import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import io.dataease.system.entity.SysUser;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;
import org.apache.ibatis.annotations.Select;

import java.util.List;
import java.util.Map;

@Mapper
public interface SysUserMapper extends BaseMapper<SysUser> {

    @Select("<script>" +
            "SELECT * FROM sys_user " +
            "WHERE del_flag = 0 " +
            "<if test='keyword != null and keyword != \"\"'>" +
            "AND (username LIKE CONCAT('%', #{keyword}, '%') " +
            "OR nick_name LIKE CONCAT('%', #{keyword}, '%') " +
            "OR email LIKE CONCAT('%', #{keyword}, '%') " +
            "OR phone LIKE CONCAT('%', #{keyword}, '%')) " +
            "</if>" +
            "<if test='orgId != null'>" +
            "AND org_id = #{orgId} " +
            "</if>" +
            "<if test='status != null'>" +
            "AND status = #{status} " +
            "</if>" +
            "ORDER BY create_time DESC " +
            "</script>")
    List<SysUser> searchUsers(@Param("keyword") String keyword,
                           @Param("orgId") Long orgId,
                           @Param("status") Integer status);

    @Select("SELECT user_id FROM sys_user WHERE username = #{username} AND del_flag = 0")
    Long getUserIdByUsername(@Param("username") String username);

    @Select("SELECT COUNT(*) FROM sys_user WHERE username = #{username} AND del_flag = 0")
    Integer countByUsername(@Param("username") String username);

    @Select("SELECT COUNT(*) FROM sys_user WHERE email = #{email} AND del_flag = 0 AND user_id != #{userId}")
    Integer checkEmailExists(@Param("email") String email, @Param("userId") Long userId);

    @Select("SELECT * FROM sys_user WHERE user_id IN " +
            "<foreach collection='orgUserIds' item='user_id' open='(' separator=','>" +
            "#{orgUserIds}" +
            "</foreach>" +
            "AND del_flag = 0 " +
            "<if test='orgId != null'>" +
            "AND org_id = #{orgId} " +
            "</if>")
    List<SysUser> listUsersByIds(@Param("orgUserIds") List<Long> orgUserIds, @Param("orgId") Long orgId);
}
