package io.dataease.audit.server;

import com.baomidou.mybatisplus.core.metadata.IPage;
import io.dataease.audit.entity.AuditLog;
import io.dataease.audit.entity.LoginFailure;
import io.dataease.audit.service.IAuditService;
import io.dataease.result.ResultMessage;
import io.dataease.utils.UserUtils;
import jakarta.servlet.http.HttpServletRequest;
import lombok.extern.slf4j.Slf4j;
import org.springframework.web.bind.annotation.*;

import java.time.LocalDateTime;
import java.util.List;

@RestController
@RequestMapping("/api/audit")
@Slf4j
public class AuditController {

    private final IAuditService auditService;

    public AuditController(IAuditService auditService) {
        this.auditService = auditService;
    }

    @PostMapping("/log")
    public ResultMessage createAuditLog(@RequestBody AuditLog auditLog, HttpServletRequest request) {
        String ipAddress = getClientIpAddress(request);
        auditLog.setIpAddress(ipAddress);
        
        AuditLog created = auditService.createAuditLog(auditLog);
        return ResultMessage.success(created);
    }

    @GetMapping("/list")
    public ResultMessage queryAuditLogs(
            @RequestParam(required = false) Long userId,
            @RequestParam(required = false) String username,
            @RequestParam(required = false) String actionType,
            @RequestParam(required = false) String resourceType,
            @RequestParam(required = false) Long organizationId,
            @RequestParam(required = false) LocalDateTime startTime,
            @RequestParam(required = false) LocalDateTime endTime,
            @RequestParam(required = false) Integer page,
            @RequestParam(required = false) Integer pageSize) {
        
        IPage<AuditLog> result = auditService.queryAuditLogs(userId, username, actionType, resourceType,
                                                            organizationId, startTime, endTime, page, pageSize);
        return ResultMessage.success(result);
    }

    @GetMapping("/user/{userId}")
    public ResultMessage getAuditLogsByUser(@PathVariable Long userId) {
        List<AuditLog> result = auditService.getAuditLogsByUser(userId);
        return ResultMessage.success(result);
    }

    @GetMapping("/{id}")
    public ResultMessage getAuditLogById(@PathVariable Long id) {
        AuditLog result = auditService.getAuditLogById(id);
        return ResultMessage.success(result);
    }

    @PostMapping("/export")
    public ResultMessage exportAuditLogs(@RequestBody List<Long> auditLogIds, 
                                     @RequestParam(defaultValue = "csv") String format) {
        auditService.exportAuditLogs(auditLogIds, format);
        return ResultMessage.success("Export completed");
    }

    @DeleteMapping("/retention")
    public ResultMessage deleteOldLogs(@RequestParam(defaultValue = "90") Integer days) {
        LocalDateTime cutoffDate = LocalDateTime.now().minusDays(days);
        auditService.deleteAuditLogsBeforeDate(cutoffDate);
        return ResultMessage.success("Deleted logs older than " + days + " days");
    }

    private String getClientIpAddress(HttpServletRequest request) {
        String ipAddress = request.getHeader("X-Forwarded-For");
        if (ipAddress == null || ipAddress.isEmpty() || ipAddress.equals("unknown")) {
            ipAddress = request.getRemoteAddr();
        }
        return ipAddress;
    }
}
