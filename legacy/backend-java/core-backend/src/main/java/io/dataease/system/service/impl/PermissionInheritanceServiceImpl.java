package io.dataease.system.service.impl;

import io.dataease.system.dao.auto.mapper.SysPermMapper;
import io.dataease.system.service.IPermissionInheritanceService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.util.List;

@Service
public class PermissionInheritanceServiceImpl implements IPermissionInheritanceService {

    @Autowired
    private SysPermMapper sysPermMapper;

    @Override
    public List<io.dataease.system.entity.SysPerm> resolvePermissionsForUser(Long userId, Long orgId) {
        if (userId == null) {
            return List.of();
        }

        return sysPermMapper.listAllPerms();
    }

    @Override
    public List<io.dataease.system.entity.SysRolePerm> resolvePermissionsForRole(Long roleId, Long orgId) {
        if (roleId == null) {
            return List.of();
        }

        return List.of();
    }

    @Override
    public List<io.dataease.system.entity.SysPerm> resolveEffectivePermissions(Long userId, Long orgId) {
        if (userId == null) {
            return List.of();
        }

        return sysPermMapper.listAllPerms();
    }

    @Override
    public List<io.dataease.system.entity.SysResourcePerm> resolvePermissionsForResource(Long resourceId, Long orgId) {
        return List.of();
    }

    @Override
    public List<io.dataease.system.entity.SysRoleMenu> resolveMenusForRole(Long roleId) {
        return List.of();
    }

    @Override
    public List<io.dataease.system.entity.SysRoleMenu> resolveMenusForOrg(Long orgId) {
        return List.of();
    }
}
