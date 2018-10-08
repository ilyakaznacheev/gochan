package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/go-redis/redis"
	_ "github.com/lib/pq"
)

const (
	redisKey           = "goboard"
	redChangeKey       = "last-change"
	redBoardKey        = "board-key"
	redBoardList       = "board-list"
	redThreadKey       = "thread-key"
	redThreadBoardKey  = "thread-board"
	redThreadAuthorKey = "thread-author"
	redPostKey         = "post-key"
	redPostAuthorKey   = "post-thread"
	redPostThreadKey   = "post-thread"
	redAuthorKey       = "author-key"
)

var (
	// ErrRedisCacheVersion error while redis cache version check
	ErrRedisCacheVersion = errors.New("cache outdated")
)

// RedisContainer is a redis main container structure
type RedisContainer struct {
	Version int
	Content string
}

type redisClient struct {
	client *redis.Client
	// input  chan redisAction
	// finish context.CancelFunc
}

func (rc *redisClient) get(entity, key string) (string, error) {
	version, _ := rc.getChangeCounter(entity)
	entityKey := fmt.Sprintf("%s:%s:%s", redisKey, entity, key)
	responseData, err := rc.client.Get(entityKey).Result()
	if err != nil {
		return "", err
	}

	container := &RedisContainer{}
	json.Unmarshal([]byte(responseData), container)
	if version > container.Version {
		return "", ErrRedisCacheVersion
	}
	return container.Content, nil
}

func (rc *redisClient) set(entity, Key, requestData string, version int) error {
	entityKey := fmt.Sprintf("%s:%s:%s", redisKey, entity, Key)

	container := &RedisContainer{
		Version: version,
		Content: requestData,
	}
	requestJSON, err := json.Marshal(container)
	if err != nil {
		log.Panic(err)
	}
	err = rc.client.Set(entityKey, string(requestJSON), 0).Err()
	if err != nil {
		return err
	}
	return nil
}

func (rc *redisClient) updateChangeCounter(entity string) int {
	entityKey := fmt.Sprintf("%s:%s:%s", redisKey, entity, redChangeKey)
	counter, err := rc.client.Incr(entityKey).Result()
	if err != nil {
		log.Println(err)
	}

	return int(counter)
}

func (rc *redisClient) getChangeCounter(entity string) (int, error) {
	entityKey := fmt.Sprintf("%s:%s:%s", redisKey, entity, redChangeKey)
	counterStr, err := rc.client.Get(entityKey).Result()
	if err != nil {
		return 0, err
	}
	counter, err := strconv.Atoi(counterStr)
	if err != nil {
		log.Panic(err)
	}
	return counter, nil
}

func newRedisClient(config *ConfigData) *redisClient {
	client := redis.NewClient(&redis.Options{
		Addr:     config.Redis.Address,
		Password: config.Redis.Password,
		DB:       config.Redis.DataBase,
	})
	return &redisClient{client}
}

type pgClient struct {
	DB *sql.DB
}

func newPGClient(config *ConfigData) (db *pgClient, err error) {
	// config = getConfig()

	connStr := fmt.Sprintf(
		"user=%s dbname=%s password=%s host=%s sslmode=%s",
		config.Database.User,
		config.Database.Name,
		config.Database.Pass,
		config.Database.Address,
		config.Database.SSL,
	)

	dbConn, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	return &pgClient{dbConn}, nil
}

type repoHandler struct {
	redis *redisClient
	pg    *pgClient
}

func newRepoHandler(config *ConfigData) *repoHandler {
	redis := newRedisClient(config)
	pg, err := newPGClient(config)
	if err != nil {
		log.Fatal(err)
	}
	return &repoHandler{
		redis: redis,
		pg:    pg,
	}
}

type (
	boardKey  string
	threadKey int
	postKey   int
	authorKey string
	imageKey  uuid.UUID
)

func (key threadKey) String() string {
	return fmt.Sprintf("%d", key)
}

func (key postKey) String() string {
	return fmt.Sprintf("%d", key)
}

// Board is a db structure of board table
type Board struct {
	Key  boardKey
	Name string
}

type boardModel struct {
	repoConnection *repoHandler
}

func (m *boardModel) getList() (boardList []*Board) {
	var boardListCache []Board
	// read from cache
	cachedData, err := m.repoConnection.redis.get(redBoardList, "")
	if err == nil {
		boardListCache = make([]Board, 0)
		json.Unmarshal([]byte(cachedData), &boardListCache)

		for idx := range boardListCache {
			boardList = append(boardList, &boardListCache[idx])
		}

		return boardList
	}

	// read from db
	rows, err := m.repoConnection.pg.DB.Query(`SELECT key, name FROM board`)
	if err != nil {
		panic(err)
	}
	boardList = make([]*Board, 0)
	for rows.Next() {
		boardItem := &Board{}
		err = rows.Scan(&boardItem.Key, &boardItem.Name)
		boardList = append(boardList, boardItem)
	}
	rows.Close()

	// update cache
	cacheVersion := m.repoConnection.redis.updateChangeCounter(redBoardList)

	boardListCache = make([]Board, 0, len(boardList))
	for idx := range boardList {
		boardListCache = append(boardListCache, *boardList[idx])
	}
	newCachedData, err := json.Marshal(&boardListCache)
	if err != nil {
		log.Panic(err)
	}
	err = m.repoConnection.redis.set(
		redBoardList,
		"",
		string(newCachedData),
		cacheVersion,
	)
	if err != nil {
		log.Panic(err)
	}

	return boardList
}

func (m *boardModel) getItem(name boardKey) (*Board, error) {
	// read from cache
	cachedData, err := m.repoConnection.redis.get(redBoardKey, string(name))
	if err == nil {
		boardCache := &Board{}
		json.Unmarshal([]byte(cachedData), boardCache)

		return boardCache, nil
	}

	// read from db
	row := m.repoConnection.pg.DB.QueryRow(
		`SELECT key, name 
			FROM board
			WHERE key = $1`,
		name,
	)
	boardItem := &Board{}
	err = row.Scan(
		&boardItem.Key,
		&boardItem.Name,
	)
	if err != nil {
		return nil, err
	}

	// update cache
	cacheVersion := m.repoConnection.redis.updateChangeCounter(redBoardKey)

	newCachedData, err := json.Marshal(boardItem)
	if err != nil {
		log.Panic(err)
	}
	err = m.repoConnection.redis.set(
		redBoardKey,
		string(name),
		string(newCachedData),
		cacheVersion,
	)
	if err != nil {
		log.Panic(err)
	}

	return boardItem, nil
}

func newBoardModel(repoConnection *repoHandler) *boardModel {
	return &boardModel{
		repoConnection: repoConnection,
	}
}

// Thread is a db structure of thread table
type Thread struct {
	Key              threadKey
	Title            string
	AuthorID         authorKey
	BoardName        boardKey
	CreationDateTime time.Time
	ImageKey         *uuid.UUID //sql.NullString
	ImagePath        *string
}

func (t *Thread) getImagePath() string {
	if t.ImagePath != nil {
		return *t.ImagePath
	} else {
		return ""
	}
}

type threadModel struct {
	repoConnection *repoHandler
}

func (m *threadModel) getTheadsByBoard(boardName boardKey) ([]*Thread, error) {
	var (
		threadListCache []Thread
		threadList      []*Thread
	)

	// read from cache
	cachedData, err := m.repoConnection.redis.get(redThreadBoardKey, string(boardName))
	if err == nil {
		threadListCache = make([]Thread, 0)
		json.Unmarshal([]byte(cachedData), &threadListCache)

		for idx := range threadListCache {
			threadList = append(threadList, &threadListCache[idx])
		}
		return threadList, nil
	}

	// read from db
	rows, err := m.repoConnection.pg.DB.Query(
		`SELECT thread.key, thread.title, thread.authorid, thread.boardname, thread.creationdatetime, image.filepath
			FROM thread
				LEFT OUTER JOIN image ON
				(thread.image = image.key)
			WHERE thread.boardname = $1`,
		boardName,
	)
	if err != nil {
		return nil, err
	}

	threadList = make([]*Thread, 0)
	for rows.Next() {
		threadItem := &Thread{}
		err = rows.Scan(
			&threadItem.Key,
			&threadItem.Title,
			&threadItem.AuthorID,
			&threadItem.BoardName,
			&threadItem.CreationDateTime,
			&threadItem.ImagePath,
		)
		threadList = append(threadList, threadItem)
	}
	rows.Close()

	if len(threadList) == 0 {
		return nil, errors.New("no threads found with board name " + string(boardName))
	}

	// update cache
	cacheVersion := m.repoConnection.redis.updateChangeCounter(redThreadBoardKey)

	threadListCache = make([]Thread, 0, len(threadList))
	for idx := range threadList {
		threadListCache = append(threadListCache, *threadList[idx])
	}
	newCachedData, err := json.Marshal(&threadListCache)
	if err != nil {
		log.Panic(err)
	}
	err = m.repoConnection.redis.set(
		redThreadBoardKey,
		string(boardName),
		string(newCachedData),
		cacheVersion,
	)
	if err != nil {
		log.Panic(err)
	}

	return threadList, nil
}

func (m *threadModel) getThreadsByAuthor(authorID authorKey) ([]*Thread, error) {
	var (
		threadListCache []Thread
		threadList      []*Thread
	)

	// read from cache
	cachedData, err := m.repoConnection.redis.get(redThreadAuthorKey, string(authorID))
	if err == nil {
		threadListCache = make([]Thread, 0)
		json.Unmarshal([]byte(cachedData), &threadListCache)

		for idx := range threadListCache {
			threadList = append(threadList, &threadListCache[idx])
		}
		return threadList, nil
	}

	// read from db
	rows, err := m.repoConnection.pg.DB.Query(
		`SELECT thread.key, thread.title, thread.authorid, thread.boardname, thread.creationdatetime, image.filepath
			FROM thread
				LEFT OUTER JOIN image ON
				(thread.image = image.key)
			WHERE thread.authorid = $1`,
		authorID,
	)
	if err != nil {
		return nil, err
	}

	threadList = make([]*Thread, 0)
	for rows.Next() {
		threadItem := &Thread{}
		err = rows.Scan(
			&threadItem.Key,
			&threadItem.Title,
			&threadItem.AuthorID,
			&threadItem.BoardName,
			&threadItem.CreationDateTime,
			&threadItem.ImagePath,
		)
		threadList = append(threadList, threadItem)
	}
	rows.Close()

	if len(threadList) == 0 {
		return nil, errors.New("no threads found with author ID " + string(authorID))
	}

	// update cache
	cacheVersion := m.repoConnection.redis.updateChangeCounter(redThreadAuthorKey)

	threadListCache = make([]Thread, 0, len(threadList))
	for idx := range threadList {
		threadListCache = append(threadListCache, *threadList[idx])
	}
	newCachedData, err := json.Marshal(&threadListCache)
	if err != nil {
		log.Panic(err)
	}
	err = m.repoConnection.redis.set(
		redThreadAuthorKey,
		string(authorID),
		string(newCachedData),
		cacheVersion,
	)
	if err != nil {
		log.Panic(err)
	}

	return threadList, nil
}

func (m *threadModel) getThread(threadID threadKey) (*Thread, error) {
	var threadCache *Thread
	// read from cache
	cachedData, err := m.repoConnection.redis.get(redThreadKey, threadID.String())
	if err == nil {
		threadCache = &Thread{}
		json.Unmarshal([]byte(cachedData), threadCache)

		return threadCache, nil
	}

	// read from db
	row := m.repoConnection.pg.DB.QueryRow(
		`SELECT thread.key, thread.title, thread.authorid, thread.boardname, thread.creationdatetime, image.filepath
			FROM thread
				LEFT OUTER JOIN image ON
				(thread.image = image.key)
			WHERE thread.key = $1`,
		threadID,
	)
	threadItem := &Thread{}
	err = row.Scan(
		&threadItem.Key,
		&threadItem.Title,
		&threadItem.AuthorID,
		&threadItem.BoardName,
		&threadItem.CreationDateTime,
		&threadItem.ImagePath,
	)
	if err != nil {
		return nil, err
	}

	// update cache
	cacheVersion := m.repoConnection.redis.updateChangeCounter(redThreadKey)

	newCachedData, err := json.Marshal(threadItem)
	if err != nil {
		log.Panic(err)
	}
	err = m.repoConnection.redis.set(
		redThreadKey,
		threadID.String(),
		string(newCachedData),
		cacheVersion,
	)
	if err != nil {
		log.Panic(err)
	}

	return threadItem, nil
}

func (m *threadModel) putThread(newThread Thread) (threadKey, error) {
	var imageKeyStr *string
	if newThread.ImageKey != nil {
		strval := newThread.ImageKey.String()
		imageKeyStr = &strval
	}
	row := m.repoConnection.pg.DB.QueryRow(
		`INSERT INTO thread (key, title, authorid, boardname, creationdatetime, image ) VALUES (
			nextval('thread_key_seq'),
			$1, $2, $3, $4, $5
			) RETURNING key;`,
		newThread.Title,
		newThread.AuthorID,
		newThread.BoardName,
		newThread.CreationDateTime,
		imageKeyStr,
	)

	var index threadKey

	err := row.Scan(&index)
	if err != nil {
		return 0, err
	}

	// update cache version
	m.repoConnection.redis.updateChangeCounter(redThreadBoardKey)
	m.repoConnection.redis.updateChangeCounter(redThreadAuthorKey)
	m.repoConnection.redis.updateChangeCounter(redThreadKey)

	return index, nil
}

func newThreadModel(repoConnection *repoHandler) *threadModel {
	return &threadModel{
		repoConnection: repoConnection,
	}
}

// Post is a db structure of post table
type Post struct {
	Key              postKey
	Author           authorKey
	Thread           threadKey
	CreationDateTime time.Time
	Text             string
	ImageKey         *uuid.UUID //sql.NullString
	ImagePath        *string
}

func (p *Post) getImagePath() string {
	if p.ImagePath != nil {
		return *p.ImagePath
	} else {
		return ""
	}
}

type postModel struct {
	repoConnection *repoHandler
}

func (m *postModel) getPostsByThread(threadID threadKey) ([]*Post, error) {
	var (
		postListCache []Post
		postList      []*Post
	)

	// read from cache
	cachedData, err := m.repoConnection.redis.get(redPostThreadKey, threadID.String())
	if err == nil {
		postListCache = make([]Post, 0)
		json.Unmarshal([]byte(cachedData), &postListCache)

		for idx := range postListCache {
			postList = append(postList, &postListCache[idx])
		}
		return postList, nil
	}

	// read from db
	rows, err := m.repoConnection.pg.DB.Query(
		`SELECT post.key, post.author, post.thread, post.creationdatetime, post.text, image.filepath
			FROM post
				LEFT OUTER JOIN image ON
				(post.image = image.key)
			WHERE post.thread = $1`,
		threadID,
	)
	if err != nil {
		return nil, err
	}

	postList = make([]*Post, 0)
	for rows.Next() {
		postItem := &Post{}
		err = rows.Scan(
			&postItem.Key,
			&postItem.Author,
			&postItem.Thread,
			&postItem.CreationDateTime,
			&postItem.Text,
			&postItem.ImagePath,
		)
		postList = append(postList, postItem)
	}
	rows.Close()

	if len(postList) == 0 {
		return nil, errors.New("no posts found with Thread ID" + threadID.String())
	}

	// update cache
	cacheVersion := m.repoConnection.redis.updateChangeCounter(redPostThreadKey)

	postListCache = make([]Post, 0, len(postList))
	for idx := range postList {
		postListCache = append(postListCache, *postList[idx])
	}
	newCachedData, err := json.Marshal(&postListCache)
	if err != nil {
		log.Panic(err)
	}
	err = m.repoConnection.redis.set(
		redPostThreadKey,
		threadID.String(),
		string(newCachedData),
		cacheVersion,
	)
	if err != nil {
		log.Panic(err)
	}

	return postList, nil
}

func (m *postModel) getPostsByAuthor(AuthorID authorKey) ([]*Post, error) {
	var (
		postListCache []Post
		postList      []*Post
	)

	// read from cache
	cachedData, err := m.repoConnection.redis.get(redPostAuthorKey, string(AuthorID))
	if err == nil {
		postListCache = make([]Post, 0)
		json.Unmarshal([]byte(cachedData), &postListCache)

		for idx := range postListCache {
			postList = append(postList, &postListCache[idx])
		}
		return postList, nil
	}

	// read from db
	rows, err := m.repoConnection.pg.DB.Query(
		`SELECT post.key, post.author, post.thread, post.creationdatetime, post.text, image.filepath
			FROM post
				LEFT OUTER JOIN image ON
				(post.image = image.key)
			WHERE post.author = $1`,
		AuthorID,
	)
	if err != nil {
		return nil, err
	}

	postList = make([]*Post, 0)
	for rows.Next() {
		postItem := &Post{}
		err = rows.Scan(
			&postItem.Key,
			&postItem.Author,
			&postItem.Thread,
			&postItem.CreationDateTime,
			&postItem.Text,
			&postItem.ImagePath,
		)
		postList = append(postList, postItem)
	}
	rows.Close()

	if len(postList) == 0 {
		return nil, errors.New("no posts found with Author ID " + string(AuthorID))
	}

	// update cache
	cacheVersion := m.repoConnection.redis.updateChangeCounter(redPostAuthorKey)

	postListCache = make([]Post, 0, len(postList))
	for idx := range postList {
		postListCache = append(postListCache, *postList[idx])
	}
	newCachedData, err := json.Marshal(&postListCache)
	if err != nil {
		log.Panic(err)
	}
	err = m.repoConnection.redis.set(
		redPostAuthorKey,
		string(AuthorID),
		string(newCachedData),
		cacheVersion,
	)
	if err != nil {
		log.Panic(err)
	}

	return postList, nil
}

func (m *postModel) getPost(postID postKey) (*Post, error) {
	var postCache *Post
	// read from cache
	cachedData, err := m.repoConnection.redis.get(redPostKey, postID.String())
	if err == nil {
		postCache = &Post{}
		json.Unmarshal([]byte(cachedData), postCache)

		return postCache, nil
	}

	// read from db
	row := m.repoConnection.pg.DB.QueryRow(
		`SELECT post.key, post.author, post.thread, post.creationdatetime, post.text, image.filepath
			FROM post
				LEFT OUTER JOIN image ON
				(post.image = image.key)
			WHERE post.key = $1`,
		postID,
	)
	postItem := &Post{}
	err = row.Scan(
		&postItem.Key,
		&postItem.Author,
		&postItem.Thread,
		&postItem.CreationDateTime,
		&postItem.Text,
		&postItem.ImagePath,
	)
	if err != nil {
		return nil, err
	}

	// update cache
	cacheVersion := m.repoConnection.redis.updateChangeCounter(redPostKey)

	newCachedData, err := json.Marshal(postItem)
	if err != nil {
		log.Panic(err)
	}
	err = m.repoConnection.redis.set(
		redPostKey,
		postID.String(),
		string(newCachedData),
		cacheVersion,
	)
	if err != nil {
		log.Panic(err)
	}

	return postItem, nil
}

func (m *postModel) putPost(newPost Post) (postKey, error) {
	var imageKeyStr *string
	if newPost.ImageKey != nil {
		strval := newPost.ImageKey.String()
		imageKeyStr = &strval
	}
	row := m.repoConnection.pg.DB.QueryRow(
		`INSERT INTO post (author, thread, creationdatetime, text, image) VALUES (
			$1, $2, $3, $4, $5
			) RETURNING key;`,
		newPost.Author,
		newPost.Thread,
		newPost.CreationDateTime,
		newPost.Text,
		imageKeyStr,
	)

	var index postKey

	err := row.Scan(&index)
	if err != nil {
		return 0, err
	}

	m.repoConnection.redis.updateChangeCounter(redPostAuthorKey)
	m.repoConnection.redis.updateChangeCounter(redPostThreadKey)
	m.repoConnection.redis.updateChangeCounter(redPostKey)

	return index, nil
}

func newPostModel(repoConnection *repoHandler) *postModel {
	return &postModel{
		repoConnection: repoConnection,
	}
}

// Author is a db structure of author table
type Author struct {
	Key authorKey
	// AdminRole bool
}

type authorModel struct {
	repoConnection *repoHandler
}

func (m *authorModel) getAuthor(authorID authorKey) (*Author, error) {
	var authorCache *Author
	// read from cache
	cachedData, err := m.repoConnection.redis.get(redAuthorKey, string(authorID))
	if err == nil {
		authorCache = &Author{}
		json.Unmarshal([]byte(cachedData), authorCache)

		return authorCache, nil
	}

	// read from db
	row := m.repoConnection.pg.DB.QueryRow(
		`SELECT Key
			FROM author
			WHERE key = $1`,
		authorID,
	)
	authorItem := &Author{}
	err = row.Scan(
		&authorItem.Key,
	)
	if err != nil {
		return nil, err
	}

	// update cache
	cacheVersion := m.repoConnection.redis.updateChangeCounter(redAuthorKey)

	newCachedData, err := json.Marshal(authorItem)
	if err != nil {
		log.Panic(err)
	}
	err = m.repoConnection.redis.set(
		redAuthorKey,
		string(authorID),
		string(newCachedData),
		cacheVersion,
	)
	if err != nil {
		log.Panic(err)
	}

	return authorItem, nil
}

func newAuthorModel(repoConnection *repoHandler) *authorModel {
	return &authorModel{
		repoConnection: repoConnection,
	}
}

// Image is a db structure of image table
type Image struct {
	Key      imageKey
	FilePath string
}

type imageModel struct {
	repoConnection *repoHandler
}

func (m *imageModel) isImageExist(image imageKey) bool {
	row := m.repoConnection.pg.DB.QueryRow(
		`SELECT EXISTS( SELECT 1
			FROM image
			WHERE key = $1
			)`,
		uuid.UUID(image).String(),
	)
	// imageExists := &bool{}
	var imageExists *bool
	err := row.Scan(
		&imageExists,
	)
	if err != nil {
		log.Println("error during sql check: ", err)
		return false
	}
	return *imageExists
}

func (m *imageModel) putImage(newImage *Image) error {
	_, err := m.repoConnection.pg.DB.Exec(
		`INSERT INTO image (key, filepath) VALUES (
			$1, $2
			)`,
		uuid.UUID(newImage.Key).String(),
		newImage.FilePath,
	)
	return err
}

func newImageModel(repoConnection *repoHandler) *imageModel {
	return &imageModel{
		repoConnection: repoConnection,
	}
}

type modelContext struct {
	repoConnection *repoHandler
	boardModel     *boardModel
	threadModel    *threadModel
	postModel      *postModel
	authorModel    *authorModel
	imageModel     *imageModel
}

var context *modelContext
var contextSingleton sync.Once

func getmodelContext() *modelContext {
	contextSingleton.Do(func() {
		config := getConfig()
		repoHnd := newRepoHandler(config)

		context = &modelContext{
			repoConnection: repoHnd,
			boardModel:     newBoardModel(repoHnd),
			threadModel:    newThreadModel(repoHnd),
			postModel:      newPostModel(repoHnd),
			authorModel:    newAuthorModel(repoHnd),
			imageModel:     newImageModel(repoHnd),
		}
	})

	return context
}
