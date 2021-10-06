package meta

import (
	mydb "distributed-fileServer/db"
	"sort"
)

// FileMeta: 文件元信息结构
type FileMeta struct {
	FileSha1 string // 文件唯一标识符
	FileName string
	FileSize int64
	Location string
	UploadAt string // 文件上传的时间
}

var fileMetas map[string]FileMeta

func init() {
	fileMetas = make(map[string]FileMeta)
}

// UpdateFileMeta: 更新文件元信息
func UpdateFileMeta(fMeta FileMeta) {
	fileMetas[fMeta.FileSha1] = fMeta
}

// UpdateFileMetaDB：新增/更新文件元信息到mysql中
func UpdateFileMetaDB(fMeta FileMeta) bool {
	return mydb.OnFileUploadFinished(fMeta.FileSha1, fMeta.FileName, fMeta.FileSize, fMeta.Location)
}

// GetFileMeta: 通过Sha1值获取元件的元信息对象
func GetFileMeta(fileSha1 string) FileMeta {
	return fileMetas[fileSha1]
}

// GetFileMetaDB：从mysql获取文件元信息
func GetFileMetaDB(fileSha1 string) (FileMeta, error) {
	tfile, err := mydb.GetFileMeta(fileSha1)
	if err != nil {
		return FileMeta{}, err
	}
	fMeta := FileMeta{
		FileSha1: tfile.FileHash,
		FileName: tfile.FileName.String,
		FileSize: tfile.FileSize.Int64,
		Location: tfile.FileAddr.String,
	}
	return fMeta, nil
}

// GetLastFileMetas: 获取批量的原件元信息列表
func GetLastFileMetas(count int) []FileMeta {
	fMetaArray := make([]FileMeta, 0)
	for _, v := range fileMetas {
		fMetaArray = append(fMetaArray, v)
	}

	// 将文件元信息列表按上传时间排序
	sort.Sort(ByUploadTime(fMetaArray))
	return fMetaArray[0:count]
}

// RemoveFileMeta: 删除文件元信息
func RemoveFileMeta(fileSha1 string) {
	delete(fileMetas, fileSha1)
}
