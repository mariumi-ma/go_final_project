// Пакет storage работает с базой данных.
package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"go_final_project/internal/model"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

const FileDB = "scheduler.db"
const LimitTasks int = 25

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

// TaskID возвращает задачу по указанному id.
func (t *TasksDB) TaskID(id string) (model.Task, error) {
	var task model.Task

	row := t.DB.QueryRow("SELECT id, date, title, comment, repeat FROM scheduler WHERE id = :id",
		sql.Named("id", id))

	err := row.Scan(&task.Id, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		if err == sql.ErrNoRows {
			return task, errors.New(`{"error":"not find the task"}`)
		}
		return task, err
	}

	if err := row.Err(); err != nil {
		return task, err
	}

	return task, nil
}

// AddTask возвращает id добавленной задачи.
func (t *TasksDB) AddTask(task model.Task) (int64, error) {
	var id int64

	result, err := t.DB.Exec(`INSERT INTO scheduler (date, title, comment, repeat)
		VALUES (:date, :title, :comment, :repeat)`,
		sql.Named("date", task.Date),
		sql.Named("title", task.Title),
		sql.Named("comment", task.Comment),
		sql.Named("repeat", task.Repeat),
	)
	if err != nil {
		return id, err
	}
	id, err = result.LastInsertId()
	if err != nil {
		return id, err
	}

	return id, nil
}

func (t *TasksDB) UptadeTaskID(task model.Task) error {

	res, err := t.DB.Exec(`UPDATE scheduler SET
	date = :date, title = :title, comment = :comment, repeat = :repeat
	WHERE id = :id`,
		sql.Named("date", task.Date),
		sql.Named("title", task.Title),
		sql.Named("comment", task.Comment),
		sql.Named("repeat", task.Repeat),
		sql.Named("id", task.Id))
	if err != nil {
		return err
	}

	result, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if result == 0 {
		return errors.New(`{"error":"not found the task"}`)
	}

	return nil
}

// DeleteTask возвращает ошибку.
func (t *TasksDB) DeleteTask(id string) error {
	task, err := t.DB.Exec("DELETE FROM scheduler WHERE id = :id", sql.Named("id", id))
	if err != nil {
		return err
	}

	rowsAffected, err := task.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New(`{"error":"not found the task"}`)
	}
	return nil
}

// SearchString возвращает записи с указанным параметром.
func (t *TasksDB) SearchString(search string) ([]model.Task, error) {
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
			if err == sql.ErrNoRows {
				return tasks, errors.New(`{"error":"not find the task"}`)
			} else {
				return tasks, err
			}
		}

		if err := rows.Err(); err != nil {
			return tasks, err
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

// SearchDate возвращает записи с указанной датой.
func (t *TasksDB) SearchDate(date string) ([]model.Task, error) {
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
			if err == sql.ErrNoRows {
				return tasks, errors.New(`{"error":"not find the task"}`)
			} else {
				return tasks, err
			}
		}

		if err := rows.Err(); err != nil {
			return tasks, err
		}

		tasks = append(tasks, task)
	}
	return tasks, nil
}

// FindTaskDone возвращает запись найденную по id.
func (t *TasksDB) FindTaskDone(id string) (model.Task, error) {
	var task model.Task

	row := t.DB.QueryRow("SELECT id, date, title, comment, repeat FROM scheduler WHERE id = :id",
		sql.Named("id", id))

	err := row.Scan(&task.Id, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		if err == sql.ErrNoRows {
			return task, errors.New(`{"error":"not find the task"}`)
		}
		return task, err
	}

	if err := row.Err(); err != nil {
		return task, err
	}
	return task, nil
}

// UpdateDateTaskDone обновляет следующую дату.
func (t *TasksDB) UpdateDateTaskDone(date, id string) error {

	res, err := t.DB.Exec(`UPDATE scheduler SET date = :date WHERE id = :id`,
		sql.Named("date", date),
		sql.Named("id", id))
	if err != nil {
		return err
	}

	result, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if result == 0 {
		return errors.New(`{"error":"not found the task"}`)
	}

	return nil
}

// FindTasks возвращает все записи. Лимит указан в константе LimitTasks.
func (t *TasksDB) FindTasks() ([]model.Task, error) {
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
			if err == sql.ErrNoRows {
				return tasks, errors.New(`{"error":"not find the task"}`)
			} else {
				return tasks, err
			}
		}

		if err := rows.Err(); err != nil {
			return tasks, err
		}

		tasks = append(tasks, task)
	}
	return tasks, nil
}
