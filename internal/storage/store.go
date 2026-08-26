// Package storage provides the application's persistent key-value store.
package storage

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/dgraph-io/badger/v4"
)

// Store wraps Badger for backend services. It is deliberately not registered
// as a Wails service.
type Store struct {
	db *badger.DB
}

func OpenDefault() (*Store, error) {
	path := "videodown.db"
	if executable, err := os.Executable(); err == nil {
		path = filepath.Join(filepath.Dir(executable), path)
	}
	return Open(path)
}

func Open(path string) (*Store, error) {
	db, err := badger.Open(badger.DefaultOptions(path).WithLoggingLevel(badger.ERROR))
	if err != nil {
		return nil, err
	}
	return &Store{db: db}, nil
}

func OpenMemory() (*Store, error) {
	db, err := badger.Open(badger.DefaultOptions("").WithInMemory(true))
	if err != nil {
		return nil, err
	}
	return &Store{db: db}, nil
}

func (s *Store) Get(key string) (string, error) {
	var value string
	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		return item.Value(func(valueBytes []byte) error {
			value = string(valueBytes)
			return nil
		})
	})
	return value, err
}

func (s *Store) Set(key, value string) error {
	return s.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), []byte(value))
	})
}

func (s *Store) Delete(key string) error {
	return s.db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(key))
	})
}

func (s *Store) DeletePrefix(prefix string) error {
	prefixBytes := []byte(prefix)
	return s.db.Update(func(txn *badger.Txn) error {
		keys := make([][]byte, 0)
		iterator := txn.NewIterator(badger.DefaultIteratorOptions)
		defer iterator.Close()

		for iterator.Seek(prefixBytes); iterator.ValidForPrefix(prefixBytes); iterator.Next() {
			keys = append(keys, iterator.Item().KeyCopy(nil))
		}
		for _, key := range keys {
			if err := txn.Delete(key); err != nil && !errors.Is(err, badger.ErrKeyNotFound) {
				return fmt.Errorf("delete key %q: %w", key, err)
			}
		}
		return nil
	})
}

func (s *Store) Update(fn func(txn *badger.Txn) error) error {
	return s.db.Update(fn)
}

func (s *Store) View(fn func(txn *badger.Txn) error) error {
	return s.db.View(fn)
}

func (s *Store) Close() error {
	return s.db.Close()
}
