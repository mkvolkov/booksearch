package main

import (
	"log"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
)

// благоприятные тесты
// тесты с авторами, у которых нет книг,
// по сути, ничем не отличаются от обычных,
// только возвращается пустой массив
// (аналогично для запроса автора по книге)
func Test_QueryBooks(t *testing.T) {
	tests := []struct {
		name          string
		expectedBooks []Title
	}{
		{
			name: "%Пушкин%",
			expectedBooks: []Title{
				{
					Title: "Евгений Онегин",
				},
			},
		},
		{
			name: "%Гоголь%",
			expectedBooks: []Title{
				{
					Title: "Тарас Бульба",
				},
				{
					Title: "Ночь перед Рождеством",
				},
			},
		},
		{
			name: "%Достоевский%",
			expectedBooks: []Title{
				{
					Title: "Преступление и наказание",
				},
				{
					Title: "Игрок",
				},
				{
					Title: "Идиот",
				},
			},
		},
	}

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		log.Fatal("error init mock", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	Dbase := Db{dbase: sqlxDB}

	for _, tc := range tests {
		var rows = sqlmock.NewRows([]string{`title`})
		for _, book := range tc.expectedBooks {
			rows.AddRow(book.Title)
		}

		mock.
			ExpectQuery(queryBooks).
			WithArgs(tc.name).
			WillReturnRows(rows)

		data, err := Dbase.GetBooks(tc.name)
		if !reflect.DeepEqual(data, tc.expectedBooks) {
			t.Error("Wrong data")
		}

		if err != nil {
			t.Error("Books req error")
		}
	}
}

func Test_QueryAuthors(t *testing.T) {
	tests := []struct {
		title           string
		expectedAuthors []Name
	}{
		{
			title: "%Онегин%",
			expectedAuthors: []Name{
				{
					Name: "Пушкин А.С.",
				},
			},
		},
		{
			title: "%Бульба%",
			expectedAuthors: []Name{
				{
					Name: "Гоголь Н.В.",
				},
			},
		},
		{
			title: "%Преступление и наказание%",
			expectedAuthors: []Name{
				{
					Name: "Достоевский Ф.М.",
				},
			},
		},
		{
			title: "%Золотой телёнок%",
			expectedAuthors: []Name{
				{
					Name: "Ильф И.А.",
				},
				{
					Name: "Петров Е.П.",
				},
			},
		},
	}

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		log.Fatal("error init mock", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	Dbase := Db{dbase: sqlxDB}

	for _, tc := range tests {
		var rows = sqlmock.NewRows([]string{`name`})
		for _, author := range tc.expectedAuthors {
			rows.AddRow(author.Name)
		}

		mock.
			ExpectQuery(queryAuthors).
			WithArgs(tc.title).
			WillReturnRows(rows)

		data, err := Dbase.GetAuthor(tc.title)
		if !reflect.DeepEqual(data, tc.expectedAuthors) {
			t.Error("Wrong data")
		}

		if err != nil {
			t.Error("Books req error")
		}
	}
}
