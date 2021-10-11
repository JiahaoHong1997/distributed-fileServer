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

// UserSignIn:判断密码是否一致
func UserSignIn(userName string, encpwd string) bool {
	stmt, err := mydb.DBConn().Prepare("select * from tbl_user where user_name=? limit 1")
	if err != nil {
		fmt.Println(err.Error())
		return false
	}

	rows, err := stmt.Query(userName)
	if err != nil {
		fmt.Println(err.Error())
		return false
	} else if rows == nil {
		fmt.Println("Username not found: " + userName)
	}

	pRows := mydb.ParseRows(rows)
	if len(pRows) > 0 && string(pRows[0]["user_pwd"].([]byte)) == encpwd {
		return true
	}
	return false
}

// UpdateToken:刷新用户登录的token
func UpdateToken(userName string, token string) bool {
	stmt, err := mydb.DBConn().Prepare("replace into tbl_user_token(`user_name`, `user_token`) values (?,?)")
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	defer stmt.Close()

	_, err = stmt.Exec(userName, token)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	return true
}
