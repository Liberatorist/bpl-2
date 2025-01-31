package controller

import (
	"bpl/auth"
	"bpl/repository"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type RouteInfo struct {
	Method        string
	Path          string
	HandlerFunc   gin.HandlerFunc
	Authenticated bool
	RequiredRoles []repository.Permission
}

func SetRoutes(r *gin.Engine, db *gorm.DB) {
	routes := make([]RouteInfo, 0)
	group := r.Group("/api")
	routes = append(routes, setupEventController(db)...)
	routes = append(routes, setupTeamController(db)...)
	routes = append(routes, setupConditionController(db)...)
	routes = append(routes, setupScoringCategoryController(db)...)
	routes = append(routes, setupObjectiveController(db)...)
	routes = append(routes, setupOauthController(db)...)
	routes = append(routes, setupUserController(db)...)
	routes = append(routes, setupScoringPresetController(db)...)
	routes = append(routes, setupSignupController(db)...)
	routes = append(routes, setupSubmissionController(db)...)
	routes = append(routes, setupScoreController(db)...)
	routes = append(routes, setupStreamController(db)...)
	for _, route := range routes {
		handlerfuncs := make([]gin.HandlerFunc, 0)
		if route.Authenticated {
			handlerfuncs = append(handlerfuncs, AuthMiddleware(route.RequiredRoles))
		}
		handlerfuncs = append(handlerfuncs, route.HandlerFunc)
		group.Handle(route.Method, route.Path, handlerfuncs...)
	}
}

func AuthMiddleware(roles []repository.Permission) gin.HandlerFunc {
	return func(r *gin.Context) {
		authCookie, err := r.Cookie("auth")
		if err != nil {
			r.AbortWithStatus(401)
			return
		}
		token, err := auth.ParseToken(authCookie)
		if err != nil {
			r.AbortWithStatus(401)
			return
		}
		claims := &auth.Claims{}
		if !token.Valid {
			r.AbortWithStatus(401)
			return
		}
		claims.FromJWTClaims(token.Claims)
		if err := claims.Valid(); err != nil {
			r.AbortWithStatus(401)
			return
		}
		if len(roles) == 0 {
			r.Next()
			return
		}

		for _, requiredRole := range roles {
			for _, userRole := range claims.Permissions {
				if requiredRole == repository.Permission(userRole) {
					r.Next()
					return
				}
			}
		}
		r.AbortWithStatus(403)
	}
}
