package cascade

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/valkey-io/valkey-go"
)

type KVConnection struct {
	Address string
}

type KVStore struct {
	details KVConnection
	client  valkey.Client
	ctx     context.Context
}

func NewKVStore(kvConn KVConnection) (*KVStore, error) {
	client, err := valkey.NewClient(valkey.ClientOption{InitAddress: []string{kvConn.Address}})
	if err != nil {
		return nil, fmt.Errorf("Failed to create valkey client for %s : %w", kvConn.Address, err)
	}
	return &KVStore{details: kvConn, client: client, ctx: context.Background()}, nil
}

func (kv *KVStore) Put(key string, value string) error {
	err := kv.client.Do(kv.ctx, kv.client.B().Set().Key(key).Value(value).Build()).Error()
	if err != nil {
		return fmt.Errorf("Failed to put %s:%s : %w", key, value, err)
	}
	return nil
}

func (kv *KVStore) Get(key string) (string, error) {
	value, err := kv.client.DoCache(kv.ctx, kv.client.B().Get().Key(key).Cache(), time.Minute).ToString()
	if err != nil {
		if valkey.IsValkeyNil(err) {
			return "", fmt.Errorf("key not found: %s", key)
		}
		return "", fmt.Errorf("Failed to get %s : %w", key, err)
	}
	return value, nil
}

func (kv *KVStore) Delete(key string) error {
	err := kv.client.Do(kv.ctx, kv.client.B().Del().Key(key).Build()).Error()
	if err != nil {
		return fmt.Errorf("Failed to delete key:%s : %w", key, err)
	}
	return nil
}

func (kv *KVStore) Exists(key string) (bool, error) {
	exists, err := kv.client.Do(kv.ctx, kv.client.B().Exists().Key(key).Build()).ToBool()
	if err != nil {
		return false, fmt.Errorf("Failed to check existence of key:%s : %w", key, err)
	}
	return exists, nil
}

func (kv *KVStore) Close() {
	kv.client.Close()
	log.Println("KV store connection closed.")
}
