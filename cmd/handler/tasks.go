package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"go_final_project/cmd/task"
	"net/http"
	"time"

	_ "modernc.org/sqlite"
)

// TasksGet возвращает весь список задач, максимум 25 строк.
// Если указан параметр "search", то вовзращает задачу с указанным параметром.
func TasksGet(w http.ResponseWriter, req *http.Request) {

	tasks := make(map[string][]task.Task)

	param := req.URL.Query().Get("search")

	db, err := sql.Open("sqlite", "scheduler.db")
	if err != nil {
		fmt.Println("open db")
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	defer db.Close()

	if param != "" {
		var tasksParam []task.Task
		tasksParam, ResponseStatus, err = TasksWithParameter(db, param)
		defer db.Close()
		if err != nil {
			http.Error(w, err.Error(), ResponseStatus)
			return
		}
		tasks["tasks"] = tasksParam

		// Если задач нет, возвращаем пустой json.
		if tasks["tasks"] == nil {
			tasks["tasks"] = []task.Task{}
		}

		response, err := json.Marshal(tasks)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		w.Write(response)
		return
	}

	rows, err := db.Query(`SELECT id, date, title, comment, repeat FROM scheduler 
	ORDER BY date LIMIT 25`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		task := task.Task{}

		err := rows.Scan(&task.Id, &task.Date, &task.Title, &task.Comment, &task.Repeat)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := rows.Err(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		tasks["tasks"] = append(tasks["tasks"], task)

	}

	// Если задач нет, возвращаем пустой json.
	if tasks["tasks"] == nil {
		tasks["tasks"] = []task.Task{}
	}

	response, err := json.Marshal(tasks)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

// TasksWithParameter возвращает список задач найденых по параметру search и статус ответа.
// Максимум 25 записей.
// search ищет соответсвие в полях title, comment, date.
// Дата подаётся в формате "02.01.2006".
func TasksWithParameter(db *sql.DB, search string) ([]task.Task, int, error) {
	var tasks []task.Task

	var date bool // Если true значит искать в поле date
	timeSearch, err := time.Parse("02.01.2006", search)
	if err == nil {
		date = true
	}

	var rows *sql.Rows

	if date {

		dateFormat := timeSearch.Format("20060102")

		rows, err = db.Query(`SELECT id, date, title, comment, repeat FROM scheduler
		WHERE date = :date LIMIT 25`,
			sql.Named("date", dateFormat))

	} else {

		rows, err = db.Query(`SELECT id, date, title, comment, repeat FROM scheduler
	WHERE title LIKE :search OR comment LIKE :search ORDER BY date LIMIT 25`,
			sql.Named("search", "%"+search+"%"))
	}

	if err != nil {
		return tasks, 500, err
	}
	defer rows.Close()

	for rows.Next() {
		task := task.Task{}

		err := rows.Scan(&task.Id, &task.Date, &task.Title, &task.Comment, &task.Repeat)
		if err != nil {
			return tasks, 500, err
		}

		if err := rows.Err(); err != nil {
			return tasks, 500, err
		}

		tasks = append(tasks, task)
	}
	return tasks, 200, nil
}
