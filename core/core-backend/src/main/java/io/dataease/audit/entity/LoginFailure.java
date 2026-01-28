package io.dataease.audit.entity;

import com.baomidou.mybatisplus.annotation.TableName;
import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import lombok.Data;
import lombok.experimental.Accessors;

import java.time.LocalDateTime;

@Data
@Accessors(chain = true)
@TableName("de_login_failure")
public class LoginFailure {

    @TableId(type = IdType.AUTO)
    private Long id;

    private String username;

    private String ipAddress;

    private String failureReason;

    private String userAgent;

    private LocalDateTime createTime;
}
