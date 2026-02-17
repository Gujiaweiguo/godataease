package io.dataease.visualization.dao.ext.po;

import com.fasterxml.jackson.databind.annotation.JsonSerialize;
import com.fasterxml.jackson.databind.ser.std.ToStringSerializer;

import java.io.Serial;
import java.io.Serializable;

public class VisualizationResourcePO implements Serializable {
    @Serial
    private static final long serialVersionUID = 627770173259978185L;

    @JsonSerialize(using = ToStringSerializer.class)
    private Long id;

    @JsonSerialize(using = ToStringSerializer.class)
    private Long resourceId;

    private String name;

    private String type;

    private Long creator;

    private Long lastEditor;

    private Long lastEditTime;

    private Boolean favorite;

    private int weight;

    private Integer extFlag;

    public VisualizationResourcePO() {
    }

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

    public String getName() {
        return name;
    }

    public void setName(String name) {
        this.name = name;
    }

    public String getType() {
        return type;
    }

    public void setType(String type) {
        this.type = type;
    }

    public Long getCreator() {
        return creator;
    }

    public void setCreator(Long creator) {
        this.creator = creator;
    }

    public Long getLastEditor() {
        return lastEditor;
    }

    public void setLastEditor(Long lastEditor) {
        this.lastEditor = lastEditor;
    }

    public Long getLastEditTime() {
        return lastEditTime;
    }

    public void setLastEditTime(Long lastEditTime) {
        this.lastEditTime = lastEditTime;
    }

    public Boolean getFavorite() {
        return favorite;
    }

    public void setFavorite(Boolean favorite) {
        this.favorite = favorite;
    }

    public int getWeight() {
        return weight;
    }

    public void setWeight(int weight) {
        this.weight = weight;
    }

    public Integer getExtFlag() {
        return extFlag;
    }

    public void setExtFlag(Integer extFlag) {
        this.extFlag = extFlag;
    }
}
