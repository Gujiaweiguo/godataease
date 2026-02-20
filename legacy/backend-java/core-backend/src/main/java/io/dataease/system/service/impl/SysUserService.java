package io.dataease.system.service.impl;

import com.baomidou.mybatisplus.core.conditions.query.QueryWrapper;
import io.dataease.audit.annotation.AuditLog;
import io.dataease.audit.constant.AuditConstants;
import io.dataease.system.dao.auto.mapper.SysUserMapper;
import io.dataease.system.entity.SysUser;
import io.dataease.system.service.ISysUserService;
import org.springframework.stereotype.Service;

import java.time.LocalDateTime;
import java.util.List;

@Service
public class SysUserService implements ISysUserService {

    private final SysUserMapper sysUserMapper;

    public SysUserService(SysUserMapper sysUserMapper) {
        this.sysUserMapper = sysUserMapper;
    }

    @Override
    @AuditLog(
        actionType = AuditConstants.ACTION_TYPE_USER_ACTION,
        actionName = "CREATE_USER",
        resourceType = AuditConstants.RESOURCE_TYPE_USER
    )
    public Long createUser(SysUser user) {
        user.setCreateTime(LocalDateTime.now());
        user.setUpdateTime(LocalDateTime.now());
        user.setFrom(0);
        user.setDelFlag(0);
        sysUserMapper.insert(user);
        return user.getUserId();
    }

    @Override
    @AuditLog(
        actionType = AuditConstants.ACTION_TYPE_USER_ACTION,
        actionName = "UPDATE_USER",
        resourceType = AuditConstants.RESOURCE_TYPE_USER
    )
    public void updateUser(SysUser user) {
        user.setUpdateTime(LocalDateTime.now());
        sysUserMapper.updateById(user);
    }

    @Override
    @AuditLog(
        actionType = AuditConstants.ACTION_TYPE_USER_ACTION,
        actionName = "DELETE_USER",
        resourceType = AuditConstants.RESOURCE_TYPE_USER
    )
    public void deleteUser(Long userId) {
        sysUserMapper.deleteById(userId);
    }

    @Override
    public SysUser getUserById(Long userId) {
        return sysUserMapper.selectById(userId);
    }

    @Override
    public SysUser getUserByUsername(String username) {
        QueryWrapper<SysUser> queryWrapper = new QueryWrapper<>();
        queryWrapper.eq("username", username);
        queryWrapper.eq("del_flag", 0);
        return sysUserMapper.selectOne(queryWrapper);
    }

    @Override
    public List<SysUser> searchUsers(String keyword, Long orgId, Integer status) {
        QueryWrapper<SysUser> queryWrapper = new QueryWrapper<>();
        queryWrapper.eq("del_flag", 0);
        if (keyword != null && !keyword.isEmpty()) {
            queryWrapper.and(wrapper -> wrapper.like("username", "%" + keyword + "%")
                    .or()
                    .like("nick_name", "%" + keyword + "%")
                    .or()
                    .like("email", "%" + keyword + "%"));
        }
        if (status != null) {
            queryWrapper.eq("status", status);
        }
        queryWrapper.orderByDesc("create_time");
        return sysUserMapper.selectList(queryWrapper);
    }

    @Override
    public Integer countByUsername(String username) {
        QueryWrapper<SysUser> queryWrapper = new QueryWrapper<>();
        queryWrapper.eq("username", username);
        queryWrapper.eq("del_flag", 0);
        return Math.toIntExact(sysUserMapper.selectCount(queryWrapper));
    }

    @Override
    public Integer checkEmailExists(String email, Long userId) {
        QueryWrapper<SysUser> queryWrapper = new QueryWrapper<>();
        queryWrapper.eq("email", email);
        queryWrapper.eq("del_flag", 0);
        if (userId != null) {
            queryWrapper.ne("user_id", userId);
        }
        return Math.toIntExact(sysUserMapper.selectCount(queryWrapper));
    }

    @Override
    @AuditLog(
        actionType = AuditConstants.ACTION_TYPE_USER_ACTION,
        actionName = "RESET_PASSWORD",
        resourceType = AuditConstants.RESOURCE_TYPE_USER
    )
    public void resetPassword(Long userId, String newPassword) {
        SysUser user = new SysUser();
        user.setUserId(userId);
        user.setPassword(newPassword);
        user.setUpdateTime(LocalDateTime.now());
        sysUserMapper.updateById(user);
    }

    @Override
    @AuditLog(
        actionType = AuditConstants.ACTION_TYPE_USER_ACTION,
        actionName = "UPDATE_USER_STATUS",
        resourceType = AuditConstants.RESOURCE_TYPE_USER
    )
    public void updateUserStatus(Long userId, Integer status) {
        SysUser user = new SysUser();
        user.setUserId(userId);
        user.setStatus(status);
        user.setUpdateTime(LocalDateTime.now());
        sysUserMapper.updateById(user);
    }

    @Override
    public List<SysUser> listUsersByIds(List<Long> userIds, Long orgId) {
        if (userIds == null || userIds.isEmpty()) {
            return List.of();
        }
        QueryWrapper<SysUser> queryWrapper = new QueryWrapper<>();
        queryWrapper.in("user_id", userIds);
        queryWrapper.eq("del_flag", 0);
        if (orgId != null) {
            queryWrapper.eq("org_id", orgId);
        }
        queryWrapper.orderByDesc("create_time");
        return sysUserMapper.selectList(queryWrapper);
    }
}
