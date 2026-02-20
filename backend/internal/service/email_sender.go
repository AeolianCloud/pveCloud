package service

import (
	"context"
	"fmt"
)

// EmailSender 定义发送邮件的最小能力，用于用户注册验证邮件通知。
type EmailSender interface {
	SendVerificationEmail(ctx context.Context, email string, verifyToken string) error
}

// ConsoleEmailSender 使用日志输出模拟邮件发送，便于无 SMTP 环境下开发联调。
type ConsoleEmailSender struct{}

// SendVerificationEmail 输出验证链接到日志，生产环境可替换为真实 SMTP/邮件服务。
func (ConsoleEmailSender) SendVerificationEmail(_ context.Context, email string, verifyToken string) error {
	fmt.Printf("send verification email to %s with token=%s\n", email, verifyToken)
	return nil
}
