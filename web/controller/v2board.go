package controller

import (
	"x-ui/web/service"
	"x-ui/web/session"

	"github.com/gin-gonic/gin"
)

type V2boardController struct {
	v2boardService service.V2boardService
	settingService service.SettingService
}

func NewV2boardController(g *gin.RouterGroup) *V2boardController {
	a := &V2boardController{}
	a.initRouter(g)
	return a
}

func (a *V2boardController) initRouter(g *gin.RouterGroup) {
	g = g.Group("/v2board")

	g.POST("/config", a.getServerConfig)
	g.POST("/users", a.getUserList)
	g.POST("/report", a.reportTraffic)
}

func (a *V2boardController) getServerConfig(c *gin.Context) {
	user := session.GetLoginUser(c)
	if user == nil {
		jsonMsg(c, "Unauthorized", nil)
		return
	}

	var req struct {
		NodeId   string `json:"nodeId" binding:"required"`
		NodeType string `json:"nodeType" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		jsonMsg(c, "Invalid request parameters", err)
		return
	}

	allSetting, err := a.settingService.GetAllSetting()
	if err != nil {
		jsonMsg(c, "Failed to get settings", err)
		return
	}

	config, err := a.v2boardService.GetServerConfig(allSetting, req.NodeId, req.NodeType)
	if err != nil {
		jsonMsg(c, "Failed to get server config", err)
		return
	}

	jsonObj(c, config, nil)
}

func (a *V2boardController) getUserList(c *gin.Context) {
	user := session.GetLoginUser(c)
	if user == nil {
		jsonMsg(c, "Unauthorized", nil)
		return
	}

	var req struct {
		NodeId   string `json:"nodeId" binding:"required"`
		NodeType string `json:"nodeType" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		jsonMsg(c, "Invalid request parameters", err)
		return
	}

	allSetting, err := a.settingService.GetAllSetting()
	if err != nil {
		jsonMsg(c, "Failed to get settings", err)
		return
	}

	userList, err := a.v2boardService.GetUserList(allSetting, req.NodeId, req.NodeType)
	if err != nil {
		jsonMsg(c, "Failed to get user list", err)
		return
	}

	jsonObj(c, userList, nil)
}

func (a *V2boardController) reportTraffic(c *gin.Context) {
	user := session.GetLoginUser(c)
	if user == nil {
		jsonMsg(c, "Unauthorized", nil)
		return
	}

	var req struct {
		NodeId   string                `json:"nodeId" binding:"required"`
		NodeType string                `json:"nodeType" binding:"required"`
		Traffic  service.TrafficReport `json:"traffic" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		jsonMsg(c, "Invalid traffic data", err)
		return
	}

	allSetting, err := a.settingService.GetAllSetting()
	if err != nil {
		jsonMsg(c, "Failed to get settings", err)
		return
	}

	err = a.v2boardService.ReportTraffic(allSetting, req.NodeId, req.NodeType, req.Traffic)
	if err != nil {
		jsonMsg(c, "Failed to report traffic", err)
		return
	}

	jsonMsg(c, "Traffic reported successfully", nil)
}
