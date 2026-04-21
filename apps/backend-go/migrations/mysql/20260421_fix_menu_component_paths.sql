-- Migration: Fix menu component paths to match actual frontend view file locations
-- Description: Update component paths for dataset, datasource, screen, and panel menus
--   to match the actual Vue view file locations under src/views/.
-- Date: 2026-04-21
--
-- The frontend's resolveViewComponent function resolves component paths as:
--   ../views/${component}.vue  OR  ../views/${component}/index.vue
-- If the component path doesn't match an actual file, the route is silently skipped.

-- Dataset: visualized/data/dataset → visualized/data/dataset/index
UPDATE core_menu SET component = 'visualized/data/dataset/index'
WHERE id = 5 AND component = 'visualized/data/dataset';

-- Datasource: visualized/data/datasource → visualized/data/datasource/index
UPDATE core_menu SET component = 'visualized/data/datasource/index'
WHERE id = 6 AND component = 'visualized/data/datasource';

-- Screen: visualized/view/screen → visualized/view/screen/index
UPDATE core_menu SET component = 'visualized/view/screen/index'
WHERE id = 3 AND component = 'visualized/view/screen';

-- Panel: visualized/view/panel → dashboard/DashboardPreviewShow
UPDATE core_menu SET component = 'dashboard/DashboardPreviewShow'
WHERE id = 2 AND component = 'visualized/view/panel';

-- Workbranch: workbranch → workbranch/index
UPDATE core_menu SET component = 'workbranch/index'
WHERE id = 1 AND component = 'workbranch';
