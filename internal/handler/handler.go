// Пакет handler реализует функции для работы с api запросами.
package handler

import (
	"encoding/json"
	"go_final_project/internal/date"
	"go_final_project/internal/model"
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
	repeat := param.Get(("repeat"))

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
			if param == "" {
				http.Error(w, `{"error":"inncorect id"}`, http.StatusBadRequest)
				return
			}
			response, err = db.TaskID(param)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

		case http.MethodPost:
			task, ResponseStatus, err := model.CheckTask(req)
			if err != nil {
				http.Error(w, err.Error(), ResponseStatus)
				return
			}

			response, err = db.AddTask(task)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

		case http.MethodPut:
			task, ResponseStatus, err := model.CheckTask(req)
			if err != nil {
				http.Error(w, err.Error(), ResponseStatus)
				return
			}

			response, ResponseStatus, err = db.UptadeTaskID(task)
			if err != nil {
				http.Error(w, err.Error(), ResponseStatus)
				return
			}

		case http.MethodDelete:
			ResponseStatus, err = db.DeleteTask(param)
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
		_, err = w.Write(response)
		if err != nil {
			logrus.WithError(err).Error("write response TaskHandler")
		}
	}
}

// GetTasks возвращает весь список задач, максимум строк указан в перемненной LimitTasks.
// Если указан параметр "search", то вовзращает задачу с указанным параметром.
// search ищет соответсвие в полях title, comment, date.
// Дата подаётся в формате "02.01.2006".
func GetTasks(db *storage.TasksDB) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		// Создаем tasks для ответа в json.
		tasks := make(map[string][]model.Task)

		search := req.URL.Query().Get("search")

		// Параметр указан, значит ищем соответсвие по параметру.
		if search != "" {

			date, isDate := date.IsDate(search)

			switch isDate {
			case true:
				allTasks, err := db.TasksWithParameterDate(date)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				tasks["tasks"] = allTasks

			case false:
				allTasks, err := db.TasksWithParameterString(search)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				tasks["tasks"] = allTasks
			}
			// Параметр не указан, значит выводим все записи.
		} else {
			allTasks, err := db.QueryAllTasks()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			tasks["tasks"] = allTasks
		}

		// Если задач нет, возвращаем пустой json.
		if tasks["tasks"] == nil {
			tasks["tasks"] = []model.Task{}
		}

		response, err := json.Marshal(tasks)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
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

		task, err := db.QueryTaskDone(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Проверяем поле repeat.
		if task.Repeat == "" {
			ResponseStatus, err := db.DeleteTask(id)
			if err != nil {
				http.Error(w, err.Error(), ResponseStatus)
				return
			}
		} else {
			now := time.Now()
			dateNew, err := date.NextDate(now, task.Date, task.Repeat)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			ResponseStatus, err := db.UpdateDateTaskDone(dateNew, id)
			if err != nil {
				http.Error(w, err.Error(), ResponseStatus)
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
		_, err = w.Write(response)
		if err != nil {
			logrus.WithError(err).Error("write response TaskDone")
		}
	}
}
