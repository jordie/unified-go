package api

import "github.com/gin-gonic/gin"

// Response is the standard API response envelope
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *APIError   `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

// ListResponse is used for list endpoints with count/pagination
type ListResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Count   int         `json:"count"`
	Total   int         `json:"total,omitempty"`
}

// CreatedResponse is used for resource creation endpoints
type CreatedResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
	ID      interface{} `json:"id"`
}

// SuccessResponse returns a success response with data
func SuccessResponse(data interface{}) *Response {
	return &Response{
		Success: true,
		Data:    data,
	}
}

// SuccessResponseWithMessage returns a success response with message and data
func SuccessResponseWithMessage(data interface{}, message string) *Response {
	return &Response{
		Success: true,
		Data:    data,
		Message: message,
	}
}

// CreatedWithID returns a creation response with ID
func CreatedWithID(id interface{}, data interface{}) *CreatedResponse {
	return &CreatedResponse{
		Success: true,
		ID:      id,
		Data:    data,
		Message: MsgCreated,
	}
}

// ErrorResponse returns an error response
func ErrorResponse(err *APIError) *Response {
	return &Response{
		Success: false,
		Error:   err,
	}
}

// RespondWith sends a success response with data
func RespondWith(c *gin.Context, statusCode int, data interface{}) {
	c.JSON(statusCode, SuccessResponse(data))
}

// RespondWithMessage sends a success response with data and message
func RespondWithMessage(c *gin.Context, statusCode int, data interface{}, message string) {
	c.JSON(statusCode, SuccessResponseWithMessage(data, message))
}

// RespondWithError sends an error response
func RespondWithError(c *gin.Context, err *APIError) {
	c.JSON(err.StatusCode, ErrorResponse(err))
}

// RespondWithCreated sends a creation response
func RespondWithCreated(c *gin.Context, id interface{}, data interface{}) {
	c.JSON(201, CreatedWithID(id, data))
}

// RespondList sends a list response with count
func RespondList(c *gin.Context, data interface{}, count int) {
	c.JSON(200, &ListResponse{
		Success: true,
		Data:    data,
		Count:   count,
	})
}

// RespondListWithTotal sends a list response with count and total
func RespondListWithTotal(c *gin.Context, data interface{}, count, total int) {
	c.JSON(200, &ListResponse{
		Success: true,
		Data:    data,
		Count:   count,
		Total:   total,
	})
}
