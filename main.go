package main

import (
	"fmt"
	"os"
	"raft/internal/config"
	"raft/internal/node"
	"raft/internal/routes"
	"time"
)

func main() {
	cfg, err := config.LoadConfig(os.Args[1])
	if err != nil {
		fmt.Printf("Erro ao carregar configuração: %v\n", err)
		return
	}

	heartbeatInterval, _ := time.ParseDuration(cfg.HeartbeatInterval)
	leaderTimeout, _ := time.ParseDuration(cfg.LeaderHeartbeatTimeout)

	node := node.NewNode(cfg)
	go routes.StartServer(node, cfg.Port)
	go node.StartElection(leaderTimeout)
	go node.StartHeartbeat(heartbeatInterval)
	go node.MonitorLeaderFailure(leaderTimeout)

	select {}
}
