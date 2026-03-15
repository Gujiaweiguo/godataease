## 1. Analysis

- [x] 1.1 分析 threshold 表结构
- [x] 1.2 Identify threshold fields that reference view IDs (`chart_id`)
- [x] 1.3 Check existing threshold sync/delete methods in `visualization_threshold_repo.go`

## 2. Implementation

- [x] 2.1 Add `CopyThresholdWithMap()` method
- [x] 2.2 Add helper function for view ID rewriting (`chart_id` → viewMap lookup)
- [x] 2.3 Add `isTableNotExistError()` helper to gracefully handle missing xpack tables
- [x] 2.4 Integrate threshold copy into `CreateCopyWithChildren()`

## 3. Testing

- [x] 3.1 Run full visualization copy test with thresholds
- [x] 3.2 Verify threshold copy handles missing table gracefully (no error when xpack_threshold_info doesn't exist)
