package io.dataease.dataset.dto;

import io.dataease.model.TreeBaseModel;

import java.io.Serial;


public class DataSetNodeBO implements TreeBaseModel {

    @Serial
    private static final long serialVersionUID = 728340676442387790L;

    private Long id;
    private String name;
    private Boolean leaf;
    private Integer weight = 3;
    private Long pid;
    private Integer extraFlag;

    public DataSetNodeBO() {
    }

    public DataSetNodeBO(Long id, String name, Boolean leaf, Integer weight, Long pid, Integer extraFlag) {
        this.id = id;
        this.name = name;
        this.leaf = leaf;
        this.weight = weight;
        this.pid = pid;
        this.extraFlag = extraFlag;
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
}
