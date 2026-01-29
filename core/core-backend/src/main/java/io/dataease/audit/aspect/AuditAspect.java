package io.dataease.audit.aspect;

import io.dataease.audit.constant.AuditConstants;
import io.dataease.audit.entity.AuditLog;
import io.dataease.audit.service.IAuditService;
import io.dataease.auth.bo.TokenUserBO;
import io.dataease.utils.AuthUtils;
import lombok.extern.slf4j.Slf4j;
import org.aspectj.lang.ProceedingJoinPoint;
import org.aspectj.lang.reflect.MethodSignature;
import org.aspectj.lang.annotation.Around;
import org.aspectj.lang.annotation.Aspect;
import org.springframework.stereotype.Component;

import jakarta.servlet.http.HttpServletRequest;
import java.lang.reflect.Method;

@Aspect
@Component
@Slf4j
public class AuditAspect {

    private final IAuditService auditService;
    private final HttpServletRequest request;

    public AuditAspect(IAuditService auditService, HttpServletRequest request) {
        this.auditService = auditService;
        this.request = request;
    }

    @Around("@annotation(io.dataease.audit.annotation.AuditLog)")
    public Object aroundAuditLog(ProceedingJoinPoint joinPoint) throws Throwable {
        Method method = ((MethodSignature) joinPoint.getSignature()).getMethod();
        Object[] args = joinPoint.getArgs();
        Object result = null;

        try {
            result = joinPoint.proceed();
            
            AuditLog auditLog = createAuditLog(method, args, result, "SUCCESS", null);
            auditService.createAuditLog(auditLog);
            return result;
        } catch (Exception e) {
            AuditLog auditLog = createAuditLog(method, args, null, "FAILED", e.getMessage());
            auditService.createAuditLog(auditLog);
            throw e;
        }
    }

    private AuditLog createAuditLog(Method method, Object[] args, Object result, 
                                     String status, String failureReason) {
        AuditLog auditLog = new AuditLog();
        
        TokenUserBO userBO = AuthUtils.getUser();
        if (userBO != null) {
            auditLog.setUserId(userBO.getUserId());
        }

        String actionType = determineActionType(method);
        auditLog.setActionType(actionType);
        auditLog.setActionName(method.getName());
        auditLog.setOperation(determineOperation(method));
        auditLog.setStatus(status);
        auditLog.setFailureReason(failureReason);
        auditLog.setIpAddress(getRequestIp());
        auditLog.setUserAgent(getUserAgent());

        if (args != null && args.length > 0) {
            ResourceInfo resourceInfo = extractResourceInfo(args);
            auditLog.setResourceType(resourceInfo.type);
            auditLog.setResourceId(resourceInfo.id);
            auditLog.setResourceName(resourceInfo.name);
            auditLog.setBeforeValue(resourceInfo.before);
            
            if (result != null) {
                auditLog.setAfterValue(result.toString());
            }
        }

        return auditLog;
    }

    private String determineActionType(Method method) {
        String className = method.getDeclaringClass().getSimpleName();
        
        if (className.contains("User")) {
            return AuditConstants.ACTION_TYPE_USER_ACTION;
        } else if (className.contains("Organization") || className.contains("Org")) {
            return AuditConstants.ACTION_TYPE_PERMISSION_CHANGE;
        } else if (className.contains("Role") || className.contains("Permission")) {
            return AuditConstants.ACTION_TYPE_PERMISSION_CHANGE;
        } else if (className.contains("Dataset") || className.contains("Dashboard")) {
            return AuditConstants.ACTION_TYPE_DATA_ACCESS;
        } else {
            return AuditConstants.ACTION_TYPE_SYSTEM_CONFIG;
        }
    }

    private String determineOperation(Method method) {
        String methodName = method.getName();
        
        if (methodName.startsWith("create") || methodName.startsWith("add")) {
            return AuditConstants.OPERATION_CREATE;
        } else if (methodName.startsWith("update") || methodName.startsWith("modify") || methodName.startsWith("change")) {
            return AuditConstants.OPERATION_UPDATE;
        } else if (methodName.startsWith("delete") || methodName.startsWith("remove")) {
            return AuditConstants.OPERATION_DELETE;
        } else if (methodName.contains("login")) {
            return AuditConstants.OPERATION_LOGIN;
        } else if (methodName.contains("logout")) {
            return AuditConstants.OPERATION_LOGOUT;
        } else if (methodName.contains("export")) {
            return AuditConstants.OPERATION_EXPORT;
        } else {
            return "VIEW";
        }
    }

    private ResourceInfo extractResourceInfo(Object[] args) {
        ResourceInfo info = new ResourceInfo();
        
        if (args != null && args.length > 0) {
            Object firstArg = args[0];
            
            if (firstArg instanceof Long) {
                info.id = (Long) firstArg;
                info.type = "UNKNOWN";
            } else if (firstArg instanceof String) {
                info.name = (String) firstArg;
                info.type = "SYSTEM_CONFIG";
            }
        }
        
        return info;
    }

    private String determineResourceType(String typeName) {
        return switch (typeName) {
            case "User", "user" -> AuditConstants.RESOURCE_TYPE_USER;
            case "Organization", "org" -> AuditConstants.RESOURCE_TYPE_ORGANIZATION;
            case "Role", "role" -> AuditConstants.RESOURCE_TYPE_ROLE;
            case "Permission", "permission" -> AuditConstants.RESOURCE_TYPE_PERMISSION;
            case "Dataset", "dataset" -> AuditConstants.RESOURCE_TYPE_DATASET;
            case "Dashboard", "dashboard" -> AuditConstants.RESOURCE_TYPE_DASHBOARD;
            default -> "UNKNOWN";
        };
    }

    private static class ResourceInfo {
        Long id;
        String type;
        String name;
        String before;
        String after;
    }

    private String getRequestIp() {
        if (request == null) {
            return "";
        }
        String forwarded = request.getHeader("X-Forwarded-For");
        if (forwarded != null && !forwarded.isEmpty() && !"unknown".equalsIgnoreCase(forwarded)) {
            return forwarded.split(",")[0].trim();
        }
        return request.getRemoteAddr();
    }

    private String getUserAgent() {
        if (request == null) {
            return "";
        }
        String agent = request.getHeader("User-Agent");
        return agent != null ? agent : "";
    }
}
