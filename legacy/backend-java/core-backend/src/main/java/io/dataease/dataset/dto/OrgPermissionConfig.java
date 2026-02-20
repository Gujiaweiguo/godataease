package io.dataease.dataset.dto;

import io.swagger.v3.oas.annotations.media.Schema;
import lombok.Data;

@Data
@Schema(description = "组织权限配置")
public class OrgPermissionConfig {

    @Schema(description = "组织字段ID")
    private Long orgFieldId;

    @Schema(description = "是否启用自动过滤")
    private Boolean enableAutoFilter;

    @Schema(description = "过滤类型: current_org(只看当前组织), current_and_sub_orgs(看当前及子组织)")
    private String orgFilterType;
}
