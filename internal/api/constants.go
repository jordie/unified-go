package api

// Error codes for standardized API responses
const (
	ErrCodeInvalidRequest    = "INVALID_REQUEST"
	ErrCodeNotFound          = "NOT_FOUND"
	ErrCodeNotAuthenticated  = "NOT_AUTHENTICATED"
	ErrCodeInternalServer    = "INTERNAL_SERVER_ERROR"
	ErrCodeUnauthorized      = "UNAUTHORIZED"
	ErrCodeConflict          = "CONFLICT"
	ErrCodeValidationFailed  = "VALIDATION_FAILED"
	ErrCodeRateLimited       = "RATE_LIMITED"
	ErrCodeBadRequest        = "BAD_REQUEST"
)

// Standard error messages
const (
	MsgSuccess                = "Operation successful"
	MsgCreated                = "Resource created successfully"
	MsgUpdated                = "Resource updated successfully"
	MsgDeleted                = "Resource deleted successfully"
	MsgInvalidRequest         = "Invalid request parameters"
	MsgNotFound               = "Resource not found"
	MsgNotAuthenticated       = "Authentication required"
	MsgUnauthorized           = "Unauthorized access"
	MsgInternalServerError    = "An internal server error occurred"
	MsgValidationFailed       = "Validation failed"
	MsgDuplicateUser          = "User already exists"
	MsgFailedToSaveSession    = "Failed to save session"
	MsgFailedToFetchStats     = "Failed to fetch statistics"
	MsgFailedToFetchLeaderboard = "Failed to fetch leaderboard"
)
