package io.dataease.system.manage;

import io.dataease.system.entity.SysOrg;
import io.dataease.system.service.IOrgService;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

import java.util.HashMap;
import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/api/system/organization")
public class OrgController {

    @Autowired
    private IOrgService orgService;

    @PostMapping("/create")
    public Map<String, Object> createOrg(@RequestBody SysOrg org) {
        Map<String, Object> result = new HashMap<>();
        try {
            orgService.createOrg(org);
            result.put("msg", "success");
        } catch (Exception e) {
            result.put("msg", "Failed: " + e.getMessage());
        }
        return result;
    }

    @PostMapping("/update")
    public Map<String, Object> updateOrg(@RequestBody SysOrg org) {
        Map<String, Object> result = new HashMap<>();
        try {
            orgService.updateOrg(org);
            result.put("msg", "success");
        } catch (Exception e) {
            result.put("msg", "Failed: " + e.getMessage());
        }
        return result;
    }

    @PostMapping("/delete/{orgId}")
    public Map<String, Object> deleteOrg(@PathVariable Long orgId) {
        Map<String, Object> result = new HashMap<>();
        try {
            orgService.deleteOrg(orgId);
            result.put("msg", "success");
        } catch (Exception e) {
            result.put("msg", "Failed: " + e.getMessage());
        }
        return result;
    }

    @GetMapping("/list")
    public Map<String, Object> listOrgs() {
        Map<String, Object> result = new HashMap<>();
        try {
            List<SysOrg> orgs = orgService.listOrgs();
            result.put("data", orgs);
            result.put("msg", "success");
        } catch (Exception e) {
            result.put("msg", "Failed: " + e.getMessage());
        }
        return result;
    }

    @GetMapping("/info/{orgId}")
    public Map<String, Object> getOrgById(@PathVariable Long orgId) {
        Map<String, Object> result = new HashMap<>();
        try {
            SysOrg org = orgService.getOrgById(orgId);
            result.put("data", org);
            result.put("msg", "success");
        } catch (Exception e) {
            result.put("msg", "Failed: " + e.getMessage());
        }
        return result;
    }

    @GetMapping("/tree")
    public Map<String, Object> getOrgTree() {
        Map<String, Object> result = new HashMap<>();
        try {
            List<SysOrg> topOrgs = orgService.listByParentId(0L);
            result.put("data", topOrgs);
            result.put("msg", "success");
        } catch (Exception e) {
            result.put("msg", "Failed: " + e.getMessage());
        }
        return result;
    }

    @GetMapping("/checkName")
    public Map<String, Object> checkOrgName(@RequestParam String orgName) {
        Map<String, Object> result = new HashMap<>();
        try {
            Integer count = orgService.checkOrgNameExists(orgName);
            result.put("exists", count != null && count > 0);
            result.put("msg", "success");
        } catch (Exception e) {
            result.put("msg", "Failed: " + e.getMessage());
        }
        return result;
    }

    @PostMapping("/updateStatus")
    public Map<String, Object> updateOrgStatus(@RequestBody Map<String, Object> params) {
        Map<String, Object> result = new HashMap<>();
        try {
            Long orgId = Long.parseLong(params.get("orgId").toString());
            Integer status = Integer.parseInt(params.get("status").toString());
            orgService.updateOrgStatus(orgId, status);
            result.put("msg", "success");
        } catch (Exception e) {
            result.put("msg", "Failed: " + e.getMessage());
        }
        return result;
    }

    @GetMapping("/children/{parentId}")
    public Map<String, Object> getChildOrgs(@PathVariable Long parentId) {
        Map<String, Object> result = new HashMap<>();
        try {
            List<SysOrg> children = orgService.listByParentId(parentId);
            result.put("data", children);
            result.put("msg", "success");
        } catch (Exception e) {
            result.put("msg", "Failed: " + e.getMessage());
        }
        return result;
    }
}
