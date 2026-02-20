package io.dataease.system.entity;

import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import io.swagger.v3.oas.annotations.media.Schema;

import java.io.Serializable;

@Schema(description = "角色权限关联表")
@TableName("sys_role_perm")
public class SysRolePerm implements Serializable {

    private static final long serialVersionUID = 1L;

    @TableId
    @Schema(description = "ID")
    private Long id;

    @Schema(description = "角色ID")
    private Long roleId;

    @Schema(description = "权限ID")
    private Long permId;

    public Long getId() {
        return id;
    }

    public void setId(Long id) {
        this.id = id;
    }

    public Long getRoleId() {
        return roleId;
    }

    public void setRoleId(Long roleId) {
        this.roleId = roleId;
    }

    public Long getPermId() {
        return permId;
    }

    public void setPermId(Long permId) {
        this.permId = permId;
    }
}
