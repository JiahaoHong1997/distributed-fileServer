package handler

import (
	"distributed-fileServer/meta"
	"distributed-fileServer/util"
	"strconv"

	"encoding/json"
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
		// 接收文件流及存储到目录，选取本地文件，Form形式上传文件
		file, head, err := r.FormFile("file") // 这里的"file"来自index.html文件
		if err != nil {
			fmt.Printf("Failed to get data,err:%s\n", err.Error())
			return
		}
		defer file.Close()

		// 保存上传文件的元信息
		// TODO: 将文件元信息保存在mysql或redis中，避免因为保存在内存中发生丢失
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

		fileMeta.FileSize, err = io.Copy(newFile, file) // io.Copy返回文件的字节数，将文件复制到新建的空文件中
		if err != nil {
			fmt.Printf("Failed to save data into file,err:%s\n", err)
			return
		}

		newFile.Seek(0, 0)                         // 将当前已打开的文件句柄的游标移到文件内容的顶部
		fileMeta.FileSha1 = util.FileSha1(newFile) // 生成该文件的唯一识别符
		//meta.UpdateFileMeta(fileMeta)              // 更新该文件的元信息(唯一识别符)
		meta.UpdateFileMetaDB(fileMeta)

		http.Redirect(w, r, "/file/upload/suc", http.StatusFound)
	}
}

// UploadSucHandler: 上传已完成
func UploadSucHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Upload finished!")
}

// GetFileMetaHandler: 获取文件元信息
func GetFileMetaHandler(w http.ResponseWriter, r *http.Request) {

	// 通过解析url的Form表单获取文件的hash值
	r.ParseForm()

	// Form返回的是一个url.Values的结构，其实是hash，hash中每个value是一个string的slice，使用当我们使用r.Form[“filehash”]获得的其实是一个slice
	filehash := r.Form["filehash"][0]
	//fMeta := meta.GetFileMeta(filehash)
	fMeta, err := meta.GetFileMetaDB(filehash) // 从mysql中获取文件元信息
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	data, err := json.Marshal(fMeta)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

// FileQueryHandler: 查询批量的文件元信息
func FileQueryHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	limitCnt, _ := strconv.Atoi(r.Form.Get("limit")) // 根据URL中的表单信息获取要查询的文件元信息数量
	fileMetas := meta.GetLastFileMetas(limitCnt)
	data, err := json.Marshal(fileMetas)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

// DownloadHandler: 文件下载
func DownloadHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	fsha1 := r.Form.Get("filehash")
	fm := meta.GetFileMeta(fsha1) // 得到文件的元信息

	f, err := os.Open(fm.Location)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f) // 针对小文件，将其全部读入到内存中去
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octect-stream")
	w.Header().Set("Content-Disposition", "attachment;filename=\""+fm.FileName+"\"")
	w.Write(data)
}

// FileMetaUpdateHandler: 更新文件元信息(重命名)
func FileMetaUpdateHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	opType := r.Form.Get("op") // 是否支持重命名
	fileSha1 := r.Form.Get("filehash")
	newFileName := r.Form.Get("filename")

	if opType != "0" { // opType在等于0的情况下才支持重命名
		w.WriteHeader(http.StatusForbidden)
		return
	}
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	curFileMeta := meta.GetFileMeta(fileSha1) // 获取要修改文件的文件元信息
	curFileMeta.FileName = newFileName        // 修改文件名
	meta.UpdateFileMeta(curFileMeta)          // 更新文件元信息

	w.WriteHeader(http.StatusOK)
	data, err := json.Marshal(curFileMeta)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

// FileDeletHandler: 删除文件及文件元信息
func FileDeleteHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	fileSha1 := r.Form.Get("filehash")

	fMeta := meta.GetFileMeta(fileSha1)
	os.Remove(fMeta.Location) // 删除文件

	meta.RemoveFileMeta(fileSha1) // 删除文件元信息

	w.WriteHeader(http.StatusOK)
}
