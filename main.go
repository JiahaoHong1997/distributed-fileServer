package main

import (
	"fmt"
	"net/http"
	"distributed-fileServer/handler"
)

func main() {
	http.HandleFunc("/file/upload", handler.UploadHandler)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Printf("Fail to start server, err:%s\n", err.Error())
	}
}
