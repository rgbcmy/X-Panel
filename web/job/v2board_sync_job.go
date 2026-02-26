package job

import (
	"encoding/json"
	"fmt"

	"x-ui/database/model"
	"x-ui/logger"
	"x-ui/web/entity"
	"x-ui/web/service"
)

type V2boardSyncJob struct {
	settingService service.SettingService
	inboundService service.InboundService
	v2boardService service.V2boardService
	xrayService    service.XrayService
}

func NewV2boardSyncJob() *V2boardSyncJob {
	return &V2boardSyncJob{}
}

func (j *V2boardSyncJob) Run() {
	allSetting, err := j.settingService.GetAllSetting()
	if err != nil {
		logger.Warning("get all setting failed:", err)
		return
	}

	if !allSetting.V2boardEnable {
		return
	}

	// 新策略: 获取用户列表，为每个有vless_config的用户创建/更新inbound
	err = j.syncUserDrivenInbounds(allSetting)
	if err != nil {
		logger.Warning("sync user-driven inbounds failed:", err)
	}
}

// syncUserDrivenInbounds 用户驱动模式：为每个用户创建专属inbound
func (j *V2boardSyncJob) syncUserDrivenInbounds(allSetting *entity.AllSetting) error {
	// 获取用户列表，使用全局配置的nodeId和nodeType
	userList, err := j.v2boardService.GetUserList(allSetting, allSetting.V2boardNodeId, allSetting.V2boardNodeType)
	if err != nil {
		return fmt.Errorf("failed to get user list: %w", err)
	}

	logger.Info("V2board sync: processing", len(userList.Users), "users")

	// 统计处理结果
	processedCount := 0
	skippedCount := 0
	errorCount := 0

	// 遍历用户列表
	for _, user := range userList.Users {
		// 只处理有vless_config的用户
		if user.VlessConfig == nil {
			logger.Debug("user", user.Id, "has no vless_config, skipping")
			skippedCount++
			continue
		}

		err := j.processUserInbound(&user, allSetting)
		if err != nil {
			logger.Warning("failed to process inbound for user", user.Id, ":", err)
			errorCount++
			continue
		}

		logger.Info("successfully processed inbound for user", user.Id, "tag:", user.VlessConfig.InboundTag)
		processedCount++
	}

	logger.Info("V2board sync completed: processed", processedCount, "users, skipped", skippedCount, "users, errors", errorCount)
	return nil
}

// processUserInbound 为单个用户创建或更新inbound
func (j *V2boardSyncJob) processUserInbound(user *service.User, allSetting *entity.AllSetting) error {
	config := user.VlessConfig
	if config == nil {
		return fmt.Errorf("user has no vless_config")
	}

	// 查找是否已存在该tag的inbound
	existingInbound, err := j.getInboundByTag(config.InboundTag)
	if err != nil && err.Error() != "record not found" {
		return fmt.Errorf("failed to query inbound by tag: %w", err)
	}

	if existingInbound != nil {
		// 更新现有inbound
		return j.updateUserInbound(existingInbound, user)
	}

	// 标签未找到，再通过端口查找（多用户共享同一端口的场景）
	portInbound, err := j.getInboundByPort(config.InboundPort)
	if err != nil && err.Error() != "record not found" {
		return fmt.Errorf("failed to query inbound by port: %w", err)
	}

	if portInbound != nil {
		// 端口已被占用，将用户加入该inbound
		logger.Info("port", config.InboundPort, "already in use by inbound", portInbound.Tag, "- adding user", user.Id, "as client")
		return j.updateUserInbound(portInbound, user)
	}

	// 标签和端口均未匹配，创建新inbound
	return j.createUserInbound(user)
}

// getInboundByTag 根据tag查找inbound
func (j *V2boardSyncJob) getInboundByTag(tag string) (*model.Inbound, error) {
	allInbounds, err := j.inboundService.GetAllInbounds()
	if err != nil {
		return nil, err
	}

	for _, inbound := range allInbounds {
		if inbound.Tag == tag {
			return inbound, nil
		}
	}

	return nil, fmt.Errorf("record not found")
}

// getInboundByPort 根据端口查找inbound
func (j *V2boardSyncJob) getInboundByPort(port int) (*model.Inbound, error) {
	allInbounds, err := j.inboundService.GetAllInbounds()
	if err != nil {
		return nil, err
	}

	for _, inbound := range allInbounds {
		if inbound.Port == port {
			return inbound, nil
		}
	}

	return nil, fmt.Errorf("record not found")
}

// createUserInbound 为用户创建新的inbound
func (j *V2boardSyncJob) createUserInbound(user *service.User) error {
	config := user.VlessConfig

	// 构建客户端配置
	client := model.Client{
		ID:         user.Uuid,
		Email:      user.Email,
		Enable:     true,
		ExpiryTime: 0,
		TotalGB:    0,
		LimitIP:    0,
		SpeedLimit: user.SpeedLimit,
		Flow:       config.Flow,
	}

	// 构建VLESS settings
	settings := model.VLESSSettings{
		Clients:    []model.Client{client},
		Decryption: "none",
	}
	settingsJSON, err := json.Marshal(settings)
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	// 构建Reality streamSettings
	// 注意：字段名需要与前端期望的格式一致，参考 web/assets/js/model/inbound.js RealityStreamSettings
	streamSettings := map[string]interface{}{
		"network":  "tcp",
		"security": "reality",
		"realitySettings": map[string]interface{}{
			"show":         false,
			"target":       config.RealityDest + ":443", // 使用target而不是dest，包含端口
			"xver":         0,
			"serverNames":  config.RealityServerNames,
			"privateKey":   config.RealityPrivateKey,
			"minClientVer": "",
			"maxClientVer": "",
			"maxTimediff":  0,
			"shortIds":     []string{config.RealityShortId},
			"mldsa65Seed":  "",
			"settings": map[string]interface{}{
				"publicKey":   config.RealityPublicKey,
				"fingerprint": config.Fingerprint,
				"serverName":  "",
				"spiderX":     config.SpiderX,
			},
		},
	}
	streamSettingsJSON, err := json.Marshal(streamSettings)
	if err != nil {
		return fmt.Errorf("failed to marshal stream settings: %w", err)
	}

	// 构建sniffing配置
	sniffing := map[string]interface{}{
		"enabled":      true,
		"destOverride": []string{"http", "tls", "quic"},
	}
	sniffingJSON, err := json.Marshal(sniffing)
	if err != nil {
		return fmt.Errorf("failed to marshal sniffing: %w", err)
	}

	// 创建inbound对象
	inbound := &model.Inbound{
		UserId:         1, // 设置为管理员用户ID，使其在面板中可见
		Enable:         true,
		Remark:         fmt.Sprintf("V2Board User %d - %s", user.Id, user.Email),
		Listen:         "",
		Port:           config.InboundPort,
		Protocol:       model.VLESS,
		Settings:       string(settingsJSON),
		StreamSettings: string(streamSettingsJSON),
		Tag:            config.InboundTag,
		Sniffing:       string(sniffingJSON),
		// V2Board标记
		V2boardEnabled:  true,
		V2boardNodeId:   fmt.Sprintf("user-%d", user.Id),
		V2boardNodeType: "vless",
	}

	// 添加到数据库
	_, _, err = j.inboundService.AddInbound(inbound)
	if err != nil {
		return fmt.Errorf("failed to add inbound: %w", err)
	}

	logger.Info("created new inbound for user", user.Id, "port:", config.InboundPort, "tag:", config.InboundTag)
	return nil
}

// updateUserInbound 更新现有inbound的用户配置
func (j *V2boardSyncJob) updateUserInbound(inbound *model.Inbound, user *service.User) error {
	// 获取当前客户端列表
	currentClients, err := j.inboundService.GetClients(inbound)
	if err != nil {
		return fmt.Errorf("failed to get clients: %w", err)
	}

	// 查找是否已存在该用户
	userExists := false
	for i, client := range currentClients {
		if client.Email == user.Email || client.ID == user.Uuid {
			// 更新现有客户端
			currentClients[i].ID = user.Uuid
			currentClients[i].Email = user.Email
			currentClients[i].Enable = true
			currentClients[i].SpeedLimit = user.SpeedLimit
			currentClients[i].Flow = user.VlessConfig.Flow
			userExists = true
			break
		}
	}

	// 如果用户不存在，添加新客户端
	if !userExists {
		newClient := model.Client{
			ID:         user.Uuid,
			Email:      user.Email,
			Enable:     true,
			ExpiryTime: 0,
			TotalGB:    0,
			LimitIP:    0,
			SpeedLimit: user.SpeedLimit,
			Flow:       user.VlessConfig.Flow,
		}
		currentClients = append(currentClients, newClient)
		logger.Info("added new client to existing inbound", inbound.Tag, "for user", user.Id)
	} else {
		logger.Debug("updated existing client in inbound", inbound.Tag, "for user", user.Id)
	}

	// 更新inbound配置
	return j.updateInboundClients(inbound, currentClients)
}

// updateInboundClients 更新inbound的客户端列表
func (j *V2boardSyncJob) updateInboundClients(inbound *model.Inbound, clients []model.Client) error {
	// 解析当前settings
	var settings map[string]interface{}
	if err := json.Unmarshal([]byte(inbound.Settings), &settings); err != nil {
		return fmt.Errorf("failed to unmarshal settings: %w", err)
	}

	// 更新客户端列表
	settings["clients"] = clients

	// 序列化回JSON
	updatedSettings, err := json.Marshal(settings)
	if err != nil {
		return fmt.Errorf("failed to marshal updated settings: %w", err)
	}

	inbound.Settings = string(updatedSettings)

	// 更新数据库
	_, _, err = j.inboundService.UpdateInbound(inbound)
	if err != nil {
		return fmt.Errorf("failed to update inbound: %w", err)
	}

	return nil
}
