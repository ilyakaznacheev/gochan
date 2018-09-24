package main

import (
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

const (
	templatePath = "static/template/"
	timeFormat   = "Mon _2 Jan 2006 15:04:05"
)

var modelCtxShared *modelContext

// MainRepr is a context for main.html template
type MainRepr struct {
	Key  string
	Name string
}

// BoardRepr is a context for board.html template
type BoardRepr struct {
	Key   string
	Title string
	Time  string
}

// BoardReprInfo is a part of board template context
type BoardReprInfo struct {
	Name string
	Key  string
}

// PostRepr is a context for post.html template
type PostRepr struct {
	Key    string
	Author string
	Time   string
	Text   string
	IsOP   bool
}

// ThreadReprBoard is a part of thread template context
type ThreadReprBoard struct {
	Key string
}

// ThreadReprInfo is a part of thread template context
type ThreadReprInfo struct {
	Key    string
	Title  string
	Author string
}

// ThreadRepr is a context for thread.html template
type ThreadRepr struct {
	Board  ThreadReprBoard
	Thread ThreadReprInfo
	Posts  []PostRepr
}

func getModelCtx() *modelContext {
	if modelCtxShared == nil {
		modelCtxShared = getmodelContext()
	}
	return modelCtxShared
}

// MainPage returns index page
func MainPage(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles(templatePath + "home.html"))
	modelCtx := getModelCtx()

	modelData := modelCtx.boardModel.getList()

	ctxBoards := make([]MainRepr, 0, len(modelData))

	for _, board := range modelData {
		ctxBoards = append(ctxBoards, MainRepr{
			Key:  string(board.key),
			Name: board.name,
		})
	}

	tmpl.Execute(w, struct{ Boards []MainRepr }{ctxBoards})

}

// BoardPage returns board page
func BoardPage(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles(templatePath + "board.html"))
	modelCtx := getModelCtx()
	requestParams := mux.Vars(r)

	boardData, err := modelCtx.boardModel.getItem(boardKey(requestParams["board"]))
	if err != nil {
		log.Println(err)
	}

	modelData, err := modelCtx.threadModel.getTheadsByBoard(boardKey(requestParams["board"]))
	if err != nil {
		log.Println(err)
	}

	ctxThreads := make([]BoardRepr, 0, len(modelData))

	for _, threadItem := range modelData {
		ctxThreads = append(ctxThreads, BoardRepr{
			Key:   strconv.Itoa(int(threadItem.key)),
			Title: threadItem.title,
			Time:  threadItem.creationDateTime.Format(timeFormat),
		})
	}

	tmpl.Execute(w, struct {
		Board   BoardReprInfo
		Threads []BoardRepr
	}{BoardReprInfo{boardData.name, string(boardData.key)}, ctxThreads})
}

// ThreadPage returns thread page
func ThreadPage(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles(templatePath + "thread.html"))
	modelCtx := getModelCtx()
	requestParams := mux.Vars(r)

	threadIDReq, _ := strconv.Atoi(requestParams["id"])

	threadData, err := modelCtx.threadModel.getThread(threadKey(threadIDReq))
	if err != nil {
		log.Println(err)
	}
	boardData, err := modelCtx.boardModel.getItem(threadData.boardName)
	if err != nil {
		log.Println(err)
	}

	postData, err := modelCtx.postModel.getPostsByThread(threadKey(threadIDReq))
	if err != nil {
		log.Println(err)
	}

	var ctxThread ThreadRepr
	ctxThread.Board = ThreadReprBoard{Key: string(boardData.key)}
	ctxThread.Thread = ThreadReprInfo{
		Key:    strconv.Itoa(int(threadData.key)),
		Title:  threadData.title,
		Author: string(threadData.authorID),
	}

	ctxThread.Posts = make([]PostRepr, 0, len(postData))

	for _, threadItem := range postData {

		ctxThread.Posts = append(ctxThread.Posts, PostRepr{
			Key:    string(threadItem.key),
			Author: string(threadItem.author),
			Time:   threadItem.creationDateTime.Format(timeFormat),
			Text:   threadItem.text,
			IsOP:   threadItem.author == threadData.authorID,
		})
	}

	tmpl.Execute(w, ctxThread)
}

// AddMessage adds new message to thread
func AddMessage(w http.ResponseWriter, r *http.Request) {
	requestParams := mux.Vars(r)
	threadID, _ := strconv.Atoi(requestParams["id"])

	authorCookie, err := r.Cookie("author_id")

	var authorID string
	loggedIn := (err != http.ErrNoCookie)

	if !loggedIn {
		expiration := time.Now().Add(1 * time.Minute)
		authorID = uuid.New().String()
		cookie := http.Cookie{
			Name:    "author_id",
			Value:   authorID,
			Expires: expiration,
		}
		http.SetCookie(w, &cookie)
	} else {
		authorID = authorCookie.Value
	}

	inputText := r.FormValue("message")

	log.Println("New message by", authorID, inputText)

	modelCtx := getModelCtx()

	newPost := post{
		author:           authorKey(authorID),
		thread:           threadKey(threadID),
		creationDateTime: time.Now(),
		text:             inputText,
	}
	_, err = modelCtx.postModel.putPost(newPost)
	if err != nil {
		log.Println(err)
	}
	http.Redirect(w, r, "/thread/"+strconv.Itoa(threadID), http.StatusFound)
}

// AddThread adds new thread
func AddThread(w http.ResponseWriter, r *http.Request) {
	requestParams := mux.Vars(r)
	boardName := boardKey(requestParams["board"])

	authorCookie, err := r.Cookie("author_id")

	var authorID string
	loggedIn := (err != http.ErrNoCookie)
	if !loggedIn {
		expiration := time.Now().Add(1 * time.Minute)
		authorID = uuid.New().String()
		cookie := http.Cookie{
			Name:    "author_id",
			Value:   authorID,
			Expires: expiration,
		}
		http.SetCookie(w, &cookie)
	} else {
		authorID = authorCookie.Value
	}

	inputTitle := r.FormValue("title")
	inputText := r.FormValue("message")

	log.Println("New thread by", authorID, inputTitle)

	modelCtx := getModelCtx()

	newThread := thread{
		title:            inputTitle,
		authorID:         authorKey(authorID),
		boardName:        boardName,
		creationDateTime: time.Now(),
	}

	threadID, err := modelCtx.threadModel.putThread(newThread)
	if err != nil {
		log.Println(err)
	}

	log.Println("New message by", authorID, inputText)

	newPost := post{
		author:           authorKey(authorID),
		thread:           threadKey(threadID),
		creationDateTime: time.Now(),
		text:             inputText,
	}
	_, err = modelCtx.postModel.putPost(newPost)
	if err != nil {
		log.Println(err)
	}
	http.Redirect(w, r, "/"+string(boardName), http.StatusFound)
}

// AuthorPage returns all messages by author selected
func AuthorPage(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles(templatePath + "author.html"))
	modelCtx := getModelCtx()
	requestParams := mux.Vars(r)

	authorID := authorKey(requestParams["author"])

	authorData, err := modelCtx.postModel.getPostsByAuthor(authorID)
	if err != nil {
		log.Println(err)
	}

	var ctxThread ThreadRepr

	ctxThread.Posts = make([]PostRepr, 0, len(authorData))

	for _, postItem := range authorData {

		ctxThread.Posts = append(ctxThread.Posts, PostRepr{
			Key:    string(postItem.key),
			Author: string(postItem.author),
			Time:   postItem.creationDateTime.Format(timeFormat),
			Text:   postItem.text,
			IsOP:   true,
		})
	}

	tmpl.Execute(w, ctxThread)
}

// AdminPage loads admin cockpit
func AdminPage(w http.ResponseWriter, r *http.Request) {

}
