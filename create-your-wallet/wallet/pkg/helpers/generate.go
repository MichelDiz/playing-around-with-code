package helpers

import (
	"encoding/json"
	"fmt"
	"os"
	"wallet/pkg/models"
)

func Gen(state *models.WalletState) {

	// Criar a estrutura para o JSON do importmulti
	var importData []map[string]interface{}
	for i := 0; i < len(state.PrivateKeys); i++ {
		if i >= len(state.Addresses) || len(state.Addresses[i]) == 0 {
			continue
		}

		entry := map[string]interface{}{
			"scriptPubKey": map[string]string{
				"address": state.Addresses[i][0], // O endereço correspondente
			},
			"keys":      []string{fmt.Sprintf("%x", state.PrivateKeys[i])},
			"timestamp": "now",
		}
		importData = append(importData, entry)
	}

	// Serializar para JSON e salvar em um arquivo
	file, err := os.Create("importmulti.json")
	if err != nil {
		fmt.Printf("Erro ao criar arquivo JSON: %v\n", err)
		return
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ") // Formata com identação para facilitar a leitura
	err = encoder.Encode(importData)
	if err != nil {
		fmt.Printf("Erro ao escrever JSON no arquivo: %v\n", err)
		return
	}

	fmt.Println("Arquivo importmulti.json criado com sucesso!")
}
