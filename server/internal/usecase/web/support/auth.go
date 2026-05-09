package websupport

import (
	mysqluser "github.com/AeolianCloud/pveCloud/server/internal/repository/mysql/user"
	webdto "github.com/AeolianCloud/pveCloud/server/internal/usecase/web/dto"
)

/**
 * UserSummary 将用户模型转换为用户端账号摘要。
 */
func UserSummary(user mysqluser.User) webdto.UserSummary {
	return webdto.UserSummary{
		ID:          user.ID,
		Username:    user.Username,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		Status:      user.Status,
	}
}

/**
 * SessionSummary 将用户端会话模型转换为会话摘要。
 */
func SessionSummary(session mysqluser.UserSession) webdto.SessionSummary {
	return webdto.SessionSummary{
		SessionID: session.SessionID,
		IssuedAt:  session.IssuedAt,
		ExpiresAt: session.ExpiresAt,
	}
}
