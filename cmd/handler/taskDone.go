package handler

import (
	"database/sql"
	"encoding/json"
	"go_final_project/cmd/date"
	"go_final_project/cmd/task"
	"net/http"
	"time"
)

// TaskDone обновляет следующую дату задачи по указанному id, при условии, что поле repeat не пустое.
// Если repeat пусто, то задача удаляется.
// В случае успеха функция возвращает пустой json.
func TaskDone(w http.ResponseWriter, req *http.Request) {
	id := req.URL.Query().Get("id")
	var taskID task.Task

	db, err := sql.Open("sqlite", "scheduler.db")
	if err != nil {
		http.Error(w, "error opening database"+err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	row := db.QueryRow("SELECT * FROM scheduler WHERE id = :id",
		sql.Named("id", id))

	err = row.Scan(&taskID.Id, &taskID.Date, &taskID.Title, &taskID.Comment, &taskID.Repeat)
	if err != nil {
		http.Error(w, `{"error":"writing date"}`+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := row.Err(); err != nil {
		http.Error(w, `{"error":"writing date"}`+err.Error(), http.StatusInternalServerError)
		return
	}
	// Проверяем поле repeat.
	if taskID.Repeat == "" {
		ResponseStatus, err := DeleteTask(db, id)
		if err != nil {
			http.Error(w, err.Error(), ResponseStatus)
			return
		}
	} else { // Если репит не пусто обновляем след дату
		now := time.Now()
		dataNew, err := date.NextDate(now, taskID.Date, taskID.Repeat)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		res, err := db.Exec(`UPDATE scheduler SET date = :date WHERE id = :id`,
			sql.Named("date", dataNew),
			sql.Named("id", taskID.Id))
		if err != nil {
			http.Error(w, `{"error":"task is not found" }`+err.Error(), http.StatusInternalServerError)
			return
		}

		result, err := res.RowsAffected()
		if err != nil {
			http.Error(w, `{"error":"task is not found" }`+err.Error(), http.StatusInternalServerError)
			return
		}
		if result == 0 {
			http.Error(w, `{"error":"task is not found"}`, http.StatusBadRequest)
			return
		}
	}

	str := map[string]interface{}{}
	response, err := json.Marshal(str)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}
