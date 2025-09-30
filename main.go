package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// Estructura de respuesta JSON
type CatResponse struct {
	URLs  []string `json:"urls"`
	Count int      `json:"count"`
	Batch int      `json:"batch"`
}

// Estructura de respuesta de error
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// Estructura para validar URLs
type CatURL struct {
	URL       string `json:"url"`
	ID        string `json:"id"`
	Timestamp int64  `json:"timestamp"`
}

// Cache simple para evitar URLs duplicadas recientes
var (
	recentURLs = make(map[string]bool)
	cacheMutex sync.RWMutex
	batchCount = 0
)

func generateCatURL() CatURL {
	timestamp := time.Now().UnixNano()
	baseURL := "https://cataas.com/cat"
	url := fmt.Sprintf("%s?timestamp=%d", baseURL, timestamp)
	id := fmt.Sprintf("cat-%d", timestamp)
	
	return CatURL{
		URL:       url,
		ID:        id,
		Timestamp: timestamp,
	}
}

// Validar que la imagen sea accesible
func validateCatURL(catURL CatURL, timeout time.Duration) bool {
	client := &http.Client{
		Timeout: timeout,
	}
	
	resp, err := client.Head(catURL.URL)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	
	return resp.StatusCode == http.StatusOK
}

// Limpiar cache de URLs antiguas (mantener solo las últimas 50)
func cleanCache() {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	
	if len(recentURLs) > 50 {
		// En producción, usarías una estructura más eficiente como LRU cache
		recentURLs = make(map[string]bool)
	}
}

func enableCORS(w http.ResponseWriter, r *http.Request) {
	// Permitir CORS para desarrollo
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
}

func catHandler(w http.ResponseWriter, r *http.Request) {
	// Habilitar CORS
	enableCORS(w, r)
	if r.Method == "OPTIONS" {
		return
	}
	
	// Solo permitir GET
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "method_not_allowed",
			Message: "Only GET requests are allowed",
		})
		return
	}
	
	// Obtener número de imágenes (default: 5, max: 10)
	n := 5
	if countParam := r.URL.Query().Get("count"); countParam != "" {
		if parsed, err := strconv.Atoi(countParam); err == nil && parsed > 0 && parsed <= 10 {
			n = parsed
		}
	}
	
	// Incrementar contador de batch
	batchCount++
	currentBatch := batchCount
	
	log.Printf("🐱 Generando lote %d con %d imágenes", currentBatch, n)
	
	// Generar URLs concurrentemente
	cats := make([]CatURL, n)
	urls := make([]string, 0, n)
	var wg sync.WaitGroup
	var urlMutex sync.Mutex
	
	// Timeout para validación de URLs
	//timeout := 3 * time.Second
	
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			
			maxRetries := 3
			for retry := 0; retry < maxRetries; retry++ {
				catURL := generateCatURL()
				
				// Verificar si la URL es nueva (evitar duplicados recientes)
				cacheMutex.RLock()
				isDuplicate := recentURLs[catURL.URL]
				cacheMutex.RUnlock()
				
				if !isDuplicate {
					// Validar que la URL sea accesible (opcional, puede ser lento)
					// if validateCatURL(catURL, timeout) {
					
					// Agregar a cache
					cacheMutex.Lock()
					recentURLs[catURL.URL] = true
					cacheMutex.Unlock()
					
					urlMutex.Lock()
					cats[index] = catURL
					urls = append(urls, catURL.URL)
					urlMutex.Unlock()
					break
					
					// } else {
					// 	log.Printf("⚠️ URL no accesible: %s (intento %d)", catURL.URL, retry+1)
					// }
				} else {
					log.Printf("🔄 URL duplicada detectada, generando nueva...")
				}
				
				// Pequeño delay entre reintentos
				time.Sleep(100 * time.Millisecond)
			}
		}(i)
	}
	
	// Esperar a que todas las goroutines terminen
	wg.Wait()
	
	// Limpiar cache periódicamente
	if currentBatch%10 == 0 {
		go cleanCache()
	}
	
	// Verificar que obtuvimos suficientes URLs
	if len(urls) == 0 {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "no_images_available",
			Message: "No se pudieron obtener imágenes de gatos",
		})
		return
	}
	
	// Respuesta exitosa
	response := CatResponse{
		URLs:  urls,
		Count: len(urls),
		Batch: currentBatch,
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
	
	log.Printf("✅ Lote %d completado: %d imágenes enviadas", currentBatch, len(urls))
}

// Health check endpoint
func healthHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w, r)
	if r.Method == "OPTIONS" {
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"batches":   batchCount,
	})
}

// Middleware de logging
func loggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("📊 %s %s - %v", r.Method, r.URL.Path, time.Since(start))
	}
}

func main() {
	// Rutas con middleware de logging
	http.HandleFunc("/cats", loggingMiddleware(catHandler))
	http.HandleFunc("/health", loggingMiddleware(healthHandler))
	
	// Ruta de información
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w, r)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"service":     "Cat Tinder API",
			"version":     "1.0.0",
			"endpoints": map[string]string{
				"/cats":        "GET - Obtener lote de imágenes de gatos",
				"/health":      "GET - Health check del servicio",
			},
			"params": map[string]string{
				"count": "Número de imágenes (1-10, default: 5)",
			},
			"example": "http://localhost:8080/cats?count=5",
		})
	})
	
	port := ":8080"
	fmt.Printf("🚀 Cat Tinder API corriendo en http://localhost%s\n", port)
	fmt.Printf("📡 Endpoints disponibles:\n")
	fmt.Printf("   • GET  /cats?count=5  - Obtener imágenes de gatos\n")
	fmt.Printf("   • GET  /health        - Health check\n")
	fmt.Printf("   • GET  /              - Información de la API\n")
	
	log.Fatal(http.ListenAndServe(port, nil))
}