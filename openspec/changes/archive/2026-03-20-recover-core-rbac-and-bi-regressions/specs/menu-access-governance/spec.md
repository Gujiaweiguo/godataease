## ADDED Requirements

### Requirement: Core Feature Recovery Matrix
The system and verification process SHALL maintain a recovery matrix for core RBAC and BI feature domains during regression remediation.

#### Scenario: Classify feature-loss symptom
- **WHEN** a core feature is reported as missing or inaccessible
- **THEN** the recovery process MUST classify the issue as route loss, menu loss, permission mismatch, API mismatch, page initialization failure, or real implementation gap
- **AND** the classification MUST be traceable to concrete file or endpoint evidence
