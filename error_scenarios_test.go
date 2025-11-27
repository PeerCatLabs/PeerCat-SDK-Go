package peercat

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// ============ Malformed Response Tests ============

func TestMalformedJSONResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("not valid json {"))
	}))
	defer server.Close()

	client := mustNew(t, "test_key", WithBaseURL(server.URL), WithMaxRetries(0))
	_, err := client.GetBalance(context.Background())

	if err == nil {
		t.Error("Expected error for malformed JSON response")
	}
}

func TestMalformedJSONErrorResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("invalid error json"))
	}))
	defer server.Close()

	client := mustNew(t, "test_key", WithBaseURL(server.URL), WithMaxRetries(0))
	_, err := client.GetBalance(context.Background())

	if err == nil {
		t.Error("Expected error for malformed JSON error response")
	}
}

func TestEmptyResponseBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		// Empty body
	}))
	defer server.Close()

	client := mustNew(t, "test_key", WithBaseURL(server.URL))
	_, err := client.GetBalance(context.Background())

	// Should either return empty result or error, not panic
	if err != nil {
		// Acceptable - empty body may cause JSON decode error
		return
	}
}

func TestErrorResponseWithoutErrorWrapper(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Something went wrong"})
	}))
	defer server.Close()

	client := mustNew(t, "test_key", WithBaseURL(server.URL), WithMaxRetries(0))
	_, err := client.GetBalance(context.Background())

	if err == nil {
		t.Error("Expected error for 500 response")
	}
}

// ============ HTTP Status Code Tests ============

func TestHTTP403Forbidden(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(apiErrorResponse{
			Error: struct {
				Type    string  `json:"type"`
				Code    string  `json:"code"`
				Message string  `json:"message"`
				Param   *string `json:"param"`
			}{
				Type:    "authentication_error",
				Code:    "forbidden",
				Message: "Access denied",
			},
		})
	}))
	defer server.Close()

	client := mustNew(t, "test_key", WithBaseURL(server.URL), WithMaxRetries(0))
	_, err := client.GetBalance(context.Background())

	if err == nil {
		t.Error("Expected error for 403 response")
	}

	apiErr, ok := err.(*Error)
	if !ok {
		t.Errorf("Expected *Error, got %T", err)
		return
	}

	if apiErr.Status != 403 {
		t.Errorf("Expected status 403, got %d", apiErr.Status)
	}
}

func TestHTTP404NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(apiErrorResponse{
			Error: struct {
				Type    string  `json:"type"`
				Code    string  `json:"code"`
				Message string  `json:"message"`
				Param   *string `json:"param"`
			}{
				Type:    "not_found",
				Code:    "resource_not_found",
				Message: "Generation not found",
			},
		})
	}))
	defer server.Close()

	client := mustNew(t, "test_key", WithBaseURL(server.URL), WithMaxRetries(0))
	_, err := client.GetOnChainStatus(context.Background(), "invalid_tx")

	if err == nil {
		t.Error("Expected error for 404 response")
	}

	apiErr, ok := err.(*Error)
	if !ok {
		t.Errorf("Expected *Error, got %T", err)
		return
	}

	if apiErr.Type != "not_found" {
		t.Errorf("Expected type 'not_found', got '%s'", apiErr.Type)
	}
}

func TestHTTP502BadGateway(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(apiErrorResponse{
			Error: struct {
				Type    string  `json:"type"`
				Code    string  `json:"code"`
				Message string  `json:"message"`
				Param   *string `json:"param"`
			}{
				Type:    "server_error",
				Code:    "bad_gateway",
				Message: "Bad gateway",
			},
		})
	}))
	defer server.Close()

	client := mustNew(t, "test_key", WithBaseURL(server.URL), WithMaxRetries(0))
	_, err := client.GetBalance(context.Background())

	if err == nil {
		t.Error("Expected error for 502 response")
	}

	apiErr, ok := err.(*Error)
	if !ok {
		t.Errorf("Expected *Error, got %T", err)
		return
	}

	if apiErr.Status != 502 {
		t.Errorf("Expected status 502, got %d", apiErr.Status)
	}
}

func TestHTTP503ServiceUnavailable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(apiErrorResponse{
			Error: struct {
				Type    string  `json:"type"`
				Code    string  `json:"code"`
				Message string  `json:"message"`
				Param   *string `json:"param"`
			}{
				Type:    "server_error",
				Code:    "service_unavailable",
				Message: "Service temporarily unavailable",
			},
		})
	}))
	defer server.Close()

	client := mustNew(t, "test_key", WithBaseURL(server.URL), WithMaxRetries(0))
	_, err := client.GetBalance(context.Background())

	if err == nil {
		t.Error("Expected error for 503 response")
	}
}

func TestHTTP504GatewayTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusGatewayTimeout)
		json.NewEncoder(w).Encode(apiErrorResponse{
			Error: struct {
				Type    string  `json:"type"`
				Code    string  `json:"code"`
				Message string  `json:"message"`
				Param   *string `json:"param"`
			}{
				Type:    "server_error",
				Code:    "gateway_timeout",
				Message: "Gateway timeout",
			},
		})
	}))
	defer server.Close()

	client := mustNew(t, "test_key", WithBaseURL(server.URL), WithMaxRetries(0))
	_, err := client.GetBalance(context.Background())

	if err == nil {
		t.Error("Expected error for 504 response")
	}
}

// ============ Retry Behavior Tests ============

func TestRetry5xxErrors(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(apiErrorResponse{
				Error: struct {
					Type    string  `json:"type"`
					Code    string  `json:"code"`
					Message string  `json:"message"`
					Param   *string `json:"param"`
				}{
					Type:    "server_error",
					Code:    "internal_error",
					Message: "Internal error",
				},
			})
			return
		}

		json.NewEncoder(w).Encode(Balance{Credits: 10.0})
	}))
	defer server.Close()

	client := mustNew(t, "test_key", WithBaseURL(server.URL), WithMaxRetries(2))
	balance, err := client.GetBalance(context.Background())

	if err != nil {
		t.Fatalf("Expected success after retries, got error: %v", err)
	}

	if balance.Credits != 10.0 {
		t.Errorf("Expected credits 10.0, got %f", balance.Credits)
	}

	if callCount != 3 {
		t.Errorf("Expected 3 calls (initial + 2 retries), got %d", callCount)
	}
}

func TestNoRetry4xxErrors(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(apiErrorResponse{
			Error: struct {
				Type    string  `json:"type"`
				Code    string  `json:"code"`
				Message string  `json:"message"`
				Param   *string `json:"param"`
			}{
				Type:    "authentication_error",
				Code:    "invalid_api_key",
				Message: "Invalid API key",
			},
		})
	}))
	defer server.Close()

	client := mustNew(t, "test_key", WithBaseURL(server.URL), WithMaxRetries(3))
	_, err := client.GetBalance(context.Background())

	if err == nil {
		t.Error("Expected authentication error")
	}

	if callCount != 1 {
		t.Errorf("Expected only 1 call (no retries for 4xx), got %d", callCount)
	}
}

// ============ Rate Limit Tests ============

func TestRateLimitHeaderParsingDetails(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-RateLimit-Limit", "1000")
		w.Header().Set("X-RateLimit-Remaining", "0")
		w.Header().Set("X-RateLimit-Reset", "1700000000")
		w.Header().Set("Retry-After", "60")
		w.WriteHeader(http.StatusTooManyRequests)
		json.NewEncoder(w).Encode(apiErrorResponse{
			Error: struct {
				Type    string  `json:"type"`
				Code    string  `json:"code"`
				Message string  `json:"message"`
				Param   *string `json:"param"`
			}{
				Type:    "rate_limit_error",
				Code:    "rate_limit_exceeded",
				Message: "Rate limit exceeded",
			},
		})
	}))
	defer server.Close()

	client := mustNew(t, "test_key", WithBaseURL(server.URL), WithMaxRetries(0))
	_, err := client.GetBalance(context.Background())

	if err == nil {
		t.Error("Expected rate limit error")
	}

	apiErr, ok := err.(*Error)
	if !ok {
		t.Errorf("Expected *Error, got %T", err)
		return
	}

	if apiErr.RateLimitInfo.Limit != 1000 {
		t.Errorf("Expected limit 1000, got %d", apiErr.RateLimitInfo.Limit)
	}

	if apiErr.RateLimitInfo.Remaining != 0 {
		t.Errorf("Expected remaining 0, got %d", apiErr.RateLimitInfo.Remaining)
	}

	if apiErr.RateLimitInfo.RetryAfter != 60 {
		t.Errorf("Expected retry after 60, got %d", apiErr.RateLimitInfo.RetryAfter)
	}
}

func TestRateLimitRetry(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(apiErrorResponse{
				Error: struct {
					Type    string  `json:"type"`
					Code    string  `json:"code"`
					Message string  `json:"message"`
					Param   *string `json:"param"`
				}{
					Type:    "rate_limit_error",
					Code:    "rate_limit_exceeded",
					Message: "Rate limit exceeded",
				},
			})
			return
		}

		json.NewEncoder(w).Encode(Balance{Credits: 10.0})
	}))
	defer server.Close()

	client := mustNew(t, "test_key", WithBaseURL(server.URL), WithMaxRetries(1))
	balance, err := client.GetBalance(context.Background())

	if err != nil {
		t.Fatalf("Expected success after rate limit retry, got error: %v", err)
	}

	if balance.Credits != 10.0 {
		t.Errorf("Expected credits 10.0, got %f", balance.Credits)
	}

	if callCount != 2 {
		t.Errorf("Expected 2 calls (initial + 1 retry), got %d", callCount)
	}
}

// ============ Error Property Tests ============

func TestErrorStatusCode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(apiErrorResponse{
			Error: struct {
				Type    string  `json:"type"`
				Code    string  `json:"code"`
				Message string  `json:"message"`
				Param   *string `json:"param"`
			}{
				Type:    "authentication_error",
				Code:    "invalid_api_key",
				Message: "Invalid API key",
			},
		})
	}))
	defer server.Close()

	client := mustNew(t, "test_key", WithBaseURL(server.URL), WithMaxRetries(0))
	_, err := client.GetBalance(context.Background())

	apiErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("Expected *Error, got %T", err)
	}

	if apiErr.Status != 401 {
		t.Errorf("Expected status 401, got %d", apiErr.Status)
	}

	if apiErr.Code != "invalid_api_key" {
		t.Errorf("Expected code 'invalid_api_key', got '%s'", apiErr.Code)
	}

	if apiErr.Type != "authentication_error" {
		t.Errorf("Expected type 'authentication_error', got '%s'", apiErr.Type)
	}
}

func TestErrorWithParam(t *testing.T) {
	param := "model"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(apiErrorResponse{
			Error: struct {
				Type    string  `json:"type"`
				Code    string  `json:"code"`
				Message string  `json:"message"`
				Param   *string `json:"param"`
			}{
				Type:    "invalid_request_error",
				Code:    "invalid_param",
				Message: "Model not found",
				Param:   &param,
			},
		})
	}))
	defer server.Close()

	client := mustNew(t, "test_key", WithBaseURL(server.URL), WithMaxRetries(0))
	_, err := client.Generate(context.Background(), &GenerateParams{
		Prompt: "test",
		Model:  "invalid_model",
	})

	apiErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("Expected *Error, got %T", err)
	}

	if apiErr.Param == nil || *apiErr.Param != "model" {
		t.Errorf("Expected param 'model', got %v", apiErr.Param)
	}
}

// ============ Edge Case Tests ============

func TestVeryLongPrompt(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(apiErrorResponse{
			Error: struct {
				Type    string  `json:"type"`
				Code    string  `json:"code"`
				Message string  `json:"message"`
				Param   *string `json:"param"`
			}{
				Type:    "invalid_request_error",
				Code:    "prompt_too_long",
				Message: "Prompt exceeds maximum length",
			},
		})
	}))
	defer server.Close()

	client := mustNew(t, "test_key", WithBaseURL(server.URL), WithMaxRetries(0))
	longPrompt := strings.Repeat("x", 10000)
	_, err := client.Generate(context.Background(), &GenerateParams{Prompt: longPrompt})

	if err == nil {
		t.Error("Expected error for long prompt")
	}
}

func TestSpecialCharactersInPrompt(t *testing.T) {
	var receivedPrompt string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var params GenerateParams
		json.Unmarshal(body, &params)
		receivedPrompt = params.Prompt

		json.NewEncoder(w).Encode(GenerateResult{
			ID:       "gen_123",
			ImageURL: "https://example.com/image.png",
			Mode:     ModeProduction,
			Usage:    GenerateUsage{CreditsUsed: 0.05, BalanceRemaining: 9.95},
		})
	}))
	defer server.Close()

	client := mustNew(t, "test_key", WithBaseURL(server.URL))
	specialPrompt := `Test with "quotes" and <tags> and émojis 🎨`

	_, err := client.Generate(context.Background(), &GenerateParams{Prompt: specialPrompt})

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if receivedPrompt != specialPrompt {
		t.Errorf("Prompt not preserved correctly.\nExpected: %s\nGot: %s", specialPrompt, receivedPrompt)
	}
}

func TestUnicodeInPrompt(t *testing.T) {
	var receivedPrompt string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var params GenerateParams
		json.Unmarshal(body, &params)
		receivedPrompt = params.Prompt

		json.NewEncoder(w).Encode(GenerateResult{
			ID:       "gen_123",
			ImageURL: "https://example.com/image.png",
			Mode:     ModeProduction,
			Usage:    GenerateUsage{CreditsUsed: 0.05, BalanceRemaining: 9.95},
		})
	}))
	defer server.Close()

	client := mustNew(t, "test_key", WithBaseURL(server.URL))
	unicodePrompt := "日本語テスト 中文测试 한국어테스트"

	_, err := client.Generate(context.Background(), &GenerateParams{Prompt: unicodePrompt})

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if receivedPrompt != unicodePrompt {
		t.Errorf("Unicode prompt not preserved correctly.\nExpected: %s\nGot: %s", unicodePrompt, receivedPrompt)
	}
}

func TestExtraFieldsInResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Response with extra fields not in schema
		w.Write([]byte(`{
			"credits": 10.50,
			"totalDeposited": 50.00,
			"totalSpent": 39.50,
			"totalWithdrawn": 0.00,
			"totalGenerated": 100,
			"unexpectedField": "should be ignored",
			"anotherUnknown": {"nested": "data"}
		}`))
	}))
	defer server.Close()

	client := mustNew(t, "test_key", WithBaseURL(server.URL))
	balance, err := client.GetBalance(context.Background())

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if balance.Credits != 10.50 {
		t.Errorf("Expected credits 10.50, got %f", balance.Credits)
	}
}

func TestVeryLargeNumericValues(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(Balance{
			Credits:        999999999.99,
			TotalDeposited: 1000000000,
			TotalSpent:     0.000001,
			TotalWithdrawn: 0,
			TotalGenerated: 9007199254740991, // Max safe integer in JS
		})
	}))
	defer server.Close()

	client := mustNew(t, "test_key", WithBaseURL(server.URL))
	balance, err := client.GetBalance(context.Background())

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if balance.Credits != 999999999.99 {
		t.Errorf("Expected credits 999999999.99, got %f", balance.Credits)
	}

	if balance.TotalGenerated != 9007199254740991 {
		t.Errorf("Expected totalGenerated 9007199254740991, got %d", balance.TotalGenerated)
	}
}

func TestZeroCredits(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(Balance{
			Credits:        0,
			TotalDeposited: 0,
			TotalSpent:     0,
			TotalWithdrawn: 0,
			TotalGenerated: 0,
		})
	}))
	defer server.Close()

	client := mustNew(t, "test_key", WithBaseURL(server.URL))
	balance, err := client.GetBalance(context.Background())

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if balance.Credits != 0 {
		t.Errorf("Expected credits 0, got %f", balance.Credits)
	}
}

// ============ Timeout Tests ============

func TestRequestTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		json.NewEncoder(w).Encode(Balance{Credits: 10.0})
	}))
	defer server.Close()

	client := mustNew(t, "test_key", WithBaseURL(server.URL), WithTimeout(50*time.Millisecond), WithMaxRetries(0))
	_, err := client.GetBalance(context.Background())

	if err == nil {
		t.Error("Expected timeout error")
	}
}

func TestContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		json.NewEncoder(w).Encode(Balance{Credits: 10.0})
	}))
	defer server.Close()

	client := mustNew(t, "test_key", WithBaseURL(server.URL))

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := client.GetBalance(ctx)

	if err == nil {
		t.Error("Expected context timeout error")
	}
}

// ============ API Key Tests ============

func TestAPIKeyInAuthorizationHeader(t *testing.T) {
	var authHeader string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader = r.Header.Get("Authorization")
		json.NewEncoder(w).Encode(Balance{Credits: 10.0})
	}))
	defer server.Close()

	client := mustNew(t, "pcat_live_test123", WithBaseURL(server.URL))
	_, _ = client.GetBalance(context.Background())

	expected := "Bearer pcat_live_test123"
	if authHeader != expected {
		t.Errorf("Expected Authorization header '%s', got '%s'", expected, authHeader)
	}
}
