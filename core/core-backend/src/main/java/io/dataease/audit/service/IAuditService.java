package io.dataease.audit.service;

import com.baomidou.mybatisplus.core.metadata.IPage;
import io.dataease.audit.entity.AuditLog;
import io.dataease.audit.entity.LoginFailure;

import java.time.LocalDateTime;
import java.util.List;

public interface IAuditService {

    AuditLog createAuditLog(AuditLog auditLog);

    void createAuditLog(Long userId, String username, String actionType, String actionName, 
                       String resourceType, Long resourceId, String resourceName, 
                       String operation, String status, String failureReason, 
                       String ipAddress, String userAgent, Long organizationId,
                       String beforeValue, String afterValue);

    IPage<AuditLog> queryAuditLogs(Long userId, String username, String actionType, String resourceType,
                                        Long organizationId, LocalDateTime startTime, LocalDateTime endTime,
                                        Integer page, Integer pageSize);

    List<AuditLog> getAuditLogsByUser(Long userId);

    AuditLog getAuditLogById(Long id);

    void recordLoginFailure(String username, String ipAddress, String userAgent, String failureReason);

    void deleteAuditLogsBeforeDate(LocalDateTime date);

    void exportAuditLogs(List<Long> auditLogIds, String format);
}
