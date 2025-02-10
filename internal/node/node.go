package node

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"raft/internal/config"
	"sync"
	"time"
)

type Node struct {
	Name         string
	ID           string
	IsLeader     bool
	Peers        []string
	CurrentTerm  int
	VotedFor     string
	LastActivity time.Time
	Mu           sync.Mutex
}

func NewNode(cfg *config.Config) *Node {
	return &Node{
		ID:           cfg.Node,
		Peers:        cfg.Peers,
		LastActivity: time.Now(),
	}
}

func (n *Node) StartElection(timeout time.Duration) {
	for {
		time.Sleep(time.Duration(rand.Intn(500)+500) * time.Millisecond)
		n.Mu.Lock()
		if n.IsLeader || time.Since(n.LastActivity) < timeout {
			n.Mu.Unlock()
			continue
		}
		n.CurrentTerm++
		n.VotedFor = n.ID
		votes := 1
		n.Mu.Unlock()

		for _, peer := range n.Peers {
			go func(peer string) {
				req, _ := json.Marshal(map[string]interface{}{"term": n.CurrentTerm, "candidate_id": n.ID})
				resp, err := http.Post("http://"+peer+"/request-vote", "application/json", bytes.NewBuffer(req))
				if err != nil {
					return
				}
				defer resp.Body.Close()

				var result map[string]interface{}
				json.NewDecoder(resp.Body).Decode(&result)
				if result["vote_granted"].(bool) {
					n.Mu.Lock()
					votes++
					if votes > len(n.Peers)/2 && !n.IsLeader {
						n.IsLeader = true
						fmt.Printf("Nó %s é agora o líder no termo %d\n", n.ID, n.CurrentTerm)
					}
					n.Mu.Unlock()
				}
			}(peer)
		}
	}
}

func (n *Node) StartHeartbeat(interval time.Duration) {
	for range time.Tick(interval) {
		n.Mu.Lock()
		if n.IsLeader {
			for _, peer := range n.Peers {
				go func(peer string) {
					req, _ := json.Marshal(map[string]interface{}{"term": n.CurrentTerm})
					resp, err := http.Post("http://"+peer+"/heartbeat", "application/json", bytes.NewBuffer(req))
					if err == nil && resp.StatusCode == http.StatusOK {
						fmt.Printf("Lider %s enviou heartbeat para %s\n", n.ID, peer)
					}
				}(peer)
			}
		}
		n.Mu.Unlock()
	}
}

func (n *Node) MonitorLeaderFailure(timeout time.Duration) {
	for range time.Tick(timeout / 2) {
		n.Mu.Lock()
		if !n.IsLeader && time.Since(n.LastActivity) > timeout {
			fmt.Printf("%s iniciou uma nova eleição...\n", n.ID)
			go n.StartElection(timeout)
		}
		n.Mu.Unlock()
	}
}
