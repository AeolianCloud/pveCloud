package adminuser

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"

	mysqliam "github.com/AeolianCloud/pveCloud/server/internal/repository/mysql/iam"
	mysqltx "github.com/AeolianCloud/pveCloud/server/internal/repository/mysql/tx"
	apperrors "github.com/AeolianCloud/pveCloud/server/internal/shared/errors"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/password"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/rbac"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/sets"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/textutil"
	admindto "github.com/AeolianCloud/pveCloud/server/internal/usecase/admin/dto"
	"github.com/AeolianCloud/pveCloud/server/internal/usecase/admin/support"
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
	iam          *mysqliam.Repository
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
	return &AdminUserService{
		db:           db,
		iam:          mysqliam.NewRepository(db),
		auditService: auditService,
	}
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
	page, perPage := adminsupport.NormalizePage(query.Page, query.PerPage)
	users, total, err := s.iam.AdminUsers(ctx, mysqliam.AdminUserListFilters{
		Keyword: query.Keyword,
		Status:  query.Status,
		RoleID:  query.RoleID,
	}, perPage, (page-1)*perPage)
	if err != nil {
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
	return adminsupport.PageResponse(items, total, page, perPage), nil
}

/**
 * Create 创建管理员账号。
 *
 * @param ctx 请求上下文
 * @param operatorID 当前操作者管理员 ID
 * @param req 创建请求
 * @return admin.AdminUserItem 创建后的管理员账号
 * @return error 创建失败原因
 */
func (s *AdminUserService) Create(ctx context.Context, operatorID uint64, operatorPermissionCodes []string, req admindto.AdminUserCreateRequest) (admindto.AdminUserItem, error) {
	email := textutil.NormalizeOptionalString(req.Email)
	if err := s.ensureAdminUserUnique(ctx, 0, req.Username, email); err != nil {
		return admindto.AdminUserItem{}, err
	}
	if err := s.ensureRolesAssignable(ctx, req.RoleIDs, operatorPermissionCodes); err != nil {
		return admindto.AdminUserItem{}, err
	}
	passwordHash, err := password.Hash(req.Password)
	if err != nil {
		return admindto.AdminUserItem{}, err
	}

	var created mysqliam.AdminUser
	if err := mysqltx.NewManager(s.db).WithinContext(ctx, func(tx *gorm.DB) error {
		created = mysqliam.AdminUser{
			Username:     strings.TrimSpace(req.Username),
			Email:        email,
			PasswordHash: passwordHash,
			DisplayName:  strings.TrimSpace(req.DisplayName),
			Status:       strings.TrimSpace(req.Status),
		}
		if err := s.iam.CreateAdminUser(ctx, tx, &created); err != nil {
			return err
		}
		if err := s.iam.ReplaceAdminUserRoles(ctx, tx, created.ID, req.RoleIDs); err != nil {
			return err
		}
		return s.auditService.Record(ctx, tx, AdminAuditWriteInput{
			AdminID:    &operatorID,
			Action:     adminUserCreateAction,
			ObjectType: adminUserObjectType,
			ObjectID:   textutil.Uint64String(created.ID),
			AfterData:  adminUserAuditSnapshot(created, req.RoleIDs),
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
	codes, err := adminsupport.PermissionCodes(ctx, s.db, id)
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
 * @return admin.AdminUserItem 更新后的管理员账号
 * @return error 更新失败原因
 */
func (s *AdminUserService) Update(ctx context.Context, operatorID uint64, operatorPermissionCodes []string, id uint64, req admindto.AdminUserUpdateRequest) (admindto.AdminUserItem, error) {
	email := textutil.NormalizeOptionalString(req.Email)
	if err := s.ensureAdminUserUnique(ctx, id, "", email); err != nil {
		return admindto.AdminUserItem{}, err
	}
	if req.RoleIDs != nil {
		if id == operatorID {
			return admindto.AdminUserItem{}, apperrors.ErrConflict.WithMessage("不能修改当前登录账号的角色")
		}
		if err := s.ensureRolesAssignable(ctx, req.RoleIDs, operatorPermissionCodes); err != nil {
			return admindto.AdminUserItem{}, err
		}
	}
	if req.Status != nil && id == operatorID && *req.Status != adminsupport.AdminStatusActive {
		return admindto.AdminUserItem{}, apperrors.ErrConflict.WithMessage("不能禁用当前登录账号")
	}

	var updated mysqliam.AdminUser
	var roles []admindto.AdminRoleSummary
	if err := mysqltx.NewManager(s.db).WithinContext(ctx, func(tx *gorm.DB) error {
		current, err := s.findAdminUserForUpdate(ctx, tx, id)
		if err != nil {
			return err
		}
		beforeRoleIDs, err := adminsupport.RoleIDs(ctx, tx, id)
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
		if err := s.iam.UpdateAdminUser(ctx, tx, id, updates); err != nil {
			return err
		}
		if len(updates) > 0 && req.Status != nil && current.Status == adminsupport.AdminStatusActive && strings.TrimSpace(*req.Status) != adminsupport.AdminStatusActive {
			now := time.Now()
			reason := "admin_disabled"
			if err := s.iam.RevokeActiveAdminSessionsByAdminID(ctx, tx, id, now, reason); err != nil {
				return err
			}
		}
		if req.RoleIDs != nil {
			if err := s.iam.ReplaceAdminUserRoles(ctx, tx, id, req.RoleIDs); err != nil {
				return err
			}
		}
		updated, err = s.iam.FindAdminUserByID(ctx, tx, id)
		if err != nil {
			return err
		}
		afterRoleIDs, err := adminsupport.RoleIDs(ctx, tx, id)
		if err != nil {
			return err
		}
		action := adminsupport.AdminUserUpdateAuditAction(current, updated, beforeRoleIDs, afterRoleIDs)
		input := AdminAuditWriteInput{
			AdminID:    &operatorID,
			Action:     action,
			ObjectType: adminUserObjectType,
			ObjectID:   textutil.Uint64String(id),
			BeforeData: adminUserAuditSnapshot(current, beforeRoleIDs),
			AfterData:  adminUserAuditSnapshot(updated, afterRoleIDs),
			Remark:     "更新管理员账号",
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
 * @return error 重置失败原因
 */
func (s *AdminUserService) ResetPassword(ctx context.Context, operatorID uint64, id uint64, req admindto.AdminUserPasswordRequest) error {
	passwordHash, err := password.Hash(req.Password)
	if err != nil {
		return err
	}

	return mysqltx.NewManager(s.db).WithinContext(ctx, func(tx *gorm.DB) error {
		current, err := s.findAdminUserForUpdate(ctx, tx, id)
		if err != nil {
			return err
		}
		if err := s.iam.UpdateAdminUserPasswordHash(ctx, tx, id, passwordHash); err != nil {
			return err
		}
		now := time.Now()
		reason := "password_reset"
		if err := s.iam.RevokeActiveAdminSessionsByAdminID(ctx, tx, id, now, reason); err != nil {
			return err
		}
		return s.auditService.Record(ctx, tx, AdminAuditWriteInput{
			AdminID:    &operatorID,
			Action:     adminUserPasswordResetAction,
			ObjectType: adminUserObjectType,
			ObjectID:   textutil.Uint64String(id),
			BeforeData: map[string]any{"id": current.ID, "username": current.Username},
			AfterData:  map[string]any{"id": current.ID, "username": current.Username, "password_reset": true},
			Remark:     "重置管理员密码",
		})
	})
}

func (s *AdminUserService) findAdminUser(ctx context.Context, db *gorm.DB, id uint64) (mysqliam.AdminUser, error) {
	user, err := s.iam.FindAdminUserByID(ctx, db, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return mysqliam.AdminUser{}, apperrors.ErrNotFound.WithMessage("管理员不存在")
	}
	return user, err
}

func (s *AdminUserService) findAdminUserForUpdate(ctx context.Context, db *gorm.DB, id uint64) (mysqliam.AdminUser, error) {
	user, err := s.iam.FindAdminUserByIDForUpdate(ctx, db, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return mysqliam.AdminUser{}, apperrors.ErrNotFound.WithMessage("管理员不存在")
	}
	return user, err
}

func (s *AdminUserService) ensureAdminUserUnique(ctx context.Context, excludeID uint64, username string, email *string) error {
	count, err := s.iam.CountAdminUsersByIdentity(ctx, excludeID, username, email)
	if err != nil {
		return err
	}
	if count > 0 {
		return apperrors.ErrConflict.WithMessage("管理员账号或邮箱已存在")
	}
	return nil
}

func (s *AdminUserService) ensureRolesAssignable(ctx context.Context, roleIDs []uint64, operatorPermissionCodes []string) error {
	roleIDs = sets.UniqueUint64s(roleIDs)
	if len(roleIDs) == 0 {
		return nil
	}
	roles, err := s.iam.ActiveAdminRolesByIDs(ctx, roleIDs, adminsupport.AdminStatusActive)
	if err != nil {
		return err
	}
	if len(roles) != len(roleIDs) {
		return apperrors.ErrValidation.WithMessage("角色不存在或已禁用")
	}
	targetCodes, err := s.permissionCodesByRoleIDs(ctx, roleIDs)
	if err != nil {
		return err
	}
	for _, code := range targetCodes {
		if !rbac.HasPermissionCode(operatorPermissionCodes, code) {
			return apperrors.ErrForbidden.WithMessage("不能分配包含当前管理员未拥有权限的角色")
		}
	}
	return nil
}

func (s *AdminUserService) permissionCodesByRoleIDs(ctx context.Context, roleIDs []uint64) ([]string, error) {
	codes, err := s.iam.PermissionCodesByRoleIDs(ctx, nil, roleIDs)
	return sets.UniqueStrings(codes), err
}

func (s *AdminUserService) roleSummariesByAdminIDs(ctx context.Context, adminIDs []uint64) (map[uint64][]admindto.AdminRoleSummary, error) {
	result := make(map[uint64][]admindto.AdminRoleSummary)
	if len(adminIDs) == 0 {
		return result, nil
	}

	rows, err := s.iam.RoleSummariesByAdminIDs(ctx, adminIDs)
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
	sessions, err := s.iam.ActiveAdminSessions(ctx, adminID, time.Now(), 10)
	if err != nil {
		return nil, err
	}
	result := make([]admindto.SessionSummary, 0, len(sessions))
	for _, session := range sessions {
		result = append(result, adminsupport.SessionSummary(session))
	}
	return result, nil
}

func adminUserIDs(users []mysqliam.AdminUser) []uint64 {
	ids := make([]uint64, 0, len(users))
	for _, user := range users {
		ids = append(ids, user.ID)
	}
	return ids
}

func adminUserItem(user mysqliam.AdminUser, roles []admindto.AdminRoleSummary) admindto.AdminUserItem {
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

func adminUserAuditSnapshot(user mysqliam.AdminUser, roleIDs []uint64) map[string]any {
	return map[string]any{
		"id":           user.ID,
		"username":     user.Username,
		"email":        user.Email,
		"display_name": user.DisplayName,
		"status":       user.Status,
		"role_ids":     sets.UniqueUint64s(roleIDs),
	}
}
