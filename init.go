package chunk_upload

import (
	"crypto/md5"
	"fmt"
	"io"
	"net/http"

	"github.com/rs/cors"
)

func startServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/upload", uploadFile)
	mux.HandleFunc("/uploadStart", uploadFileStart)
	mux.HandleFunc("/uploadFinish", uploadFileFinish)
	c := cors.AllowAll()
	handler := c.Handler(mux)
	http.ListenAndServe(":3000", handler)

}
func calMd5(reader io.Reader) (string, error) {
	hash := md5.New()
	if _, err := io.Copy(hash, reader); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}
func calMd5Bytes(data []byte) string {

	return fmt.Sprintf("%x", md5.Sum(data))

}

