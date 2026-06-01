package luogusdk

import (
	"fmt"
	"io"
	"net/http"
)

// AuthService 认证服务
type AuthService struct {
	client *Client
}

// RefreshCSRF 刷新 CSRF token（需要先调用才能登录）
func (a *AuthService) RefreshCSRF() error {
	return a.client.refreshCSRF()
}

// GetCaptcha 获取验证码图片，返回 JPEG 字节
func (a *AuthService) GetCaptcha() ([]byte, error) {
	resp, err := a.client.get("/lg4/captcha")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

// Login 使用用户名、密码、验证码登录
func (a *AuthService) Login(username, password, captcha string) (*LoginResponse, error) {
	body := &LoginRequest{
		Username: username,
		Password: password,
		Captcha:  captcha,
	}

	resp, err := a.client.post("/do-auth/password", body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		}
		if parseErr := parseBody(resp, &errResp); parseErr == nil && errResp.Message != "" {
			return nil, &AuthError{Code: errResp.Code, Message: errResp.Message}
		}
		return nil, &AuthError{Code: resp.StatusCode, Message: "login failed"}
	}

	var result LoginResponse
	if err := parseBody(resp, &result); err != nil {
		return nil, err
	}

	if err := a.client.saveCookiesToFile(); err != nil {
		return nil, fmt.Errorf("save cookies: %w", err)
	}

	return &result, nil
}

// LoginWithSolver 使用 CaptchaSolver 自动获取验证码并登录
func (a *AuthService) LoginWithSolver(username, password string, solver CaptchaSolver) (*LoginResponse, error) {
	image, err := a.GetCaptcha()
	if err != nil {
		return nil, err
	}

	captcha, err := solver(image)
	if err != nil {
		return nil, &AuthError{Code: 0, Message: "captcha solve failed: " + err.Error()}
	}

	return a.Login(username, password, captcha)
}

// Logout 登出当前会话
func (a *AuthService) Logout() error {
	resp, err := a.client.post("/auth/logout", nil)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

// Verify 验证当前登录状态是否有效
func (a *AuthService) Verify() error {
	return a.client.verifyAuth()
}

// IsAuthenticated 检查是否已经登录
func (a *AuthService) IsAuthenticated() bool {
	return a.client.verifyAuth() == nil
}

// SaveCookies 手动持久化当前 cookie 到文件
func (a *AuthService) SaveCookies() error {
	return a.client.saveCookiesToFile()
}

// CookiePath 返回 cookie 持久化文件的路径
func (a *AuthService) CookiePath() string {
	return a.client.cookieFile
}
