package io.dataease.system.service;

import io.dataease.system.entity.SysRole;

import java.util.List;

public interface IRoleService {

    void createRole(SysRole role);

    void updateRole(SysRole role);

    void deleteRole(Long roleId);

    SysRole getRoleById(Long roleId);

    List<SysRole> listRoles();

    Integer checkRoleCodeExists(String roleCode);

    void updateRoleStatus(Long roleId, Integer status);

    void assignUsersToRole(Long roleId, List<Long> userIds, Long orgId);

    void assignPermissionsToRole(Long roleId, List<Long> permIds);

    List<Long> getRoleUserIds(Long roleId);

    List<Long> getRolePermIds(Long roleId);
}
