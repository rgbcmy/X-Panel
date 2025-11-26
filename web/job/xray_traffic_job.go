package job

import (
	"encoding/json"
	"x-ui/database/model"
	"x-ui/logger"
	"x-ui/web/service"
	"x-ui/xray"

	"github.com/valyala/fasthttp"
)

type XrayTrafficJob struct {
	settingService  service.SettingService
	xrayService     service.XrayService
	inboundService  service.InboundService
	outboundService service.OutboundService
	v2boardService  service.V2boardService
}

func NewXrayTrafficJob() *XrayTrafficJob {
	return new(XrayTrafficJob)
}

func (j *XrayTrafficJob) Run() {
	if !j.xrayService.IsXrayRunning() {
		return
	}
	traffics, clientTraffics, err := j.xrayService.GetXrayTraffic()
	if err != nil {
		return
	}
	err, needRestart0 := j.inboundService.AddTraffic(traffics, clientTraffics)
	if err != nil {
		logger.Warning("add inbound traffic failed:", err)
	}
	err, needRestart1 := j.outboundService.AddTraffic(traffics, clientTraffics)
	if err != nil {
		logger.Warning("add outbound traffic failed:", err)
	}
	if ExternalTrafficInformEnable, err := j.settingService.GetExternalTrafficInformEnable(); ExternalTrafficInformEnable {
		j.informTrafficToExternalAPI(traffics, clientTraffics)
	} else if err != nil {
		logger.Warning("get ExternalTrafficInformEnable failed:", err)
	}
	if v2boardEnable, err := j.settingService.GetV2boardEnable(); v2boardEnable {
		j.reportTrafficToV2board(clientTraffics)
	} else if err != nil {
		logger.Warning("get V2boardEnable failed:", err)
	}
	if needRestart0 || needRestart1 {
		j.xrayService.SetToNeedRestart()
	}
}

func (j *XrayTrafficJob) informTrafficToExternalAPI(inboundTraffics []*xray.Traffic, clientTraffics []*xray.ClientTraffic) {
	informURL, err := j.settingService.GetExternalTrafficInformURI()
	if err != nil {
		logger.Warning("get ExternalTrafficInformURI failed:", err)
		return
	}
	requestBody, err := json.Marshal(map[string]any{"clientTraffics": clientTraffics, "inboundTraffics": inboundTraffics})
	if err != nil {
		logger.Warning("parse client/inbound traffic failed:", err)
		return
	}
	request := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(request)
	request.Header.SetMethod("POST")
	request.Header.SetContentType("application/json; charset=UTF-8")
	request.SetBody([]byte(requestBody))
	request.SetRequestURI(informURL)
	response := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(response)
	if err := fasthttp.Do(request, response); err != nil {
		logger.Warning("POST ExternalTrafficInformURI failed:", err)
	}
}

func (j *XrayTrafficJob) reportTrafficToV2board(clientTraffics []*xray.ClientTraffic) {
	allSetting, err := j.settingService.GetAllSetting()
	if err != nil {
		logger.Warning("get all setting failed:", err)
		return
	}

	if !allSetting.V2boardEnable {
		return
	}

	// Group traffic by inbound and then by user
	inboundTraffic := make(map[int]service.TrafficReport) // inboundId -> trafficReport

	for _, clientTraffic := range clientTraffics {
		// Find the inbound for this client
		traffic, inbound, err := j.inboundService.GetClientInboundByTrafficID(clientTraffic.Id)
		if err != nil {
			logger.Warning("get inbound for client", clientTraffic.Id, "failed:", err)
			continue
		}

		// Check if this inbound has v2board enabled
		if !inbound.V2boardEnabled || inbound.V2boardNodeId == "" {
			continue
		}

		// Initialize traffic report for this inbound if not exists
		if inboundTraffic[inbound.Id] == nil {
			inboundTraffic[inbound.Id] = make(service.TrafficReport)
		}

		// Use email as user identifier
		userId := traffic.Email
		if userId == "" {
			continue
		}

		// v2board expects [upload, download] in bytes
		inboundTraffic[inbound.Id][userId] = []int64{clientTraffic.Up, clientTraffic.Down}
	}

	// Report traffic for each inbound
	for inboundId, trafficReport := range inboundTraffic {
		if len(trafficReport) == 0 {
			continue
		}

		// Get inbound info
		allInbounds, err := j.inboundService.GetAllInbounds()
		if err != nil {
			logger.Warning("get all inbounds failed:", err)
			continue
		}

		var targetInbound *model.Inbound
		for _, ib := range allInbounds {
			if ib.Id == inboundId {
				targetInbound = ib
				break
			}
		}

		if targetInbound == nil {
			logger.Warning("inbound not found:", inboundId)
			continue
		}

		err = j.v2boardService.ReportTraffic(allSetting, targetInbound.V2boardNodeId, targetInbound.V2boardNodeType, trafficReport)
		if err != nil {
			logger.Warning("report traffic to v2board for inbound", inboundId, "failed:", err)
		} else {
			logger.Info("successfully reported traffic to v2board for inbound", inboundId)
		}
	}
}
