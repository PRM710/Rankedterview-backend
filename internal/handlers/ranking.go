package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/PRM710/Rankedterview-backend/internal/services"
	"github.com/PRM710/Rankedterview-backend/internal/utils"
)

type RankingHandler struct {
	rankingService *services.RankingService
}

func NewRankingHandler(rankingService *services.RankingService) *RankingHandler {
	return &RankingHandler{
		rankingService: rankingService,
	}
}

// GetGlobalLeaderboard retrieves the global leaderboard
func (h *RankingHandler) GetGlobalLeaderboard(c *gin.Context) {
	limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "100"), 10, 64)

	rankings, err := h.rankingService.GetGlobalLeaderboard(c.Request.Context(), limit)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to retrieve leaderboard")
		return
	}

	// Convert to response format
	responses := make([]interface{}, len(rankings))
	for i, ranking := range rankings {
		responses[i] = ranking.ToResponse()
	}

	utils.SuccessResponse(c, responses)
}

// GetCategoryLeaderboard retrieves a category-specific leaderboard
func (h *RankingHandler) GetCategoryLeaderboard(c *gin.Context) {
	category := c.Param("category")
	limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "100"), 10, 64)

	// Validate category
	validCategories := map[string]bool{
		"overall":       true,
		"communication": true,
		"technical":     true,
		"confidence":    true,
		"structure":     true,
	}

	if !validCategories[category] {
		utils.BadRequestResponse(c, "Invalid category")
		return
	}

	rankings, err := h.rankingService.GetCategoryLeaderboard(c.Request.Context(), category, limit)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to retrieve leaderboard")
		return
	}

	// Convert to response format
	responses := make([]interface{}, len(rankings))
	for i, ranking := range rankings {
		responses[i] = ranking.ToResponse()
	}

	utils.SuccessResponse(c, responses)
}

// GetUserRank retrieves a user's current rank
func (h *RankingHandler) GetUserRank(c *gin.Context) {
	userID := c.Param("userId")
	category := c.DefaultQuery("category", "overall")

	rank, err := h.rankingService.GetUserRank(c.Request.Context(), userID, category)
	if err != nil {
		utils.NotFoundResponse(c, "Rank not found for user")
		return
	}

	utils.SuccessResponse(c, gin.H{
		"userId":   userID,
		"category": category,
		"rank":     rank,
	})
}

// GetRankHistory retrieves a user's ranking history
func (h *RankingHandler) GetRankHistory(c *gin.Context) {
	userID := c.Param("userId")

	ranking, err := h.rankingService.GetRankHistory(c.Request.Context(), userID)
	if err != nil {
		utils.NotFoundResponse(c, "Ranking history not found")
		return
	}

	utils.SuccessResponse(c, gin.H{
		"currentRank": ranking.Rank,
		"currentElo":  ranking.Elo,
		"history":     ranking.History,
	})
}
