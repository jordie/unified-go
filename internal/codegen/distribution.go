package codegen

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jgirmay/GAIA_GO/internal/api"
)

// DistributionHandler manages SDK distribution and packaging
type DistributionHandler struct {
	generator *ClientGenerator
	config    PackageConfig
}

// NewDistributionHandler creates a new distribution handler
func NewDistributionHandler(generator *ClientGenerator, config PackageConfig) *DistributionHandler {
	return &DistributionHandler{
		generator: generator,
		config:    config,
	}
}

// RegisterRoutes registers SDK distribution endpoints
func (h *DistributionHandler) RegisterRoutes(engine *gin.Engine) {
	dist := engine.Group("/api/sdk/dist")
	{
		dist.GET("", h.handleDistributionIndex)
		dist.GET("/info", h.handleDistributionInfo)
		dist.GET("/typescript/package.json", h.handleTypeScriptPackageJSON)
		dist.GET("/python/setup.py", h.handlePythonSetupPy)
		dist.GET("/python/pyproject.toml", h.handlePythonPyprojectToml)
		dist.GET("/go/go.mod", h.handleGoModFile)
		dist.GET("/all/:language", h.handleGetDistributionPackage)
	}
}

// handleDistributionIndex returns distribution information index
func (h *DistributionHandler) handleDistributionIndex(c *gin.Context) {
	distributions := []gin.H{
		{
			"language":  "typescript",
			"name":      "TypeScript/JavaScript",
			"registry":  "npm",
			"package":   "@gaia/client",
			"url":       "https://www.npmjs.com/package/@gaia/client",
			"docs":      "/api/sdk/dist/info?lang=typescript",
			"endpoints": []string{
				"/api/sdk/typescript",
				"/api/sdk/dist/typescript/package.json",
				"/api/sdk/dist/all/typescript",
			},
		},
		{
			"language": "python",
			"name":     "Python",
			"registry": "PyPI",
			"package":  "gaia-client",
			"url":      "https://pypi.org/project/gaia-client/",
			"docs":     "/api/sdk/dist/info?lang=python",
			"endpoints": []string{
				"/api/sdk/python",
				"/api/sdk/dist/python/setup.py",
				"/api/sdk/dist/python/pyproject.toml",
				"/api/sdk/dist/all/python",
			},
		},
		{
			"language": "go",
			"name":     "Go",
			"registry": "Go Modules",
			"package":  "github.com/jgirmay/gaia-go-client",
			"url":      "https://pkg.go.dev/github.com/jgirmay/gaia-go-client",
			"docs":     "/api/sdk/dist/info?lang=go",
			"endpoints": []string{
				"/api/sdk/go",
				"/api/sdk/dist/go/go.mod",
				"/api/sdk/dist/all/go",
			},
		},
	}

	api.RespondWith(c, http.StatusOK, gin.H{
		"message":        "SDK Distribution and Package Management",
		"description":    "Manage SDK packages across multiple languages and registries",
		"distributions":  distributions,
		"documentation": "/api/docs",
	})
}

// handleDistributionInfo returns detailed distribution information
func (h *DistributionHandler) handleDistributionInfo(c *gin.Context) {
	lang := c.DefaultQuery("lang", "typescript")

	info := GetDistributionInfo(lang)
	if info.Language == "" {
		api.RespondWithError(c, api.NewError(
			api.ErrCodeBadRequest,
			"Unsupported language: "+lang,
			http.StatusBadRequest,
		))
		return
	}

	api.RespondWith(c, http.StatusOK, gin.H{
		"language":          info.Language,
		"package_name":      info.PackageName,
		"registry":          info.Registry,
		"registry_url":      info.RegistryURL,
		"publish_command":   info.PublishCommand,
		"install_command":   info.InstallCommand,
		"distribution_docs": "https://docs.gaia.io/sdk/distribution",
		"tutorial":          "https://docs.gaia.io/sdk/publish-" + lang,
	})
}

// handleTypeScriptPackageJSON returns TypeScript package.json
func (h *DistributionHandler) handleTypeScriptPackageJSON(c *gin.Context) {
	content := GeneratePackageJSON(h.config)
	c.Header("Content-Type", "application/json; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=package.json")
	c.String(http.StatusOK, content)
}

// handlePythonSetupPy returns Python setup.py
func (h *DistributionHandler) handlePythonSetupPy(c *gin.Context) {
	content := GenerateSetupPy(h.config)
	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=setup.py")
	c.String(http.StatusOK, content)
}

// handlePythonPyprojectToml returns Python pyproject.toml
func (h *DistributionHandler) handlePythonPyprojectToml(c *gin.Context) {
	content := GeneratePyprojectToml(h.config)
	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=pyproject.toml")
	c.String(http.StatusOK, content)
}

// handleGoModFile returns Go go.mod
func (h *DistributionHandler) handleGoModFile(c *gin.Context) {
	content := GenerateGoModFile(h.config)
	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=go.mod")
	c.String(http.StatusOK, content)
}

// handleGetDistributionPackage returns a complete package structure
func (h *DistributionHandler) handleGetDistributionPackage(c *gin.Context) {
	lang := c.Param("language")

	packageStructure := gin.H{
		"language": lang,
		"files": gin.H{
			"source": map[string]string{
				"main":      "Client SDK source code",
				"types":     "Type definitions (TypeScript/Python)",
				"examples":  "Usage examples",
			},
			"config": map[string]string{},
			"docs": map[string]string{
				"README":    "Project documentation",
				"CHANGELOG": "Version history",
				"LICENSE":   "MIT License",
			},
			"build": map[string]string{},
		},
		"publishing_guide": "Step-by-step guide to publish to registries",
	}

	// Add language-specific files
	switch lang {
	case "typescript":
		packageStructure["files"].(gin.H)["config"] = gin.H{
			"package.json": "npm package configuration",
			"tsconfig.json": "TypeScript compiler options",
			".gitignore":   "Git ignore patterns",
		}
		packageStructure["files"].(gin.H)["build"] = gin.H{
			"dist":       "Compiled JavaScript output",
			"src":        "TypeScript source files",
		}

	case "python":
		packageStructure["files"].(gin.H)["config"] = gin.H{
			"setup.py":      "Package configuration (legacy)",
			"pyproject.toml": "Modern Python packaging (PEP 517)",
			".gitignore":    "Git ignore patterns",
		}
		packageStructure["files"].(gin.H)["build"] = gin.H{
			"dist":   "Build artifacts",
			"build":  "Build directory",
			"src":    "Python source files",
		}

	case "go":
		packageStructure["files"].(gin.H)["config"] = gin.H{
			"go.mod": "Go module definition",
			"go.sum": "Go module checksums",
		}
		packageStructure["files"].(gin.H)["build"] = gin.H{
			"bin": "Compiled binaries",
		}

	default:
		api.RespondWithError(c, api.NewError(
			api.ErrCodeBadRequest,
			"Unsupported language: "+lang,
			http.StatusBadRequest,
		))
		return
	}

	api.RespondWith(c, http.StatusOK, packageStructure)
}
