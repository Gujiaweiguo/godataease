package io.dataease.system.service.impl;

import io.dataease.system.dao.auto.mapper.SysPermMapper;
import io.dataease.system.dao.auto.mapper.SysResourceMapper;
import io.dataease.system.dao.auto.mapper.SysMenuMapper;
import io.dataease.system.entity.SysPerm;
import io.dataease.system.entity.SysResource;
import io.dataease.system.entity.SysMenu;
import io.dataease.system.service.IPermService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.List;

@Service
public class PermServiceImpl implements IPermService {

    @Autowired
    private SysPermMapper permMapper;

    @Autowired
    private SysResourceMapper resourceMapper;

    @Autowired
    private SysMenuMapper menuMapper;

    @Override
    @Transactional(rollbackFor = Exception.class)
    public void createPerm(SysPerm perm) {
        perm.setCreateTime(java.time.LocalDateTime.now());
        perm.setUpdateTime(java.time.LocalDateTime.now());
        perm.setStatus(1);
        perm.setDelFlag(0);
        permMapper.insert(perm);
    }

    @Override
    @Transactional(rollbackFor = Exception.class)
    public void updatePerm(SysPerm perm) {
        if (perm.getPermId() == null) {
            return;
        }
        perm.setUpdateTime(java.time.LocalDateTime.now());
        permMapper.updateById(perm);
    }

    @Override
    @Transactional(rollbackFor = Exception.class)
    public void deletePerm(Long permId) {
        if (permId == null) {
            return;
        }
        SysPerm perm = new SysPerm();
        perm.setPermId(permId);
        perm.setStatus(0);
        perm.setDelFlag(1);
        perm.setUpdateTime(java.time.LocalDateTime.now());
        permMapper.updateById(perm);
    }

    @Override
    public SysPerm getPermById(Long permId) {
        if (permId == null) {
            return null;
        }
        return permMapper.selectById(permId);
    }

    @Override
    public List<SysPerm> listPerms() {
        return permMapper.selectList(null);
    }

    @Override
    public List<SysPerm> listByType(String permType) {
        if (permType == null || permType.trim().isEmpty()) {
            return List.of();
        }
        return permMapper.listByPermType(permType);
    }

    @Override
    public List<SysResource> listResourcesByType(String resourceType) {
        if (resourceType == null || resourceType.trim().isEmpty()) {
            return List.of();
        }
        return resourceMapper.selectList(
            new com.baomidou.mybatisplus.core.conditions.query.QueryWrapper<SysResource>()
                .eq("resource_type", resourceType)
        );
    }

    @Override
    public List<SysMenu> listMenus() {
        return menuMapper.selectList(null);
    }
}
