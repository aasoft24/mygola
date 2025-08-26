// pkg/helpers/ids.go
package helpers

import (
	"github.com/google/uuid"
)

func GenerateID() string {
	return uuid.New().String()
}
