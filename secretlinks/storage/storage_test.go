package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateUnique(t *testing.T) {
	memoryStorage := NewMemoryStorage()
	unique := memoryStorage.Create("key", Link{}, true)
	unique2 := memoryStorage.Create("key2", Link{}, true)

	assert.Equal(t, true, unique)
	assert.Equal(t, true, unique2)
	assert.Equal(t, 2, len(memoryStorage.links))
}

func TestCreateNotUnique(t *testing.T) {
	memoryStorage := NewMemoryStorage()
	unique := memoryStorage.Create("key", Link{}, true)
	unique2 := memoryStorage.Create("key", Link{}, true)

	assert.Equal(t, true, unique)
	assert.Equal(t, false, unique2)
	assert.Equal(t, 1, len(memoryStorage.links))
}

func TestUpdate(t *testing.T) {
	memoryStorage := NewMemoryStorage()
	memoryStorage.Create("key", Link{Secret: "1"}, true)

	memoryStorage.Update("key", Link{Secret: "2"})

	assert.Equal(t, "2", memoryStorage.links["key"].Secret)
}

func TestGetExist(t *testing.T) {
	memoryStorage := NewMemoryStorage()
	memoryStorage.Create("key", Link{Secret: "1"}, true)

	link, exist := memoryStorage.Get("key")

	assert.Equal(t, Link{Secret: "1"}, link)
	assert.Equal(t, true, exist)
}

func TestGetNotExist(t *testing.T) {
	memoryStorage := NewMemoryStorage()
	memoryStorage.Create("key", Link{Secret: "1"}, true)

	link, exist := memoryStorage.Get("key2")

	assert.Equal(t, Link{}, link)
	assert.Equal(t, false, exist)
}

func TestDelete(t *testing.T) {
	memoryStorage := NewMemoryStorage()
	memoryStorage.Create("key", Link{Secret: "1"}, true)

	memoryStorage.Delete("key")

	assert.Equal(t, 0, len(memoryStorage.links))

}
