package utils

import (
	"crypto/rand"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

const fieldKeyPrefix = "fld_"
const fieldKeyRandLen = 8
const fieldKeyAlphabet = "abcdefghijklmnopqrstuvwxyz0123456789"

// GenerateFieldKey returns a background key like fld_a3x9kp2q.
// Random, not derived from label. 62^8 combos per form scope;
// uniqueness enforced by partial UNIQUE index with 23505 retry in service.
func GenerateFieldKey() (string, error) {
	buf := make([]byte, fieldKeyRandLen)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	out := make([]byte, 0, len(fieldKeyPrefix)+fieldKeyRandLen)
	out = append(out, fieldKeyPrefix...)
	for _, b := range buf {
		out = append(out, fieldKeyAlphabet[int(b)%len(fieldKeyAlphabet)])
	}
	return string(out), nil
}

// IsUniqueViolation reports Postgres 23505 (partial UNIQUE on (form_id,key)/(template_id,key)).
func IsUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
