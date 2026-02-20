package io.dataease.share.dao.ext.po;

import java.io.Serial;
import java.io.Serializable;

public class XpackSharePO implements Serializable {
    @Serial
    private static final long serialVersionUID = 7929343371768885789L;

    private Long shareId;

    private Long resourceId;

    private String name;

    private String type;

    private Long creator;

    private Long time;

    private Long exp;

    private Integer extFlag;

    private Integer extFlag1;

    public XpackSharePO() {
    }

    public Long getShareId() {
        return shareId;
    }

    public void setShareId(Long shareId) {
        this.shareId = shareId;
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

    public Long getTime() {
        return time;
    }

    public void setTime(Long time) {
        this.time = time;
    }

    public Long getExp() {
        return exp;
    }

    public void setExp(Long exp) {
        this.exp = exp;
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
