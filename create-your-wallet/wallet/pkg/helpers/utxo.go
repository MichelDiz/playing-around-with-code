package helpers

import (
	"fmt"
	"wallet/pkg/models"
)

func UpdateUTXO(state *models.WalletState, txid string, voutIndex int, value float64, address string, privateKey []byte) {
	utxoKey := fmt.Sprintf("%s:%d", txid, voutIndex)

	// Inicializa o mapa se ele for nil
	if state.UTXOs == nil {
		state.UTXOs = make(map[string]models.UTXO)
	}

	// Verifique se o UTXO já existe
	if _, exists := state.UTXOs[utxoKey]; exists {
		return // Já processado
	}

	// Adicione o UTXO completo
	state.UTXOs[utxoKey] = models.UTXO{
		TxID:       txid,
		VoutIndex:  voutIndex,
		Address:    address,
		PrivateKey: privateKey,
		Value:      value,
	}

	fmt.Printf("Atualizando UTXO: %s com valor %.8f e endereço %s\n", utxoKey, value, address)
}

func CalculateBalance(state *models.WalletState) {
	state.Balance = 0
	for _, utxo := range state.UTXOs {
		state.Balance += utxo.Value // Soma o valor de cada UTXO
	}
}
