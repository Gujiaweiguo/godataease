package io.dataease.datasource.dto;

import com.fasterxml.jackson.databind.annotation.JsonSerialize;
import com.fasterxml.jackson.databind.ser.std.ToStringSerializer;
import io.dataease.model.TreeBaseModel;

import java.io.Serial;


public class DatasourceNodeBO implements TreeBaseModel<DatasourceNodeBO> {

    @Serial
    private static final long serialVersionUID = 728340676442387790L;

    @JsonSerialize(using = ToStringSerializer.class)
    private Long id;
    private String name;
    private Boolean leaf;
    private Integer weight = 3;
    private Long pid;
    private Integer extraFlag;
    private String type;

    public DatasourceNodeBO() {
    }

    public DatasourceNodeBO(Long id, String name, Boolean leaf, Integer weight, Long pid, Integer extraFlag, String type) {
        this.id = id;
        this.name = name;
        this.leaf = leaf;
        this.weight = weight;
        this.pid = pid;
        this.extraFlag = extraFlag;
        this.type = type;
    }

    public Long getId() {
        return id;
    }

    public void setId(Long id) {
        this.id = id;
    }

    public Long getPid() {
        return pid;
    }

    public void setPid(Long pid) {
        this.pid = pid;
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

    public Integer getExtraFlag() {
        return extraFlag;
    }

    public void setExtraFlag(Integer extraFlag) {
        this.extraFlag = extraFlag;
    }

    public String getType() {
        return type;
    }

    public void setType(String type) {
        this.type = type;
    }
}
