package handler

import (
	dblayer "distributed-fileServer/db"
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
		_ = meta.UpdateFileMetaDB(fileMeta)

		// 更新用户文件表记录
		r.ParseForm()
		userName := r.Form.Get("username")
		suc := dblayer.OnUserFileUploadFinished(userName, fileMeta.FileSha1,fileMeta.FileName,fileMeta.FileSize)
		if suc {
			http.Redirect(w, r, "/static/view/home.html", http.StatusFound)  // 上传完成后直接跳转完成页面
		} else {
			w.Write([]byte("Upload Failed."))
		}
	}
}

// TryFastUploadHandler：尝试秒传的接口
func TryFastUploadHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	// 1.解析请求参数
	userName := r.Form.Get("username")
	fileHash := r.Form.Get("filehash")
	fileName := r.Form.Get("filename")
	filesize,_ := strconv.Atoi(r.Form.Get("filesize"))

	// 2.从文件表中查询相同hash的文件记录
	fileMeta,err := meta.GetFileMetaDB(fileHash)
	if err != nil {
		fmt.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// 3.查不到记录则返回秒传失败
	if fileMeta.FileSha1 == "" {
		resp := util.RespMsg{
			Code: -1,
			Msg:"秒传失败，请访问普通上传接口",
		}
		w.Write(resp.JSONBytes())
		return
	}

	// 4.上传过则将文件信息写入用户文件表，返回成功
	suc := dblayer.OnUserFileUploadFinished(userName,fileHash,fileName,int64(filesize))
	if suc {
		resp := util.RespMsg{
			Code: 0,
			Msg:"秒传成功",
		}
		w.Write(resp.JSONBytes())
		return
	} else {
		resp := util.RespMsg{
			Code: -2,
			Msg:"秒传失败，请稍后重试",
		}
		w.Write(resp.JSONBytes())
		return
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
	//userFile, err := meta.GetLastFileMetasDB(limitCnt)

	userName := r.Form.Get("username")
	userFile,err := dblayer.QueryUserFileMetas(userName,limitCnt)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// 封装为json格式返回给客户端
	data, err := json.Marshal(userFile)
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
	//fm := meta.GetFileMeta(fsha1) // 得到文件的元信息
	fm, err := meta.GetFileMetaDB(fsha1)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
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

	//curFileMeta := meta.GetFileMeta(fileSha1) // 获取要修改文件的文件元信息
	curFileMeta, err := meta.GetFileMetaDB(fileSha1)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	curFileMeta.FileName = newFileName        // 修改文件名
	suc := meta.UpdateFileMetaDB(curFileMeta)          // 更新文件元信息
	if suc != true {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

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

	fMeta,err := meta.GetFileMetaDB(fileSha1)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	os.Remove(fMeta.Location) // 删除文件

	suc := meta.OnFileRemovedDB(fileSha1) // 删除文件元信息
	if !suc {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
