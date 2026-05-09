package adminsession

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"

	domainiam "github.com/AeolianCloud/pveCloud/server/internal/domain/iam"
	mysqliam "github.com/AeolianCloud/pveCloud/server/internal/repository/mysql/iam"
	mysqltx "github.com/AeolianCloud/pveCloud/server/internal/repository/mysql/tx"
	apperrors "github.com/AeolianCloud/pveCloud/server/internal/shared/errors"
	admindto "github.com/AeolianCloud/pveCloud/server/internal/usecase/admin/dto"
	"github.com/AeolianCloud/pveCloud/server/internal/usecase/admin/support"
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
	iam          *mysqliam.Repository
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
	return &AdminSessionService{
		db:           db,
		iam:          mysqliam.NewRepository(db),
		auditService: auditService,
	}
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

	page, perPage := adminsupport.NormalizePage(query.Page, query.PerPage)
	rows, total, err := s.iam.AdminSessionRows(ctx, mysqliam.AdminSessionListFilters{
		Keyword: query.Keyword,
		Status:  query.Status,
	}, perPage, (page-1)*perPage)
	if err != nil {
		return admindto.PageResponse[admindto.AdminSessionItem]{}, err
	}

	items := make([]admindto.AdminSessionItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, adminSessionItem(row, currentSessionID))
	}
	return adminsupport.PageResponse(items, total, page, perPage), nil
}

/**
 * Revoke 吊销指定管理员会话。
 *
 * @param ctx 请求上下文
 * @param operatorID 当前操作者管理员 ID
 * @param currentSessionID 当前登录会话 ID
 * @param targetSessionID 目标会话 ID
 * @return error 吊销失败原因
 */
func (s *AdminSessionService) Revoke(ctx context.Context, operatorID uint64, currentSessionID string, targetSessionID string) error {
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

	return mysqltx.NewManager(s.db).WithinContext(ctx, func(tx *gorm.DB) error {
		session, err := s.iam.FindAdminSessionBySessionIDForUpdate(ctx, tx, targetSessionID)
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
		if !session.ExpiresAt.After(now) && session.Status == domainiam.SessionStatusActive {
			if err := s.iam.UpdateAdminSessionState(ctx, tx, session.ID, domainiam.SessionStatusExpired, now, domainiam.RevokeReasonExpired); err != nil {
				return err
			}
			return apperrors.ErrConflict.WithMessage("会话已过期")
		}
		if session.Status != domainiam.SessionStatusActive {
			return apperrors.ErrConflict.WithMessage("会话已不是活跃状态")
		}

		before := adminSessionAuditSnapshot(session)
		afterSession := session
		afterSession.Status = domainiam.SessionStatusRevoked
		afterSession.RevokedAt = &now
		afterSession.RevokeReason = stringPtr(adminSessionRevokeReasonAdmin)

		if err := s.iam.UpdateAdminSessionState(ctx, tx, session.ID, domainiam.SessionStatusRevoked, now, domainiam.RevokeReasonAdmin); err != nil {
			return err
		}

		return s.auditService.Record(ctx, tx, AdminAuditWriteInput{
			AdminID:    &operatorID,
			Action:     adminSessionRevokeAction,
			ObjectType: adminSessionObjectType,
			ObjectID:   session.SessionID,
			BeforeData: before,
			AfterData:  adminSessionAuditSnapshot(afterSession),
			Remark:     "吊销管理员会话",
		})
	})
}

func (s *AdminSessionService) expireStaleSessions(ctx context.Context) error {
	now := time.Now()
	return s.iam.ExpireStaleAdminSessions(
		ctx,
		nil,
		now,
		domainiam.SessionStatusActive,
		domainiam.SessionStatusExpired,
		domainiam.RevokeReasonExpired,
	)
}

func adminSessionItem(row mysqliam.AdminSessionListRow, currentSessionID string) admindto.AdminSessionItem {
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

func adminSessionAuditSnapshot(session mysqliam.AdminSession) map[string]any {
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
