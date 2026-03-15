SET NAMES utf8mb4;
SET @now_ms = CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED);

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
VALUES (
  985188400292302861,
  '茶饮订单明细',
  985188400292302860,
  1,
  'dataset',
  'db',
  0
)
ON DUPLICATE KEY UPDATE
  name = VALUES(name),
  pid = VALUES(pid),
  level = VALUES(level),
  node_type = VALUES(node_type),
  type = VALUES(type),
  del_flag = VALUES(del_flag);

INSERT INTO core_dataset_group (
  id,
  name,
  pid,
  level,
  node_type,
  type,
  del_flag
)
VALUES (
  985189703189925888,
  '茶饮原料费用',
  985188400292302860,
  1,
  'dataset',
  'db',
  0
)
ON DUPLICATE KEY UPDATE
  name = VALUES(name),
  pid = VALUES(pid),
  level = VALUES(level),
  node_type = VALUES(node_type),
  type = VALUES(type),
  del_flag = VALUES(del_flag);

INSERT INTO core_dataset_table (
  id,
  name,
  table_name,
  datasource_id,
  dataset_group_id,
  type,
  info,
  sql_variable_details
)
VALUES (
  985188400292302862,
  '茶饮订单明细',
  'demo_tea_order',
  985188400292302848,
  985188400292302861,
  'db',
  'demo_tea_order',
  NULL
)
ON DUPLICATE KEY UPDATE
  name = VALUES(name),
  table_name = VALUES(table_name),
  datasource_id = VALUES(datasource_id),
  dataset_group_id = VALUES(dataset_group_id),
  type = VALUES(type),
  info = VALUES(info),
  sql_variable_details = VALUES(sql_variable_details);

INSERT INTO core_dataset_table (
  id,
  name,
  table_name,
  datasource_id,
  dataset_group_id,
  type,
  info,
  sql_variable_details
)
VALUES (
  7193457660727922688,
  '茶饮原料费用',
  'demo_tea_material',
  985188400292302848,
  985189703189925888,
  'db',
  'demo_tea_material',
  NULL
)
ON DUPLICATE KEY UPDATE
  name = VALUES(name),
  table_name = VALUES(table_name),
  datasource_id = VALUES(datasource_id),
  dataset_group_id = VALUES(dataset_group_id),
  type = VALUES(type),
  info = VALUES(info),
  sql_variable_details = VALUES(sql_variable_details);

INSERT INTO core_dataset_table_field (
  id,
  datasource_id,
  dataset_table_id,
  dataset_group_id,
  origin_name,
  name,
  dataease_name,
  field_short_name,
  group_type,
  type,
  de_type,
  de_extract_type,
  ext_field,
  checked,
  column_index,
  last_sync_time
)
VALUES
  (1715053944934, 985188400292302848, 7193457660727922688, 985189703189925888, '店铺', '店铺', 'f_4a4cd188441bb10a', 'f_4a4cd188441bb10a', 'd', 'LONGTEXT', 0, 0, 0, 1, 1, @now_ms),
  (1715053944935, 985188400292302848, 7193457660727922688, 985189703189925888, '日期', '日期', 'f_7fedb6b454fd0ddb', 'f_7fedb6b454fd0ddb', 'd', 'DATETIME', 1, 1, 0, 1, 2, @now_ms),
  (1715053944936, 985188400292302848, 7193457660727922688, 985189703189925888, '用途', '用途', 'f_703aac67af8ea53d', 'f_703aac67af8ea53d', 'd', 'LONGTEXT', 0, 0, 0, 1, 3, @now_ms),
  (1715053944937, 985188400292302848, 7193457660727922688, 985189703189925888, '金额', '金额', 'f_8cc276e515d2de6d', 'f_8cc276e515d2de6d', 'q', 'BIGINT', 2, 2, 0, 1, 4, @now_ms)
ON DUPLICATE KEY UPDATE
  datasource_id = VALUES(datasource_id),
  dataset_table_id = VALUES(dataset_table_id),
  dataset_group_id = VALUES(dataset_group_id),
  origin_name = VALUES(origin_name),
  name = VALUES(name),
  dataease_name = VALUES(dataease_name),
  field_short_name = VALUES(field_short_name),
  group_type = VALUES(group_type),
  type = VALUES(type),
  de_type = VALUES(de_type),
  de_extract_type = VALUES(de_extract_type),
  ext_field = VALUES(ext_field),
  checked = VALUES(checked),
  column_index = VALUES(column_index),
  last_sync_time = VALUES(last_sync_time);
