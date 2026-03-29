package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"golang.org/x/crypto/pbkdf2"
)

const (
	SaltSize       = 32
	NonceSize      = 12
	TagSize        = 16
	DefaultIterations = 100000
	KeySize        = 32
)

type VaultData struct {
	Salt       []byte
	Iterations uint32
	Nonce      []byte
	Tag        []byte
	Ciphertext []byte
}

type Vault struct {
	key []byte
}

func NewVault(password string, salt []byte, iterations uint32) *Vault {
	key := pbkdf2.Key([]byte(password), salt, int(iterations), KeySize, sha256.New)
	return &Vault{key: key}
}

func GenerateSalt() ([]byte, error) {
	salt := make([]byte, SaltSize)
	_, err := rand.Read(salt)
	return salt, err
}

func (v *Vault) Encrypt(plaintext []byte) (*VaultData, error) {
	block, err := aes.NewCipher(v.key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, NonceSize)
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := aesGCM.Seal(nil, nonce, plaintext, nil)
	
	tagSize := aesGCM.Overhead()
	if len(ciphertext) < tagSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	actualCiphertext := ciphertext[:len(ciphertext)-tagSize]
	tag := ciphertext[len(ciphertext)-tagSize:]

	return &VaultData{
		Nonce:      nonce,
		Tag:        tag,
		Ciphertext: actualCiphertext,
	}, nil
}

func (v *Vault) Decrypt(data *VaultData) ([]byte, error) {
	block, err := aes.NewCipher(v.key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	ciphertext := append(data.Ciphertext, data.Tag...)
	
	plaintext, err := aesGCM.Open(nil, data.Nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	return plaintext, nil
}

func (vd *VaultData) Serialize() []byte {
	result := make([]byte, 0, 4+len(vd.Salt)+4+len(vd.Nonce)+len(vd.Tag)+len(vd.Ciphertext))
	
	result = append(result, vd.Salt...)
	
	iterBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(iterBytes, vd.Iterations)
	result = append(result, iterBytes...)
	
	result = append(result, vd.Nonce...)
	result = append(result, vd.Tag...)
	result = append(result, vd.Ciphertext...)
	
	return result
}

func DeserializeVaultData(data []byte) (*VaultData, error) {
	if len(data) < SaltSize+4+NonceSize+TagSize {
		return nil, fmt.Errorf("data too short")
	}
	
	offset := 0
	
	salt := make([]byte, SaltSize)
	copy(salt, data[offset:offset+SaltSize])
	offset += SaltSize
	
	iterations := binary.LittleEndian.Uint32(data[offset : offset+4])
	offset += 4
	
	nonce := make([]byte, NonceSize)
	copy(nonce, data[offset:offset+NonceSize])
	offset += NonceSize
	
	tag := make([]byte, TagSize)
	copy(tag, data[offset:offset+TagSize])
	offset += TagSize
	
	ciphertext := make([]byte, len(data)-offset)
	copy(ciphertext, data[offset:])
	
	return &VaultData{
		Salt:       salt,
		Iterations: iterations,
		Nonce:      nonce,
		Tag:        tag,
		Ciphertext: ciphertext,
	}, nil
}