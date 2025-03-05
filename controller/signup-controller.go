package controller

import (
	"bpl/repository"
	"bpl/service"
	"bpl/utils"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type SignupController struct {
	signupService *service.SignupService
	userService   *service.UserService
}

func NewSignupController() *SignupController {
	return &SignupController{
		signupService: service.NewSignupService(),
		userService:   service.NewUserService(),
	}
}

func setupSignupController() []RouteInfo {
	e := NewSignupController()
	basePath := "/events/:event_id/signups"
	routes := []RouteInfo{
		{Method: "GET", Path: "", HandlerFunc: e.getEventSignupsHandler(), Authenticated: true},
		{Method: "GET", Path: "/self", HandlerFunc: e.getPersonalSignupHandler(), Authenticated: true},
		{Method: "PUT", Path: "/self", HandlerFunc: e.createSignupHandler(), Authenticated: true},
		{Method: "DELETE", Path: "/self", HandlerFunc: e.deleteSignupHandler(), Authenticated: true},
	}
	for i, route := range routes {
		routes[i].Path = basePath + route.Path
	}
	return routes
}

// @id GetPersonalSignup
// @Description Fetches an authenticated user's signup for the event
// @Tags signup
// @Produce json
// @Success 200 {object} Signup
// @Param event_id path int true "Event Id"
// @Router /events/{event_id}/signups/self [get]
func (e *SignupController) getPersonalSignupHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		eventId, err := strconv.Atoi(c.Param("event_id"))
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		user, err := e.userService.GetUserFromAuthCookie(c)
		if err != nil {
			c.JSON(401, gin.H{"error": "Not authenticated"})
			return
		}
		signup, err := e.signupService.GetSignupForUser(user.Id, eventId)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(404, gin.H{"error": "Not signed up"})
			} else {
				c.JSON(400, gin.H{"error": err.Error()})
			}
			return
		}
		c.JSON(200, toSignupResponse(signup))
	}
}

// @id CreateSignup
// @Description Creates a signup for the authenticated user
// @Tags signup
// @Accept json
// @Produce json
// @Success 201 {object} Signup
// @Param event_id path int true "Event Id"
// @Param body body SignupCreate true "Signup"
// @Router /events/{event_id}/signups/self [put]
func (e *SignupController) createSignupHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		eventId, err := strconv.Atoi(c.Param("event_id"))
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		user, err := e.userService.GetUserFromAuthCookie(c)
		if err != nil {
			c.JSON(401, gin.H{"error": "Not authenticated"})
			return
		}
		err = e.userService.DiscordServerCheck(user)
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		var signupCreate SignupCreate
		if err := c.BindJSON(&signupCreate); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		signup := &repository.Signup{
			UserId:           user.Id,
			EventId:          eventId,
			Timestamp:        time.Now(),
			ExpectedPlayTime: signupCreate.ExpectedPlaytime,
		}
		signup, err = e.signupService.CreateSignup(signup)
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		c.JSON(201, toSignupResponse(signup))
	}
}

// @id DeleteSignup
// @Description Deletes the authenticated user's signup for the event
// @Tags signup
// @Produce json
// @Success 204
// @Param event_id path int true "Event Id"
// @Router /events/{event_id}/signups/self [delete]
func (e *SignupController) deleteSignupHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		eventId, err := strconv.Atoi(c.Param("event_id"))
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		user, err := e.userService.GetUserFromAuthCookie(c)
		if err != nil {
			c.JSON(401, gin.H{"error": "Not authenticated"})
			return
		}
		err = e.signupService.RemoveSignup(user.Id, eventId)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{})
	}
}

// @id GetEventSignups
// @Description Fetches all signups for the event
// @Tags signup
// @Produce json
// @Success 200 {object} map[int][]Signup
// @Param event_id path int true "Event Id"
// @Router /events/{event_id}/signups [get]
func (e *SignupController) getEventSignupsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		eventId, err := strconv.Atoi(c.Param("event_id"))
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		signups, err := e.signupService.GetSignupsForEvent(eventId)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		signupsResponse := make(map[int][]*Signup, 0)
		for teamId, teamSignups := range signups {
			signupsResponse[teamId] = utils.Map(teamSignups, toSignupResponse)
		}
		c.JSON(200, signupsResponse)
	}
}

type Signup struct {
	Id               int                         `json:"id" binding:"required"`
	User             *NonSensitiveUser           `json:"user" binding:"required"`
	Timestamp        time.Time                   `json:"timestamp" binding:"required"`
	ExpectedPlaytime repository.ExpectedPlayTime `json:"expected_playtime" binding:"required"`
	TeamId           *int                        `json:"team_id"`
}

type SignupCreate struct {
	ExpectedPlaytime repository.ExpectedPlayTime `json:"expected_playtime" binding:"required"`
}

func toSignupResponse(signup *repository.Signup) *Signup {
	if signup == nil {
		return nil
	}

	return &Signup{
		Id:               signup.Id,
		User:             toNonSensitiveUserResponse(signup.User),
		Timestamp:        signup.Timestamp,
		ExpectedPlaytime: signup.ExpectedPlayTime,
	}
}
