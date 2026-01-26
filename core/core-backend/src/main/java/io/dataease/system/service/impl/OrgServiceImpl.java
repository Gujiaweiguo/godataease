package io.dataease.system.service.impl;

import io.dataease.system.dao.auto.mapper.SysOrgMapper;
import io.dataease.system.entity.SysOrg;
import io.dataease.system.service.IOrgService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.List;

@Service
public class OrgServiceImpl implements IOrgService {

    @Autowired
    private SysOrgMapper orgMapper;

    @Override
    @Transactional(rollbackFor = Exception.class)
    public void createOrg(SysOrg org) {
        org.setCreateTime(java.time.LocalDateTime.now());
        org.setUpdateTime(java.time.LocalDateTime.now());
        org.setLevel(1);
        org.setStatus(1);
        org.setDelFlag(0);
        orgMapper.insert(org);
    }

    @Override
    @Transactional(rollbackFor = Exception.class)
    public void updateOrg(SysOrg org) {
        if (org.getOrgId() == null) {
            return;
        }
        org.setUpdateTime(java.time.LocalDateTime.now());
        orgMapper.updateById(org);
    }

    @Override
    @Transactional(rollbackFor = Exception.class)
    public void deleteOrg(Long orgId) {
        if (orgId == null) {
            return;
        }
        orgMapper.deleteOrg(orgId);
    }

    @Override
    public SysOrg getOrgById(Long orgId) {
        if (orgId == null) {
            return null;
        }
        return orgMapper.getOrgById(orgId);
    }

    @Override
    public List<SysOrg> listOrgs() {
        return orgMapper.listOrgs();
    }

    @Override
    public List<SysOrg> listByParentId(Long parentId) {
        if (parentId == null) {
            parentId = 0L;
        }
        return orgMapper.listByParentId(parentId);
    }

    @Override
    public Integer checkOrgNameExists(String orgName) {
        if (orgName == null || orgName.trim().isEmpty()) {
            return 0;
        }
        return orgMapper.checkOrgNameExists(orgName);
    }

    @Override
    @Transactional(rollbackFor = Exception.class)
    public void updateOrgStatus(Long orgId, Integer status) {
        if (orgId == null) {
            return;
        }
        orgMapper.updateOrgStatus(orgId, status != null ? status : 1);
    }

    @Override
    public Integer countChildOrgs(Long orgId) {
        if (orgId == null) {
            return 0;
        }
        return orgMapper.countChildOrgs(orgId);
    }
}
