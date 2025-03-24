package helpers

import (
	"fmt"
	"os"
	"path/filepath"
)

var configBasePath string

func init() {
	var err error
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Println("Erro ao obter o caminho atual:", err)
		os.Exit(1)
	}

	configBasePath = filepath.Join(currentDir, "..")
}

func GetConfigBasePath() string {
	return configBasePath
}
