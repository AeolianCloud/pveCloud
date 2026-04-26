package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	admindto "github.com/AeolianCloud/pveCloud/server/internal/dto/admin"
	"github.com/AeolianCloud/pveCloud/server/internal/models"
	apperrors "github.com/AeolianCloud/pveCloud/server/internal/pkg/errors"
)

const (
	adminSessionObjectType       = "admin_session"
	adminSessionRevokeAction     = "admin.session.revoke"
	adminRevokeReasonAdminRevoke = "admin_revoke"
)

/**
 * AdminSessionService 处理管理端登录会话管理。
 */
type AdminSessionService struct {
	db           *gorm.DB
	auditService *AdminAuditService
}

/**
 * NewAdminSessionService 创建管理端登录会话服务。
 *
 * @param db 数据库连接
 * @param auditService 后台审计服务
 * @return *AdminSessionService 管理端登录会话服务
 */
func NewAdminSessionService(db *gorm.DB, auditService *AdminAuditService) *AdminSessionService {
	if auditService == nil {
		auditService = NewAdminAuditService(db)
	}
	return &AdminSessionService{db: db, auditService: auditService}
}

/**
 * Sessions 分页查询管理端登录会话。
 *
 * @param ctx 请求上下文
 * @param query 查询参数
 * @return admin.PageResponse[admin.AdminSessionItem] 分页结果
 * @return error 查询失败原因
 */
func (s *AdminSessionService) Sessions(ctx context.Context, query admindto.AdminSessionListQuery) (admindto.PageResponse[admindto.AdminSessionItem], error) {
	page, perPage := normalizePage(query.Page, query.PerPage)
	db := s.db.WithContext(ctx).Table("admin_sessions")
	if query.AdminID > 0 {
		db = db.Where("admin_sessions.admin_id = ?", query.AdminID)
	}
	if query.Status != "" {
		db = db.Where("admin_sessions.status = ?", query.Status)
	}
	if query.Keyword != "" {
		keyword := "%" + strings.TrimSpace(query.Keyword) + "%"
		db = db.Joins("LEFT JOIN admin_users ON admin_users.id = admin_sessions.admin_id").
			Where("admin_sessions.session_id LIKE ? OR admin_sessions.last_seen_ip LIKE ? OR admin_users.username LIKE ? OR admin_users.display_name LIKE ? OR admin_users.email LIKE ?", keyword, keyword, keyword, keyword, keyword)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return admindto.PageResponse[admindto.AdminSessionItem]{}, err
	}

	var rows []adminSessionRow
	selectDB := db
	if query.Keyword == "" {
		selectDB = selectDB.Joins("LEFT JOIN admin_users ON admin_users.id = admin_sessions.admin_id")
	}
	if err := selectDB.
		Select("admin_sessions.*, admin_users.username AS admin_username, admin_users.display_name AS admin_display_name, admin_users.email AS admin_email").
		Order("admin_sessions.id DESC").
		Limit(perPage).
		Offset((page - 1) * perPage).
		Scan(&rows).Error; err != nil {
		return admindto.PageResponse[admindto.AdminSessionItem]{}, err
	}

	items := make([]admindto.AdminSessionItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, row.sessionItem())
	}
	return pageResponse(items, total, page, perPage), nil
}

/**
 * Revoke 吊销指定活跃会话。
 *
 * @param ctx 请求上下文
 * @param operatorID 当前操作者管理员 ID
 * @param currentSessionID 当前会话 ID
 * @param id 会话记录 ID
 * @param clientIP 客户端 IP
 * @param userAgent 浏览器 User-Agent
 * @return error 吊销失败原因
 */
func (s *AdminSessionService) Revoke(ctx context.Context, operatorID uint64, currentSessionID string, id uint64, clientIP string, userAgent string) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var session models.AdminSession
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", id).First(&session).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrNotFound.WithMessage("会话不存在")
		}
		if err != nil {
			return err
		}
		if session.SessionID == currentSessionID {
			return apperrors.ErrConflict.WithMessage("不能通过列表吊销当前会话")
		}
		if session.Status != adminSessionStatusActive || !session.ExpiresAt.After(time.Now()) {
			return apperrors.ErrConflict.WithMessage("会话已失效")
		}

		now := time.Now()
		reason := adminRevokeReasonAdminRevoke
		if err := tx.Model(&models.AdminSession{}).Where("id = ?", id).Updates(map[string]interface{}{
			"status":        adminSessionStatusRevoked,
			"revoked_at":    now,
			"revoke_reason": reason,
		}).Error; err != nil {
			return err
		}

		return s.auditService.RecordRisk(ctx, tx, AdminRiskWriteInput{
			AdminAuditWriteInput: AdminAuditWriteInput{
				AdminID:    &operatorID,
				Action:     adminSessionRevokeAction,
				ObjectType: adminSessionObjectType,
				ObjectID:   uintIDString(id),
				BeforeData: adminSessionAuditSnapshot(session),
				AfterData:  map[string]any{"id": session.ID, "status": adminSessionStatusRevoked, "revoke_reason": reason},
				IP:         clientIP,
				UserAgent:  userAgent,
				Remark:     "吊销管理端会话",
			},
			RiskLevel:  "high",
			RiskReason: "吊销他人管理端会话",
		})
	})
}

type adminSessionRow struct {
	models.AdminSession
	AdminUsername    *string `gorm:"column:admin_username"`
	AdminDisplayName *string `gorm:"column:admin_display_name"`
	AdminEmail       *string `gorm:"column:admin_email"`
}

func (row adminSessionRow) sessionItem() admindto.AdminSessionItem {
	return admindto.AdminSessionItem{
		ID:           row.ID,
		SessionID:    row.SessionID,
		Admin:        row.adminSummary(),
		Status:       row.Status,
		IssuedAt:     row.IssuedAt,
		ExpiresAt:    row.ExpiresAt,
		LastSeenAt:   row.LastSeenAt,
		LastSeenIP:   row.LastSeenIP,
		UserAgent:    row.UserAgent,
		RevokedAt:    row.RevokedAt,
		RevokeReason: row.RevokeReason,
	}
}

func (row adminSessionRow) adminSummary() *admindto.AuditAdminSummary {
	if row.AdminUsername == nil {
		return nil
	}
	return &admindto.AuditAdminSummary{
		ID:          row.AdminID,
		Username:    *row.AdminUsername,
		DisplayName: valueOrEmpty(row.AdminDisplayName),
		Email:       row.AdminEmail,
	}
}

func adminSessionAuditSnapshot(session models.AdminSession) map[string]any {
	return map[string]any{
		"id":            session.ID,
		"session_id":    session.SessionID,
		"admin_id":      session.AdminID,
		"status":        session.Status,
		"issued_at":     session.IssuedAt,
		"expires_at":    session.ExpiresAt,
		"last_seen_at":  session.LastSeenAt,
		"last_seen_ip":  session.LastSeenIP,
		"revoked_at":    session.RevokedAt,
		"revoke_reason": session.RevokeReason,
	}
}
