SET @now_ms = CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED);

SET NAMES utf8mb4;

INSERT INTO data_visualization_info (
  id,
  name,
  pid,
  org_id,
  level,
  node_type,
  type,
  canvas_style_data,
  component_data,
  mobile_layout,
  status,
  sort,
  create_time,
  create_by,
  update_time,
  update_by,
  delete_flag
)
SELECT
  '985188400292302870',
  '连锁茶饮销售看板',
  '0',
  NULL,
  0,
  'panel',
  'dashboard',
  '{}',
  '{}',
  0,
  1,
  0,
  @now_ms,
  '1',
  @now_ms,
  '1',
  0
FROM dual
WHERE NOT EXISTS (
  SELECT 1 FROM data_visualization_info WHERE id = '985188400292302870'
);

INSERT INTO data_visualization_info (
  id,
  name,
  pid,
  org_id,
  level,
  node_type,
  type,
  canvas_style_data,
  component_data,
  mobile_layout,
  status,
  sort,
  create_time,
  create_by,
  update_time,
  update_by,
  delete_flag
)
SELECT
  '985188400292302871',
  '官方示例数据大屏',
  '0',
  NULL,
  0,
  'panel',
  'dataV',
  '{}',
  '{}',
  0,
  1,
  0,
  @now_ms,
  '1',
  @now_ms,
  '1',
  0
FROM dual
WHERE NOT EXISTS (
  SELECT 1 FROM data_visualization_info WHERE id = '985188400292302871'
);

UPDATE data_visualization_info
SET
  name = '连锁茶饮销售看板',
  node_type = 'panel',
  type = 'dashboard',
  status = 1,
  delete_flag = 0,
  update_time = @now_ms,
  update_by = '1'
WHERE id = '985188400292302870';

UPDATE data_visualization_info
SET
  name = '官方示例数据大屏',
  node_type = 'panel',
  type = 'dataV',
  status = 1,
  delete_flag = 0,
  update_time = @now_ms,
  update_by = '1'
WHERE id = '985188400292302871';
