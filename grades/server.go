package grades

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func RegisterHandlers() {
	handler := new(studentsHandler)
	http.Handle("/students", handler)
	http.Handle("/students/", handler)
}

type studentsHandler struct{}

// /students - entire class
// /students/{id} - a single student's record
// /students/{id}/grades - a single student's grades
func (sh studentsHandler) ServeHTTP(wr http.ResponseWriter, r *http.Request) {
	pathSegments := strings.Split(r.URL.Path, "/")
	switch len(pathSegments) {
	case 2:
		sh.getAll(wr, r)
	case 3:
		id, err := strconv.Atoi(pathSegments[2])
		if err != nil {
			wr.WriteHeader(http.StatusNotFound)
			return
		}
		sh.getOne(wr, r, id)
	case 4:
		id, err := strconv.Atoi(pathSegments[3])
		if err != nil {
			wr.WriteHeader(http.StatusNotFound)
			return
		}
		sh.addGrade(wr, r, id)
	default:
		wr.WriteHeader(http.StatusNotFound)
	}
}

func (sh studentsHandler) getAll(wr http.ResponseWriter, r *http.Request) {
	studentsMutex.Lock()
	defer studentsMutex.Unlock()

	data, err := sh.toJson(students)
	if err != nil {
		wr.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}
	wr.Header().Add("content-Type", "application/json")
	wr.Write(data)
}

func (sh studentsHandler) toJson(obj interface{}) ([]byte, error) {
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	err := enc.Encode(obj)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize students: %v", err)
	}
	return b.Bytes(), nil
}

func (sh studentsHandler) getOne(wr http.ResponseWriter, r *http.Request, id int) {
	studentsMutex.Lock()
	defer studentsMutex.Unlock()

	student, err := students.GetByID(id)
	if err != nil {
		wr.WriteHeader(http.StatusNotFound)
		log.Println(err)
		return
	}

	data, err := sh.toJson(student)
	if err != nil {
		wr.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}
	wr.Header().Add("content-Type", "application/json")
	wr.Write(data)
}

func (sh studentsHandler) addGrade(wr http.ResponseWriter, r *http.Request, id int) {
	studentsMutex.Lock()
	defer studentsMutex.Unlock()

	student, err := students.GetByID(id)
	if err != nil {
		wr.WriteHeader(http.StatusNotFound)
		log.Println(err)
		return
	}

	var g Grade
	dec := json.NewDecoder(r.Body)
	err = dec.Decode(&g)
	if err != nil {
		wr.WriteHeader(http.StatusBadRequest)
		log.Println(err)
		return
	}

	student.Grades = append(student.Grades, g)
	wr.WriteHeader(http.StatusCreated)

	data, err := sh.toJson(student)
	if err != nil {
		log.Println(err)
		return
	}
	wr.Header().Add("content-Type", "application/json")
	wr.Write(data)
}
