package io.dataease.audit.service.impl;

import com.baomidou.mybatisplus.core.conditions.query.QueryWrapper;
import io.dataease.audit.constant.AuditConstants;
import io.dataease.audit.dao.auto.mapper.AuditLogMapper;
import io.dataease.audit.dao.auto.mapper.LoginFailureMapper;
import io.dataease.audit.entity.AuditLog;
import io.dataease.audit.entity.LoginFailure;
import io.dataease.audit.service.IAlertService;
import lombok.extern.slf4j.Slf4j;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Service;

import java.time.LocalDateTime;
import java.util.List;

@Service
public class AlertServiceImpl implements IAlertService {

    private static final Logger log = LoggerFactory.getLogger(AlertServiceImpl.class);

    private final AuditLogMapper auditLogMapper;
    private final LoginFailureMapper loginFailureMapper;

    private static final int FAILED_LOGIN_THRESHOLD = 5;
    private static final int FAILED_LOGIN_WINDOW_MINUTES = 5;
    private static final int BATCH_OPERATION_THRESHOLD = 50;
    private static final int BATCH_OPERATION_WINDOW_MINUTES = 60;

    public AlertServiceImpl(AuditLogMapper auditLogMapper, LoginFailureMapper loginFailureMapper) {
        this.auditLogMapper = auditLogMapper;
        this.loginFailureMapper = loginFailureMapper;
    }

    @Override
    public void checkSuspiciousActivity(AuditLog auditLog) {
        if (shouldTriggerAlert(auditLog)) {
            String alertType = determineAlertType(auditLog);
            String message = buildAlertMessage(alertType, auditLog);
            sendAlert(alertType, message, auditLog);
        }
    }

    @Override
    public void detectFailedLoginAttempts(String username, String ipAddress) {
        LocalDateTime cutoff = LocalDateTime.now().minusMinutes(FAILED_LOGIN_WINDOW_MINUTES);

        QueryWrapper<LoginFailure> queryWrapper = new QueryWrapper<>();
        queryWrapper.eq("username", username);
        queryWrapper.eq("ip_address", ipAddress);
        queryWrapper.ge("create_time", cutoff);

        Long failureCount = loginFailureMapper.selectCount(queryWrapper);

        if (failureCount >= FAILED_LOGIN_THRESHOLD) {
            String message = String.format(
                "检测到异常登录尝试：用户 %s 在 %d 分钟内从 IP %s 失败登录 %d 次",
                username,
                FAILED_LOGIN_WINDOW_MINUTES,
                ipAddress,
                failureCount
            );
            log.warn(message);
            sendAlert("FAILED_LOGIN_ALERT", message, null);
        }
    }

    @Override
    public List<AuditLog> getRecentSuspiciousActivity(int minutes) {
        LocalDateTime cutoff = LocalDateTime.now().minusMinutes(minutes);

        QueryWrapper<AuditLog> queryWrapper = new QueryWrapper<>();
        queryWrapper.eq("status", AuditConstants.STATUS_FAILED);
        queryWrapper.ge("create_time", cutoff);
        queryWrapper.orderByDesc("create_time");
        queryWrapper.last("LIMIT 100");

        return auditLogMapper.selectList(queryWrapper);
    }

    @Override
    public boolean shouldTriggerAlert(AuditLog auditLog) {
        return AuditConstants.STATUS_FAILED.equals(auditLog.getStatus()) ||
               isBatchOperation(auditLog) ||
               isPermissionChange(auditLog);
    }

    @Override
    public void sendAlert(String alertType, String message, AuditLog auditLog) {
        log.warn("Audit Alert [{}]: {}", alertType, message);

        if (auditLog != null) {
            log.info("Alert details - User: {}, Action: {}, IP: {}",
                     auditLog.getUsername(),
                     auditLog.getActionName(),
                     auditLog.getIpAddress());
        }
    }

    private String determineAlertType(AuditLog auditLog) {
        if (AuditConstants.STATUS_FAILED.equals(auditLog.getStatus())) {
            return "FAILED_OPERATION";
        }

        if (AuditConstants.ACTION_TYPE_PERMISSION_CHANGE.equals(auditLog.getActionType())) {
            return "PERMISSION_CHANGE";
        }

        if (isBatchOperation(auditLog)) {
            return "BATCH_OPERATION";
        }

        return "SUSPICIOUS_ACTIVITY";
    }

    private String buildAlertMessage(String alertType, AuditLog auditLog) {
        return switch (alertType) {
            case "FAILED_OPERATION" -> String.format(
                "操作失败：用户 %s 执行 %s 操作时失败，原因：%s",
                auditLog.getUsername(),
                auditLog.getActionName(),
                auditLog.getFailureReason()
            );
            case "PERMISSION_CHANGE" -> String.format(
                "权限变更：用户 %s 修改了 %s 的权限配置",
                auditLog.getUsername(),
                auditLog.getResourceType()
            );
            case "BATCH_OPERATION" -> String.format(
                "批量操作：用户 %s 在短时间内执行了大量 %s 操作",
                auditLog.getUsername(),
                auditLog.getActionType()
            );
            default -> String.format(
                "可疑活动：用户 %s 从 IP %s 执行了 %s 操作",
                auditLog.getUsername(),
                auditLog.getIpAddress(),
                auditLog.getActionName()
            );
        };
    }

    private boolean isBatchOperation(AuditLog auditLog) {
        LocalDateTime cutoff = LocalDateTime.now().minusMinutes(BATCH_OPERATION_WINDOW_MINUTES);

        QueryWrapper<AuditLog> queryWrapper = new QueryWrapper<>();
        queryWrapper.eq("user_id", auditLog.getUserId());
        queryWrapper.eq("action_type", auditLog.getActionType());
        queryWrapper.ge("create_time", cutoff);

        Long count = auditLogMapper.selectCount(queryWrapper);
        return count >= BATCH_OPERATION_THRESHOLD;
    }

    private boolean isPermissionChange(AuditLog auditLog) {
        return AuditConstants.ACTION_TYPE_PERMISSION_CHANGE.equals(auditLog.getActionType());
    }
}
