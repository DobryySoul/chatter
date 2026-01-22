package domain

import "time"

type RefreshToken struct {
	ExpiresAt time.Time
	UpdatedAt time.Time
	ID        string
	TokenHash string
	DeviceID  string
	Revoked   bool
	UserID    uint64
}
