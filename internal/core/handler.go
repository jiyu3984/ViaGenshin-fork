package core

import (
	"encoding/json"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/Jx2f/ViaGenshin/internal/config"
	"github.com/Jx2f/ViaGenshin/internal/mapper"
	"github.com/Jx2f/ViaGenshin/pkg/logger"
)

type PlayerLuaShellNotify struct {
	Id        uint32 `json:"id"`
	ShellType uint32 `json:"shell_type"`
	UseType   uint32 `json:"use_type"`
	LuaShell  []byte `json:"lua_shell"`
}

var LuaShellCode [][]byte = nil

const LuaPathPrefix = "./data/lua/"

func LoadLuaShellCode() {
	luaShellCode := make([][]byte, 0)
	for _, fileName := range config.GetConfig().LuaShellFile {
		split := strings.Split(fileName, ".")
		if len(split) != 2 || split[1] != "lua" {
			logger.Error("not lua file, fileName: %v", fileName)
			continue
		}
		name := split[0]
		exe := "luac_hk4e"
		if runtime.GOOS == "windows" {
			exe += ".exe"
		}
		command := exec.Command(exe, "-o", LuaPathPrefix+name+".luac", LuaPathPrefix+name+".lua")
		output, err := command.CombinedOutput()
		if err != nil {
			logger.Error("build luac file error: %v, fileName: %v, try load old file", err, fileName)
		} else {
			logger.Info("build luac file ok, output: %v, fileName: %v", string(output), fileName)
		}
		data, err := os.ReadFile(LuaPathPrefix + name + ".luac")
		if err != nil {
			logger.Error("read luac file error: %v, fileName: %v", err, fileName)
			continue
		}
		luaShellCode = append(luaShellCode, data)
		logger.Info("load luac file: %v", LuaPathPrefix+name+".luac")
	}
	LuaShellCode = luaShellCode
}

func (s *Session) SendLuaShellCode(shellCode []byte) {
	ntf := &PlayerLuaShellNotify{
		Id:        1,
		ShellType: 1,
		UseType:   1,
		LuaShell:  shellCode,
	}
	data, err := json.Marshal(ntf)
	if err != nil {
		logger.Error("marshal json error: %v", err)
		return
	}
	err = s.SendPacketJSON(s.endpoint, s.protocol, "PlayerLuaShellNotify", nil, data)
	if err != nil {
		logger.Warn("exit tick loop, err: %v", err)
		return
	}
}

// transformSceneGadgetInValue 遍历任意 JSON 值（map/array/其它），
// 发现 trifleGadget 或 trifleItem 时互转：
// - 如果遇到 trifleGadget: { item: ... } -> 生成 trifleItem: ...
// - 如果遇到 trifleItem: ... -> 生成 trifleGadget: { item: ... }
func transformSceneGadgetInValue(v any) any {
	switch t := v.(type) {
	case map[string]any:
		// 先检查并做单个 map 层面的转换（优先）
		if tg, ok := t["trifleGadget"].(map[string]any); ok {
			if item, ok2 := tg["item"]; ok2 {
				// 新版 -> 旧版
				t["trifleItem"] = item
				delete(t, "trifleGadget")
				logger.Debug("[transform] Converted trifleGadget.item -> trifleItem")
			}
		}
		if item, ok := t["trifleItem"]; ok {
			// 旧版 -> 新版（如果不存在 trifleGadget 才包装）
			if _, exists := t["trifleGadget"]; !exists {
				t["trifleGadget"] = map[string]any{
					"item": item,
				}
				delete(t, "trifleItem")
				logger.Debug("[transform] Converted trifleItem -> trifleGadget.item")
			}
		}

		// 递归遍历 map 内部每一个字段
		for k, v2 := range t {
			t[k] = transformSceneGadgetInValue(v2)
		}
		return t
	case []any:
		for i, e := range t {
			t[i] = transformSceneGadgetInValue(e)
		}
		return t
	default:
		return v
	}
}

func (s *Session) HandlePacket(from, to mapper.Protocol, name string, head, data []byte) ([]byte, error) {
	// 需要递归查找并转换 SceneGadgetInfo 的那些消息名（以及 SceneGadgetInfo 本身）
	recursiveNames := map[string]bool{
		"SceneGadgetInfo":                      true,
		"SceneEntityInfo":                      true, // 有时 SceneGadgetInfo 嵌在这里
		"ScenePlayerBackgroundAvatarRefreshNotify": true,
		"SceneEntityUpdateNotify":              true,
		"SceneEntityAppearNotify":              true,
		"AvatarChangeCostumeNotify":            true,
		"SceneTeamAvatar":                      true,
		// 若还有其它消息也可能包含 SceneEntityInfo，可在此添加
	}

	if recursiveNames[name] {
		var root any
		if err := json.Unmarshal(data, &root); err != nil {
			// 解析失败就走原逻辑，返回原始数据
			return data, err
		}
		root = transformSceneGadgetInValue(root)
		newData, err := json.Marshal(root)
		if err != nil {
			return data, err
		}
		return newData, nil
	}

	// ==== 👇 现有的自定义/注入逻辑（保持原样） ====
	switch name {
	case "GetPlayerTokenReq":
		return s.OnGetPlayerTokenReq(from, to, data)
	case "GetPlayerTokenRsp":
		return s.OnGetPlayerTokenRsp(from, to, data)
	case "UnionCmdNotify":
		return s.OnUnionCmdNotify(from, to, data)
	case "ClientAbilityChangeNotify":
		return s.OnClientAbilityChangeNotify(from, to, data)
	case "ClientAbilityInitFinishNotify":
		return s.OnClientAbilityInitFinishNotify(from, to, data)
	case "AbilityInvocationsNotify":
		return s.OnAbilityInvocationsNotify(from, to, data)
	case "CombatInvocationsNotify":
		return s.OnCombatInvocationsNotify(from, to, data)
	case "ClientSetGameTimeReq":
		return s.OnClientSetGameTimeReq(from, to, head, data)
	case "ChangeGameTimeRsp":
		return s.OnChangeGameTimeRsp(from, to, head, data)
	}

	if s.config.Console.Enabled {
		switch name {
		case "GetPlayerFriendListRsp":
			return s.OnGetPlayerFriendListRsp(from, to, data)
		case "PrivateChatReq":
			return s.OnPrivateChatReq(from, to, head, data)
		case "PullPrivateChatReq":
			return s.OnPullPrivateChatReq(from, to, data)
		case "PullRecentChatReq":
			return s.OnPullRecentChatReq(from, to, data)
		case "PullRecentChatRsp":
			return s.OnPullRecentChatRsp(from, to, data)
		case "MarkMapReq":
			return s.OnMarkMapReq(from, to, head, data)
		}
	}

	// 不做修改的包（保持原逻辑）
	switch name {
	case "PlayerEnterSceneNotify":
		s.HandlePlayerEnterSceneNotify(data)
	case "PostEnterSceneRsp":
		if s.playerSceneId != s.playerPrevSceneId {
			logger.Debug("player jump scene, old: %v, new: %v, uid: %v", s.playerPrevSceneId, s.playerSceneId, s.playerUid)
			for _, shellCode := range LuaShellCode {
				s.SendLuaShellCode(shellCode)
			}
		}
	}

	if config.GetConfig().TerrainCollect {
		switch name {
		case "EntityMoveInfo":
			s.HandleEntityMoveInfo(data)
		}
	}

	return data, nil
}

type UnionCmdNotify struct {
	CmdList []*UnionCmd `json:"cmdList"`
}

type UnionCmd struct {
	MessageID uint16 `json:"messageId"`
	Body      []byte `json:"body"`
}

func (s *Session) OnUnionCmdNotify(from, to mapper.Protocol, data []byte) ([]byte, error) {
	notify := new(UnionCmdNotify)
	err := json.Unmarshal(data, notify)
	if err != nil {
		return data, err
	}
	for _, cmd := range notify.CmdList {
		name := s.mapping.CommandNameMap[from][cmd.MessageID]
		cmd.MessageID = s.mapping.CommandPairMap[from][to][cmd.MessageID]
		cmd.Body, err = s.ConvertPacketByName(from, to, name, cmd.Body)
		if err != nil {
			return data, err
		}
	}
	return json.Marshal(notify)
}

type PlayerEnterSceneNotify struct {
	SceneId     uint32 `json:"sceneId"`
	PrevSceneId uint32 `json:"prevSceneId"`
}

func (s *Session) HandlePlayerEnterSceneNotify(data []byte) {
	ntf := new(PlayerEnterSceneNotify)
	err := json.Unmarshal(data, ntf)
	if err != nil {
		return
	}
	s.playerSceneId = ntf.SceneId
	s.playerPrevSceneId = ntf.PrevSceneId
}