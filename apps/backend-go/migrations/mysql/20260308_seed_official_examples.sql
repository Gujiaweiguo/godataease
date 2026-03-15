SET @now_ms = CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED);

SET NAMES utf8mb4;

INSERT INTO core_datasource (
  id,
  name,
  description,
  type,
  pid,
  edit_type,
  configuration,
  create_time,
  update_time,
  update_by,
  create_by,
  status,
  qrtz_instance,
  task_status,
  enable_data_fill,
  del_flag
)
SELECT
  985188400292302840,
  '官方示例-数据源',
  'Official sample datasource folder',
  'folder',
  0,
  '0',
  '{}',
  @now_ms,
  @now_ms,
  NULL,
  '1',
  'Success',
  NULL,
  NULL,
  0,
  0
FROM dual
WHERE NOT EXISTS (
  SELECT 1 FROM core_datasource WHERE id = 985188400292302840
);

UPDATE core_datasource
SET pid = 985188400292302840,
    del_flag = 0,
    type = 'MySQL'
WHERE id = 985188400292302848;

INSERT INTO core_dataset_group (
  id,
  name,
  pid,
  level,
  node_type,
  type,
  del_flag
)
SELECT
  985188400292302860,
  '官方示例-数据集',
  0,
  0,
  'folder',
  NULL,
  0
FROM dual
WHERE NOT EXISTS (
  SELECT 1 FROM core_dataset_group WHERE id = 985188400292302860
);

INSERT INTO core_dataset_group (
  id,
  name,
  pid,
  level,
  node_type,
  type,
  del_flag
)
SELECT
  985188400292302861,
  '官方示例数据集',
  985188400292302860,
  1,
  'dataset',
  'sql',
  0
FROM dual
WHERE NOT EXISTS (
  SELECT 1 FROM core_dataset_group WHERE id = 985188400292302861
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
