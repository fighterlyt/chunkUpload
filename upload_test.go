package chunk_upload

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClientUpload(t *testing.T) {
	err := clientUpload("/Users/apple/Downloads/githublisten1-listen1_desktop_mac-v1.2.0.zip", 10000)
	assert.NoError(t, err)
}

func TestMain(m *testing.M) {
	go startServer()
	m.Run()
}
