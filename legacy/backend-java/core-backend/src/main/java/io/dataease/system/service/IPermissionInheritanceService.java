package io.dataease.system.service;

import io.dataease.system.entity.SysPerm;
import io.dataease.system.entity.SysRolePerm;
import io.dataease.system.entity.SysResourcePerm;
import io.dataease.system.entity.SysRoleMenu;

import java.util.List;

public interface IPermissionInheritanceService {

    List<SysPerm> resolvePermissionsForUser(Long userId, Long orgId);

    List<SysRolePerm> resolvePermissionsForRole(Long roleId, Long orgId);

    List<SysPerm> resolveEffectivePermissions(Long userId, Long orgId);

    List<SysResourcePerm> resolvePermissionsForResource(Long resourceId, Long orgId);

    List<SysRoleMenu> resolveMenusForRole(Long roleId);

    List<SysRoleMenu> resolveMenusForOrg(Long orgId);
}

