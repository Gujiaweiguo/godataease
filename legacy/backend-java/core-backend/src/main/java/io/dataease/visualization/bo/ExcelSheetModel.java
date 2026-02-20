package io.dataease.visualization.bo;

import java.io.Serial;
import java.io.Serializable;
import java.util.List;

public class ExcelSheetModel implements Serializable {
    @Serial
    private static final long serialVersionUID = 1122095875367371623L;

    private String sheetName;

    private List<String> heads;

    private List<List<String>> data;

    private List<Integer> filedTypes;

    public String getSheetName() {
        return sheetName;
    }

    public void setSheetName(String sheetName) {
        this.sheetName = sheetName;
    }

    public List<String> getHeads() {
        return heads;
    }

    public void setHeads(List<String> heads) {
        this.heads = heads;
    }

    public List<List<String>> getData() {
        return data;
    }

    public void setData(List<List<String>> data) {
        this.data = data;
    }

    public List<Integer> getFiledTypes() {
        return filedTypes;
    }

    public void setFiledTypes(List<Integer> filedTypes) {
        this.filedTypes = filedTypes;
    }
}
