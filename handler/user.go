package handler

import (
	dblayer "distributed-fileServer/db"
	"distributed-fileServer/util"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	pwd_salt = "*#890"
)

// SignUpHandler:处理用户注册请求
func SignUpHandler(w http.ResponseWriter, r *http.Request) {

	// 1.判断如果是http Get请求，直接返回注册页面内容
	if r.Method == http.MethodGet {
		data, err := ioutil.ReadFile("./static/view/signup.html")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write(data)
		return
	}

	r.ParseForm()
	userName := r.Form.Get("username")
	passwd := r.Form.Get("password")

	// 2.校验参数的有效性
	if len(userName) < 3 || len(passwd) < 5 {
		w.Write([]byte("Invilid parameter"))
		return
	}

	// 3.加密用户名密码
	enc_passwd := util.Sha1([]byte(passwd + pwd_salt))

	// 4.存入数据库表tbl_user并返回结果
	suc := dblayer.UserSignUp(userName, enc_passwd)
	if suc {
		w.Write([]byte("SUCCESS"))
	} else {
		w.Write([]byte("Failed"))
	}
}

// SignInHandler：登录接口
func SignInHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodGet {
		// data, err := ioutil.ReadFile("./static/view/signin.html")
		// if err != nil {
		//  w.WriteHeader(http.StatusInternalServerError)
		//  return
		// }
		// w.Write(data)
		http.Redirect(w, r, "/static/view/signin.html", http.StatusFound)
		return
	}

	r.ParseForm()
	userName := r.Form.Get("username")
	passWord := r.Form.Get("password")
	encPasswd := util.Sha1([]byte(passWord + pwd_salt))

	// 1.校验用户名及密码
	pwdChecked := dblayer.UserSignIn(userName, encPasswd)
	if !pwdChecked {
		w.Write([]byte("Failed!"))
		return
	}

	// 2.生成访问凭证（token）
	token := GenToken(userName)

	// 3.存储token到tbl_user_token表
	upRes := dblayer.UpdateToken(userName, token)
	if !upRes {
		w.Write([]byte("Failed!"))
		return
	}

	// 4.登录成功后重定向到首页
	resp := util.RespMsg{
		Code: 0,
		Msg:  "OK",
		Data: struct {
			Location string
			Username string
			Token    string
		}{
			Location: "http://" + r.Host + "/static/view/home.html",
			Username: userName,
			Token:    token,
		},
	}
	w.Write(resp.JSONBytes())

}

// UserInfoHandler:查新用户信息接口
func UserInfoHandler(w http.ResponseWriter, r *http.Request) {
	// 1.解析请求参数
	r.ParseForm()
	userName := r.Form.Get("username")
	token := r.Form.Get("token")

	//2.验证token是否有效
	isValidToken := IsTokenValid(token)
	if !isValidToken {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// 3.查询用户信息
	user, err := dblayer.GetUserInfo(userName)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	// 4.组装并且响应用户数据
	resp := util.RespMsg{
		Code: 0,
		Msg:  "OK",
		Data: user,
	}
	w.Write(resp.JSONBytes())
}

func GenToken(userName string) string {
	// 40位字符md5(userName+timestamp+token_salt)+timestamp[:8]
	ts := fmt.Sprintf("%x", time.Now().Unix())
	tokenPrefix := util.MD5([]byte(userName + ts + "_tokensalt"))
	return tokenPrefix + ts[:8]
}

// IsTokenValid：token是否有效
func IsTokenValid(token string) bool {
	if len(token) != 40 {
		return false
	}
	// 判断token的时效性，是否过期
	// example，假设token的有效期为1天
	//n := len(token)
	//tokenTS := token[n-8:]
	//if util.Hex2Dec(tokenTS) < time.Now().Unix()-86400 {
	//	return false
	//}
	// TODO:从数据库表tbl-user_token查询对应的username的token信息
	// TODO:对比两个token是否一致
	return true
}
