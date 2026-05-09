package systemconfig

import auditpkg "github.com/AeolianCloud/pveCloud/server/internal/usecase/admin/audit"

type AdminAuditService = auditpkg.AdminAuditService
type AdminAuditWriteInput = auditpkg.AdminAuditWriteInput

var NewAdminAuditService = auditpkg.NewAdminAuditService
