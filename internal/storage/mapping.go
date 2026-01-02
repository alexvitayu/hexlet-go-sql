package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

const createTable = `CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(100),
    age INTEGER,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
    );`

type User struct {
	ID         int            `json:"id" db:"id"`
	Email      string         `json:"email" db:"email"`
	Name       sql.NullString `json:"name" db:"name"` //колонка может быть NULL
	Age        sql.NullInt64  `json:"age" db:"age"`   //колонка может быть NULL
	Created_at time.Time      `json:"created_at" db:"created_at"`
}

// DTO - data transfer object
type CreateUserDTO struct {
	Email string         `json:"email" db:"email"`
	Name  sql.NullString `json:"name" db:"name"` //колонка может быть NULL
	Age   sql.NullInt64  `json:"age" db:"age"`   //колонка может быть NULL
}

// DTO - data transfer object
type UpdateUserDTO struct {
	ID    int            `json:"id" db:"id"`
	Email string         `json:"email" db:"email"`
	Name  sql.NullString `json:"name" db:"name"` //колонка может быть NULL
	Age   sql.NullInt64  `json:"age" db:"age"`   //колонка может быть NULL
}

func CreateUser(ctx context.Context, db *sql.DB, dto CreateUserDTO) (User, error) {
	if _, err := db.ExecContext(ctx, createTable); err != nil {
		return User{}, fmt.Errorf("create table: %w", err)
	}

	const query = `INSERT INTO users (email, name, age, created_at) VALUES (?, ?, ?, CURRENT_TIMESTAMP) RETURNING id, email, name, age, created_at`

	var u User
	var createdAtStr string
	err := db.QueryRowContext(ctx, query, dto.Email, dto.Name, dto.Age).Scan(&u.ID, &u.Email, &u.Name, &u.Age, &createdAtStr)
	if err != nil {
		return User{}, fmt.Errorf("createUser: %w", err)
	}
	u.Created_at, err = time.Parse("2006-01-02 15:04:05", createdAtStr)
	if err != nil {
		return User{}, fmt.Errorf("createUser: %w", err)
	}

	return u, nil
}

func UpdateUser(ctx context.Context, db *sql.DB, dto UpdateUserDTO) (User, error) {
	//COALESCE возвращает первый не NULL аргумент функции и читает строку слева направо
	const query = `
				UPDATE users SET
                 email = COALESCE(?, email),
                 name = COALESCE(?, name),
                 age = COALESCE(?, age)
             	WHERE id = ?
             	RETURNING id, email, name, age, created_at`

	var u User

	if err := db.QueryRowContext(ctx,
		query, dto.Email, dto.Name, dto.Age, dto.ID).Scan(&u.ID, &u.Email, &u.Name, &u.Age, &u.Created_at); err != nil {
		return User{}, fmt.Errorf("updateUser: %w", err)
	}
	return u, nil
}

func GetUser(ctx context.Context, db *sql.DB, id int64) (User, error) {
	const query = `SELECT id, email, name, age, created_at FROM users WHERE id = ?`

	var u User
	var createdAtStr string

	err := db.QueryRowContext(ctx,
		query, id).Scan(&u.ID, &u.Email, &u.Name, &u.Age, &createdAtStr)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, fmt.Errorf("user not found: %w", err)
		}
		return User{}, fmt.Errorf("getUser: %w", err)
	}
	u.Created_at, err = time.Parse("2006-01-02 15:04:05", createdAtStr)
	if err != nil {
		return User{}, fmt.Errorf("getUser: %w", err)
	}
	return u, nil
}

func ListUsers(ctx context.Context, db *sql.DB) ([]User, error) {

	const query = `SELECT id, email, name, age, created_at FROM users ORDER BY id`

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return []User{}, fmt.Errorf("listUSer: %w", err)
	}
	defer rows.Close()

	var users []User

	for rows.Next() {
		var u User
		if err = rows.Scan(&u.ID, &u.Email, &u.Name, &u.Age, &u.Created_at); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, u)
	}
	return users, rows.Err()
}
