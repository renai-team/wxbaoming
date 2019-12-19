package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"gopkg.in/yaml.v2"
)

//ApplyMSG 报名信息
type ApplyMSG struct {
	Name      string `json:"name,omitempty"`
	Snumber   string `json:"snumber,omitempty"`
	Major     string `json:"major,omitempty"`
	Class     string `json:"class,omitempty"`
	Sex       int    `json:"sex,omitempty"`
	Telephone string `json:"telephone,omitempty"`
	QQ        string `json:"qq,omitempty"`
	MSG       string `json:"msg,omitempty"`
}

//Config 配置文件
type Config struct {
	Drivername string `yaml:"drivername"`
	DNS        string `yaml:"dns"`
}

//Apply 报名
type Apply struct {
	db *sql.DB //数据库连接
	a  *ApplyMSG
}

var config = new(Config)

//OpenDB 打开连接
func OpenDB() *sql.DB {

	db, err := sql.Open(config.Drivername, config.DNS)
	if err != nil {
		log.Println("数据库初始化错误：", err)
	}
	err = db.Ping()
	if err != nil {
		log.Println("数据库打开错误：", err)
	}
	return db
}

//NewApplyMSG 实例
func NewApplyMSG() *ApplyMSG {
	var a = new(ApplyMSG)
	return a
}

//NewApply 实例化
func NewApply() *Apply {

	GetConfig()
	apply := new(Apply)
	apply.db = OpenDB()
	apply.a = NewApplyMSG()
	return apply
}

//GetConfig 解析yaml文件
func GetConfig() {

	buffer, err := ioutil.ReadFile("config.yaml")
	if err != nil {

		log.Println("打开yaml文件出错：", err)
		return
	}
	err = yaml.Unmarshal(buffer, &config)
	if err != nil {
		log.Println("解析yaml文件出错:\n", err)
	}
}

func (apply *Apply) selectMsg() (bool, int) {

	var times int
	err := apply.db.QueryRow("select time from apply where snumber=?", apply.a.Snumber).Scan(&times)
	if err != nil {
		if err == sql.ErrNoRows {
			return true, 0
		}
		log.Fatal("查询错误：", err)
	}
	return false, times
}

func (apply *Apply) addMsg(times int) {

	stmt, stmtErr := apply.db.Prepare("insert into apply values(?,?,?,?,?,?,?,?)")
	if stmtErr != nil {
		log.Println("准备sql出错：", stmtErr)
	}
	_, resErr := stmt.Exec(apply.a.Name, apply.a.Snumber, apply.a.Major, apply.a.Class, apply.a.Sex, apply.a.Telephone, apply.a.QQ, times)
	if resErr != nil {
		log.Println("添加信息错误：", resErr)
	}
}

func (apply *Apply) deleteMsg() {

	stmt, stmtErr := apply.db.Prepare("delete from apply where snumber=?")
	if stmtErr != nil {
		log.Println("准备sql出错：", stmtErr)
	}
	_, resErr := stmt.Exec(apply.a.Snumber)
	if resErr != nil {
		log.Println("删除信息错误：", resErr)
	}
}

func wxapply(w http.ResponseWriter, r *http.Request) {

	apply := NewApply()
	err := json.NewDecoder(r.Body).Decode(apply.a)
	if err != nil {
		log.Println("解码错误：", err)
		response(w, "提交错误", 500)
	}
	if apply.a.Name == "" || apply.a.Snumber == "" || apply.a.Major == "" || apply.a.Class == "" || apply.a.QQ == "" {
		response(w, "请完善信息", 500)
		return
	}
	//查找
	var ok bool
	var times int
	if ok, times = apply.selectMsg(); !ok {

		//报过名了
		if times <= 3 {
			fmt.Println("报名了 ", times, " 了")
			//不超过三次
			apply.deleteMsg()
		} else {
			response(w, "重复报名", 500)
			return
		}
	}
	apply.addMsg(times + 1)
	response(w, "报名成功", 200)
	return
}

func response(w http.ResponseWriter, msg string, status int) {

	w.Header().Set("Content-Type", "application/json")
	b, err := json.Marshal(&ApplyMSG{MSG: msg})
	if err != nil {
		log.Println("编码错误：", err)
	}
	w.Write(b)
	w.WriteHeader(status)
}

func main() {

	http.HandleFunc("/wxapply", wxapply)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
