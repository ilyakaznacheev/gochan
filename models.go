package main

import (
	"database/sql"
	"errors"
	"time"

	_ "github.com/lib/pq"
)

const (
	dbUser    = "gochanuser"
	dbPass    = "gochanpass"
	dbName    = "gochandb"
	dbSSL     = "disable"
	dbAddress = "localhost"
)

type dbHandler struct {
	DB *sql.DB
}

func getDBConnection() (db *dbHandler, err error) {
	connStr := "user=" + dbUser +
		" dbname=" + dbName +
		" password=" + dbPass +
		" host=" + dbAddress +
		" sslmode=" + dbSSL

	dbConn, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	return &dbHandler{dbConn}, nil
}

type (
	boardKey  string
	threadKey int
	postKey   int
	authorKey string
)

type board struct {
	key  boardKey
	name string
}

type boardModel struct {
	// boardSet []board
	dbConn *dbHandler
}

func (m *boardModel) getList() (boardList []*board) {
	rows, err := m.dbConn.DB.Query(`SELECT key, name FROM board`)
	if err != nil {
		panic(err)
	}
	boardList = make([]*board, 0)
	for rows.Next() {
		boardItem := &board{}
		err = rows.Scan(&boardItem.key, &boardItem.name)
		boardList = append(boardList, boardItem)
	}
	rows.Close()
	return boardList
}

func (m *boardModel) getItem(name boardKey) (*board, error) {
	row := m.dbConn.DB.QueryRow(
		`SELECT key, name 
		FROM board
		WHERE key = $1`,
		name)
	boardItem := &board{}
	err := row.Scan(
		&boardItem.key,
		&boardItem.name,
	)
	if err != nil {
		return nil, err
	}
	return boardItem, nil
}

func getBoardModel(dbConn *dbHandler) *boardModel {
	return &boardModel{
		dbConn: dbConn,
	}
}

type thread struct {
	key              threadKey
	title            string
	authorID         authorKey
	boardName        boardKey
	creationDateTime time.Time
}

type threadModel struct {
	// ThreadSet []thread
	dbConn *dbHandler
}

func (m *threadModel) getTheadsByBoard(boardName boardKey) ([]*thread, error) {
	rows, err := m.dbConn.DB.Query(
		`SELECT key, title, authorid, boardname, creationdatetime 
			FROM thread
			WHERE boardname = $1`,
		boardName)
	if err != nil {
		return nil, err
	}

	threadList := make([]*thread, 0)
	for rows.Next() {
		threadItem := &thread{}
		err = rows.Scan(
			&threadItem.key,
			&threadItem.title,
			&threadItem.authorID,
			&threadItem.boardName,
			&threadItem.creationDateTime,
		)
		threadList = append(threadList, threadItem)
	}
	rows.Close()

	if len(threadList) == 0 {
		return nil, errors.New("no threads found with board name " + string(boardName))
	}
	return threadList, nil
}

func (m *threadModel) getThreadsByAuthor(authorID authorKey) ([]*thread, error) {
	rows, err := m.dbConn.DB.Query(
		`SELECT key, title, authorid, boardname, creationdatetime 
			FROM thread
			WHERE authorid = $1`,
		authorID)
	if err != nil {
		return nil, err
	}

	threadList := make([]*thread, 0)
	for rows.Next() {
		threadItem := &thread{}
		err = rows.Scan(
			&threadItem.key,
			&threadItem.title,
			&threadItem.authorID,
			&threadItem.boardName,
			&threadItem.creationDateTime,
		)
		threadList = append(threadList, threadItem)
	}
	rows.Close()

	if len(threadList) == 0 {
		return nil, errors.New("no threads found with author ID" + string(authorID))
	}
	return threadList, nil
}

func (m *threadModel) getThread(threadID threadKey) (*thread, error) {
	row := m.dbConn.DB.QueryRow(
		`SELECT key, title, authorid, boardname, creationdatetime 
			FROM thread
			WHERE key = $1`,
		threadID)
	threadItem := &thread{}
	err := row.Scan(
		&threadItem.key,
		&threadItem.title,
		&threadItem.authorID,
		&threadItem.boardName,
		&threadItem.creationDateTime,
	)
	if err != nil {
		return nil, err
	}
	return threadItem, nil
}

func (m *threadModel) putThread(newThread thread) (threadKey, error) {
	row := m.dbConn.DB.QueryRow(
		`INSERT INTO thread (key, title, authorid, boardname, creationdatetime ) VALUES (
			nextval('thread_key_seq'),
			$1, 
			$2, 
			$3, 
			$4
			) RETURNING key;`,
		newThread.title,
		newThread.authorID,
		newThread.boardName,
		newThread.creationDateTime,
	)

	var index threadKey

	err := row.Scan(&index)
	if err != nil {
		return 0, err
	}
	return index, nil
}

// func (m *threadModel) GetLastThreads(number int) []thread {
// }

func getThreadModel(dbConn *dbHandler) *threadModel {
	return &threadModel{
		dbConn: dbConn,
	}
}

type post struct {
	key              postKey
	author           authorKey
	thread           threadKey
	creationDateTime time.Time
	text             string
}

type postModel struct {
	// PostSet []post
	dbConn *dbHandler
}

func (m *postModel) getPostsByThread(threadID threadKey) ([]*post, error) {
	rows, err := m.dbConn.DB.Query(
		`SELECT key, author, thread, creationdatetime, text
			FROM post
			WHERE thread = $1`,
		threadID)
	if err != nil {
		return nil, err
	}

	postList := make([]*post, 0)
	for rows.Next() {
		postItem := &post{}
		err = rows.Scan(
			&postItem.key,
			&postItem.author,
			&postItem.thread,
			&postItem.creationDateTime,
			&postItem.text,
		)
		postList = append(postList, postItem)
	}
	rows.Close()

	if len(postList) == 0 {
		return nil, errors.New("no posts found with thread ID" + string(threadID))
	}
	return postList, nil
}

func (m *postModel) getPostsByAuthor(authorID authorKey) ([]*post, error) {
	rows, err := m.dbConn.DB.Query(
		`SELECT key, author, thread, creationdatetime, text
			FROM post
			WHERE author = $1`,
		authorID)
	if err != nil {
		return nil, err
	}

	postList := make([]*post, 0)
	for rows.Next() {
		postItem := &post{}
		err = rows.Scan(
			&postItem.key,
			&postItem.author,
			&postItem.thread,
			&postItem.creationDateTime,
			&postItem.text,
		)
		postList = append(postList, postItem)
	}
	rows.Close()

	if len(postList) == 0 {
		return nil, errors.New("no posts found with author ID" + string(authorID))
	}
	return postList, nil
}

func (m *postModel) getPost(postID postKey) (*post, error) {
	row := m.dbConn.DB.QueryRow(
		`SELECT key, author, thread, creationdatetime, text
		FROM post
			WHERE key = $1`,
		postID)
	postItem := &post{}
	err := row.Scan(
		&postItem.key,
		&postItem.author,
		&postItem.thread,
		&postItem.creationDateTime,
		&postItem.text,
	)
	if err != nil {
		return nil, err
	}
	return postItem, nil
}

func (m *postModel) putPost(newPost post) (postKey, error) {
	row := m.dbConn.DB.QueryRow(
		`INSERT INTO post (author, thread, creationdatetime, text) VALUES (
			$1, 
			$2, 
			$3, 
			$4
			) RETURNING key;`,
		newPost.author,
		newPost.thread,
		newPost.creationDateTime,
		newPost.text,
	)

	var index postKey

	err := row.Scan(&index)
	if err != nil {
		return 0, err
	}
	return index, nil
}

func getPostModel(dbConn *dbHandler) *postModel {
	return &postModel{
		dbConn: dbConn,
	}
}

type author struct {
	key authorKey
	// AdminRole bool
}

type authorModel struct {
	// authorSet []author
	dbConn *dbHandler
}

func (m *authorModel) getAuthor(authorID authorKey) (*author, error) {
	row := m.dbConn.DB.QueryRow(
		`SELECT key
		FROM author
			WHERE key = $1`,
		authorID)
	authorItem := &author{}
	err := row.Scan(
		&authorItem.key,
	)
	if err != nil {
		return nil, err
	}
	return authorItem, nil
}

func getAuthorModel(dbConn *dbHandler) *authorModel {
	return &authorModel{
		dbConn: dbConn,
	}
}

type modelContext struct {
	dbConnection *dbHandler
	boardModel   *boardModel
	threadModel  *threadModel
	postModel    *postModel
	authorModel  *authorModel
}

func getmodelContext() *modelContext {
	dbConn, err := getDBConnection()
	if err != nil {
		panic(err)
	}

	modelContext := &modelContext{
		dbConnection: dbConn,
		boardModel:   getBoardModel(dbConn),
		threadModel:  getThreadModel(dbConn),
		postModel:    getPostModel(dbConn),
		authorModel:  getAuthorModel(dbConn),
	}

	return modelContext
}
