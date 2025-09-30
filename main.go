package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"

	h "github.com/ChrisTheAbysswalker/meownder-backend/handlers"
	s "github.com/ChrisTheAbysswalker/meownder-backend/services"
)

func main() {
	gin.SetMode(gin.ReleaseMode)

	router := gin.Default()

	router.Use(corsMiddleware())

	catService := s.NewCatService()

	catHandler := h.NewCatHandler(catService)

	api := router.Group("/api")
	{
		api.GET("/cats", catHandler.GetCats)
		api.GET("/health", catHandler.Health)
		api.GET("/profiles", catHandler.GetCatProfiles)
		api.GET("/profiles/:id", catHandler.GetCatProfileByID)
		api.POST("/profiles/refresh", catHandler.RefreshImages)
	}

	router.GET("/", func(c *gin.Context) {
		c.File("./public/index.html")
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" 
	}

	baseURL := os.Getenv("RENDER_EXTERNAL_URL")
	if baseURL == "" {
		baseURL = fmt.Sprintf("http://localhost:%s", port)
	}

	fmt.Printf("🚀 Meownder API corriendo en %s\n", baseURL)
	fmt.Printf("📡 Endpoints disponibles:\n")
	fmt.Printf("   • GET  %s/api/profiles         - Obtener todos los perfiles de gatos\n", baseURL)
	fmt.Printf("   • GET  %s/api/profiles/:id     - Obtener perfil por ID\n", baseURL)
	fmt.Printf("   • POST %s/api/profiles/refresh - Refrescar imágenes\n", baseURL)
	fmt.Printf("   • GET  %s/api/cats?count=5     - Obtener imágenes de gatos (legacy)\n", baseURL)
	fmt.Printf("   • GET  %s/api/health           - Health check\n", baseURL)
	fmt.Printf("   • GET  %s/                 - Información de la API\n", baseURL)

	if err := router.Run(":" + port); err != nil {
		log.Fatal("Error al iniciar el servidor:", err)
	}
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
