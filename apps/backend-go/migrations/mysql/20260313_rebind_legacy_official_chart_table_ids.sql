SET NAMES utf8mb4;

SET @legacy_order_group_id = 985189053949415424;
SET @canonical_order_group_id = 985188400292302861;
SET @canonical_order_table_id = 985188400292302862;

START TRANSACTION;

UPDATE core_chart_view
SET table_id = @canonical_order_group_id
WHERE table_id = @legacy_order_group_id;

SELECT CASE
         WHEN EXISTS (
           SELECT 1
           FROM core_dataset_group
           WHERE id = @canonical_order_group_id
         ) THEN 1
         ELSE (1 / 0)
       END AS assert_canonical_order_group_exists;

SELECT CASE
         WHEN EXISTS (
           SELECT 1
           FROM core_dataset_table
           WHERE id = @canonical_order_table_id
             AND dataset_group_id = @canonical_order_group_id
         ) THEN 1
         ELSE (1 / 0)
       END AS assert_canonical_order_table_exists;

COMMIT;
