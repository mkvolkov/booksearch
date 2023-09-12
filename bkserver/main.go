package main

import (
	grbook "booksearch/bkfind"
	cfg "booksearch/cfg"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/go-sql-driver/mysql"

	"github.com/jmoiron/sqlx"
	"google.golang.org/grpc"
)

var mainConn *sqlx.DB

type Db struct {
	dbase *sqlx.DB
}

// запрос: поиск книг по автору
// имя автора не обязано полностью совпадать с именем в таблице
// например, вместо "Достоевский Ф.М." работает запрос
// по имени "Достоевский"
var queryBooks string = `SELECT title
	FROM authors
	JOIN indexes ON authors.author_id = indexes.author
	JOIN books ON books.book_id = indexes.book
	WHERE name LIKE ?`

// запрос: поиск автора по книге
// название книги не обязано полностью совпадать с названием в таблице
var queryAuthors string = `SELECT name
	FROM authors
	JOIN indexes ON authors.author_id = indexes.author
	JOIN books ON books.book_id = indexes.book
	WHERE title LIKE ?`

// вспомогательный метод, исполняющий запрос "поиск книг по автору"
func (db *Db) GetBooks(name string) ([]Title, error) {
	var data []Title
	err := db.dbase.Select(&data, queryBooks, name)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// вспомогательный метод, исполняющий запрос "поиск автора по книге"
func (db *Db) GetAuthor(title string) ([]Name, error) {
	var data []Name
	err := db.dbase.Select(&data, queryAuthors, title)
	if err != nil {
		return nil, err
	}

	return data, nil
}

type Title struct {
	Title string `db:"title"`
}

type Name struct {
	Name string `db:"name"`
}

type server struct {
	grbook.UnimplementedFinderServer
}

// служба gRPC: поиск книг по автору
// функция получает массив книг, формирует из них одну строку, где книги
// разделены переносом строки
// эта строка возвращается клиенту
func (s *server) FindBooks(ctx context.Context, msg *grbook.BReq) (*grbook.BReply, error) {
	var arg string = msg.GetAuthor()

	log.Printf("Req type: Get Book, author: %s\n", arg)

	var argMode string = fmt.Sprintf("%%%s%%", arg)

	Dbase := Db{dbase: mainConn}
	data, err := Dbase.GetBooks(argMode)
	if err != nil {
		log.Fatalln(err)
	}

	var resp string
	for i := 0; i < len(data); i++ {
		resp += data[i].Title + string('\n')
	}

	return &grbook.BReply{Books: resp}, nil
}

// служба gRPC: поиск автора по книге
// функция получает автора в виде списка из одного элемента
// возвращает строку
func (s *server) FindAuthors(ctx context.Context, msg *grbook.AReq) (*grbook.AReply, error) {
	var arg string = msg.GetBook()

	log.Printf("Req type: Get Author, book: %s\n", arg)

	var argMode string = fmt.Sprintf("%%%s%%", arg)

	Dbase := Db{dbase: mainConn}
	data, err := Dbase.GetAuthor(argMode)
	if err != nil {
		log.Fatalln(err)
	}

	/*
		var resp string
		if len(data) == 0 {
			resp = ""
		} else {
			resp = data[0].Name + string('\n')
		}
	*/

	var resp string
	for i := 0; i < len(data); i++ {
		resp += data[i].Name + string('\n')
	}

	return &grbook.AReply{Authors: resp}, nil
}

func InitDB(cfg *cfg.Cfg) (*sqlx.DB, error) {
	// адрес подключения к базе данных
	connectionUrl := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s",
		cfg.Mysql.User,
		cfg.Mysql.Password,
		cfg.Mysql.Host,
		cfg.Mysql.Port,
		cfg.Mysql.Dbname,
	)

	// подключение к базе данных
	conn, err := sqlx.Connect(cfg.Mysql.Driver, connectionUrl)
	if err != nil {
		return nil, err
	}

	// проверка базы данных
	err = conn.Ping()
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func main() {
	// чтение конфигурации базы данных
	dbCfg := &cfg.Cfg{}
	err := cfg.LoadConfig(dbCfg)
	if err != nil {
		log.Fatalln("Error in LoadConfig: ", err)
	}

	// инициализация подключения к базе данных
	mainConn, err = InitDB(dbCfg)
	if err != nil {
		log.Fatalln(err)
	}

	// создание сервера, на основе которого будет построен сервер gRPC
	strAddr := fmt.Sprintf(":%s", dbCfg.Server.Port)
	listener, err := net.Listen("tcp", strAddr)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// создание сервера gRPC
	grpcServer := grpc.NewServer()
	grbook.RegisterFinderServer(grpcServer, &server{})
	log.Printf("Listening at %v...", listener.Addr())

	// канал, принимающий сигнал прерывания для изящного завершения программы
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)

	// изящное завершение программы
	finishCh := make(chan struct{})
	go func() {
		s := <-signalCh
		log.Printf("got signal %v, graceful shutdown...", s)
		grpcServer.GracefulStop()
		mainConn.Close()
		finishCh <- struct{}{}
	}()

	// запуск сервера gRPC
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}

	// ожидание изящного завершения программы
	<-finishCh
	fmt.Println("Finished shutdown")
}
