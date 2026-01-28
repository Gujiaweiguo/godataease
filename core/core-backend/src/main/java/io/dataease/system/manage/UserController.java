package io.dataease.system.manage;

import io.dataease.system.entity.SysUser;
import io.dataease.system.service.ISysUserService;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

import java.util.HashMap;
import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/api/system/user")
public class UserController {

    @Autowired
    private io.dataease.system.service.ISysUserService userService;

    @PostMapping("/list")
    public Map<String, Object> listUsers(@RequestBody Map<String, Object> params) {
        Map<String, Object> result = new HashMap<>();
        try {
            Integer current = params.containsKey("current") ? Integer.parseInt(params.get("current").toString()) : 1;
            Integer size = params.containsKey("size") ? Integer.parseInt(params.get("size").toString()) : 10;
            String keyword = params.containsKey("keyword") ? params.get("keyword").toString() : null;
            Long orgId = params.containsKey("orgId") ? Long.parseLong(params.get("orgId").toString()) : null;
            Integer status = params.containsKey("status") ? Integer.parseInt(params.get("status").toString()) : null;

            List<SysUser> users = userService.searchUsers(keyword, orgId, status);

            int start = (current - 1) * size;
            int end = Math.min(start + size, users.size());
            List<SysUser> pageData = users.subList(start, end);

            Map<String, Object> data = new HashMap<>();
            data.put("list", pageData);
            data.put("total", users.size());
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
    public Map<String, Object> createUser(@RequestBody Map<String, Object> userMap) {
        Map<String, Object> result = new HashMap<>();
        try {
            SysUser user = new SysUser();
            user.setUsername((String) userMap.get("username"));
            user.setNickName((String) userMap.get("realName"));
            user.setEmail((String) userMap.get("email"));
            user.setPhone(userMap.get("phone") != null ? userMap.get("phone").toString() : null);
            user.setPassword((String) userMap.get("password"));
            user.setFrom(0);
            user.setStatus(userMap.get("status") != null ? Integer.parseInt(userMap.get("status").toString()) : 1);
            user.setDelFlag(0);

            Long userId = userService.createUser(user);
            result.put("code", "000000");
            result.put("data", userId);
            result.put("msg", "success");
        } catch (Exception e) {
            result.put("code", "500000");
            result.put("msg", "Failed: " + e.getMessage());
        }
        return result;
    }

    @PostMapping("/update")
    public Map<String, Object> updateUser(@RequestBody Map<String, Object> userMap) {
        Map<String, Object> result = new HashMap<>();
        try {
            SysUser user = new SysUser();
            user.setUserId(Long.parseLong(userMap.get("id").toString()));
            user.setUsername((String) userMap.get("username"));
            user.setNickName((String) userMap.get("realName"));
            user.setEmail((String) userMap.get("email"));
            user.setPhone(userMap.get("phone") != null ? userMap.get("phone").toString() : null);

            if (userMap.get("password") != null && !((String) userMap.get("password")).isEmpty()) {
                user.setPassword((String) userMap.get("password"));
            }

            user.setStatus(userMap.get("status") != null ? Integer.parseInt(userMap.get("status").toString()) : 1);

            userService.updateUser(user);
            result.put("code", "000000");
            result.put("msg", "success");
        } catch (Exception e) {
            result.put("code", "500000");
            result.put("msg", "Failed: " + e.getMessage());
        }
        return result;
    }

    @PostMapping("/delete/{id}")
    public Map<String, Object> deleteUser(@PathVariable Long id) {
        Map<String, Object> result = new HashMap<>();
        try {
            userService.deleteUser(id);
            result.put("code", "000000");
            result.put("msg", "success");
        } catch (Exception e) {
            result.put("code", "500000");
            result.put("msg", "Failed: " + e.getMessage());
        }
        return result;
    }

    @GetMapping("/options")
    public Map<String, Object> getUserOptions() {
        Map<String, Object> result = new HashMap<>();
        try {
            List<SysUser> users = userService.searchUsers(null, null, null);
            result.put("code", "000000");
            result.put("data", users);
            result.put("msg", "success");
        } catch (Exception e) {
            result.put("code", "500000");
            result.put("msg", "Failed: " + e.getMessage());
        }
        return result;
    }
}
