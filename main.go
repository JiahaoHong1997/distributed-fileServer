package main

import (
	"fmt"
	"net/http"
	"distributed-fileServer/handler"
)

func main() {

	// 静态资源处理
	http.Handle("/static/",
		http.StripPrefix("/static/",
			http.FileServer(http.Dir("./static"))))

	http.HandleFunc("/file/upload", handler.UploadHandler)  // 处理上传文件
	http.HandleFunc("/file/upload/suc", handler.UploadSucHandler) // 上传完成
	// 监听端口
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Printf("Fail to start server, err:%s\n", err.Error())
	}
}
