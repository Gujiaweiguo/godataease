package io.dataease.system.manage;

import com.baomidou.mybatisplus.core.conditions.query.QueryWrapper;
import io.dataease.system.dao.auto.mapper.DataPermRowMapper;
import io.dataease.system.entity.DataPermRow;
import io.dataease.system.service.IDataPermRowService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

import java.time.LocalDateTime;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/api/system/rowPermission")
public class RowPermissionController {

    @Autowired
    private IDataPermRowService dataPermRowService;

    @PostMapping("/list")
    public Map<String, Object> listRowPermissions(@RequestBody Map<String, Object> params) {
        Map<String, Object> result = new HashMap<>();
        try {
            Long datasetId = params.containsKey("datasetId") ? Long.parseLong(params.get("datasetId").toString()) : null;
            Integer current = params.containsKey("current") ? Integer.parseInt(params.get("current").toString()) : 1;
            Integer size = params.containsKey("size") ? Integer.parseInt(params.get("size").toString()) : 10;

            QueryWrapper<DataPermRow> queryWrapper = new QueryWrapper<>();
            queryWrapper.eq("status", 1);
            queryWrapper.orderByDesc("create_time");

            List<DataPermRow> allPerms = dataPermRowService.listByDatasetId(datasetId);

            int start = (current - 1) * size;
            int end = Math.min(start + size, allPerms.size());
            List<DataPermRow> pageData = allPerms.subList(start, end);

            Map<String, Object> data = new HashMap<>();
            data.put("list", pageData);
            data.put("total", allPerms.size());
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
    public Map<String, Object> createRowPermission(@RequestBody Map<String, Object> permMap) {
        Map<String, Object> result = new HashMap<>();
        try {
            DataPermRow dataPermRow = new DataPermRow();
            dataPermRow.setDatasetId(Long.parseLong(permMap.get("datasetId").toString()));
            dataPermRow.setAuthTargetType((String) permMap.get("authTargetType"));
            dataPermRow.setAuthTargetId(Long.parseLong(permMap.get("authTargetId").toString()));
            dataPermRow.setExpressionTree((String) permMap.get("expressionTree"));
            dataPermRow.setStatus(permMap.get("status") != null ? Integer.parseInt(permMap.get("status").toString()) : 1);
            dataPermRow.setCreateBy((String) permMap.get("createBy"));

            Long id = dataPermRowService.create(dataPermRow);
            result.put("code", "000000");
            result.put("data", id);
            result.put("msg", "success");
        } catch (Exception e) {
            result.put("code", "500000");
            result.put("msg", "Failed: " + e.getMessage());
        }
        return result;
    }

    @PostMapping("/update")
    public Map<String, Object> updateRowPermission(@RequestBody Map<String, Object> permMap) {
        Map<String, Object> result = new HashMap<>();
        try {
            DataPermRow dataPermRow = new DataPermRow();
            dataPermRow.setId(Long.parseLong(permMap.get("id").toString()));
            dataPermRow.setDatasetId(Long.parseLong(permMap.get("datasetId").toString()));
            dataPermRow.setAuthTargetType((String) permMap.get("authTargetType"));
            dataPermRow.setAuthTargetId(Long.parseLong(permMap.get("authTargetId").toString()));
            dataPermRow.setExpressionTree((String) permMap.get("expressionTree"));
            if (permMap.get("status") != null) {
                dataPermRow.setStatus(Integer.parseInt(permMap.get("status").toString()));
            }

            dataPermRowService.update(dataPermRow);
            result.put("code", "000000");
            result.put("msg", "success");
        } catch (Exception e) {
            result.put("code", "500000");
            result.put("msg", "Failed: " + e.getMessage());
        }
        return result;
    }

    @PostMapping("/delete/{id}")
    public Map<String, Object> deleteRowPermission(@PathVariable Long id) {
        Map<String, Object> result = new HashMap<>();
        try {
            dataPermRowService.delete(id);
            result.put("code", "000000");
            result.put("msg", "success");
        } catch (Exception e) {
            result.put("code", "500000");
            result.put("msg", "Failed: " + e.getMessage());
        }
        return result;
    }

    @PostMapping("/updateStatus/{id}")
    public Map<String, Object> updateStatus(@PathVariable Long id, @RequestBody Map<String, Object> params) {
        Map<String, Object> result = new HashMap<>();
        try {
            Integer status = params.containsKey("status") ? Integer.parseInt(params.get("status").toString()) : 1;
            dataPermRowService.updateStatus(id, status);
            result.put("code", "000000");
            result.put("msg", "success");
        } catch (Exception e) {
            result.put("code", "500000");
            result.put("msg", "Failed: " + e.getMessage());
        }
        return result;
    }
}
