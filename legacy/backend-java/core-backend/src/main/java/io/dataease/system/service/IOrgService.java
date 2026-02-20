package io.dataease.system.service;

import io.dataease.system.entity.SysOrg;
import java.util.List;

public interface IOrgService {

    void createOrg(SysOrg org);

    void updateOrg(SysOrg org);

    void deleteOrg(Long orgId);

    SysOrg getOrgById(Long orgId);

    List<SysOrg> listOrgs();

    List<SysOrg> listByParentId(Long parentId);

    Integer checkOrgNameExists(String orgName);

    Integer countChildOrgs(Long parentId);

    void updateOrgStatus(Long orgId, Integer status);
}
