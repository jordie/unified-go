package codegen

import (
	"fmt"
	"strings"
)

// PackageManager represents different package managers
type PackageManager string

const (
	NPM   PackageManager = "npm"
	PyPI  PackageManager = "pypi"
	GoMod PackageManager = "go"
)

// PackageConfig holds configuration for SDK distribution
type PackageConfig struct {
	Name        string
	Version     string
	Description string
	Author      string
	License     string
	Homepage    string
	Repository  string
	Repository_go string
	BugTracker  string
	Tags        []string
}

// GeneratePackageJSON generates npm package.json
func GeneratePackageJSON(config PackageConfig) string {
	return fmt.Sprintf(`{
  "name": "@gaia/%s",
  "version": "%s",
  "description": "%s",
  "main": "dist/index.js",
  "types": "dist/index.d.ts",
  "files": [
    "dist",
    "src",
    "README.md",
    "LICENSE"
  ],
  "scripts": {
    "build": "tsc",
    "test": "jest",
    "prepublish": "npm run build",
    "release": "npm publish --access public"
  },
  "keywords": %s,
  "author": "%s",
  "license": "%s",
  "homepage": "%s",
  "repository": {
    "type": "git",
    "url": "%s"
  },
  "bugs": {
    "url": "%s"
  },
  "dependencies": {},
  "devDependencies": {
    "typescript": "^5.0.0",
    "@types/node": "^20.0.0",
    "jest": "^29.0.0",
    "@types/jest": "^29.0.0",
    "ts-jest": "^29.0.0"
  },
  "engines": {
    "node": ">=14.0.0"
  }
}
`, config.Name, config.Version, config.Description, tagsToJSON(config.Tags),
		config.Author, config.License, config.Homepage, config.Repository, config.BugTracker)
}

// GenerateSetupPy generates Python setup.py
func GenerateSetupPy(config PackageConfig) string {
	return fmt.Sprintf(`from setuptools import setup, find_packages

with open("README.md", "r", encoding="utf-8") as fh:
    long_description = fh.read()

setup(
    name="gaia-%s",
    version="%s",
    author="%s",
    description="%s",
    long_description=long_description,
    long_description_content_type="text/markdown",
    url="%s",
    project_urls={
        "Bug Tracker": "%s",
        "Documentation": "%s/wiki",
        "Source Code": "%s",
    },
    packages=find_packages(),
    classifiers=[
        "Programming Language :: Python :: 3",
        "Programming Language :: Python :: 3.8",
        "Programming Language :: Python :: 3.9",
        "Programming Language :: Python :: 3.10",
        "Programming Language :: Python :: 3.11",
        "License :: OSI Approved :: MIT License",
        "Operating System :: OS Independent",
    ],
    python_requires=">=3.8",
    install_requires=[
        "requests>=2.28.0",
        "aiohttp>=3.8.0",
    ],
    extras_require={
        "async": ["aiohttp>=3.8.0"],
        "dev": ["pytest>=7.0.0", "pytest-asyncio>=0.20.0"],
    },
)
`, config.Name, config.Version, config.Author, config.Description,
		config.Homepage, config.BugTracker, config.Homepage, config.Repository)
}

// GenerateGoModFile generates go.mod file for Go SDK
func GenerateGoModFile(config PackageConfig) string {
	modulePath := "github.com/jgirmay/gaia-go-client"
	return fmt.Sprintf(`module %s

go 1.21

require (
)

// %s
// Version: %s
`, modulePath, config.Description, config.Version)
}

// GenerateTsConfigJSON generates TypeScript tsconfig.json
func GenerateTsConfigJSON() string {
	return `{
  "compilerOptions": {
    "target": "ES2020",
    "module": "ESNext",
    "lib": ["ES2020"],
    "declaration": true,
    "outDir": "./dist",
    "rootDir": "./src",
    "strict": true,
    "esModuleInterop": true,
    "skipLibCheck": true,
    "forceConsistentCasingInFileNames": true,
    "resolveJsonModule": true,
    "moduleResolution": "node",
    "allowSyntheticDefaultImports": true,
    "declarationMap": true,
    "sourceMap": true
  },
  "include": ["src"],
  "exclude": ["node_modules", "dist"]
}
`
}

// GeneratePyprojectToml generates Python pyproject.toml
func GeneratePyprojectToml(config PackageConfig) string {
	return fmt.Sprintf(`[build-system]
requires = ["setuptools>=65.0", "wheel"]
build-backend = "setuptools.build_meta"

[project]
name = "gaia-%s"
version = "%s"
description = "%s"
readme = "README.md"
requires-python = ">=3.8"
license = {text = "%s"}
authors = [
  {name = "%s"}
]
keywords = %s

dependencies = [
  "requests>=2.28.0",
  "aiohttp>=3.8.0",
]

[project.optional-dependencies]
async = ["aiohttp>=3.8.0"]
dev = [
  "pytest>=7.0.0",
  "pytest-asyncio>=0.20.0",
  "black>=23.0.0",
  "mypy>=1.0.0",
]

[project.urls]
Homepage = "%s"
Documentation = "%s/wiki"
Repository = "%s"
"Bug Tracker" = "%s"

[tool.black]
line-length = 100
target-version = ["py38", "py39", "py310", "py311"]

[tool.mypy]
python_version = "3.8"
warn_return_any = true
warn_unused_configs = true
disallow_untyped_defs = true
`, config.Name, config.Version, config.Description, config.License, config.Author,
		tagsToJSON(config.Tags), config.Homepage, config.Homepage, config.Repository, config.BugTracker)
}

// GenerateGitignore generates .gitignore for SDKs
func GenerateGitignore(lang string) string {
	switch lang {
	case "typescript":
		return `# Dependencies
node_modules/
package-lock.json
yarn.lock

# Build
dist/
build/
*.tsbuildinfo

# IDE
.vscode/
.idea/
*.swp
*.swo
*~

# Environment
.env
.env.local

# Test
coverage/
.nyc_output/

# OS
.DS_Store
Thumbs.db
`
	case "python":
		return `# Byte-compiled / optimized / DLL files
__pycache__/
*.py[cod]
*$py.class

# C extensions
*.so

# Distribution / packaging
.Python
build/
develop-eggs/
dist/
downloads/
eggs/
.eggs/
lib/
lib64/
parts/
sdist/
var/
wheels/
pip-wheel-metadata/
share/python-wheels/

# PyInstaller
*.manifest
*.spec

# Unit test / coverage reports
htmlcov/
.tox/
.nox/
.coverage
.coverage.*
.cache
nosetests.xml
coverage.xml
*.cover
*.py,cover
.hypothesis/
.pytest_cache/

# Environments
.env
.venv
env/
venv/
ENV/
env.bak/
venv.bak/

# IDEs
.vscode/
.idea/
*.swp
*.swo
*~

# OS
.DS_Store
Thumbs.db
`
	case "go":
		return `# Binaries for programs and plugins
*.exe
*.exe~
*.dll
*.so
*.so.*
*.dylib

# Test binary, built with go test -c
*.test

# Output of the go coverage tool
*.out

# Go workspace file
go.work

# IDEs
.vscode/
.idea/
*.swp
*.swo
*~

# OS
.DS_Store
Thumbs.db
`
	default:
		return ""
	}
}

// GenerateREADME generates SDK README
func GenerateREADME(lang, config string) string {
	titles := map[string]string{
		"typescript": "GAIA TypeScript/JavaScript SDK",
		"python":     "GAIA Python SDK",
		"go":         "GAIA Go Client Library",
	}

	installations := map[string]string{
		"typescript": "```bash\nnpm install @gaia/client\n```",
		"python":     "```bash\npip install gaia-client\n```",
		"go":         "```bash\ngo get github.com/jgirmay/gaia-go-client\n```",
	}

	usages := map[string]string{
		"typescript": "### TypeScript\n```typescript\nimport { GAIAClient } from '@gaia/client';\n\nconst client = new GAIAClient({\n  baseURL: 'http://localhost:8080',\n});\n\nconst health = await client.getHealth();\nconsole.log(health);\n```",
		"python": "### Python\n```python\nfrom gaia import GAIAClient, ClientConfig\n\nconfig = ClientConfig(base_url='http://localhost:8080')\nclient = GAIAClient(config)\n\nhealth = client.get_health()\nprint(health)\n```",
		"go": "### Go\n```go\npackage main\n\nimport (\n\t\"fmt\"\n\t\"log\"\n\t\"github.com/jgirmay/gaia-go-client\"\n)\n\nfunc main() {\n\tconfig := gaia.ClientConfig{\n\t\tBaseURL: \"http://localhost:8080\",\n\t}\n\tclient := gaia.NewClient(config)\n\n\thealth, err := client.GetHealth()\n\tif err != nil {\n\t\tlog.Fatal(err)\n\t}\n\tfmt.Println(health)\n}\n```",
	}

	return fmt.Sprintf(`# %s

Auto-generated client SDK for the GAIA API framework.

## Installation

%s

## Quick Start

%s

## Documentation

- [API Documentation](http://localhost:8080/api/docs)
- [OpenAPI Specification](http://localhost:8080/api/docs/openapi.json)

## Features

- ✅ Full type safety
- ✅ Error handling
- ✅ Authentication support
- ✅ Request/response handling
- ✅ Timeout configuration

## License

MIT

## Contributing

Contributions are welcome! Please open an issue or submit a pull request.
`, titles[lang], installations[lang], usages[lang])
}

// GenerateLICENSE generates MIT license
func GenerateLICENSE(author string) string {
	return fmt.Sprintf(`MIT License

Copyright (c) 2024 %s

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
`, author)
}

// Helper function to convert tags to JSON array string
func tagsToJSON(tags []string) string {
	if len(tags) == 0 {
		return "[]"
	}
	quoted := make([]string, len(tags))
	for i, tag := range tags {
		quoted[i] = fmt.Sprintf(`"%s"`, tag)
	}
	return fmt.Sprintf("[%s]", strings.Join(quoted, ", "))
}

// SDKDistributionInfo contains information about SDK distribution
type SDKDistributionInfo struct {
	Language      string
	PackageName   string
	Registry      string
	RegistryURL   string
	PublishCommand string
	InstallCommand string
}

// GetDistributionInfo returns distribution info for a language
func GetDistributionInfo(lang string) SDKDistributionInfo {
	infos := map[string]SDKDistributionInfo{
		"typescript": {
			Language:       "TypeScript/JavaScript",
			PackageName:    "@gaia/client",
			Registry:       "npm",
			RegistryURL:    "https://www.npmjs.com/package/@gaia/client",
			PublishCommand: "npm publish --access public",
			InstallCommand: "npm install @gaia/client",
		},
		"python": {
			Language:       "Python",
			PackageName:    "gaia-client",
			Registry:       "PyPI",
			RegistryURL:    "https://pypi.org/project/gaia-client/",
			PublishCommand: "twine upload dist/*",
			InstallCommand: "pip install gaia-client",
		},
		"go": {
			Language:       "Go",
			PackageName:    "github.com/jgirmay/gaia-go-client",
			Registry:       "Go Modules",
			RegistryURL:    "https://pkg.go.dev/github.com/jgirmay/gaia-go-client",
			PublishCommand: "git tag v{version} && git push origin v{version}",
			InstallCommand: "go get github.com/jgirmay/gaia-go-client",
		},
	}

	if info, exists := infos[lang]; exists {
		return info
	}

	return SDKDistributionInfo{}
}
