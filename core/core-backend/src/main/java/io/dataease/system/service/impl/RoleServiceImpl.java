import io.dataease.system.dao.auto.mapper.SysRoleMapper;
import io.dataease.system.dao.auto.mapper.SysRoleMenuMapper;
import io.dataease.system.entity.SysRolePerm;
import io.dataease.system.entity.SysRole;
import io.dataease.system.service.IRoleService;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.List;

@Service
@Slf4j
public class RoleServiceImpl implements IRoleService {

    @Autowired
    private SysRoleMapper roleMapper;
    @Autowired
    private SysRoleMenuMapper roleMenuMapper;
    @Autowired
    private SysRolePerm rolePermMapper;

    @Override
    @io.dataease.audit.annotation.AuditLog(actionType = AuditConstants.ACTION_TYPE_PERMISSION_CHANGE, actionName = "CREATE_ROLE", resourceType = AuditConstants.RESOURCE_TYPE_ROLE)
    @Transactional(rollbackFor = Exception.class)
    public void createRole(SysRole role) {
        role.setCreateTime(LocalDateTime.now());
        role.setUpdateTime(LocalDateTime.now());
        role.setLevel(1);
        role.setStatus(1);
        roleMapper.insert(role);
        log.info("Role created: {}", role.getRoleCode());
    }

    @Override
    @Transactional(rollbackFor = Exception.class)
    public void updateRole(SysRole role) {
        if (role.getRoleId() == null) {
            return;
        }
        role.setUpdateTime(java.time.LocalDateTime.now());
        roleMapper.updateById(role);
    }

    @Override
    @Transactional(rollbackFor = Exception.class)
    public void deleteRole(Long roleId) {
        if (roleId == null) {
            return;
        }
        SysRole role = new SysRole();
        role.setRoleId(roleId);
        role.setStatus(0);
        role.setUpdateTime(java.time.LocalDateTime.now());
        roleMapper.updateById(role);
    }

    @Override
    public SysRole getRoleById(Long roleId) {
        if (roleId == null) {
            return null;
        }
        return roleMapper.selectById(roleId);
    }

    @Override
    public List<SysRole> listRoles() {
        return roleMapper.selectList(null);
    }

    @Override
    public Integer checkRoleCodeExists(String roleCode) {
        if (roleCode == null || roleCode.trim().isEmpty()) {
            return 0;
        }
        return roleMapper.selectCount(
            new com.baomidou.mybatisplus.core.conditions.query.QueryWrapper<SysRole>()
                .eq("role_code", roleCode)
        ).intValue();
    }

    @Override
    @Transactional(rollbackFor = Exception.class)
    public void updateRoleStatus(Long roleId, Integer status) {
        if (roleId == null) {
            return;
        }
        SysRole role = new SysRole();
        role.setRoleId(roleId);
        role.setStatus(status != null ? status : 1);
        role.setUpdateTime(java.time.LocalDateTime.now());
        roleMapper.updateById(role);
    }

    @Override
    @Transactional(rollbackFor = Exception.class)
    public void assignUsersToRole(Long roleId, List<Long> userIds, Long orgId) {
        if (roleId == null || userIds == null || userIds.isEmpty()) {
            return;
        }
        for (Long userId : userIds) {
            io.dataease.system.dao.auto.mapper.SysUserRoleMapper mapper = userRoleMapper;
            com.baomidou.mybatisplus.core.conditions.query.QueryWrapper<io.dataease.system.entity.SysUserRole> wrapper =
                new com.baomidou.mybatisplus.core.conditions.query.QueryWrapper<>();
            wrapper.eq("user_id", userId);
            wrapper.eq("role_id", roleId);
            wrapper.eq("org_id", orgId);
            Long existingCount = mapper.selectCount(wrapper);
            if (existingCount != null && existingCount > 0) {
                continue;
            }
            io.dataease.system.entity.SysUserRole userRole = new io.dataease.system.entity.SysUserRole();
            userRole.setUserId(userId);
            userRole.setRoleId(roleId);
            userRole.setOrgId(orgId);
            userRole.setCreateTime(java.time.LocalDateTime.now());
            userRoleMapper.insert(userRole);
        }
    }

    @Override
    @Transactional(rollbackFor = Exception.class)
    public void assignPermissionsToRole(Long roleId, List<Long> permIds) {
        if (roleId == null || permIds == null || permIds.isEmpty()) {
            return;
        }
        for (Long permId : permIds) {
            SysRolePerm rolePerm = new SysRolePerm();
            rolePerm.setRoleId(roleId);
            rolePerm.setPermId(permId);
            rolePermMapper.insert(rolePerm);
        }
    }

    @Override
    public List<Long> getRoleUserIds(Long roleId) {
        if (roleId == null) {
            return List.of();
        }
        return userRoleMapper.selectList(
            new com.baomidou.mybatisplus.core.conditions.query.QueryWrapper<io.dataease.system.entity.SysUserRole>()
                .eq("role_id", roleId)
        ).stream().map(io.dataease.system.entity.SysUserRole::getUserId).toList();
    }

    @Override
    public List<Long> getRolePermIds(Long roleId) {
        if (roleId == null) {
            return List.of();
        }
        return rolePermMapper.selectList(
            new com.baomidou.mybatisplus.core.conditions.query.QueryWrapper<SysRolePerm>()
                .eq("role_id", roleId)
        ).stream().map(SysRolePerm::getPermId).toList();
    }
}
