package io.dataease.system.entity;

import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import io.swagger.v3.oas.annotations.media.Schema;

import java.io.Serializable;

@Schema(description = "资源权限关联表")
@TableName("sys_resource_perm")
public class SysResourcePerm implements Serializable {

    private static final long serialVersionUID = 1L;

    @TableId
    @Schema(description = "ID")
    private Long id;

    @Schema(description = "资源ID")
    private Long resourceId;

    @Schema(description = "权限ID")
    private Long permId;

    public Long getId() {
        return id;
    }

    public void setId(Long id) {
        this.id = id;
    }

    public Long getResourceId() {
        return resourceId;
    }

    public void setResourceId(Long resourceId) {
        this.resourceId = resourceId;
    }

    public Long getPermId() {
        return permId;
    }

    public void setPermId(Long permId) {
        this.permId = permId;
    }
}
