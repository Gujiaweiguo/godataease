package io.dataease.dataset.dao.ext.po;

import java.io.Serial;
import java.io.Serializable;

public class DataSetNodePO implements Serializable {

    @Serial
    private static final long serialVersionUID = -4457506330575500164L;

    private Long id;
    private String name;
    private String nodeType;
    private Long pid;

    public DataSetNodePO() {
    }

    public DataSetNodePO(Long id, String name, String nodeType, Long pid) {
        this.id = id;
        this.name = name;
        this.nodeType = nodeType;
        this.pid = pid;
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

    public String getNodeType() {
        return nodeType;
    }

    public void setNodeType(String nodeType) {
        this.nodeType = nodeType;
    }

    public Long getPid() {
        return pid;
    }

    public void setPid(Long pid) {
        this.pid = pid;
    }
}
