# Baseline Naming Rules

## Overview

This document defines the directory structure, file naming conventions, and path resolution rules for API baseline fixtures used in contract diffing.

## Directory Structure

```
baselines/
├── auth/           # Authentication and authorization APIs
├── chart/          # Chart and visualization APIs
├── dataset/        # Dataset management APIs
├── datasource/     # Data source configuration APIs
├── export/         # Export and download APIs
├── org/            # Organization management APIs
├── share/          # Sharing and permission APIs
├── template/       # Template management APIs
└── user/           # User management APIs
```

### Capability Groups

| Capability | Description | Example Endpoints |
|------------|-------------|-------------------|
| `auth` | Login, logout, token management | `/api/auth/login`, `/api/auth/refresh` |
| `chart` | Chart CRUD, rendering | `/api/chart/create`, `/api/chart/view` |
| `dataset` | Dataset operations | `/api/dataset/list`, `/api/dataset/field` |
| `datasource` | Data source management | `/api/datasource/list`, `/api/datasource/validate` |
| `export` | PDF, image, Excel exports | `/api/export/pdf`, `/api/export/panel` |
| `org` | Organization settings | `/api/org/list`, `/api/org/setting` |
| `share` | Sharing and links | `/api/share/create`, `/api/share/validate` |
| `template` | Template management | `/api/template/list`, `/api/template/apply` |
| `user` | User CRUD operations | `/api/user/list`, `/api/user/profile` |

## File Naming Rules

### Conversion Rules

1. **Remove leading `/api/` prefix** - All API paths start with `/api/`, which is omitted from filenames
2. **Replace `/` with `_`** - Path segments become underscore-separated
3. **Append method suffix** - `_{METHOD}.json` (uppercase)
4. **Handle path parameters** - Replace `:param` with `_param`

### Path to Filename

| API Path | HTTP Method | Filename |
|----------|-------------|----------|
| `/api/datasource/list` | POST | `datasource_list_POST.json` |
| `/api/dataset/detail/:id` | GET | `dataset_detail_id_GET.json` |
| `/api/chart/view/:chartId` | POST | `chart_view_chartId_POST.json` |
| `/api/user/profile` | GET | `user_profile_GET.json` |
| `/api/auth/login` | POST | `auth_login_POST.json` |
| `/api/export/pdf/panel/:id` | GET | `export_pdf_panel_id_GET.json` |

### Filename to Path

To derive the API path from a filename:

1. Remove the method suffix (`_{METHOD}.json`)
2. Replace `_` with `/`
3. Prepend `/api/`
4. Restore path parameters: `_param` → `/:param`

## Path Resolution Rules

### Resolve File from Path + Method

```
Input: path="/api/dataset/list", method="POST"
1. Strip prefix: dataset/list
2. Replace slashes: dataset_list
3. Append method: dataset_list_POST.json
4. Locate in capability directory: baselines/dataset/dataset_list_POST.json
```

### Resolve Path from File

```
Input: baselines/chart/chart_view_chartId_POST.json
1. Extract relative path: chart/chart_view_chartId_POST.json
2. Capability: chart
3. Filename without extension: chart_view_chartId_POST
4. Remove method suffix: chart_view_chartId
5. Replace underscores with slashes: chart/view/chartId
6. Prepend /api/: /api/chart/view/chartId
7. Identify path params (lowercase segments): /api/chart/view/:chartId
```

## Example Mappings

| File Path | API Path | Method | Capability |
|-----------|----------|--------|------------|
| `baselines/auth/auth_login_POST.json` | `/api/auth/login` | POST | auth |
| `baselines/datasource/datasource_list_POST.json` | `/api/datasource/list` | POST | datasource |
| `baselines/dataset/dataset_list_POST.json` | `/api/dataset/list` | POST | dataset |
| `baselines/dataset/dataset_detail_id_GET.json` | `/api/dataset/detail/:id` | GET | dataset |
| `baselines/chart/chart_create_POST.json` | `/api/chart/create` | POST | chart |
| `baselines/chart/chart_view_chartId_POST.json` | `/api/chart/view/:chartId` | POST | chart |
| `baselines/export/export_pdf_panel_id_GET.json` | `/api/export/pdf/panel/:id` | GET | export |
| `baselines/user/user_profile_GET.json` | `/api/user/profile` | GET | user |
| `baselines/share/share_create_POST.json` | `/api/share/create` | POST | share |
| `baselines/org/org_list_GET.json` | `/api/org/list` | GET | org |

## Validation Rules

1. **Capability must be valid** - Must match one of the defined capability groups
2. **Method must be uppercase** - GET, POST, PUT, DELETE only
3. **Filename must end with .json** - All baseline fixtures are JSON
4. **Path params use underscore prefix** - e.g., `_id`, `_chartId`
5. **No consecutive underscores** - Path segments never result in `__`
