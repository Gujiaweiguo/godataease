## MODIFIED Requirements

### Requirement: Column Permission Control

系统 SHALL 在 Go 实现中保持与 Java 版本完全一致的列级权限控制。

#### Scenario: Permission center exposes keep-middle column desensitization
- **WHEN** an administrator configures column permission masking from the unified permission center
- **THEN** the UI MUST expose the existing keep-middle desensitization rule supported by backend masking
- **AND** saving and reloading that rule MUST preserve the selected keep-middle semantics without degrading to another mask rule
