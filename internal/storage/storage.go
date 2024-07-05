// Пакет storage работает с базой данных.
package storage

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"go_final_project/internal/model"
	"net/http"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

const FileDB = "scheduler.db"
const LimitTasks int = 25

type ResponseForPostTask struct {
	Id int64 `json:"id"`
}

type TasksDB struct {
	DB *sql.DB
}

func NewTasksDB(db *sql.DB) *TasksDB {
	return &TasksDB{DB: db}
}

// CheckDB проверяет существует ли в директории приложения файл scheduler.db.
// Если файла нет, то функция создаёт файл с таблицей scheduler.
func CheckDB() error {

	appPath, err := os.Executable()
	if err != nil {
		return err
	}

	dbFile := filepath.Join(filepath.Dir(appPath), FileDB)
	_, err = os.Stat(dbFile)

	//Если install == true, то выполнится sql-запрос с CREATE TABLE.
	var install bool
	if err != nil {
		install = true
	}

	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		return err
	}
	defer db.Close()

	if install {
		statement, err := db.Prepare(`CREATE TABLE IF NOT EXISTS scheduler 
		(id INTEGER PRIMARY KEY AUTOINCREMENT,
		date CHAR(8) NOT NULL DEFAULT "",
   		title VARCHAR(128) NOT NULL DEFAULT "",
   		comment TEXT NOT NULL DEFAULT "",
  		repeat VARCHAR(128) NOT NULL DEFAULT "");

		CREATE INDEX IF NOT EXISTS date_indx ON scheduler (date);
		`)
		if err != nil {
			return fmt.Errorf("error create db. %v", err)
		}
		statement.Exec()
	}
	return nil
}

// TaskID возвращает задачу по указанному id и статус ответа.
func (t *TasksDB) TaskID(id string) ([]byte, error) {
	var task model.Task

	row := t.DB.QueryRow("SELECT id, date, title, comment, repeat FROM scheduler WHERE id = :id",
		sql.Named("id", id))

	err := row.Scan(&task.Id, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		return []byte{}, fmt.Errorf(`{"error":"writing data %v"}`, err)
	}

	if err := row.Err(); err != nil {
		return []byte{}, fmt.Errorf(`{"error":"writing data %v"}`, err)
	}

	result, err := json.Marshal(task)
	if err != nil {
		return []byte{}, err
	}

	return result, nil
}

// AddTask возвращает id добавленной задачи и статус ответа.
func (t *TasksDB) AddTask(task model.Task) ([]byte, error) {
	var resp ResponseForPostTask

	result, err := t.DB.Exec(`INSERT INTO scheduler (date, title, comment, repeat)
		VALUES (:date, :title, :comment, :repeat)`,
		sql.Named("date", task.Date),
		sql.Named("title", task.Title),
		sql.Named("comment", task.Comment),
		sql.Named("repeat", task.Repeat),
	)
	if err != nil {
		return []byte{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return []byte{}, err
	}

	resp.Id = id

	idResult, err := json.Marshal(resp)
	if err != nil {
		return []byte{}, err
	}
	return idResult, nil
}

// UptadeTaskID возвращает пустой json в случаее успешного обновления данных и статус ответа.
func (t *TasksDB) UptadeTaskID(task model.Task) ([]byte, int, error) {

	res, err := t.DB.Exec(`UPDATE scheduler SET
	date = :date, title = :title, comment = :comment, repeat = :repeat
	WHERE id = :id`,
		sql.Named("date", task.Date),
		sql.Named("title", task.Title),
		sql.Named("comment", task.Comment),
		sql.Named("repeat", task.Repeat),
		sql.Named("id", task.Id))
	if err != nil {
		return []byte{}, http.StatusInternalServerError, fmt.Errorf(`{"error":"task is not found" %s}`, err)
	}

	result, err := res.RowsAffected()
	if err != nil {
		return []byte{}, http.StatusInternalServerError, fmt.Errorf(`{"error":"task is not found" %s}`, err)
	}
	if result == 0 {
		return []byte{}, http.StatusBadRequest, fmt.Errorf(`{"error":"task is not found"}`)
	}
	var str model.Task
	response, err := json.Marshal(str)
	if err != nil {
		return []byte{}, http.StatusInternalServerError, err
	}

	return response, http.StatusOK, nil
}

// DeleteTask возвращает статус ответа.
func (t *TasksDB) DeleteTask(id string) (int, error) {
	task, err := t.DB.Exec("DELETE FROM scheduler WHERE id = :id", sql.Named("id", id))
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf(`{"error":"%s"}`, err)
	}

	rowsAffected, err := task.RowsAffected()
	if err != nil {
		return http.StatusInternalServerError, err
	}

	if rowsAffected == 0 {
		return http.StatusBadRequest, errors.New(`{"error":"not found the task"}`)
	}
	return http.StatusOK, nil
}

// TasksWithParameterString возвращает записи с указанным параметром.
func (t *TasksDB) TasksWithParameterString(search string) ([]model.Task, error) {
	var tasks []model.Task
	rows, err := t.DB.Query(`SELECT id, date, title, comment, repeat FROM scheduler
	WHERE title LIKE :search OR comment LIKE :search ORDER BY date LIMIT :limit`,
		sql.Named("search", "%"+search+"%"),
		sql.Named("limit", LimitTasks))
	if err != nil {
		return tasks, err
	}
	defer rows.Close()

	for rows.Next() {
		task := model.Task{}

		err := rows.Scan(&task.Id, &task.Date, &task.Title, &task.Comment, &task.Repeat)
		if err != nil {
			return tasks, err
		}

		if err := rows.Err(); err != nil {
			return tasks, err
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

// TasksWithParameterDate возвращает записи с указанной датой.
func (t *TasksDB) TasksWithParameterDate(date string) ([]model.Task, error) {
	var tasks []model.Task

	rows, err := t.DB.Query(`SELECT id, date, title, comment, repeat FROM scheduler
	WHERE date = :date LIMIT :limit`,
		sql.Named("date", date),
		sql.Named("limit", LimitTasks))

	if err != nil {
		return tasks, err
	}
	defer rows.Close()

	for rows.Next() {
		task := model.Task{}

		err := rows.Scan(&task.Id, &task.Date, &task.Title, &task.Comment, &task.Repeat)
		if err != nil {
			return tasks, err
		}

		if err := rows.Err(); err != nil {
			return tasks, err
		}

		tasks = append(tasks, task)
	}
	return tasks, nil
}

// QueryTaskDone записывает в task данные из базы данных по указанному id.
func (t *TasksDB) QueryTaskDone(id string) (model.Task, error) {
	var task model.Task

	row := t.DB.QueryRow("SELECT id, date, title, comment, repeat FROM scheduler WHERE id = :id",
		sql.Named("id", id))

	err := row.Scan(&task.Id, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		return task, fmt.Errorf(`{"error":"writing date"} %v`, err)
	}

	if err := row.Err(); err != nil {
		return task, fmt.Errorf(`{"error":"writing date"} %v`, err)
	}
	return task, nil
}

// UpdateDateTaskDone обновляет следующую дату. Возвращает статус ответа.
func (t *TasksDB) UpdateDateTaskDone(date, id string) (int, error) {

	res, err := t.DB.Exec(`UPDATE scheduler SET date = :date WHERE id = :id`,
		sql.Named("date", date),
		sql.Named("id", id))
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf(`{"error":"task is not found"}%v`, err)
	}

	result, err := res.RowsAffected()
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf(`{"error":"task is not found"}%v`, err)
	}
	if result == 0 {
		return http.StatusBadRequest, fmt.Errorf(`{"error":"task is not found"}%v`, err)
	}

	return http.StatusOK, nil
}

// QueryAllTasks возвращает все записи. Лимит указан в константе LimitTasks.
func (t *TasksDB) QueryAllTasks() ([]model.Task, error) {
	var tasks []model.Task
	rows, err := t.DB.Query(`SELECT id, date, title, comment, repeat FROM scheduler
	ORDER BY date LIMIT :limit`,
		sql.Named("limit", LimitTasks))
	if err != nil {
		return tasks, err
	}
	defer rows.Close()

	for rows.Next() {
		task := model.Task{}

		err := rows.Scan(&task.Id, &task.Date, &task.Title, &task.Comment, &task.Repeat)
		if err != nil {
			return tasks, err
		}

		if err := rows.Err(); err != nil {
			return tasks, err
		}

		tasks = append(tasks, task)
	}
	return tasks, nil
}
