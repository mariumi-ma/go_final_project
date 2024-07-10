// Пакет helper реализует функции для обработки параметров полученных от клиента.
package helper

import (
	"encoding/json"
	"errors"
	"go_final_project/internal/date"
	"go_final_project/internal/model"
	"go_final_project/internal/storage"
	"net/http"
	"time"
)

type ResponseForPostTask struct {
	Id int64 `json:"id"`
}

// GetTaskID обрабатывает параметр id. Серилизует и возвращает запись по указанному id.
func GetTaskID(db *storage.TasksDB, id string) ([]byte, int, error) {

	if id == "" {
		return []byte{}, http.StatusBadRequest, errors.New(`{"error":"inncorect id"}`)
	}

	task, ResponseStatus, err := db.TaskID(id)
	if err != nil {
		return []byte{}, ResponseStatus, err
	}

	response, err := json.Marshal(task)
	if err != nil {
		return []byte{}, http.StatusInternalServerError, err
	}

	return response, http.StatusOK, nil
}

// FindParameter возвращает весь список зада и статус ответа.
// Максимум строк указан в перемненной LimitTasks.
// Если указан параметр "search", то вовзращает задачу с указанным параметром.
// search ищет соответсвие в полях title, comment, date.
// Дата подаётся в формате "02.01.2006".
func FindParameter(search string, db *storage.TasksDB) ([]byte, int, error) {
	// Создаем map для ответа клиенту.
	tasks := make(map[string][]model.Task)

	if search != "" {

		date, isDate := date.IsDate(search)

		switch isDate {
		case true:
			allTasks, ResponseStatus, err := db.SearchDate(date)
			if err != nil {
				return []byte{}, ResponseStatus, err
			}
			tasks["tasks"] = allTasks

		case false:
			allTasks, ResponseStatus, err := db.SearchString(search)
			if err != nil {
				return []byte{}, ResponseStatus, err
			}
			tasks["tasks"] = allTasks
		}
		// Параметр не указан, значит выводим все записи.
	} else {
		allTasks, ResponseStatus, err := db.FindTasks()
		if err != nil {
			return []byte{}, ResponseStatus, err
		}
		tasks["tasks"] = allTasks
	}

	// Если задач нет, возвращаем пустой json.
	if tasks["tasks"] == nil {
		tasks["tasks"] = []model.Task{}
	}

	response, err := json.Marshal(tasks)
	if err != nil {
		return []byte{}, http.StatusInternalServerError, err
	}

	return response, http.StatusOK, nil
}

// SearchTaskDone обновляет следующую дату задачи по указанному id, при условии, что поле repeat не пустое.
// Если repeat пусто, то задача удаляется.
// Возвращает ответ статуса и, в случае успеха, пустой json.
func SearchTaskDone(id string, db *storage.TasksDB) ([]byte, int, error) {

	task, ResponseStatus, err := db.FindTaskDone(id)
	if err != nil {
		return []byte{}, ResponseStatus, err
	}

	// Проверяем поле repeat.
	if task.Repeat == "" {
		ResponseStatus, err := db.DeleteTask(id)
		if err != nil {
			return []byte{}, ResponseStatus, err
		}
	} else {
		now := time.Now()
		dateNew, err := date.NextDate(now, task.Date, task.Repeat)
		if err != nil {
			return []byte{}, http.StatusBadRequest, err
		}

		ResponseStatus, err := db.UpdateDateTaskDone(dateNew, id)
		if err != nil {
			return []byte{}, ResponseStatus, err
		}
	}

	str := map[string]interface{}{}
	response, err := json.Marshal(str)
	if err != nil {
		return []byte{}, http.StatusInternalServerError, err
	}

	return response, http.StatusOK, nil
}

// CheckAndUpdateTask проверяет валидность данных и обновляяет запись по указанному id.
// Возвращает пустой json в случае успеха и статус ответа.
func CheckAndUpdateTask(db *storage.TasksDB, req *http.Request) ([]byte, int, error) {
	task, ResponseStatus, err := model.CheckTask(req)
	if err != nil {
		return []byte{}, ResponseStatus, err
	}

	ResponseStatus, err = db.UptadeTaskID(task)
	if err != nil {
		return []byte{}, ResponseStatus, err
	}

	// Создаем пустой объект для ответа в json.
	var str model.Task
	response, err := json.Marshal(str)
	if err != nil {
		return []byte{}, http.StatusInternalServerError, err
	}

	return response, http.StatusOK, nil
}

// CheckAndAddTask проверят валидность данных и добавляет запись в базу данных.
// Возвращает id записи, статуст ответа.
func CheckAndAddTask(db *storage.TasksDB, req *http.Request) ([]byte, int, error) {
	var resp ResponseForPostTask

	task, ResponseStatus, err := model.CheckTask(req)
	if err != nil {
		return []byte{}, ResponseStatus, err
	}

	resp.Id, err = db.AddTask(task)
	if err != nil {
		return []byte{}, http.StatusInternalServerError, err
	}

	response, err := json.Marshal(resp)
	if err != nil {
		return []byte{}, http.StatusInternalServerError, err
	}

	return response, http.StatusOK, nil
}

// DelTask удаляет запись по указанному id. Возвращает пустой json и статус ответа.
func DelTask(db *storage.TasksDB, id string) ([]byte, int, error) {
	ResponseStatus, err := db.DeleteTask(id)
	if err != nil {
		return []byte{}, ResponseStatus, err
	}

	// если прошло всё успешно, то возвращаем пустой json
	str := map[string]interface{}{}
	response, err := json.Marshal(str)
	if err != nil {
		return []byte{}, http.StatusInternalServerError, err
	}
	return response, http.StatusOK, nil
}
