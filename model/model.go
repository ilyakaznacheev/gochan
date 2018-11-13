package model

import (
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

// DB model interfaces

// BoardModelDB is a board model DB interaction interface
type BoardModelDB interface {
	GetBoardList() ([]*Board, error)
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

// Cache model interfaces

// BoardModelCache is a board model cache interaction interface
type BoardModelCache interface {
	GetBoardList() ([]*Board, error)
	GetBoard(BoardKey) (*Board, error)
	SetBoardList([]*Board) error
	SetBoard(BoardKey, *Board) error
}

// ThreadModelCache is a thread model cache interaction interface
type ThreadModelCache interface {
	GetTheadsByBoard(BoardKey) ([]*Thread, error)
	GetThreadsByAuthor(AuthorKey) ([]*Thread, error)
	GetThread(ThreadKey) (*Thread, error)
	SetTheadsByBoard(BoardKey, []*Thread) error
	SetThreadsByAuthor(AuthorKey, []*Thread) error
	SetThread(ThreadKey, *Thread) error
	InvalidateCache()
}

// PostModelCache is a post model cache interaction interface
type PostModelCache interface {
	GetPostsByThread(ThreadKey) ([]*Post, error)
	GetPostsByAuthor(AuthorKey) ([]*Post, error)
	GetPost(PostKey) (*Post, error)
	InvalidateCache()
	SetPostsByThread(ThreadKey, []*Post) error
	SetPostsByAuthor(AuthorKey, []*Post) error
	SetPost(PostKey, *Post) error
}

// AuthorModelCache is a author model cache interaction interface
type AuthorModelCache interface {
	GetAuthor(AuthorKey) (*Author, error)
	SetAuthor(AuthorKey, *Author) error
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

// BoardModel is a board model
type BoardModel struct {
	repoConnection *RepoHandler
	modelDAC       BoardModelDB
}

// NewBoardModel creates new BoardModel
func NewBoardModel(repoConnection *RepoHandler, modelDAC BoardModelDB) *BoardModel {
	return &BoardModel{
		repoConnection: repoConnection,
		modelDAC:       modelDAC,
	}
}

// GetList returns all boards
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
	boardList, _ = m.modelDAC.GetBoardList()

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

// GetItem returns certain board by key
func (m *BoardModel) GetItem(name BoardKey) (*Board, error) {
	// read from cache
	cachedData, err := m.repoConnection.redis.get(redBoardKey, string(name))
	if err == nil {
		boardCache := &Board{}
		json.Unmarshal([]byte(cachedData), boardCache)

		return boardCache, nil
	}

	// read from db
	boardItem, err := m.modelDAC.GetBoard(name)
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

// GetImagePath returns path to image
func (t *Thread) GetImagePath() string {
	if t.ImagePath != nil {
		return *t.ImagePath
	}
	return ""
}

// ThreadModel is a thread model
type ThreadModel struct {
	repoConnection *RepoHandler
	modelDAC       ThreadModelDB
}

// NewThreadModel creates new ThreadModel
func NewThreadModel(repoConnection *RepoHandler, modelDAC ThreadModelDB) *ThreadModel {
	return &ThreadModel{
		repoConnection: repoConnection,
		modelDAC:       modelDAC,
	}
}

// GetTheadsByBoard returns threads by certain board
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
	threadList, err = m.modelDAC.GetTheadsByBoard(boardName)
	if err != nil {
		return nil, err
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

// GetThreadsByAuthor returns threads by certain author
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
	threadList, err = m.modelDAC.GetThreadsByAuthor(authorID)
	if err != nil {
		return nil, err
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

// GetThread returns certain thread by key
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
	threadItem, err := m.modelDAC.GetThread(threadID)
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

// PutThread adds new post into db
func (m *ThreadModel) PutThread(newThread Thread) (ThreadKey, error) {
	index, err := m.modelDAC.PutThread(newThread)
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

// GetImagePath returns path to image
func (p *Post) GetImagePath() string {
	if p.ImagePath != nil {
		return *p.ImagePath
	}
	return ""
}

// PostModel is a post model
type PostModel struct {
	repoConnection *RepoHandler
	modelDAC       PostModelDB
}

// NewPostModel creates new PostModel
func NewPostModel(repoConnection *RepoHandler, modelDAC PostModelDB) *PostModel {
	return &PostModel{
		repoConnection: repoConnection,
		modelDAC:       modelDAC,
	}
}

// GetPostsByThread returns posts by certain thread
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
	postList, err = m.modelDAC.GetPostsByThread(threadID)
	if err != nil {
		return nil, err
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

// GetPostsByAuthor returns posts by certain author
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
	postList, err = m.modelDAC.GetPostsByAuthor(AuthorID)
	if err != nil {
		return nil, err
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

// GetPost certain returns post by key
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
	postItem, err := m.modelDAC.GetPost(postID)
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

// PutPost adds new post into db
func (m *PostModel) PutPost(newPost Post) (PostKey, error) {
	index, err := m.modelDAC.PutPost(newPost)
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

// AuthorModel is an author model
type AuthorModel struct {
	repoConnection *RepoHandler
	modelDAC       AuthorModelDB
}

// NewAuthorModel creates new AuthorModel
func NewAuthorModel(repoConnection *RepoHandler, modelDAC AuthorModelDB) *AuthorModel {
	return &AuthorModel{
		repoConnection: repoConnection,
		modelDAC:       modelDAC,
	}
}

// GetAuthor returns author data
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
	authorItem, err := m.modelDAC.GetAuthor(authorID)
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

// ImageModel is an image model
type ImageModel struct {
	repoConnection *RepoHandler
	modelDAC       ImageModelDB
}

// NewImageModel returns new ImageModel
func NewImageModel(repoConnection *RepoHandler, modelDAC ImageModelDB) *ImageModel {
	return &ImageModel{
		repoConnection: repoConnection,
		modelDAC:       modelDAC,
	}
}

// IsImageExist returns image existance status
func (m *ImageModel) IsImageExist(image ImageKey) bool {
	return m.modelDAC.IsImageExist(image)
}

// PutImage adds new image into table
func (m *ImageModel) PutImage(newImage *Image) error {
	return m.modelDAC.PutImage(newImage)
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

// RepoHandler is a repository handler
type RepoHandler struct {
	redis *redisClient
}

// NewRepoHandler creates new repository handler
func NewRepoHandler(config *config.ConfigData) *RepoHandler {
	redis := newRedisClient(config)

	return &RepoHandler{
		redis: redis,
	}
}
