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

	// 2. 检查是否已经登录
	if client.Auth.IsAuthenticated() {
		fmt.Println("已登录（通过持久化 cookie 恢复）")
	} else {
		fmt.Println("未登录，开始登录流程...")

		// 3. 刷新 CSRF token
		if err := client.Auth.RefreshCSRF(); err != nil {
			fmt.Printf("获取 CSRF token 失败: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("✓ CSRF token 获取成功")

		// 4. 使用内置 OCR 自动识别验证码并登录
		// 需要提前安装: pip install ddddocr
		result, err := client.Auth.LoginWithSolver(username, password, luoguSDK.DDDDOCRSolver())
		if err != nil {
			fmt.Printf("登录失败: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✓ 登录成功 (UID: %d)\n", result.UID)
	}

	// 5. 获取题目示例
	fmt.Println("\n--- 获取题目 P1001 ---")
	problem, err := client.Problem.Get("P1001")
	if err != nil {
		fmt.Printf("获取题目失败: %v\n", err)
	} else {
		fmt.Printf("标题: %s\n", problem.Title)
		fmt.Printf("难度: %d\n", problem.Difficulty)
		fmt.Printf("时间限制: %dms\n", problem.TimeLimit)
		fmt.Printf("内存限制: %dKB\n", problem.MemoryLimit)
		if len(problem.Tags) > 0 {
			fmt.Print("标签: ")
			for i, tag := range problem.Tags {
				if i > 0 {
					fmt.Print(", ")
				}
				fmt.Print(tag.Name)
			}
			fmt.Println()
		}
	}

	// 6. 搜索题目示例
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
