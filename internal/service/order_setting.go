package service

import (
	"encoding/json"
	"time"

	"github.com/dujiao-next/internal/config"
	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/models"
)

const (
	orderConfigFieldPaymentExpireMinutes = "payment_expire_minutes"
	orderConfigFieldMaxRefundDays        = "max_refund_days"

	orderPaymentExpireMinutesDefault = 15
	orderPaymentExpireMinutesMin     = 1
	orderPaymentExpireMinutesMax     = 10080

	orderRefundMaxDaysDefault = 30
	orderRefundMaxDaysMin     = 0
	orderRefundMaxDaysMax     = 3650
)

// OrderConfig 订单配置。
type OrderConfig struct {
	PaymentExpireMinutes int `json:"payment_expire_minutes"`
	MaxRefundDays        int `json:"max_refund_days"`
}

// OrderRefundConfig 订单退款配置。
type OrderRefundConfig struct {
	MaxRefundDays int `json:"max_refund_days"`
}

// DefaultOrderConfig 默认订单配置。
func DefaultOrderConfig() OrderConfig {
	return OrderConfig{
		PaymentExpireMinutes: orderPaymentExpireMinutesDefault,
		MaxRefundDays:        orderRefundMaxDaysDefault,
	}
}

// NormalizeOrderConfig 归一化订单配置。
func NormalizeOrderConfig(cfg OrderConfig) OrderConfig {
	if cfg.PaymentExpireMinutes < orderPaymentExpireMinutesMin {
		cfg.PaymentExpireMinutes = orderPaymentExpireMinutesDefault
	}
	if cfg.PaymentExpireMinutes > orderPaymentExpireMinutesMax {
		cfg.PaymentExpireMinutes = orderPaymentExpireMinutesMax
	}
	if cfg.MaxRefundDays < orderRefundMaxDaysMin {
		cfg.MaxRefundDays = orderRefundMaxDaysDefault
	}
	if cfg.MaxRefundDays > orderRefundMaxDaysMax {
		cfg.MaxRefundDays = orderRefundMaxDaysMax
	}
	return cfg
}

// DefaultOrderRefundConfig 默认订单退款配置。
func DefaultOrderRefundConfig() OrderRefundConfig {
	return OrderRefundConfig{
		MaxRefundDays: DefaultOrderConfig().MaxRefundDays,
	}
}

// NormalizeOrderRefundConfig 归一化订单退款配置。
func NormalizeOrderRefundConfig(cfg OrderRefundConfig) OrderRefundConfig {
	normalized := NormalizeOrderConfig(OrderConfig{
		PaymentExpireMinutes: DefaultOrderConfig().PaymentExpireMinutes,
		MaxRefundDays:        cfg.MaxRefundDays,
	})
	return OrderRefundConfig{MaxRefundDays: normalized.MaxRefundDays}
}

// orderConfigFromJSON 从 JSON map 解析订单配置。
func orderConfigFromJSON(raw models.JSON, fallback OrderConfig) OrderConfig {
	result := NormalizeOrderConfig(fallback)
	if raw == nil {
		return result
	}
	if parsed, err := parseSettingInt(raw[orderConfigFieldPaymentExpireMinutes]); err == nil {
		result.PaymentExpireMinutes = parsed
	}
	if parsed, err := parseSettingInt(raw[orderConfigFieldMaxRefundDays]); err == nil {
		result.MaxRefundDays = parsed
	}
	return NormalizeOrderConfig(result)
}

// OrderConfigToMap 将订单配置转为 map 用于存储。
func OrderConfigToMap(cfg OrderConfig) models.JSON {
	normalized := NormalizeOrderConfig(cfg)
	data, err := json.Marshal(normalized)
	if err != nil {
		return models.JSON{}
	}
	var result models.JSON
	_ = json.Unmarshal(data, &result)
	return result
}

func defaultOrderConfigWithFallback(defaultCfg config.OrderConfig) OrderConfig {
	cfg := DefaultOrderConfig()
	if defaultCfg.PaymentExpireMinutes > 0 {
		cfg.PaymentExpireMinutes = defaultCfg.PaymentExpireMinutes
	}
	return NormalizeOrderConfig(cfg)
}

// GetOrderConfig 获取订单配置。
func (s *SettingService) GetOrderConfig(defaultCfg config.OrderConfig) (OrderConfig, error) {
	fallback := defaultOrderConfigWithFallback(defaultCfg)
	if s == nil {
		return fallback, nil
	}
	value, err := s.GetByKey(constants.SettingKeyOrderConfig)
	if err != nil {
		return fallback, err
	}
	return orderConfigFromJSON(value, fallback), nil
}

// GetOrderRefundConfig 获取订单退款配置。
func (s *SettingService) GetOrderRefundConfig() (OrderRefundConfig, error) {
	fallback := DefaultOrderRefundConfig()
	cfg, err := s.GetOrderConfig(config.OrderConfig{})
	if err != nil {
		return fallback, err
	}
	return OrderRefundConfig{MaxRefundDays: cfg.MaxRefundDays}, nil
}

// isOrderRefundWindowExpired 判断订单是否已超过可退款时间窗口（优先 paid_at，其次 created_at）。
func isOrderRefundWindowExpired(order *models.Order, maxRefundDays int, now time.Time) bool {
	normalizedDays := NormalizeOrderRefundConfig(OrderRefundConfig{
		MaxRefundDays: maxRefundDays,
	}).MaxRefundDays
	if order == nil || normalizedDays == 0 {
		return false
	}
	baseAt := order.CreatedAt
	if order.PaidAt != nil && !order.PaidAt.IsZero() {
		baseAt = *order.PaidAt
	}
	if baseAt.IsZero() {
		return false
	}
	deadline := baseAt.AddDate(0, 0, normalizedDays)
	return now.After(deadline)
}
