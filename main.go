package main

import (
	"distributed-fileServer/handler"
	"fmt"
	"net/http"
)

func main() {

	// 静态资源处理
	http.Handle("/static/",
		http.StripPrefix("/static/",
			http.FileServer(http.Dir("./static"))))

	http.HandleFunc("/file/upload", handler.UploadHandler)        // 处理上传文件
	http.HandleFunc("/file/upload/suc", handler.UploadSucHandler) // 上传完成
	http.HandleFunc("/file/meta", handler.GetFileMetaHandler)     // 获取上传文件元信息
	http.HandleFunc("/file/query", handler.FileQueryHandler)      // 批量获取上传文件的元信息
	http.HandleFunc("/file/download", handler.DownloadHandler)    // 文件下载

	// 监听端口
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Printf("Fail to start server, err:%s\n", err.Error())
	}
}
