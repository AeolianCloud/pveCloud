package adminuser

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	admindto "github.com/AeolianCloud/pveCloud/server/internal/admin/dto"
	"github.com/AeolianCloud/pveCloud/server/internal/admin/models"
	"github.com/AeolianCloud/pveCloud/server/internal/admin/support"
	apperrors "github.com/AeolianCloud/pveCloud/server/internal/shared/errors"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/password"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/sets"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/textutil"
)

const (
	adminUserObjectType          = "admin_user"
	adminUserCreateAction        = "admin.user.create"
	adminUserUpdateAction        = "admin.user.update"
	adminUserDisableAction       = "admin.user.disable"
	adminUserRoleUpdateAction    = "admin.user.role_update"
	adminUserPasswordResetAction = "admin.user.password_reset"
)

/**
 * AdminUserService 处理管理员账号管理。
 */
type AdminUserService struct {
	db           *gorm.DB
	auditService *AdminAuditService
}

/**
 * NewAdminUserService 创建管理员账号服务。
 *
 * @param db 数据库连接
 * @param auditService 后台审计服务
 * @return *AdminUserService 管理员账号服务
 */
func NewAdminUserService(db *gorm.DB, auditService *AdminAuditService) *AdminUserService {
	if auditService == nil {
		auditService = NewAdminAuditService(db)
	}
	return &AdminUserService{db: db, auditService: auditService}
}

/**
 * List 分页查询管理员账号。
 *
 * @param ctx 请求上下文
 * @param query 查询参数
 * @return admin.PageResponse[admin.AdminUserItem] 分页结果
 * @return error 查询失败原因
 */
func (s *AdminUserService) List(ctx context.Context, query admindto.AdminUserListQuery) (admindto.PageResponse[admindto.AdminUserItem], error) {
	page, perPage := support.NormalizePage(query.Page, query.PerPage)
	db := s.db.WithContext(ctx).Model(&models.AdminUser{}).Where("deleted_at IS NULL")
	db = applyAdminUserFilters(db, query)

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return admindto.PageResponse[admindto.AdminUserItem]{}, err
	}

	var users []models.AdminUser
	if err := db.Order("id DESC").Limit(perPage).Offset((page - 1) * perPage).Find(&users).Error; err != nil {
		return admindto.PageResponse[admindto.AdminUserItem]{}, err
	}
	roleMap, err := s.roleSummariesByAdminIDs(ctx, adminUserIDs(users))
	if err != nil {
		return admindto.PageResponse[admindto.AdminUserItem]{}, err
	}

	items := make([]admindto.AdminUserItem, 0, len(users))
	for _, user := range users {
		items = append(items, adminUserItem(user, roleMap[user.ID]))
	}
	return support.PageResponse(items, total, page, perPage), nil
}

/**
 * Create 创建管理员账号。
 *
 * @param ctx 请求上下文
 * @param operatorID 当前操作者管理员 ID
 * @param req 创建请求
 * @param clientIP 客户端 IP
 * @param userAgent 浏览器 User-Agent
 * @return admin.AdminUserItem 创建后的管理员账号
 * @return error 创建失败原因
 */
func (s *AdminUserService) Create(ctx context.Context, operatorID uint64, req admindto.AdminUserCreateRequest, clientIP string, userAgent string) (admindto.AdminUserItem, error) {
	email := textutil.NormalizeOptionalString(req.Email)
	if err := s.ensureAdminUserUnique(ctx, 0, req.Username, email); err != nil {
		return admindto.AdminUserItem{}, err
	}
	if err := s.ensureRolesAssignable(ctx, req.RoleIDs); err != nil {
		return admindto.AdminUserItem{}, err
	}
	passwordHash, err := password.Hash(req.Password)
	if err != nil {
		return admindto.AdminUserItem{}, err
	}

	var created models.AdminUser
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		created = models.AdminUser{
			Username:     strings.TrimSpace(req.Username),
			Email:        email,
			PasswordHash: passwordHash,
			DisplayName:  strings.TrimSpace(req.DisplayName),
			Status:       strings.TrimSpace(req.Status),
		}
		if err := tx.Create(&created).Error; err != nil {
			return err
		}
		if err := replaceAdminUserRoles(ctx, tx, created.ID, req.RoleIDs); err != nil {
			return err
		}
		return s.auditService.Record(ctx, tx, AdminAuditWriteInput{
			AdminID:    &operatorID,
			Action:     adminUserCreateAction,
			ObjectType: adminUserObjectType,
			ObjectID:   textutil.Uint64String(created.ID),
			AfterData:  adminUserAuditSnapshot(created, req.RoleIDs),
			IP:         clientIP,
			UserAgent:  userAgent,
			Remark:     "创建管理员账号",
		})
	}); err != nil {
		return admindto.AdminUserItem{}, err
	}

	roles, err := s.roleSummariesByAdminIDs(ctx, []uint64{created.ID})
	if err != nil {
		return admindto.AdminUserItem{}, err
	}
	return adminUserItem(created, roles[created.ID]), nil
}

/**
 * Detail 查询管理员账号详情。
 *
 * @param ctx 请求上下文
 * @param id 管理员 ID
 * @return admin.AdminUserDetail 管理员详情
 * @return error 查询失败原因
 */
func (s *AdminUserService) Detail(ctx context.Context, id uint64) (admindto.AdminUserDetail, error) {
	user, err := s.findAdminUser(ctx, s.db, id)
	if err != nil {
		return admindto.AdminUserDetail{}, err
	}
	roleMap, err := s.roleSummariesByAdminIDs(ctx, []uint64{id})
	if err != nil {
		return admindto.AdminUserDetail{}, err
	}
	codes, err := support.PermissionCodes(ctx, s.db, id)
	if err != nil {
		return admindto.AdminUserDetail{}, err
	}
	sessions, err := s.activeSessionSummaries(ctx, id)
	if err != nil {
		return admindto.AdminUserDetail{}, err
	}

	return admindto.AdminUserDetail{
		AdminUserItem:   adminUserItem(user, roleMap[id]),
		PermissionCodes: codes,
		Sessions:        sessions,
	}, nil
}

/**
 * Update 更新管理员账号资料、状态和角色。
 *
 * @param ctx 请求上下文
 * @param operatorID 当前操作者管理员 ID
 * @param id 管理员 ID
 * @param req 更新请求
 * @param clientIP 客户端 IP
 * @param userAgent 浏览器 User-Agent
 * @return admin.AdminUserItem 更新后的管理员账号
 * @return error 更新失败原因
 */
func (s *AdminUserService) Update(ctx context.Context, operatorID uint64, id uint64, req admindto.AdminUserUpdateRequest, clientIP string, userAgent string) (admindto.AdminUserItem, error) {
	email := textutil.NormalizeOptionalString(req.Email)
	if err := s.ensureAdminUserUnique(ctx, id, "", email); err != nil {
		return admindto.AdminUserItem{}, err
	}
	if req.RoleIDs != nil {
		if err := s.ensureRolesAssignable(ctx, req.RoleIDs); err != nil {
			return admindto.AdminUserItem{}, err
		}
	}
	if req.Status != nil && id == operatorID && *req.Status != support.AdminStatusActive {
		return admindto.AdminUserItem{}, apperrors.ErrConflict.WithMessage("不能禁用当前登录账号")
	}

	var updated models.AdminUser
	var roles []admindto.AdminRoleSummary
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		current, err := s.findAdminUserForUpdate(ctx, tx, id)
		if err != nil {
			return err
		}
		beforeRoleIDs, err := support.RoleIDs(ctx, tx, id)
		if err != nil {
			return err
		}

		updates := map[string]interface{}{}
		if req.Email != nil {
			updates["email"] = email
		}
		if req.DisplayName != nil {
			updates["display_name"] = strings.TrimSpace(*req.DisplayName)
		}
		if req.Status != nil {
			updates["status"] = strings.TrimSpace(*req.Status)
		}
		if len(updates) > 0 {
			if err := tx.Model(&models.AdminUser{}).Where("id = ?", id).Updates(updates).Error; err != nil {
				return err
			}
		}
		if req.RoleIDs != nil {
			if err := replaceAdminUserRoles(ctx, tx, id, req.RoleIDs); err != nil {
				return err
			}
		}
		if err := tx.Where("id = ?", id).First(&updated).Error; err != nil {
			return err
		}
		afterRoleIDs, err := support.RoleIDs(ctx, tx, id)
		if err != nil {
			return err
		}
		action, riskLevel, riskReason := support.AdminUserUpdateRisk(current, updated, beforeRoleIDs, afterRoleIDs)
		input := AdminAuditWriteInput{
			AdminID:    &operatorID,
			Action:     action,
			ObjectType: adminUserObjectType,
			ObjectID:   textutil.Uint64String(id),
			BeforeData: adminUserAuditSnapshot(current, beforeRoleIDs),
			AfterData:  adminUserAuditSnapshot(updated, afterRoleIDs),
			IP:         clientIP,
			UserAgent:  userAgent,
			Remark:     "更新管理员账号",
		}
		if riskLevel != "" {
			return s.auditService.RecordRisk(ctx, tx, AdminRiskWriteInput{
				AdminAuditWriteInput: input,
				RiskLevel:            riskLevel,
				RiskReason:           riskReason,
			})
		}
		return s.auditService.Record(ctx, tx, input)
	}); err != nil {
		return admindto.AdminUserItem{}, err
	}

	roleMap, err := s.roleSummariesByAdminIDs(ctx, []uint64{id})
	if err != nil {
		return admindto.AdminUserItem{}, err
	}
	roles = roleMap[id]
	return adminUserItem(updated, roles), nil
}

/**
 * ResetPassword 重置管理员密码。
 *
 * @param ctx 请求上下文
 * @param operatorID 当前操作者管理员 ID
 * @param id 管理员 ID
 * @param req 密码请求
 * @param clientIP 客户端 IP
 * @param userAgent 浏览器 User-Agent
 * @return error 重置失败原因
 */
func (s *AdminUserService) ResetPassword(ctx context.Context, operatorID uint64, id uint64, req admindto.AdminUserPasswordRequest, clientIP string, userAgent string) error {
	passwordHash, err := password.Hash(req.Password)
	if err != nil {
		return err
	}

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		current, err := s.findAdminUserForUpdate(ctx, tx, id)
		if err != nil {
			return err
		}
		if err := tx.Model(&models.AdminUser{}).Where("id = ?", id).Update("password_hash", passwordHash).Error; err != nil {
			return err
		}
		return s.auditService.RecordRisk(ctx, tx, AdminRiskWriteInput{
			AdminAuditWriteInput: AdminAuditWriteInput{
				AdminID:    &operatorID,
				Action:     adminUserPasswordResetAction,
				ObjectType: adminUserObjectType,
				ObjectID:   textutil.Uint64String(id),
				BeforeData: map[string]any{"id": current.ID, "username": current.Username},
				AfterData:  map[string]any{"id": current.ID, "username": current.Username, "password_reset": true},
				IP:         clientIP,
				UserAgent:  userAgent,
				Remark:     "重置管理员密码",
			},
			RiskLevel:  "high",
			RiskReason: "重置管理员密码",
		})
	})
}

func applyAdminUserFilters(db *gorm.DB, query admindto.AdminUserListQuery) *gorm.DB {
	if query.Keyword != "" {
		keyword := "%" + strings.TrimSpace(query.Keyword) + "%"
		db = db.Where("username LIKE ? OR email LIKE ? OR display_name LIKE ?", keyword, keyword, keyword)
	}
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}
	if query.RoleID > 0 {
		db = db.Joins("JOIN admin_user_roles ON admin_user_roles.admin_id = admin_users.id").
			Where("admin_user_roles.role_id = ?", query.RoleID)
	}
	return db
}

func (s *AdminUserService) findAdminUser(ctx context.Context, db *gorm.DB, id uint64) (models.AdminUser, error) {
	var user models.AdminUser
	err := db.WithContext(ctx).Where("deleted_at IS NULL").Where("id = ?", id).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.AdminUser{}, apperrors.ErrNotFound.WithMessage("管理员不存在")
	}
	return user, err
}

func (s *AdminUserService) findAdminUserForUpdate(ctx context.Context, db *gorm.DB, id uint64) (models.AdminUser, error) {
	var user models.AdminUser
	err := db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("deleted_at IS NULL").Where("id = ?", id).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.AdminUser{}, apperrors.ErrNotFound.WithMessage("管理员不存在")
	}
	return user, err
}

func (s *AdminUserService) ensureAdminUserUnique(ctx context.Context, excludeID uint64, username string, email *string) error {
	query := s.db.WithContext(ctx).Model(&models.AdminUser{})
	if excludeID > 0 {
		query = query.Where("id <> ?", excludeID)
	}
	if username != "" && email != nil {
		query = query.Where("username = ? OR email = ?", strings.TrimSpace(username), *email)
	} else if username != "" {
		query = query.Where("username = ?", strings.TrimSpace(username))
	} else if email != nil {
		query = query.Where("email = ?", *email)
	} else {
		return nil
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return apperrors.ErrConflict.WithMessage("管理员账号或邮箱已存在")
	}
	return nil
}

func (s *AdminUserService) ensureRolesAssignable(ctx context.Context, roleIDs []uint64) error {
	roleIDs = sets.UniqueUint64s(roleIDs)
	if len(roleIDs) == 0 {
		return nil
	}
	var count int64
	if err := s.db.WithContext(ctx).Model(&models.AdminRole{}).
		Where("id IN ?", roleIDs).
		Where("status = ?", support.AdminStatusActive).
		Count(&count).Error; err != nil {
		return err
	}
	if count != int64(len(roleIDs)) {
		return apperrors.ErrValidation.WithMessage("角色不存在或已禁用")
	}
	return nil
}

func replaceAdminUserRoles(ctx context.Context, db *gorm.DB, adminID uint64, roleIDs []uint64) error {
	if err := db.WithContext(ctx).Exec("DELETE FROM admin_user_roles WHERE admin_id = ?", adminID).Error; err != nil {
		return err
	}
	for _, roleID := range sets.UniqueUint64s(roleIDs) {
		if err := db.WithContext(ctx).Table("admin_user_roles").Create(map[string]any{
			"admin_id": adminID,
			"role_id":  roleID,
		}).Error; err != nil {
			return err
		}
	}
	return nil
}

func (s *AdminUserService) roleSummariesByAdminIDs(ctx context.Context, adminIDs []uint64) (map[uint64][]admindto.AdminRoleSummary, error) {
	result := make(map[uint64][]admindto.AdminRoleSummary)
	if len(adminIDs) == 0 {
		return result, nil
	}

	var rows []adminUserRoleRow
	err := s.db.WithContext(ctx).
		Table("admin_user_roles").
		Select("admin_user_roles.admin_id, admin_roles.id AS role_id, admin_roles.code, admin_roles.name").
		Joins("JOIN admin_roles ON admin_roles.id = admin_user_roles.role_id").
		Where("admin_user_roles.admin_id IN ?", adminIDs).
		Order("admin_roles.id ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		result[row.AdminID] = append(result[row.AdminID], admindto.AdminRoleSummary{
			ID:   row.RoleID,
			Code: row.Code,
			Name: row.Name,
		})
	}
	return result, nil
}

func (s *AdminUserService) activeSessionSummaries(ctx context.Context, adminID uint64) ([]admindto.SessionSummary, error) {
	var sessions []models.AdminSession
	err := s.db.WithContext(ctx).
		Where("admin_id = ? AND status = ?", adminID, support.AdminSessionStatusActive).
		Where("expires_at > ?", time.Now()).
		Order("issued_at DESC").
		Limit(10).
		Find(&sessions).Error
	if err != nil {
		return nil, err
	}
	result := make([]admindto.SessionSummary, 0, len(sessions))
	for _, session := range sessions {
		result = append(result, support.SessionSummary(session))
	}
	return result, nil
}

type adminUserRoleRow struct {
	AdminID uint64 `gorm:"column:admin_id"`
	RoleID  uint64 `gorm:"column:role_id"`
	Code    string `gorm:"column:code"`
	Name    string `gorm:"column:name"`
}

func adminUserIDs(users []models.AdminUser) []uint64 {
	ids := make([]uint64, 0, len(users))
	for _, user := range users {
		ids = append(ids, user.ID)
	}
	return ids
}

func adminUserItem(user models.AdminUser, roles []admindto.AdminRoleSummary) admindto.AdminUserItem {
	roleIDs := make([]uint64, 0, len(roles))
	for _, role := range roles {
		roleIDs = append(roleIDs, role.ID)
	}
	return admindto.AdminUserItem{
		ID:          user.ID,
		Username:    user.Username,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		Status:      user.Status,
		RoleIDs:     roleIDs,
		Roles:       roles,
		LastLoginAt: user.LastLoginAt,
		LastLoginIP: user.LastLoginIP,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}
}

func adminUserAuditSnapshot(user models.AdminUser, roleIDs []uint64) map[string]any {
	return map[string]any{
		"id":           user.ID,
		"username":     user.Username,
		"email":        user.Email,
		"display_name": user.DisplayName,
		"status":       user.Status,
		"role_ids":     sets.UniqueUint64s(roleIDs),
	}
}
