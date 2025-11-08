// internal/core/protocol_issue.go

package core

import (
	"fmt"
	"github.com/Jx2f/ViaGenshin/pkg/logger"
)

// ProtocolIssue 协议版本不匹配问题
type ProtocolIssue struct {
	Type         string                 `json:"type"`          // "protocol_version_mismatch"
	Severity     string                 `json:"severity"`      // "high", "medium", "low"
	PacketName   string                 `json:"packet_name"`   // "SceneEntityInfo"
	Description  string                 `json:"description"`   // 问题描述
	Location     string                 `json:"location"`      // 问题位置
	FixStrategy  string                 `json:"fix_strategy"`  // 修复策略标识
	Context      map[string]interface{} `json:"context"`       // 上下文信息
}

// ProtocolIssueDetector 协议问题检测器
type ProtocolIssueDetector struct {
	issues []ProtocolIssue
}

func NewProtocolIssueDetector() *ProtocolIssueDetector {
	return &ProtocolIssueDetector{
		issues: make([]ProtocolIssue, 0),
	}
}

func (d *ProtocolIssueDetector) AddIssue(issue ProtocolIssue) {
	d.issues = append(d.issues, issue)
	logger.Warn("[ProtocolIssue] %s: %s - %s", issue.Severity, issue.Description, issue.Location)
}

func (d *ProtocolIssueDetector) GetIssues() []ProtocolIssue {
	return d.issues
}

func (d *ProtocolIssueDetector) HasIssues() bool {
	return len(d.issues) > 0
}