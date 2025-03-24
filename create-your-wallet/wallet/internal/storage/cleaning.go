package storage

import (
	"fmt"

	badger "github.com/dgraph-io/badger/v4"
)

// ClearBlocks remove todas as chaves relacionadas a blocos ("block-").
func (db *DB) ClearBlocks() error {
	return db.Update(func(txn *badger.Txn) error {
		prefix := []byte("block-")
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			if err := txn.Delete(it.Item().Key()); err != nil {
				return fmt.Errorf("erro ao remover bloco: %w", err)
			}
		}
		return nil
	})
}

// ClearWalletHashes remove todas as chaves relacionadas a hashes de carteiras ("found-wallet-").
func (db *DB) ClearWalletHashes() error {
	return db.Update(func(txn *badger.Txn) error {
		prefix := []byte("found-wallet-")
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			key := it.Item().Key()
			fmt.Printf("Removendo chave: %s\n", key) // Log para depuração
			if err := txn.Delete(key); err != nil {
				return fmt.Errorf("erro ao remover chave %s: %w", key, err)
			}
		}
		return nil
	})
}

// ClearProcessedBlocks remove todas as chaves relacionadas a blocos processados ("processed-block-").
func (db *DB) ClearProcessedBlocks() error {
	return db.Update(func(txn *badger.Txn) error {
		prefix := []byte("processed-block-")
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			key := it.Item().Key()
			fmt.Printf("Removendo chave: %s\n", key) // Log para depuração
			if err := txn.Delete(key); err != nil {
				return fmt.Errorf("erro ao remover chave %s: %w", key, err)
			}
		}
		return nil
	})
}

// ClearHashes remove todas as chaves relacionadas a hashes de blocos ("hash-").
func (db *DB) ClearHashes() error {
	return db.Update(func(txn *badger.Txn) error {
		prefix := []byte("hash-")
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			key := it.Item().Key()
			fmt.Printf("Removendo chave: %s\n", key) // Log para depuração
			if err := txn.Delete(key); err != nil {
				return fmt.Errorf("erro ao remover chave %s: %w", key, err)
			}
		}
		return nil
	})
}

// ClearAll limpa todas as chaves do banco de dados.
func (db *DB) ClearAll() error {
	return db.Update(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			if err := txn.Delete(it.Item().Key()); err != nil {
				return fmt.Errorf("erro ao limpar chave: %w", err)
			}
		}
		return nil
	})
}
