package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"example.com/go-sql/internal/storage"
	_ "modernc.org/sqlite"
)

var menuVariants = []string{
	" 1. Создать курс;",
	" 2. Посмотреть все курсы;",
	" 3. Посмотреть курсы по ids;",
	" 4. Выход",
	"выберите вариант",
}

var menu = map[string]func(*sql.DB, context.Context){
	"1": createCourse,
	"2": listCourses,
	"3": listCoursesByIDs,
}

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
Menu:
	for {
		fmt.Println("-- Меню работы с курсами --")
		variant := promptData(menuVariants...)
		menuFunc := menu[variant]
		if menuFunc == nil {
			break Menu
		}
		menuFunc(db, ctx)
	}
}

func promptData(prompt ...string) string {
	for i, line := range prompt {
		if i == len(prompt)-1 {
			fmt.Printf("%v:", line)
		} else {
			fmt.Println(line)
		}
	}
	var info string
	fmt.Scanln(&info)
	return info
}

func createCourse(db *sql.DB, ctx context.Context) {
	fmt.Print("Введите слоган: ")
	var slug string
	if _, err := fmt.Scanln(&slug); err != nil {
		log.Fatalf("input_slug: %v", err)
	}
	fmt.Print("Введите титул: ")
	var title string
	if _, err := fmt.Scanln(&title); err != nil {
		log.Fatalf("input_title: %v", err)
	}
	fmt.Print("Введите цену: ")
	var price int
	if _, err := fmt.Scanln(&price); err != nil {
		log.Fatalf("input_price: %v", err)
	}

	id, err := storage.CreateCourse(ctx, db, slug, title, price)
	if err != nil {
		log.Fatalf("create_course: %v", err)
	}
	fmt.Printf("course_id = %v\n", id)
}

func listCourses(db *sql.DB, ctx context.Context) {
	fmt.Print("Введите вариант сортировки, напимер id_asc: ")
	var sort string
	if _, err := fmt.Scanln(&sort); err != nil {
		log.Fatalf("input_userSort: %v", err)
	}
	courses, err := storage.ListCourses(ctx, db, sort)
	if err != nil {
		log.Printf("list_courses %v", err)
	}
	printCourses(courses)
}

func listCoursesByIDs(db *sql.DB, ctx context.Context) {
	var id int
	ids := make([]int, 0, 5)
Menu:
	for {
		fmt.Print("Введите ID: ")
		if _, err := fmt.Scan(&id); err != nil {
			log.Fatalf("input_ids: %v", err)
		}
		ids = append(ids, id)
		if id == 0 {
			break Menu
		}
	}

	courses, err := storage.FindCoursesByIDs(ctx, db, ids...)
	if err != nil {
		log.Printf("list_courses %v", err)
	}
	printCourses(courses)
}

func printCourses(courses []storage.Course) {
	for _, course := range courses {
		fmt.Printf("course_id = %v\n course_slug = %v\n course_title = %v\n price = %v\n", course.ID, course.Slug, course.Title, course.Price)
	}
}
