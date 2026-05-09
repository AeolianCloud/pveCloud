package userprofile

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"

	domainuser "github.com/AeolianCloud/pveCloud/server/internal/domain/user"
	mysqltx "github.com/AeolianCloud/pveCloud/server/internal/repository/mysql/tx"
	mysqluser "github.com/AeolianCloud/pveCloud/server/internal/repository/mysql/user"
	apperrors "github.com/AeolianCloud/pveCloud/server/internal/shared/errors"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/password"
	webdto "github.com/AeolianCloud/pveCloud/server/internal/usecase/web/dto"
	websupport "github.com/AeolianCloud/pveCloud/server/internal/usecase/web/support"
)

/**
 * UserProfileService 处理当前登录用户的资料编辑和密码修改。
 */
type UserProfileService struct {
	db    *gorm.DB
	users *mysqluser.Repository
}

/**
 * NewUserProfileService 创建用户资料服务。
 */
func NewUserProfileService(db *gorm.DB) *UserProfileService {
	return &UserProfileService{db: db, users: mysqluser.NewRepository(db)}
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
	err := mysqltx.NewManager(s.db).WithinContext(ctx, func(tx *gorm.DB) error {
		user, err := s.users.FindUserByID(ctx, tx, userID)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrUnauthorized
		}
		if err != nil {
			return err
		}
		if !domainuser.IsActive(user.Status) {
			return apperrors.ErrForbidden.WithMessage("用户账号已被禁用")
		}
		exists, err := s.users.EmailExistsForOtherUser(ctx, tx, email, userID)
		if err != nil {
			return err
		}
		if exists {
			return apperrors.ErrConflict.WithMessage("邮箱已被使用")
		}
		if err := s.users.UpdateUser(ctx, tx, userID, map[string]any{"email": email, "display_name": displayName}); err != nil {
			if isDuplicateEntry(err) {
				return apperrors.ErrConflict.WithMessage("邮箱已被使用")
			}
			return err
		}
		user, err = s.users.FindUserByID(ctx, tx, userID)
		if err != nil {
			return err
		}
		session, err := s.users.FindUserSessionBySessionID(ctx, sessionID, userID)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrUnauthorized
		}
		if err != nil {
			return err
		}
		if session.Status != domainuser.SessionStatusActive {
			return apperrors.ErrUnauthorized
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
	return mysqltx.NewManager(s.db).WithinContext(ctx, func(tx *gorm.DB) error {
		user, err := s.users.FindUserByIDForUpdate(ctx, tx, userID)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrUnauthorized
		}
		if err != nil {
			return err
		}
		if !domainuser.IsActive(user.Status) {
			return apperrors.ErrForbidden.WithMessage("用户账号已被禁用")
		}
		session, err := s.users.FindActiveUserSessionForUpdate(ctx, tx, sessionID, userID, domainuser.SessionStatusActive)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrUnauthorized
		}
		if err != nil {
			return err
		}
		if !domainuser.IsSessionActiveAt(session.Status, session.ExpiresAt, now) {
			return apperrors.ErrUnauthorized
		}
		if !password.Verify(user.PasswordHash, req.CurrentPassword) {
			return apperrors.ErrValidation.WithMessage("当前密码错误")
		}
		newHash, err := password.Hash(req.Password)
		if err != nil {
			return err
		}
		if err := s.users.UpdateUserPasswordHash(ctx, tx, userID, newHash); err != nil {
			return err
		}
		return s.users.RevokeOtherActiveUserSessions(ctx, tx, userID, sessionID, now, "password_change", domainuser.SessionStatusActive, domainuser.SessionStatusRevoked)
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
