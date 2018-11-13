package cache

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/go-redis/redis"
	"github.com/ilyakaznacheev/gochan/config"
	"github.com/ilyakaznacheev/gochan/model"
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

// RedisContainer is a redis main container structure
type RedisContainer struct {
	Version int
	Content string
}

type redisClient struct {
	client *redis.Client
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
		return "", model.ErrCacheOutdated
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

// BoardCache is a board table cache manager
type BoardCache struct {
	rc *redisClient
}

// NewBoardCache returns new BoardCache
func NewBoardCache(client *redis.Client) *BoardCache {
	return &BoardCache{&redisClient{client}}
}

// GetBoardList returns board list model cache
func (c *BoardCache) GetBoardList() ([]*model.Board, error) {
	var (
		boardListCache []model.Board
		boardList      []*model.Board
	)

	cachedData, err := c.rc.get(redBoardList, "")
	if err != nil {
		return nil, err
	}

	json.Unmarshal([]byte(cachedData), &boardListCache)

	for idx := range boardListCache {
		boardList = append(boardList, &boardListCache[idx])
	}

	return boardList, nil

}

// GetBoard returns board model cache
func (c *BoardCache) GetBoard(boardKey model.BoardKey) (*model.Board, error) {
	cachedData, err := c.rc.get(redBoardKey, string(boardKey))
	if err != nil {
		return nil, err
	}
	boardCache := &model.Board{}
	json.Unmarshal([]byte(cachedData), boardCache)

	return boardCache, nil
}

// SetBoardList updates board list model cache
func (c *BoardCache) SetBoardList(boardList []*model.Board) error {
	cacheVersion := c.rc.updateChangeCounter(redBoardList)

	boardListCache := make([]model.Board, 0, len(boardList))
	for idx := range boardList {
		boardListCache = append(boardListCache, *boardList[idx])
	}
	newCachedData, err := json.Marshal(&boardListCache)
	if err != nil {
		log.Panic(err)
	}
	err = c.rc.set(
		redBoardList,
		"",
		string(newCachedData),
		cacheVersion,
	)

	return err
}

// SetBoard updates board model cache
func (c *BoardCache) SetBoard(boardKey model.BoardKey, board *model.Board) error {
	cacheVersion := c.rc.updateChangeCounter(redBoardKey)

	newCachedData, err := json.Marshal(board)
	if err != nil {
		log.Panic(err)
	}
	err = c.rc.set(
		redBoardKey,
		string(boardKey),
		string(newCachedData),
		cacheVersion,
	)

	return err
}

// ThreadCache is a thread table cache manager
type ThreadCache struct {
	rc *redisClient
}

// NewThreadCache returns new ThreadCache
func NewThreadCache(client *redis.Client) *ThreadCache {
	return &ThreadCache{&redisClient{client}}
}

// GetTheadsByBoard returns thread model cache by board
func (c *ThreadCache) GetTheadsByBoard(boardKey model.BoardKey) ([]*model.Thread, error) {
	var (
		threadListCache []model.Thread
		threadList      []*model.Thread
	)

	cachedData, err := c.rc.get(redThreadBoardKey, string(boardKey))
	if err != nil {
		return nil, err
	}

	threadListCache = make([]model.Thread, 0)
	json.Unmarshal([]byte(cachedData), &threadListCache)

	for idx := range threadListCache {
		threadList = append(threadList, &threadListCache[idx])
	}
	return threadList, nil

}

// GetThreadsByAuthor returns thread model cache by author
func (c *ThreadCache) GetThreadsByAuthor(authorKey model.AuthorKey) ([]*model.Thread, error) {
	var (
		threadListCache []model.Thread
		threadList      []*model.Thread
	)

	// read from cache
	cachedData, err := c.rc.get(redThreadAuthorKey, string(authorKey))
	if err != nil {
		return nil, err
	}

	threadListCache = make([]model.Thread, 0)
	json.Unmarshal([]byte(cachedData), &threadListCache)

	for idx := range threadListCache {
		threadList = append(threadList, &threadListCache[idx])
	}
	return threadList, nil
}

// GetThread returns thread model cache
func (c *ThreadCache) GetThread(threadKey model.ThreadKey) (*model.Thread, error) {
	cachedData, err := c.rc.get(redThreadKey, string(threadKey))
	if err != nil {
		return nil, err
	}

	threadCache := &model.Thread{}
	json.Unmarshal([]byte(cachedData), threadCache)

	return threadCache, nil
}

// SetTheadsByBoard updates thread model cache by board
func (c *ThreadCache) SetTheadsByBoard(boardKey model.BoardKey, threadList []*model.Thread) error {
	cacheVersion := c.rc.updateChangeCounter(redThreadBoardKey)

	threadListCache := make([]model.Thread, 0, len(threadList))
	for idx := range threadList {
		threadListCache = append(threadListCache, *threadList[idx])
	}
	newCachedData, err := json.Marshal(&threadListCache)
	if err != nil {
		return err
	}
	err = c.rc.set(
		redThreadBoardKey,
		string(boardKey),
		string(newCachedData),
		cacheVersion,
	)

	return err
}

// SetThreadsByAuthor updates thread model cache by author
func (c *ThreadCache) SetThreadsByAuthor(authorkey model.AuthorKey, threadList []*model.Thread) error {
	cacheVersion := c.rc.updateChangeCounter(redThreadAuthorKey)

	threadListCache := make([]model.Thread, 0, len(threadList))
	for idx := range threadList {
		threadListCache = append(threadListCache, *threadList[idx])
	}
	newCachedData, err := json.Marshal(&threadListCache)
	if err != nil {
		return err
	}
	err = c.rc.set(
		redThreadAuthorKey,
		string(authorkey),
		string(newCachedData),
		cacheVersion,
	)

	return err
}

// SetThread updates thread model cache
func (c *ThreadCache) SetThread(threadKey model.ThreadKey, thread *model.Thread) error {
	cacheVersion := c.rc.updateChangeCounter(redThreadKey)

	newCachedData, err := json.Marshal(thread)
	if err != nil {
		return err
	}
	err = c.rc.set(
		redThreadKey,
		string(threadKey),
		string(newCachedData),
		cacheVersion,
	)
	return err
}

// InvalidateCache invalidates old cache
func (c *ThreadCache) InvalidateCache() {
	c.rc.updateChangeCounter(redThreadBoardKey)
	c.rc.updateChangeCounter(redThreadAuthorKey)
	c.rc.updateChangeCounter(redThreadKey)
}

// PostCache is a post table cache manager
type PostCache struct {
	rc *redisClient
}

// NewPostCache returns new PostCache
func NewPostCache(client *redis.Client) *PostCache {
	return &PostCache{&redisClient{client}}
}

// GetPostsByThread returns post model cache by thread
func (c *PostCache) GetPostsByThread(threadKey model.ThreadKey) ([]*model.Post, error) {
	var (
		postListCache []model.Post
		postList      []*model.Post
	)

	cachedData, err := c.rc.get(redPostThreadKey, string(threadKey))
	if err != nil {
		return nil, err
	}
	postListCache = make([]model.Post, 0)
	json.Unmarshal([]byte(cachedData), &postListCache)

	for idx := range postListCache {
		postList = append(postList, &postListCache[idx])
	}
	return postList, nil
}

// GetPostsByAuthor returns post model cache by author
func (c *PostCache) GetPostsByAuthor(authorKey model.AuthorKey) ([]*model.Post, error) {
	var (
		postListCache []model.Post
		postList      []*model.Post
	)

	cachedData, err := c.rc.get(redPostAuthorKey, string(authorKey))
	if err != nil {
		return nil, err
	}

	postListCache = make([]model.Post, 0)
	json.Unmarshal([]byte(cachedData), &postListCache)

	for idx := range postListCache {
		postList = append(postList, &postListCache[idx])
	}
	return postList, nil
}

// GetPost returns post model cache
func (c *PostCache) GetPost(postKey model.PostKey) (*model.Post, error) {
	var postCache *model.Post

	cachedData, err := c.rc.get(redPostKey, string(postKey))
	if err != nil {
		return nil, err
	}

	postCache = &model.Post{}
	json.Unmarshal([]byte(cachedData), postCache)

	return postCache, nil
}

// SetPostsByThread updates post model cache by thread
func (c *PostCache) SetPostsByThread(threadKey model.ThreadKey, postList []*model.Post) error {
	cacheVersion := c.rc.updateChangeCounter(redPostThreadKey)

	postListCache := make([]model.Post, 0, len(postList))
	for idx := range postList {
		postListCache = append(postListCache, *postList[idx])
	}
	newCachedData, err := json.Marshal(&postListCache)
	if err != nil {
		return err
	}
	err = c.rc.set(
		redPostThreadKey,
		string(threadKey),
		string(newCachedData),
		cacheVersion,
	)
	return err
}

// SetPostsByAuthor updates post model cache by author
func (c *PostCache) SetPostsByAuthor(authorKey model.AuthorKey, postList []*model.Post) error {
	cacheVersion := c.rc.updateChangeCounter(redPostAuthorKey)

	postListCache := make([]model.Post, 0, len(postList))
	for idx := range postList {
		postListCache = append(postListCache, *postList[idx])
	}
	newCachedData, err := json.Marshal(&postListCache)
	if err != nil {
		return err
	}
	err = c.rc.set(
		redPostAuthorKey,
		string(authorKey),
		string(newCachedData),
		cacheVersion,
	)
	return err
}

// SetPost updates thread model cache
func (c *PostCache) SetPost(postKey model.PostKey, post *model.Post) error {
	cacheVersion := c.rc.updateChangeCounter(redPostKey)

	newCachedData, err := json.Marshal(post)
	if err != nil {
		return err
	}
	err = c.rc.set(
		redPostKey,
		string(postKey),
		string(newCachedData),
		cacheVersion,
	)
	return err
}

// InvalidateCache invalidates old cache
func (c *PostCache) InvalidateCache() {
	c.rc.updateChangeCounter(redPostAuthorKey)
	c.rc.updateChangeCounter(redPostThreadKey)
	c.rc.updateChangeCounter(redPostKey)
}

// AuthorCache is a author table cache manager
type AuthorCache struct {
	rc *redisClient
}

// NewAuthorCache returns new AuthorCache
func NewAuthorCache(client *redis.Client) *AuthorCache {
	return &AuthorCache{&redisClient{client}}
}

// GetAuthor returns post author cache
func (c *AuthorCache) GetAuthor(authorKey model.AuthorKey) (*model.Author, error) {
	var authorCache *model.Author

	cachedData, err := c.rc.get(redAuthorKey, string(authorKey))
	if err != nil {
		return nil, err
	}

	authorCache = &model.Author{}
	json.Unmarshal([]byte(cachedData), authorCache)

	return authorCache, nil
}

// SetAuthor updates author model cache
func (c *AuthorCache) SetAuthor(authorKey model.AuthorKey, author *model.Author) error {
	cacheVersion := c.rc.updateChangeCounter(redAuthorKey)

	newCachedData, err := json.Marshal(author)
	if err != nil {
		return err
	}
	err = c.rc.set(
		redAuthorKey,
		string(authorKey),
		string(newCachedData),
		cacheVersion,
	)

	return err
}
