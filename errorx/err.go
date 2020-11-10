package errorx

import "fmt"

type ServiceError struct {
	ErrCode string
	ErrMsg  string
}

func (se *ServiceError) Error() string {
	return se.ErrMsg
}

func (se *ServiceError) Add(msg string) *ServiceError {
	err := &ServiceError{
		ErrCode: se.ErrCode,
		ErrMsg:  se.ErrMsg + fmt.Sprintf(msg),
	}
	return err
}
