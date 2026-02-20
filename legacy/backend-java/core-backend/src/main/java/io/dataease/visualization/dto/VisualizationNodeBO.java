package io.dataease.visualization.dto;

import io.dataease.model.TreeBaseModel;

import java.io.Serial;

public class VisualizationNodeBO implements TreeBaseModel {
    @Serial
    private static final long serialVersionUID = -4998292096597683628L;

    private Long id;
    private String name;
    private Boolean leaf;
    private Integer weight = 3;
    private Long pid;
    private Integer extraFlag;
    private Integer extraFlag1;

    public VisualizationNodeBO() {
    }

    public VisualizationNodeBO(Long id, String name, Boolean leaf, Integer weight, Long pid, Integer extraFlag, Integer extraFlag1) {
        this.id = id;
        this.name = name;
        this.leaf = leaf;
        this.weight = weight;
        this.pid = pid;
        this.extraFlag = extraFlag;
        this.extraFlag1 = extraFlag1;
    }

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

    public Boolean getLeaf() {
        return leaf;
    }

    public void setLeaf(Boolean leaf) {
        this.leaf = leaf;
    }

    public Integer getWeight() {
        return weight;
    }

    public void setWeight(Integer weight) {
        this.weight = weight;
    }

    public Long getPid() {
        return pid;
    }

    public void setPid(Long pid) {
        this.pid = pid;
    }

    public Integer getExtraFlag() {
        return extraFlag;
    }

    public void setExtraFlag(Integer extraFlag) {
        this.extraFlag = extraFlag;
    }

    public Integer getExtraFlag1() {
        return extraFlag1;
    }

    public void setExtraFlag1(Integer extraFlag1) {
        this.extraFlag1 = extraFlag1;
    }
}
