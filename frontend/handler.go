package frontend

import (
	"bytes"
	"embed"
	"io/fs"
	"net/http"
	"strconv"
	"strings"

	"donetick.com/core/config"
	"github.com/gin-gonic/gin"
)

//go:embed dist
var embeddedFiles embed.FS

type Handler struct {
	ServeFrontend bool
	// indexHTML is nil unless BasePath is set, in which case it's the
	// embedded index.html with window.__BASE_PATH__ injected -- computed
	// once here rather than per-request.
	indexHTML []byte
}

func NewHandler(config *config.Config) *Handler {
	h := &Handler{
		ServeFrontend: config.Server.ServeFrontend,
	}
	if config.BasePath != "" {
		if raw, err := embeddedFiles.ReadFile("dist/index.html"); err == nil {
			injected := `<head><script>window.__BASE_PATH__=` + strconv.Quote(config.BasePath) + `;</script>`
			h.indexHTML = bytes.Replace(raw, []byte("<head>"), []byte(injected), 1)
		}
	}
	return h
}

func Routes(router *gin.Engine, h *Handler) {
	// this whole logic is walkaround for serving frontend files
	// TODO: figure out better way to improve it. main issue i run into is failing over to index.html when file does not exist

	if h.ServeFrontend {
		// if file exists in dist folder, serve it
		router.Use(staticMiddleware("dist", h.indexHTML))
		// if file does not exist in dist folder fallback to index.html
		router.NoRoute(staticMiddlewareNoRoute("dist", h.indexHTML))

	}

}

func staticMiddleware(root string, indexHTML []byte) gin.HandlerFunc {
	fileServer := http.FileServer(getFileSystem(root))

	return func(c *gin.Context) {
		_, err := fs.Stat(embeddedFiles, "dist"+c.Request.URL.Path)
		if err != nil {
			c.Next()
			return
		}
		// Cache-busted assets (filenames contain content hashes) can be
		// cached indefinitely. Other files like index.html must not be
		// cached so that updates are picked up immediately.
		if strings.HasPrefix(c.Request.URL.Path, "/assets/") {
			c.Header("Cache-Control", "public, max-age=31536000, immutable")
		}
		if indexHTML != nil && (c.Request.URL.Path == "/" || c.Request.URL.Path == "/index.html") {
			c.Data(http.StatusOK, "text/html; charset=utf-8", indexHTML)
			return
		}
		fileServer.ServeHTTP(c.Writer, c.Request)

	}
}
func staticMiddlewareNoRoute(root string, indexHTML []byte) gin.HandlerFunc {
	fileServer := http.FileServer(getFileSystem(root))

	// always serve index.html for any route does not match:
	return func(c *gin.Context) {
		if indexHTML != nil {
			c.Data(http.StatusOK, "text/html; charset=utf-8", indexHTML)
			return
		}
		// Rewrite all requests to serve index.html
		c.Request.URL.Path = "/"
		fileServer.ServeHTTP(c.Writer, c.Request)

	}
}

func getFileSystem(path string) http.FileSystem {
	fs, err := fs.Sub(embeddedFiles, path)
	if err != nil {
		panic(err)
	}
	return http.FS(fs)
}
