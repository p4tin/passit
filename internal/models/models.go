package models

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"io"
	"time"
)

type Account struct {
	ID       string    `json:"id"`
	Username string    `json:"username"`
	Password string    `json:"password"`
	Notes    string    `json:"notes"`
	Updated  time.Time `json:"updated"`
}

type Site struct {
	ID       string     `json:"id"`
	Name     string     `json:"name"`
	URL      string     `json:"url"`
	Accounts []*Account `json:"accounts"`
	Updated  time.Time  `json:"updated"`
}

type Vault struct {
	Sites   []*Site   `json:"sites"`
	Updated time.Time `json:"updated"`
}

func NewVault() *Vault {
	return &Vault{
		Sites:   make([]*Site, 0),
		Updated: time.Now(),
	}
}

func NewSite(name, url string) *Site {
	return &Site{
		ID:       generateID(),
		Name:     name,
		URL:      url,
		Accounts: make([]*Account, 0),
		Updated:  time.Now(),
	}
}

func NewAccount(username, password, notes string) *Account {
	return &Account{
		ID:       generateID(),
		Username: username,
		Password: password,
		Notes:    notes,
		Updated:  time.Now(),
	}
}

func (v *Vault) AddSite(site *Site) {
	v.Sites = append(v.Sites, site)
	v.Updated = time.Now()
}

func (v *Vault) RemoveSite(siteID string) bool {
	for i, site := range v.Sites {
		if site.ID == siteID {
			v.Sites = append(v.Sites[:i], v.Sites[i+1:]...)
			v.Updated = time.Now()
			return true
		}
	}
	return false
}

func (v *Vault) FindSite(siteID string) *Site {
	for _, site := range v.Sites {
		if site.ID == siteID {
			return site
		}
	}
	return nil
}

func (s *Site) AddAccount(account *Account) {
	s.Accounts = append(s.Accounts, account)
	s.Updated = time.Now()
}

func (s *Site) RemoveAccount(accountID string) bool {
	for i, account := range s.Accounts {
		if account.ID == accountID {
			s.Accounts = append(s.Accounts[:i], s.Accounts[i+1:]...)
			s.Updated = time.Now()
			return true
		}
	}
	return false
}

func (s *Site) FindAccount(accountID string) *Account {
	for _, account := range s.Accounts {
		if account.ID == accountID {
			return account
		}
	}
	return nil
}

func (v *Vault) Search(query string) ([]*Site, []*Account) {
	var matchingSites []*Site
	var matchingAccounts []*Account

	query = toLower(query)

	for _, site := range v.Sites {
		siteMatches := contains(toLower(site.Name), query)

		if siteMatches {
			matchingSites = append(matchingSites, site)
		}

		for _, account := range site.Accounts {
			if contains(toLower(account.Username), query) {
				matchingAccounts = append(matchingAccounts, account)
			}
		}
	}

	return matchingSites, matchingAccounts
}

func (v *Vault) ToJSON() ([]byte, error) {
	return json.Marshal(v)
}

func VaultFromJSON(data []byte) (*Vault, error) {
	var vault Vault
	err := json.Unmarshal(data, &vault)
	if err != nil {
		return nil, err
	}
	return &vault, nil
}

func generateID() string {
	return generateRandomString(16)
}

func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[randomInt(len(charset))]
	}
	return string(b)
}

func randomInt(max int) int {
	// Generate a random uint64 using crypto/rand
	var b [8]byte
	if _, err := io.ReadFull(rand.Reader, b[:]); err != nil {
		// fallback to time-based if crypto/rand fails (shouldn't happen)
		return int(time.Now().UnixNano()) % max
	}
	// Convert to int and map to [0, max)
	n := int(binary.BigEndian.Uint64(b[:]) % uint64(max))
	if n < 0 {
		n = -n
	}
	return n
}

func toLower(s string) string {
	result := make([]rune, 0, len(s))
	for _, r := range s {
		if r >= 'A' && r <= 'Z' {
			result = append(result, r+32)
		} else {
			result = append(result, r)
		}
	}
	return string(result)
}

func contains(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(substr) > len(s) {
		return false
	}

	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if s[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}
