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

var db *sql.DB
var err error

func main() {
	db, err = sql.Open("mysql", "root:@/facebookdb")
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()
	
	router := mux.NewRouter()

	router.HandleFunc("/posts/{id}", getPosts).Methods("POST")
	router.HandleFunc("/posts/create/{id}", createPost).Methods("POST")

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