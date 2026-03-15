SET NAMES utf8mb4;

UPDATE data_visualization_info
SET
  canvas_style_data = CASE
    WHEN canvas_style_data IS NULL OR TRIM(canvas_style_data) = '' OR canvas_style_data = '{}' OR JSON_VALID(canvas_style_data) = 0 THEN '{"width":1920,"height":1080,"scale":60,"scaleWidth":60,"scaleHeight":60,"fontFamily":"PingFang","dashboard":{"gap":"yes","gapSize":5,"resultMode":"all","resultCount":1000,"themeColor":"light","mobileSetting":{"customSetting":false,"imageUrl":null,"backgroundType":"image","color":"#000"},"showGrid":false,"matrixBase":4,"gapMode":"middle"},"component":{"seniorStyleSetting":{"linkageIconColor":"#646A73","drillLayerColor":"#3370ff"}}}'
    ELSE canvas_style_data
  END,
  component_data = CASE
    WHEN component_data IS NULL OR TRIM(component_data) = '' OR component_data = '{}' OR JSON_VALID(component_data) = 0 THEN '[{"id":"official-demo-dashboard-text","component":"UserView","name":"富文本","label":"富文本","innerType":"rich-text","canvasId":"canvas-main","isShow":true,"show":true,"x":1,"y":1,"sizeX":24,"sizeY":6,"style":{"left":0,"top":0,"width":320,"height":120,"rotate":0,"opacity":1},"propValue":{"textValue":"<p style=\"text-align:center;line-height:1.8;\"><span style=\"font-size:28px;\"><strong>连锁茶饮销售看板</strong></span></p><p style=\"text-align:center;\">官方示例内容已恢复，可继续编辑和拖拽组件。</p>"}}]'
    ELSE component_data
  END
WHERE id = '985188400292302870';

UPDATE data_visualization_info
SET
  canvas_style_data = CASE
    WHEN canvas_style_data IS NULL OR TRIM(canvas_style_data) = '' OR canvas_style_data = '{}' OR JSON_VALID(canvas_style_data) = 0 THEN '{"width":1920,"height":1080,"scale":60,"scaleWidth":60,"scaleHeight":60,"fontFamily":"PingFang","dashboard":{"gap":"no","gapSize":0,"resultMode":"all","resultCount":1000,"themeColor":"dark","mobileSetting":{"customSetting":false,"imageUrl":null,"backgroundType":"image","color":"#fff"},"showGrid":false,"matrixBase":4,"gapMode":"middle"},"component":{"seniorStyleSetting":{"linkageIconColor":"#DDE3EA","drillLayerColor":"#3370ff"}}}'
    ELSE canvas_style_data
  END,
  component_data = CASE
    WHEN component_data IS NULL OR TRIM(component_data) = '' OR component_data = '{}' OR JSON_VALID(component_data) = 0 THEN '[{"id":"official-demo-screen-text","component":"UserView","name":"富文本","label":"富文本","innerType":"rich-text","canvasId":"canvas-main","isShow":true,"show":true,"x":1,"y":1,"sizeX":24,"sizeY":6,"style":{"left":0,"top":0,"width":320,"height":120,"rotate":0,"opacity":1},"propValue":{"textValue":"<p style=\"text-align:center;line-height:1.8;\"><span style=\"font-size:28px;\"><strong>官方示例数据大屏</strong></span></p><p style=\"text-align:center;\">官方示例内容已恢复，可继续编辑和拖拽组件。</p>"}}]'
    ELSE component_data
  END
WHERE id = '985188400292302871';
