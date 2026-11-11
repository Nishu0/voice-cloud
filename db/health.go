package db

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

// HealthConfigHash binds a probe to the exact configuration it observed.
func HealthConfigHash(deployment Deployment, connection Connection) string {
	encoded, _ := json.Marshal(struct {
		Deployment Deployment
		Connection Connection
	}{deployment, connection})
	hash := sha256.Sum256(encoded)
	return hex.EncodeToString(hash[:])
}
