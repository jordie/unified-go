package models

import (
	"fmt"
	"time"
)

// Response represents a standard API response
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
	Time    time.Time   `json:"timestamp"`
}

// PaginationParams represents pagination parameters
type PaginationParams struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
	Offset   int `json:"offset"`
	Limit    int `json:"limit"`
}

// PaginatedResponse represents a paginated API response
type PaginatedResponse struct {
	Success      bool        `json:"success"`
	Data         interface{} `json:"data"`
	Total        int64       `json:"total"`
	Page         int         `json:"page"`
	PageSize     int         `json:"page_size"`
	TotalPages   int         `json:"total_pages"`
	HasNextPage  bool        `json:"has_next_page"`
	HasPrevPage  bool        `json:"has_prev_page"`
	Error        string      `json:"error,omitempty"`
	Time         time.Time   `json:"timestamp"`
}

// ErrorResponse represents an error API response
type ErrorResponse struct {
	Success   bool      `json:"success"`
	Error     string    `json:"error"`
	Code      string    `json:"code,omitempty"`
	Details   string    `json:"details,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// BatchOperation represents a batch operation request
type BatchOperation struct {
	Operation string          `json:"operation"`
	Items     []interface{}   `json:"items"`
	Options   map[string]interface{} `json:"options,omitempty"`
}

// BatchResult represents the result of a batch operation
type BatchResult struct {
	Operation     string        `json:"operation"`
	TotalItems    int           `json:"total_items"`
	SuccessCount  int           `json:"success_count"`
	FailureCount  int           `json:"failure_count"`
	Results       []BatchItemResult `json:"results"`
	Duration      string        `json:"duration"`
}

// BatchItemResult represents the result of a single batch item
type BatchItemResult struct {
	Index     int         `json:"index"`
	ID        interface{} `json:"id,omitempty"`
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Error     string      `json:"error,omitempty"`
}

// FilterOptions represents filtering options
type FilterOptions struct {
	Status    string                 `json:"status,omitempty"`
	Priority  int                    `json:"priority,omitempty"`
	StartTime time.Time              `json:"start_time,omitempty"`
	EndTime   time.Time              `json:"end_time,omitempty"`
	Search    string                 `json:"search,omitempty"`
	Tags      []string               `json:"tags,omitempty"`
	Custom    map[string]interface{} `json:"custom,omitempty"`
}

// SortOptions represents sorting options
type SortOptions struct {
	Field     string `json:"field"`
	Direction string `json:"direction"` // "asc" or "desc"
}

// QueryParams represents general query parameters
type QueryParams struct {
	Filter FilterOptions  `json:"filter,omitempty"`
	Sort   SortOptions    `json:"sort,omitempty"`
	Pagination PaginationParams `json:"pagination,omitempty"`
}

// HealthStatus represents component health status
type HealthStatus struct {
	Component   string    `json:"component"`
	Status      string    `json:"status"` // "healthy", "degraded", "unhealthy"
	LastChecked time.Time `json:"last_checked"`
	Details     string    `json:"details,omitempty"`
}

// SystemStatus represents overall system status
type SystemStatus struct {
	Overall     string          `json:"overall"`
	Components  []HealthStatus  `json:"components"`
	Uptime      int64           `json:"uptime_seconds"`
	Version     string          `json:"version"`
	Environment string          `json:"environment"`
	Timestamp   time.Time       `json:"timestamp"`
}

// Event represents a system event
type Event struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Severity  string                 `json:"severity"` // "info", "warning", "error"
	Source    string                 `json:"source"`
	Message   string                 `json:"message"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// Notification represents a notification to be sent
type Notification struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Target    string                 `json:"target"`
	Subject   string                 `json:"subject"`
	Body      string                 `json:"body"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
	SentAt    *time.Time             `json:"sent_at,omitempty"`
	ReadAt    *time.Time             `json:"read_at,omitempty"`
}

// Webhook represents a webhook registration
type Webhook struct {
	ID       string   `json:"id"`
	Event    string   `json:"event"`
	URL      string   `json:"url"`
	Active   bool     `json:"active"`
	Retry    int      `json:"retry"`
	Headers  map[string]string `json:"headers,omitempty"`
	Secret   string   `json:"secret,omitempty"`
}

// Config represents a configuration item
type Config struct {
	Key         string      `json:"key"`
	Value       interface{} `json:"value"`
	Description string      `json:"description,omitempty"`
	Type        string      `json:"type,omitempty"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

// NewResponse creates a new success response
func NewResponse(data interface{}) Response {
	return Response{
		Success: true,
		Data:    data,
		Time:    time.Now(),
	}
}

// NewErrorResponse creates a new error response
func NewErrorResponse(err string) ErrorResponse {
	return ErrorResponse{
		Success:   false,
		Error:     err,
		Timestamp: time.Now(),
	}
}

// NewErrorResponseWithCode creates a new error response with code
func NewErrorResponseWithCode(err, code string) ErrorResponse {
	return ErrorResponse{
		Success:   false,
		Error:     err,
		Code:      code,
		Timestamp: time.Now(),
	}
}

// NewPaginatedResponse creates a new paginated response
func NewPaginatedResponse(data interface{}, total int64, page, pageSize int) PaginatedResponse {
	totalPages := (int(total) + pageSize - 1) / pageSize
	return PaginatedResponse{
		Success:     true,
		Data:        data,
		Total:       total,
		Page:        page,
		PageSize:    pageSize,
		TotalPages:  totalPages,
		HasNextPage: page < totalPages,
		HasPrevPage: page > 1,
		Time:        time.Now(),
	}
}

// NewEvent creates a new event
func NewEvent(eventType, source, message, severity string, metadata map[string]interface{}) Event {
	return Event{
		ID:        fmt.Sprintf("evt_%d", time.Now().UnixNano()),
		Type:      eventType,
		Severity:  severity,
		Source:    source,
		Message:   message,
		Metadata:  metadata,
		Timestamp: time.Now(),
	}
}

// NewNotification creates a new notification
func NewNotification(notifType, target, subject, body string) Notification {
	return Notification{
		ID:        fmt.Sprintf("notif_%d", time.Now().UnixNano()),
		Type:      notifType,
		Target:    target,
		Subject:   subject,
		Body:      body,
		CreatedAt: time.Now(),
	}
}
