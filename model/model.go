package model

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/go-redis/redis"
	"github.com/google/uuid"
	"github.com/ilyakaznacheev/gochan/config"
)

type (
	// BoardKey represents unique board model key
	BoardKey string
	// ThreadKey represents unique thread model key
	ThreadKey int
	// PostKey represents unique post model key
	PostKey int
	// AuthorKey represents unique author model key
	AuthorKey string
	// ImageKey represents unique board model key
	ImageKey uuid.UUID
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

// BoardModelDB is a board model DB interaction interface
type BoardModelDB interface {
	GetBoard(BoardKey) (*Board, error)
}

// ThreadModelDB is a thread model DB interaction interface
type ThreadModelDB interface {
	GetTheadsByBoard(BoardKey) ([]*Thread, error)
	GetThreadsByAuthor(AuthorKey) ([]*Thread, error)
	GetThread(ThreadKey) (*Thread, error)
	PutThread(Thread) (ThreadKey, error)
}

// PostModelDB is a post model DB interaction interface
type PostModelDB interface {
	GetPostsByThread(ThreadKey) ([]*Post, error)
	GetPostsByAuthor(AuthorKey) ([]*Post, error)
	GetPost(PostKey) (*Post, error)
	PutPost(Post) (PostKey, error)
}

// ImageModelDB is a image model DB interaction interface
type ImageModelDB interface {
	IsImageExist(ImageKey) bool
	PutImage(*Image) error
}

// AuthorModelDB is a author model DB interaction interface
type AuthorModelDB interface {
	GetAuthor(AuthorKey) (*Author, error)
}

func (key ThreadKey) String() string {
	return fmt.Sprintf("%d", key)
}

func (key PostKey) String() string {
	return fmt.Sprintf("%d", key)
}

// Board model

// Board is a db structure of board table
type Board struct {
	Key  BoardKey
	Name string
}

type BoardModel struct {
	repoConnection *RepoHandler
	modelDAC       BoardModelDB
}

func NewBoardModel(repoConnection *RepoHandler, modelDAC BoardModelDB) *BoardModel {
	return &BoardModel{
		repoConnection: repoConnection,
		modelDAC:       modelDAC,
	}
}

func (m *BoardModel) GetList() (boardList []*Board) {
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

func (m *BoardModel) GetItem(name BoardKey) (*Board, error) {
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
	go func() {
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
	}()

	return boardItem, nil
}

// Thread model

// Thread is a db structure of thread table
type Thread struct {
	Key              ThreadKey
	Title            string
	AuthorID         AuthorKey
	BoardName        BoardKey
	CreationDateTime time.Time
	ImageKey         *uuid.UUID //sql.NullString
	ImagePath        *string
}

func (t *Thread) GetImagePath() string {
	if t.ImagePath != nil {
		return *t.ImagePath
	} else {
		return ""
	}
}

type ThreadModel struct {
	repoConnection *RepoHandler
	modelDAC       ThreadModelDB
}

func NewThreadModel(repoConnection *RepoHandler, modelDAC ThreadModelDB) *ThreadModel {
	return &ThreadModel{
		repoConnection: repoConnection,
		modelDAC:       modelDAC,
	}
}

func (m *ThreadModel) GetTheadsByBoard(boardName BoardKey) ([]*Thread, error) {
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
	go func() {
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
	}()

	return threadList, nil
}

func (m *ThreadModel) GetThreadsByAuthor(authorID AuthorKey) ([]*Thread, error) {
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
	go func() {
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
	}()

	return threadList, nil
}

func (m *ThreadModel) GetThread(threadID ThreadKey) (*Thread, error) {
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
	go func() {
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
	}()

	return threadItem, nil
}

func (m *ThreadModel) PutThread(newThread Thread) (ThreadKey, error) {
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

	var index ThreadKey

	err := row.Scan(&index)
	if err != nil {
		return 0, err
	}

	// update cache version
	go func() {
		m.repoConnection.redis.updateChangeCounter(redThreadBoardKey)
		m.repoConnection.redis.updateChangeCounter(redThreadAuthorKey)
		m.repoConnection.redis.updateChangeCounter(redThreadKey)
	}()

	return index, nil
}

// Post is a db structure of post table
type Post struct {
	Key              PostKey
	Author           AuthorKey
	Thread           ThreadKey
	CreationDateTime time.Time
	Text             string
	ImageKey         *uuid.UUID //sql.NullString
	ImagePath        *string
}

func (p *Post) GetImagePath() string {
	if p.ImagePath != nil {
		return *p.ImagePath
	} else {
		return ""
	}
}

type PostModel struct {
	repoConnection *RepoHandler
	modelDAC       PostModelDB
}

func NewPostModel(repoConnection *RepoHandler, modelDAC PostModelDB) *PostModel {
	return &PostModel{
		repoConnection: repoConnection,
		modelDAC:       modelDAC,
	}
}

func (m *PostModel) GetPostsByThread(threadID ThreadKey) ([]*Post, error) {
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
	go func() {
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
	}()

	return postList, nil
}

func (m *PostModel) GetPostsByAuthor(AuthorID AuthorKey) ([]*Post, error) {
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
	go func() {
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
	}()

	return postList, nil
}

func (m *PostModel) GetPost(postID PostKey) (*Post, error) {
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
	go func() {
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
	}()

	return postItem, nil
}

func (m *PostModel) PutPost(newPost Post) (PostKey, error) {
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

	var index PostKey

	err := row.Scan(&index)
	if err != nil {
		return 0, err
	}

	go func() {
		m.repoConnection.redis.updateChangeCounter(redPostAuthorKey)
		m.repoConnection.redis.updateChangeCounter(redPostThreadKey)
		m.repoConnection.redis.updateChangeCounter(redPostKey)
	}()

	return index, nil
}

// Author model

// Author is a db structure of author table
type Author struct {
	Key AuthorKey
	// AdminRole bool
}

type AuthorModel struct {
	repoConnection *RepoHandler
	modelDAC       AuthorModelDB
}

func NewAuthorModel(repoConnection *RepoHandler, modelDAC AuthorModelDB) *AuthorModel {
	return &AuthorModel{
		repoConnection: repoConnection,
		modelDAC:       modelDAC,
	}
}

func (m *AuthorModel) GetAuthor(authorID AuthorKey) (*Author, error) {
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
	go func() {
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
	}()

	return authorItem, nil
}

// Image is a db structure of image table
type Image struct {
	Key      ImageKey
	FilePath string
}

type ImageModel struct {
	repoConnection *RepoHandler
	modelDAC       ImageModelDB
}

func NewImageModel(repoConnection *RepoHandler, modelDAC ImageModelDB) *ImageModel {
	return &ImageModel{
		repoConnection: repoConnection,
		modelDAC:       modelDAC,
	}
}

func (m *ImageModel) IsImageExist(image ImageKey) bool {
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

func (m *ImageModel) PutImage(newImage *Image) error {
	_, err := m.repoConnection.pg.DB.Exec(
		`INSERT INTO image (key, filepath) VALUES (
			$1, $2
			)`,
		uuid.UUID(newImage.Key).String(),
		newImage.FilePath,
	)
	return err
}

// Redis

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

func newRedisClient(config *config.ConfigData) *redisClient {
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

func NewPGClient(config *config.ConfigData) (db *pgClient, err error) {

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

type RepoHandler struct {
	redis *redisClient
	pg    *pgClient
}

func NewRepoHandler(config *config.ConfigData) *RepoHandler {
	redis := newRedisClient(config)
	pg, err := NewPGClient(config)
	if err != nil {
		log.Fatal(err)
	}

	return &RepoHandler{
		redis: redis,
		pg:    pg,
	}
}

// TEMP!!!
func (r *RepoHandler) GetDB() *sql.DB {
	return r.pg.DB
}
