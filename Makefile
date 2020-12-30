agentd:
	GOOS=linux GOARCH=amd64 go build -o deployments/iptv-agentd iptv-agent.go