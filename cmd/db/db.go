// Пакет db работает с базой данных.
package db

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

// CheckAndOpenDB проверяет существует ли в директории приложения файл scheduler.db.
// Если файла нет, то функция создаёт файл с таблицей scheduler.
func CheckAndOpenDB() {

	appPath, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}

	dbFile := filepath.Join(filepath.Dir(appPath), "scheduler.db")
	_, err = os.Stat(dbFile)

	//Если install == true, то выполнится sql-запрос с CREATE TABLE.
	var install bool
	if err != nil {
		install = true
	}

	DB, err = sql.Open("sqlite", dbFile)
	if err != nil {
		log.Println(err)
	}
	defer DB.Close()

	if install {
		statement, err := DB.Prepare(`CREATE TABLE IF NOT EXISTS scheduler 
		(id INTEGER PRIMARY KEY AUTOINCREMENT,
		date CHAR(8) NOT NULL DEFAULT "",
   		title VARCHAR(128) NOT NULL DEFAULT "",
   		comment TEXT NOT NULL DEFAULT "",
  		repeat VARCHAR(128) NOT NULL DEFAULT "");

		CREATE INDEX IF NOT EXISTS date_indx ON scheduler (date);
		`)
		if err != nil {
			log.Println("Error create db", err)
		}
		statement.Exec()
	}
}
