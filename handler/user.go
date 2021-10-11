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

	if len(userName) < 3 || len(passwd) < 5 {
		w.Write([]byte("Invilid parameter"))
		return
	}

	enc_passwd := util.Sha1([]byte(passwd + pwd_salt))
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
	upRes := dblayer.UpdateToken(userName, token)
	if !upRes {
		w.Write([]byte("Failed!"))
		return
	}

	// 3.登录成功后重定向到首页
	//w.Write([]byte("http://" + r.Host + "/static/view/home.html"))
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

func GenToken(userName string) string {
	// 40位字符md5(userName+timestamp+token_salt)+timestamp[:8]
	ts := fmt.Sprintf("%x", time.Now().Unix())
	tokenPrefix := util.MD5([]byte(userName + ts + "_tokensalt"))
	return tokenPrefix + ts[:8]
}
