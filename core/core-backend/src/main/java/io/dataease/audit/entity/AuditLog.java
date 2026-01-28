package io.dataease.audit.entity;

import com.baomidou.mybatisplus.annotation.TableName;
import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import lombok.Data;
import lombok.experimental.Accessors;

import java.time.LocalDateTime;

@Data
@Accessors(chain = true)
@TableName("de_audit_log")
public class AuditLog {

    @TableId(type = IdType.AUTO)
    private Long id;

    private Long userId;

    private String username;

    private String actionType;

    private String actionName;

    private String resourceType;

    private Long resourceId;

    private String resourceName;

    private String operation;

    private String status;

    private String failureReason;

    private String ipAddress;

    private String userAgent;

    private String beforeValue;

    private String afterValue;

    private Long organizationId;

    private LocalDateTime createTime;
}
