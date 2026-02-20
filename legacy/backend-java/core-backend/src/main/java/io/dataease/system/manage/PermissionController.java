package io.dataease.system.manage;

import io.dataease.system.entity.SysPerm;
import io.dataease.system.service.IPermService;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

import java.util.HashMap;
import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/api/system/permission")
public class PermissionController {

    @Autowired
    private IPermService permService;

    @PostMapping("/list")
    public Map<String, Object> listPerms(@RequestBody Map<String, Object> params) {
        Map<String, Object> result = new HashMap<>();
        try {
            Integer current = params.containsKey("current") ? Integer.parseInt(params.get("current").toString()) :1;
            Integer size = params.containsKey("size") ? Integer.parseInt(params.get("size").toString()) : 10;

            List<SysPerm> perms = permService.listPerms();

            int start = (current -1) * size;
            int end = Math.min(start + size, perms.size());
            List<SysPerm> pageData = perms.subList(start, end);

            Map<String, Object> data = new HashMap<>();
            data.put("list", pageData);
            data.put("total", perms.size());
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
    public Map<String, Object> createPerm(@RequestBody Map<String, Object> permMap) {
        Map<String, Object> result = new HashMap<>();
        try {
            SysPerm perm = new SysPerm();
            perm.setPermName((String) permMap.get("permName"));
            perm.setPermKey((String) permMap.get("permKey"));
            perm.setPermType((String) permMap.get("permType"));
            perm.setPermDesc((String) permMap.get("permDesc"));
            perm.setStatus(permMap.get("status") != null ? Integer.parseInt(permMap.get("status").toString()) :1);

            permService.createPerm(perm);

            result.put("code", "000000");
            result.put("data", perm.getPermId());
            result.put("msg", "success");
        } catch (Exception e) {
            result.put("code", "500000");
            result.put("msg", "Failed: " + e.getMessage());
        }
        return result;
    }

    @PostMapping("/update")
    public Map<String, Object> updatePerm(@RequestBody Map<String, Object> permMap) {
        Map<String, Object> result = new HashMap<>();
        try {
            SysPerm perm = new SysPerm();
            perm.setPermId(Long.parseLong(permMap.get("permId").toString()));
            perm.setPermName((String) permMap.get("permName"));
            perm.setPermKey((String) permMap.get("permKey"));
            perm.setPermType((String) permMap.get("permType"));
            perm.setPermDesc((String) permMap.get("permDesc"));
            perm.setStatus(permMap.get("status") != null ? Integer.parseInt(permMap.get("status").toString()) :1);

            permService.updatePerm(perm);

            result.put("code", "000000");
            result.put("msg", "success");
        } catch (Exception e) {
            result.put("code", "500000");
            result.put("msg", "Failed: " + e.getMessage());
        }
        return result;
    }

    @PostMapping("/delete/{permId}")
    public Map<String, Object> deletePerm(@PathVariable Long permId) {
        Map<String, Object> result = new HashMap<>();
        try {
            permService.deletePerm(permId);

            result.put("code", "000000");
            result.put("msg", "success");
        } catch (Exception e) {
            result.put("code", "500000");
            result.put("msg", "Failed: " + e.getMessage());
        }
        return result;
    }
}
