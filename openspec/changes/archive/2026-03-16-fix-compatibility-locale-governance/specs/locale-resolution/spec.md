## ADDED Requirements

### Requirement: Compatibility Locale Normalization
The system SHALL normalize inbound locale signals for compatibility-facing responses into the supported locale set: `zh-CN`, `en`, and `tw`.

#### Scenario: Normalize English locale header
- **WHEN** a compatibility request carries `Accept-Language` values such as `en`, `en-US`, or `en-GB`
- **THEN** the effective locale MUST be normalized to `en`
- **AND** localized compatibility fields MUST be rendered in English

#### Scenario: Normalize Traditional Chinese locale header
- **WHEN** a compatibility request carries `Accept-Language` values such as `zh-TW` or `zh-HK`
- **THEN** the effective locale MUST be normalized to `tw`
- **AND** localized compatibility fields MUST be rendered in Traditional Chinese

### Requirement: Deterministic Compatibility Locale Fallback
The system SHALL resolve a deterministic effective locale for compatibility responses using governed precedence and fallback behavior.

#### Scenario: Use explicit request locale when supported
- **WHEN** a compatibility request provides a supported `Accept-Language` value
- **THEN** the response MUST use that normalized locale as the effective locale
- **AND** lower-priority fallback sources MUST NOT override it

#### Scenario: Fall back when request locale is absent or unsupported
- **WHEN** a compatibility request omits `Accept-Language` or provides an unsupported locale value
- **THEN** the system MUST fall back to the normalized user language preference when available
- **AND** MUST fall back to `zh-CN` when no supported user language preference is available
