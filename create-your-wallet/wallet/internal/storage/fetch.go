package storage

import (
	"encoding/json"
	"fmt"
	"wallet/pkg/helpers"
)

func FetchAndStoreBlock(db *DB, blockHeight int) (map[string]interface{}, error) {
	// Verifique se o bloco já está no banco
	blockData, err := db.GetBlock(blockHeight)
	if err == nil {
		var block map[string]interface{}
		if err := json.Unmarshal(blockData, &block); err != nil {
			return nil, fmt.Errorf("erro ao deserializar bloco: %v", err)
		}
		return block, nil
	}

	blockHash, err := getBlockHash(blockHeight)
	if err != nil {
		return nil, fmt.Errorf("erro ao obter hash do bloco %d: %v", blockHeight, err)
	}

	blockResult, err := helpers.RunBitcoinCLI("getblock", blockHash, "2")
	if err != nil {
		return nil, fmt.Errorf("erro ao obter bloco %s: %v", blockHash, err)
	}

	block, ok := blockResult.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("resultado inesperado do bloco: %v", blockResult)
	}

	blockData, err = json.Marshal(block)
	if err != nil {
		return nil, fmt.Errorf("erro ao serializar bloco: %v", err)
	}

	if err := db.StoreBlock(blockHeight, blockData); err != nil {
		return nil, fmt.Errorf("erro ao armazenar bloco: %v", err)
	}

	return block, nil
}

func getBlockHash(blockHeight int) (string, error) {
	result, err := helpers.RunBitcoinCLI("getblockhash", fmt.Sprintf("%d", blockHeight))
	if err != nil {
		return "", err
	}

	// Cast para string, pois sabemos que o resultado será um hash
	blockHash, ok := result.(string)
	if !ok {
		return "", fmt.Errorf("esperado string, mas recebeu: %v", result)
	}
	return blockHash, nil
}
