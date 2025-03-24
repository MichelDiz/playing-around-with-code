package storage

import (
	"encoding/json"
	"fmt"

	badger "github.com/dgraph-io/badger/v4"
)

type DB struct {
	badgerDB *badger.DB
}

// Setup inicializa o BadgerDB e retorna uma instância de DB
func Setup(path string) (*DB, error) {
	opts := badger.DefaultOptions(path)
	opts.Logger = nil
	badgerDB, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}
	return &DB{badgerDB: badgerDB}, nil
}

// Close fecha o banco de dados
func (db *DB) Close() {
	db.badgerDB.Close()
}

// StoreBlock armazena o bloco no banco
func (db *DB) StoreBlock(height int, blockData []byte) error {
	return db.Update(func(txn *badger.Txn) error {
		key := []byte(fmt.Sprintf("block-%d", height))
		return txn.Set(key, blockData)
	})
}

func (s *DB) GetBadgerDB() *badger.DB {
	return s.badgerDB
}

// GetBlock recupera o bloco do banco
func (db *DB) GetBlock(height int) ([]byte, error) {
	var blockData []byte
	err := db.View(func(txn *badger.Txn) error {
		key := []byte(fmt.Sprintf("block-%d", height))
		item, err := txn.Get(key)
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			blockData = append([]byte{}, val...) // Copia os dados
			return nil
		})
	})
	return blockData, err
}

// View permite executar uma função de leitura dentro de uma transação
func (db *DB) View(fn func(txn *badger.Txn) error) error {
	return db.badgerDB.View(fn)
}

// Update permite executar uma função de escrita dentro de uma transação
func (db *DB) Update(fn func(txn *badger.Txn) error) error {
	return db.badgerDB.Update(fn)
}

func (db *DB) Compact() error {
	return db.badgerDB.RunValueLogGC(0.7)
}

//! Daqui pra baixo.

func (db *DB) StoreAddressesInBlock(height int, addressesFound map[string][]string) error {
	key := []byte(fmt.Sprintf("block-h-%d", height))

	// Serializa o mapa de endereços em JSON
	addressesJSON, err := json.Marshal(addressesFound)
	if err != nil {
		return fmt.Errorf("erro ao serializar endereços: %w", err)
	}

	// Armazena os dados no BadgerDB
	return db.Update(func(txn *badger.Txn) error {
		return txn.Set(key, addressesJSON)
	})
}

func (db *DB) GetAddressesInBlock(height int) (map[string][]string, error) {
	key := []byte(fmt.Sprintf("block-%d", height))
	var addressesFound map[string][]string

	// Recupera os dados do BadgerDB
	err := db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}

		// Deserializa o JSON de volta para o mapa
		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &addressesFound)
		})
	})

	return addressesFound, err
}
