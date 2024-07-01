package handler

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"go_final_project/cmd/date"
	"go_final_project/cmd/task"
	"net/http"
	"time"
)

type ResponseForPostTask struct {
	Id int64 `json:"id"`
}

var ResponseStatus int

// TaskHandler используется для пути "/api/task" с методами Get, Post, Put, Delete.
func TaskHandler(w http.ResponseWriter, req *http.Request) {
	param := req.URL.Query().Get("id")

	db, err := sql.Open("sqlite", "scheduler.db")
	if err != nil {
		http.Error(w, "error opening database"+err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var response []byte

	switch req.Method {

	case http.MethodGet:
		if param == "" {
			http.Error(w, `{"error":"inncorect id"}`, http.StatusBadRequest)
			return
		}
		response, ResponseStatus, err = TaskID(db, param)
		defer db.Close()
		if err != nil {
			http.Error(w, err.Error(), ResponseStatus)
			return
		}

	case http.MethodPost:
		response, ResponseStatus, err = AddTask(db, req)
		defer db.Close()
		if err != nil {
			http.Error(w, err.Error(), ResponseStatus)
			return
		}
	case http.MethodPut:
		response, ResponseStatus, err = UptadeTaskID(db, req)
		defer db.Close()
		if err != nil {
			http.Error(w, err.Error(), ResponseStatus)
			return
		}
	case http.MethodDelete:
		ResponseStatus, err = DeleteTask(db, param)
		if err != nil {
			http.Error(w, err.Error(), ResponseStatus)
			return
		}
		// если прошло всё успешно, то возвращаем пустой json
		str := map[string]interface{}{}
		response, err = json.Marshal(str)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

// TaskID возвращает задачу по указанному id и статус ответа.
func TaskID(db *sql.DB, id string) ([]byte, int, error) {
	var task task.Task

	row := db.QueryRow("SELECT * FROM scheduler WHERE id = :id",
		sql.Named("id", id))

	err := row.Scan(&task.Id, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		return []byte{}, 500, fmt.Errorf(`{"error":"writing data %v"}`, err)
	}

	if err := row.Err(); err != nil {
		return []byte{}, 500, fmt.Errorf(`{"error":"writing data %v"}`, err)
	}

	result, err := json.Marshal(task)
	if err != nil {
		return []byte{}, 500, err
	}

	return result, 400, nil
}

// AddTask возвращает id добавленной задачи и статус ответа.
func AddTask(db *sql.DB, req *http.Request) ([]byte, int, error) {
	var resp ResponseForPostTask

	task, ResponseStatus, err := CheckTask(req)
	if err != nil {
		return []byte{}, ResponseStatus, err
	}

	result, err := db.Exec(`INSERT INTO scheduler (date, title, comment, repeat)
		VALUES (:date, :title, :comment, :repeat)`,
		sql.Named("date", task.Date),
		sql.Named("title", task.Title),
		sql.Named("comment", task.Comment),
		sql.Named("repeat", task.Repeat),
	)
	if err != nil {
		return []byte{}, 500, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return []byte{}, 500, err
	}

	resp.Id = id

	idResult, err := json.Marshal(resp)
	if err != nil {
		return []byte{}, 500, err
	}
	return idResult, 200, nil
}

// CheckTask проверяет корректность заполненых данных. Возвращает задачу и статус ответа.
// Если дата текущего дня, то task.Date присваивается текущий день, т.е. напоминание
// начинает отсчёт от сегоднянего дня.
func CheckTask(req *http.Request) (task.Task, int, error) {
	var task task.Task
	var buf bytes.Buffer

	_, err := buf.ReadFrom(req.Body)
	if err != nil {
		return task, 500, err
	}

	if err = json.Unmarshal(buf.Bytes(), &task); err != nil {
		return task, 500, err
	}
	// Поле title обязательно должно быть указано, иначе возвращаем ошибку.
	if task.Title == "" {
		return task, 400, errors.New(`{"error":"task title is not specified"}`)
	}

	now := time.Now()
	now = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)

	// Если дата не указана, полю date присваивается сегодняшняя дата.
	if task.Date == "" {
		task.Date = now.Format("20060102")
	}

	dateParse, err := time.Parse("20060102", task.Date)
	if err != nil {
		return task, 400, errors.New(`{"error":"incorrect date"}`)
	}
	var dateNew string
	if task.Repeat != "" {
		dateNew, err = date.NextDate(now, task.Date, task.Repeat) // Проверяем корректность поля repeat
		if err != nil {
			return task, 400, err
		}
	}

	// Если поле date равен текущему дню, то date присваивается сегодняшний день.
	if task.Date == now.Format("20060102") {
		task.Date = now.Format("20060102")
	}

	// Если дата раньше сегодняшней, есть два варианта:
	// 1. Если поле repeat пусто, то полю date присваиваетя сегодняшняя дата.
	// 2. Иначе полю date присваиваетя следующая дата повторения, высчитанная ранее ф. NextDate.
	if dateParse.Before(now) {
		if task.Repeat == "" {
			task.Date = now.Format("20060102")
		} else {
			task.Date = dateNew
		}
	}

	return task, 200, nil
}

// UptadeTaskID возвращает пустой json в случаее успешного обновления данных и статус ответа.
func UptadeTaskID(db *sql.DB, req *http.Request) ([]byte, int, error) {

	taskID, ResponseStatus, err := CheckTask(req)
	if err != nil {
		return []byte{}, ResponseStatus, err
	}

	res, err := db.Exec(`UPDATE scheduler SET
	date = :date, title = :title, comment = :comment, repeat = :repeat
	WHERE id = :id`,
		sql.Named("date", taskID.Date),
		sql.Named("title", taskID.Title),
		sql.Named("comment", taskID.Comment),
		sql.Named("repeat", taskID.Repeat),
		sql.Named("id", taskID.Id))
	if err != nil {
		return []byte{}, 500, fmt.Errorf(`{"error":"task is not found" %s}`, err)
	}

	result, err := res.RowsAffected()
	if err != nil {
		return []byte{}, 500, fmt.Errorf(`{"error":"task is not found" %s}`, err)
	}
	if result == 0 {
		return []byte{}, 400, fmt.Errorf(`{"error":"task is not found"}`)
	}
	var str task.Task
	response, err := json.Marshal(str)
	if err != nil {
		return []byte{}, 500, err
	}

	return response, 200, nil
}

// DeleteTask возвращает статус ответа
func DeleteTask(db *sql.DB, id string) (int, error) {
	task, err := db.Exec("DELETE FROM scheduler WHERE id = :id", sql.Named("id", id))
	if err != nil {
		return 500, fmt.Errorf(`{"error":"%s"}`, err)
	}

	rowsAffected, err := task.RowsAffected()
	if err != nil {
		return 500, err
	}

	if rowsAffected == 0 {
		return 400, errors.New(`{"error":"not found the task"}`)
	}
	return 200, nil
}
