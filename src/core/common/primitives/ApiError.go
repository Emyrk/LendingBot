package primitives

import (
	"fmt"
)

type ApiError struct {
	LogError  error
	UserError error
}

func (a *ApiError) Error() string {
	return a.LogError.Error()
}

func NewAPIError(logError error, userError error) *ApiError {
	apiError := new(ApiError)
	apiError.LogError = logError
	apiError.UserError = userError
	return apiError
}

func NewAPIErrorFromOne(err error) *ApiError {
	apiError := new(ApiError)
	apiError.LogError = err
	apiError.UserError = err
	return apiError
}

func NewAPIErrorInternalError(err error) *ApiError {
	apiError := new(ApiError)
	apiError.LogError = err
	apiError.UserError = fmt.Errorf("Internal Error. Please contact support at: support@hodl.zone")
	return apiError
}
