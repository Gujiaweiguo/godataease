SET NAMES utf8mb4;

SET @demo_ds_id = 985188400292302848;
SET @order_group_id = 985188400292302861;
SET @material_group_id = 985189703189925888;
SET @order_table_id = 985188400292302862;
SET @material_table_id = 7193457660727922688;

START TRANSACTION;

UPDATE core_dataset_group
SET name = '茶饮订单明细',
    pid = 985188400292302860,
    level = 1,
    node_type = 'dataset',
    type = 'db',
    del_flag = 0
WHERE id = @order_group_id;

UPDATE core_dataset_group
SET name = '茶饮原料费用',
    pid = 985188400292302860,
    level = 1,
    node_type = 'dataset',
    type = 'db',
    del_flag = 0
WHERE id = @material_group_id;

UPDATE core_dataset_table
SET name = '茶饮订单明细',
    table_name = 'demo_tea_order',
    datasource_id = @demo_ds_id,
    dataset_group_id = @order_group_id,
    type = 'db',
    info = 'demo_tea_order'
WHERE id = @order_table_id;

UPDATE core_dataset_table
SET name = '茶饮原料费用',
    table_name = 'demo_tea_material',
    datasource_id = @demo_ds_id,
    dataset_group_id = @material_group_id,
    type = 'db',
    info = 'demo_tea_material'
WHERE id = @material_table_id;

UPDATE core_dataset_table_field
SET datasource_id = @demo_ds_id,
    dataset_group_id = @order_group_id,
    dataset_table_id = @order_table_id
WHERE dataset_table_id = @order_table_id;

UPDATE core_dataset_table_field
SET datasource_id = @demo_ds_id,
    dataset_group_id = @material_group_id,
    dataset_table_id = @material_table_id
WHERE dataset_table_id = @material_table_id;

DROP TEMPORARY TABLE IF EXISTS tmp_tea_legacy_tables;
CREATE TEMPORARY TABLE tmp_tea_legacy_tables AS
SELECT t.id, t.table_name, t.dataset_group_id
FROM core_dataset_table t
LEFT JOIN core_dataset_group g ON g.id = t.dataset_group_id
WHERE t.table_name IN ('demo_tea_order', 'demo_tea_material')
  AND t.id NOT IN (@order_table_id, @material_table_id)
  AND (
    t.datasource_id = @demo_ds_id
    OR t.name IS NULL
    OR t.name IN ('官方示例数据集-订单', '茶饮订单明细', '茶饮原料费用')
    OR g.id IS NULL
  );

UPDATE core_dataset_table_field f
JOIN tmp_tea_legacy_tables t ON f.dataset_table_id = t.id
SET f.datasource_id = @demo_ds_id,
    f.dataset_group_id = CASE
      WHEN t.table_name = 'demo_tea_order' THEN @order_group_id
      ELSE @material_group_id
    END,
    f.dataset_table_id = CASE
      WHEN t.table_name = 'demo_tea_order' THEN @order_table_id
      ELSE @material_table_id
    END;

DELETE f
FROM core_dataset_table_field f
JOIN (
  SELECT DISTINCT dataset_group_id
  FROM tmp_tea_legacy_tables
  WHERE dataset_group_id IS NOT NULL
) lg ON lg.dataset_group_id = f.dataset_group_id
LEFT JOIN core_dataset_group g ON g.id = f.dataset_group_id
WHERE g.id IS NULL;

UPDATE core_dataset_group
SET del_flag = 1
WHERE id IN (
  SELECT dataset_group_id
  FROM (
    SELECT DISTINCT dataset_group_id
    FROM tmp_tea_legacy_tables
    WHERE dataset_group_id IS NOT NULL
  ) x
)
  AND id NOT IN (@order_group_id, @material_group_id);

DELETE t
FROM core_dataset_table t
JOIN tmp_tea_legacy_tables l ON l.id = t.id;

SET @dup_count = (
  SELECT COUNT(*)
  FROM core_dataset_table
  WHERE table_name IN ('demo_tea_order', 'demo_tea_material')
    AND id NOT IN (@order_table_id, @material_table_id)
);
SELECT CASE WHEN @dup_count = 0 THEN 1 ELSE (1 / 0) END AS assert_no_legacy_duplicates;

SET @canonical_bind_count = (
  SELECT COUNT(*)
  FROM core_dataset_table
  WHERE id IN (@order_table_id, @material_table_id)
    AND datasource_id = @demo_ds_id
);
SELECT CASE WHEN @canonical_bind_count = 2 THEN 1 ELSE (1 / 0) END AS assert_canonical_tables_bound;

SET @missing_field_bind_count = (
  SELECT COUNT(*)
  FROM core_dataset_table_field
  WHERE dataset_table_id IN (@order_table_id, @material_table_id)
    AND (datasource_id IS NULL OR datasource_id = 0)
);
SELECT CASE WHEN @missing_field_bind_count = 0 THEN 1 ELSE (1 / 0) END AS assert_canonical_fields_bound;

DROP TEMPORARY TABLE IF EXISTS tmp_tea_legacy_tables;

COMMIT;
