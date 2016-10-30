package main

import (
	"fmt"
	"net/http"
	"html/template"
	"github.com/jinzhu/gorm"
	//_ "github.com/go-sql-driver/mysql"
	"strconv"
	"github.com/go-web-framework/gflux/mux"
	//"io/ioutil"
	//"./mux"
	//"github.com/google/uuid"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

var db *gorm.DB
var userName string
var userID uint
var templates = template.Must(template.ParseFiles("goblog.html", "page.html"))

// table posts (
//   Post_id: int (autoincrement)
//   Author: varchar(30)
//   Aext: varchar(200)
// )
func main(){

	var err error
	db, err = gorm.Open("sqlite3", "goblog.db")
	if err != nil {
		panic("failed to connect database")
	}
	defer db.Close()
	
	// Migrate the schema
	//db.CreateTable(&User{})
 	db.AutoMigrate(&Post{})
 	db.AutoMigrate(&User{})
 	db.Delete(Post{})
	db.Delete(User{})
	
	//create default user
	userName = "default"
	var user = User{Name: userName}
	db.Create(&user)
	var user2 User
	db.Where("Name= ?", userName).First(&user2)
	userID = user2.ID
  
	testMux := mux.New()
	homeHandler := homeHandler{}
	pageHandler := pageHandler{}
	newHandler := newHandler{}
	jsHandler := jsHandler{}
	userHandler := userHandler{}
	newUserHandler := newUserHandler{}
	changeUserHandler := changeUserHandler{}
	testMux.GET("/home", nil, homeHandler)
	testMux.GET("/page/{id}", nil, pageHandler)
	testMux.POST("/page/new", nil, newHandler)
	testMux.GET("/newuser", nil, userHandler)
	testMux.POST("/makeNewUser", nil, newUserHandler)
	testMux.GET("/count.js", nil, jsHandler)
	testMux.POST("/changeUser", nil, changeUserHandler)
	fmt.Println("Listening on :8080")
	http.ListenAndServe(":8080", testMux)
	
}
type changeUserHandler struct{}
func (t changeUserHandler) ServeHTTP(w http.ResponseWriter, r *http.Request){
	name := r.FormValue("Name")
	var user User
	db.Where("Name= ?", name).First(&user)
	userName = user.Name
	userID = user.ID
	http.Redirect(w, r, "/home", http.StatusFound)
}
type userHandler struct{}
func (t userHandler) ServeHTTP(w http.ResponseWriter, r *http.Request){
	http.ServeFile(w, r, "newuser.html")
}

type newUserHandler struct{}
func (t newUserHandler) ServeHTTP(w http.ResponseWriter, r *http.Request){
	name := r.FormValue("Name")
	//no name
	//unique name
	if (name == ""){
		http.Redirect(w, r, "/newUser", http.StatusFound)
	}
	//store user
	var userList []User
	db.Find(&userList)
	var user = User{Name:name}
	db.Create(&user)
	
	http.Redirect(w, r, "/home", http.StatusFound)
}

type jsHandler struct{}
func (t jsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request){
	http.ServeFile(w, r, "count.js");
}
type homeHandler struct{
}

func (t homeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request){
	//make database call
	var postList []Post
	var user User
	db.Where("Name = ?", userName).First(&user)
	//db.Find(&postList)
	db.Model(&user).Related(&postList)
	err := templates.ExecuteTemplate(w, "goblog.html", &goBlog{Username: userName, PostList: postList})
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
	
	return
}

type pageHandler struct{
}

func (t pageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request){
	//database call
	/*url := r.URL.Path
	urls := strings.Split(url, "/")
	idURL, err := strconv.Atoi(urls[len(urls)-1])
	if (err != nil){
		idURL = 0
	}*/
	params := mux.GetParams(r)
	idURL, err := strconv.Atoi(params["id"])
	if (err != nil){
		idURL = 0
	}
	var post Post
	db.Where("Post_id = ?", idURL).First(&post)
	err = templates.ExecuteTemplate(w, "page.html", &post)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
	return
}

type newHandler struct{
}

func (t newHandler) ServeHTTP(w http.ResponseWriter, r *http.Request){
	author := r.FormValue("Author")
	if (author == ""){
		author = "anon"
	}
	text := r.FormValue("Text")
	var postList []Post
	db.Find(&postList)
	//store post
	var post = Post{UserID: userID, Author:author, Text:text}
	db.Create(&post)
	
	http.Redirect(w, r, "/home", http.StatusFound)
}


//ajax
type goBlog struct{
	Username string
	PostList []Post
}

type Post struct{
	gorm.Model
	User User			`gorm:"ForeignKey:UserID"`
	UserID 		uint
	//Post_id 	int 	`gorm:""primary_key"`
	Author 		string `gorm:"type:varchar(20)"`
	Text 			string	`gorm:"type:varchar(200)"`
	
}

type User struct{
	gorm.Model
	Posts []Post
	//User_id 	int `gorm:"primary_key"`
	Name		string `gorm:"type:varchar(20)"`
}

type handler404 struct{
}

func (t handler404) ServeHTTP(w http.ResponseWriter, r *http.Request){
	fmt.Fprintf(w, "<h1>You've reached a custom 404!</h1>")
	return
}
