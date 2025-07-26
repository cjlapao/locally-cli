// Package encryption provides encryption and decryption capabilities
package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strconv"
	"sync"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/pbkdf2"
)

var (
	instance *EncryptionService
	once     sync.Once
)

// Service provides encryption and decryption capabilities
type Service interface {
	// Hash functions (one-way, for passwords)
	HashPassword(password string) (string, error)
	VerifyPassword(password, hash string) error

	// Symmetric encryption (reversible, for sensitive data)
	EncryptForTenant(tenantID, data string) (string, error)
	DecryptForTenant(tenantID, encryptedData string) (string, error)
	EncryptGlobal(data string) (string, error)
	DecryptGlobal(encryptedData string) (string, error)

	// Utility functions for different data types
	EncryptNumber(tenantID string, number int64) (string, error)
	DecryptNumber(tenantID, encryptedData string) (int64, error)
	EncryptFloat(tenantID string, number float64) (string, error)
	DecryptFloat(tenantID, encryptedData string) (float64, error)
}

type EncryptionService struct {
	masterSecret []byte
	globalKey    []byte
}

// Config represents the encryption service configuration
type Config struct {
	// MasterSecret is used to derive tenant-specific keys
	// This should be a strong, randomly generated secret
	MasterSecret string
	// GlobalSecret is used for global encryption
	GlobalSecret string
}

const (
	// Key derivation parameters
	keyLength  = 32     // AES-256
	saltSize   = 16     // 128-bit salt
	iterations = 100000 // PBKDF2 iterations

	// bcrypt cost (adjust based on your security/performance needs)
	bcryptCost = 12
)

// Initialize initializes the encryption service singleton
func Initialize(cfg Config) (*EncryptionService, error) {
	var initErr error
	once.Do(func() {
		var svc Service
		svc, initErr = newService(cfg)
		if initErr == nil {
			instance = svc.(*EncryptionService)
		}
	})
	return instance, initErr
}

// GetInstance returns the singleton instance of the encryption service
func GetInstance() Service {
	if instance == nil {
		panic("encryption service not initialized")
	}
	return instance
}

// newService creates a new encryption service instance
func newService(cfg Config) (Service, error) {
	if cfg.MasterSecret == "" {
		return nil, fmt.Errorf("master secret is required")
	}
	if cfg.GlobalSecret == "" {
		return nil, fmt.Errorf("global secret is required")
	}

	// Derive global key from global secret
	globalKey := pbkdf2.Key([]byte(cfg.GlobalSecret), []byte("global-salt"), iterations, keyLength, sha256.New)

	return &EncryptionService{
		masterSecret: []byte(cfg.MasterSecret),
		globalKey:    globalKey,
	}, nil
}

// Hash functions

func (s *EncryptionService) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hash), nil
}

func (s *EncryptionService) VerifyPassword(password, hash string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		return fmt.Errorf("password verification failed: %w", err)
	}
	return nil
}

// Symmetric encryption functions

func (s *EncryptionService) EncryptForTenant(tenantID, data string) (string, error) {
	key := s.deriveTenantKey(tenantID)
	return s.encrypt(data, key)
}

func (s *EncryptionService) DecryptForTenant(tenantID, encryptedData string) (string, error) {
	key := s.deriveTenantKey(tenantID)
	return s.decrypt(encryptedData, key)
}

func (s *EncryptionService) EncryptGlobal(data string) (string, error) {
	return s.encrypt(data, s.globalKey)
}

func (s *EncryptionService) DecryptGlobal(encryptedData string) (string, error) {
	return s.decrypt(encryptedData, s.globalKey)
}

// Utility functions for different data types

func (s *EncryptionService) EncryptNumber(tenantID string, number int64) (string, error) {
	return s.EncryptForTenant(tenantID, strconv.FormatInt(number, 10))
}

func (s *EncryptionService) DecryptNumber(tenantID, encryptedData string) (int64, error) {
	decrypted, err := s.DecryptForTenant(tenantID, encryptedData)
	if err != nil {
		return 0, err
	}

	number, err := strconv.ParseInt(decrypted, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse decrypted number: %w", err)
	}

	return number, nil
}

func (s *EncryptionService) EncryptFloat(tenantID string, number float64) (string, error) {
	return s.EncryptForTenant(tenantID, strconv.FormatFloat(number, 'f', -1, 64))
}

func (s *EncryptionService) DecryptFloat(tenantID, encryptedData string) (float64, error) {
	decrypted, err := s.DecryptForTenant(tenantID, encryptedData)
	if err != nil {
		return 0, err
	}

	number, err := strconv.ParseFloat(decrypted, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse decrypted float: %w", err)
	}

	return number, nil
}

// Private helper functions

// deriveTenantKey derives a unique encryption key for a tenant
func (s *EncryptionService) deriveTenantKey(tenantID string) []byte {
	// Use tenant ID as salt for key derivation
	salt := sha256.Sum256([]byte(tenantID))
	return pbkdf2.Key(s.masterSecret, salt[:], iterations, keyLength, sha256.New)
}

// encrypt encrypts data using AES-256-GCM
func (s *EncryptionService) encrypt(data string, key []byte) (string, error) {
	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate random nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt the data
	ciphertext := gcm.Seal(nonce, nonce, []byte(data), nil)

	// Encode to base64 for storage
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decrypt decrypts data using AES-256-GCM
func (s *EncryptionService) decrypt(encryptedData string, key []byte) (string, error) {
	// Decode from base64
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Extract nonce and ciphertext
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Decrypt the data
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plaintext), nil
}
