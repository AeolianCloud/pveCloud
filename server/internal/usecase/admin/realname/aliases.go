package realname

import "github.com/AeolianCloud/pveCloud/server/internal/usecase/admin/audit"

type AdminAuditService = audit.AdminAuditService
type AdminAuditWriteInput = audit.AdminAuditWriteInput

var NewAdminAuditService = audit.NewAdminAuditService
