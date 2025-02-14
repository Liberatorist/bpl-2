package controller

import (
	"bpl/repository"
	"bpl/service"
	"bpl/utils"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ScoringCategoryController struct {
	service *service.ScoringCategoryService
}

func NewScoringCategoryController() *ScoringCategoryController {
	return &ScoringCategoryController{service: service.NewScoringCategoryService()}
}

func setupScoringCategoryController() []RouteInfo {
	e := NewScoringCategoryController()
	routes := []RouteInfo{
		{Method: "GET", Path: "/events/:event_id/rules", HandlerFunc: e.getRulesForEventHandler()},
		{Method: "PUT", Path: "/scoring/categories", HandlerFunc: e.createCategoryHandler(), Authenticated: true, RequiredRoles: []repository.Permission{repository.PermissionAdmin}},
		{Method: "GET", Path: "/scoring/categories/:id", HandlerFunc: e.getScoringCategoryHandler(), Authenticated: true, RequiredRoles: []repository.Permission{repository.PermissionAdmin}},
		{Method: "DELETE", Path: "/scoring/categories/:id", HandlerFunc: e.deleteCategoryHandler(), Authenticated: true, RequiredRoles: []repository.Permission{repository.PermissionAdmin}}}
	return routes
}

// @id GetRulesForEvent
// @Description Fetches the rules for the current event
// @Tags scoring
// @Produce json
// @Success 200 {array} CategoryResponse
// @Router /scoring/categories [get]
func (e *ScoringCategoryController) getRulesForEventHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		event_id, err := strconv.Atoi(c.Param("event_id"))
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		rules, err := e.service.GetRulesForEvent(event_id, "Objectives", "Objectives.Conditions")
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, toPublicCategoryResponse(rules))
	}
}

// @id GetScoringCategory
// @Description Fetches a scoring category by id
// @Tags scoring
// @Produce json
// @Param id path int true "Category ID"
// @Success 200 {object} CategoryResponse
// @Router /scoring/categories/{id} [get]
func (e *ScoringCategoryController) getScoringCategoryHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		category, err := e.service.GetCategoryById(id, "Objectives", "Objectives.Conditions")
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(404, gin.H{"error": "Category not found"})
			} else {
				c.JSON(500, gin.H{"error": err.Error()})
			}
			return
		}
		c.JSON(200, toCategoryResponse(category))
	}
}

// @id CreateCategory
// @Description Creates a new scoring category
// @Tags scoring
// @Accept json
// @Produce json
// @Param body body CategoryCreate true "Category to create"
// @Success 201 {object} CategoryResponse
func (e *ScoringCategoryController) createCategoryHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var categoryCreate CategoryCreate
		if err := c.BindJSON(&categoryCreate); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		category, err := e.service.CreateCategory(categoryCreate.toModel())
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(404, gin.H{"error": "Parent category not found"})
			} else {
				c.JSON(500, gin.H{"error": err.Error()})
			}
			return
		}
		c.JSON(201, toCategoryResponse(category))
	}
}

// @id DeleteCategory
// @Description Deletes a scoring category
// @Tags scoring
// @Produce json
// @Param id path int true "Category ID"
// @Success 204
// @Router /scoring/categories/{id} [delete]
func (e *ScoringCategoryController) deleteCategoryHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		err = e.service.DeleteCategory(id)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(404, gin.H{"error": "Category not found"})
			} else {
				c.JSON(500, gin.H{"error": err.Error()})
			}
			return
		}
		c.JSON(204, nil)
	}
}

type CategoryCreate struct {
	ID        *int   `json:"id"`
	ParentID  int    `json:"parent_id" binding:"required"`
	Name      string `json:"name" binding:"required"`
	ScoringId *int   `json:"scoring_preset_id"`
}

type CategoryResponse struct {
	ID              int                  `json:"id" binding:"required"`
	Name            string               `json:"name" binding:"required"`
	SubCategories   []*CategoryResponse  `json:"sub_categories" binding:"required"`
	Objectives      []*ObjectiveResponse `json:"objectives" binding:"required"`
	ScoringPresetID *int                 `json:"scoring_preset_id"`
}

func (e *CategoryCreate) toModel() *repository.ScoringCategory {
	category := &repository.ScoringCategory{
		ParentID:  &e.ParentID,
		Name:      e.Name,
		ScoringId: e.ScoringId,
	}
	if e.ID != nil {
		category.ID = *e.ID
	}
	return category
}

type ScoringMethodResponse struct {
	Type   repository.ScoringMethod `json:"type" binding:"required"`
	Points []int                    `json:"points" binding:"required"`
}

func toCategoryResponse(category *repository.ScoringCategory) *CategoryResponse {
	if category == nil {
		return nil
	}
	return &CategoryResponse{
		ID:              category.ID,
		Name:            category.Name,
		SubCategories:   utils.Map(category.SubCategories, toCategoryResponse),
		Objectives:      utils.Map(category.Objectives, toObjectiveResponse),
		ScoringPresetID: category.ScoringId,
	}
}
func toPublicCategoryResponse(category *repository.ScoringCategory) *CategoryResponse {
	if category == nil {
		return nil
	}
	return &CategoryResponse{
		ID:              category.ID,
		Name:            category.Name,
		SubCategories:   utils.Map(category.SubCategories, toPublicCategoryResponse),
		Objectives:      utils.Map(category.Objectives, toPublicObjectiveResponse),
		ScoringPresetID: category.ScoringId,
	}
}
