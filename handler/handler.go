package handler

import (
	"distributed-fileServer/meta"
	"distributed-fileServer/util"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

// UploadHandler: 处理文件上传
func UploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		// 返回上传html页面
		data, err := ioutil.ReadFile("./static/view/index.html")
		if err != nil {
			io.WriteString(w, "internel server eeror")
			return
		}
		io.WriteString(w, string(data))
	} else if r.Method == "POST" {
		// 接收文件流及存储到目录
		file, head, err := r.FormFile("file") // 这里的"file"来自index.html文件
		if err != nil {
			fmt.Printf("Failed to get data,err:%s\n", err.Error())
			return
		}
		defer file.Close()

		// 保存上传文件的元信息
		fileMeta := meta.FileMeta{
			FileName: head.Filename,
			Location: "/tmp/" + head.Filename,
			UploadAt: time.Now().Format("2006-01-02 15:04:05"),
		}

		newFile, err := os.Create(fileMeta.Location) // 根据路径创建一个空文件，如果该路径下存在这个文件，则重置它。可用于I/O
		if err != nil {
			fmt.Printf("Failed to create file,err:%s\n", err)
			return
		}
		defer newFile.Close()

		fileMeta.FileSize, err = io.Copy(newFile, file) // io.Copy返回文件的字节数
		if err != nil {
			fmt.Printf("Failed to save data into file,err:%s\n", err)
			return
		}

		newFile.Seek(0, 0)
		fileMeta.FileSha1 = util.FileSha1(newFile) // 生成该文件的唯一识别符
		meta.UpdateFileMeta(fileMeta)              // 更新该文件的元信息(唯一识别符)

		http.Redirect(w, r, "/file/upload/suc", http.StatusFound)
	}
}

// UploadSucHandler: 上传已完成
func UploadSucHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Upload finished!")
}
