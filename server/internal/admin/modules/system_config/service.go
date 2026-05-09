package systemconfig

import (
	"context"
	"errors"
	"sort"
	"strconv"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	admindto "github.com/AeolianCloud/pveCloud/server/internal/admin/dto"
	"github.com/AeolianCloud/pveCloud/server/internal/admin/models"
	apperrors "github.com/AeolianCloud/pveCloud/server/internal/shared/errors"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/textutil"
)

const (
	systemConfigObjectType         = "system_config"
	systemConfigUpdateAction       = "system.config.update"
	systemConfigSecretUpdateAction = "system.config.secret_update"
	adminAuditMaskedValue          = "[已脱敏]"
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
 * @return admin.SystemConfigItem 更新后的配置
 * @return error 更新失败原因
 */
func (s *SystemConfigService) Update(ctx context.Context, operatorID uint64, id uint64, req admindto.SystemConfigUpdateRequest) (admindto.SystemConfigItem, error) {
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
		if err := validateConfigValue(current.ValueType, value); err != nil {
			return err
		}
		if err := s.validateRealNameConfigUpdate(ctx, tx, current, value); err != nil {
			return err
		}
		if current.IsSecret && strings.TrimSpace(value) == "" {
			updated = current
		} else {
			if err := tx.Model(&models.SystemConfig{}).Where("id = ?", id).Update("config_value", value).Error; err != nil {
				return err
			}
			if err := tx.Where("id = ?", id).First(&updated).Error; err != nil {
				return err
			}
		}
		input := AdminAuditWriteInput{
			AdminID:    &operatorID,
			Action:     systemConfigUpdateAction,
			ObjectType: systemConfigObjectType,
			ObjectID:   textutil.Uint64String(id),
			BeforeData: systemConfigAuditSnapshot(current),
			AfterData:  systemConfigAuditSnapshot(updated),
			Remark:     "更新系统配置",
		}
		if current.IsSecret {
			input.Action = systemConfigSecretUpdateAction
		}
		return s.auditService.Record(ctx, tx, input)
	}); err != nil {
		return admindto.SystemConfigItem{}, err
	}
	return systemConfigItem(updated), nil
}

func validateConfigValue(valueType string, value string) error {
	switch strings.TrimSpace(valueType) {
	case "bool":
		trimmed := strings.TrimSpace(value)
		if trimmed != "true" && trimmed != "false" {
			return apperrors.ErrValidation.WithMessage("布尔配置只能填写 true 或 false")
		}
	case "int":
		parsed, err := strconv.Atoi(strings.TrimSpace(value))
		if err != nil || parsed <= 0 {
			return apperrors.ErrValidation.WithMessage("数字配置必须为正整数")
		}
	}
	return nil
}

func (s *SystemConfigService) validateRealNameConfigUpdate(ctx context.Context, tx *gorm.DB, current models.SystemConfig, value string) error {
	key := strings.TrimSpace(current.ConfigKey)
	trimmedValue := strings.TrimSpace(value)
	if current.IsSecret && trimmedValue == "" {
		return nil
	}
	switch key {
	case "real_name.allowed_providers":
		for _, provider := range splitProviders(trimmedValue) {
			if provider != "alipay" && provider != "wechat" {
				return apperrors.ErrValidation.WithMessage("实名供应商只允许 alipay 或 wechat")
			}
		}
	case "real_name.default_provider":
		if trimmedValue != "" && trimmedValue != "alipay" && trimmedValue != "wechat" && trimmedValue != "manual" {
			return apperrors.ErrValidation.WithMessage("默认实名方式只允许 alipay、wechat 或 manual")
		}
	case "real_name.identity_digest_secret":
		if err := ensureIdentityDigestSecretMutable(ctx, tx, current, trimmedValue); err != nil {
			return err
		}
	case "real_name.alipay.enabled":
		if trimmedValue == "true" {
			configs, err := systemConfigValueMap(ctx, tx, current, value)
			if err != nil {
				return err
			}
			if !alipayConfigComplete(configs) {
				return apperrors.ErrValidation.WithMessage("启用支付宝实名前必须补齐应用ID、网关、私钥、公钥、返回地址和回调地址")
			}
		}
	case "real_name.wechat.enabled":
		if trimmedValue == "true" {
			configs, err := systemConfigValueMap(ctx, tx, current, value)
			if err != nil {
				return err
			}
			if !wechatConfigComplete(configs) {
				return apperrors.ErrValidation.WithMessage("启用微信实名前必须补齐腾讯云密钥、地域、端点、规则ID和返回地址")
			}
		}
	}
	return nil
}

func ensureIdentityDigestSecretMutable(ctx context.Context, tx *gorm.DB, current models.SystemConfig, nextValue string) error {
	currentValue := ""
	if current.ConfigValue != nil {
		currentValue = strings.TrimSpace(*current.ConfigValue)
	}
	if nextValue == "" || currentValue == "" || nextValue == currentValue {
		return nil
	}
	var count int64
	if err := tx.WithContext(ctx).Model(&models.UserRealNameApplication{}).
		Where("id_number_digest_version = ?", "hmac-sha256-v1").
		Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return apperrors.ErrConflict.WithMessage("已有当前HMAC版本实名记录，不能直接修改证件摘要密钥")
	}
	return nil
}

func systemConfigValueMap(ctx context.Context, tx *gorm.DB, current models.SystemConfig, nextValue string) (map[string]string, error) {
	var configs []models.SystemConfig
	if err := tx.WithContext(ctx).Where("config_key LIKE ?", "real_name.%").Find(&configs).Error; err != nil {
		return nil, err
	}
	result := make(map[string]string, len(configs))
	for _, config := range configs {
		value := ""
		if config.ConfigValue != nil {
			value = strings.TrimSpace(*config.ConfigValue)
		}
		if config.ID == current.ID {
			value = strings.TrimSpace(nextValue)
		}
		result[config.ConfigKey] = value
	}
	return result, nil
}

func alipayConfigComplete(configs map[string]string) bool {
	return strings.TrimSpace(configs["real_name.alipay.app_id"]) != "" &&
		strings.TrimSpace(configs["real_name.alipay.gateway_url"]) != "" &&
		strings.TrimSpace(configs["real_name.alipay.return_url"]) != "" &&
		(strings.TrimSpace(configs["real_name.alipay.notify_url"]) != "" || strings.TrimSpace(configs["real_name.callback_base_url"]) != "") &&
		strings.TrimSpace(configs["real_name.alipay.app_private_key"]) != "" &&
		strings.TrimSpace(configs["real_name.alipay.alipay_public_key"]) != ""
}

func wechatConfigComplete(configs map[string]string) bool {
	return strings.TrimSpace(configs["real_name.wechat.secret_id"]) != "" &&
		strings.TrimSpace(configs["real_name.wechat.secret_key"]) != "" &&
		strings.TrimSpace(configs["real_name.wechat.region"]) != "" &&
		strings.TrimSpace(configs["real_name.wechat.endpoint"]) != "" &&
		strings.TrimSpace(configs["real_name.wechat.rule_id"]) != "" &&
		strings.TrimSpace(configs["real_name.wechat.redirect_url"]) != ""
}

func splitProviders(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		item := strings.ToLower(strings.TrimSpace(part))
		if item != "" {
			result = append(result, item)
		}
	}
	return result
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
