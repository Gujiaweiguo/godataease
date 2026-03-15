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
SELECT
  985188400292302861,
  '官方示例数据集',
  985188400292302860,
  1,
  'dataset',
  'db',
  0
FROM dual
WHERE NOT EXISTS (
  SELECT 1 FROM core_dataset_group WHERE id = 985188400292302861
);

UPDATE core_dataset_group
SET
  name = '官方示例-数据集',
  pid = 0,
  level = 0,
  node_type = 'folder',
  del_flag = 0
WHERE id = 985188400292302860;

UPDATE core_dataset_group
SET
  name = '官方示例数据集',
  pid = 985188400292302860,
  level = 1,
  node_type = 'dataset',
  type = 'db',
  del_flag = 0
WHERE id = 985188400292302861;

UPDATE core_datasource
SET
  name = '官方示例-数据源',
  pid = 0,
  type = 'folder',
  del_flag = 0
WHERE id = 985188400292302840;

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
  '官方示例数据集-订单',
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
  (985188400292302863, 985188400292302848, 985188400292302862, 985188400292302861, '店铺', '店铺', 'f_demo_shop', 'f_demo_shop', 'd', 'LONGTEXT', 0, 0, 0, 1, 1, @now_ms),
  (985188400292302864, 985188400292302848, 985188400292302862, 985188400292302861, '品线', '品线', 'f_demo_line', 'f_demo_line', 'd', 'LONGTEXT', 0, 0, 0, 1, 2, @now_ms),
  (985188400292302865, 985188400292302848, 985188400292302862, 985188400292302861, '菜品名称', '菜品名称', 'f_demo_product_name', 'f_demo_product_name', 'd', 'LONGTEXT', 0, 0, 0, 1, 3, @now_ms),
  (985188400292302866, 985188400292302848, 985188400292302862, 985188400292302861, '冷/热', '冷/热', 'f_demo_temp', 'f_demo_temp', 'd', 'LONGTEXT', 0, 0, 0, 1, 4, @now_ms),
  (985188400292302867, 985188400292302848, 985188400292302862, 985188400292302861, '规格', '规格', 'f_demo_spec', 'f_demo_spec', 'd', 'LONGTEXT', 0, 0, 0, 1, 5, @now_ms),
  (985188400292302868, 985188400292302848, 985188400292302862, 985188400292302861, '销售数量', '销售数量', 'f_demo_sale_qty', 'f_demo_sale_qty', 'q', 'BIGINT', 2, 2, 0, 1, 6, @now_ms),
  (985188400292302869, 985188400292302848, 985188400292302862, 985188400292302861, '单价', '单价', 'f_demo_price', 'f_demo_price', 'q', 'BIGINT', 2, 2, 0, 1, 7, @now_ms),
  (985188400292302870, 985188400292302848, 985188400292302862, 985188400292302861, '账单流水号', '账单流水号', 'f_demo_bill_no', 'f_demo_bill_no', 'd', 'LONGTEXT', 0, 0, 0, 1, 8, @now_ms),
  (985188400292302871, 985188400292302848, 985188400292302862, 985188400292302861, '销售日期', '销售日期', 'f_demo_sale_date', 'f_demo_sale_date', 'd', 'DATETIME', 1, 1, 0, 1, 9, @now_ms)
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
