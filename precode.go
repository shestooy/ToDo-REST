package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Task struct {
	ID           string   `json:"id"`
	Description  string   `json:"description"`
	Note         string   `json:"note"`
	Applications []string `json:"applications"`
}

var tasks = map[string]Task{
	"1": {
		ID:          "1",
		Description: "Сделать финальное задание темы REST API",
		Note:        "Если сегодня сделаю, то завтра будет свободный день. Ура!",
		Applications: []string{
			"VS Code",
			"Terminal",
			"git",
		},
	},
	"2": {
		ID:          "2",
		Description: "Протестировать финальное задание с помощью Postmen",
		Note:        "Лучше это делать в процессе разработки, каждый раз, когда запускаешь сервер и проверяешь хендлер",
		Applications: []string{
			"VS Code",
			"Terminal",
			"git",
			"Postman",
		},
	},
}

// В общем и целом, важный комментарий об устройстве обработчиков
// Выяснилось, что под капотом обработчик всегда возвращает код 200
// И если попытаться в конце обработчика назначить еще раз код ответа 200
// В логгах отмечается как избыточное действие.
// В связи с этим, было принято решение использовать функцию WriteHeader
// Лишь в одном обработчике, который возвращает код 201.

// GetTasks -  Обработчик, возвращает все задачи в виде json файла
// через GET запрос к серверу, в случае ошибки возвращает код 500, в случае успеха - 200
func GetTasks(res http.ResponseWriter, req *http.Request) {
	jsonSlice, err := json.Marshal(tasks)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	res.Header().Set("Content-Type", "application/json")
	_, err = res.Write(jsonSlice)
	if err != nil {
		log.Println(err)
		return
	}

}

// PostTask - Обработчик, обрабатывает POST запрос к северу, добавляет новую задачу в мапу
// в случае успеха возвращает статус 201, в случае ошибки - 400
func PostTask(res http.ResponseWriter, req *http.Request) {
	var newTask Task
	var buf bytes.Buffer

	_, err := buf.ReadFrom(req.Body)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}
	err = json.Unmarshal(buf.Bytes(), &newTask)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
	}
	if _, ok := tasks[newTask.ID]; ok {
		http.Error(res, "запись с таким id уже есть", http.StatusBadRequest)
		return
	}
	tasks[newTask.ID] = newTask
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusCreated)

}

// GetTaskId - Обработчик запросов типа GET c паттерном /{id}
// Возвращает json объект с соотвесвующим id и кодом 200, в случае ошибки - код 400
// Паттерн запроса /{id}, вместо id - число
func GetTaskId(res http.ResponseWriter, req *http.Request) {
	id := chi.URLParam(req, "id")
	if _, ok := tasks[id]; !ok {
		http.Error(res, "задача с данным ID не найдена", http.StatusBadRequest)
		return
	}

	jsonSlice, err := json.Marshal(tasks[id])
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
	}

	res.Header().Set("Content-Type", "application/json")
	_, err = res.Write(jsonSlice)
	if err != nil {
		log.Println(err.Error())
		return
	}
}

// DeleteTaskID - Обработчик запросов типа DELETE c паттерном /{id}
// Удаляет из мапы запись с соответствующим id и
// Отвечает кодом 200 в случае успеха, в случае ошибки - код 400.
// Паттерн запроса /{id}, вместо id - число
func DeleteTaskID(res http.ResponseWriter, req *http.Request) {
	id := chi.URLParam(req, "id")
	if _, ok := tasks[id]; !ok {
		http.Error(res, "задача с данным ID не найдена", http.StatusBadRequest)
		return
	}
	delete(tasks, id)
}

func main() {
	r := chi.NewRouter()

	r.Get("/tasks", GetTasks)
	r.Post("/tasks", PostTask)

	r.Get("/tasks/{id}", GetTaskId)
	r.Delete("/tasks/{id}", DeleteTaskID)

	if err := http.ListenAndServe(":8080", r); err != nil {
		fmt.Printf("Ошибка при запуске сервера: %s", err.Error())
		return
	}
}
