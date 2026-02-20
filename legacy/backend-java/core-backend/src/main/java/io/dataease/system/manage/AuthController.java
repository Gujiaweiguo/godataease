package io.dataease.system.manage;

import io.dataease.system.dao.auto.mapper.SysPermMapper;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/auth")
public class AuthController {

    @Autowired
    private SysPermMapper permMapper;

    @GetMapping("/menuResource")
    public Map<String, Object> getMenuResource() {
        Map<String, Object> result = new HashMap<>();
        try {
            List<Map<String, Object>> menuTree = new ArrayList<>();

            Map<String, Object> userMenu = new HashMap<>();
            userMenu.put("path", "user");
            userMenu.put("meta", new HashMap<String, Object>() {{
                put("title", "用户管理");
                put("icon", "peoples");
            }});

            Map<String, Object> roleMenu = new HashMap<>();
            roleMenu.put("path", "role");
            roleMenu.put("meta", new HashMap<String, Object>() {{
                put("title", "角色管理");
                put("icon", "auth");
            }});

            Map<String, Object> orgMenu = new HashMap<>();
            orgMenu.put("path", "org");
            orgMenu.put("meta", new HashMap<String, Object>() {{
                put("title", "组织管理");
                put("icon", "org");
            }});

            Map<String, Object> permMenu = new HashMap<>();
            permMenu.put("path", "permission");
            permMenu.put("meta", new HashMap<String, Object>() {{
                put("title", "权限管理");
                put("icon", "icon_security");
            }});

            menuTree.add(userMenu);
            menuTree.add(roleMenu);
            menuTree.add(orgMenu);
            menuTree.add(permMenu);

            result.put("code", "000000");
            result.put("data", menuTree);
            result.put("msg", "success");
        } catch (Exception e) {
            result.put("code", "500000");
            result.put("msg", "Failed: " + e.getMessage());
        }
        return result;
    }

    @GetMapping("/busiResource/{flag}")
    public Map<String, Object> getBusiResource(@PathVariable String flag) {
        Map<String, Object> result = new HashMap<>();
        try {
            List<Map<String, Object>> resourceTree = new ArrayList<>();

            result.put("code", "000000");
            result.put("data", resourceTree);
            result.put("msg", "success");
        } catch (Exception e) {
            result.put("code", "500000");
            result.put("msg", "Failed: " + e.getMessage());
        }
        return result;
    }

    @PostMapping("/menuPermission")
    public Map<String, Object> getMenuPermission(@RequestBody Map<String, Object> params) {
        Map<String, Object> result = new HashMap<>();
        try {
            Long roleId = Long.parseLong(params.get("roleId").toString());

            List<Long> permissions = new ArrayList<>();

            result.put("code", "000000");
            result.put("data", permissions);
            result.put("msg", "success");
        } catch (Exception e) {
            result.put("code", "500000");
            result.put("msg", "Failed: " + e.getMessage());
        }
        return result;
    }

    @PostMapping("/saveMenuPer")
    public Map<String, Object> saveMenuPermission(@RequestBody Map<String, Object> params) {
        Map<String, Object> result = new HashMap<>();
        try {
            Long roleId = Long.parseLong(params.get("roleId").toString());
            List<Long> permIds = (List<Long>) params.get("permIds");

            result.put("code", "000000");
            result.put("msg", "success");
        } catch (Exception e) {
            result.put("code", "500000");
            result.put("msg", "Failed: " + e.getMessage());
        }
        return result;
    }

    @PostMapping("/saveBusiPer")
    public Map<String, Object> saveBusiPermission(@RequestBody Map<String, Object> params) {
        Map<String, Object> result = new HashMap<>();
        try {
            Long roleId = Long.parseLong(params.get("roleId").toString());
            List<Long> permIds = (List<Long>) params.get("permIds");

            result.put("code", "000000");
            result.put("msg", "success");
        } catch (Exception e) {
            result.put("code", "500000");
            result.put("msg", "Failed: " + e.getMessage());
        }
        return result;
    }

    @PostMapping("/menuTargetPermission")
    public Map<String, Object> getMenuTargetPermission(@RequestBody Map<String, Object> params) {
        Map<String, Object> result = new HashMap<>();
        try {
            Long roleId = Long.parseLong(params.get("roleId").toString());

            List<Map<String, Object>> targetPermissions = new ArrayList<>();

            result.put("code", "000000");
            result.put("data", targetPermissions);
            result.put("msg", "success");
        } catch (Exception e) {
            result.put("code", "500000");
            result.put("msg", "Failed: " + e.getMessage());
        }
        return result;
    }

    @PostMapping("/saveMenuTargetPer")
    public Map<String, Object> saveMenuTargetPermission(@RequestBody Map<String, Object> params) {
        Map<String, Object> result = new HashMap<>();
        try {
            Long roleId = Long.parseLong(params.get("roleId").toString());
            List<Map<String, Object>> targetPerms = (List<Map<String, Object>>) params.get("targetPerms");

            result.put("code", "000000");
            result.put("msg", "success");
        } catch (Exception e) {
            result.put("code", "500000");
            result.put("msg", "Failed: " + e.getMessage());
        }
        return result;
    }

    @PostMapping("/busiPermission")
    public Map<String, Object> getBusiPermission(@RequestBody Map<String, Object> params) {
        Map<String, Object> result = new HashMap<>();
        try {
            Long roleId = Long.parseLong(params.get("roleId").toString());

            List<Long> permissions = new ArrayList<>();

            result.put("code", "000000");
            result.put("data", permissions);
            result.put("msg", "success");
        } catch (Exception e) {
            result.put("code", "500000");
            result.put("msg", "Failed: " + e.getMessage());
        }
        return result;
    }

    @PostMapping("/busiTargetPermission")
    public Map<String, Object> getBusiTargetPermission(@RequestBody Map<String, Object> params) {
        Map<String, Object> result = new HashMap<>();
        try {
            Long roleId = Long.parseLong(params.get("roleId").toString());

            List<Map<String, Object>> targetPermissions = new ArrayList<>();

            result.put("code", "000000");
            result.put("data", targetPermissions);
            result.put("msg", "success");
        } catch (Exception e) {
            result.put("code", "500000");
            result.put("msg", "Failed: " + e.getMessage());
        }
        return result;
    }

    @PostMapping("/saveBusiTargetPer")
    public Map<String, Object> saveBusiTargetPermission(@RequestBody Map<String, Object> params) {
        Map<String, Object> result = new HashMap<>();
        try {
            Long roleId = Long.parseLong(params.get("roleId").toString());
            List<Map<String, Object>> targetPerms = (List<Map<String, Object>>) params.get("targetPerms");

            result.put("code", "000000");
            result.put("msg", "success");
        } catch (Exception e) {
            result.put("code", "500000");
            result.put("msg", "Failed: " + e.getMessage());
        }
        return result;
    }
}
