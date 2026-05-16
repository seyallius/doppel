package object

import "time"

type PasswordHistoryEntry struct {
	Hash          string
	PasswordType  string
	CreatedAt     time.Time
	HashAlgorithm string
}

type PasswordHistory struct {
	Entries []*PasswordHistoryEntry
}
