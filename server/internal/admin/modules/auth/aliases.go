package auth

import auditpkg "github.com/AeolianCloud/pveCloud/server/internal/admin/modules/audit"

type AdminAuditService = auditpkg.AdminAuditService
type AdminAuditWriteInput = auditpkg.AdminAuditWriteInput
type AdminRiskWriteInput = auditpkg.AdminRiskWriteInput

var NewAdminAuditService = auditpkg.NewAdminAuditService
