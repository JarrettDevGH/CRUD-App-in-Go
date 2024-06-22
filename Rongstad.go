package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"strconv"

	"github.com/gorilla/sessions"
	_ "github.com/gorilla/sessions"
	_ "github.com/mattn/go-sqlite3"
)

type User struct {
	Id       int
	Username string
	Password string
}

type Recommendaton struct {
	Id       int
	Url      string
	Category string
	User_id  int
}

var (
	key   = []byte("super-secret-key")
	store = sessions.NewCookieStore(key)
)

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
func connect() *sql.DB {
	db, err := sql.Open("sqlite3", "./Rongstad.sqlite")
	check(err)
	return db
}
func home(writer http.ResponseWriter, request *http.Request) {
	html, _ := template.ParseFiles("index.html")
	session, err := store.Get(request, "cookie-name")
	check(err)
	wow := session.Values["authenticated"]
	html.Execute(writer, wow)
}
func InsertUser(username string, password string) {
	db := connect()
	cmd := "INSERT INTO users " + "(username, password) " + "VALUES ('" + username + "', '" + password + "')"
	fmt.Println(cmd)
	_, err := db.Exec(cmd)
	check(err)
}

func user(writer http.ResponseWriter, request *http.Request) {
	if request.Method == "POST" {
		username := request.FormValue("username")
		password := request.FormValue("password")
		InsertUser(username, password)
		http.Redirect(writer, request, "/login", http.StatusFound)
	} else {
		html, _ := template.ParseFiles("user.html")
		html.Execute(writer, home)
	}
}
func login(writer http.ResponseWriter, request *http.Request) {
	session, _ := store.Get(request, "cookie-name")
	if request.Method == "POST" {
		db := connect()
		defer db.Close()
		eUsername := request.FormValue("username")
		ePassword := request.FormValue("password")
		fmt.Println(eUsername)
		fmt.Println(ePassword)
		cmd := "SELECT * FROM users WHERE username = ?" //and password=?"
		stmt, err2 := db.Prepare(cmd)
		check(err2)
		defer stmt.Close()
		var myUser User
		var id int
		var username string
		var password string
		stmt.QueryRow(eUsername).Scan(&username, &password, &id)
		myUser.Id = id
		myUser.Username = username
		myUser.Password = password
		fmt.Println(myUser.Username)
		fmt.Println(myUser.Password)

		if myUser.Password == ePassword && myUser.Username == eUsername {
			session.Values["authenticated"] = true
			session.Values["id"] = id
			//fmt.Println("gang2 gang")
			session.Save(request, writer)
			http.Redirect(writer, request, "/", http.StatusFound)
		} else {
			//fmt.Println("gang3 ga2222222ng")
			http.Redirect(writer, request, "/login", http.StatusFound)
		}
	} else {
		html, _ := template.ParseFiles("login.html")
		html.Execute(writer, nil)
	}
}
func logout(writer http.ResponseWriter, request *http.Request) {
	session, err := store.Get(request, "cookie-name")
	check(err)
	session.Values["authenticated"] = false
	session.Save(request, writer)
	http.Redirect(writer, request, "/", http.StatusFound)
}
func InsertData(recURL string, category string, user_id int) {
	db := connect()
	cmd := fmt.Sprintf("INSERT INTO data (recURL, category, user_id) VALUES ('%s', '%s', %d)", recURL, category, user_id)
	fmt.Println(cmd)
	_, err := db.Exec(cmd)
	check(err)
}
func create(writer http.ResponseWriter, request *http.Request) {
	session, err := store.Get(request, "cookie-name")
	check(err)
	auth, ok := session.Values["authenticated"].(bool)
	if !ok || !auth {
		http.Redirect(writer, request, "/login", http.StatusFound)
		return
	}
	if request.Method == "POST" {
		recURL := request.FormValue("recomendation")
		category := request.FormValue("category")
		user_id := session.Values["id"].(int)
		InsertData(recURL, category, user_id)
		http.Redirect(writer, request, "/read", http.StatusFound)
	} else {
		html, _ := template.ParseFiles("create.html")
		html.Execute(writer, create)
	}
}
func read(writer http.ResponseWriter, request *http.Request) {
	session, err := store.Get(request, "cookie-name")
	check(err)
	auth, ok := session.Values["authenticated"].(bool)
	if !ok || !auth {
		http.Redirect(writer, request, "/login", http.StatusFound)
		return
	}
	if request.Method == "GET" {
		db := connect()
		var recSlice = make([]Recommendaton, 0)
		cmd := "SELECT * FROM data"
		rows, err := db.Query(cmd)
		check(err)
		for rows.Next() {
			var recommendation Recommendaton
			var id int
			var recURL string
			var category string
			var user_id int
			err := rows.Scan(&id, &recURL, &category, &user_id)
			check(err)
			recommendation.Id = id
			recommendation.Url = recURL
			recommendation.Category = category
			recommendation.User_id = user_id
			//if err != nil {log.Fatal(err)}
			if session.Values["id"].(int) == recommendation.User_id {
				recSlice = append(recSlice, Recommendaton{Id: id, Url: recURL, Category: category, User_id: user_id})
			}
		}
		html, _ := template.ParseFiles("read.html")
		html.Execute(writer, recSlice)
	}

}
func update(writer http.ResponseWriter, request *http.Request) {
	session, err := store.Get(request, "cookie-name")
	check(err)
	auth, ok := session.Values["authenticated"].(bool)
	if !ok || !auth {
		http.Redirect(writer, request, "/login", http.StatusFound)
		return
	}
	if request.Method == "GET" {
		db := connect()
		defer db.Close()
		idd := request.FormValue("id")
		cmd := "SELECT * FROM data WHERE id=?"
		stmt, err2 := db.Prepare(cmd)
		check(err2)
		defer stmt.Close()
		var recommendation Recommendaton
		var id int
		var recURL string
		var category string
		var user_id int
		stmt.QueryRow(idd).Scan(&id, &recURL, &category, &user_id)
		recommendation.Id = id
		recommendation.Url = recURL
		recommendation.Category = category
		recommendation.User_id = user_id
		fmt.Println(request.FormValue("hidden"))
		//wow, err := strconv.Atoi(request.FormValue("hidden"))
		//check(err)
		fmt.Println(session.Values["id"].(int))
		//fmt.Println(wow)
		if session.Values["id"] == recommendation.User_id {
			html, _ := template.ParseFiles("update.html")
			html.Execute(writer, recommendation)
		} else {
			http.Redirect(writer, request, "/logout", http.StatusFound)
			//http.Redirect(writer, request, "/login", http.StatusFound)
		}
	} else {
		db := connect()
		defer db.Close()
		recURL := request.FormValue("recURL")
		category := request.FormValue("category")
		id, err := strconv.Atoi(request.FormValue("id"))
		check(err)
		fmt.Println(recURL)
		fmt.Println(category)
		wow, err := strconv.Atoi(request.FormValue("hidden"))
		check(err)
		if session.Values["id"].(int) == wow {
			cmd := "UPDATE data SET recURL = ?, category = ? WHERE id = ?"
			db.Exec(cmd, recURL, category, id)
			http.Redirect(writer, request, "/read", http.StatusFound)
		} else {
			fmt.Print("hey")
			//http.Redirect(writer, request, "/logout", http.StatusFound)
			//http.Redirect(writer, request, "/login", http.StatusFound)
		}
	}
}
func delete(writer http.ResponseWriter, request *http.Request) {
	session, err := store.Get(request, "cookie-name")
	check(err)
	auth, ok := session.Values["authenticated"].(bool)
	if !ok || !auth {
		http.Redirect(writer, request, "/login", http.StatusFound)
		return
	}
	db := connect()
	defer db.Close()
	id := request.FormValue("id")
	cmd := "DELETE from data where id=?"
	db.Exec(cmd, id)
	http.Redirect(writer, request, "/read", http.StatusFound)
}

func main() {
	http.HandleFunc("/", home)
	http.HandleFunc("/user", user)
	http.HandleFunc("/login", login)
	http.HandleFunc("/logout", logout)
	http.HandleFunc("/create", create)
	http.HandleFunc("/update", update)
	http.HandleFunc("/delete", delete)
	http.HandleFunc("/read", read)
	http.ListenAndServe("localhost:5000", nil)
}
