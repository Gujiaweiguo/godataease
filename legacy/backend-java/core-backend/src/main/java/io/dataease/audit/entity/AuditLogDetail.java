package io.dataease.audit.entity;

import com.baomidou.mybatisplus.annotation.TableName;
import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import lombok.Data;
import lombok.experimental.Accessors;

import java.time.LocalDateTime;

@Data
@Accessors(chain = true)
@TableName("de_audit_log_detail")
public class AuditLogDetail {

    @TableId(type = IdType.AUTO)
    private Long id;

    private Long auditLogId;

    private String detailType;

    private String detailKey;

    private String detailValue;

    private LocalDateTime createTime;
}
