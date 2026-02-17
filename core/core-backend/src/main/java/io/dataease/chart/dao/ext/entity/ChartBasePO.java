package io.dataease.chart.dao.ext.entity;

import java.io.Serial;
import java.io.Serializable;

public class ChartBasePO implements Serializable {
    @Serial
    private static final long serialVersionUID = 183064537525500481L;

    private Long chartId;

    private String chartType;

    private String chartName;

    private Long resourceId;

    private String resourceType;

    private String resourceName;

    private Long tableId;

    private String xAxis;

    private String xAxisExt;

    private String yAxis;

    private String yAxisExt;

    private String extStack;

    private String extBubble;

    private String flowMapStartName;

    private String flowMapEndName;

    private String extColor;

    private String extLabel;

    private String extTooltip;

    public Long getChartId() {
        return chartId;
    }

    public void setChartId(Long chartId) {
        this.chartId = chartId;
    }

    public String getChartType() {
        return chartType;
    }

    public void setChartType(String chartType) {
        this.chartType = chartType;
    }

    public String getChartName() {
        return chartName;
    }

    public void setChartName(String chartName) {
        this.chartName = chartName;
    }

    public Long getResourceId() {
        return resourceId;
    }

    public void setResourceId(Long resourceId) {
        this.resourceId = resourceId;
    }

    public String getResourceType() {
        return resourceType;
    }

    public void setResourceType(String resourceType) {
        this.resourceType = resourceType;
    }

    public String getResourceName() {
        return resourceName;
    }

    public void setResourceName(String resourceName) {
        this.resourceName = resourceName;
    }

    public Long getTableId() {
        return tableId;
    }

    public void setTableId(Long tableId) {
        this.tableId = tableId;
    }

    public String getXAxis() {
        return xAxis;
    }

    public void setXAxis(String xAxis) {
        this.xAxis = xAxis;
    }

    public String getXAxisExt() {
        return xAxisExt;
    }

    public void setXAxisExt(String xAxisExt) {
        this.xAxisExt = xAxisExt;
    }

    public String getYAxis() {
        return yAxis;
    }

    public void setYAxis(String yAxis) {
        this.yAxis = yAxis;
    }

    public String getYAxisExt() {
        return yAxisExt;
    }

    public void setYAxisExt(String yAxisExt) {
        this.yAxisExt = yAxisExt;
    }

    public String getExtStack() {
        return extStack;
    }

    public void setExtStack(String extStack) {
        this.extStack = extStack;
    }

    public String getExtBubble() {
        return extBubble;
    }

    public void setExtBubble(String extBubble) {
        this.extBubble = extBubble;
    }

    public String getFlowMapStartName() {
        return flowMapStartName;
    }

    public void setFlowMapStartName(String flowMapStartName) {
        this.flowMapStartName = flowMapStartName;
    }

    public String getFlowMapEndName() {
        return flowMapEndName;
    }

    public void setFlowMapEndName(String flowMapEndName) {
        this.flowMapEndName = flowMapEndName;
    }

    public String getExtColor() {
        return extColor;
    }

    public void setExtColor(String extColor) {
        this.extColor = extColor;
    }

    public String getExtLabel() {
        return extLabel;
    }

    public void setExtLabel(String extLabel) {
        this.extLabel = extLabel;
    }

    public String getExtTooltip() {
        return extTooltip;
    }

    public void setExtTooltip(String extTooltip) {
        this.extTooltip = extTooltip;
    }
}
