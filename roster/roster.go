package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"reflect"

	_ "github.com/mattn/go-sqlite3"
)

type User struct {
	Name   string
	Course string
	Role   float64
}

func main() {
	os.Remove("./rosterdb.sqlite")
	db, err := sql.Open("sqlite3", "./rosterdb.sqlite")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	CreateUserDB(db)

	file, err := os.Open("./roster_data_sample.json")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	jsonBytes, err := ioutil.ReadAll(file)

	var arr interface{}
	err = json.Unmarshal([]byte(jsonBytes), &arr)

	roster := ReadJsonArray(arr)
	for i := 0; i < len(roster); i++ {
		UpdateUserDB(roster[i], db)
	}

	ans, err := GetAssignmentAnswer(db)
	fmt.Println(ans)
}

func CreateUserDB(db *sql.DB) error {
	createStmt := `
	CREATE TABLE User (
		id     INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT UNIQUE,
		name   TEXT UNIQUE
	);

	CREATE TABLE Course (
			id     INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT UNIQUE,
			title  TEXT UNIQUE
	);

	CREATE TABLE Member (
			user_id     INTEGER,
			course_id   INTEGER,
			role        INTEGER,
			PRIMARY KEY (user_id, course_id)
	);`

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec(createStmt)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func ReadJsonArray(json interface{}) []User {
	var roster []User
	switch reflect.TypeOf(json).Kind() {
	case reflect.Slice:
		item := reflect.ValueOf(json)

		for i := 0; i < item.Len(); i++ {
			a := item.Index(i).Interface()
			b := a.([]interface{})

			user := User{
				Name:   b[0].(string),
				Course: b[1].(string),
				Role:   b[2].(float64),
			}

			roster = append(roster, user)
		}
	}

	return roster
}

func UpdateUserDB(u User, db *sql.DB) error {
	var (
		user_id   int
		course_id int
	)
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	newUser, err := tx.Prepare("INSERT OR IGNORE INTO User (name) VALUES (?)")
	if err != nil {
		return err
	}
	defer newUser.Close()
	newUser.Exec(u.Name)

	row := tx.QueryRow("SELECT id FROM User WHERE name = ?", u.Name)
	err = row.Scan(&user_id)
	if err != nil {
		return err
	}

	newCourse, err := tx.Prepare("INSERT OR IGNORE INTO Course (title) VALUES (?)")
	if err != nil {
		return err
	}
	defer newCourse.Close()
	newCourse.Exec(u.Course)

	row = tx.QueryRow("SELECT id FROM Course WHERE title = ?", u.Course)
	err = row.Scan(&course_id)
	if err != nil {
		return err
	}

	newMember, err := tx.Prepare("INSERT OR IGNORE INTO Member (user_id, course_id, role) VALUES (?, ?, ?)")
	if err != nil {
		return err
	}
	defer newMember.Close()
	newMember.Exec(user_id, course_id, u.Role)

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func GetAssignmentAnswer(db *sql.DB) (string, error) {
	var answer string

	tx, err := db.Begin()
	if err != nil {
		return "", err
	}

	answerStmt := `
	SELECT hex(User.name || Course.title || Member.role ) AS X FROM 
		User JOIN Member JOIN Course 
		ON User.id = Member.user_id AND Member.course_id = Course.id
		ORDER BY X;`
	row := tx.QueryRow(answerStmt)
	err = row.Scan(&answer)
	if err != nil {
		return "", err
	}

	err = tx.Commit()
	if err != nil {
		return "", err
	}

	return answer, nil
}
