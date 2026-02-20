package io.dataease.visualization.dto;

public class WatermarkContentDTO {

    private Boolean enable;

    private Boolean excelEnable = false;

    private Boolean enablePanelCustom;

    private String type;

    private String content;

    private String watermark_color;

    private Integer watermark_x_space;

    private Integer watermark_y_space;

    private Integer watermark_fontsize;

    public Boolean getEnable() { return enable; }
    public void setEnable(Boolean enable) { this.enable = enable; }

    public Boolean getExcelEnable() { return excelEnable; }
    public void setExcelEnable(Boolean excelEnable) { this.excelEnable = excelEnable; }

    public Boolean getEnablePanelCustom() { return enablePanelCustom; }
    public void setEnablePanelCustom(Boolean enablePanelCustom) { this.enablePanelCustom = enablePanelCustom; }

    public String getType() { return type; }
    public void setType(String type) { this.type = type; }

    public String getContent() { return content; }
    public void setContent(String content) { this.content = content; }

    public String getWatermark_color() { return watermark_color; }
    public void setWatermark_color(String watermark_color) { this.watermark_color = watermark_color; }

    public Integer getWatermark_x_space() { return watermark_x_space; }
    public void setWatermark_x_space(Integer watermark_x_space) { this.watermark_x_space = watermark_x_space; }

    public Integer getWatermark_y_space() { return watermark_y_space; }
    public void setWatermark_y_space(Integer watermark_y_space) { this.watermark_y_space = watermark_y_space; }

    public Integer getWatermark_fontsize() { return watermark_fontsize; }
    public void setWatermark_fontsize(Integer watermark_fontsize) { this.watermark_fontsize = watermark_fontsize; }
}
