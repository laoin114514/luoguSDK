package main

import (
	"fmt"
	"os"

	luoguSDK "github.com/laoin114514/luoguSDK"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("用法: go run . <用户名> <密码>")
		fmt.Println("示例: go run . myuser mypassword")
		os.Exit(1)
	}
	username := os.Args[1]
	password := os.Args[2]

	// 1. 创建客户端（自动加载持久化的 cookie，若有效则跳过登录）
	client, err := luoguSDK.NewClient()
	if err != nil {
		fmt.Printf("创建客户端失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Cookie 文件路径: %s\n", client.Auth.CookiePath())

	// 2. 检查是否已经登录
	if client.Auth.IsAuthenticated() {
		fmt.Println("已登录（通过持久化 cookie 恢复）")
	} else {
		fmt.Println("未登录，开始登录流程...")

		if err := client.Auth.RefreshCSRF(); err != nil {
			fmt.Printf("获取 CSRF token 失败: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("✓ CSRF token 获取成功")

		result, err := client.Auth.LoginWithSolver(username, password, captchaByHand)
		if err != nil {
			fmt.Printf("登录失败: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✓ 登录成功\n")
		_ = result
	}

	// 3. 获取题目示例
	fmt.Println("\n--- 获取题目 P1001 ---")
	problem, err := client.Problem.Get("P1001")
	if err != nil {
		fmt.Printf("获取题目失败: %v\n", err)
	} else {
		fmt.Printf("标题: %s\n", problem.Title)
		fmt.Printf("难度: %d\n", problem.Difficulty)
		fmt.Printf("描述: %s\n", problem.DescText())
		fmt.Printf("输入格式: %s\n", problem.InputText())
		fmt.Printf("输出格式: %s\n", problem.OutputText())
		fmt.Printf("时间限制: %dms\n", problem.TimeLimit())
		fmt.Printf("内存限制: %dKB\n", problem.MemoryLimit())
		if len(problem.Tags) > 0 {
			fmt.Printf("标签ID: %v\n", problem.Tags)
		}
	}

	// 4. 搜索题目示例
	fmt.Println("\n--- 搜索题目 (关键词: 排序) ---")
	results, err := client.Problem.Search(luoguSDK.SearchParams{
		Keyword:  "排序",
		Page:     1,
		PageSize: 5,
	})
	if err != nil {
		fmt.Printf("搜索失败: %v\n", err)
	} else {
		fmt.Printf("共 %d 个结果，当前第 %d 页:\n", results.Total, results.Page)
		for _, p := range results.Problems {
			fmt.Printf("  %s - %s (难度: %d)\n", p.PID, p.Title, p.Difficulty)
		}
	}
}
func captchaByHand(image []byte) (string, error) {
	if err := os.WriteFile("captcha.jpg", image, 0644); err != nil {
		return "", fmt.Errorf("保存验证码图片失败: %w", err)
	}
	fmt.Print("验证码已保存到 captcha.jpg，请输入验证码: ")
	var captcha string
	if _, err := fmt.Scan(&captcha); err != nil {
		return "", fmt.Errorf("读取验证码输入失败: %w", err)
	}
	return captcha, nil
}
