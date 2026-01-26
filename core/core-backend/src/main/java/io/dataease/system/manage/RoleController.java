package io.dataease.system.manage;

import io.dataease.system.entity.SysRole;
import io.dataease.system.service.IRoleService;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

import java.util.HashMap;
import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/api/system/role")
public class RoleController {

    @Autowired
    private IRoleService roleService;

    @PostMapping("/list")
    public Map<String, Object> listRoles(@RequestBody Map<String, Object> params) {
        Map<String, Object> result = new HashMap<>();
        try {
            Integer current = params.containsKey("current") ? Integer.parseInt(params.get("current").toString()) :1;
            Integer size = params.containsKey("size") ? Integer.parseInt(params.get("size").toString()) : 10;

            List<SysRole> roles = roleService.listRoles();

            int start = (current -1) * size;
            int end = Math.min(start + size, roles.size());
            List<SysRole> pageData = roles.subList(start, end);

            Map<String, Object> data = new HashMap<>();
            data.put("list", pageData);
            data.put("total", roles.size());
            data.put("current", current);
            data.put("size", size);

            result.put("code", "000000");
            result.put("data", data);
            result.put("msg", "success");
        } catch (Exception e) {
            result.put("code", "500000");
            result.put("msg", "Failed: " + e.getMessage());
        }
        return result;
    }

    @PostMapping("/create")
    public Map<String, Object> createRole(@RequestBody Map<String, Object> roleMap) {
        Map<String, Object> result = new HashMap<>();
        try {
            SysRole role = new SysRole();
            role.setRoleName((String) roleMap.get("roleName"));
            role.setRoleCode((String) roleMap.get("roleKey"));
            role.setRoleDesc((String) roleMap.get("roleDesc"));
            role.setStatus(roleMap.get("status") != null ? Integer.parseInt(roleMap.get("status").toString()) :1);
            role.setDataScope("all");
            role.setParentId(0L);
            role.setLevel(1);

            roleService.createRole(role);

            result.put("code", "000000");
            result.put("data", role.getRoleId());
            result.put("msg", "success");
        } catch (Exception e) {
            result.put("code", "500000");
            result.put("msg", "Failed: " + e.getMessage());
        }
        return result;
    }

    @PostMapping("/update")
    public Map<String, Object> updateRole(@RequestBody Map<String, Object> roleMap) {
        Map<String, Object> result = new HashMap<>();
        try {
            SysRole role = new SysRole();
            role.setRoleId(Long.parseLong(roleMap.get("roleId").toString()));
            role.setRoleName((String) roleMap.get("roleName"));
            role.setRoleCode((String) roleMap.get("roleKey"));
            role.setRoleDesc((String) roleMap.get("roleDesc"));
            role.setStatus(roleMap.get("status") != null ? Integer.parseInt(roleMap.get("status").toString()) :1);

            roleService.updateRole(role);

            result.put("code", "000000");
            result.put("msg", "success");
        } catch (Exception e) {
            result.put("code", "500000");
            result.put("msg", "Failed: " + e.getMessage());
        }
        return result;
    }

    @PostMapping("/delete/{roleId}")
    public Map<String, Object> deleteRole(@PathVariable Long roleId) {
        Map<String, Object> result = new HashMap<>();
        try {
            roleService.deleteRole(roleId);

            result.put("code", "000000");
            result.put("msg", "success");
        } catch (Exception e) {
            result.put("code", "500000");
            result.put("msg", "Failed: " + e.getMessage());
        }
        return result;
    }

    @PostMapping("/permission/save")
    public Map<String, Object> saveRolePermissions(@RequestBody Map<String, Object> params) {
        Map<String, Object> result = new HashMap<>();
        try {
            Long roleId = Long.parseLong(params.get("roleId").toString());
            List<Long> permIds = (List<Long>) params.get("permIds");

            roleService.assignPermissionsToRole(roleId, permIds);

            result.put("code", "000000");
            result.put("msg", "success");
        } catch (Exception e) {
            result.put("code", "500000");
            result.put("msg", "Failed: " + e.getMessage());
        }
        return result;
    }
}
