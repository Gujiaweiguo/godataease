SET @now_ms = CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED);

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
  985188400292302848,
  'Demo MySQL',
  'Official sample datasource',
  'MySQL',
  0,
  '0',
  'eyJkYXRhQmFzZSI6ImRhdGFlYXNlX2RldiIsImhvc3QiOiJteXNxbDgiLCJwb3J0IjozMzA2LCJ1c2VybmFtZSI6InJvb3QiLCJwYXNzd29yZCI6IkFkbWluMTY4IiwiamRiY1VybCI6ImpkYmM6bXlzcWw6Ly9teXNxbDg6MzMwNi9kYXRhZWFzZV9kZXY/dXNlVW5pY29kZT10cnVlJmNoYXJhY3RlckVuY29kaW5nPVVURi04JnNlcnZlclRpbWV6b25lPUFzaWEvU2hhbmdoYWkifQ==',
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
  SELECT 1 FROM core_datasource WHERE id = 985188400292302848
);

UPDATE core_datasource
SET del_flag = 0,
    type = 'MySQL'
WHERE id = 985188400292302848;
