package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	m "github.com/ChrisTheAbysswalker/meownder-backend/models"
	s "github.com/ChrisTheAbysswalker/meownder-backend/services"
)

type CatHandler struct {
	service *s.CatService
}

func NewCatHandler(service *s.CatService) *CatHandler {
	return &CatHandler{
		service: service,
	}
}

func (h *CatHandler) GetCatProfiles(c *gin.Context) {
	profiles := h.service.GetCatProfiles()

	if len(profiles) == 0 {
		c.JSON(http.StatusNotFound, m.ErrorResponse{
			Error:   "no_profiles_found",
			Message: "No se encontraron perfiles de gatos",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"cats":  profiles,
		"count": len(profiles),
	})
}

func (h *CatHandler) GetCatProfileByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, m.ErrorResponse{
			Error:   "invalid_id",
			Message: "El ID debe ser un número válido",
		})
		return
	}

	profile, err := h.service.GetCatProfileByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, m.ErrorResponse{
			Error:   "profile_not_found",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, profile)
}

func (h *CatHandler) RefreshImages(c *gin.Context) {
	if err := h.service.RefreshCatImages(); err != nil {
		c.JSON(http.StatusInternalServerError, m.ErrorResponse{
			Error:   "refresh_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Imágenes actualizadas correctamente",
	})
}

func (h *CatHandler) GetCats(c *gin.Context) {
	// * Obtener parámetro count (default: 5, max: 10)
	count := 5
	if countStr := c.Query("count"); countStr != "" {
		if parsed, err := strconv.Atoi(countStr); err == nil && parsed > 0 && parsed <= 10 {
			count = parsed
		}
	}

	urls, batch, err := h.service.GenerateCatURLs(count)
	if err != nil {
		c.JSON(http.StatusInternalServerError, m.ErrorResponse{
			Error:   "no_images_available",
			Message: err.Error(),
		})
		return
	}

	response := m.CatResponse{
		URLs:  urls,
		Count: len(urls),
		Batch: batch,
	}

	c.JSON(http.StatusOK, response)
}

func (h *CatHandler) Health(c *gin.Context) {
	profiles := h.service.GetCatProfiles()
	
	response := m.HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().Unix(),
		Batches:   h.service.GetBatchCount(),
	}

	if len(profiles) > 0 {
		c.JSON(http.StatusOK, gin.H{
			"status":    response.Status,
			"timestamp": response.Timestamp,
			"batches":   response.Batches,
			"profiles_loaded": len(profiles),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}