package middleware

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	// Import the new, recommended Sui SDK package
	"github.com/coming-chat/go-sui-sdk/sui_types"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/blake2b"
)

// verifySuiSignature checks if a Sui signature is valid for a given message and wallet address.
// It returns true and nil on success, or false and an error on failure.
func verifySuiSignature(walletAddress, signatureB64, message string) (bool, error) {
	// 1. Parse the signature from the Base64 string.
	// The new SDK's NewSignature function handles Base64 decoding and parsing
	// of the signature scheme, signature bytes, and public key automatically.
	signature, err := sui_types.NewSignature(signatureB64)
	if err != nil {
		return false, fmt.Errorf("failed to parse signature from base64: %w", err)
	}

	// 2. Verify that the address matches the public key embedded in the signature.
	// This prevents a user from submitting a valid signature from a different address.
	derivedAddress := signature.PubKey.SuiAddress().String()
	if derivedAddress != walletAddress {
		return false, fmt.Errorf("address mismatch: provided %s, derived %s", walletAddress, derivedAddress)
	}

	// 3. Prepare the message for verification by wrapping it with a Sui intent.
	// This is a standard procedure for all personal messages signed on Sui.
	messageBytes := []byte(message)
	intent := sui_types.Intent{
		Scope:   sui_types.IntentScopePersonalMessage,
		Version: sui_types.IntentVersionV0,
		AppId:   sui_types.AppIdSui,
	}

	// Combine the intent prefix with the message bytes.
	intentMessage := intent.GetIntentMessage(messageBytes)

	// Hash the combined message using Blake2b-256.
	msgHash := blake2b.Sum256(intentMessage)

	// 4. Verify the signature against the hashed message.
	// The Verify method on the signature object uses its embedded public key.
	if !signature.Verify(msgHash[:]) {
		return false, fmt.Errorf("signature verification failed")
	}

	return true, nil
}

// WalletAuthMiddleware is a Gin middleware for authenticating users via a Sui wallet signature.
func WalletAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Missing authorization header"})
			c.Abort()
			return
		}

		// The expected header format is: Bearer <wallet_address>:<signature>:<nonce>
		prefix := "Bearer "
		if !strings.HasPrefix(authHeader, prefix) {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid authorization format: must start with Bearer"})
			c.Abort()
			return
		}

		parts := strings.Split(strings.TrimPrefix(authHeader, prefix), ":")
		if len(parts) != 3 {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid authorization format: expected <address>:<signature>:<nonce>"})
			c.Abort()
			return
		}

		walletAddress := parts[0]
		signature := parts[1]
		nonce := parts[2]

		// TODO: Validate the nonce to prevent replay attacks.
		// isValidNonce

		// The server constructs the message to be verified based on the nonce.
		// This prevents the client from signing arbitrary messages and ensures consistency.
		// The date is added to potentially limit the signature's validity period.
		message := fmt.Sprintf("Sign in to our service with nonce: %s at %s", nonce, time.Now().UTC().Format("2006-01-02"))

		// Call the verification function and handle the error.
		isValid, err := verifySuiSignature(walletAddress, signature, message)
		if !isValid {
			// Log the detailed error on the server for debugging purposes.
			log.Printf("Signature verification failed for address %s: %v", walletAddress, err)

			// Return a generic error message to the client.
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "Invalid signature",
			})
			c.Abort()
			return
		}

		// On successful verification, set the wallet address in the context for later use.
		c.Set("WalletAddress", walletAddress)
		c.Next()
	}
}
