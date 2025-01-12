# SpotyToYTM

El propósito de SpotyToYTM (Spotify To YouTube Music) es el de migrar listas de
reproducción de Spotify a YouTube Music.

## Estructura del Proyecto

El proyecto tiene la siguiente estructura de archivos:

```text
spotytotym
├── cmd
│   └── main.go             # Punto de entrada de la aplicación
├── internal
│   ├── spotify
│   │   └── spotify.go      # Interacción con la API de Spotify
│   ├── youtubemusic
│   │   └── youtubemusic.go # Interacción con la API de YouTube Music
│   └── migrator
│       └── migrator.go     # Manejo del proceso de migración
├── go.mod                  # Definición del módulo y dependencias
└── README.md               # Documentación del proyecto
```

## Uso

1. Clona el repositorio.
2. Instala las dependencias usando `go mod tidy`.
3. Ejecuta la aplicación con `go run cmd/main.go`.

A medida que se desarrolle la herramienta, se agregarán más detalles y ejemplos
de uso.