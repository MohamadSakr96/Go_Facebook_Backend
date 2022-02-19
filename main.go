package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

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

	router.HandleFunc("/posts", getPosts).Methods("POST")

	http.ListenAndServe(":8000", router)
}

func getPosts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var posts []Post

	result, err := db.Query("SELECT users.user_name,posts.post_id,posts.post_content,posts.post_date,(SELECT COUNT(post_likes.id) from post_likes WHERE post_likes.post_id=posts.post_id) AS nb_likes FROM posts,users WHERE posts.user_id IN(SELECT friend_id FROM user_friends WHERE user_id=19 UNION SELECT user_id FROM user_friends WHERE friend_id=19 UNION SELECT user_id FROM users WHERE user_id=19) AND users.user_id=posts.user_id ORDER BY posts.post_date DESC;")
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
	fmt.Println(posts)
	json.NewEncoder(w) .Encode(posts)
}