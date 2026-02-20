package io.dataease.system.entity;

import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import io.swagger.v3.oas.annotations.media.Schema;

import java.io.Serializable;
import java.time.LocalDateTime;

@Schema(description = "用户权限关联表")
@TableName("sys_user_perm")
public class SysUserPerm implements Serializable {

    private static final long serialVersionUID = 1L;

    @TableId
    @Schema(description = "ID")
    private Long id;

    @Schema(description = "用户ID")
    private Long userId;

    @Schema(description = "组织ID")
    private Long orgId;

    @Schema(description = "权限ID")
    private Long permId;

    @Schema(description = "创建人")
    private String createBy;

    @Schema(description = "创建时间")
    private LocalDateTime createTime;

    @Schema(description = "状态：0-禁用，1-启用")
    private Integer status;

    @Schema(description = "删除标记：0-未删除，1-已删除")
    private Integer delFlag;

    public Long getId() {
        return id;
    }

    public void setId(Long id) {
        this.id = id;
    }

    public Long getUserId() {
        return userId;
    }

    public void setUserId(Long userId) {
        this.userId = userId;
    }

    public Long getOrgId() {
        return orgId;
    }

    public void setOrgId(Long orgId) {
        this.orgId = orgId;
    }

    public Long getPermId() {
        return permId;
    }

    public void setPermId(Long permId) {
        this.permId = permId;
    }

    public String getCreateBy() {
        return createBy;
    }

    public void setCreateBy(String createBy) {
        this.createBy = createBy;
    }

    public LocalDateTime getCreateTime() {
        return createTime;
    }

    public void setCreateTime(LocalDateTime createTime) {
        this.createTime = createTime;
    }

    public Integer getStatus() {
        return status;
    }

    public void setStatus(Integer status) {
        this.status = status;
    }

    public Integer getDelFlag() {
        return delFlag;
    }

    public void setDelFlag(Integer delFlag) {
        this.delFlag = delFlag;
    }
}
