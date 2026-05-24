package systemconfig

import (
	"context"
	"errors"
	"sort"
	"strings"

	"gorm.io/gorm"

	domainsystemconfig "github.com/AeolianCloud/pveCloud/server/internal/domain/systemconfig"
	integrationpayment "github.com/AeolianCloud/pveCloud/server/internal/integration/payment"
	mysqlsystemconfig "github.com/AeolianCloud/pveCloud/server/internal/repository/mysql/systemconfig"
	mysqltx "github.com/AeolianCloud/pveCloud/server/internal/repository/mysql/tx"
	apperrors "github.com/AeolianCloud/pveCloud/server/internal/shared/errors"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/textutil"
	admindto "github.com/AeolianCloud/pveCloud/server/internal/usecase/admin/dto"
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
	configs      *mysqlsystemconfig.Repository
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
	return &SystemConfigService{
		db:           db,
		configs:      mysqlsystemconfig.NewRepository(db),
		auditService: auditService,
	}
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
	configs, err := s.configs.ConfigRows(ctx, strings.TrimSpace(query.GroupName))
	if err != nil {
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
	var updated mysqlsystemconfig.SystemConfig
	if err := mysqltx.NewManager(s.db).WithinContext(ctx, func(tx *gorm.DB) error {
		current, err := s.configs.FindByIDForUpdate(ctx, tx, id)
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
		if err := s.validatePaymentConfigUpdate(ctx, tx, current, value); err != nil {
			return err
		}
		if domainsystemconfig.PreserveSecretWhenBlank(current.IsSecret, value) {
			updated = current
		} else {
			updatedConfig, err := s.configs.UpdateValueAndReload(ctx, tx, id, value)
			if err != nil {
				return err
			}
			updated = updatedConfig
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

func (s *SystemConfigService) validatePaymentConfigUpdate(ctx context.Context, tx *gorm.DB, current mysqlsystemconfig.SystemConfig, value string) error {
	key := strings.TrimSpace(current.ConfigKey)
	trimmedValue := strings.TrimSpace(value)
	if current.IsSecret && trimmedValue == "" {
		return nil
	}
	if key != "payment.alipay.enabled" && key != "payment.wechat.enabled" {
		return nil
	}
	if trimmedValue != "true" {
		return nil
	}
	configs, err := s.paymentConfigValueMap(ctx, tx, current, value)
	if err != nil {
		return err
	}
	provider := integrationpayment.ProviderAlipay
	method := integrationpayment.MethodAlipayPage
	if key == "payment.wechat.enabled" {
		provider = integrationpayment.ProviderWechat
		method = integrationpayment.MethodWechatH5
	}
	if err := integrationpayment.ValidateProductionConfig(integrationpayment.Config{Provider: provider, Values: configs}, method); err != nil {
		return apperrors.ErrValidation.WithMessage("启用支付渠道前必须补齐商户号、密钥、公钥、回调地址和支付场景配置")
	}
	return nil
}

func validateConfigValue(valueType string, value string) error {
	if err := domainsystemconfig.ValidatePrimitiveValue(valueType, value); err != nil {
		return apperrors.ErrValidation.WithMessage(err.Error())
	}
	return nil
}

func (s *SystemConfigService) validateRealNameConfigUpdate(ctx context.Context, tx *gorm.DB, current mysqlsystemconfig.SystemConfig, value string) error {
	key := strings.TrimSpace(current.ConfigKey)
	trimmedValue := strings.TrimSpace(value)
	if current.IsSecret && trimmedValue == "" {
		return nil
	}
	switch key {
	case "real_name.allowed_providers":
		if err := domainsystemconfig.ValidateRealNameProviders(trimmedValue); err != nil {
			return apperrors.ErrValidation.WithMessage(err.Error())
		}
	case "real_name.default_provider":
		if trimmedValue != "" && trimmedValue != "alipay" && trimmedValue != "wechat" && trimmedValue != "manual" {
			return apperrors.ErrValidation.WithMessage("默认实名方式只允许 alipay、wechat 或 manual")
		}
	case "real_name.identity_digest_secret":
		if err := s.ensureIdentityDigestSecretMutable(ctx, tx, current, trimmedValue); err != nil {
			return err
		}
	case "real_name.alipay.enabled":
		if trimmedValue == "true" {
			configs, err := s.systemConfigValueMap(ctx, tx, current, value)
			if err != nil {
				return err
			}
			if !domainsystemconfig.AlipayConfigComplete(configs) {
				return apperrors.ErrValidation.WithMessage("启用支付宝实名前必须补齐应用ID、网关、私钥、公钥、返回地址和回调地址")
			}
		}
	case "real_name.wechat.enabled":
		if trimmedValue == "true" {
			configs, err := s.systemConfigValueMap(ctx, tx, current, value)
			if err != nil {
				return err
			}
			if !domainsystemconfig.WechatConfigComplete(configs) {
				return apperrors.ErrValidation.WithMessage("启用微信实名前必须补齐腾讯云密钥、地域、端点、规则ID和返回地址")
			}
		}
	}
	return nil
}

func (s *SystemConfigService) ensureIdentityDigestSecretMutable(ctx context.Context, tx *gorm.DB, current mysqlsystemconfig.SystemConfig, nextValue string) error {
	currentValue := ""
	if current.ConfigValue != nil {
		currentValue = strings.TrimSpace(*current.ConfigValue)
	}
	if nextValue == "" || currentValue == "" || nextValue == currentValue {
		return nil
	}
	count, err := s.configs.CountRealNameApplicationsByDigestVersion(ctx, tx, "hmac-sha256-v1")
	if err != nil {
		return err
	}
	if count > 0 {
		return apperrors.ErrConflict.WithMessage("已有当前HMAC版本实名记录，不能直接修改证件摘要密钥")
	}
	return nil
}

func (s *SystemConfigService) systemConfigValueMap(ctx context.Context, tx *gorm.DB, current mysqlsystemconfig.SystemConfig, nextValue string) (map[string]string, error) {
	configs, err := s.configs.RealNameConfigRows(ctx, tx)
	if err != nil {
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

func (s *SystemConfigService) paymentConfigValueMap(ctx context.Context, tx *gorm.DB, current mysqlsystemconfig.SystemConfig, nextValue string) (map[string]string, error) {
	configs, err := s.configs.ConfigRowsByPrefix(ctx, tx, "payment.")
	if err != nil {
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

func systemConfigGroups(configs []mysqlsystemconfig.SystemConfig) []admindto.SystemConfigGroup {
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

func systemConfigItem(config mysqlsystemconfig.SystemConfig) admindto.SystemConfigItem {
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

func systemConfigAuditSnapshot(config mysqlsystemconfig.SystemConfig) map[string]any {
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
