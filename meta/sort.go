package meta

import "time"

const baseFormate = "2006-01-02 15:04:05"

type ByUploadTime []FileMeta

func (a ByUploadTime) Len() int {
	return len(a)
}

func (a ByUploadTime) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a ByUploadTime) Less(i, j int) bool {
	iTime, _ := time.Parse(baseFormate, a[i].UploadAt)
	jTime, _ := time.Parse(baseFormate, a[j].UploadAt)
	return iTime.UnixNano() > jTime.UnixNano()
}
