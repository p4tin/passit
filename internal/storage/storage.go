package storage

import (
	"fmt"
	"os"
	"path/filepath"

	"passit/internal/crypto"
	"passit/internal/models"
)

type Storage struct {
	vaultPath string
	vault     *crypto.Vault
	data      *models.Vault
}

func New() *Storage {
	return &Storage{
		vaultPath: getVaultPath(),
	}
}

func (s *Storage) VaultExists() bool {
	_, err := os.Stat(s.vaultPath)
	return err == nil
}

func (s *Storage) CreateVault(password string) error {
	salt, err := crypto.GenerateSalt()
	if err != nil {
		return fmt.Errorf("failed to generate salt: %w", err)
	}

	s.vault = crypto.NewVault(password, salt, crypto.DefaultIterations)
	s.data = models.NewVault()

	if err := s.saveVault(salt, crypto.DefaultIterations); err != nil {
		return fmt.Errorf("failed to save vault: %w", err)
	}

	return nil
}

func (s *Storage) UnlockVault(password string) error {
	if !s.VaultExists() {
		return fmt.Errorf("vault does not exist")
	}

	fileData, err := os.ReadFile(s.vaultPath)
	if err != nil {
		return fmt.Errorf("failed to read vault file: %w", err)
	}

	vaultData, err := crypto.DeserializeVaultData(fileData)
	if err != nil {
		return fmt.Errorf("failed to deserialize vault data: %w", err)
	}

	s.vault = crypto.NewVault(password, vaultData.Salt, vaultData.Iterations)

	decrypted, err := s.vault.Decrypt(vaultData)
	if err != nil {
		return fmt.Errorf("failed to decrypt vault: %w", err)
	}

	vault, err := models.VaultFromJSON(decrypted)
	if err != nil {
		return fmt.Errorf("failed to parse vault data: %w", err)
	}

	s.data = vault
	return nil
}

func (s *Storage) Save() error {
	if s.vault == nil || s.data == nil {
		return fmt.Errorf("vault not initialized")
	}

	fileData, err := os.ReadFile(s.vaultPath)
	if err != nil {
		return fmt.Errorf("failed to read vault file: %w", err)
	}

	vaultData, err := crypto.DeserializeVaultData(fileData)
	if err != nil {
		return fmt.Errorf("failed to deserialize vault data: %w", err)
	}

	return s.saveVault(vaultData.Salt, vaultData.Iterations)
}

func (s *Storage) saveVault(salt []byte, iterations uint32) error {
	jsonData, err := s.data.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to serialize vault: %w", err)
	}

	encryptedData, err := s.vault.Encrypt(jsonData)
	if err != nil {
		return fmt.Errorf("failed to encrypt vault: %w", err)
	}

	encryptedData.Salt = salt
	encryptedData.Iterations = iterations

	serialized := encryptedData.Serialize()

	if err := os.MkdirAll(filepath.Dir(s.vaultPath), 0700); err != nil {
		return fmt.Errorf("failed to create vault directory: %w", err)
	}

	if err := os.WriteFile(s.vaultPath, serialized, 0600); err != nil {
		return fmt.Errorf("failed to write vault file: %w", err)
	}

	return nil
}

func (s *Storage) GetVault() *models.Vault {
	return s.data
}

func (s *Storage) Lock() {
	s.vault = nil
	s.data = nil
}

func (s *Storage) IsUnlocked() bool {
	return s.vault != nil && s.data != nil
}

func (s *Storage) ChangePassword(oldPassword, newPassword string) error {
	if !s.IsUnlocked() {
		return fmt.Errorf("vault is locked")
	}

	salt, err := crypto.GenerateSalt()
	if err != nil {
		return fmt.Errorf("failed to generate new salt: %w", err)
	}

	s.vault = crypto.NewVault(newPassword, salt, crypto.DefaultIterations)

	return s.saveVault(salt, crypto.DefaultIterations)
}

func getVaultPath() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = os.TempDir()
	}

	appDir := filepath.Join(configDir, "passit")
	return filepath.Join(appDir, "vault.enc")
}