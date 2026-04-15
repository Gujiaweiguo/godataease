## ADDED Requirements

### Requirement: Migrate Dataset Query Routes to Canonical Paths
The system SHALL provide 10 canonical dataset query endpoints under `/api/dataset/*` that replace legacy compatibility routes.

#### Migrated Routes
| Canonical Route | Method | Service Method |
|---|---|---|
| /api/dataset/get/:id | POST | buildDatasetDetail |
| /api/dataset/details/:id | POST | buildDatasetDetail |
| /api/dataset/dsDetails | POST | buildDatasetDetail (batch) |
| /api/dataset/getSqlParams | POST | GetSQLParams |
| /api/dataset/barInfo/:id | GET | GetGroupByID |
| /api/dataset/getDatasetTotal | POST | Preview (limit=1) |
| /api/dataset/previewSql | POST | PreviewSQLWithUser |
| /api/dataset/enumValueObj | POST | GetFieldEnumObj |
| /api/dataset/enumValueDs | POST | GetFieldEnumDs |
| /api/dataset/enumValue | POST | GetFieldEnum |

#### Skipped Routes
- detailWithPerm: requires permission middleware
- exportDataset: requires chart export service

#### Frontend Changes
All frontend API functions updated to use canonical `/dataset/` paths.
All compatibility routes preserved for backward compatibility.
