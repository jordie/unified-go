package codegen

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jgirmay/GAIA_GO/internal/api"
	"github.com/jgirmay/GAIA_GO/internal/docs"
)

// CodegenHandler manages SDK generation endpoints
type CodegenHandler struct {
	generator          *ClientGenerator
	distributionHandler *DistributionHandler
}

// NewCodegenHandler creates a new codegen handler
func NewCodegenHandler(spec *docs.OpenAPISpec) *CodegenHandler {
	generator := NewClientGenerator(spec)
	config := PackageConfig{
		Name:        "client",
		Version:     "1.0.0",
		Description: "GAIA API Client SDK",
		Author:      "GAIA Development Team",
		License:     "MIT",
		Homepage:    "https://github.com/jgirmay/gaia-go-client",
		Repository:  "https://github.com/jgirmay/gaia-go-client.git",
		BugTracker:  "https://github.com/jgirmay/gaia-go-client/issues",
		Tags: []string{
			"api",
			"client",
			"sdk",
			"gaia",
			"education",
		},
	}

	return &CodegenHandler{
		generator:          generator,
		distributionHandler: NewDistributionHandler(generator, config),
	}
}

// RegisterRoutes registers SDK generation endpoints
func (h *CodegenHandler) RegisterRoutes(engine *gin.Engine) {
	sdks := engine.Group("/api/sdk")
	{
		sdks.GET("", h.handleSDKIndex)
		sdks.GET("/typescript", h.handleGenerateTypeScript)
		sdks.GET("/go", h.handleGenerateGo)
		sdks.GET("/python", h.handleGeneratePython)
		sdks.GET("/endpoints", h.handleListEndpoints)
	}

	// Register distribution routes
	h.distributionHandler.RegisterRoutes(engine)
}

// handleSDKIndex returns SDK generation index
func (h *CodegenHandler) handleSDKIndex(c *gin.Context) {
	languages := []gin.H{
		{
			"language":    "typescript",
			"name":        "TypeScript/JavaScript",
			"description": "Complete TypeScript client with full type safety",
			"url":         "/api/sdk/typescript",
			"filename":    "gaia-client.ts",
		},
		{
			"language":    "go",
			"name":        "Go",
			"description": "Native Go client library with proper error handling",
			"url":         "/api/sdk/go",
			"filename":    "client.go",
		},
		{
			"language":    "python",
			"name":        "Python",
			"description": "Python client package with async support",
			"url":         "/api/sdk/python",
			"filename":    "client.py",
		},
	}

	api.RespondWith(c, http.StatusOK, gin.H{
		"message":         "GAIA SDK Code Generator",
		"description":     "Generate native client SDKs for various programming languages",
		"supported_languages": languages,
		"documentation":   "/api/docs",
		"endpoints":       "/api/sdk/endpoints",
	})
}

// handleGenerateTypeScript generates TypeScript SDK
func (h *CodegenHandler) handleGenerateTypeScript(c *gin.Context) {
	sdk, err := h.generator.GenerateTypeScriptSDK()
	if err != nil {
		api.RespondWithError(c, api.NewError(
			api.ErrCodeInternalServer,
			"Failed to generate TypeScript SDK: "+err.Error(),
			http.StatusInternalServerError,
		))
		return
	}

	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=gaia-client.ts")
	c.String(http.StatusOK, sdk)
}

// handleGenerateGo generates Go SDK
func (h *CodegenHandler) handleGenerateGo(c *gin.Context) {
	sdk, err := h.generator.GenerateGoSDK()
	if err != nil {
		api.RespondWithError(c, api.NewError(
			api.ErrCodeInternalServer,
			"Failed to generate Go SDK: "+err.Error(),
			http.StatusInternalServerError,
		))
		return
	}

	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=client.go")
	c.String(http.StatusOK, sdk)
}

// handleGeneratePython generates Python SDK
func (h *CodegenHandler) handleGeneratePython(c *gin.Context) {
	sdk, err := h.generator.GeneratePythonSDK()
	if err != nil {
		api.RespondWithError(c, api.NewError(
			api.ErrCodeInternalServer,
			"Failed to generate Python SDK: "+err.Error(),
			http.StatusInternalServerError,
		))
		return
	}

	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=client.py")
	c.String(http.StatusOK, sdk)
}

// handleListEndpoints returns all API endpoints available in SDKs
func (h *CodegenHandler) handleListEndpoints(c *gin.Context) {
	endpoints := h.generator.ExtractEndpoints()

	// Group by app
	appEndpoints := make(map[string][]gin.H)
	for _, ep := range endpoints {
		appTag := "core"
		if len(ep.Tags) > 0 {
			appTag = ep.Tags[0]
		}

		endpointInfo := gin.H{
			"method":      ep.Method,
			"path":        ep.Path,
			"summary":     ep.Summary,
			"description": ep.Description,
		}

		appEndpoints[appTag] = append(appEndpoints[appTag], endpointInfo)
	}

	api.RespondWith(c, http.StatusOK, gin.H{
		"total_endpoints": len(endpoints),
		"endpoints":       appEndpoints,
		"documentation":   "/api/docs/openapi.json",
	})
}
