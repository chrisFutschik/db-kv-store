package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"

	"github.com/gin-gonic/gin"
)

type KeyValue struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type KVStore struct {
	mu       sync.RWMutex
	data     map[string]string
	filePath string
}

func (s *KVStore) LoadDiskBackedData() {

	j, err := os.Open(s.filePath)
	if err != nil {
		fmt.Println(err)
	}

	defer j.Close()

	byteValue, err := io.ReadAll(j)
	if err != nil {
		fmt.Println(err)
	}

	json.Unmarshal(byteValue, &s.data)

}

func (s *KVStore) WriteToDisk() {
	d, err := json.Marshal(s.data)
	if err != nil {
		fmt.Println(err)
	}

	err = os.WriteFile(s.filePath, d, 0644) // -rw-r--r--
	if err != nil {
		fmt.Println(err)
	}
}

func CreateKVStore(fp string) *KVStore {

	kv := &KVStore{
		data:     make(map[string]string),
		filePath: fp,
	}

	kv.LoadDiskBackedData()

	return kv
}

func (s *KVStore) GetAll() map[string]string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	x := make(map[string]string)
	for k, v := range s.data {
		x[k] = v
	}

	return x
}

func GetAllKeyValues(c *gin.Context) {

	c.IndentedJSON(http.StatusOK, keyValueStore.GetAll())
}

func ValidKeyInput(k string) bool {

	//Check for Valid ASCII a-z,A-Z,0-9
	for _, r := range k {
		if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') && (r < '0' || r > '9') {
			return false
		}
	}
	return true
}

func KeyOver255Bytes(s string) bool {
	return len(s) > 255
}

func KeyEmpty(s string) bool {
	return len(s) <= 0
}

func ValueOver1024Bytes(s string) bool {

	return len(s) > 1024

}

func (s *KVStore) Put(key string, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data[key] = value
}

func PutKey(c *gin.Context) {

	var kv KeyValue

	if err := c.ShouldBindJSON(&kv); err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": "key not found"})
		return
	}

	if !ValidKeyInput(kv.Key) {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "key can only contain ascii 0-9,a-z,A-Z"})
		return
	}

	if KeyOver255Bytes(kv.Key) {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "key over 255 bytes"})
		return
	}

	if KeyEmpty(kv.Key) {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "key empty"})
		return
	}

	if ValueOver1024Bytes(kv.Value) {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "value over 1024 bytes"})
		return
	}

	keyValueStore.Put(kv.Key, kv.Value)
	keyValueStore.WriteToDisk()
	c.IndentedJSON(http.StatusCreated, kv)

}

func (s *KVStore) Get(k string) (string, bool) {

	s.mu.RLock()
	defer s.mu.RUnlock()

	value, ok := s.data[k]

	return value, ok
}

func GetKeyValue(c *gin.Context) {

	key := c.Param("key")

	if value, ok := keyValueStore.Get(key); ok {
		c.IndentedJSON(http.StatusOK, value)
		return
	}

	c.IndentedJSON(http.StatusNotFound, gin.H{"error": " key not found"})

}

func (s *KVStore) Delete(k string) {

	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.data, k)
}

func DeleteKey(c *gin.Context) {

	key := c.Param("key")

	if _, ok := keyValueStore.data[key]; ok {
		keyValueStore.Delete(key)
		keyValueStore.WriteToDisk()
		return
	}

	c.IndentedJSON(http.StatusNotFound, gin.H{"error": " key not found"})
}

var keyValueStore *KVStore

func main() {

	keyValueStore = CreateKVStore("kvBackUp.json")

	router := gin.Default()

	router.GET("/", GetAllKeyValues)
	router.GET("/:key", GetKeyValue)
	router.POST("/", PutKey)
	router.DELETE("/:key", DeleteKey)

	router.Run("localhost:3000")
}
