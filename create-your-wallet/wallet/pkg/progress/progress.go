package progress

import (
	"encoding/json"
	"fmt"

	"wallet/pkg/models"

	badger "github.com/dgraph-io/badger/v4"
)

const (
	progressKey    = "block_progress_state"
	walletStateKey = "wallet_state"
)

// SaveProgress salva o progresso do bloco atual no BadgerDB
func SaveProgress(db *badger.DB, blockHeight int, state *models.WalletState) error {
	return db.Update(func(txn *badger.Txn) error {
		// Salvar altura do bloco
		progressData, err := json.Marshal(blockHeight)
		if err != nil {
			return fmt.Errorf("erro ao serializar progresso: %w", err)
		}
		if err := txn.Set([]byte(progressKey), progressData); err != nil {
			return fmt.Errorf("erro ao salvar progresso: %w", err)
		}

		// Salvar estado da carteira
		stateData, err := json.Marshal(state)
		if err != nil {
			return fmt.Errorf("erro ao serializar estado da carteira: %w", err)
		}
		if err := txn.Set([]byte(walletStateKey), stateData); err != nil {
			return fmt.Errorf("erro ao salvar estado da carteira: %w", err)
		}

		return nil
	})
}

// LoadProgress carrega o progresso e o estado da carteira do BadgerDB
func LoadProgress(db *badger.DB) (int, *models.WalletState, error) {
	var blockHeight int
	var state models.WalletState

	err := db.View(func(txn *badger.Txn) error {
		// Carregar altura do bloco
		item, err := txn.Get([]byte(progressKey))
		if err != nil {
			if err == badger.ErrKeyNotFound {
				return nil // Nenhum progresso salvo
			}
			return err
		}
		if err := item.Value(func(val []byte) error {
			return json.Unmarshal(val, &blockHeight)
		}); err != nil {
			return fmt.Errorf("erro ao desserializar progresso: %w", err)
		}

		// Carregar estado da carteira
		item, err = txn.Get([]byte(walletStateKey))
		if err != nil {
			if err == badger.ErrKeyNotFound {
				return nil // Nenhum estado salvo
			}
			return err
		}
		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &state)
		})
	})

	if err != nil {
		return 0, nil, fmt.Errorf("erro ao carregar progresso e estado: %w", err)
	}

	return blockHeight, &state, nil
}
