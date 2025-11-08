package core

import (
	"github.com/Jx2f/ViaGenshin/pkg/logger"
)

// ApplyFix 根据 FixStrategy 应用修复
func ApplyFix(issue ProtocolIssue, data map[string]interface{}) bool {
	switch issue.FixStrategy {
	case "convert_trifle_item_to_gadget":
		// 这个已经在检测时自动修复了
		logger.Debug("[ProtocolFix] Applied fix: %s", issue.FixStrategy)
		return true
	
	default:
		logger.Warn("[ProtocolFix] Unknown fix strategy: %s", issue.FixStrategy)
		return false
	}
}