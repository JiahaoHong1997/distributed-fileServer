package db

import (
	"database/sql"
	mydb "distributed-fileServer/db/mysql"
	"fmt"
)

// OnFileUploadFinished：文件上传完成，保存meta
func OnFileUploadFinished(fileHash string, fileName string, fileSize int64, fileAddr string) bool {
	stmt, err := mydb.DBConn().Prepare(
		"insert ignore into tbl_file(`file_sha1`,`file_name`,`file_size`," +
			"`file_addr`,`status`) values (?,?,?,?,1)",
	)
	if err != nil {
		fmt.Println("Failed to prepare statement,err:" + err.Error())
		return false
	}
	defer stmt.Close()

	ret, err := stmt.Exec(fileHash, fileName, fileSize, fileAddr)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	if rf, err := ret.RowsAffected(); err == nil {
		if rf <= 0 {
			fmt.Printf("File with hash:%s has been uploaded before", fileHash)
		}
		return true
	}
	return false
}

type TableFile struct {
	FileHash string
	FileName sql.NullString
	FileSize sql.NullInt64
	FileAddr sql.NullString
}

// GetFileMeta：从mysql获取文件元信息
func GetFileMeta(fileHash string) (*TableFile, error) {
	stmt, err := mydb.DBConn().Prepare(
		"select file_sha1,file_addr,file_name,file_size from tbl_file " +
			"where file_sha1=? and status=1 limit 1",
	)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	defer stmt.Close()

	tfile := TableFile{}
	err = stmt.QueryRow(fileHash).Scan(&tfile.FileHash, &tfile.FileAddr, &tfile.FileName, &tfile.FileSize)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	return &tfile, nil
}

// IsFileUploaded : 文件是否已经上传过
func IsFileUploaded(fileHash string) bool {
	stmt, err := mydb.DBConn().Prepare("select 1 from tbl_file where file_sha1=? and status=1 limit 1")
	rows, err := stmt.Query(fileHash)
	if err != nil {
		return false
	} else if rows == nil || !rows.Next() {
		return false
	}
	return true
}

// GetFileMetaList : 从mysql批量获取文件元信息
func GetFileMetaList(limit int) ([]TableFile,error) {
	stmt, err := mydb.DBConn().Prepare(
		"select file_sha1,file_addr,file_name,file_size from tbl_file " +
			"where status=1 limit ?")
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	defer stmt.Close()

	rows,err := stmt.Query(limit)
	if err != nil {
		fmt.Println(err.Error())
		return nil,err
	}

	var tfiles []TableFile
	for rows.Next() {
		tfile := TableFile{}
		err := rows.Scan(&tfile.FileHash,&tfile.FileAddr,&tfile.FileName,&tfile.FileSize)
		if err != nil {
			fmt.Println(err.Error())
			break
		}
		tfiles = append(tfiles,tfile)
	}
	fmt.Println(len(tfiles))
	return tfiles,nil
}


// OnFileRemoved : 文件删除(这里只做标记删除，即改为status=2)
func OnFileRemoved(fileHash string) bool {
	stmt, err := mydb.DBConn().Prepare(
		"update tbl_file set status=2 where file_sha1=? and status=1 limit 1")
	if err != nil {
		fmt.Println("Failed to prepare statement, err:" + err.Error())
		return false
	}
	defer stmt.Close()
	ret, err := stmt.Exec(fileHash)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	if rf, err := ret.RowsAffected(); nil == err {
		if rf <= 0 {
			fmt.Printf("File with hash:%s not uploaded", fileHash)
		}
		return true
	}
	return false
}