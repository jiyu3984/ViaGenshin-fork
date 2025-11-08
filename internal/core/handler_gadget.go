// internal/core/handler_gadget.go (完整重写)

package core

import (
	"encoding/json"
	"fmt"
	"github.com/Jx2f/ViaGenshin/internal/mapper"
	"github.com/Jx2f/ViaGenshin/pkg/logger"
)

// ConvertSceneEntityInfo 处理 SceneEntityInfo（包含检测和修复）
func (s *Session) ConvertSceneEntityInfo(from, to mapper.Protocol, data []byte) ([]byte, error) {
	// 解析为 map
	var entityInfo map[string]interface{}
	if err := json.Unmarshal(data, &entityInfo); err != nil {
		logger.Error("Failed to unmarshal SceneEntityInfo: %v", err)
		return data, err
	}

	// 检查是否包含 gadget 字段
	if gadget, ok := entityInfo["gadget"].(map[string]interface{}); ok {
		// 检测并修复 Gadget 字段
		detector := NewProtocolIssueDetector()
		convertedGadget := s.detectAndFixGadget(from, to, gadget, detector)
		
		// 如果有问题，记录日志
		if detector.HasIssues() {
			for _, issue := range detector.GetIssues() {
				logger.Info("[ProtocolFix] %s | %s | %s", 
					issue.Severity, 
					issue.Description, 
					issue.FixStrategy)
			}
		}
		
		entityInfo["gadget"] = convertedGadget
	}

	return json.Marshal(entityInfo)
}

// detectAndFixGadget 检测并修复 Gadget 字段（参考原逻辑）
func (s *Session) detectAndFixGadget(from, to mapper.Protocol, gadget map[string]interface{}, detector *ProtocolIssueDetector) map[string]interface{} {
	// 定义可能的字段名（按优先级）
	possibleFields := []string{"trifleItem", "trifle_item", "item"}
	
	// 检测 gadget 字段
	hasGadget := gadget != nil
	if !hasGadget {
		return gadget
	}
	
	// 遍历检测 trifle_item 相关字段
	for i, fieldName := range possibleFields {
		if _, exists := gadget[fieldName]; exists {
			// 检测到字段存在
			listFieldName := fmt.Sprintf("gadget.%s", fieldName)
			
			// 判断是否需要转换
			if from > to {
				// 新版本 → 旧版本 (v4.2 → v3.2)
				// 检查是否需要从 trifleGadget 提取
				if fieldName == "trifleGadget" || fieldName == "trifle_gadget" {
					detector.AddIssue(ProtocolIssue{
						Type:        "protocol_version_mismatch",
						Severity:    "high",
						PacketName:  "SceneGadgetInfo",
						Description: fmt.Sprintf("检测到3.2版本的%s字段需要转换为5.0版本的trifle_gadget", fieldName),
						Location:    fmt.Sprintf("%%s[%%d].gadget.%%s", listFieldName, i, fieldName),
						FixStrategy: "convert_trifle_item_to_gadget",
						Context: map[string]interface{}{
							"entity_index": 1,
							"field_name":   fieldName,
							"list_name":    listFieldName,
						},
					})
					
					// 执行转换
					return s.convertGadgetNewToOld(gadget)
				}
			} else {
				// 旧版本 → 新版本 (v3.2 → v4.2)
				// 检查是否是 trifleItem
				if fieldName == "trifleItem" || fieldName == "trifle_item" || fieldName == "item" {
					detector.AddIssue(ProtocolIssue{
						Type:        "protocol_version_mismatch",
						Severity:    "high",
						PacketName:  "SceneGadgetInfo",
						Description: fmt.Sprintf("检测到3.2版本的%s字段需要转换为5.0版本的trifle_gadget", fieldName),
						Location:    fmt.Sprintf("%%s[%%d].gadget.%%s", listFieldName, i, fieldName),
						FixStrategy: "convert_trifle_item_to_gadget",
						Context: map[string]interface{}{
							"entity_index": 1,
							"field_name":   fieldName,
							"list_name":    listFieldName,
						},
					})
					
					// 执行转换
					return s.convertGadgetOldToNew(gadget)
				}
			}
			
			// 只处理第一个匹配的字段
			break
		}
	}
	
	return gadget
}

// convertGadgetNewToOld 新版本 → 旧版本（带详细日志）
func (s *Session) convertGadgetNewToOld(gadget map[string]interface{}) map[string]interface{} {
	// 可能的新版本字段名
	newFieldNames := []string{"trifleGadget", "trifle_gadget"}
	
	for _, newField := range newFieldNames {
		if trifleGadget, ok := gadget[newField].(map[string]interface{}); ok {
			// 提取内部的 trifleItem
			oldFieldNames := []string{"trifleItem", "trifle_item", "item"}
			
			for _, oldField := range oldFieldNames {
				if trifleItem, ok := trifleGadget[oldField]; ok {
					// 删除新版本字段
					delete(gadget, newField)
					
					// 添加旧版本字段（统一使用 trifleItem）
					gadget["trifleItem"] = trifleItem
					
					logger.Info("[ProtocolConvert] %s.%s -> trifleItem (v4.2->v3.2)", newField, oldField)
					return gadget
				}
			}
		}
	}
	
	return gadget
}

// convertGadgetOldToNew 旧版本 → 新版本（带详细日志）
func (s *Session) convertGadgetOldToNew(gadget map[string]interface{}) map[string]interface{} {
	// 可能的旧版本字段名
	oldFieldNames := []string{"trifleItem", "trifle_item", "item"}
	
	for _, oldField := range oldFieldNames {
		if trifleItem, ok := gadget[oldField]; ok {
			// 删除旧版本字段
			delete(gadget, oldField)
			
			// 包装成新版本结构（统一使用 trifleGadget）
			gadget["trifleGadget"] = map[string]interface{}{
				"trifleItem": trifleItem,
			}
			
			logger.Info("[ProtocolConvert] %s -> trifleGadget.trifleItem (v3.2->v4.2)", oldField)
			return gadget
		}
	}
	
	return gadget
}