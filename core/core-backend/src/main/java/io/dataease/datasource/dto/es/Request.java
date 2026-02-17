package io.dataease.datasource.dto.es;

public class Request {
    private String query;
    private Integer fetch_size = 10000;
    private boolean field_multi_value_leniency = true;

    public String getQuery() {
        return query;
    }

    public void setQuery(String query) {
        this.query = query;
    }

    public Integer getFetch_size() {
        return fetch_size;
    }

    public void setFetch_size(Integer fetch_size) {
        this.fetch_size = fetch_size;
    }

    public boolean isField_multi_value_leniency() {
        return field_multi_value_leniency;
    }

    public void setField_multi_value_leniency(boolean field_multi_value_leniency) {
        this.field_multi_value_leniency = field_multi_value_leniency;
    }
}
