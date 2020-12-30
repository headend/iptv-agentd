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
