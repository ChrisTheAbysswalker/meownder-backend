 # Meownder Backend

  !Go (https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white)
  !Gin (https://img.shields.io/badge/Gin-0082C9?style=for-the-badge&logo=gin&logoColor=white)

  Backend de la API de Meownder, un servicio para obtener perfiles e imágenes de gatos adorables. Este proyecto está construido
  en Go utilizando el framework Gin.

  ## Diagrama de Flujo

  El siguiente diagrama ilustra el flujo básico para obtener un perfil de gato.

  `mermaid
  sequenceDiagram
      participant Client as Cliente
      participant API as Meownder API
      participant CatJSON as cats.json
      participant Cataas as cataas.com


      Client->>API: GET /api/profiles/1
      API->>CatJSON: Cargar datos del perfil (ID: 1)
      CatJSON-->>API: Retorna datos (Nombre, Raza, etc.)
      API->>Cataas: Generar y obtener URL de imagen
      Cataas-->>API: Retorna URL de la imagen
      API-->>Client: Retorna perfil completo (JSON)

  `

  ## API Endpoints

  A continuación se detallan los endpoints disponibles en la API.

  ---

  ### 1. Obtener Todos los Perfiles

  Obtiene una lista de todos los perfiles de gatos disponibles.

  - Método: GET
  - Endpoint: /api/profiles

  Respuesta Exitosa (200 OK):
  `json
  {
      "cats": [
          {
              "id": 1,
              "name": "Whiskers",
              "breed": "Siamese",
              "age": 2,
              "description": "A curious and playful cat.",
              "img": "https://cataas.com/cat?timestamp=1678886400&rand=123"
          },
          {
              "id": 2,
              "name": "Shadow",
              "breed": "Bombay",
              "age": 4,
              "description": "Loves to nap in the sun.",
              "img": "https://cataas.com/cat?timestamp=1678886401&rand=456"
          }
      ],
      "count": 2
  }
  `

  Respuesta de Error (404 Not Found):
  `json
  {
      "error": "no_profiles_found",
      "message": "No se encontraron perfiles de gatos"
  }

  `

  ---

  ### 2. Obtener Perfil por ID

  Obtiene un perfil de gato específico según su id.

  - Método: GET
  - Endpoint: /api/profiles/:id

  Respuesta Exitosa (200 OK):
  `json
  {
      "id": 1,
      "name": "Whiskers",
      "breed": "Siamese",
      "age": 2,
      "description": "A curious and playful cat.",
      "img": "https://cataas.com/cat?timestamp=1678886400&rand=123"
  }
  `

  Respuesta de Error (404 Not Found):
  `json
  {
      "error": "profile_not_found",
      "message": "gato con ID 99 no encontrado"
  }

  `

  ---

  ### 3. Refrescar Imágenes

  Actualiza las imágenes de todos los perfiles de gatos, obteniendo nuevas URLs desde cataas.com.

  - Método: POST
  - Endpoint: /api/profiles/refresh

  Respuesta Exitosa (200 OK):
  `json
  {
      "message": "Imágenes actualizadas correctamente"
  }
  `

  Respuesta de Error (500 Internal Server Error):
  `json
  {
      "error": "refresh_failed",
      "message": "error al contactar el servicio de imágenes"
  }

  `

  ---

  ### 4. Health Check

  Verifica el estado del servicio.

  - Método: GET
  - Endpoint: /api/health

  Respuesta Exitosa (200 OK):
  `json
  {
      "status": "healthy",
      "timestamp": 1678886400,
      "batches": 15,
      "profiles_loaded": 10
  }

  `

  ---

  ### 5. Obtener Imágenes de Gatos (Legacy)

  Endpoint legacy para obtener un lote de URLs de imágenes de gatos.

  - Método: GET
  - Endpoint: /api/cats
  - Query Params: count (número de imágenes, 1-10, default: 5)

  Respuesta Exitosa (200 OK):
  `json
  {
      "urls": [
          "https://cataas.com/cat?timestamp=1678886402&rand=789",
          "https://cataas.com/cat?timestamp=1678886403&rand=101"
      ],
      "count": 2,
      "batch": 16
  }

  `

  ## Configuración

  ### Variables de entorno

  El backend utiliza las siguientes variables de entorno:

  - PORT: Puerto en el que se ejecutará el servidor. Por defecto es 8080.
  - RENDER_EXTERNAL_URL: URL base para el entorno de producción. Por defecto es http://localhost:8080.

  ## Ejecución

  Para ejecutar el backend, sigue los siguientes pasos:

  1.  Instalar Go en tu sistema.
  2.  Clonar este repositorio.
  3.  Ejecutar go run main.go en la raíz del proyecto.

  ## Despliegue

  Este proyecto está configurado para un despliegue sencillo en Render a través del archivo render.yaml.