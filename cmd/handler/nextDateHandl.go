// Пакет handler реализует функции для работы с api запросами.
package handler

import (
	"go_final_project/cmd/date"
	"net/http"
	"time"

	_ "modernc.org/sqlite"
)

// NextDateHandl вызывает функцию NextDate и возвращает её результат.
func NextDateHandl(w http.ResponseWriter, req *http.Request) {

	param := req.URL.Query()

	now := param.Get("now")
	day := param.Get("date")
	repeat := param.Get(("repeat"))

	timeNow, err := time.Parse("20060102", now)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	nextDay, err := date.NextDate(timeNow, day, repeat)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Write([]byte(nextDay))
}
