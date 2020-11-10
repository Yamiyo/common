package errorx

var (
	SystemErr = &ServiceError{
		ErrMsg:  "系統異常",
		ErrCode: "10000",
	}
	AccessErr = &ServiceError{
		ErrMsg:  "Access denied",
		ErrCode: "10001",
	}
	VerifyErr = &ServiceError{
		ErrMsg:  "參數錯誤: ",
		ErrCode: "20000",
	}
)
