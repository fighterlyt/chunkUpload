package chunk_upload

import (
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"time"

	"crypto/md5"
	"hash"

	"io"

	"github.com/pkg/errors"
	"log"
)

var (
	tasks               = &UploadTasks{}
	defaultLimit uint64 = 1000 * 1000 * 10 //10MB
)

type Data struct {
	writedCount uint64
	file        *os.File
	mem         []byte
	limit       uint64
	md5         hash.Hash
}

//Write 实现了io.Writer 接口
func (d *Data) Write(data []byte) (int, error) {
	if d.writedCount+uint64(len(data)) > d.limit {
		if d.file == nil { //如果尚未开始写入临时文件
			// 创建临时文件
			if file, err := ioutil.TempFile("", "upload_data"); err != nil {
				return 0, errors.Wrap(err, "写入数据")
			} else {
				//将内存数据写入文件
				if _, err = file.Write(d.mem); err != nil {
					return 0, errors.Wrap(err, "写入数据")
				}
				d.mem = d.mem[:0]
				d.file = file
			}
		}
		if count, err := d.file.Write(data); err != nil {
			return 0, errors.Wrap(err, "写入数据")
		} else {
			d.md5.Write(data)
			return count, err
		}
	} else {
		d.md5.Write(data)
		d.mem = append(d.mem, data...)
		return len(data), nil
	}
}

func (d *Data) Close() error {
	log.Println(d.file,d.file.Name())
	if d.file != nil {
		if err := d.file.Close(); err != nil {
			return err
		} else {
			return os.Remove(d.file.Name())
		}
	}
	return nil
}
func (d *Data) Read(data []byte) (int, error) {
	if d.file != nil {
		return d.file.Read(data)
	} else {
		if d.writedCount > uint64(len(data)) {
			copy(data, d.mem[:len(data)])
			return len(data), nil
		} else {
			copy(data, d.mem)
			return len(data), io.EOF
		}
	}
}
//Md5 返回已经上传数据的 md5值
func (d Data) Md5() string {
	return fmt.Sprintf("%x", d.md5.Sum(nil))
}
//NewData 创建一个新的*Data,如果 limit==0,那么使用默认大小
func NewData(limit uint64) *Data {
	if limit == 0 {
		limit = defaultLimit
	}
	return &Data{
		mem: make([]byte, 0, limit),
		md5: md5.New(),
	}
}
//UploadTask表示一个文件的多次上传
type UploadTask struct {
	id        string
	data      *Data
	startTime time.Time
	md5       string
}

func newUploadTask(id string, md5 string) *UploadTask {
	return &UploadTask{
		id:        id,
		data:      NewData(0),
		startTime: time.Now(),
		md5:       md5,
	}
}
/*  Append 添加一个文件块*/
func (u *UploadTask) Append(data []byte, md5 string) error {
	if realMd5 := calMd5Bytes(data); realMd5 == md5 {
		_, err := u.data.Write(data)
		return err
	} else {
		return newError(ErrDisMatch, md5+"/"+realMd5)
	}
}

//Finish 完成一个文件上传，返回io.ReadCloser用于访问文件
func (u *UploadTask) Finish() (io.ReadCloser,error) {
	if realMd5 := u.data.Md5(); realMd5 == u.md5 {
		return u.data, nil
	} else {
		return nil, newError(ErrDisMatch, u.md5+"/"+realMd5)
	}
}

type UploadTasks struct {
	tasks map[string]*UploadTask
	sync.Mutex
}

//Add 添加一个上传任务
func (u *UploadTasks) Add(task *UploadTask) {
	u.Lock()
	defer u.Unlock()

	if u.tasks == nil {
		u.tasks = make(map[string]*UploadTask, 10)
	}
	u.tasks[task.id] = task
}

// AddChunk 指定任务添加一个上传块
func (u *UploadTasks) AddChunk(id, md5 string, data []byte) error {
	u.Lock()
	task := u.tasks[id]
	u.Unlock()
	if task == nil {
		return newError(ErrNotFound, id)
	} else {
		return task.Append(data, md5)
	}
}

// Finish 完成一个上传任务
func (u *UploadTasks) Finish(id string) (io.ReadCloser, error) {
	u.Lock()
	task := u.tasks[id]
	u.Unlock()
	if readCloser, err := task.Finish(); err != nil {
		return nil, err
	} else {
		return readCloser, nil
	}

}
