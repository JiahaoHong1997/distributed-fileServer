package meta

import "sort"

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

// GetFileMeta: 通过Sha1值获取元件的元信息对象
func GetFileMeta(fileSha1 string) FileMeta {
	return fileMetas[fileSha1]
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
