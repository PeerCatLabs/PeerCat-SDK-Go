package peercat

// GenerationMode represents the generation mode
type GenerationMode string

const (
	// ModeProduction is production mode - uses credits
	ModeProduction GenerationMode = "production"
	// ModeDemo is demo mode - free, returns placeholder images
	ModeDemo GenerationMode = "demo"
)

// Model represents an available image generation model
type Model struct {
	ID               string  `json:"id"`
	Name             string  `json:"name"`
	Description      string  `json:"description"`
	Provider         string  `json:"provider"`
	MaxPromptLength  int     `json:"maxPromptLength"`
	OutputFormat     string  `json:"outputFormat"`
	OutputResolution string  `json:"outputResolution"`
	PriceUSD         float64 `json:"priceUsd"`
}

// ModelsResponse is the response from GetModels
type ModelsResponse struct {
	Models []Model `json:"models"`
}

// ModelPrice represents pricing for a specific model
type ModelPrice struct {
	Model                string  `json:"model"`
	PriceUSD             float64 `json:"priceUsd"`
	PriceSOL             float64 `json:"priceSol"`
	PriceSOLWithSlippage float64 `json:"priceSolWithSlippage"`
}

// PriceResponse is the response from GetPrices
type PriceResponse struct {
	SOLPrice          float64      `json:"solPrice"`
	SlippageTolerance float64      `json:"slippageTolerance"`
	UpdatedAt         string       `json:"updatedAt"`
	Treasury          string       `json:"treasury"`
	Models            []ModelPrice `json:"models"`
}

// GenerateParams are the parameters for Generate
type GenerateParams struct {
	Prompt  string                 `json:"prompt"`
	Model   string                 `json:"model,omitempty"`
	Mode    GenerationMode         `json:"mode,omitempty"`
	Options map[string]interface{} `json:"options,omitempty"`
}

// GenerateUsage represents usage information from a generation
type GenerateUsage struct {
	CreditsUsed      float64 `json:"creditsUsed"`
	BalanceRemaining float64 `json:"balanceRemaining"`
}

// GenerateResult is the result of Generate
type GenerateResult struct {
	ID       string         `json:"id"`
	ImageURL string         `json:"imageUrl"`
	IPFSHash *string        `json:"ipfsHash"`
	Model    string         `json:"model"`
	Mode     GenerationMode `json:"mode"`
	Usage    GenerateUsage  `json:"usage"`
}

// Balance represents account balance information
type Balance struct {
	Credits        float64 `json:"credits"`
	TotalDeposited float64 `json:"totalDeposited"`
	TotalSpent     float64 `json:"totalSpent"`
	TotalWithdrawn float64 `json:"totalWithdrawn"`
	TotalGenerated int64   `json:"totalGenerated"`
}

// HistoryParams are the parameters for GetHistory
type HistoryParams struct {
	Limit  int `json:"limit,omitempty"`
	Offset int `json:"offset,omitempty"`
}

// HistoryStatus represents the status of a history item
type HistoryStatus string

const (
	HistoryStatusPending   HistoryStatus = "pending"
	HistoryStatusCompleted HistoryStatus = "completed"
	HistoryStatusRefunded  HistoryStatus = "refunded"
)

// HistoryItem represents a single usage history item
type HistoryItem struct {
	ID          string        `json:"id"`
	Endpoint    string        `json:"endpoint"`
	Model       *string       `json:"model"`
	CreditsUsed float64       `json:"creditsUsed"`
	RequestID   *string       `json:"requestId"`
	Status      HistoryStatus `json:"status"`
	CreatedAt   string        `json:"createdAt"`
	CompletedAt *string       `json:"completedAt"`
}

// Pagination represents pagination information
type Pagination struct {
	Total   int  `json:"total"`
	Limit   int  `json:"limit"`
	Offset  int  `json:"offset"`
	HasMore bool `json:"hasMore"`
}

// HistoryResponse is the response from GetHistory
type HistoryResponse struct {
	Items      []HistoryItem `json:"items"`
	Pagination Pagination    `json:"pagination"`
}

// CreateKeyParams are the parameters for CreateKey
type CreateKeyParams struct {
	Name      string `json:"name,omitempty"`
	Message   string `json:"message"`
	Signature string `json:"signature"`
	PublicKey string `json:"publicKey"`
}

// KeyEnvironment represents the environment for an API key
type KeyEnvironment string

const (
	KeyEnvironmentLive KeyEnvironment = "live"
	KeyEnvironmentTest KeyEnvironment = "test"
)

// APIKey represents an API key
type APIKey struct {
	ID            string         `json:"id"`
	Name          *string        `json:"name"`
	KeyPrefix     string         `json:"keyPrefix"`
	Environment   KeyEnvironment `json:"environment"`
	RateLimitTier string         `json:"rateLimitTier"`
	CreatedAt     string         `json:"createdAt"`
	LastUsedAt    *string        `json:"lastUsedAt"`
	Revoked       bool           `json:"revoked"`
}

// CreateKeyResult is the result of CreateKey
type CreateKeyResult struct {
	ID          string         `json:"id"`
	Key         string         `json:"key"`
	KeyPrefix   string         `json:"keyPrefix"`
	Name        *string        `json:"name"`
	Environment KeyEnvironment `json:"environment"`
	CreatedAt   string         `json:"createdAt"`
	Warning     string         `json:"warning"`
}

// KeysResponse is the response from ListKeys
type KeysResponse struct {
	Keys []APIKey `json:"keys"`
}

// SubmitPromptParams are the parameters for SubmitPrompt
type SubmitPromptParams struct {
	Prompt      string                 `json:"prompt"`
	Model       string                 `json:"model,omitempty"`
	Options     map[string]interface{} `json:"options,omitempty"`
	CallbackURL string                 `json:"callbackUrl,omitempty"`
}

// RequiredAmount represents the required payment amount
type RequiredAmount struct {
	SOL      float64 `json:"sol"`
	Lamports int64   `json:"lamports"`
	USD      float64 `json:"usd"`
}

// PromptSubmission is the result of SubmitPrompt
type PromptSubmission struct {
	SubmissionID      string            `json:"submissionId"`
	PromptHash        string            `json:"promptHash"`
	PaymentAddress    string            `json:"paymentAddress"`
	RequiredAmount    RequiredAmount    `json:"requiredAmount"`
	Memo              string            `json:"memo"`
	Model             string            `json:"model"`
	SlippageTolerance float64           `json:"slippageTolerance"`
	ExpiresAt         string            `json:"expiresAt"`
	Instructions      map[string]string `json:"instructions"`
}

// OnChainStatus represents the status of an on-chain generation
type OnChainStatus string

const (
	OnChainStatusPending    OnChainStatus = "pending"
	OnChainStatusProcessing OnChainStatus = "processing"
	OnChainStatusCompleted  OnChainStatus = "completed"
	OnChainStatusFailed     OnChainStatus = "failed"
	OnChainStatusRefunded   OnChainStatus = "refunded"
)

// OnChainGenerationStatus is the status of an on-chain generation
type OnChainGenerationStatus struct {
	TxSignature string        `json:"txSignature"`
	Status      OnChainStatus `json:"status"`
	Model       *string       `json:"model,omitempty"`
	CreatedAt   *string       `json:"createdAt,omitempty"`
	ImageURL    *string       `json:"imageUrl,omitempty"`
	IPFSHash    *string       `json:"ipfsHash,omitempty"`
	CompletedAt *string       `json:"completedAt,omitempty"`
	Error       *string       `json:"error,omitempty"`
	Message     *string       `json:"message,omitempty"`
}
