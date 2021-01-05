package utils

import (
	"github.com/headend/share-module/configuration"
	static_config "github.com/headend/share-module/configuration/static-config"
)

func GetMasterConnectionInfo(conf configuration.Conf) (string, uint16) {
	var masterHost string
	if conf.Agentd.Master.Host != "" {
		masterHost = conf.Agentd.Master.Host
	} else {
		masterHost = static_config.MasterHost
	}

	var masterPort uint16
	if conf.AgentGateway.Port != 0 {
		masterPort = conf.Agentd.Master.Port
	} else {
		masterPort = static_config.MasterPort
	}
	return masterHost, masterPort
}


func GetGatewayPort(conf configuration.Conf) uint16 {
	var gwPort uint16
	if conf.AgentGateway.Port != 0 {
		gwPort = conf.AgentGateway.Port
	} else {
		gwPort = static_config.GatewayPort
	}
	return gwPort
}

func GetGatewayHost(conf configuration.Conf) string {
	var gwHost string
	if conf.AgentGateway.Gateway != "" {
		gwHost = conf.AgentGateway.Gateway
	} else {
		if conf.AgentGateway.Host != "" {
			gwHost = conf.AgentGateway.Host
		} else {
			gwHost = static_config.GatewayHost
		}
	}
	return gwHost
}