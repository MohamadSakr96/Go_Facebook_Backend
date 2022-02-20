package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

type Post struct {
	User string `json:"user_name"`
	ID string `json:"post_id"`
	Content string `json:"post_content"`
	Date string `json:"post_date"`
	Likes string `json:"nb_likes"`
}

type User struct {
	ID string `json:"user_id"`
	Name string `json:"user_name"`
}

var db *sql.DB
var err error

func main() {
	db, err = sql.Open("mysql", "root:@/facebookdb")
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()
	
	router := mux.NewRouter()

	router.HandleFunc("/signup", signup).Methods("POST")
	router.HandleFunc("/login", login).Methods("POST")
	router.HandleFunc("/users/{id}", getUsers).Methods("POST")
	// router.HandleFunc("/posts/create/{id}", createPost).Methods("POST")
	// router.HandleFunc("/posts/update/{id}", updatePost).Methods("PUT")
	// router.HandleFunc("/posts/delete/{id}", deletePost).Methods("DELETE")
	
	router.HandleFunc("/posts/{id}", getPosts).Methods("POST")
	router.HandleFunc("/posts/create/{id}", createPost).Methods("POST")
	router.HandleFunc("/posts/like/{id}", likePost).Methods("POST")
	router.HandleFunc("/posts/update/{id}", updatePost).Methods("PUT")
	router.HandleFunc("/posts/delete/{id}", deletePost).Methods("DELETE")
	
	http.ListenAndServe(":8000", router)
}

func getPosts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	var posts []Post

	result, err := db.Query("SELECT users.user_name,posts.post_id,posts.post_content,posts.post_date,(SELECT COUNT(post_likes.id) from post_likes WHERE post_likes.post_id=posts.post_id) AS nb_likes FROM posts,users WHERE posts.user_id IN(SELECT friend_id FROM user_friends WHERE user_id=? UNION SELECT user_id FROM user_friends WHERE friend_id=? UNION SELECT user_id FROM users WHERE user_id=?) AND users.user_id=posts.user_id ORDER BY posts.post_date DESC;",params["id"], params["id"],params["id"])
	if err != nil {
		panic(err.Error())
	}
	

	defer result.Close()

	for result.Next() {
		var post Post
		err := result.Scan(&post.User, &post.ID, &post.Content, &post.Date, &post.Likes)
		if err != nil {
			panic(err.Error())
		}
		posts = append(posts, post)
	}
	
	json.NewEncoder(w) .Encode(posts)
}

func createPost(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	params := mux.Vars(r)

	stmt, err := db.Prepare("INSERT INTO posts (post_content,post_date,user_id) VALUES (?,?,?);")
	if err != nil {
	  panic(err.Error())
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
	  panic(err.Error())
	}
	current_time := time.Now().Format("2006/01/02 15:04:05")

	keyVal := make(map[string]string)
	json.Unmarshal(body, &keyVal)
	contant := keyVal["post_content"]

	_, err = stmt.Exec(contant,current_time,params["id"])
	if err != nil {
	  panic(err.Error())
	}

	fmt.Fprintf(w, "New post was created")
  }

  func updatePost(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	stmt, err := db.Prepare("UPDATE posts SET post_content =? WHERE post_id=?;")
	if err != nil {
	  panic(err.Error())
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
	  panic(err.Error())
	}
	keyVal := make(map[string]string)
	json.Unmarshal(body, &keyVal)
	contant := keyVal["post_contant"]
	_, err = stmt.Exec(contant, params["id"])
	if err != nil {
	  panic(err.Error())
	}
	fmt.Fprintf(w, "Post was Updated!")
  }

  func deletePost(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	stmt, err := db.Prepare("DELETE FROM posts WHERE post_id=?")
	if err != nil {
	  panic(err.Error())
	}
	_, err = stmt.Exec(params["id"])
	if err != nil {
	  panic(err.Error())
	}
	fmt.Fprintf(w, "Post Deleted!")
  }
  
  func likePost(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	stmt, err := db.Prepare("INSERT INTO post_likes (post_id,user_id) VALUES (?, ?);")
	if err != nil {
	  panic(err.Error())
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
	  panic(err.Error())
	}
	keyVal := make(map[string]string)
	json.Unmarshal(body, &keyVal)
	post_id := keyVal["post_id"]

	_, err = stmt.Exec(post_id, params["id"])
	if err != nil {
	  panic(err.Error())
	}
	fmt.Fprintf(w, "+1 like!")
  }

  
func getUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	var users []User

	result, err := db.Query("SELECT user_id,user_name FROM users WHERE user_id NOT IN(SELECT from_user_id FROM blocks WHERE to_user_id=? UNION SELECT user_id FROM users where user_id=?) AND user_id NOT IN (SELECT friend_id FROM user_friends WHERE user_id=? UNION SELECT user_id FROM user_friends WHERE friend_id=? UNION SELECT user_id FROM users WHERE user_id=?);",params["id"], params["id"],params["id"],params["id"],params["id"])
	if err != nil {
		panic(err.Error())
	}
	
	defer result.Close()

	for result.Next() {
		var user User
		err := result.Scan(&user.ID, &user.Name)
		if err != nil {
			panic(err.Error())
		}
		users = append(users, user)
	}
	
	json.NewEncoder(w) .Encode(users)
}

func login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var users []User

	stmt, err := db.Prepare("SELECT user_id FROM users WHERE user_email = ? AND password = ?;")
	if err != nil {
		panic(err.Error())
	}
	
	defer stmt.Close()

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
	  panic(err.Error())
	}
	keyVal := make(map[string]string)
	json.Unmarshal(body, &keyVal)
	email := keyVal["email"]
	password := keyVal["password"]

	result, err := stmt.Query(email, password)
	if err != nil {
	  panic(err.Error())
	}

	for result.Next() {
		var user User
		err := result.Scan(&user.ID)
		if err != nil {
			panic(err.Error())
		}
		users = append(users, user)
	}
	
	json.NewEncoder(w) .Encode(users)
}



func signup(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	stmt, err := db.Prepare("INSERT INTO users(user_name,user_email,password) VALUES (?,?,?)")
	if err != nil {
	  panic(err.Error())
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
	  panic(err.Error())
	}

	keyVal := make(map[string]string)
	json.Unmarshal(body, &keyVal)
	user_name := keyVal["user_name"]
	user_email := keyVal["user_email"]
	password := keyVal["password"]

	_, err = stmt.Exec(user_name,user_email,password)
	if err != nil {
	  panic(err.Error())
	}

	fmt.Fprintf(w, "User was created")
  }
