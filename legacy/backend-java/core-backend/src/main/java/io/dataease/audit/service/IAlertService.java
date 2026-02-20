package io.dataease.audit.service;

import io.dataease.audit.entity.AuditLog;

import java.util.List;

public interface IAlertService {

    void checkSuspiciousActivity(AuditLog auditLog);

    void detectFailedLoginAttempts(String username, String ipAddress);

    List<AuditLog> getRecentSuspiciousActivity(int minutes);

    boolean shouldTriggerAlert(AuditLog auditLog);

    void sendAlert(String alertType, String message, AuditLog auditLog);
}
