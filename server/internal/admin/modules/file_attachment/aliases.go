package fileattachment

import auditpkg "github.com/AeolianCloud/pveCloud/server/internal/admin/modules/audit"

type AdminAuditService = auditpkg.AdminAuditService
type AdminAuditWriteInput = auditpkg.AdminAuditWriteInput

var NewAdminAuditService = auditpkg.NewAdminAuditService
