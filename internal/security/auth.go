package security

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type LockoutState struct {
	FailedAttempts int       `json:"failed_attempts"`
	LastFailure    time.Time `json:"last_failure"`
	LockoutUntil   time.Time `json:"lockout_until"`
}

type AuthManager struct {
	lockoutFile string
	state       *LockoutState
}

func NewAuthManager() *AuthManager {
	return &AuthManager{
		lockoutFile: getLockoutPath(),
		state:       &LockoutState{},
	}
}

func (am *AuthManager) LoadLockoutState() error {
	data, err := os.ReadFile(am.lockoutFile)
	if os.IsNotExist(err) {
		am.state = &LockoutState{}
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to read lockout file: %w", err)
	}

	if err := json.Unmarshal(data, am.state); err != nil {
		am.state = &LockoutState{}
		return nil
	}

	return nil
}

func (am *AuthManager) saveLockoutState() error {
	data, err := json.Marshal(am.state)
	if err != nil {
		return fmt.Errorf("failed to marshal lockout state: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(am.lockoutFile), 0700); err != nil {
		return fmt.Errorf("failed to create lockout directory: %w", err)
	}

	if err := os.WriteFile(am.lockoutFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write lockout file: %w", err)
	}

	return nil
}

func (am *AuthManager) IsLockedOut() (bool, time.Duration) {
	now := time.Now()
	if now.Before(am.state.LockoutUntil) {
		return true, am.state.LockoutUntil.Sub(now)
	}
	return false, 0
}

func (am *AuthManager) RecordFailedAttempt() error {
	am.state.FailedAttempts++
	am.state.LastFailure = time.Now()

	var lockoutDuration time.Duration
	switch am.state.FailedAttempts {
	case 1, 2:
		lockoutDuration = 0
	case 3:
		lockoutDuration = 10 * time.Second
	case 4:
		lockoutDuration = 30 * time.Second
	default:
		lockoutDuration = 60 * time.Second
	}

	if lockoutDuration > 0 {
		am.state.LockoutUntil = time.Now().Add(lockoutDuration)
	}

	return am.saveLockoutState()
}

func (am *AuthManager) RecordSuccessfulAttempt() error {
	am.state.FailedAttempts = 0
	am.state.LockoutUntil = time.Time{}
	return am.saveLockoutState()
}

func (am *AuthManager) ClearLockout() error {
	am.state = &LockoutState{}
	return am.saveLockoutState()
}

func (am *AuthManager) GetFailedAttempts() int {
	return am.state.FailedAttempts
}

func (am *AuthManager) GetRemainingLockoutTime() time.Duration {
	if time.Now().Before(am.state.LockoutUntil) {
		return am.state.LockoutUntil.Sub(time.Now())
	}
	return 0
}

func ValidatePassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}
	return nil
}

func getLockoutPath() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = os.TempDir()
	}

	appDir := filepath.Join(configDir, "passit")
	return filepath.Join(appDir, "lockout.json")
}