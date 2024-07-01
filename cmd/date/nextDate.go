// Пакет date реализует функции вычисления следующей даты для задачи
// в соответствии с указанным правилом.
package date

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

// MonthsDays устанавливает количество дней в каждом месяце.
var MonthsDays = map[int]int{
	1:  31, // Jan
	2:  28, // Feb
	3:  31, // Mar
	4:  30, // Apr
	5:  31, // May
	6:  30, // Jun
	7:  31, // Jul
	8:  31, // Aug
	9:  30, // Sep
	10: 31, // Oct
	11: 30, // Nov
	12: 31, // Dec
}

// LeapYear определяет високосный год или нет. Если true - год високосный.
func LeapYear(year int) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}

// DayLessZero возвращает последний или предпоследний день месяца.
func DayLessZero(year, month, day int) int {

	february := 2
	if month == february {

		if LeapYear(year) {
			MonthsDays[month] = 29
		}
	}

	if day == 40 { // возвращаем последний день месяца
		return MonthsDays[month]
	}

	return MonthsDays[month] - 1 // возвращаем предпоследний день месяца
}

// ConvertInt преобразовывает даты и месяцы из string в тип int.
// str - список дней или месяцев;
// days - если передать true - это дни, иначе месяцы.
func ConvertInt(str []string, days bool) ([]int, error) {

	//daysInt содержит преобразованные даты или месяцы в типе int.
	daysInt := make([]int, 0)

	for _, day := range str {
		dayInt, err := strconv.Atoi(day)
		if err != nil {
			return []int{}, err
		}

		// Проверяем это месяц или день.
		switch days {
		case true:
			if dayInt > 31 || dayInt < -2 {
				return []int{}, errors.New(`{"error":"incorrect day of month"}`)
			}
			// Если день дан со знаком минус, то присваиваем новое значение dayInt
			// для дальнейшей корректной сортировки, т.к. -1 и -2 это последние дни месяца.
			if dayInt == -1 {
				dayInt = 40 // Последний день месяца.
			} else if dayInt == -2 {
				dayInt = 35 // Предпоследний день месяца.
			}

		case false:
			if dayInt > 12 || dayInt <= 0 {
				return []int{}, errors.New(`{"error":"incorrect month"}`)
			}
		}
		daysInt = append(daysInt, dayInt)
	}
	sort.Ints(daysInt)
	return daysInt, nil
}

// NextDate высчитывает и возвращает следующую дату задачи. Возвращаемая дата должна быть
// больше даты, указанной в переменной now.
// now - время от которого ищется ближайшая дата;
// date - исходное время в формате '20060102', от которого начинается отсчёт повторений;
// repeat - правило повторения. Может быть в следующем формате:
//
// d <число> — задача переносится на указанное число дней. Максимальное допустимое число равно 400.
// y — задача выполняется ежегодно. Указывается только символ.
// w <через запятую от 1 до 7> — задача назначается в указанные дни недели,
// где 1 — понедельник, 7 — воскресенье.
// m <через запятую от 1 до 31,-1,-2> [через запятую от 1 до 12] — задача назначается в
// указанные дни месяца. При этом вторая последовательность чисел опциональна и указывает
// на определённые месяцы. Если месяцы не указаны, то задача выполняется в указанные дни
// каждый месяц.
// '-1' - последний день месяца.
// '-2' - предпоследний день месяца.
func NextDate(now time.Time, date string, repeat string) (string, error) {

	dateParse, err := time.Parse("20060102", date)
	if err != nil {
		return "", err
	}

	if repeat == "" {
		return "", errors.New(`{"error":"'repeat' is empty"}`)
	}

	// Отделяем символы от дат и месяцев
	yearMonthDay := strings.Split(repeat, " ")
	symbolRepeat := yearMonthDay[0]

	if !strings.ContainsAny(symbolRepeat, "dywm") {
		return "", fmt.Errorf(`{"error":"incorrect symbol %s"}`, symbolRepeat)
	}

	switch symbolRepeat {
	case "y":
		next := dateParse

	loop0:
		for {
			if next.After(now) && next.After(dateParse) { // проверяем, чтобы следующая дата была больше now
				break loop0
			}
			next = next.AddDate(1, 0, 0)
		}

		return next.Format("20060102"), nil

	case "d":
		if len(yearMonthDay) != 2 { // проверяем, чтобы были указаны дни повторений
			return "", errors.New(`{"error":"the day is not specified"}`)
		}
		dayRepeat := yearMonthDay[1]
		dayInt, err := strconv.Atoi(dayRepeat)
		if err != nil {
			return "", err
		}

		if dayInt > 400 {
			return "", errors.New(`{"error":"the maximum allowed interval is exceeded"}`)
		}

		next := dateParse.AddDate(0, 0, dayInt)
		// в цикле высчитываем дату, пока она не будет больше переменной now
		for {
			if next.After(now) {
				break
			}
			next = next.AddDate(0, 0, dayInt)
		}

		return next.Format("20060102"), nil

	case "w":
		if len(yearMonthDay) != 2 { // проверяем, чтобы были указаны дни недели
			return "", errors.New(`{"error":"the day is not specified"}`)
		}

		daysOfWeek := strings.Split(yearMonthDay[1], ",")

		for _, day := range daysOfWeek {
			if !strings.ContainsAny(day, "1234567") {
				return "", errors.New(`{"error":"incorrect day of the week"}`)
			}
		}
		// week определяет дни недели. Т.к. в константе типа Weekday воскреснье это 0,
		// но в нашем случае воскрененье = 7.
		// Переменная week позволяет корректно сравнивать дни недели.
		week := map[string]int{
			"1": 1, //Monday
			"2": 2, //Tuesday
			"3": 3, //Wednesday
			"4": 4, //Thursday
			"5": 5, //Friday
			"6": 6, //Saturday
			"7": 0, //Sunday
		}

		next := dateParse.AddDate(0, 0, 1)

	loop1:
		for {
			for _, day := range daysOfWeek {
				if next.Weekday() == time.Weekday(week[day]) && next.After(now) {
					break loop1
				}
			}
			next = next.AddDate(0, 0, 1)
		}

		return next.Format("20060102"), nil

	case "m":
		if len(yearMonthDay) < 2 {
			return "", errors.New(`{"error":"the day is not specified"}`)
		}

		// Это условие определения, что даны только числа повторений задачи, т.е. ежемесячный повтор.
		if len(yearMonthDay) == 2 {
			daysRepeat := strings.Split(yearMonthDay[1], ",")

			daysInt, err := ConvertInt(daysRepeat, true)
			if err != nil {
				return "", err
			}

			year, month, _ := dateParse.Date()
			next := dateParse

		loop2:
			for {
				for _, dayNew := range daysInt {

					if dayNew == 35 || dayNew == 40 {
						dayNew = DayLessZero(year, int(month), dayNew)
					}

					// Если этой даты нет в месяце (например 31-е), то задача переходит
					// на следующий месяц и т.д. в цикле.
					if MonthsDays[int(month)] < dayNew {
						for {
							next = time.Date(year, month+1, dayNew, 0, 0, 0, 0, time.Local)
							if next.After(now) {
								break loop2
							}
						}
					}

					next = time.Date(year, month, dayNew, 0, 0, 0, 0, time.Local)
					if next.After(now) {
						break loop2
					}
				}
				// Если в декабре нет подходящей даты, то задача переносится на январь следующего года
				if month == 12 {
					year += 1
					month = 1
				} else {
					month++
				}
			}
			return next.Format("20060102"), nil
		}
		// Это услосвие определения, что указаны дни повторений и месяцев.
		if len(yearMonthDay) == 3 {
			daysRepeat := strings.Split(yearMonthDay[1], ",")
			monthsRepeat := strings.Split(yearMonthDay[2], ",")

			daysInt, err := ConvertInt(daysRepeat, true)
			if err != nil {
				return "", err
			}

			monthsInt, err := ConvertInt(monthsRepeat, false)
			if err != nil {
				return "", err
			}

			year, _, _ := dateParse.Date()
			next := dateParse

		loop3:
			for {
				for _, month := range monthsInt {
					for _, dayNew := range daysInt {
						// 40 - это '-1' (последний день месяца)
						// 35 - это '-2' (предпоследний день месяца)
						if dayNew == 35 || dayNew == 40 {
							dayNew = DayLessZero(year, int(month), dayNew)
						}

						next = time.Date(year, time.Month(month), dayNew, 0, 0, 0, 0, time.Local)
						if next.After(now) {
							break loop3
						}
					}
				}
				year++ // Если за текущий год нет подходящей даты, цикл переходит на следующий год.
			}
			return next.Format("20060102"), nil
		}

	}
	return "", nil
}
