package errno

import (
	"testing"
)

func TestResultCode_GetCode(t *testing.T) {
	code := SUCCESS.GetCode()
	if code != 0 {
		t.Errorf("Expected SUCCESS code 0, got %d", code)
	}

	code = PARAM_IS_INVALID.GetCode()
	if code != 10001 {
		t.Errorf("Expected PARAM_IS_INVALID code 10001, got %d", code)
	}
}

func TestResultCode_GetMessage(t *testing.T) {
	msg := SUCCESS.GetMessage()
	if msg != "success" {
		t.Errorf("Expected SUCCESS message 'success', got '%s'", msg)
	}

	msg = USER_NOT_LOGGED_IN.GetMessage()
	if msg != "用户未登录" {
		t.Errorf("Expected USER_NOT_LOGGED_IN message '用户未登录', got '%s'", msg)
	}
}

func TestIsSuccess(t *testing.T) {
	if !IsSuccess(0) {
		t.Error("IsSuccess(0) should return true")
	}

	if IsSuccess(10001) {
		t.Error("IsSuccess(10001) should return false")
	}

	if IsSuccess(-1) {
		t.Error("IsSuccess(-1) should return false")
	}
}

func TestParamErrorCodes(t *testing.T) {
	tests := []struct {
		name     string
		code     ResultCode
		expected int
	}{
		{"PARAM_IS_INVALID", PARAM_IS_INVALID, 10001},
		{"PARAM_IS_BLANK", PARAM_IS_BLANK, 10002},
		{"PARAM_TYPE_ERROR", PARAM_TYPE_ERROR, 10003},
		{"PARAM_NOT_COMPLETE", PARAM_NOT_COMPLETE, 10004},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.code.Code != tt.expected {
				t.Errorf("%s: expected code %d, got %d", tt.name, tt.expected, tt.code.Code)
			}
		})
	}
}

func TestUserErrorCodes(t *testing.T) {
	tests := []struct {
		name     string
		code     ResultCode
		expected int
	}{
		{"USER_NOT_LOGGED_IN", USER_NOT_LOGGED_IN, 20001},
		{"USER_LOGIN_ERROR", USER_LOGIN_ERROR, 20002},
		{"USER_ACCOUNT_FORBIDDEN", USER_ACCOUNT_FORBIDDEN, 20003},
		{"USER_NOT_EXIST", USER_NOT_EXIST, 20004},
		{"USER_HAS_EXISTED", USER_HAS_EXISTED, 20005},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.code.Code != tt.expected {
				t.Errorf("%s: expected code %d, got %d", tt.name, tt.expected, tt.code.Code)
			}
		})
	}
}

func TestBusinessErrorCode(t *testing.T) {
	if BUSINESS_ERROR.Code != 30001 {
		t.Errorf("Expected BUSINESS_ERROR code 30001, got %d", BUSINESS_ERROR.Code)
	}
}

func TestSystemErrorCode(t *testing.T) {
	if SYSTEM_INNER_ERROR.Code != 40001 {
		t.Errorf("Expected SYSTEM_INNER_ERROR code 40001, got %d", SYSTEM_INNER_ERROR.Code)
	}
}

func TestDataErrorCodes(t *testing.T) {
	tests := []struct {
		name     string
		code     ResultCode
		expected int
	}{
		{"RESULE_DATA_NONE", RESULE_DATA_NONE, 50001},
		{"DATA_IS_WRONG", DATA_IS_WRONG, 50002},
		{"DATA_ALREADY_EXISTED", DATA_ALREADY_EXISTED, 50003},
		{"DS_RESOURCE_UNCHECKED", DS_RESOURCE_UNCHECKED, 50004},
		{"DV_RESOURCE_UNCHECKED", DV_RESOURCE_UNCHECKED, 50005},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.code.Code != tt.expected {
				t.Errorf("%s: expected code %d, got %d", tt.name, tt.expected, tt.code.Code)
			}
		})
	}
}

func TestInterfaceErrorCodes(t *testing.T) {
	tests := []struct {
		name     string
		code     ResultCode
		expected int
	}{
		{"INTERFACE_INNER_INVOKE_ERROR", INTERFACE_INNER_INVOKE_ERROR, 60001},
		{"INTERFACE_OUTER_INVOKE_ERROR", INTERFACE_OUTER_INVOKE_ERROR, 60002},
		{"INTERFACE_FORBID_VISIT", INTERFACE_FORBID_VISIT, 60003},
		{"INTERFACE_ADDRESS_INVALID", INTERFACE_ADDRESS_INVALID, 60004},
		{"INTERFACE_REQUEST_TIMEOUT", INTERFACE_REQUEST_TIMEOUT, 60005},
		{"INTERFACE_EXCEED_LOAD", INTERFACE_EXCEED_LOAD, 60006},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.code.Code != tt.expected {
				t.Errorf("%s: expected code %d, got %d", tt.name, tt.expected, tt.code.Code)
			}
		})
	}
}

func TestPermissionErrorCode(t *testing.T) {
	if PERMISSION_NO_ACCESS.Code != 70001 {
		t.Errorf("Expected PERMISSION_NO_ACCESS code 70001, got %d", PERMISSION_NO_ACCESS.Code)
	}
}

func TestQuotaErrorCode(t *testing.T) {
	if USER_NO_QUOTA.Code != 80001 {
		t.Errorf("Expected USER_NO_QUOTA code 80001, got %d", USER_NO_QUOTA.Code)
	}
}
