package peercat

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	client := New("test_key")

	if client.apiKey != "test_key" {
		t.Errorf("expected apiKey to be 'test_key', got '%s'", client.apiKey)
	}
	if client.baseURL != DefaultBaseURL {
		t.Errorf("expected baseURL to be '%s', got '%s'", DefaultBaseURL, client.baseURL)
	}
	if client.maxRetries != DefaultMaxRetries {
		t.Errorf("expected maxRetries to be %d, got %d", DefaultMaxRetries, client.maxRetries)
	}
}

func TestNewWithOptions(t *testing.T) {
	client := New("test_key",
		WithBaseURL("https://custom.url"),
		WithTimeout(30*time.Second),
		WithMaxRetries(5),
	)

	if client.baseURL != "https://custom.url" {
		t.Errorf("expected baseURL to be 'https://custom.url', got '%s'", client.baseURL)
	}
	if client.maxRetries != 5 {
		t.Errorf("expected maxRetries to be 5, got %d", client.maxRetries)
	}
}

func TestGenerate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/v1/generate" {
			t.Errorf("expected /v1/generate, got %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test_key" {
			t.Errorf("expected Authorization header 'Bearer test_key'")
		}

		var params GenerateParams
		json.NewDecoder(r.Body).Decode(&params)

		if params.Prompt != "test prompt" {
			t.Errorf("expected prompt 'test prompt', got '%s'", params.Prompt)
		}

		json.NewEncoder(w).Encode(GenerateResult{
			ID:       "gen_123",
			ImageURL: "https://example.com/image.png",
			Model:    "stable-diffusion-xl",
			Mode:     ModeProduction,
			Usage: GenerateUsage{
				CreditsUsed:      0.05,
				BalanceRemaining: 9.95,
			},
		})
	}))
	defer server.Close()

	client := New("test_key", WithBaseURL(server.URL))
	result, err := client.Generate(context.Background(), &GenerateParams{
		Prompt: "test prompt",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != "gen_123" {
		t.Errorf("expected ID 'gen_123', got '%s'", result.ID)
	}
	if result.Usage.CreditsUsed != 0.05 {
		t.Errorf("expected CreditsUsed 0.05, got %f", result.Usage.CreditsUsed)
	}
}

func TestGenerateDemo(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var params GenerateParams
		json.NewDecoder(r.Body).Decode(&params)

		if params.Mode != ModeDemo {
			t.Errorf("expected mode 'demo', got '%s'", params.Mode)
		}

		json.NewEncoder(w).Encode(GenerateResult{
			ID:   "demo_123",
			Mode: ModeDemo,
			Usage: GenerateUsage{
				CreditsUsed: 0,
			},
		})
	}))
	defer server.Close()

	client := New("test_key", WithBaseURL(server.URL))
	result, err := client.Generate(context.Background(), &GenerateParams{
		Prompt: "test",
		Mode:   ModeDemo,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Mode != ModeDemo {
		t.Errorf("expected mode 'demo', got '%s'", result.Mode)
	}
	if result.Usage.CreditsUsed != 0 {
		t.Errorf("expected CreditsUsed 0, got %f", result.Usage.CreditsUsed)
	}
}

func TestGetModels(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/v1/models" {
			t.Errorf("expected /v1/models, got %s", r.URL.Path)
		}

		json.NewEncoder(w).Encode(ModelsResponse{
			Models: []Model{
				{ID: "model-1", Name: "Model 1"},
				{ID: "model-2", Name: "Model 2"},
			},
		})
	}))
	defer server.Close()

	client := New("test_key", WithBaseURL(server.URL))
	models, err := client.GetModels(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(models) != 2 {
		t.Errorf("expected 2 models, got %d", len(models))
	}
}

func TestGetBalance(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(Balance{
			Credits:        10.00,
			TotalDeposited: 50.00,
			TotalSpent:     40.00,
		})
	}))
	defer server.Close()

	client := New("test_key", WithBaseURL(server.URL))
	balance, err := client.GetBalance(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if balance.Credits != 10.00 {
		t.Errorf("expected Credits 10.00, got %f", balance.Credits)
	}
}

func TestGetHistory(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("limit") != "10" {
			t.Errorf("expected limit=10, got %s", r.URL.Query().Get("limit"))
		}
		if r.URL.Query().Get("offset") != "20" {
			t.Errorf("expected offset=20, got %s", r.URL.Query().Get("offset"))
		}

		json.NewEncoder(w).Encode(HistoryResponse{
			Items: []HistoryItem{},
			Pagination: Pagination{
				Total:   100,
				Limit:   10,
				Offset:  20,
				HasMore: true,
			},
		})
	}))
	defer server.Close()

	client := New("test_key", WithBaseURL(server.URL))
	history, err := client.GetHistory(context.Background(), &HistoryParams{
		Limit:  10,
		Offset: 20,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if history.Pagination.Total != 100 {
		t.Errorf("expected Total 100, got %d", history.Pagination.Total)
	}
}

func TestAuthenticationError(t *testing.T) {
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

	client := New("bad_key", WithBaseURL(server.URL), WithMaxRetries(0))
	_, err := client.GetBalance(context.Background())

	if err == nil {
		t.Fatal("expected error")
	}

	apiErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}
	if !apiErr.IsAuthenticationError() {
		t.Errorf("expected authentication error")
	}
	if apiErr.Status != 401 {
		t.Errorf("expected status 401, got %d", apiErr.Status)
	}
}

func TestInsufficientCreditsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusPaymentRequired)
		json.NewEncoder(w).Encode(apiErrorResponse{
			Error: struct {
				Type    string  `json:"type"`
				Code    string  `json:"code"`
				Message string  `json:"message"`
				Param   *string `json:"param"`
			}{
				Type:    "insufficient_credits",
				Code:    "insufficient_balance",
				Message: "Insufficient credits",
			},
		})
	}))
	defer server.Close()

	client := New("test_key", WithBaseURL(server.URL), WithMaxRetries(0))
	_, err := client.Generate(context.Background(), &GenerateParams{Prompt: "test"})

	if err == nil {
		t.Fatal("expected error")
	}

	apiErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}
	if !apiErr.IsInsufficientCredits() {
		t.Errorf("expected insufficient credits error")
	}
}

func TestRateLimitError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

	client := New("test_key", WithBaseURL(server.URL), WithMaxRetries(0))
	_, err := client.GetBalance(context.Background())

	if err == nil {
		t.Fatal("expected error")
	}

	apiErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}
	if !apiErr.IsRateLimitError() {
		t.Errorf("expected rate limit error")
	}
	if !apiErr.IsRetryable() {
		t.Errorf("expected error to be retryable")
	}
}

func TestSubmitPrompt(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(PromptSubmission{
			SubmissionID:   "sub_123",
			PromptHash:     "abc123",
			PaymentAddress: "9JKi...",
			RequiredAmount: RequiredAmount{
				SOL:      0.001,
				Lamports: 1000000,
				USD:      0.05,
			},
			Memo: "PCAT:v1:sdxl:abc123",
		})
	}))
	defer server.Close()

	client := New("test_key", WithBaseURL(server.URL))
	result, err := client.SubmitPrompt(context.Background(), &SubmitPromptParams{
		Prompt: "test",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.SubmissionID != "sub_123" {
		t.Errorf("expected SubmissionID 'sub_123', got '%s'", result.SubmissionID)
	}
}

func TestGetOnChainStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/generate/txSig123" {
			t.Errorf("expected /v1/generate/txSig123, got %s", r.URL.Path)
		}

		imageURL := "https://example.com/image.png"
		json.NewEncoder(w).Encode(OnChainGenerationStatus{
			TxSignature: "txSig123",
			Status:      OnChainStatusCompleted,
			ImageURL:    &imageURL,
		})
	}))
	defer server.Close()

	client := New("test_key", WithBaseURL(server.URL))
	status, err := client.GetOnChainStatus(context.Background(), "txSig123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status.Status != OnChainStatusCompleted {
		t.Errorf("expected status 'completed', got '%s'", status.Status)
	}
}

func TestErrorMessage(t *testing.T) {
	err := &Error{
		Type:    "invalid_request_error",
		Code:    "invalid_prompt",
		Message: "Prompt too long",
	}

	expected := "peercat: Prompt too long (invalid_prompt)"
	if err.Error() != expected {
		t.Errorf("expected '%s', got '%s'", expected, err.Error())
	}

	param := "prompt"
	err.Param = &param
	expected = "peercat: Prompt too long (invalid_prompt, param: prompt)"
	if err.Error() != expected {
		t.Errorf("expected '%s', got '%s'", expected, err.Error())
	}
}
