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

// inicia a eleicaoo quando o nó detecta a falta de um lider
func (n *Node) StartElection(timeout time.Duration) {
	for {
		// tempo aleatorio para iniciar a eleicao
		time.Sleep(time.Duration(rand.Intn(500)+500) * time.Millisecond)

		n.Mu.Lock()
		// Caso ja for lider ou tiver um ativo nao inicia a eleicao
		if n.IsLeader || time.Since(n.LastActivity) < timeout {
			n.Mu.Unlock()
			continue
		}

		// inicia a eleicao
		n.CurrentTerm++
		n.VotedFor = n.ID // Vota em si mesmo
		votes := 1
		n.Mu.Unlock()

		// envia o pedido de voto para cada nó
		for _, peer := range n.Peers {
			go func(peer string) {
				req, _ := json.Marshal(map[string]interface{}{
					"term":         n.CurrentTerm,
					"candidate_id": n.ID,
				})

				// faz a request para solicitar os votos
				resp, err := http.Post("http://"+peer+"/request-vote", "application/json", bytes.NewBuffer(req))
				if err != nil {
					return
				}
				defer resp.Body.Close()

				var result map[string]interface{}
				json.NewDecoder(resp.Body).Decode(&result)

				// se receber voto aumenta a contagem
				if result["vote_granted"].(bool) {
					n.Mu.Lock()
					votes++
					// se a quantidade de votos for maior que a quantidade de no divido por 2 vira lider
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

// envia os Heartbeats para cada nó
func (n *Node) StartHeartbeat(interval time.Duration) {
	for range time.Tick(interval) {
		n.Mu.Lock()

		if n.IsLeader {
			for _, peer := range n.Peers {
				go func(peer string) {
					req, _ := json.Marshal(map[string]interface{}{"term": n.CurrentTerm})
					resp, err := http.Post("http://"+peer+"/heartbeat", "application/json", bytes.NewBuffer(req))
					if err == nil && resp.StatusCode == http.StatusOK {
						fmt.Printf("Líder %s enviou heartbeat para %s\n", n.ID, peer)
					}
				}(peer)
			}
		}
		n.Mu.Unlock()
	}
}

// fica monitorando o lider
func (n *Node) MonitorLeaderFailure(timeout time.Duration) {
	for range time.Tick(timeout / 2) {
		n.Mu.Lock()
		// se nao for lider e nao recebeu heartbeat dentro do tempo
		// inicia uma nova eleicao
		if !n.IsLeader && time.Since(n.LastActivity) > timeout {
			fmt.Printf("%s iniciou uma nova eleição...\n", n.ID)
			go n.StartElection(timeout)
		}
		n.Mu.Unlock()
	}
}
