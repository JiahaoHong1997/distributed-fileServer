package db

import (
	mydb "distributed-fileServer/db/mysql"
	"fmt"
)

// UserSignUp：通过用户名及密码完成user表的注册操作
func UserSignUp(userName string, passwd string) bool {
	stmt, err := mydb.DBConn().Prepare("insert ignore into tbl_user(`user_name`,`user_pwd`)values(?,?)")
	if err != nil {
		fmt.Println("Failed to insert,err:", err.Error())
		return false
	}
	defer stmt.Close()

	ret, err := stmt.Exec(userName, passwd)
	if err != nil {
		fmt.Println("Failed to insert,err:", err.Error())
		return false
	}

	if rowAffected, err := ret.RowsAffected(); err == nil && rowAffected > 0 { // 判断是否重复注册
		return true
	}
	return false
}
