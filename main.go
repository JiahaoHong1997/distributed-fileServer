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

	// 动态接口路由设置
	http.HandleFunc("/file/upload", handler.HTTPInterceptor(handler.UploadHandler))         // 处理上传文件
	http.HandleFunc("/file/upload/suc", handler.HTTPInterceptor(handler.UploadSucHandler))  // 上传完成
	http.HandleFunc("/file/meta", handler.HTTPInterceptor(handler.GetFileMetaHandler))      // 获取上传文件元信息
	http.HandleFunc("/file/query", handler.HTTPInterceptor(handler.FileQueryHandler))       // 批量获取上传文件的元信息
	http.HandleFunc("/file/download", handler.HTTPInterceptor(handler.DownloadHandler))     // 文件下载
	http.HandleFunc("/file/update", handler.HTTPInterceptor(handler.FileMetaUpdateHandler)) // 更新文件元信息(重命名)
	http.HandleFunc("/file/delete", handler.HTTPInterceptor(handler.FileDeleteHandler))     // 删除文件以及文件元信息

	// 秒传接口
	http.HandleFunc("/file/fastupload", handler.HTTPInterceptor(handler.TryFastUploadHandler)) // 尝试秒传的接口

	// 分块上传接口
	http.HandleFunc("/file/mpupload/init", handler.HTTPInterceptor(handler.InitialMultipartUploadHandler)) // 初始化分块上传
	http.HandleFunc("/file/mpupload/uppart", handler.HTTPInterceptor(handler.UploadPartHandler))           // 上传文件分块
	http.HandleFunc("/file/mpupload/complete", handler.HTTPInterceptor(handler.CompleteUploadHandler))     // 通知完成上传接口
	http.HandleFunc("/file/mpload/cancel", handler.HTTPInterceptor(handler.CancelUploadHandler))           // 取消文件上传接口

	// 用户相关接口
	http.HandleFunc("/user/signup", handler.SignUpHandler)                          // 用户注册
	http.HandleFunc("/user/signin", handler.SignInHandler)                          // 用户登录
	http.HandleFunc("/user/info", handler.HTTPInterceptor(handler.UserInfoHandler)) // 用户信息查询

	// 监听端口
	fmt.Println("上传服务正在启动, 监听端口:8080...")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Printf("Fail to start server, err:%s\n", err.Error())
	}
}
