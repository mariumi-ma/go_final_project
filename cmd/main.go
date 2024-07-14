package main

import (
	"database/sql"
	"go_final_project/internal/handler"
	"go_final_project/internal/storage"
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

	err = storage.CheckDB()
	if err != nil {
		log.Fatal(err)
	}

	dbOpen, err := sql.Open("sqlite", storage.FileDB)
	if err != nil {
		return
	}
	defer dbOpen.Close()

	DB := storage.NewTasksDB(dbOpen)

	http.Handle("/", http.FileServer(http.Dir("./web")))
	http.HandleFunc("/api/nextdate", handler.GetNextDate)
	http.HandleFunc("/api/task", handler.TaskHandler(DB))
	http.HandleFunc("/api/tasks", handler.GetTasks(DB))
	http.HandleFunc("/api/task/done", handler.TaskDone(DB))

	log.Println("Запуск сервера: http://localhost:7540/")
	err = http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Printf("Ошибка при запуске сервера: %s", err)
	}

}
