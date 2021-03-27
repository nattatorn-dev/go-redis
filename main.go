package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/xid"
)

type Service struct {
	redis *redisClient
}

func (s *Service) Handler(fn func(*Service, http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		fn(s, w, r)
	}
}

type CreateTaskRequest struct {
	Body string `json:"body"`
}

type CreateTaskResponse struct {
	TaskId string `json:"taskId"`
}

type Error struct {
	Error        int    `json:"error"`
	ErrorMessage string `json:"message,omitempty"`
}

func (service *Service) Request(w http.ResponseWriter, r *http.Request) {
	var task CreateTaskRequest

	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	guid := xid.New()
	taskId := guid.String()

	err = service.redis.setKey(taskId, task, 5*time.Minute)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-type", "application/json; charset=UTF-8;")

		json.NewEncoder(w).Encode(Error{
			http.StatusInternalServerError,
			"",
		})
		return
	}

	w.WriteHeader(http.StatusAccepted)
	w.Header().Set("Content-type", "application/json; charset=UTF-8;")

	json.NewEncoder(w).Encode(CreateTaskResponse{
		taskId,
	})
}

type GetTaskRequest struct {
	TaskId string `json:"taskId"`
}

type GetTaskResponse struct {
	TaskId string      `json:"taskId"`
	Result interface{} `json:"result"`
}

func (service *Service) GetRequest(w http.ResponseWriter, r *http.Request) {
	var taskReq GetTaskRequest

	err := json.NewDecoder(r.Body).Decode(&taskReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var taskResponse GetTaskResponse
	err = service.redis.getKey(taskReq.TaskId, &taskResponse)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Header().Set("Content-type", "application/json; charset=UTF-8;")

		json.NewEncoder(w).Encode(Error{
			http.StatusNotFound,
			"",
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-type", "application/json; charset=UTF-8;")

	json.NewEncoder(w).Encode(taskResponse)
}

func main() {
	redis := initialize()

	service := &Service{redis: redis}
	router := mux.NewRouter()

	router.HandleFunc("/", service.Request).Methods("POST")
	router.HandleFunc("/", service.GetRequest).Methods("GET")

	log.Println("Listin port :80")
	log.Fatal(http.ListenAndServe(":80", router))
}
