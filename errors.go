package peercat

import "fmt"

// Error represents an API error
type Error struct {
	Status  int     `json:"-"`
	Type    string  `json:"type"`
	Code    string  `json:"code"`
	Message string  `json:"message"`
	Param   *string `json:"param,omitempty"`
}

// Error implements the error interface
func (e *Error) Error() string {
	if e.Param != nil {
		return fmt.Sprintf("peercat: %s (%s, param: %s)", e.Message, e.Code, *e.Param)
	}
	return fmt.Sprintf("peercat: %s (%s)", e.Message, e.Code)
}

// IsRetryable returns true if the error is retryable
func (e *Error) IsRetryable() bool {
	return e.Status >= 500 || e.Status == 429
}

// IsAuthenticationError returns true if this is an authentication error
func (e *Error) IsAuthenticationError() bool {
	return e.Type == "authentication_error"
}

// IsInsufficientCredits returns true if this is an insufficient credits error
func (e *Error) IsInsufficientCredits() bool {
	return e.Type == "insufficient_credits"
}

// IsRateLimitError returns true if this is a rate limit error
func (e *Error) IsRateLimitError() bool {
	return e.Type == "rate_limit_error"
}

// IsInvalidRequestError returns true if this is an invalid request error
func (e *Error) IsInvalidRequestError() bool {
	return e.Type == "invalid_request_error"
}

// IsNotFoundError returns true if this is a not found error
func (e *Error) IsNotFoundError() bool {
	return e.Type == "not_found"
}

// apiErrorResponse is the API error response format
type apiErrorResponse struct {
	Error struct {
		Type    string  `json:"type"`
		Code    string  `json:"code"`
		Message string  `json:"message"`
		Param   *string `json:"param"`
	} `json:"error"`
}

// errorFromResponse creates an Error from an API error response
func errorFromResponse(status int, resp *apiErrorResponse) *Error {
	return &Error{
		Status:  status,
		Type:    resp.Error.Type,
		Code:    resp.Error.Code,
		Message: resp.Error.Message,
		Param:   resp.Error.Param,
	}
}
