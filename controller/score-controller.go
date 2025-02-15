package controller

import (
	"bpl/client"
	"bpl/scoring"
	"bpl/service"
	"bpl/utils"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type ScoreController struct {
	scoringCategoryService *service.ScoringCategoryService
	eventService           *service.EventService
	scoreService           *scoring.ScoreService
	poeClient              *client.PoEClient
	mu                     sync.Mutex
	connections            map[int]map[*websocket.Conn]bool
	cancelFetching         context.CancelFunc
	cancelEvaluating       context.CancelFunc
}

func NewScoreController() *ScoreController {
	scoringCategoryService := service.NewScoringCategoryService()
	eventService := service.NewEventService()
	poeClient := client.NewPoEClient(os.Getenv("POE_CLIENT_AGENT"), 10, false, 10)
	controller := &ScoreController{
		scoringCategoryService: scoringCategoryService,
		eventService:           eventService,
		scoreService:           scoring.NewScoreService(),
		poeClient:              poeClient,
		connections:            make(map[int]map[*websocket.Conn]bool),
		cancelFetching:         nil,
		cancelEvaluating:       nil,
	}
	controller.StartScoreUpdater()
	return controller
}

func setupScoreController() []RouteInfo {
	e := NewScoreController()
	baseUrl := "events/:event_id/scores"
	routes := []RouteInfo{
		{Method: "GET", Path: "/latest", HandlerFunc: e.getLatestScoresForEventHandler()},
		{Method: "POST", Path: "/fetch/:seconds", HandlerFunc: e.FetchStashChangesHandler()},
		{Method: "POST", Path: "/evaluate/:seconds", HandlerFunc: e.EvaluateStashChangesHandler()},
		{Method: "GET", Path: "/ws", HandlerFunc: e.WebSocketHandler},
	}
	for i, route := range routes {
		routes[i].Path = baseUrl + route.Path
	}
	return routes
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// allow any host origin to connect to the websocket
		return true
	},
}

func (e *ScoreController) WebSocketHandler(c *gin.Context) {
	eventID, err := strconv.Atoi(c.Param("event_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		http.NotFound(c.Writer, c.Request)
		return
	}
	defer conn.Close()

	e.mu.Lock()
	if _, ok := e.connections[eventID]; !ok {
		e.connections[eventID] = make(map[*websocket.Conn]bool)
	}
	e.connections[eventID][conn] = true
	e.mu.Unlock()

	if _, ok := e.scoreService.LatestScores[eventID]; !ok {
		e.scoreService.LatestScores[eventID] = make(scoring.ScoreMap)
	}
	// Send the latest score to the new subscriber
	serialized, err := json.Marshal(toScoreMapResponse(e.scoreService.LatestScores[eventID]))
	if err != nil {
		return
	}

	if err := conn.WriteMessage(websocket.TextMessage, serialized); err != nil {
		return
	}

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			e.mu.Lock()
			delete(e.connections[eventID], conn)
			if len(e.connections[eventID]) == 0 {
				delete(e.connections, eventID)
			}
			e.mu.Unlock()
			return
		}
	}
}

func (e *ScoreController) StartScoreUpdater() {
	go func() {
		for {
			e.mu.Lock()
			// calculate scores for events with active websocket connections
			for eventID, conns := range e.connections {
				if len(utils.Values(conns)) == 0 {
					continue
				}
				diff, err := e.scoreService.GetNewDiff(eventID)
				if err != nil {
					continue
				}
				serializedDiff, err := json.Marshal(toScoreMapResponse(diff))
				if err != nil {
					log.Fatal(err)
					continue
				}
				for conn := range conns {
					if err := conn.WriteMessage(websocket.TextMessage, serializedDiff); err != nil {
						conn.Close()
						delete(conns, conn)
					}
				}
			}
			e.mu.Unlock()
			time.Sleep(5 * time.Second)
		}
	}()
}

// @id GetLatestScoresForEvent
// @Description Fetches the latest scores for the current event
// @Tags scores
// @Produce json
// @Success 200 {array} ScoreResponse
// @Param event_id path int true "Event ID"
// @Router /events/{event_id}/scores/latest [get]
func (e *ScoreController) getLatestScoresForEventHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		eventID, err := strconv.Atoi(c.Param("event_id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
			return
		}
		scores := e.scoreService.LatestScores[eventID]
		if scores == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "No scores found for event"})
			return
		}
		c.JSON(http.StatusOK, scores)
	}
}

func (e *ScoreController) FetchStashChangesHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		seconds, err := strconv.Atoi(c.Param("seconds"))
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		eventId, err := strconv.Atoi(c.Param("event_id"))
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		event, err := e.eventService.GetEventById(eventId)
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		if e.cancelFetching != nil {
			e.cancelFetching()
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(seconds)*time.Second)
		e.cancelFetching = cancel
		scoring.FetchLoop(ctx, event, e.poeClient)
		c.JSON(200, gin.H{"message": "Stash change fetch started"})
	}
}
func (e *ScoreController) EvaluateStashChangesHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		seconds, err := strconv.Atoi(c.Param("seconds"))
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		eventId, err := strconv.Atoi(c.Param("event_id"))
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		event, err := e.eventService.GetEventById(eventId, "Teams", "Teams.Users")
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		if e.cancelEvaluating != nil {
			e.cancelEvaluating()
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(seconds)*time.Second)
		e.cancelEvaluating = cancel
		ctx.Done()
		err = scoring.StashLoop(ctx, e.poeClient, event)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"message": "Stash change evaluation started"})
	}
}

type ScoreResponse struct {
	Points    int       `json:"points" binding:"required"`
	UserID    int       `json:"user_id" binding:"required"`
	Rank      int       `json:"rank" binding:"required"`
	Timestamp time.Time `json:"timestamp" binding:"required"`
	Number    int       `json:"number" binding:"required"`
	Finished  bool      `json:"finished" binding:"required"`
}

type ScoreDiffResponse struct {
	Score     *ScoreResponse   `json:"score"`
	FieldDiff []string         `json:"field_diff"`
	DiffType  scoring.Difftype `json:"diff_type"`
}

type ScoreMapResponse map[string]*ScoreDiffResponse

func toScoreDiffResponse(scoreDiff *scoring.ScoreDifference) *ScoreDiffResponse {
	return &ScoreDiffResponse{
		Score:     toScoreResponse(scoreDiff.Score),
		FieldDiff: scoreDiff.FieldDiff,
		DiffType:  scoreDiff.DiffType,
	}
}

func toScoreMapResponse(scoreMap scoring.ScoreMap) ScoreMapResponse {
	response := make(ScoreMapResponse)
	for id, score := range scoreMap {
		response[id] = toScoreDiffResponse(score)
	}
	return response
}

func toScoreResponse(score *scoring.Score) *ScoreResponse {
	return &ScoreResponse{
		Points:    score.Points,
		UserID:    score.UserID,
		Rank:      score.Rank,
		Timestamp: score.Timestamp,
		Number:    score.Number,
		Finished:  score.Finished,
	}
}
