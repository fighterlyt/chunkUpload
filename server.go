package chunk_upload

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/davecgh/go-spew/spew"
)



func uploadFileStart(w http.ResponseWriter, r *http.Request) {
	log.Println("接收到开始信息")

	defer r.Body.Close()

	req := &Request{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		WriteResult(w, "", err.Error(), err.Error())
	} else {
		spew.Print("接收到开始信息",)

		id := fmt.Sprintf("%d", time.Now().UnixNano())
		task := newUploadTask(id, req.Md5)
		tasks.Add(task)
		WriteResult(w, id, "ok", "")
	}
}

func uploadFile(w http.ResponseWriter, r *http.Request) {

	defer r.Body.Close()

	req := &Request{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		WriteResult(w, req.Session_id, err.Error(), err.Error())
	} else {
		spew.Println("接收到块信息", len(req.Data), calMd5Bytes(req.Data), req.Md5)
		tasks.AddChunk(req.Session_id, req.Md5, req.Data)
		WriteResult(w, req.Session_id, "ok", "")
	}
}

func uploadFileFinish(w http.ResponseWriter, r *http.Request) {
	req := &Request{}
	spew.Print("接收到结束信息")

	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Write([]byte(err.Error()))
		w.WriteHeader(400)
	} else {
		spew.Print("接收到结束信息", req.Session_id)

		if closer, err := tasks.Finish(req.Session_id); err != nil {
			WriteResult(w, req.Session_id, err.Error(), err.Error())
		} else {
			closer.Close()
			WriteResult(w, req.Session_id, "ok", "")
		}
	}
}

func WriteResult(w http.ResponseWriter, id, result, detail string) {
	resp := Result{
		Id:     id,
		Result: result,
		Detail: detail,
	}
	data, _ := json.Marshal(resp)
	w.Write(data)
}

type Result struct {
	Id     string `json:"id"`
	Result string `json:"result"`
	Detail string `json:"detail`
}
