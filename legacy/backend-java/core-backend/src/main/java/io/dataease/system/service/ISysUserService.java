package io.dataease.system.service;

import io.dataease.system.entity.SysUser;
import java.util.List;

public interface ISysUserService {

    Long createUser(SysUser user);

    void updateUser(SysUser user);

    void deleteUser(Long userId);

    SysUser getUserById(Long userId);

    SysUser getUserByUsername(String username);

    List<SysUser> searchUsers(String keyword, Long orgId, Integer status);

    Integer countByUsername(String username);

    Integer checkEmailExists(String email, Long userId);

    void resetPassword(Long userId, String newPassword);

    void updateUserStatus(Long userId, Integer status);

    List<SysUser> listUsersByIds(List<Long> userIds, Long orgId);
}
