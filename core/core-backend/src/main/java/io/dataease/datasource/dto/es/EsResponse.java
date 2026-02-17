package io.dataease.datasource.dto.es;


import java.util.ArrayList;
import java.util.List;

public class EsResponse {
    private List<Column> columns = new ArrayList<>();
    private List<String[]> rows = new ArrayList<>();
    private String cursor;
    private Integer status;
    private Error error;
    private String version;

    public List<Column> getColumns() { return columns; }
    public void setColumns(List<Column> columns) { this.columns = columns; }

    public List<String[]> getRows() { return rows; }
    public void setRows(List<String[]> rows) { this.rows = rows; }

    public String getCursor() { return cursor; }
    public void setCursor(String cursor) { this.cursor = cursor; }

    public Integer getStatus() { return status; }
    public void setStatus(Integer status) { this.status = status; }

    public Error getError() { return error; }
    public void setError(Error error) { this.error = error; }

    public String getVersion() { return version; }
    public void setVersion(String version) { this.version = version; }

    public static class Error {
        private String type;
        private String reason;

        public String getType() { return type; }
        public void setType(String type) { this.type = type; }

        public String getReason() { return reason; }
        public void setReason(String reason) { this.reason = reason; }
    }

    public static class Column {
        private String name;
        private String type;

        public String getName() { return name; }
        public void setName(String name) { this.name = name; }

        public String getType() { return type; }
        public void setType(String type) { this.type = type; }
    }

}
