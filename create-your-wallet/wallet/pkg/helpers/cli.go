package helpers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

func RunBitcoinCLI(command string, args ...string) (interface{}, error) {
	configPath := filepath.Join(GetConfigBasePath(), "config", "bitcoin.conf")
	cliArgs := append([]string{"-conf=" + configPath, "-signet", command}, args...)
	cmd := exec.Command("bitcoin-cli", cliArgs...)

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("erro ao executar bitcoin-cli: %v. Output: %s", err, out.String())
	}

	output := strings.TrimSpace(out.String())

	// Tente decodificar como JSON; caso contrário, retorne a saída bruta
	var result interface{}
	if err := json.Unmarshal([]byte(output), &result); err == nil {
		// Se for JSON, retorna o resultado
		return result, nil
	}

	// Se não for JSON, retorna a string
	return output, nil
}
