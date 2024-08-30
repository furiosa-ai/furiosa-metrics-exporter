package e2e

import (
	"fmt"
	"path/filepath"

	"github.com/google/uuid"
)

func generateRandomString() string {
	UUID, _ := uuid.NewUUID()

	return UUID.String()[:8]
}

func GetAbsolutePath(path string) string {
	absPath, err := filepath.Abs(path)
	if err != nil {
		panic(fmt.Errorf("failed to resolve absolute path: %v", err))
	}

	return absPath
}
