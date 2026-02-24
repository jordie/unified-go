package codegen

const typeScriptSDKTemplate = `/**
 * {{.Title}} - Client SDK
 *
 * {{.Description}}
 * Version: {{.Version}}
 *
 * Auto-generated from OpenAPI specification.
 * DO NOT EDIT - Changes will be overwritten.
 */

export interface ClientConfig {
  baseURL: string;
  timeout?: number;
  headers?: Record<string, string>;
}

export interface Response<T> {
  status: number;
  data: T;
  error?: string;
}

/**
 * GAIA Client SDK
 *
 * Complete TypeScript client for all GAIA API endpoints.
 */
export class GAIAClient {
  private baseURL: string;
  private timeout: number;
  private headers: Record<string, string>;

  constructor(config: ClientConfig) {
    this.baseURL = config.baseURL || 'http://localhost:8080';
    this.timeout = config.timeout || 30000;
    this.headers = {
      'Content-Type': 'application/json',
      ...config.headers,
    };
  }

  /**
   * Set authentication token
   */
  setAuthToken(token: string): void {
    this.headers['Authorization'] = 'Bearer ' + token;
  }

  /**
   * Make an HTTP request
   */
  private async request<T>(
    method: string,
    path: string,
    data?: any
  ): Promise<Response<T>> {
    const url = this.baseURL + path;
    const options: RequestInit = {
      method,
      headers: this.headers,
      timeout: this.timeout,
    };

    if (data) {
      options.body = JSON.stringify(data);
    }

    try {
      const response = await fetch(url, options);
      const responseData = await response.json();

      return {
        status: response.status,
        data: responseData,
      };
    } catch (error) {
      return {
        status: 500,
        data: null as any,
        error: error instanceof Error ? error.message : "Unknown error",
      };
    }
  }

  // ============================================================================
  // DOCUMENTATION ENDPOINTS
  // ============================================================================

  /**
   * Get API documentation index
   */
  async getDocsIndex(): Promise<Response<any>> {
    return this.request('GET', '/api/docs');
  }

  /**
   * Get OpenAPI specification
   */
  async getOpenAPISpec(): Promise<Response<any>> {
    return this.request('GET', '/api/docs/openapi.json');
  }

  /**
   * Get application directory
   */
  async getApps(): Promise<Response<any>> {
    return this.request('GET', '/api/docs/apps');
  }

  /**
   * Get details for a specific application
   */
  async getAppDetails(appName: string): Promise<Response<any>> {
    return this.request('GET', '/api/docs/apps/' + appName);
  }

  // ============================================================================
  // HEALTH CHECK ENDPOINTS
  // ============================================================================

  /**
   * Get complete system health status
   */
  async getHealth(): Promise<Response<any>> {
    return this.request('GET', '/api/health');
  }

  /**
   * Get liveness status (Kubernetes)
   */
  async getLiveness(): Promise<Response<any>> {
    return this.request('GET', '/api/health/live');
  }

  /**
   * Get readiness status (Kubernetes)
   */
  async getReadiness(): Promise<Response<any>> {
    return this.request('GET', '/api/health/ready');
  }

  /**
   * Get health status for a specific app
   */
  async getAppHealth(appName: string): Promise<Response<any>> {
    return this.request('GET', '/api/health/apps/' + appName);
  }

  // ============================================================================
  // APP-SPECIFIC ENDPOINTS
  // ============================================================================

  // Add app-specific endpoints here based on the OpenAPI spec
  // Example:
  // async getMathProblems(difficulty: string): Promise<Response<any>> {
  //   return this.request('GET', '/api/math/problems/generate?difficulty=' + difficulty);
  // }
}

/**
 * Default export - Create and export a singleton client
 */
export const gaiaClient = new GAIAClient({
  baseURL: 'http://localhost:8080',
});

export default gaiaClient;
`

const goSDKTemplate = `package gaia

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

/*
Package gaia provides a complete Go client for the {{.Title}} API.

{{.Description}}

Version: {{.Version}}

Auto-generated from OpenAPI specification.
DO NOT EDIT - Changes will be overwritten.
*/

// ClientConfig holds configuration for the GAIA client
type ClientConfig struct {
	BaseURL string
	Timeout time.Duration
	Headers map[string]string
}

// Response represents an API response
type Response struct {
	Status int
	Data   interface{}
	Error  string
}

// Client is the main GAIA API client
type Client struct {
	baseURL string
	timeout time.Duration
	headers map[string]string
	client  *http.Client
}

// NewClient creates a new GAIA API client
func NewClient(config ClientConfig) *Client {
	if config.BaseURL == "" {
		config.BaseURL = "http://localhost:8080"
	}
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	return &Client{
		baseURL: config.BaseURL,
		timeout: config.Timeout,
		headers: config.Headers,
		client: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// SetAuthToken sets the authentication token
func (c *Client) SetAuthToken(token string) {
	if c.headers == nil {
		c.headers = make(map[string]string)
	}
	c.headers["Authorization"] = "Bearer " + token
}

// request makes an HTTP request and returns the response
func (c *Client) request(method string, path string, body interface{}) (*Response, error) {
	url := c.baseURL + path

	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(data)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	for key, value := range c.headers {
		req.Header.Set(key, value)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var data interface{}
	if err := json.Unmarshal(respBody, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &Response{
		Status: resp.StatusCode,
		Data:   data,
	}, nil
}

// ============================================================================
// DOCUMENTATION ENDPOINTS
// ============================================================================

// GetDocsIndex returns the documentation index
func (c *Client) GetDocsIndex() (*Response, error) {
	return c.request("GET", "/api/docs", nil)
}

// GetOpenAPISpec returns the OpenAPI specification
func (c *Client) GetOpenAPISpec() (*Response, error) {
	return c.request("GET", "/api/docs/openapi.json", nil)
}

// GetApps returns the application directory
func (c *Client) GetApps() (*Response, error) {
	return c.request("GET", "/api/docs/apps", nil)
}

// GetAppDetails returns details for a specific application
func (c *Client) GetAppDetails(appName string) (*Response, error) {
	return c.request("GET", fmt.Sprintf("/api/docs/apps/%s", appName), nil)
}

// ============================================================================
// HEALTH CHECK ENDPOINTS
// ============================================================================

// GetHealth returns the complete system health status
func (c *Client) GetHealth() (*Response, error) {
	return c.request("GET", "/api/health", nil)
}

// GetLiveness returns liveness status (Kubernetes)
func (c *Client) GetLiveness() (*Response, error) {
	return c.request("GET", "/api/health/live", nil)
}

// GetReadiness returns readiness status (Kubernetes)
func (c *Client) GetReadiness() (*Response, error) {
	return c.request("GET", "/api/health/ready", nil)
}

// GetAppHealth returns health status for a specific app
func (c *Client) GetAppHealth(appName string) (*Response, error) {
	return c.request("GET", fmt.Sprintf("/api/health/apps/%s", appName), nil)
}

// ============================================================================
// APP-SPECIFIC ENDPOINTS
// ============================================================================

// Add app-specific methods here based on the OpenAPI spec
// Example:
// func (c *Client) GetMathProblems(difficulty string) (*Response, error) {
//     return c.request("GET", fmt.Sprintf("/api/math/problems/generate?difficulty=%s", difficulty), nil)
// }
`

const pythonSDKTemplate = `"""
{{.Title}} - Python Client SDK

{{.Description}}

Version: {{.Version}}

Auto-generated from OpenAPI specification.
DO NOT EDIT - Changes will be overwritten.
"""

import asyncio
import json
from typing import Any, Dict, Optional
from urllib.parse import urljoin

try:
    import aiohttp
except ImportError:
    aiohttp = None

import requests


class ClientConfig:
    """Configuration for GAIA client."""

    def __init__(
        self,
        base_url: str = "http://localhost:8080",
        timeout: int = 30,
        headers: Optional[Dict[str, str]] = None,
    ):
        self.base_url = base_url
        self.timeout = timeout
        self.headers = headers or {}


class Response:
    """API response wrapper."""

    def __init__(self, status: int, data: Any = None, error: Optional[str] = None):
        self.status = status
        self.data = data
        self.error = error

    def __repr__(self) -> str:
        return f"Response(status={self.status}, data={self.data}, error={self.error})"


class GAIAClient:
    """
    GAIA Client SDK

    Complete Python client for all GAIA API endpoints.
    Supports both synchronous and asynchronous operations.
    """

    def __init__(self, config: ClientConfig):
        self.base_url = config.base_url
        self.timeout = config.timeout
        self.headers = {"Content-Type": "application/json", **config.headers}
        self.session = requests.Session()

    def set_auth_token(self, token: str) -> None:
        """Set authentication token."""
        self.headers["Authorization"] = f"Bearer {token}"

    def _request(
        self, method: str, path: str, data: Optional[Dict[str, Any]] = None
    ) -> Response:
        """Make an HTTP request."""
        url = urljoin(self.base_url, path)

        try:
            response = self.session.request(
                method,
                url,
                json=data,
                headers=self.headers,
                timeout=self.timeout,
            )

            try:
                response_data = response.json()
            except json.JSONDecodeError:
                response_data = response.text

            return Response(status=response.status_code, data=response_data)

        except requests.RequestException as e:
            return Response(status=500, error=str(e))

    # ========================================================================
    # DOCUMENTATION ENDPOINTS
    # ========================================================================

    def get_docs_index(self) -> Response:
        """Get API documentation index."""
        return self._request("GET", "/api/docs")

    def get_openapi_spec(self) -> Response:
        """Get OpenAPI specification."""
        return self._request("GET", "/api/docs/openapi.json")

    def get_apps(self) -> Response:
        """Get application directory."""
        return self._request("GET", "/api/docs/apps")

    def get_app_details(self, app_name: str) -> Response:
        """Get details for a specific application."""
        return self._request("GET", f"/api/docs/apps/{app_name}")

    # ========================================================================
    # HEALTH CHECK ENDPOINTS
    # ========================================================================

    def get_health(self) -> Response:
        """Get complete system health status."""
        return self._request("GET", "/api/health")

    def get_liveness(self) -> Response:
        """Get liveness status (Kubernetes)."""
        return self._request("GET", "/api/health/live")

    def get_readiness(self) -> Response:
        """Get readiness status (Kubernetes)."""
        return self._request("GET", "/api/health/ready")

    def get_app_health(self, app_name: str) -> Response:
        """Get health status for a specific app."""
        return self._request("GET", f"/api/health/apps/{app_name}")

    # ========================================================================
    # APP-SPECIFIC ENDPOINTS
    # ========================================================================

    # Add app-specific methods here based on the OpenAPI spec
    # Example:
    # def get_math_problems(self, difficulty: str) -> Response:
    #     return self._request("GET", f"/api/math/problems/generate?difficulty={difficulty}")


class AsyncGAIAClient:
    """
    Asynchronous GAIA Client SDK

    Async version of the GAIA client using aiohttp.
    Requires: pip install aiohttp
    """

    def __init__(self, config: ClientConfig):
        if not aiohttp:
            raise ImportError("aiohttp is required for async client. Install with: pip install aiohttp")

        self.base_url = config.base_url
        self.timeout = aiohttp.ClientTimeout(total=config.timeout)
        self.headers = {"Content-Type": "application/json", **config.headers}
        self.session: Optional[aiohttp.ClientSession] = None

    async def __aenter__(self):
        """Async context manager entry."""
        self.session = aiohttp.ClientSession(timeout=self.timeout)
        return self

    async def __aexit__(self, exc_type, exc_val, exc_tb):
        """Async context manager exit."""
        if self.session:
            await self.session.close()

    def set_auth_token(self, token: str) -> None:
        """Set authentication token."""
        self.headers["Authorization"] = f"Bearer {token}"

    async def _request(
        self, method: str, path: str, data: Optional[Dict[str, Any]] = None
    ) -> Response:
        """Make an async HTTP request."""
        if not self.session:
            raise RuntimeError("Client not initialized. Use 'async with' context manager.")

        url = urljoin(self.base_url, path)

        try:
            async with self.session.request(
                method, url, json=data, headers=self.headers
            ) as response:
                try:
                    response_data = await response.json()
                except json.JSONDecodeError:
                    response_data = await response.text()

                return Response(status=response.status, data=response_data)

        except Exception as e:
            return Response(status=500, error=str(e))

    async def get_health(self) -> Response:
        """Get complete system health status."""
        return await self._request("GET", "/api/health")

    async def get_apps(self) -> Response:
        """Get application directory."""
        return await self._request("GET", "/api/docs/apps")


# Default client instance
_default_config = ClientConfig()
gaia_client = GAIAClient(_default_config)


def set_auth_token(token: str) -> None:
    """Set authentication token for the default client."""
    gaia_client.set_auth_token(token)
`
