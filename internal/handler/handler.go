// Пакет handler реализует функции для работы с api запросами.
package handler

import (
	"go_final_project/internal/date"
	"go_final_project/internal/helper"
	"go_final_project/internal/storage"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	_ "modernc.org/sqlite"
)

var ResponseStatus int

// GetNextDate вызывает функцию NextDate и возвращает её результат.
func GetNextDate(w http.ResponseWriter, req *http.Request) {

	param := req.URL.Query()

	now := param.Get("now")
	day := param.Get("date")
	repeat := param.Get("repeat")

	timeNow, err := time.Parse(date.DateFormat, now)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	nextDay, err := date.NextDate(timeNow, day, repeat)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err = w.Write([]byte(nextDay))
	if err != nil {
		logrus.WithError(err).Error("write next date")
	}
}

// TaskHandler используется для пути "/api/task" с методами Get, Post, Put, Delete.
func TaskHandler(db *storage.TasksDB) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		param := req.URL.Query().Get("id")

		var response []byte
		var err error

		switch req.Method {

		case http.MethodGet:
			response, ResponseStatus, err = helper.GetTaskID(db, param)
			if err != nil {
				http.Error(w, err.Error(), ResponseStatus)
				return
			}

		case http.MethodPost:
			response, ResponseStatus, err = helper.CheckAndAddTask(db, req)
			if err != nil {
				http.Error(w, err.Error(), ResponseStatus)
				return
			}
		case http.MethodPut:
			response, ResponseStatus, err = helper.CheckAndUpdateTask(db, req)
			if err != nil {
				http.Error(w, err.Error(), ResponseStatus)
				return
			}

		case http.MethodDelete:
			response, ResponseStatus, err = helper.DelTask(db, param)
			if err != nil {
				http.Error(w, err.Error(), ResponseStatus)
				return
			}
		}

		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		_, err = w.Write(response)
		if err != nil {
			logrus.WithError(err).Error("write response TaskHandler")
		}
	}
}

// GetTasks возвращает весь список задач.
// Если указан параметр "search", то вовзращает задачу с указанным параметром.
func GetTasks(db *storage.TasksDB) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		search := req.URL.Query().Get("search")

		response, ResponseStatus, err := helper.FindParameter(search, db)
		if err != nil {
			http.Error(w, err.Error(), ResponseStatus)
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		_, err = w.Write(response)
		if err != nil {
			logrus.WithError(err).Error("write response GetTasks")
		}
	}
}

// TaskDone обновляет следующую дату задачи по указанному id, при условии, что поле repeat не пустое.
// Если repeat пусто, то задача удаляется.
// В случае успеха функция возвращает пустой json.
func TaskDone(db *storage.TasksDB) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		id := req.URL.Query().Get("id")

		response, ResponseStatus, err := helper.SearchTaskDone(id, db)
		if err != nil {
			http.Error(w, err.Error(), ResponseStatus)
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		_, err = w.Write(response)
		if err != nil {
			logrus.WithError(err).Error("write response TaskDone")
		}
	}
}
