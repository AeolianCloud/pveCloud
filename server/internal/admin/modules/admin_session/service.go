package adminsession

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
)

const (
	adminSessionObjectType         = "admin_session"
	adminSessionRevokeAction       = "admin.session.revoke"
	adminSessionRevokeReasonAdmin  = "admin_revoke"
	adminSessionRevokeReasonExpire = "expired"
)

/**
 * AdminSessionService 处理管理员会话管理。
 */
type AdminSessionService struct {
	db           *gorm.DB
	auditService *AdminAuditService
}

/**
 * NewAdminSessionService 创建管理员会话服务。
 *
 * @param db 数据库连接
 * @param auditService 后台审计服务
 * @return *AdminSessionService 管理员会话服务
 */
func NewAdminSessionService(db *gorm.DB, auditService *AdminAuditService) *AdminSessionService {
	if auditService == nil {
		auditService = NewAdminAuditService(db)
	}
	return &AdminSessionService{db: db, auditService: auditService}
}

/**
 * List 分页查询管理员会话。
 *
 * @param ctx 请求上下文
 * @param query 查询参数
 * @param currentSessionID 当前登录会话 ID
 * @return admin.PageResponse[admin.AdminSessionItem] 分页结果
 * @return error 查询失败原因
 */
func (s *AdminSessionService) List(ctx context.Context, query admindto.AdminSessionListQuery, currentSessionID string) (admindto.PageResponse[admindto.AdminSessionItem], error) {
	if err := s.expireStaleSessions(ctx); err != nil {
		return admindto.PageResponse[admindto.AdminSessionItem]{}, err
	}

	page, perPage := support.NormalizePage(query.Page, query.PerPage)
	db := s.db.WithContext(ctx).
		Table("admin_sessions AS sessions").
		Joins("JOIN admin_users ON admin_users.id = sessions.admin_id").
		Where("admin_users.deleted_at IS NULL")

	db = applyAdminSessionFilters(db, query)

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return admindto.PageResponse[admindto.AdminSessionItem]{}, err
	}

	var rows []adminSessionListRow
	err := db.
		Select(`
			sessions.session_id,
			sessions.admin_id,
			admin_users.username AS admin_username,
			admin_users.display_name AS admin_display_name,
			admin_users.email AS admin_email,
			sessions.status,
			sessions.issued_at,
			sessions.expires_at,
			sessions.last_seen_at,
			sessions.last_seen_ip,
			sessions.user_agent,
			sessions.revoked_at,
			sessions.revoke_reason
		`).
		Order("CASE sessions.status WHEN 'active' THEN 0 WHEN 'revoked' THEN 1 ELSE 2 END").
		Order("sessions.issued_at DESC").
		Limit(perPage).
		Offset((page - 1) * perPage).
		Scan(&rows).Error
	if err != nil {
		return admindto.PageResponse[admindto.AdminSessionItem]{}, err
	}

	items := make([]admindto.AdminSessionItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, adminSessionItem(row, currentSessionID))
	}
	return support.PageResponse(items, total, page, perPage), nil
}

/**
 * Revoke 吊销指定管理员会话。
 *
 * @param ctx 请求上下文
 * @param operatorID 当前操作者管理员 ID
 * @param currentSessionID 当前登录会话 ID
 * @param targetSessionID 目标会话 ID
 * @param clientIP 客户端 IP
 * @param userAgent 浏览器 User-Agent
 * @return error 吊销失败原因
 */
func (s *AdminSessionService) Revoke(ctx context.Context, operatorID uint64, currentSessionID string, targetSessionID string, clientIP string, userAgent string) error {
	targetSessionID = strings.TrimSpace(targetSessionID)
	if targetSessionID == "" {
		return apperrors.ErrValidation.WithMessage("会话 ID 不能为空")
	}
	if strings.TrimSpace(currentSessionID) == targetSessionID {
		return apperrors.ErrConflict.WithMessage("不能吊销当前会话")
	}
	if err := s.expireStaleSessions(ctx); err != nil {
		return err
	}

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var session models.AdminSession
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("session_id = ?", targetSessionID).
			First(&session).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrNotFound.WithMessage("管理员会话不存在")
		}
		if err != nil {
			return err
		}

		if session.SessionID == currentSessionID {
			return apperrors.ErrConflict.WithMessage("不能吊销当前会话")
		}

		now := time.Now()
		if !session.ExpiresAt.After(now) && session.Status == support.AdminSessionStatusActive {
			if err := tx.Model(&models.AdminSession{}).
				Where("id = ?", session.ID).
				Updates(map[string]any{
					"status":        support.AdminSessionStatusExpired,
					"revoked_at":    now,
					"revoke_reason": adminSessionRevokeReasonExpire,
				}).Error; err != nil {
				return err
			}
			return apperrors.ErrConflict.WithMessage("会话已过期")
		}
		if session.Status != support.AdminSessionStatusActive {
			return apperrors.ErrConflict.WithMessage("会话已不是活跃状态")
		}

		before := adminSessionAuditSnapshot(session)
		afterSession := session
		afterSession.Status = support.AdminSessionStatusRevoked
		afterSession.RevokedAt = &now
		afterSession.RevokeReason = stringPtr(adminSessionRevokeReasonAdmin)

		if err := tx.Model(&models.AdminSession{}).
			Where("id = ?", session.ID).
			Updates(map[string]any{
				"status":        support.AdminSessionStatusRevoked,
				"revoked_at":    now,
				"revoke_reason": adminSessionRevokeReasonAdmin,
			}).Error; err != nil {
			return err
		}

		return s.auditService.RecordRisk(ctx, tx, AdminRiskWriteInput{
			AdminAuditWriteInput: AdminAuditWriteInput{
				AdminID:    &operatorID,
				Action:     adminSessionRevokeAction,
				ObjectType: adminSessionObjectType,
				ObjectID:   session.SessionID,
				BeforeData: before,
				AfterData:  adminSessionAuditSnapshot(afterSession),
				IP:         clientIP,
				UserAgent:  userAgent,
				Remark:     "吊销管理员会话",
			},
			RiskLevel:  "high",
			RiskReason: "吊销管理员会话",
		})
	})
}

func (s *AdminSessionService) expireStaleSessions(ctx context.Context) error {
	now := time.Now()
	return s.db.WithContext(ctx).
		Model(&models.AdminSession{}).
		Where("status = ? AND expires_at <= ?", support.AdminSessionStatusActive, now).
		Updates(map[string]any{
			"status":        support.AdminSessionStatusExpired,
			"revoked_at":    now,
			"revoke_reason": adminSessionRevokeReasonExpire,
		}).Error
}

func applyAdminSessionFilters(db *gorm.DB, query admindto.AdminSessionListQuery) *gorm.DB {
	if query.Keyword != "" {
		keyword := "%" + strings.TrimSpace(query.Keyword) + "%"
		db = db.Where(
			"sessions.session_id LIKE ? OR admin_users.username LIKE ? OR admin_users.display_name LIKE ? OR sessions.last_seen_ip LIKE ?",
			keyword,
			keyword,
			keyword,
			keyword,
		)
	}
	if query.Status != "" {
		db = db.Where("sessions.status = ?", strings.TrimSpace(query.Status))
	}
	return db
}

type adminSessionListRow struct {
	SessionID        string     `gorm:"column:session_id"`
	AdminID          uint64     `gorm:"column:admin_id"`
	AdminUsername    string     `gorm:"column:admin_username"`
	AdminDisplayName string     `gorm:"column:admin_display_name"`
	AdminEmail       *string    `gorm:"column:admin_email"`
	Status           string     `gorm:"column:status"`
	IssuedAt         time.Time  `gorm:"column:issued_at"`
	ExpiresAt        time.Time  `gorm:"column:expires_at"`
	LastSeenAt       *time.Time `gorm:"column:last_seen_at"`
	LastSeenIP       *string    `gorm:"column:last_seen_ip"`
	UserAgent        *string    `gorm:"column:user_agent"`
	RevokedAt        *time.Time `gorm:"column:revoked_at"`
	RevokeReason     *string    `gorm:"column:revoke_reason"`
}

func adminSessionItem(row adminSessionListRow, currentSessionID string) admindto.AdminSessionItem {
	return admindto.AdminSessionItem{
		SessionID:        row.SessionID,
		AdminID:          row.AdminID,
		AdminUsername:    row.AdminUsername,
		AdminDisplayName: row.AdminDisplayName,
		AdminEmail:       row.AdminEmail,
		Status:           row.Status,
		IssuedAt:         row.IssuedAt,
		ExpiresAt:        row.ExpiresAt,
		LastSeenAt:       row.LastSeenAt,
		LastSeenIP:       row.LastSeenIP,
		UserAgent:        row.UserAgent,
		RevokedAt:        row.RevokedAt,
		RevokeReason:     row.RevokeReason,
		IsCurrent:        strings.TrimSpace(currentSessionID) != "" && row.SessionID == strings.TrimSpace(currentSessionID),
	}
}

func adminSessionAuditSnapshot(session models.AdminSession) map[string]any {
	return map[string]any{
		"session_id":    session.SessionID,
		"admin_id":      session.AdminID,
		"status":        session.Status,
		"issued_at":     session.IssuedAt,
		"expires_at":    session.ExpiresAt,
		"last_seen_at":  session.LastSeenAt,
		"last_seen_ip":  session.LastSeenIP,
		"user_agent":    session.UserAgent,
		"revoked_at":    session.RevokedAt,
		"revoke_reason": session.RevokeReason,
	}
}

func stringPtr(value string) *string {
	return &value
}
