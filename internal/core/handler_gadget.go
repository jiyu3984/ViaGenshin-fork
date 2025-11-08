// internal/core/handler_gadget.go

package core

import (
	"encoding/json"
	"github.com/Jx2f/ViaGenshin/internal/mapper"
	"github.com/Jx2f/ViaGenshin/pkg/logger"
)

// ConvertSceneEntityInfo 处理 SceneEntityInfo（包含 SceneGadgetInfo）
func (s *Session) ConvertSceneEntityInfo(from, to mapper.Protocol, data []byte) ([]byte, error) {
	// 解析为 map
	var entityInfo map[string]interface{}
	if err := json.Unmarshal(data, &entityInfo); err != nil {
		logger.Error("Failed to unmarshal SceneEntityInfo: %v", err)
		return data, err
	}

	// 检查是否包含 gadget 字段
	if gadget, ok := entityInfo["gadget"].(map[string]interface{}); ok {
		// 转换 SceneGadgetInfo
		convertedGadget, err := s.convertSceneGadgetInfo(from, to, gadget)
		if err != nil {
			logger.Error("Failed to convert gadget: %v", err)
			return data, err
		}
		entityInfo["gadget"] = convertedGadget
		
		logger.Debug("Converted SceneEntityInfo.gadget successfully")
	}

	return json.Marshal(entityInfo)
}

// convertSceneGadgetInfo 转换 SceneGadgetInfo 内部逻辑
func (s *Session) convertSceneGadgetInfo(from, to mapper.Protocol, gadgetInfo map[string]interface{}) (map[string]interface{}, error) {
	// 判断转换方向
	if from > to {
		// 新 → 旧 (v4.2 → v3.2): TrifleGadget → TrifleItem
		return s.convertGadgetNewToOld(gadgetInfo), nil
	} else {
		// 旧 → 新 (v3.2 → v4.2): TrifleItem → TrifleGadget
		return s.convertGadgetOldToNew(gadgetInfo), nil
	}
}

// convertGadgetNewToOld 新版本 → 旧版本
// v4.2: trifleGadget { trifleItem: {...} }
// v3.2: trifleItem: {...}
func (s *Session) convertGadgetNewToOld(gadgetInfo map[string]interface{}) map[string]interface{} {
	// 检查是否存在 trifleGadget
	if trifleGadget, ok := gadgetInfo["trifleGadget"].(map[string]interface{}); ok {
		// 提取内部的 trifleItem
		if trifleItem, ok := trifleGadget["trifleItem"]; ok {
			// 删除新版本字段
			delete(gadgetInfo, "trifleGadget")
			
			// 添加旧版本字段
			gadgetInfo["trifleItem"] = trifleItem
			
			logger.Debug("Converted: trifleGadget.trifleItem -> trifleItem (v4.2->v3.2)")
		} else {
			logger.Warn("trifleGadget exists but no trifleItem inside, data: %v", trifleGadget)
		}
	}
	
	return gadgetInfo
}

// convertGadgetOldToNew 旧版本 → 新版本
// v3.2: trifleItem: {...}
// v4.2: trifleGadget { trifleItem: {...} }
func (s *Session) convertGadgetOldToNew(gadgetInfo map[string]interface{}) map[string]interface{} {
	// 检查是否存在 trifleItem
	if trifleItem, ok := gadgetInfo["trifleItem"]; ok {
		// 删除旧版本字段
		delete(gadgetInfo, "trifleItem")
		
		// 包装成新版本结构
		gadgetInfo["trifleGadget"] = map[string]interface{}{
			"trifleItem": trifleItem,
			// 注意：如果 TrifleGadget 还有其他字段（如 LKLPOHNKLNF），
			// 可能需要添加默认值，目前先不添加
		}
		
		logger.Debug("Converted: trifleItem -> trifleGadget.trifleItem (v3.2->v4.2)")
	}
	
	return gadgetInfo
}