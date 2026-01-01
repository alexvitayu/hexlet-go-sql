package storage

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
)

type Course struct {
	ID    int
	Slug  string
	Title string
	Price int
}

func CreateCourse(ctx context.Context, db *sql.DB, slug, title string, price int) (int, error) {
	const schema = `CREATE TABLE IF NOT EXISTS courses(
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    slug TEXT NOT NULL UNIQUE,
    title TEXT NOT NULL,
    price INTEGER NOT NULL DEFAULT 0
);`

	if _, err := db.ExecContext(ctx, schema); err != nil {
		log.Fatalf("create table: %v", err)
	}

	var id int

	err := db.QueryRowContext(ctx,
		`INSERT INTO courses (slug, title, price) VALUES(?, ?, ?) RETURNING id`,
		slug, title, price).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("create course: %v", err)
	}
	return id, nil
}

func ListCourses(ctx context.Context, db *sql.DB, userSort string) ([]Course, error) {
	var courses []Course

	// Белый список допустимых вариантов сортировки.
	allowed := map[string]string{
		"id_asc":    "id ASC",
		"id_desc":   "id DESC",
		"name_asc":  "name ASC",
		"name_desc": "name DESC",
	}

	order, ok := allowed[userSort]
	if !ok {
		order = "id ASC"
	}

	query := `SELECT id, slug, title, price FROM courses ORDER BY ` + order + ``

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list_courses: %w", err)
	}

	defer rows.Close()

	for rows.Next() {
		var c Course
		if err = rows.Scan(&c.ID, &c.Slug, &c.Title, &c.Price); err != nil {
			return nil, fmt.Errorf("query_listCourses: %w", err)
		}
		courses = append(courses, c)
	}
	return courses, nil
}

func FindCoursesByIDs(ctx context.Context, db *sql.DB, ids ...int) ([]Course, error) {
	// Генерируем строку вида "(?,?,?)".
	placeholders := strings.TrimRight(strings.Repeat("?,", len(ids)), ",")

	// Превращаем []int в []interface{} для передачи в ExecContext/QueryContext.
	args := make([]interface{}, len(ids))
	for i, v := range ids {
		args[i] = v
	}

	query := fmt.Sprintf("SELECT id, slug, title, price FROM courses WHERE id IN (%s)", placeholders)

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("find_courses: %w", err)
	}
	defer rows.Close()

	var courses []Course
	var c Course
	for rows.Next() {
		if err = rows.Scan(&c.ID, &c.Slug, &c.Title, &c.Price); err != nil {
			return nil, fmt.Errorf("find_corses: %w", err)
		}
		courses = append(courses, c)
	}
	return courses, nil
}
