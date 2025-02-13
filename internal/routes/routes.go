package routes

import (
	"fmt"
	"net/http"
	"raft/internal/node"
	"time"

	"github.com/gin-gonic/gin"
)

func StartServer(n *node.Node, port int) {

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	r.POST("/request-vote", func(c *gin.Context) {

		var req struct {
			Term        int
			CandidateID string
		}

		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		n.Mu.Lock()
		defer n.Mu.Unlock()

		// se o termo do candidato for maior que o do nó atual da o voto
		if req.Term > n.CurrentTerm {
			n.CurrentTerm, n.VotedFor, n.IsLeader = req.Term, req.CandidateID, false
			c.JSON(http.StatusOK, gin.H{"term": n.CurrentTerm, "vote_granted": true})
		} else {
			// nega o voto.
			c.JSON(http.StatusOK, gin.H{"term": n.CurrentTerm, "vote_granted": false})
		}
	})

	r.POST("/heartbeat", func(c *gin.Context) {

		var req struct{ Term int }

		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		n.Mu.Lock()
		defer n.Mu.Unlock()

		// verifica se o termo do lider for maior ou igual ao atual e atualiza o estado do nó
		if req.Term >= n.CurrentTerm {
			n.CurrentTerm, n.IsLeader, n.LastActivity = req.Term, false, time.Now()
			c.JSON(http.StatusOK, gin.H{"term": n.CurrentTerm, "success": true})
		} else {
			c.JSON(http.StatusOK, gin.H{"term": n.CurrentTerm, "success": false})
		}
	})

	r.Run(fmt.Sprintf(":%d", port))
}
