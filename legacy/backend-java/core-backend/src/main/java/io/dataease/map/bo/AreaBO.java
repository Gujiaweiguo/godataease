package io.dataease.map.bo;

import io.dataease.map.dao.auto.entity.Area;
import lombok.EqualsAndHashCode;

import java.io.Serializable;

@EqualsAndHashCode(callSuper = true)
public class AreaBO extends Area implements Serializable {
    private boolean custom = false;

    public boolean isCustom() {
        return custom;
    }

    public void setCustom(boolean custom) {
        this.custom = custom;
    }

    public boolean getCustom() {
        return custom;
    }
}
