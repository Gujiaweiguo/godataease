package errno

// ResultCode 对齐 Java 侧 io.dataease.result.ResultCode
// 保持与 Java 侧完全一致的错误码定义
type ResultCode struct {
	Code    int
	Message string
}

// 成功状态码
var SUCCESS = ResultCode{Code: 0, Message: "success"}

// 参数错误：10001-19999
var (
	PARAM_IS_INVALID   = ResultCode{Code: 10001, Message: "参数无效"}
	PARAM_IS_BLANK     = ResultCode{Code: 10002, Message: "参数为空"}
	PARAM_TYPE_ERROR   = ResultCode{Code: 10003, Message: "参数类型错误"}
	PARAM_NOT_COMPLETE = ResultCode{Code: 10004, Message: "参数缺失"}
)

// 用户错误：20001-29999
var (
	USER_NOT_LOGGED_IN     = ResultCode{Code: 20001, Message: "用户未登录"}
	USER_LOGIN_ERROR       = ResultCode{Code: 20002, Message: "账号不存在或密码错误"}
	USER_ACCOUNT_FORBIDDEN = ResultCode{Code: 20003, Message: "账号已被禁用"}
	USER_NOT_EXIST         = ResultCode{Code: 20004, Message: "用户不存在"}
	USER_HAS_EXISTED       = ResultCode{Code: 20005, Message: "用户已存在"}
)

// 业务错误：30001-39999
var (
	BUSINESS_ERROR = ResultCode{Code: 30001, Message: "业务处理异常"}
)

// 系统错误：40001-49999
var (
	SYSTEM_INNER_ERROR = ResultCode{Code: 40001, Message: "系统错误"}
)

// 数据错误：50001-59999
var (
	RESULE_DATA_NONE      = ResultCode{Code: 50001, Message: "数据未找到"}
	DATA_IS_WRONG         = ResultCode{Code: 50002, Message: "数据有误"}
	DATA_ALREADY_EXISTED  = ResultCode{Code: 50003, Message: "数据已存在"}
	DS_RESOURCE_UNCHECKED = ResultCode{Code: 50004, Message: "资源被使用，无法删除"}
	DV_RESOURCE_UNCHECKED = ResultCode{Code: 50005, Message: "仪表板或大屏正在使用，无法删除"}
)

// 接口错误：60001-69999
var (
	INTERFACE_INNER_INVOKE_ERROR = ResultCode{Code: 60001, Message: "内部系统接口调用异常"}
	INTERFACE_OUTER_INVOKE_ERROR = ResultCode{Code: 60002, Message: "外部系统接口调用异常"}
	INTERFACE_FORBID_VISIT       = ResultCode{Code: 60003, Message: "该接口禁止访问"}
	INTERFACE_ADDRESS_INVALID    = ResultCode{Code: 60004, Message: "接口地址无效"}
	INTERFACE_REQUEST_TIMEOUT    = ResultCode{Code: 60005, Message: "接口请求超时"}
	INTERFACE_EXCEED_LOAD        = ResultCode{Code: 60006, Message: "接口负载过高"}
)

// 权限错误：70001-79999
var (
	PERMISSION_NO_ACCESS = ResultCode{Code: 70001, Message: "无访问权限"}
)

// 用户配额错误
var (
	USER_NO_QUOTA = ResultCode{Code: 80001, Message: "没有用户配额"}
)

// GetCode 获取错误码
func (r ResultCode) GetCode() int {
	return r.Code
}

// GetMessage 获取错误消息
func (r ResultCode) GetMessage() string {
	return r.Message
}

// IsSuccess 判断是否成功
func IsSuccess(code int) bool {
	return code == SUCCESS.Code
}
