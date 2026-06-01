package luogusdk

import "fmt"

// AuthError 登录/认证失败
type AuthError struct {
	Code    int
	Message string
}

func (e *AuthError) Error() string {
	return fmt.Sprintf("auth error [%d]: %s", e.Code, e.Message)
}

// CSRFError CSRF token 获取或过期
type CSRFError struct {
	Err error
}

func (e *CSRFError) Error() string {
	return fmt.Sprintf("csrf error: %v", e.Err)
}

func (e *CSRFError) Unwrap() error {
	return e.Err
}

// NetworkError 网络请求失败
type NetworkError struct {
	Err error
}

func (e *NetworkError) Error() string {
	return fmt.Sprintf("network error: %v", e.Err)
}

func (e *NetworkError) Unwrap() error {
	return e.Err
}

// UnauthorizedError 未登录调用需认证的 API
type UnauthorizedError struct{}

func (e *UnauthorizedError) Error() string {
	return "unauthorized: please login first"
}
