# peercat-sdk-go

Official Go SDK for the PeerCat AI image generation API.

[![Go Reference](https://pkg.go.dev/badge/github.com/peercat/peercat-sdk-go.svg)](https://pkg.go.dev/github.com/peercat/peercat-sdk-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/peercat/peercat-sdk-go)](https://goreportcard.com/report/github.com/peercat/peercat-sdk-go)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## Installation

```bash
go get github.com/peercat/peercat-sdk-go
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/peercat/peercat-sdk-go"
)

func main() {
    client := peercat.New("pcat_live_xxx")

    result, err := client.Generate(context.Background(), &peercat.GenerateParams{
        Prompt: "A beautiful sunset over mountains",
        Model:  "stable-diffusion-xl",
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("Image URL:", result.ImageURL)
}
```

## Features

- Context support for cancellation and timeouts
- Automatic retries with exponential backoff
- Functional options pattern for configuration
- Strongly typed API responses
- Comprehensive error handling
- On-chain SOL payment support

## Configuration

```go
import "time"

client := peercat.New("pcat_live_xxx",
    peercat.WithBaseURL("https://custom.api.url"),
    peercat.WithTimeout(30 * time.Second),
    peercat.WithMaxRetries(5),
)

// Or use a custom HTTP client
client := peercat.New("pcat_live_xxx",
    peercat.WithHTTPClient(&http.Client{
        Transport: customTransport,
    }),
)
```

## API Reference

### Image Generation

```go
// Basic generation
result, err := client.Generate(ctx, &peercat.GenerateParams{
    Prompt: "A futuristic cityscape",
})

// With options
result, err := client.Generate(ctx, &peercat.GenerateParams{
    Prompt: "A majestic dragon",
    Model:  "stable-diffusion-xl",
    Mode:   peercat.ModeDemo,  // Free, returns placeholder
})

fmt.Println("Image:", result.ImageURL)
fmt.Println("Credits used:", result.Usage.CreditsUsed)
```

### Models & Pricing

```go
// List available models
models, err := client.GetModels(ctx)
for _, model := range models {
    fmt.Printf("%s: $%.2f\n", model.ID, model.PriceUSD)
}

// Get current prices (including SOL conversion)
prices, err := client.GetPrices(ctx)
fmt.Printf("SOL/USD: $%.2f\n", prices.SOLPrice)
```

### Account

```go
// Get balance
balance, err := client.GetBalance(ctx)
fmt.Printf("Credits: $%.2f\n", balance.Credits)

// Get usage history
history, err := client.GetHistory(ctx, &peercat.HistoryParams{
    Limit:  10,
    Offset: 0,
})
for _, item := range history.Items {
    fmt.Printf("%s: %.4f credits\n", item.Endpoint, item.CreditsUsed)
}
```

### API Keys

```go
// Create a new key (requires wallet signature)
newKey, err := client.CreateKey(ctx, &peercat.CreateKeyParams{
    Name:      "Production App",
    Message:   "Create API key for PeerCat",
    Signature: "base58signature...",
    PublicKey: "walletPublicKey...",
})
// Warning: Full key only shown once!
fmt.Println("API Key:", newKey.Key)

// List keys
keys, err := client.ListKeys(ctx)

// Revoke a key
err := client.RevokeKey(ctx, "key_id")

// Update key name
err := client.UpdateKeyName(ctx, "key_id", "New Name")
```

### On-Chain Payments

For direct SOL payments without credits:

```go
// Step 1: Submit prompt and get payment details
submission, err := client.SubmitPrompt(ctx, &peercat.SubmitPromptParams{
    Prompt: "A majestic dragon",
    Model:  "stable-diffusion-xl",
})

fmt.Printf("Send %.6f SOL to %s\n", submission.RequiredAmount.SOL, submission.PaymentAddress)
fmt.Println("Include memo:", submission.Memo)

// Step 2: After sending payment, check status
status, err := client.GetOnChainStatus(ctx, "txSignature...")

switch status.Status {
case peercat.OnChainStatusCompleted:
    fmt.Println("Image:", *status.ImageURL)
case peercat.OnChainStatusPending, peercat.OnChainStatusProcessing:
    fmt.Println("Still processing...")
case peercat.OnChainStatusFailed:
    fmt.Println("Failed:", *status.Error)
}
```

## Error Handling

```go
result, err := client.Generate(ctx, &peercat.GenerateParams{Prompt: "test"})
if err != nil {
    if apiErr, ok := err.(*peercat.Error); ok {
        switch {
        case apiErr.IsAuthenticationError():
            log.Println("Invalid API key")
        case apiErr.IsInsufficientCredits():
            log.Println("Add more credits")
        case apiErr.IsRateLimitError():
            log.Println("Rate limited, please wait")
        case apiErr.IsInvalidRequestError():
            log.Printf("Invalid request: %s (param: %v)", apiErr.Message, apiErr.Param)
        default:
            log.Printf("API error: %s", apiErr.Message)
        }

        // Check if error is retryable
        if apiErr.IsRetryable() {
            // Implement retry logic
        }
    } else {
        log.Printf("Request failed: %v", err)
    }
    return
}
```

## Context Usage

All methods accept a context for cancellation and timeouts:

```go
// With timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

result, err := client.Generate(ctx, &peercat.GenerateParams{
    Prompt: "test",
})

// Context cancellation is respected during retries
```

## Requirements

- Go 1.18+

## License

MIT
