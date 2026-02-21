// tools/genpwd/main.go
// 一次性工具：生成 bcrypt 密码哈希，用于手动插入初始管理员账号。
// 用法：go run tools/genpwd/main.go 你的密码
package main

import (
	"fmt"
	"os"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("用法: go run tools/genpwd/main.go <密码>")
		os.Exit(1)
	}
	pwd := os.Args[1]
	hash, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	if err != nil {
		fmt.Fprintf(os.Stderr, "生成失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(hash))
}
