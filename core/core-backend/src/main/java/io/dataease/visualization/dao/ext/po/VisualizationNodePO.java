package io.dataease.visualization.dao.ext.po;

import io.swagger.v3.oas.annotations.media.Schema;

import java.io.Serializable;


public class VisualizationNodePO implements Serializable {

    private Long id;
    private String name;
    private Long pid;
    private String nodeType;
    @Schema(description = "额外标识")
    private int extraFlag;
    @Schema(description = "额外标识1")
    private int extraFlag1;

    public Long getId() {
        return id;
    }

    public void setId(Long id) {
        this.id = id;
    }

    public String getName() {
        return name;
    }

    public void setName(String name) {
        this.name = name;
    }

    public Long getPid() {
        return pid;
    }

    public void setPid(Long pid) {
        this.pid = pid;
    }

    public String getNodeType() {
        return nodeType;
    }

    public void setNodeType(String nodeType) {
        this.nodeType = nodeType;
    }

    public int getExtraFlag() {
        return extraFlag;
    }

    public void setExtraFlag(int extraFlag) {
        this.extraFlag = extraFlag;
    }

    public int getExtraFlag1() {
        return extraFlag1;
    }

    public void setExtraFlag1(int extraFlag1) {
        this.extraFlag1 = extraFlag1;
    }
}
