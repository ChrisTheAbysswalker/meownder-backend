package services

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"
	"sync"
	"time"

	m "github.com/ChrisTheAbysswalker/meownder-backend/models"
)

type CatService struct {
	recentURLs map[string]bool
	cacheMutex sync.RWMutex
	batchCount int
	countMutex sync.Mutex
	catProfiles []m.CatProfile 
	profilesMutex sync.RWMutex
}

func NewCatService() *CatService {
	service := &CatService{
		recentURLs: make(map[string]bool),
		batchCount: 0,
	}
	
	// * Cargar perfiles de gatos al iniciar
	if err := service.loadCatProfiles(); err != nil {
		log.Printf("⚠️ Error cargando perfiles de gatos: %v", err)
	} else {
		log.Printf("✅ Perfiles de gatos cargados: %d", len(service.catProfiles))
	}
	
	return service
}

func (s *CatService) loadCatProfiles() error {
	data, err := os.ReadFile("cats.json")
	if err != nil {
		return fmt.Errorf("error leyendo cats.json: %w", err)
	}

	var catsData struct {
		Cats []m.CatProfile `json:"cats"`
	}

	if err := json.Unmarshal(data, &catsData); err != nil {
		return fmt.Errorf("error parseando JSON: %w", err)
	}

	// * Llenar imágenes desde Cat as a Service
	for i := range catsData.Cats {
		catURL := s.generateCatURL()
		catsData.Cats[i].Img = catURL.URL
		log.Printf("🖼️ Imagen asignada a %s: %s", catsData.Cats[i].Name, catURL.URL)
	}

	s.profilesMutex.Lock()
	s.catProfiles = catsData.Cats
	s.profilesMutex.Unlock()

	return nil
}

func (s *CatService) GetCatProfiles() []m.CatProfile {
	s.profilesMutex.RLock()
	defer s.profilesMutex.RUnlock()
	return s.catProfiles
}

func (s *CatService) GetCatProfileByID(id int) (*m.CatProfile, error) {
	s.profilesMutex.RLock()
	defer s.profilesMutex.RUnlock()

	for _, cat := range s.catProfiles {
		if cat.ID == id {
			return &cat, nil
		}
	}

	return nil, fmt.Errorf("gato con ID %d no encontrado", id)
}

func (s *CatService) RefreshCatImages() error {
	s.profilesMutex.Lock()
	defer s.profilesMutex.Unlock()

	for i := range s.catProfiles {
		catURL := s.generateCatURL()
		s.catProfiles[i].Img = catURL.URL
	}

	log.Println("🔄 Imágenes de perfiles actualizadas")
	return nil
}

func (s *CatService) GenerateCatURLs(count int) ([]string, int, error) {
	s.countMutex.Lock()
	s.batchCount++
	currentBatch := s.batchCount
	s.countMutex.Unlock()

	log.Printf("🐱 Generando lote %d con %d imágenes", currentBatch, count)

	urls := make([]string, 0, count)
	var wg sync.WaitGroup
	var urlMutex sync.Mutex

	for i := 0; i < count; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			maxRetries := 3
			for retry := 0; retry < maxRetries; retry++ {
				catURL := s.generateCatURL()

				s.cacheMutex.RLock()
				isDuplicate := s.recentURLs[catURL.URL]
				s.cacheMutex.RUnlock()

				if !isDuplicate {
					s.cacheMutex.Lock()
					s.recentURLs[catURL.URL] = true
					s.cacheMutex.Unlock()

					urlMutex.Lock()
					urls = append(urls, catURL.URL)
					urlMutex.Unlock()
					break
				} else {
					log.Printf("🔄 URL duplicada detectada, generando nueva...")
				}

				time.Sleep(100 * time.Millisecond)
			}
		}(i)
	}

	wg.Wait()

	if currentBatch%10 == 0 {
		go s.cleanCache()
	}

	if len(urls) == 0 {
		return nil, 0, fmt.Errorf("no se pudieron obtener imágenes de gatos")
	}

	log.Printf("✅ Lote %d completado: %d imágenes enviadas", currentBatch, len(urls))
	return urls, currentBatch, nil
}

func (s *CatService) generateCatURL() m.CatURL {
	timestamp := time.Now().UnixNano()
	baseURL := "https://cataas.com/cat"

	randNum, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		log.Printf("Error generando número aleatorio: %v", err)
		randNum = big.NewInt(0)
	}

	url := fmt.Sprintf("%s?timestamp=%d&rand=%s", baseURL, timestamp, randNum.String())
	id := fmt.Sprintf("cat-%d-%s", timestamp, randNum.String())

	return m.CatURL{
		URL:       url,
		ID:        id,
		Timestamp: timestamp,
	}
}

// ! valida que la imagen sea accesible (no implementado por ahora)
func (s *CatService) validateCatURL(catURL m.CatURL, timeout time.Duration) bool {
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
func (s *CatService) cleanCache() {
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()

	if len(s.recentURLs) > 50 {
		s.recentURLs = make(map[string]bool)
		log.Println("🧹 Cache limpiado")
	}
}

func (s *CatService) GetBatchCount() int {
	s.countMutex.Lock()
	defer s.countMutex.Unlock()
	return s.batchCount
}