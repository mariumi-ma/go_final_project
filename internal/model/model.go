// Пакет model содержит сущность типа Task и функцию валидации данных типа Task.
package model

import (
	"bytes"
	"encoding/json"
	"errors"
	"go_final_project/internal/date"
	"net/http"
	"time"
)

type Task struct {
	Id      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

// CheckTask проверяет корректность заполненых данных. Возвращает задачу и статус ответа.
// Если дата текущего дня, то task.Date присваивается текущий день, т.е. напоминание
// начинает отсчёт от сегоднянего дня.
func CheckTask(req *http.Request) (Task, int, error) {
	var task Task
	var buf bytes.Buffer

	_, err := buf.ReadFrom(req.Body)
	if err != nil {
		return task, http.StatusInternalServerError, err
	}

	if err = json.Unmarshal(buf.Bytes(), &task); err != nil {
		return task, http.StatusInternalServerError, err
	}
	// Поле title обязательно должно быть указано, иначе возвращаем ошибку.
	if task.Title == "" {
		return task, http.StatusBadRequest, errors.New(`{"error":"task title is not specified"}`)
	}

	now := time.Now()
	// Приводим время к нулям, чтобы корректно использовать ф. 'dateParse.Before(now)'
	now = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)

	// Если дата не указана, полю date присваивается сегодняшняя дата.
	if task.Date == "" {
		task.Date = now.Format(date.DateFormat)
	}

	dateParse, err := time.Parse(date.DateFormat, task.Date)
	if err != nil {
		return task, http.StatusBadRequest, errors.New(`{"error":"incorrect date"}`)
	}
	var dateNew string
	if task.Repeat != "" {
		dateNew, err = date.NextDate(now, task.Date, task.Repeat) // Проверяем корректность поля repeat
		if err != nil {
			return task, http.StatusBadRequest, err
		}
	}

	// Если поле date равен текущему дню, то date присваивается сегодняшний день.
	if task.Date == now.Format(date.DateFormat) {
		task.Date = now.Format(date.DateFormat)
	}

	// Если дата раньше сегодняшней, есть два варианта:
	// 1. Если поле repeat пусто, то полю date присваиваетя сегодняшняя дата.
	// 2. Иначе полю date присваиваетя следующая дата повторения, высчитанная ранее ф. NextDate.
	if dateParse.Before(now) {
		if task.Repeat == "" {
			task.Date = now.Format(date.DateFormat)
		} else {
			task.Date = dateNew
		}
	}

	return task, http.StatusOK, nil
}
