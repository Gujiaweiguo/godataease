package io.dataease.datasource.dao.ext.po;

import lombok.Data;

@Data
public class Ctimestamp {
    private Long currentTimestamp;

    public Long getCurrentTimestamp() {
        return currentTimestamp;
    }

    public void setCurrentTimestamp(Long currentTimestamp) {
        this.currentTimestamp = currentTimestamp;
    }
}
