package main

import (
	"fmt"
	"go_final_project/cmd/db"
	"go_final_project/cmd/handler"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	_ "modernc.org/sqlite"
)

func main() {

	err := godotenv.Load()

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	port := os.Getenv("TODO_PORT")

	db.CheckAndOpenDB()
	defer db.DB.Close()

	http.Handle("/", http.FileServer(http.Dir("./web")))
	http.HandleFunc("/api/nextdate", handler.NextDateHandl)
	http.HandleFunc("/api/task", handler.TaskHandler)
	http.HandleFunc("/api/tasks", handler.TasksGet)
	http.HandleFunc("/api/task/done", handler.TaskDone)

	fmt.Println("Запуск сервера")
	err = http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Printf("Ошибка при запуске сервера: %s", err)
	}

}
