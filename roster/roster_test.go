package main

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestCreateUserDB(t *testing.T) {
	tests := []struct {
		testCase string
	}{
		{
			testCase: "Unit test",
		},
	}

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

	for _, test := range tests {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		if err != nil {
			t.Fatalf("an error '%s' was not expected during %s", err, test.testCase)
		}
		defer db.Close()

		mock.ExpectBegin()
		mock.ExpectExec(createStmt).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err = CreateUserDB(db)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	}
}

func TestUpdateUserDB(t *testing.T) {
	tests := []struct {
		u            User
		expUserRow   *sqlmock.Rows
		expCourseRow *sqlmock.Rows
	}{
		{
			u: User{
				Name:   "Bobby Tables",
				Course: "eecs 281",
				Role:   0,
			},
			expUserRow:   sqlmock.NewRows([]string{"id"}).AddRow(1),
			expCourseRow: sqlmock.NewRows([]string{"id"}).AddRow(1),
		},
	}

	for _, test := range tests {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected during mock database initialization", err)
		}
		defer db.Close()

		mock.ExpectBegin()
		mock.ExpectPrepare("INSERT OR IGNORE INTO").ExpectExec().WithArgs(test.u.Name).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectQuery("SELECT").WillReturnRows(test.expUserRow)
		mock.ExpectPrepare("INSERT OR IGNORE INTO").ExpectExec().WithArgs(test.u.Course).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectQuery("SELECT").WillReturnRows(test.expCourseRow)
		mock.ExpectPrepare("INSERT OR IGNORE INTO").ExpectExec().WithArgs(1, 1, test.u.Role).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err = UpdateUserDB(test.u, db)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	}
}
