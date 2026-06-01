package luogusdk

import (
	"math"
	"net"
	"time"
)

// defaultBackoff 默认指数退避: 1s, 2s, 4s, ...
func defaultBackoff(attempt int) time.Duration {
	return time.Duration(math.Pow(2, float64(attempt))) * time.Second
}

// shouldRetry 判断错误是否应该重试（仅网络错误，不重试业务错误）
func shouldRetry(err error) bool {
	if err == nil {
		return false
	}
	if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
		return true
	}
	return false
}

// shouldRetryStatus 判断 HTTP 状态码是否应该重试（仅 5xx）
func shouldRetryStatus(statusCode int) bool {
	return statusCode >= 500
}
