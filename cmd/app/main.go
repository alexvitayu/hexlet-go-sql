package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"example.com/go-sql/internal/storage"
	_ "modernc.org/sqlite"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	db, err := sql.Open("sqlite", "file:data.db?_foreign_keys=on&_busy_timeout=5000")
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("ping db: %v", err)
	}

	// create/update/get/list/delete
	crud := flag.String("crud", "get", "defines an operation with DB")

	id := flag.Int("id", 1, "receives id")

	email := flag.String("email", "example@yandex.ru", "")
	name := flag.String("name", "Vasia", "")
	age := flag.Int("age", 30, "")

	flag.Parse()

	if *crud == "get" {
		getUser(ctx, db, int64(*id))
	} else if *crud == "create" {
		createUser(ctx, db, *email, *name, int64(*age))
	} else {
		fmt.Println("smth else")
	}
}

func createUser(ctx context.Context, db *sql.DB, email, name string, age int64) {
	dto := storage.CreateUserDTO{
		Email: email,
		Name: sql.NullString{
			String: name,
			Valid:  true,
		},
		Age: sql.NullInt64{
			Int64: age,
			Valid: true,
		},
	}
	u := storage.User{}

	u, err := storage.CreateUser(ctx, db, dto)
	if err != nil {
		os.Stderr.WriteString(err.Error())
	}
	res, err := json.MarshalIndent(u, "", "")
	if err != nil {
		os.Stderr.WriteString(err.Error())
	}
	fmt.Println(string(res))
}

func getUser(ctx context.Context, db *sql.DB, id int64) {
	u := storage.User{}
	u, err := storage.GetUser(ctx, db, id)
	if err != nil {
		os.Stderr.WriteString(err.Error())
	}
	out, err := json.MarshalIndent(u, "", "")
	if err != nil {
		os.Stderr.WriteString(err.Error())
	}
	fmt.Println(string(out))
}
