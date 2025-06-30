package main

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"

	// "net/http"
	"log"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

type User struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func main() {
	var user User
	var query string

	connStr := "postgres://localhost/go_http_gin_db?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	logFile, err := os.OpenFile("log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(logFile)

	logAction := func(msg string) {
		log.Println(msg)
	}

	r := gin.Default()

	r.POST("/user", func(c *gin.Context) {
		c.BindJSON(&user)
		query = "INSERT INTO users (name, age) VALUES ($1, $2) RETURNING id"
		err = db.QueryRow(query, user.Name, user.Age).Scan(&user.Id)
		if err != nil {
			c.JSON(400, gin.H{"error": "Bad request"})
			return
		}

		c.JSON(200, gin.H{
			"message": fmt.Sprintf("Имя: %s. Возраст: %d. ID: %d", user.Name, user.Age, user.Id),
		})
		logAction(fmt.Sprintf("Создан пользователь: %s, %d лет (id=%d)", user.Name, user.Age, user.Id))
	})

	r.GET("/users", func(c *gin.Context) {
		query = "SELECT * FROM users ORDER BY id"
		rows, err := db.Query(query)
		if err != nil {
			c.JSON(500, gin.H{"error": "Ошибка чтения из БД"})
			return
		}
		defer rows.Close()

		var users []User

		for rows.Next() {
			var u User
			if err := rows.Scan(&u.Id, &u.Name, &u.Age); err != nil {
				c.JSON(500, gin.H{"error": "Ошибка при сканировании строки"})
				return
			}

			users = append(users, u)
		}

		c.IndentedJSON(200, users)
		logAction("Запрошены все пользователи в базе данных")
	})

	r.DELETE("/user/:id", func(c *gin.Context) {
		query = "DELETE FROM users WHERE id = $1"
		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			c.JSON(400, gin.H{"error": "Bad request"})
			return
		}

		res, err := db.Exec(query, id)
		if err != nil {
			c.JSON(500, gin.H{"error": "Internal server error"})
			return
		}

		rowsAffected, err := res.RowsAffected()
		if err != nil {
			c.JSON(500, gin.H{"error": "Internal server error"})
			return
		}

		if rowsAffected > 0 {
			c.JSON(200, gin.H{"message": fmt.Sprintf("Пользователь %d удален из базы данных", id)})
			logAction(fmt.Sprintf("Удален пользователь с id: %d", id))
		} else {
			c.JSON(404, gin.H{"error": "Пользователь не найден"})
		}
	})

	r.GET("/user/:id", func(c *gin.Context) {
		var user User
		query = "SELECT * FROM users WHERE id = $1"
		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			c.JSON(400, gin.H{"error": "Bad request"})
			return
		}

		err = db.QueryRow(query, id).Scan(&user.Id, &user.Name, &user.Age)
		if err == sql.ErrNoRows {
			c.JSON(404, gin.H{"error": "Bad request"})
			return
		} else if err != nil {
			c.JSON(500, gin.H{"error": "Internal server error"})
			return
		}

		c.JSON(200, user)
		logAction(fmt.Sprintf("Запрошен пользователь с id: %d", id))
	})

	r.PUT("/user/:id", func(c *gin.Context) {
		var user User
		query = "UPDATE users SET name = $1, age = $2 WHERE id = $3"
		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			c.JSON(400, gin.H{"error": "Bad request"})
			return
		}

		if err := c.BindJSON(&user); err != nil {
			c.JSON(400, gin.H{"error": "Bad request"})
			return
		}

		res, err := db.Exec(query, user.Name, user.Age, id)
		if err != nil {
			c.JSON(500, gin.H{"error": "Internal server error"})
			return
		}

		rowsAffected, err := res.RowsAffected()
		if err != nil {
			c.JSON(500, gin.H{"error": "Internal server error"})
			return
		}

		if rowsAffected > 0 {
			user.Id = id
			c.JSON(200, user)
			logAction(fmt.Sprintf("Обновлен пользователь с id: %d -> имя: %s; возраст: %d", user.Id, user.Name, user.Age))
			return
		} else {
			c.JSON(404, gin.H{"error": "Пользователь не найден"})
			return
		}
	})

	r.Run()
}
