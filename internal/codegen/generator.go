package codegen

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/jgirmay/GAIA_GO/internal/docs"
)

// ClientGenerator generates SDK code from OpenAPI specifications
type ClientGenerator struct {
	spec *docs.OpenAPISpec
}

// NewClientGenerator creates a new code generator
func NewClientGenerator(spec *docs.OpenAPISpec) *ClientGenerator {
	return &ClientGenerator{
		spec: spec,
	}
}

// GenerateTypeScriptSDK generates a TypeScript client SDK
func (g *ClientGenerator) GenerateTypeScriptSDK() (string, error) {
	tmpl := template.Must(template.New("ts-sdk").Parse(typeScriptSDKTemplate))

	data := map[string]interface{}{
		"Title":       g.spec.Info.Title,
		"Description": g.spec.Info.Description,
		"Version":     g.spec.Info.Version,
		"Paths":       g.spec.Paths,
		"Servers":     g.spec.Servers,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to generate TypeScript SDK: %w", err)
	}

	return buf.String(), nil
}

// GenerateGoSDK generates a Go client library
func (g *ClientGenerator) GenerateGoSDK() (string, error) {
	tmpl := template.Must(template.New("go-sdk").Parse(goSDKTemplate))

	data := map[string]interface{}{
		"Title":       g.spec.Info.Title,
		"Description": g.spec.Info.Description,
		"Version":     g.spec.Info.Version,
		"Paths":       g.spec.Paths,
		"Servers":     g.spec.Servers,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to generate Go SDK: %w", err)
	}

	return buf.String(), nil
}

// GeneratePythonSDK generates a Python client package
func (g *ClientGenerator) GeneratePythonSDK() (string, error) {
	tmpl := template.Must(template.New("py-sdk").Parse(pythonSDKTemplate))

	data := map[string]interface{}{
		"Title":       g.spec.Info.Title,
		"Description": g.spec.Info.Description,
		"Version":     g.spec.Info.Version,
		"Paths":       g.spec.Paths,
		"Servers":     g.spec.Servers,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to generate Python SDK: %w", err)
	}

	return buf.String(), nil
}

// Language represents a supported SDK language
type Language string

const (
	TypeScript Language = "typescript"
	Go         Language = "go"
	Python     Language = "python"
)

// GenerateSDK generates SDK code for the specified language
func (g *ClientGenerator) GenerateSDK(lang Language) (string, error) {
	switch lang {
	case TypeScript:
		return g.GenerateTypeScriptSDK()
	case Go:
		return g.GenerateGoSDK()
	case Python:
		return g.GeneratePythonSDK()
	default:
		return "", fmt.Errorf("unsupported language: %s", lang)
	}
}

// SDKInfo contains information about a generated SDK
type SDKInfo struct {
	Language    Language
	FileName    string
	FileType    string
	Description string
	Size        int
}

// GetSDKInfo returns information about an SDK for a specific language
func GetSDKInfo(lang Language) SDKInfo {
	switch lang {
	case TypeScript:
		return SDKInfo{
			Language:    TypeScript,
			FileName:    "gaia-client.ts",
			FileType:    "TypeScript",
			Description: "Complete TypeScript/JavaScript client library with full type safety",
			Size:        0, // Will be set after generation
		}
	case Go:
		return SDKInfo{
			Language:    Go,
			FileName:    "client.go",
			FileType:    "Go",
			Description: "Native Go client library with proper error handling",
			Size:        0,
		}
	case Python:
		return SDKInfo{
			Language:    Python,
			FileName:    "client.py",
			FileType:    "Python",
			Description: "Python client package with async support",
			Size:        0,
		}
	default:
		return SDKInfo{}
	}
}

// ExtractEndpoints extracts all endpoints from the OpenAPI spec
func (g *ClientGenerator) ExtractEndpoints() []EndpointInfo {
	endpoints := make([]EndpointInfo, 0)

	for path, pathItem := range g.spec.Paths {
		if pathItem.Get != nil {
			endpoints = append(endpoints, EndpointInfo{
				Path:        path,
				Method:      "GET",
				Summary:     pathItem.Get.Summary,
				Description: pathItem.Get.Description,
				Tags:        pathItem.Get.Tags,
			})
		}
		if pathItem.Post != nil {
			endpoints = append(endpoints, EndpointInfo{
				Path:        path,
				Method:      "POST",
				Summary:     pathItem.Post.Summary,
				Description: pathItem.Post.Description,
				Tags:        pathItem.Post.Tags,
			})
		}
		if pathItem.Put != nil {
			endpoints = append(endpoints, EndpointInfo{
				Path:        path,
				Method:      "PUT",
				Summary:     pathItem.Put.Summary,
				Description: pathItem.Put.Description,
				Tags:        pathItem.Put.Tags,
			})
		}
		if pathItem.Delete != nil {
			endpoints = append(endpoints, EndpointInfo{
				Path:        path,
				Method:      "DELETE",
				Summary:     pathItem.Delete.Summary,
				Description: pathItem.Delete.Description,
				Tags:        pathItem.Delete.Tags,
			})
		}
		if pathItem.Patch != nil {
			endpoints = append(endpoints, EndpointInfo{
				Path:        path,
				Method:      "PATCH",
				Summary:     pathItem.Patch.Summary,
				Description: pathItem.Patch.Description,
				Tags:        pathItem.Patch.Tags,
			})
		}
	}

	return endpoints
}

// EndpointInfo represents information about an API endpoint
type EndpointInfo struct {
	Path        string
	Method      string
	Summary     string
	Description string
	Tags        []string
}

// MethodName converts an endpoint path and method to a method name
func MethodName(path, method string) string {
	// Remove /api/ prefix and convert to camelCase
	path = strings.TrimPrefix(path, "/api/")
	parts := strings.Split(path, "/")

	// Start with method in lowercase
	name := strings.ToLower(method)

	// Add path parts in camelCase
	for _, part := range parts {
		if part != "" {
			name += strings.ToUpper(part[:1]) + part[1:]
		}
	}

	return name
}

// TypeScriptType converts a Go type to TypeScript type
func TypeScriptType(goType string) string {
	switch goType {
	case "string":
		return "string"
	case "int", "int64", "float", "float64":
		return "number"
	case "bool":
		return "boolean"
	case "[]string":
		return "string[]"
	case "map[string]interface{}":
		return "Record<string, any>"
	default:
		return "any"
	}
}

// GoType maps interface{} to proper Go type for SDK
func GoType(jsonType string) string {
	switch jsonType {
	case "string":
		return "string"
	case "integer":
		return "int64"
	case "number":
		return "float64"
	case "boolean":
		return "bool"
	case "array":
		return "[]interface{}"
	case "object":
		return "map[string]interface{}"
	default:
		return "interface{}"
	}
}

// PythonType maps interface{} to proper Python type for SDK
func PythonType(jsonType string) string {
	switch jsonType {
	case "string":
		return "str"
	case "integer":
		return "int"
	case "number":
		return "float"
	case "boolean":
		return "bool"
	case "array":
		return "list"
	case "object":
		return "dict"
	default:
		return "Any"
	}
}
