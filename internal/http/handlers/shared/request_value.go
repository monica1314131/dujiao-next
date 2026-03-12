package shared

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// ParseParamUint 解析路径参数中的正整数 ID。
func ParseParamUint(c *gin.Context, key string) (uint, error) {
	if c == nil {
		return 0, errors.New("context is nil")
	}
	raw := strings.TrimSpace(c.Param(key))
	if raw == "" {
		return 0, errors.New("param is empty")
	}
	parsed, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || parsed == 0 {
		return 0, errors.New("param is invalid")
	}
	return uint(parsed), nil
}

// ParseQueryUint 解析可选查询参数中的正整数 ID。
func ParseQueryUint(raw string, zeroInvalid bool) (uint, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return 0, nil
	}
	parsed, err := strconv.ParseUint(trimmed, 10, 64)
	if err != nil {
		return 0, err
	}
	if zeroInvalid && parsed == 0 {
		return 0, errors.New("query value is invalid")
	}
	return uint(parsed), nil
}

// GetContextUintOrZero 从上下文读取 uint 值，读取失败时返回 0。
func GetContextUintOrZero(c *gin.Context, key string) uint {
	if c == nil {
		return 0
	}
	value, exists := c.Get(key)
	if !exists {
		return 0
	}
	switch v := value.(type) {
	case uint:
		return v
	case int:
		if v > 0 {
			return uint(v)
		}
	case float64:
		if v > 0 {
			return uint(v)
		}
	}
	return 0
}

// GetContextString 从上下文读取字符串值，读取失败时返回空串。
func GetContextString(c *gin.Context, key string) string {
	if c == nil {
		return ""
	}
	value, exists := c.Get(key)
	if !exists {
		return ""
	}
	text, ok := value.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(text)
}
