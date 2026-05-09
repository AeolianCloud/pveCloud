package iam

import (
	"context"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AdminUserListFilters struct {
	Keyword string
	Status  string
	RoleID  uint64
}

type AdminRoleListFilters struct {
	Keyword string
	Status  string
}

type AdminUserRoleRow struct {
	AdminID uint64 `gorm:"column:admin_id"`
	RoleID  uint64 `gorm:"column:role_id"`
	Code    string `gorm:"column:code"`
	Name    string `gorm:"column:name"`
}

type AdminRolePermissionRow struct {
	RoleID uint64 `gorm:"column:role_id"`
	Code   string `gorm:"column:code"`
}

func (r *Repository) AdminUsers(ctx context.Context, filters AdminUserListFilters, limit int, offset int) ([]AdminUser, int64, error) {
	query := r.applyAdminUserListFilters(r.db.WithContext(ctx).Model(&AdminUser{}).Where("deleted_at IS NULL"), filters)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var users []AdminUser
	if err := query.Order("id DESC").Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

func (r *Repository) AdminRoles(ctx context.Context, filters AdminRoleListFilters, limit int, offset int) ([]AdminRole, int64, error) {
	query := r.applyAdminRoleListFilters(r.db.WithContext(ctx).Model(&AdminRole{}), filters)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var roles []AdminRole
	if err := query.Order("id ASC").Limit(limit).Offset(offset).Find(&roles).Error; err != nil {
		return nil, 0, err
	}
	return roles, total, nil
}

func (r *Repository) CreateAdminUser(ctx context.Context, db *gorm.DB, user *AdminUser) error {
	return r.queryDB(db).WithContext(ctx).Create(user).Error
}

func (r *Repository) CreateAdminRole(ctx context.Context, db *gorm.DB, role *AdminRole) error {
	return r.queryDB(db).WithContext(ctx).Create(role).Error
}

func (r *Repository) FindAdminUserByID(ctx context.Context, db *gorm.DB, id uint64) (AdminUser, error) {
	var user AdminUser
	err := r.queryDB(db).WithContext(ctx).
		Where("deleted_at IS NULL").
		Where("id = ?", id).
		First(&user).Error
	return user, err
}

func (r *Repository) FindAdminUserByIDForUpdate(ctx context.Context, db *gorm.DB, id uint64) (AdminUser, error) {
	var user AdminUser
	err := r.queryDB(db).WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("deleted_at IS NULL").
		Where("id = ?", id).
		First(&user).Error
	return user, err
}

func (r *Repository) FindAdminByAccount(ctx context.Context, account string) (AdminUser, error) {
	account = strings.TrimSpace(account)
	var admin AdminUser
	err := r.db.WithContext(ctx).
		Where("deleted_at IS NULL").
		Where("username = ? OR email = ?", account, account).
		First(&admin).Error
	return admin, err
}

func (r *Repository) FindAdminRoleByID(ctx context.Context, db *gorm.DB, id uint64) (AdminRole, error) {
	var role AdminRole
	err := r.queryDB(db).WithContext(ctx).Where("id = ?", id).First(&role).Error
	return role, err
}

func (r *Repository) FindAdminRoleByIDForUpdate(ctx context.Context, db *gorm.DB, id uint64) (AdminRole, error) {
	var role AdminRole
	err := r.queryDB(db).WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ?", id).
		First(&role).Error
	return role, err
}

func (r *Repository) UpdateAdminUser(ctx context.Context, db *gorm.DB, id uint64, updates map[string]any) error {
	if len(updates) == 0 {
		return nil
	}
	return r.queryDB(db).WithContext(ctx).
		Model(&AdminUser{}).
		Where("id = ?", id).
		Updates(updates).Error
}

func (r *Repository) UpdateAdminRole(ctx context.Context, db *gorm.DB, id uint64, updates map[string]any) error {
	if len(updates) == 0 {
		return nil
	}
	return r.queryDB(db).WithContext(ctx).
		Model(&AdminRole{}).
		Where("id = ?", id).
		Updates(updates).Error
}

func (r *Repository) UpdateAdminUserPasswordHash(ctx context.Context, db *gorm.DB, id uint64, passwordHash string) error {
	return r.queryDB(db).WithContext(ctx).
		Model(&AdminUser{}).
		Where("id = ?", id).
		Update("password_hash", passwordHash).Error
}

func (r *Repository) UpdateAdminLastLogin(ctx context.Context, db *gorm.DB, id uint64, at time.Time, ip string) error {
	return r.queryDB(db).WithContext(ctx).
		Model(&AdminUser{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"last_login_at": at,
			"last_login_ip": ip,
		}).Error
}

func (r *Repository) CountAdminUsersByIdentity(ctx context.Context, excludeID uint64, username string, email *string) (int64, error) {
	query := r.db.WithContext(ctx).Model(&AdminUser{})
	if excludeID > 0 {
		query = query.Where("id <> ?", excludeID)
	}
	username = strings.TrimSpace(username)
	if username != "" && email != nil {
		query = query.Where("username = ? OR email = ?", username, *email)
	} else if username != "" {
		query = query.Where("username = ?", username)
	} else if email != nil {
		query = query.Where("email = ?", *email)
	} else {
		return 0, nil
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *Repository) CountAdminRolesByCode(ctx context.Context, excludeID uint64, code string) (int64, error) {
	query := r.db.WithContext(ctx).Model(&AdminRole{}).Where("code = ?", strings.TrimSpace(code))
	if excludeID > 0 {
		query = query.Where("id <> ?", excludeID)
	}
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *Repository) ActiveAdminRolesByIDs(ctx context.Context, ids []uint64, activeStatus string) ([]AdminRole, error) {
	var roles []AdminRole
	if len(ids) == 0 {
		return roles, nil
	}
	err := r.db.WithContext(ctx).
		Model(&AdminRole{}).
		Where("id IN ?", ids).
		Where("status = ?", activeStatus).
		Find(&roles).Error
	return roles, err
}

func (r *Repository) AdminPermissions(ctx context.Context, groupName string) ([]AdminPermission, error) {
	db := r.db.WithContext(ctx).Model(&AdminPermission{})
	if strings.TrimSpace(groupName) != "" {
		db = db.Where("group_name = ?", strings.TrimSpace(groupName))
	}
	var permissions []AdminPermission
	if err := db.Order("sort_order ASC, id ASC").Find(&permissions).Error; err != nil {
		return nil, err
	}
	return permissions, nil
}

func (r *Repository) AllAdminPermissions(ctx context.Context) ([]AdminPermission, error) {
	var permissions []AdminPermission
	if err := r.db.WithContext(ctx).Find(&permissions).Error; err != nil {
		return nil, err
	}
	return permissions, nil
}

func (r *Repository) AdminPermissionsByCodes(ctx context.Context, codes []string) ([]AdminPermission, error) {
	var permissions []AdminPermission
	if len(codes) == 0 {
		return permissions, nil
	}
	err := r.db.WithContext(ctx).Where("code IN ?", codes).Find(&permissions).Error
	return permissions, err
}

func (r *Repository) RolePermissionCodeMap(ctx context.Context, db *gorm.DB, roleIDs []uint64) (map[uint64][]string, error) {
	result := make(map[uint64][]string)
	if len(roleIDs) == 0 {
		return result, nil
	}
	var rows []AdminRolePermissionRow
	err := r.queryDB(db).WithContext(ctx).
		Table("admin_role_permissions").
		Select("admin_role_permissions.role_id, admin_permissions.code").
		Joins("JOIN admin_permissions ON admin_permissions.id = admin_role_permissions.permission_id").
		Where("admin_role_permissions.role_id IN ?", roleIDs).
		Order("admin_permissions.code ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		result[row.RoleID] = append(result[row.RoleID], row.Code)
	}
	return result, nil
}

func (r *Repository) PermissionCodesByRoleIDs(ctx context.Context, db *gorm.DB, roleIDs []uint64) ([]string, error) {
	var codes []string
	if len(roleIDs) == 0 {
		return codes, nil
	}
	err := r.queryDB(db).WithContext(ctx).
		Table("admin_role_permissions").
		Distinct("admin_permissions.code").
		Joins("JOIN admin_permissions ON admin_permissions.id = admin_role_permissions.permission_id").
		Where("admin_role_permissions.role_id IN ?", roleIDs).
		Order("admin_permissions.code ASC").
		Scan(&codes).Error
	return codes, err
}

func (r *Repository) PermissionCodesByRoleID(ctx context.Context, db *gorm.DB, roleID uint64) ([]string, error) {
	var codes []string
	err := r.queryDB(db).WithContext(ctx).
		Table("admin_role_permissions").
		Select("admin_permissions.code").
		Joins("JOIN admin_permissions ON admin_permissions.id = admin_role_permissions.permission_id").
		Where("admin_role_permissions.role_id = ?", roleID).
		Order("admin_permissions.code ASC").
		Scan(&codes).Error
	return codes, err
}

func (r *Repository) ReplaceAdminRolePermissions(ctx context.Context, db *gorm.DB, roleID uint64, permissionIDs []uint64) error {
	target := r.queryDB(db).WithContext(ctx)
	if err := target.Exec("DELETE FROM admin_role_permissions WHERE role_id = ?", roleID).Error; err != nil {
		return err
	}
	for _, permissionID := range uniqueUint64s(permissionIDs) {
		if err := target.Table("admin_role_permissions").Create(map[string]any{
			"role_id":       roleID,
			"permission_id": permissionID,
		}).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) ReplaceAdminUserRoles(ctx context.Context, db *gorm.DB, adminID uint64, roleIDs []uint64) error {
	target := r.queryDB(db).WithContext(ctx)
	if err := target.Exec("DELETE FROM admin_user_roles WHERE admin_id = ?", adminID).Error; err != nil {
		return err
	}
	for _, roleID := range uniqueUint64s(roleIDs) {
		if err := target.Table("admin_user_roles").Create(map[string]any{
			"admin_id": adminID,
			"role_id":  roleID,
		}).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) RoleSummariesByAdminIDs(ctx context.Context, adminIDs []uint64) ([]AdminUserRoleRow, error) {
	var rows []AdminUserRoleRow
	if len(adminIDs) == 0 {
		return rows, nil
	}
	err := r.db.WithContext(ctx).
		Table("admin_user_roles").
		Select("admin_user_roles.admin_id, admin_roles.id AS role_id, admin_roles.code, admin_roles.name").
		Joins("JOIN admin_roles ON admin_roles.id = admin_user_roles.role_id").
		Where("admin_user_roles.admin_id IN ?", adminIDs).
		Order("admin_roles.id ASC").
		Scan(&rows).Error
	return rows, err
}

func (r *Repository) AdminUsersByIDs(ctx context.Context, ids []uint64) ([]AdminUser, error) {
	var users []AdminUser
	if len(ids) == 0 {
		return users, nil
	}
	err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&users).Error
	return users, err
}

func (r *Repository) VisibleMenuPermissions(ctx context.Context, db *gorm.DB) ([]AdminPermission, error) {
	var permissions []AdminPermission
	err := r.queryDB(db).WithContext(ctx).
		Where("type = ?", "menu").
		Where("visible_in_menu = ?", true).
		Order("sort_order ASC, id ASC").
		Find(&permissions).Error
	return permissions, err
}

func (r *Repository) applyAdminUserListFilters(db *gorm.DB, filters AdminUserListFilters) *gorm.DB {
	if strings.TrimSpace(filters.Keyword) != "" {
		keyword := "%" + strings.TrimSpace(filters.Keyword) + "%"
		db = db.Where("username LIKE ? OR email LIKE ? OR display_name LIKE ?", keyword, keyword, keyword)
	}
	if strings.TrimSpace(filters.Status) != "" {
		db = db.Where("status = ?", strings.TrimSpace(filters.Status))
	}
	if filters.RoleID > 0 {
		db = db.Joins("JOIN admin_user_roles ON admin_user_roles.admin_id = admin_users.id").
			Where("admin_user_roles.role_id = ?", filters.RoleID)
	}
	return db
}

func (r *Repository) applyAdminRoleListFilters(db *gorm.DB, filters AdminRoleListFilters) *gorm.DB {
	if strings.TrimSpace(filters.Keyword) != "" {
		keyword := "%" + strings.TrimSpace(filters.Keyword) + "%"
		db = db.Where("code LIKE ? OR name LIKE ? OR description LIKE ?", keyword, keyword, keyword)
	}
	if strings.TrimSpace(filters.Status) != "" {
		db = db.Where("status = ?", strings.TrimSpace(filters.Status))
	}
	return db
}

func uniqueUint64s(values []uint64) []uint64 {
	seen := map[uint64]bool{}
	result := make([]uint64, 0, len(values))
	for _, value := range values {
		if value == 0 || seen[value] {
			continue
		}
		seen[value] = true
		result = append(result, value)
	}
	return result
}
