package job

import (
	"encoding/json"

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

	// Get all inbounds
	inbounds, err := j.inboundService.GetAllInbounds()
	if err != nil {
		logger.Warning("get all inbounds failed:", err)
		return
	}

	// Sync users for each inbound that has v2board enabled
	for _, inbound := range inbounds {
		if !inbound.Enable || !inbound.V2boardEnabled || inbound.V2boardNodeId == "" {
			continue
		}

		err = j.syncInboundUsers(inbound, allSetting)
		if err != nil {
			logger.Warning("sync users for inbound", inbound.Id, "failed:", err)
		}
	}
}

func (j *V2boardSyncJob) syncInboundUsers(inbound *model.Inbound, allSetting *entity.AllSetting) error {
	// Get user list from v2board for this specific node
	userList, err := j.v2boardService.GetUserList(allSetting, inbound.V2boardNodeId, inbound.V2boardNodeType)
	if err != nil {
		return err
	}

	// Get current clients for this inbound
	currentClients, err := j.inboundService.GetClients(inbound)
	if err != nil {
		return err
	}

	// Create a map of existing clients by email
	existingClients := make(map[string]*model.Client)
	for i, client := range currentClients {
		existingClients[client.Email] = &currentClients[i]
	}

	// Sync users - add new clients or update existing ones
	updatedClients := make([]model.Client, 0)
	for _, user := range userList.Users {
		client, exists := existingClients[user.Email]

		if !exists {
			// Create new client
			newClient := model.Client{
				ID:         user.Uuid, // Use UUID from v2board
				Email:      user.Email,
				Enable:     true,
				ExpiryTime: 0, // No expiry
				TotalGB:    0, // Unlimited
				LimitIP:    0, // No limit
				SpeedLimit: user.SpeedLimit,
				//set default flow
				Flow: "xtls-rprx-vision",
			}
			updatedClients = append(updatedClients, newClient)
			logger.Info("added new client for user", user.Id, "in inbound", inbound.Id)
		} else {
			// Update existing client
			client.Enable = true
			client.SpeedLimit = user.SpeedLimit
			updatedClients = append(updatedClients, *client)
		}
	}

	// Disable clients not in V2board list
	v2boardEmails := make(map[string]bool)
	for _, user := range userList.Users {
		v2boardEmails[user.Email] = true
	}

	for _, client := range currentClients {
		if !v2boardEmails[client.Email] {
			client.Enable = false
			updatedClients = append(updatedClients, client)
			logger.Info("disabled client", client.Email, "in inbound", inbound.Id, "as not in V2board")
		}
	}

	// Update inbound settings with new client list
	return j.updateInboundClients(inbound, updatedClients)
}

func (j *V2boardSyncJob) updateInboundClients(inbound *model.Inbound, clients []model.Client) error {
	// Parse current settings
	var settings map[string]interface{}
	if err := json.Unmarshal([]byte(inbound.Settings), &settings); err != nil {
		return err
	}

	// Update clients
	settings["clients"] = clients

	// Serialize back to JSON
	updatedSettings, err := json.Marshal(settings)
	if err != nil {
		return err
	}

	inbound.Settings = string(updatedSettings)

	// Update inbound in database
	_, _, err = j.inboundService.UpdateInbound(inbound)
	return err
}
