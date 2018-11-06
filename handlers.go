package gochan

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"html/template"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

const (
	templatePath = "static/template/"
	imgPath      = "media/img/"
	timeFormat   = "Mon _2 Jan 2006 15:04:05"
)

// MainRepr is a context for main.html template
type MainRepr struct {
	Key  string
	Name string
}

// BoardRepr is a context for board.html template
type BoardRepr struct {
	Key       string
	Title     string
	Time      string
	ImagePath string
	HasImage  bool
}

// BoardReprInfo is a part of board template context
type BoardReprInfo struct {
	Name string
	Key  string
}

// PostRepr is a context for post.html template
type PostRepr struct {
	Key       string
	Author    string
	Time      string
	Text      string
	IsOP      bool
	ImagePath string
	HasImage  bool
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

// RequestHandler is a common request handler interface
type RequestHandler interface {
	MainPage(http.ResponseWriter, *http.Request)
	BoardPage(http.ResponseWriter, *http.Request)
	ThreadPage(http.ResponseWriter, *http.Request)
	AddMessage(http.ResponseWriter, *http.Request)
	AddThread(http.ResponseWriter, *http.Request)
	AuthorPage(http.ResponseWriter, *http.Request)
	AdminPage(http.ResponseWriter, *http.Request)
}

// ChanRequestHandler handles http requests
type ChanRequestHandler struct {
	model *modelContext
}

func (rh *ChanRequestHandler) uploadImage(r *http.Request) (*uuid.UUID, error) {

	file, handler, err := r.FormFile("picture")
	if err != nil {
		return nil, errors.New("picture doesn's load: " + err.Error())
	}

	defer file.Close()

	if handler.Size == 0 {
		return nil, errors.New("file is empty")
	}

	tmpName := RandStringRunes(32)

	// fileEndings, err := mime.ExtensionsByType(http.DetectContentType(fileBytes))
	if err != nil {
		return nil, err
	}

	fileExt := path.Ext(handler.Filename)
	tmpFile := filepath.Join(imgPath, tmpName+fileExt)
	newFile, err := os.Create(tmpFile)
	if err != nil {
		return nil, errors.New("cant open file: " + err.Error())
	}

	hasher := md5.New()
	_, err = io.Copy(newFile, io.TeeReader(file, hasher))
	if err != nil {
		return nil, errors.New("cant save file: " + err.Error())
	}
	newFile.Sync()
	newFile.Close()

	// md5Sum := hex.EncodeToString(hasher.Sum(nil))
	md5SumHEX := make([]byte, hex.EncodedLen(len(hasher.Sum(nil))))
	hex.Encode(md5SumHEX, hasher.Sum(nil))
	fileUUID, err := uuid.ParseBytes(md5SumHEX)
	if err != nil {
		return nil, errors.New("cant generate uuid: " + err.Error())
	}
	md5Sum := fileUUID.String()

	if rh.model.imageModel.isImageExist(ImageKey(fileUUID)) {
		os.Remove(tmpFile)
		return &fileUUID, nil
	}

	realFile := filepath.Join(imgPath, md5Sum+fileExt)
	err = os.Rename(tmpFile, realFile)
	if err != nil {
		return nil, errors.New("cant raname file: " + err.Error())
	}

	log.Println("new file upload:", realFile)

	err = rh.model.imageModel.putImage(&Image{
		Key:      ImageKey(fileUUID),
		FilePath: realFile,
	})
	if err != nil {
		return nil, err
	}

	return &fileUUID, nil
}

// MainPage returns index page
func (rh *ChanRequestHandler) MainPage(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles(templatePath + "home.html"))

	modelData := rh.model.boardModel.getList()

	ctxBoards := make([]MainRepr, 0, len(modelData))

	for _, board := range modelData {
		ctxBoards = append(ctxBoards, MainRepr{
			Key:  string(board.Key),
			Name: board.Name,
		})
	}

	tmpl.Execute(w, struct{ Boards []MainRepr }{ctxBoards})

}

// BoardPage returns board page
func (rh *ChanRequestHandler) BoardPage(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles(templatePath + "board.html"))
	requestParams := mux.Vars(r)

	boardData, err := rh.model.boardModel.getItem(BoardKey(requestParams["board"]))
	if err != nil {
		log.Println(err)
	}

	modelData, err := rh.model.threadModel.getTheadsByBoard(BoardKey(requestParams["board"]))
	if err != nil {
		log.Println(err)
	}

	ctxThreads := make([]BoardRepr, 0, len(modelData))

	for _, threadItem := range modelData {
		ctxThreads = append(ctxThreads, BoardRepr{
			Key:       strconv.Itoa(int(threadItem.Key)),
			Title:     threadItem.Title,
			Time:      threadItem.CreationDateTime.Format(timeFormat),
			ImagePath: threadItem.getImagePath(),
			HasImage:  threadItem.ImagePath != nil,
		})
	}

	tmpl.Execute(w, struct {
		Board   BoardReprInfo
		Threads []BoardRepr
	}{BoardReprInfo{boardData.Name, string(boardData.Key)}, ctxThreads})
}

// ThreadPage returns thread page
func (rh *ChanRequestHandler) ThreadPage(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles(templatePath + "thread.html"))
	requestParams := mux.Vars(r)

	threadIDReq, _ := strconv.Atoi(requestParams["id"])

	threadData, err := rh.model.threadModel.getThread(ThreadKey(threadIDReq))
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	boardData, err := rh.model.boardModel.getItem(threadData.BoardName)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	postData, err := rh.model.postModel.getPostsByThread(ThreadKey(threadIDReq))
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var ctxThread ThreadRepr
	ctxThread.Board = ThreadReprBoard{Key: string(boardData.Key)}
	ctxThread.Thread = ThreadReprInfo{
		Key:    strconv.Itoa(int(threadData.Key)),
		Title:  threadData.Title,
		Author: string(threadData.AuthorID),
	}

	ctxThread.Posts = make([]PostRepr, 0, len(postData))

	for _, threadItem := range postData {

		ctxThread.Posts = append(ctxThread.Posts, PostRepr{
			Key:       strconv.Itoa(int(threadItem.Key)),
			Author:    string(threadItem.Author),
			Time:      threadItem.CreationDateTime.Format(timeFormat),
			Text:      threadItem.Text,
			IsOP:      threadItem.Author == threadData.AuthorID,
			ImagePath: threadItem.getImagePath(),
			HasImage:  threadItem.ImagePath != nil,
		})
	}

	tmpl.Execute(w, ctxThread)
}

// AddMessage adds new message to thread
func (rh *ChanRequestHandler) AddMessage(w http.ResponseWriter, r *http.Request) {
	requestParams := mux.Vars(r)
	ThreadID, _ := strconv.Atoi(requestParams["id"])

	// read file
	fileUUID, err := rh.uploadImage(r)
	if err != nil {
		log.Println("error while file upload", err)
	}

	// check cookie
	authorCookie, err := r.Cookie("author_id")

	var AuthorID string
	loggedIn := (err != http.ErrNoCookie)

	if !loggedIn {
		expiration := time.Now().Add(1 * time.Minute)
		AuthorID = uuid.New().String()
		cookie := http.Cookie{
			Name:    "author_id",
			Value:   AuthorID,
			Expires: expiration,
		}
		http.SetCookie(w, &cookie)
	} else {
		AuthorID = authorCookie.Value
	}

	inputText := r.FormValue("message")

	log.Println("New message by", AuthorID, inputText)

	// save post data
	newPost := Post{
		Author:           AuthorKey(AuthorID),
		Thread:           ThreadKey(ThreadID),
		CreationDateTime: time.Now(),
		Text:             inputText,
		ImageKey:         fileUUID,
	}
	_, err = rh.model.postModel.putPost(newPost)
	if err != nil {
		log.Println(err)
	}
	http.Redirect(w, r, "/thread/"+strconv.Itoa(ThreadID), http.StatusFound)
}

// AddThread adds new thread
func (rh *ChanRequestHandler) AddThread(w http.ResponseWriter, r *http.Request) {
	requestParams := mux.Vars(r)
	BoardName := BoardKey(requestParams["board"])

	// read file
	fileUUID, err := rh.uploadImage(r)
	if err != nil {
		log.Println("error whila file upload", err)
	}

	authorCookie, err := r.Cookie("author_id")

	var AuthorID string
	loggedIn := (err != http.ErrNoCookie)
	if !loggedIn {
		expiration := time.Now().Add(1 * time.Minute)
		AuthorID = uuid.New().String()
		cookie := http.Cookie{
			Name:    "author_id",
			Value:   AuthorID,
			Expires: expiration,
		}
		http.SetCookie(w, &cookie)
	} else {
		AuthorID = authorCookie.Value
	}

	inputTitle := r.FormValue("title")
	inputText := r.FormValue("message")

	log.Println("New thread by", AuthorID, inputTitle)

	newThread := Thread{
		Title:            inputTitle,
		AuthorID:         AuthorKey(AuthorID),
		BoardName:        BoardName,
		CreationDateTime: time.Now(),
		ImageKey:         fileUUID,
	}

	ThreadID, err := rh.model.threadModel.putThread(newThread)
	if err != nil {
		log.Println(err)
	}

	log.Println("New message by", AuthorID, inputText)

	newPost := Post{
		Author:           AuthorKey(AuthorID),
		Thread:           ThreadKey(ThreadID),
		CreationDateTime: time.Now(),
		Text:             inputText,
		ImageKey:         fileUUID,
	}
	_, err = rh.model.postModel.putPost(newPost)
	if err != nil {
		log.Println(err)
	}
	http.Redirect(w, r, "/"+string(BoardName), http.StatusFound)
}

// AuthorPage returns all messages by Author selected
func (rh *ChanRequestHandler) AuthorPage(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles(templatePath + "author.html"))
	requestParams := mux.Vars(r)

	AuthorID := AuthorKey(requestParams["author"])

	authorData, err := rh.model.postModel.getPostsByAuthor(AuthorID)
	if err != nil {
		log.Println(err)
	}

	var ctxThread ThreadRepr

	ctxThread.Posts = make([]PostRepr, 0, len(authorData))

	for _, postItem := range authorData {

		ctxThread.Posts = append(ctxThread.Posts, PostRepr{
			Key:       strconv.Itoa(int(postItem.Key)),
			Author:    string(postItem.Author),
			Time:      postItem.CreationDateTime.Format(timeFormat),
			Text:      postItem.Text,
			IsOP:      true,
			ImagePath: postItem.getImagePath(),
			HasImage:  postItem.ImagePath != nil,
		})
	}

	tmpl.Execute(w, ctxThread)
}

// AdminPage loads admin cockpit
func (rh *ChanRequestHandler) AdminPage(w http.ResponseWriter, r *http.Request) {

}

func newRequestHandler(model *modelContext) RequestHandler {
	return &ChanRequestHandler{model}
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

// RandStringRunes returns random string of given length
func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
