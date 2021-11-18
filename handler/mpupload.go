package handler

import (
	"distributed-fileServer/util"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"

	rPool "distributed-fileServer/cache/redis"
)

// MutiPartUploadInfo:初始化信息
type MutiPartUploadInfo struct {
	FileHash	string
	FileSize	int
	UploadID	string
	ChunkSize	int
	ChunkCount	int
}

// 初始化分块上传
func InitialMultipartUploadHandler(w http.ResponseWriter, r *http.Request) {
	// 1.解析用户请求参数
	r.ParseForm()
	userName := r.Form.Get("username")
	fileHash := r.Form.Get("filehash")
	fileSize, err := strconv.Atoi(r.Form.Get("filesize"))
	if err != nil {
		w.Write(util.NewRespMsg(-1, "params invalid", nil).JSONBytes())
		return
	}

	// 2.获得Redis的一个连接
	rConn := rPool.RedisPool().Get()
	defer rConn.Close()

	// 3.生成分块上传的初始化信息
	upInfo := MutiPartUploadInfo{
		FileHash: fileHash,
		FileSize: fileSize,
		UploadID: userName + fmt.Sprintf("%x", time.Now().UnixNano()),
		ChunkSize: 5*1024*1024,	// 5MB
		ChunkCount: int(math.Ceil(float64(fileSize)/(5*1024*1024))),
	}

	// 4.将初始化信息写入Redis缓存
	rConn.Do("HSET", "MP_"+upInfo.UploadID,"chunkcount",upInfo.ChunkCount)
	rConn.Do("HSET", "MP_"+upInfo.UploadID,"filehash",upInfo.FileHash)
	rConn.Do("HSET", "MP_"+upInfo.UploadID,"filesize",upInfo.FileSize)

	// 5.将响应初始化数据返回到客户端
	w.Write(util.NewRespMsg(0,"OK",upInfo).JSONBytes())
}