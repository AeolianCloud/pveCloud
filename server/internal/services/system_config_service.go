package services

import (
	"context"
	"errors"
	"sort"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	admindto "github.com/AeolianCloud/pveCloud/server/internal/dto/admin"
	"github.com/AeolianCloud/pveCloud/server/internal/models"
	apperrors "github.com/AeolianCloud/pveCloud/server/internal/pkg/errors"
)

const (
	systemConfigObjectType         = "system_config"
	systemConfigUpdateAction       = "system.config.update"
	systemConfigSecretUpdateAction = "system.config.secret_update"
)

/**
 * SystemConfigService 处理基础后台系统配置管理。
 */
type SystemConfigService struct {
	db           *gorm.DB
	auditService *AdminAuditService
}

/**
 * NewSystemConfigService 创建系统配置服务。
 *
 * @param db 数据库连接
 * @param auditService 后台审计服务
 * @return *SystemConfigService 系统配置服务
 */
func NewSystemConfigService(db *gorm.DB, auditService *AdminAuditService) *SystemConfigService {
	if auditService == nil {
		auditService = NewAdminAuditService(db)
	}
	return &SystemConfigService{db: db, auditService: auditService}
}

/**
 * Configs 按分组查询系统配置。
 *
 * @param ctx 请求上下文
 * @param query 查询参数
 * @return []admin.SystemConfigGroup 配置分组
 * @return error 查询失败原因
 */
func (s *SystemConfigService) Configs(ctx context.Context, query admindto.SystemConfigListQuery) ([]admindto.SystemConfigGroup, error) {
	db := s.db.WithContext(ctx).Model(&models.SystemConfig{})
	if query.GroupName != "" {
		db = db.Where("group_name = ?", strings.TrimSpace(query.GroupName))
	}

	var configs []models.SystemConfig
	if err := db.Order("group_name ASC, id ASC").Find(&configs).Error; err != nil {
		return nil, err
	}
	return systemConfigGroups(configs), nil
}

/**
 * Update 更新系统配置值。
 *
 * @param ctx 请求上下文
 * @param operatorID 当前操作者管理员 ID
 * @param id 配置 ID
 * @param req 更新请求
 * @param clientIP 客户端 IP
 * @param userAgent 浏览器 User-Agent
 * @return admin.SystemConfigItem 更新后的配置
 * @return error 更新失败原因
 */
func (s *SystemConfigService) Update(ctx context.Context, operatorID uint64, id uint64, req admindto.SystemConfigUpdateRequest, clientIP string, userAgent string) (admindto.SystemConfigItem, error) {
	var updated models.SystemConfig
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var current models.SystemConfig
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", id).First(&current).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrNotFound.WithMessage("配置不存在")
		}
		if err != nil {
			return err
		}
		value := req.ConfigValue
		if err := tx.Model(&models.SystemConfig{}).Where("id = ?", id).Update("config_value", value).Error; err != nil {
			return err
		}
		if err := tx.Where("id = ?", id).First(&updated).Error; err != nil {
			return err
		}

		input := AdminAuditWriteInput{
			AdminID:    &operatorID,
			Action:     systemConfigUpdateAction,
			ObjectType: systemConfigObjectType,
			ObjectID:   uintIDString(id),
			BeforeData: systemConfigAuditSnapshot(current),
			AfterData:  systemConfigAuditSnapshot(updated),
			IP:         clientIP,
			UserAgent:  userAgent,
			Remark:     "更新系统配置",
		}
		if current.IsSecret {
			input.Action = systemConfigSecretUpdateAction
			return s.auditService.RecordRisk(ctx, tx, AdminRiskWriteInput{
				AdminAuditWriteInput: input,
				RiskLevel:            "high",
				RiskReason:           "修改敏感系统配置",
			})
		}
		return s.auditService.Record(ctx, tx, input)
	}); err != nil {
		return admindto.SystemConfigItem{}, err
	}
	return systemConfigItem(updated), nil
}

func systemConfigGroups(configs []models.SystemConfig) []admindto.SystemConfigGroup {
	groupMap := make(map[string][]admindto.SystemConfigItem)
	groupNames := make([]string, 0)
	for _, config := range configs {
		if _, ok := groupMap[config.GroupName]; !ok {
			groupNames = append(groupNames, config.GroupName)
		}
		groupMap[config.GroupName] = append(groupMap[config.GroupName], systemConfigItem(config))
	}
	sort.Strings(groupNames)

	groups := make([]admindto.SystemConfigGroup, 0, len(groupNames))
	for _, groupName := range groupNames {
		groups = append(groups, admindto.SystemConfigGroup{
			GroupName: groupName,
			Items:     groupMap[groupName],
		})
	}
	return groups
}

func systemConfigItem(config models.SystemConfig) admindto.SystemConfigItem {
	configValue := config.ConfigValue
	if config.IsSecret {
		configValue = nil
	}
	return admindto.SystemConfigItem{
		ID:          config.ID,
		ConfigKey:   config.ConfigKey,
		ConfigValue: configValue,
		ValueType:   config.ValueType,
		GroupName:   config.GroupName,
		IsSecret:    config.IsSecret,
		HasValue:    config.ConfigValue != nil && strings.TrimSpace(*config.ConfigValue) != "",
		Description: config.Description,
		UpdatedAt:   config.UpdatedAt,
	}
}

func systemConfigAuditSnapshot(config models.SystemConfig) map[string]any {
	value := config.ConfigValue
	if config.IsSecret && value != nil {
		masked := adminAuditMaskedValue
		value = &masked
	}
	return map[string]any{
		"id":           config.ID,
		"config_key":   config.ConfigKey,
		"config_value": value,
		"value_type":   config.ValueType,
		"group_name":   config.GroupName,
		"is_secret":    config.IsSecret,
		"has_value":    config.ConfigValue != nil && strings.TrimSpace(*config.ConfigValue) != "",
	}
}
