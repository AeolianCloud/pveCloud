package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"strings"

	"gorm.io/gorm"

	"github.com/AeolianCloud/pveCloud/server/internal/bootstrap"
	"github.com/AeolianCloud/pveCloud/server/internal/models"
	"github.com/AeolianCloud/pveCloud/server/internal/pkg/password"
)

func main() {
	configPath := flag.String("config", "config.yaml", "YAML 配置文件路径")
	username := flag.String("username", "", "管理员用户名")
	email := flag.String("email", "", "管理员邮箱，可选")
	displayName := flag.String("display-name", "超级管理员", "管理员显示名称")
	plainPassword := flag.String("password", "", "管理员密码")
	flag.Parse()

	if strings.TrimSpace(*username) == "" {
		log.Fatal("管理员用户名不能为空")
	}

	passwordValue := strings.TrimSpace(*plainPassword)
	if passwordValue == "" {
		log.Fatal("管理员密码不能为空，请通过 -password 传入")
	}
	if len(passwordValue) < 6 || len(passwordValue) > 72 {
		log.Fatal("管理员密码长度必须在 6 到 72 个字符之间")
	}

	ctx := context.Background()
	app, err := bootstrap.NewApp(ctx, *configPath)
	if err != nil {
		log.Fatalf("初始化应用失败：%v", err)
	}

	if err := createAdmin(ctx, app.DB, *username, *email, *displayName, passwordValue); err != nil {
		log.Fatalf("创建管理员失败：%v", err)
	}

	log.Printf("管理员 %s 创建成功", *username)
}

func createAdmin(ctx context.Context, db *gorm.DB, username string, email string, displayName string, plainPassword string) error {
	username = strings.TrimSpace(username)
	email = strings.TrimSpace(email)
	displayName = strings.TrimSpace(displayName)
	if displayName == "" {
		displayName = username
	}

	hash, err := password.Hash(plainPassword)
	if err != nil {
		return err
	}

	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing models.AdminUser
		err := tx.Where("username = ?", username).First(&existing).Error
		if err == nil {
			return errors.New("管理员用户名已存在")
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		var emailPtr *string
		if email != "" {
			err = tx.Where("email = ?", email).First(&existing).Error
			if err == nil {
				return errors.New("管理员邮箱已存在")
			}
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			emailPtr = &email
		}

		admin := models.AdminUser{
			Username:     username,
			Email:        emailPtr,
			PasswordHash: hash,
			DisplayName:  displayName,
			Status:       "active",
		}
		if err := tx.Create(&admin).Error; err != nil {
			return err
		}

		var role models.AdminRole
		if err := tx.Where("code = ?", "super_admin").First(&role).Error; err != nil {
			return err
		}

		// 首个管理员默认绑定超级管理员角色，权限明细仍以数据库 RBAC 关系为准。
		return tx.Exec(
			"INSERT INTO admin_user_roles (admin_id, role_id) VALUES (?, ?) ON DUPLICATE KEY UPDATE admin_id = VALUES(admin_id)",
			admin.ID,
			role.ID,
		).Error
	})
}
