package io.dataease.system.entity;

import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import io.swagger.v3.oas.annotations.media.Schema;

import java.io.Serializable;
import java.time.LocalDateTime;

@Schema(description = "数据集行级权限表")
@TableName("data_perm_row")
public class DataPermRow implements Serializable {

    private static final long serialVersionUID = 1L;

    @TableId
    @Schema(description = "主键ID")
    private Long id;

    @Schema(description = "数据集ID")
    private Long datasetId;

    @Schema(description = "授权对象类型：user-用户，role-角色")
    private String authTargetType;

    @Schema(description = "授权对象ID（用户ID或角色ID）")
    private Long authTargetId;

    @Schema(description = "权限表达式树：JSON格式")
    private String expressionTree;

    @Schema(description = "状态：0-禁用，1-启用")
    private Integer status;

    @Schema(description = "创建人")
    private String createBy;

    @Schema(description = "创建时间")
    private LocalDateTime createTime;

    @Schema(description = "更新人")
    private String updateBy;

    @Schema(description = "更新时间")
    private LocalDateTime updateTime;

    public Long getId() {
        return id;
    }

    public void setId(Long id) {
        this.id = id;
    }

    public Long getDatasetId() {
        return datasetId;
    }

    public void setDatasetId(Long datasetId) {
        this.datasetId = datasetId;
    }

    public String getAuthTargetType() {
        return authTargetType;
    }

    public void setAuthTargetType(String authTargetType) {
        this.authTargetType = authTargetType;
    }

    public Long getAuthTargetId() {
        return authTargetId;
    }

    public void setAuthTargetId(Long authTargetId) {
        this.authTargetId = authTargetId;
    }

    public String getExpressionTree() {
        return expressionTree;
    }

    public void setExpressionTree(String expressionTree) {
        this.expressionTree = expressionTree;
    }

    public Integer getStatus() {
        return status;
    }

    public void setStatus(Integer status) {
        this.status = status;
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

    public String getUpdateBy() {
        return updateBy;
    }

    public void setUpdateBy(String updateBy) {
        this.updateBy = updateBy;
    }

    public LocalDateTime getUpdateTime() {
        return updateTime;
    }

    public void setUpdateTime(LocalDateTime updateTime) {
        this.updateTime = updateTime;
    }
}
