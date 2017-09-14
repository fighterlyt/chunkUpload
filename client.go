package chunk_upload

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/pkg/errors"
)

type Kind int

const (
	start Kind = iota
	upload
	finish

	startUrl  = "http://localhost:3000/uploadStart"
	uploadUrl = "http://localhost:3000/upload"
	finishUrl = "http://localhost:3000/uploadFinish"
)

func clientUpload(name string, chunksize int64) error {
	if file, err := os.Open(name); err != nil {
		return errors.Wrap(err, "打开文件错误")
	} else {
		hash := ""
		log.Println("计算md5")
		if hash, err = calMd5(file); err != nil {
			return errors.Wrap(err, "计算md5")
		}
		file.Seek(0, 0)
		if info, err := file.Stat(); err != nil {
			return errors.Wrap(err, "读取文件信息")
		} else {
			chunks := info.Size() / chunksize
			if info.Size()%chunksize != 0 {
				chunks++
			}
			var id string
			log.Println("开始")

			if id, err = uploadStart(name, startUrl, hash, uint64(info.Size())); err != nil {
				return errors.Wrap(err, "传输启动")
			}

			buf := make([]byte, chunksize)
			for i := 0; int64(i) < chunks; i++ {
				count, _ := file.Read(buf)
				log.Println("块", i+1, count, calMd5Bytes(buf[:count]))

				if err = uploadProcess(id, calMd5Bytes(buf[:count]), int64(i), buf[:count], uploadUrl); err != nil {
					return errors.Wrap(err, "传输块")
				}
			}
			return uploadFinish(id, hash, finishUrl)

		}
	}
}

//第一步
func uploadStart(name, url, md5 string, size uint64) (string, error) {
	if resp, err := makeRequest(nil, md5, "", name, size, url); err != nil {
		return "", err
	} else {
		if resp.Result != "ok" {
			return "", errors.New(resp.Result)
		} else {
			return resp.Id, nil
		}
	}
}

//第二步，上传块
func uploadProcess(id, md5 string, index int64, data []byte, url string) error {
	if resp, err := makeRequest(data, md5, id, "", 0,  url); err != nil {
		return err
	} else {
		if resp.Result != "ok" {
			return errors.New(resp.Result)
		} else {
			return nil
		}
	}
}

func uploadFinish(id, md5, url string) error {
	if resp, err := makeRequest(nil, md5, id, "", 0, url); err != nil {
		return err
	} else {
		if resp.Result != "ok" {
			return errors.New(resp.Result)
		} else {
			return nil
		}
	}
}
func makeRequest(data []byte, md5, id, fileName string, size uint64, url string) (*Response, error) {
	r := Request{
		Session_id: id,
		Data:       data,
		Md5:        md5,
		Size:       size,
		FileName:   fileName,
	}

	buffer := &bytes.Buffer{}
	json.NewEncoder(buffer).Encode(r)

	if resp, err := http.Post(url, "application/json", buffer); err != nil {
		return nil, err
	} else {
		result := &Response{}

		defer resp.Body.Close()
		return result, json.NewDecoder(resp.Body).Decode(result)
	}
}

type Response struct {
	Id     string `json:"id"`
	Result string `json:"result"`
}

type Request struct {
	Session_id string `json:"session_id"`
	Data       []byte `json:"data"`
	Md5        string `json:"md5"`
	Size       uint64 `json:"size"`
	FileName   string `json:"fileName"`
}
