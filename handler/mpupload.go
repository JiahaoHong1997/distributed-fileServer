package handler

import (
	"distributed-fileServer/util"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"math"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	rPool "distributed-fileServer/cache/redis"
	dblayer "distributed-fileServer/db"
)

const (
	// ChunkDir：上传的分块所在目录
	ChunkDir = "/data/chunks"
	// MergeDir：合并后的文件所在目录
	MergeDir = "/data/merge"
	// ChunkKeyPrefix：分块信息对应的Redis键前缀
	ChunkKeyPrefix = "MP_"
	// HashUpIDKeyPrefix：文件hash映射uploadid随营的redis键前缀
	HashUpIDKeyPrefix = "HASH_UPID_"
)

func init() {
	if err := os.MkdirAll(ChunkDir,0744);err != nil {
		fmt.Println("无法指定目录用于存储分块文件"+ChunkDir)
		os.Exit(1)
	}

	if err := os.MkdirAll(MergeDir,0744);err != nil {
		fmt.Println("无法指定目录用于存储合并后文件"+ChunkDir)
		os.Exit(1)
	}
}

// MutiPartUploadInfo:初始化信息
type MultiPartUploadInfo struct {
	FileHash    string
	FileSize    int
	UploadID    string
	ChunkSize   int
	ChunkCount  int
	ChunkExists []int // 已经上传完成的分块索引列表
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

	// 3.通过文件hash判断是否断点续传，并获取uploadid
	uploadID := ""
	keyExists,_ := redis.Bool(rConn.Do("EXISTS", HashUpIDKeyPrefix+fileHash))
	if keyExists {
		uploadID, err = redis.String(rConn.Do("GET", HashUpIDKeyPrefix+fileHash))
		if err != nil {
			w.Write(util.NewRespMsg(-1,"Upload part failed",err.Error()).JSONBytes())
			return
		}
	}

	// 4.1 首次上传则新建uploaID
	// 4.2 断点续传则根据uploadIB获取已经上传的文件分块列表
	chunksExist := []int{}
	if uploadID == "" {
		uploadID = userName + fmt.Sprintf("%x", time.Now().UnixNano())
	} else {
		chunks,err := redis.Values(rConn.Do("HGETALL", ChunkKeyPrefix+uploadID)) // 读出已经上传的分块列表
		if err != nil {
			w.Write(util.NewRespMsg(-2, "Upload part failed", err.Error()).JSONBytes())
			return
		}
		for i:=0; i<len(chunks); i+=2 {
			k := string(chunks[i].([]byte))
			v := string(chunks[i+1].([]byte))
			if strings.HasPrefix(k, "chkidx_") && v == "1" {
				chunkIdx,_ := strconv.Atoi(k[7:len(k)])
				chunksExist = append(chunksExist,chunkIdx)
			}
		}
	}

	// 5.生成分块上传的初始化信息
	upInfo := MultiPartUploadInfo{
		FileHash:   fileHash,
		FileSize:   fileSize,
		UploadID:   userName + fmt.Sprintf("%x", time.Now().UnixNano()),
		ChunkSize:  5 * 1024 * 1024, // 5MB
		ChunkCount: int(math.Ceil(float64(fileSize) / (5 * 1024 * 1024))),
		ChunkExists: chunksExist,
	}

	// 6.首次上传的话将初始化信息写入Redis缓存
	if len(upInfo.ChunkExists) <= 0 {
		hkey := "MP_"+upInfo.UploadID
		rConn.Do("HSET", hkey, "chunkcount", upInfo.ChunkCount)
		rConn.Do("HSET", hkey, "filehash", upInfo.FileHash)
		rConn.Do("HSET", hkey, "filesize", upInfo.FileSize)
		rConn.Do("EXPIRE", hkey, 43200)
		rConn.Do("SET", HashUpIDKeyPrefix+fileHash, upInfo.UploadID, "EX", 43200)
	}

	// 7.将响应初始化数据返回到客户端
	w.Write(util.NewRespMsg(0, "OK", upInfo).JSONBytes())
}

// UploadPartHandler: 上传文件分块
func UploadPartHandler(w http.ResponseWriter, r *http.Request) {
	// 1.解析用户请求参数
	r.ParseForm()
	//userName := r.Form.Get("username")
	uploadID := r.Form.Get("uploadid")
	chunkIndex := r.Form.Get("index")

	// 2.获得Redis连接池的连接
	rConn := rPool.RedisPool().Get()
	defer rConn.Close()

	// 3.获得文件句柄，用于存储分块内容
	fpath := "/data/" + uploadID + "/" + chunkIndex
	os.MkdirAll(path.Dir(fpath), 0744)
	fd, err := os.Create(fpath)
	if err != nil {
		w.Write(util.NewRespMsg(-1, "Upload part failed", nil).JSONBytes())
		return
	}
	defer fd.Close()

	buf := make([]byte, 1024*1024) // 1MB
	for {
		n, err := r.Body.Read(buf)
		fd.Write(buf[:n])
		if err != nil {
			break
		}
	}

	// 4.更新Redis缓存状态
	rConn.Do("HSET", "MP_"+uploadID, "chkidx_"+chunkIndex, 1)

	// 5.返回处理结果到客户端
	w.Write(util.NewRespMsg(0, "OK", nil).JSONBytes())
}

// CompleteUploadHandler: 通知上传合并
func CompleteUploadHandler(w http.ResponseWriter, r *http.Request) {
	// 1.解析请求参数
	r.ParseForm()
	uploadID := r.Form.Get("uploadid")
	userName := r.Form.Get("username")
	fileHash := r.Form.Get("filehash")
	filseSize := r.Form.Get("filesize")
	fileName := r.Form.Get("filename")

	// 2.获得redis连接池的一个连接
	rConn := rPool.RedisPool().Get()
	defer rConn.Close()

	// 3.通过uploadid查询redis并判断是否所有的分块完成上传
	data, err := redis.Values(rConn.Do("HGETALL", "MP_"+uploadID))
	if err != nil {
		w.Write(util.NewRespMsg(-1, "complete upload failed", nil).JSONBytes())
		return
	}
	totalCount := 0
	chunkCount := 0
	for i := 0; i < len(data); i += 2 {
		k := string(data[i].([]byte))
		v := string(data[i+1].([]byte))
		if k == "chunkcount" {
			totalCount, _ = strconv.Atoi(v)
		} else if strings.HasPrefix(k, "chkidx_") && v == "1" {
			chunkCount++
		}
	}
	if totalCount != chunkCount {
		w.Write(util.NewRespMsg(-2, "invalid request", nil).JSONBytes())
		return
	}
	// 4.TODO：合并分块
	if mergeSuc := util.MergeChuncksByShell(ChunkDir+uploadID, MergeDir+fileHash, fileHash); !mergeSuc {
		w.Write(util.NewRespMsg(-3, "Complete upload failed", nil).JSONBytes())
		return
	}

	// 5.更新唯一文件表积用户文件表
	fsize, _ := strconv.Atoi(filseSize)
	dblayer.OnFileUploadFinished(fileHash, fileName, int64(fsize), "")
	dblayer.OnUserFileUploadFinished(userName, fileHash, fileName, int64(fsize))

	// 6.响应处理结果
	w.Write(util.NewRespMsg(0, "OK", nil).JSONBytes())
}

// CancelUploadHandler：文件取消上传接口
func CancelUploadHandler(w http.ResponseWriter, r *http.Request) {
	// 1.解析用户请求参数
	r.ParseForm()
	fileHash := r.Form.Get("filehash")

	// 2.获得redis的一个连接
	rConn := rPool.RedisPool().Get()
	defer rConn.Close()

	// 3.检查uploadid是否存在，如果存在则删除
	uploadID,err := redis.String(rConn.Do("GET", HashUpIDKeyPrefix+fileHash))
	if err != nil || uploadID == "" {
		w.Write(util.NewRespMsg(-1,"Cancel upload part failed",nil).JSONBytes())
		return
	}

	_,delHashErr := rConn.Do("DEL", HashUpIDKeyPrefix+fileHash)
	_,delUploadInfoErr := rConn.Do("DEL", ChunkKeyPrefix+uploadID)
	if delHashErr != nil || delUploadInfoErr != nil {
		w.Write(util.NewRespMsg(-2,"Cancel upload part failed",nil).JSONBytes())
		return
	}

	// 4.删除已经上传的分块文件
	delChkRes := util.RemovePathByShell(ChunkDir+uploadID)
	if !delChkRes {
		fmt.Printf("Failed to delete chunks as upload canceled, uploadID:%s\n",uploadID)
	}

	// 5.响应客户端
	w.Write(util.NewRespMsg(0,"OK",nil).JSONBytes())
}