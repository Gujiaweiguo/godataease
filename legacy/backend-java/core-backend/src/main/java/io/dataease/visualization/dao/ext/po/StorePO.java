package io.dataease.visualization.dao.ext.po;

import java.io.Serial;
import java.io.Serializable;

public class StorePO implements Serializable {
    @Serial
    private static final long serialVersionUID = 9130790627765997999L;

    private Long resourceId;

    private String type;

    private Long creator;

    private Long editor;

    private Long editTime;

    private Long storeId;

    private String name;

    private Integer extFlag;

    private Integer extFlag1;

    public Long getResourceId() {
        return resourceId;
    }

    public void setResourceId(Long resourceId) {
        this.resourceId = resourceId;
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

    public Long getEditor() {
        return editor;
    }

    public void setEditor(Long editor) {
        this.editor = editor;
    }

    public Long getEditTime() {
        return editTime;
    }

    public void setEditTime(Long editTime) {
        this.editTime = editTime;
    }

    public Long getStoreId() {
        return storeId;
    }

    public void setStoreId(Long storeId) {
        this.storeId = storeId;
    }

    public String getName() {
        return name;
    }

    public void setName(String name) {
        this.name = name;
    }

    public Integer getExtFlag() {
        return extFlag;
    }

    public void setExtFlag(Integer extFlag) {
        this.extFlag = extFlag;
    }

    public Integer getExtFlag1() {
        return extFlag1;
    }

    public void setExtFlag1(Integer extFlag1) {
        this.extFlag1 = extFlag1;
    }
}
