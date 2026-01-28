package io.dataease.audit.service.impl;

import com.baomidou.mybatisplus.core.conditions.query.QueryWrapper;
import com.baomidou.mybatisplus.core.metadata.IPage;
import com.baomidou.mybatisplus.extension.plugins.pagination.Page;
import com.baomidou.mybatisplus.core.conditions.query.QueryWrapper;
import com.baomidou.mybatisplus.core.metadata.IPage;
import com.baomidou.mybatisplus.extension.plugins.pagination.Page;
import io.dataease.audit.constant.AuditConstants;
import io.dataease.audit.dao.auto.mapper.AuditLogMapper;
import io.dataease.audit.dao.auto.mapper.LoginFailureMapper;
import io.dataease.audit.entity.AuditLog;
import io.dataease.audit.entity.LoginFailure;
import io.dataease.audit.service.IAuditService;
import io.dataease.auth.bo.TokenUserBO;
import io.dataease.utils.AuthUtils;

import com.baomidou.mybatisplus.core.conditions.query.QueryWrapper;
import com.baomidou.mybatisplus.core.metadata.IPage;
import com.baomidou.mybatisplus.extension.plugins.pagination.Page;
import io.dataease.audit.constant.AuditConstants;
import io.dataease.audit.dao.auto.mapper.AuditLogMapper;
import io.dataease.audit.dao.auto.mapper.LoginFailureMapper;
import io.dataease.audit.entity.AuditLog;
import io.dataease.audit.entity.LoginFailure;
import io.dataease.audit.service.IAuditService;
import io.dataease.auth.bo.TokenUserBO;
import io.dataease.utils.AuthUtils;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.time.LocalDateTime;
import java.util.List;

@Service
@Slf4j
public class AuditServiceImpl implements IAuditService {

    private final AuditLogMapper auditLogMapper;
    private final LoginFailureMapper loginFailureMapper;

    public AuditServiceImpl(AuditLogMapper auditLogMapper, LoginFailureMapper loginFailureMapper) {
        this.auditLogMapper = auditLogMapper;
        this.loginFailureMapper = loginFailureMapper;
    }

    @Override
    @Transactional(rollbackFor = Exception.class)
    public AuditLog createAuditLog(AuditLog auditLog) {
        if (auditLog.getCreateTime() == null) {
            auditLog.setCreateTime(LocalDateTime.now());
        }
        if (auditLog.getStatus() == null) {
            auditLog.setStatus(AuditConstants.STATUS_SUCCESS);
        }
        if (auditLog.getIpAddress() == null) {
            auditLog.setIpAddress("127.0.0.1");
        }
        if (auditLog.getUserAgent() == null) {
            auditLog.setUserAgent("");
        }

        auditLogMapper.insert(auditLog);
        log.info("Audit log created: {}", auditLog);
        return auditLog;
    }

    @Override
    public void createAuditLog(Long userId, String username, String actionType, String actionName,
                              String resourceType, Long resourceId, String resourceName,
                              String operation, String status, String failureReason,
                              String ipAddress, String userAgent, Long organizationId,
                              String beforeValue, String afterValue) {
        AuditLog auditLog = new AuditLog();
        auditLog.setUserId(userId);
        auditLog.setUsername(username);
        auditLog.setActionType(actionType);
        auditLog.setActionName(actionName);
        auditLog.setResourceType(resourceType);
        auditLog.setResourceId(resourceId);
        auditLog.setResourceName(resourceName);
        auditLog.setOperation(operation);
        auditLog.setStatus(status);
        auditLog.setFailureReason(failureReason);
        auditLog.setIpAddress(ipAddress);
        auditLog.setUserAgent(userAgent);
        auditLog.setOrganizationId(organizationId);
        auditLog.setBeforeValue(beforeValue);
        auditLog.setAfterValue(afterValue);
        auditLog.setCreateTime(LocalDateTime.now());

        createAuditLog(auditLog);
    }

    @Override
    public IPage<AuditLog> queryAuditLogs(Long userId, String username, String actionType, String resourceType,
                                            Long organizationId, LocalDateTime startTime, LocalDateTime endTime,
                                            Integer page, Integer pageSize) {
        QueryWrapper<AuditLog> queryWrapper = new QueryWrapper<>();

        if (userId != null) {
            queryWrapper.eq("user_id", userId);
        }
        if (username != null && !username.isEmpty()) {
            queryWrapper.like("username", "%" + username + "%");
        }
        if (actionType != null && !actionType.isEmpty()) {
            queryWrapper.eq("action_type", actionType);
        }
        if (resourceType != null && !resourceType.isEmpty()) {
            queryWrapper.eq("resource_type", resourceType);
        }
        if (organizationId != null) {
            queryWrapper.eq("organization_id", organizationId);
        }
        if (startTime != null) {
            queryWrapper.ge("create_time", startTime);
        }
        if (endTime != null) {
            queryWrapper.le("create_time", endTime);
        }

        queryWrapper.orderByDesc("create_time");

        Page<AuditLog> pageParam = new Page<>(page != null ? page : 1, pageSize != null ? pageSize : 10);
        return auditLogMapper.selectPage(pageParam, queryWrapper);
    }

    @Override
    public List<AuditLog> getAuditLogsByUser(Long userId) {
        QueryWrapper<AuditLog> queryWrapper = new QueryWrapper<>();
        queryWrapper.eq("user_id", userId);
        queryWrapper.orderByDesc("create_time");
        queryWrapper.last("SELECT 10");
        return auditLogMapper.selectList(queryWrapper);
    }

    @Override
    public AuditLog getAuditLogById(Long id) {
        return auditLogMapper.selectById(id);
    }

    @Override
    @Transactional(rollbackFor = Exception.class)
    public void recordLoginFailure(String username, String ipAddress, String userAgent, String failureReason) {
        LoginFailure loginFailure = new LoginFailure();
        loginFailure.setUsername(username);
        loginFailure.setIpAddress(ipAddress);
        loginFailure.setFailureReason(failureReason);
        loginFailure.setUserAgent(userAgent);
        loginFailure.setCreateTime(LocalDateTime.now());

        loginFailureMapper.insert(loginFailure);
        log.warn("Login failure recorded: username={}, ip={}, reason={}", username, ipAddress, failureReason);
    }

    @Override
    @Transactional(rollbackFor = Exception.class)
    public void deleteAuditLogsBeforeDate(LocalDateTime date) {
        QueryWrapper<AuditLog> queryWrapper = new QueryWrapper<>();
        queryWrapper.lt("create_time", date);
        List<AuditLog> logsToDelete = auditLogMapper.selectList(queryWrapper);
        
        if (!logsToDelete.isEmpty()) {
            for (AuditLog log : logsToDelete) {
                auditLogMapper.deleteById(log.getId());
            }
            log.info("Deleted {} audit logs older than {}", logsToDelete.size(), date);
        }
    }

    @Override
    @Transactional(rollbackFor = Exception.class)
    public void exportAuditLogs(List<Long> auditLogIds, String format) {
        QueryWrapper<AuditLog> queryWrapper = new QueryWrapper<>();
        queryWrapper.in("id", auditLogIds);
        List<AuditLog> auditLogs = auditLogMapper.selectList(queryWrapper);
        
        log.info("Exporting {} audit logs in {} format", auditLogs.size(), format);
    }
}
