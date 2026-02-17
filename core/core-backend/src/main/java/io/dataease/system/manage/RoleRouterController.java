package io.dataease.system.manage;

import org.springframework.web.bind.annotation.*;

import java.util.*;

@RestController
@RequestMapping("/roleRouter")
public class RoleRouterController {

    @GetMapping("/query")
    public List<Map<String, Object>> getRoleRouters() {
        List<Map<String, Object>> routers = new ArrayList<>();

        Map<String, Object> systemRoute = new HashMap<>();
        systemRoute.put("path", "/system");
        systemRoute.put("name", "system");
        systemRoute.put("redirect", "/system/user");
        systemRoute.put("meta", new HashMap<String, Object>() {{
            put("title", "系统管理");
        }});

        List<Map<String, Object>> children = new ArrayList<>();

        Map<String, Object> userRoute = new HashMap<>();
        userRoute.put("path", "user");
        userRoute.put("name", "system-user");
        userRoute.put("component", "system/user");
        userRoute.put("meta", new HashMap<String, Object>() {{
            put("title", "用户管理");
            put("icon", "peoples");
        }});

        Map<String, Object> roleRoute = new HashMap<>();
        roleRoute.put("path", "role");
        roleRoute.put("name", "system-role");
        roleRoute.put("component", "system/role");
        roleRoute.put("meta", new HashMap<String, Object>() {{
            put("title", "角色管理");
            put("icon", "auth");
        }});

        Map<String, Object> orgRoute = new HashMap<>();
        orgRoute.put("path", "org");
        orgRoute.put("name", "system-org");
        orgRoute.put("component", "system/org");
        orgRoute.put("meta", new HashMap<String, Object>() {{
            put("title", "组织管理");
            put("icon", "org");
        }});

        Map<String, Object> permRoute = new HashMap<>();
        permRoute.put("path", "permission");
        permRoute.put("name", "system-permission");
        permRoute.put("component", "system/permission");
        permRoute.put("meta", new HashMap<String, Object>() {{
            put("title", "权限管理");
            put("icon", "icon_security");
        }});

        children.add(userRoute);
        children.add(roleRoute);
        children.add(orgRoute);
        children.add(permRoute);

        systemRoute.put("children", children);
        routers.add(systemRoute);

        return routers;
    }
}
