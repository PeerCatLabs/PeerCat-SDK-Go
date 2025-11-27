package peercat

import (
	"reflect"
	"testing"
)

// Schema validation tests to ensure SDK types match OpenAPI specification
// These tests validate:
// 1. All required fields are present in response types
// 2. Field types match the OpenAPI schema
// 3. JSON field names match OpenAPI property names

// TestModelSchema validates Model struct matches OpenAPI Model schema
func TestModelSchema(t *testing.T) {
	// Required fields per OpenAPI spec
	requiredFields := []string{"id", "name", "description", "provider", "maxPromptLength", "outputFormat", "outputResolution", "priceUsd"}

	modelType := reflect.TypeOf(Model{})
	checkRequiredJSONFields(t, modelType, requiredFields, "Model")

	// Verify field types
	assertFieldType(t, modelType, "ID", "string")
	assertFieldType(t, modelType, "Name", "string")
	assertFieldType(t, modelType, "Description", "string")
	assertFieldType(t, modelType, "Provider", "string")
	assertFieldType(t, modelType, "MaxPromptLength", "int")
	assertFieldType(t, modelType, "OutputFormat", "string")
	assertFieldType(t, modelType, "OutputResolution", "string")
	assertFieldType(t, modelType, "PriceUSD", "float64")
}

// TestBalanceSchema validates Balance struct matches OpenAPI Balance schema
func TestBalanceSchema(t *testing.T) {
	requiredFields := []string{"credits", "totalDeposited", "totalSpent", "totalWithdrawn", "totalGenerated"}

	balanceType := reflect.TypeOf(Balance{})
	checkRequiredJSONFields(t, balanceType, requiredFields, "Balance")

	assertFieldType(t, balanceType, "Credits", "float64")
	assertFieldType(t, balanceType, "TotalDeposited", "float64")
	assertFieldType(t, balanceType, "TotalSpent", "float64")
	assertFieldType(t, balanceType, "TotalWithdrawn", "float64")
	assertFieldType(t, balanceType, "TotalGenerated", "int64")
}

// TestGenerateResultSchema validates GenerateResult struct matches OpenAPI GenerateResponse schema
func TestGenerateResultSchema(t *testing.T) {
	requiredFields := []string{"id", "imageUrl", "model", "mode", "usage"}

	resultType := reflect.TypeOf(GenerateResult{})
	checkRequiredJSONFields(t, resultType, requiredFields, "GenerateResult")

	// ipfsHash is nullable (pointer type)
	ipfsField, found := resultType.FieldByName("IPFSHash")
	if !found {
		t.Error("GenerateResult missing IPFSHash field")
	} else if ipfsField.Type.Kind() != reflect.Ptr {
		t.Error("GenerateResult.IPFSHash should be a pointer (nullable)")
	}
}

// TestPriceResponseSchema validates PriceResponse struct matches OpenAPI PriceResponse schema
func TestPriceResponseSchema(t *testing.T) {
	// Treasury is now required per OpenAPI spec
	requiredFields := []string{"solPrice", "slippageTolerance", "updatedAt", "treasury", "models"}

	priceType := reflect.TypeOf(PriceResponse{})
	checkRequiredJSONFields(t, priceType, requiredFields, "PriceResponse")

	assertFieldType(t, priceType, "SOLPrice", "float64")
	assertFieldType(t, priceType, "SlippageTolerance", "float64")
	assertFieldType(t, priceType, "UpdatedAt", "string")
	assertFieldType(t, priceType, "Treasury", "string")
}

// TestModelPriceSchema validates ModelPrice struct matches OpenAPI ModelPrice schema
func TestModelPriceSchema(t *testing.T) {
	requiredFields := []string{"model", "priceUsd", "priceSol", "priceSolWithSlippage"}

	priceType := reflect.TypeOf(ModelPrice{})
	checkRequiredJSONFields(t, priceType, requiredFields, "ModelPrice")

	assertFieldType(t, priceType, "Model", "string")
	assertFieldType(t, priceType, "PriceUSD", "float64")
	assertFieldType(t, priceType, "PriceSOL", "float64")
	assertFieldType(t, priceType, "PriceSOLWithSlippage", "float64")
}

// TestHistoryItemSchema validates HistoryItem struct matches OpenAPI HistoryItem schema
func TestHistoryItemSchema(t *testing.T) {
	requiredFields := []string{"id", "endpoint", "creditsUsed", "status", "createdAt"}

	itemType := reflect.TypeOf(HistoryItem{})
	checkRequiredJSONFields(t, itemType, requiredFields, "HistoryItem")

	// Check nullable fields are pointers
	assertNullableField(t, itemType, "Model")
	assertNullableField(t, itemType, "RequestID")
	assertNullableField(t, itemType, "CompletedAt")
}

// TestAPIKeySchema validates APIKey struct matches OpenAPI ApiKey schema
func TestAPIKeySchema(t *testing.T) {
	requiredFields := []string{"id", "keyPrefix", "environment", "rateLimitTier", "createdAt", "revoked"}

	keyType := reflect.TypeOf(APIKey{})
	checkRequiredJSONFields(t, keyType, requiredFields, "APIKey")

	// Check nullable fields are pointers
	assertNullableField(t, keyType, "Name")
	assertNullableField(t, keyType, "LastUsedAt")

	assertFieldType(t, keyType, "Revoked", "bool")
}

// TestOnChainGenerationStatusSchema validates OnChainGenerationStatus struct matches OpenAPI schema
func TestOnChainGenerationStatusSchema(t *testing.T) {
	requiredFields := []string{"txSignature", "status"}

	statusType := reflect.TypeOf(OnChainGenerationStatus{})
	checkRequiredJSONFields(t, statusType, requiredFields, "OnChainGenerationStatus")

	// All other fields are optional (nullable)
	assertNullableField(t, statusType, "Model")
	assertNullableField(t, statusType, "CreatedAt")
	assertNullableField(t, statusType, "ImageURL")
	assertNullableField(t, statusType, "IPFSHash")
	assertNullableField(t, statusType, "CompletedAt")
	assertNullableField(t, statusType, "Error")
	assertNullableField(t, statusType, "Message")
}

// TestGenerationModeEnum validates GenerationMode values match OpenAPI enum
func TestGenerationModeEnum(t *testing.T) {
	// OpenAPI spec: enum: [production, demo]
	validModes := []GenerationMode{ModeProduction, ModeDemo}
	expectedValues := []string{"production", "demo"}

	for i, mode := range validModes {
		if string(mode) != expectedValues[i] {
			t.Errorf("GenerationMode value mismatch: got %s, want %s", mode, expectedValues[i])
		}
	}
}

// TestHistoryStatusEnum validates HistoryStatus values match OpenAPI enum
func TestHistoryStatusEnum(t *testing.T) {
	// OpenAPI spec: enum: [pending, completed, refunded]
	validStatuses := []HistoryStatus{HistoryStatusPending, HistoryStatusCompleted, HistoryStatusRefunded}
	expectedValues := []string{"pending", "completed", "refunded"}

	for i, status := range validStatuses {
		if string(status) != expectedValues[i] {
			t.Errorf("HistoryStatus value mismatch: got %s, want %s", status, expectedValues[i])
		}
	}
}

// TestKeyEnvironmentEnum validates KeyEnvironment values match OpenAPI enum
func TestKeyEnvironmentEnum(t *testing.T) {
	// OpenAPI spec: enum: [live, test]
	validEnvs := []KeyEnvironment{KeyEnvironmentLive, KeyEnvironmentTest}
	expectedValues := []string{"live", "test"}

	for i, env := range validEnvs {
		if string(env) != expectedValues[i] {
			t.Errorf("KeyEnvironment value mismatch: got %s, want %s", env, expectedValues[i])
		}
	}
}

// TestOnChainStatusEnum validates OnChainStatus values match OpenAPI enum
func TestOnChainStatusEnum(t *testing.T) {
	// OpenAPI spec: enum: [pending, processing, completed, failed, refunded]
	validStatuses := []OnChainStatus{
		OnChainStatusPending,
		OnChainStatusProcessing,
		OnChainStatusCompleted,
		OnChainStatusFailed,
		OnChainStatusRefunded,
	}
	expectedValues := []string{"pending", "processing", "completed", "failed", "refunded"}

	for i, status := range validStatuses {
		if string(status) != expectedValues[i] {
			t.Errorf("OnChainStatus value mismatch: got %s, want %s", status, expectedValues[i])
		}
	}
}

// TestPromptSubmissionSchema validates PromptSubmission struct matches OpenAPI schema
func TestPromptSubmissionSchema(t *testing.T) {
	requiredFields := []string{"submissionId", "promptHash", "paymentAddress", "requiredAmount", "memo", "model", "slippageTolerance", "expiresAt", "instructions"}

	submissionType := reflect.TypeOf(PromptSubmission{})
	checkRequiredJSONFields(t, submissionType, requiredFields, "PromptSubmission")
}

// TestRequiredAmountSchema validates RequiredAmount struct matches OpenAPI schema
func TestRequiredAmountSchema(t *testing.T) {
	requiredFields := []string{"sol", "lamports", "usd"}

	amountType := reflect.TypeOf(RequiredAmount{})
	checkRequiredJSONFields(t, amountType, requiredFields, "RequiredAmount")

	assertFieldType(t, amountType, "SOL", "float64")
	assertFieldType(t, amountType, "Lamports", "int64")
	assertFieldType(t, amountType, "USD", "float64")
}

// Helper functions

// checkRequiredJSONFields verifies that all required JSON fields exist in the struct
func checkRequiredJSONFields(t *testing.T, structType reflect.Type, requiredFields []string, typeName string) {
	t.Helper()

	// Build map of json tags to field names
	jsonFieldMap := make(map[string]string)
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		jsonTag := field.Tag.Get("json")
		if jsonTag != "" && jsonTag != "-" {
			// Handle "field,omitempty" format
			if idx := len(jsonTag); idx > 0 {
				for j := 0; j < len(jsonTag); j++ {
					if jsonTag[j] == ',' {
						jsonTag = jsonTag[:j]
						break
					}
				}
			}
			jsonFieldMap[jsonTag] = field.Name
		}
	}

	for _, required := range requiredFields {
		if _, found := jsonFieldMap[required]; !found {
			t.Errorf("%s missing required JSON field: %s", typeName, required)
		}
	}
}

// assertFieldType verifies a struct field has the expected type
func assertFieldType(t *testing.T, structType reflect.Type, fieldName string, expectedType string) {
	t.Helper()

	field, found := structType.FieldByName(fieldName)
	if !found {
		t.Errorf("Field %s not found in struct", fieldName)
		return
	}

	actualType := field.Type.String()
	if actualType != expectedType {
		t.Errorf("Field %s type mismatch: got %s, want %s", fieldName, actualType, expectedType)
	}
}

// assertNullableField verifies a struct field is a pointer (nullable)
func assertNullableField(t *testing.T, structType reflect.Type, fieldName string) {
	t.Helper()

	field, found := structType.FieldByName(fieldName)
	if !found {
		t.Errorf("Nullable field %s not found in struct", fieldName)
		return
	}

	if field.Type.Kind() != reflect.Ptr {
		t.Errorf("Field %s should be a pointer (nullable), got %s", fieldName, field.Type.Kind())
	}
}

// Contract tests - validate complete objects can be created that match OpenAPI

func TestContractModelResponse(t *testing.T) {
	// This test ensures we can create a complete Model matching OpenAPI
	model := Model{
		ID:               "stable-diffusion-xl",
		Name:             "Stable Diffusion XL",
		Description:      "High quality image generation",
		Provider:         "stability",
		MaxPromptLength:  2000,
		OutputFormat:     "png",
		OutputResolution: "1024x1024",
		PriceUSD:         0.28,
	}

	if model.ID == "" {
		t.Error("Model.ID should not be empty")
	}
}

func TestContractPriceResponse(t *testing.T) {
	// This test ensures PriceResponse includes treasury (our fix)
	response := PriceResponse{
		SOLPrice:          185.50,
		SlippageTolerance: 0.05,
		UpdatedAt:         "2024-01-15T12:00:00Z",
		Treasury:          "9JKi6Tr7JdsTJw1zNedF5vML9GpPnjHD9DWuZq1oE6nV",
		Models: []ModelPrice{
			{
				Model:                "stable-diffusion-xl",
				PriceUSD:             0.28,
				PriceSOL:             0.00151,
				PriceSOLWithSlippage: 0.00159,
			},
		},
	}

	if response.Treasury == "" {
		t.Error("PriceResponse.Treasury should not be empty (OpenAPI compliance)")
	}
}

func TestContractGenerateResult(t *testing.T) {
	// Production mode with IPFS hash
	ipfsHash := "QmXyz123"
	result := GenerateResult{
		ID:       "gen_123",
		ImageURL: "https://cdn.peerc.at/images/gen_123.png",
		IPFSHash: &ipfsHash,
		Model:    "stable-diffusion-xl",
		Mode:     ModeProduction,
		Usage: GenerateUsage{
			CreditsUsed:      0.28,
			BalanceRemaining: 9.72,
		},
	}

	if result.IPFSHash == nil {
		t.Error("Production mode should have IPFS hash")
	}

	// Demo mode without IPFS hash
	demoResult := GenerateResult{
		ID:       "demo_123",
		ImageURL: "https://cdn.peerc.at/demo/placeholder.png",
		IPFSHash: nil, // Null in demo mode
		Model:    "stable-diffusion-xl",
		Mode:     ModeDemo,
		Usage: GenerateUsage{
			CreditsUsed:      0,
			BalanceRemaining: 10,
		},
	}

	if demoResult.Mode != ModeDemo {
		t.Errorf("Expected mode 'demo', got %s", demoResult.Mode)
	}
}
