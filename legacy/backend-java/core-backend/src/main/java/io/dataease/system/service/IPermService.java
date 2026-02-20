package io.dataease.system.service;

import io.dataease.system.entity.SysPerm;
import io.dataease.system.entity.SysResource;
import io.dataease.system.entity.SysMenu;

import java.util.List;

public interface IPermService {

    void createPerm(SysPerm perm);

    void updatePerm(SysPerm perm);

    void deletePerm(Long permId);

    SysPerm getPermById(Long permId);

    List<SysPerm> listPerms();

    List<SysPerm> listByType(String permType);

    List<SysResource> listResourcesByType(String resourceType);

    List<SysMenu> listMenus();
}
