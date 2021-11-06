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

	http.HandleFunc("/file/upload", handler.UploadHandler)         // 处理上传文件
	http.HandleFunc("/file/upload/suc", handler.UploadSucHandler)  // 上传完成
	http.HandleFunc("/file/meta", handler.GetFileMetaHandler)      // 获取上传文件元信息
	http.HandleFunc("/file/query", handler.FileQueryHandler)       // 批量获取上传文件的元信息
	http.HandleFunc("/file/download", handler.DownloadHandler)     // 文件下载
	http.HandleFunc("/file/update", handler.FileMetaUpdateHandler) // 更新文件元信息(重命名)
	http.HandleFunc("/file/delete", handler.FileDeleteHandler)     // 删除文件以及文件元信息

	http.HandleFunc("/user/signup", handler.SignUpHandler) // 用户注册
	http.HandleFunc("/user/signin", handler.SignInHandler) // 用户登录
	http.HandleFunc("/user/info", handler.HTTPInterceptor(handler.UserInfoHandler)) // 用户信息查询

	// 监听端口
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Printf("Fail to start server, err:%s\n", err.Error())
	}
}
