package io.dataease.datasource.dao.ext.po;

import java.io.Serial;
import java.io.Serializable;

public class DsItem implements Serializable {
    @Serial
    private static final long serialVersionUID = 370886385725258461L;

    private Long id;

    private Long pid;

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
}
