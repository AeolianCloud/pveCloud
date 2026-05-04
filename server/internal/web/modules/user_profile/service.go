package userprofile

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/AeolianCloud/pveCloud/server/internal/admin/models"
	apperrors "github.com/AeolianCloud/pveCloud/server/internal/shared/errors"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/password"
	webdto "github.com/AeolianCloud/pveCloud/server/internal/web/dto"
	websupport "github.com/AeolianCloud/pveCloud/server/internal/web/support"
)

/**
 * UserProfileService 处理当前登录用户的资料编辑和密码修改。
 */
type UserProfileService struct {
	db *gorm.DB
}

/**
 * NewUserProfileService 创建用户资料服务。
 */
func NewUserProfileService(db *gorm.DB) *UserProfileService {
	return &UserProfileService{db: db}
}

/**
 * UpdateProfile 更新当前登录用户的基础资料。
 */
func (s *UserProfileService) UpdateProfile(ctx context.Context, userID uint64, sessionID string, req webdto.UpdateProfileRequest) (webdto.AuthStateResponse, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))
	displayName := trimOptional(req.DisplayName)
	if email == "" {
		return webdto.AuthStateResponse{}, apperrors.ErrValidation.WithMessage("邮箱不能为空")
	}

	var result webdto.AuthStateResponse
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var user models.User
		if err := tx.Where("id = ?", userID).First(&user).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return apperrors.ErrUnauthorized
			}
			return err
		}
		if user.Status != "active" {
			return apperrors.ErrForbidden.WithMessage("用户账号已被禁用")
		}
		var existing models.User
		if err := tx.Where("email = ? AND id <> ?", email, userID).First(&existing).Error; err == nil {
			return apperrors.ErrConflict.WithMessage("邮箱已被使用")
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if err := tx.Model(&models.User{}).
			Where("id = ?", userID).
			Updates(map[string]interface{}{"email": email, "display_name": displayName}).Error; err != nil {
			if isDuplicateEntry(err) {
				return apperrors.ErrConflict.WithMessage("邮箱已被使用")
			}
			return err
		}
		if err := tx.Where("id = ?", userID).First(&user).Error; err != nil {
			return err
		}
		var session models.UserSession
		if err := tx.Where("session_id = ? AND user_id = ? AND status = ?", sessionID, userID, "active").First(&session).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return apperrors.ErrUnauthorized
			}
			return err
		}
		result = webdto.AuthStateResponse{
			User:    websupport.UserSummary(user),
			Session: websupport.SessionSummary(session),
		}
		return nil
	})
	return result, err
}

/**
 * ChangePassword 修改当前登录用户密码。
 */
func (s *UserProfileService) ChangePassword(ctx context.Context, userID uint64, sessionID string, req webdto.ChangePasswordRequest) error {
	now := time.Now()
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var user models.User
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", userID).
			First(&user).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return apperrors.ErrUnauthorized
			}
			return err
		}
		if user.Status != "active" {
			return apperrors.ErrForbidden.WithMessage("用户账号已被禁用")
		}
		var session models.UserSession
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("session_id = ? AND user_id = ? AND status = ?", sessionID, userID, "active").
			First(&session).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return apperrors.ErrUnauthorized
			}
			return err
		}
		if !session.ExpiresAt.After(now) {
			return apperrors.ErrUnauthorized
		}
		if !password.Verify(user.PasswordHash, req.CurrentPassword) {
			return apperrors.ErrValidation.WithMessage("当前密码错误")
		}
		newHash, err := password.Hash(req.Password)
		if err != nil {
			return err
		}
		if err := tx.Model(&models.User{}).
			Where("id = ?", userID).
			Update("password_hash", newHash).Error; err != nil {
			return err
		}
		return tx.Model(&models.UserSession{}).
			Where("user_id = ? AND status = ? AND session_id <> ?", userID, "active", sessionID).
			Updates(map[string]interface{}{"status": "revoked", "revoked_at": now, "revoke_reason": "password_change"}).Error
	})
}

func trimOptional(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func isDuplicateEntry(err error) bool {
	var mysqlErr *mysql.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}
