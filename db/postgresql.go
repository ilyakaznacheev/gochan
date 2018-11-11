package db

import (
	"database/sql"
	"errors"
	"log"

	"github.com/google/uuid"
	_ "github.com/lib/pq" // use Postgres driver

	"github.com/ilyakaznacheev/gochan/model"
)

// BoardDAC is a board table DAC
type BoardDAC struct {
	db *sql.DB
}

// NewBoardDAC creates BoardDAC instance
func NewBoardDAC(db *sql.DB) *BoardDAC {
	return &BoardDAC{db}
}

// GetBoardList returns board list
func (m *BoardDAC) GetBoardList() ([]*model.Board, error) {
	rows, err := m.db.Query(`SELECT key, name FROM board`)
	if err != nil {
		return nil, err
	}
	boardList := make([]*model.Board, 0)
	for rows.Next() {
		boardItem := &model.Board{}
		err = rows.Scan(&boardItem.Key, &boardItem.Name)
		boardList = append(boardList, boardItem)
	}
	rows.Close()
	return boardList, nil
}

// GetBoard returns board data
func (m *BoardDAC) GetBoard(key model.BoardKey) (*model.Board, error) {
	row := m.db.QueryRow(
		`SELECT key, name 
			FROM board
			WHERE key = $1`,
		key,
	)

	boardItem := &model.Board{}
	err := row.Scan(
		&boardItem.Key,
		&boardItem.Name,
	)

	return boardItem, err
}

// ThreadDAC is a thread table DAC
type ThreadDAC struct {
	db *sql.DB
}

// NewThreadDAC creates ThreadDAC instance
func NewThreadDAC(db *sql.DB) *ThreadDAC {
	return &ThreadDAC{db}
}

// GetTheadsByBoard returns threads of certain board
func (m *ThreadDAC) GetTheadsByBoard(boardName model.BoardKey) ([]*model.Thread, error) {
	rows, err := m.db.Query(
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

	threadList := make([]*model.Thread, 0)
	for rows.Next() {
		threadItem := &model.Thread{}
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
	return threadList, nil
}

// GetThreadsByAuthor returns threads of certain author
func (m *ThreadDAC) GetThreadsByAuthor(authorKey model.AuthorKey) ([]*model.Thread, error) {
	rows, err := m.db.Query(
		`SELECT thread.key, thread.title, thread.authorid, thread.boardname, thread.creationdatetime, image.filepath
			FROM thread
				LEFT OUTER JOIN image ON
				(thread.image = image.key)
			WHERE thread.authorid = $1`,
		authorKey,
	)
	if err != nil {
		return nil, err
	}

	threadList := make([]*model.Thread, 0)
	for rows.Next() {
		threadItem := &model.Thread{}
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
		return nil, errors.New("no threads found with author ID " + string(authorKey))
	}
	return threadList, nil
}

// GetThread returns thread data
func (m *ThreadDAC) GetThread(threadKey model.ThreadKey) (*model.Thread, error) {
	row := m.db.QueryRow(
		`SELECT thread.key, thread.title, thread.authorid, thread.boardname, thread.creationdatetime, image.filepath
			FROM thread
				LEFT OUTER JOIN image ON
				(thread.image = image.key)
			WHERE thread.key = $1`,
		threadKey,
	)
	threadItem := &model.Thread{}
	err := row.Scan(
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
	return threadItem, nil
}

// PutThread creates new thread
func (m *ThreadDAC) PutThread(newThread model.Thread) (model.ThreadKey, error) {
	var imageKeyStr *string
	if newThread.ImageKey != nil {
		strval := newThread.ImageKey.String()
		imageKeyStr = &strval
	}
	row := m.db.QueryRow(
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

	var index model.ThreadKey

	err := row.Scan(&index)
	if err != nil {
		return 0, err
	}
	return index, nil
}

// PostDAC is a post table DAC
type PostDAC struct {
	db *sql.DB
}

// NewPostDAC creates PostDAC instance
func NewPostDAC(db *sql.DB) *PostDAC {
	return &PostDAC{db}
}

// GetPostsByThread returns posts of certain thread
func (m *PostDAC) GetPostsByThread(threadKey model.ThreadKey) ([]*model.Post, error) {
	rows, err := m.db.Query(
		`SELECT post.key, post.author, post.thread, post.creationdatetime, post.text, image.filepath
			FROM post
				LEFT OUTER JOIN image ON
				(post.image = image.key)
			WHERE post.thread = $1`,
		threadKey,
	)
	if err != nil {
		return nil, err
	}

	postList := make([]*model.Post, 0)
	for rows.Next() {
		postItem := &model.Post{}
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
		return nil, errors.New("no posts found with Thread ID" + threadKey.String())
	}
	return postList, nil

}

// GetPostsByAuthor returns posts of certain author
func (m *PostDAC) GetPostsByAuthor(authorKey model.AuthorKey) ([]*model.Post, error) {
	rows, err := m.db.Query(
		`SELECT post.key, post.author, post.thread, post.creationdatetime, post.text, image.filepath
			FROM post
				LEFT OUTER JOIN image ON
				(post.image = image.key)
			WHERE post.author = $1`,
		authorKey,
	)
	if err != nil {
		return nil, err
	}

	postList := make([]*model.Post, 0)
	for rows.Next() {
		postItem := &model.Post{}
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
		return nil, errors.New("no posts found with Author ID " + string(authorKey))
	}
	return postList, nil
}

// GetPost returns post data
func (m *PostDAC) GetPost(postKey model.PostKey) (*model.Post, error) {
	row := m.db.QueryRow(
		`SELECT post.key, post.author, post.thread, post.creationdatetime, post.text, image.filepath
			FROM post
				LEFT OUTER JOIN image ON
				(post.image = image.key)
			WHERE post.key = $1`,
		postKey,
	)
	postItem := &model.Post{}
	err := row.Scan(
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
	return postItem, nil
}

// PutPost creates a new post
func (m *PostDAC) PutPost(newPost model.Post) (model.PostKey, error) {
	var imageKeyStr *string
	if newPost.ImageKey != nil {
		strval := newPost.ImageKey.String()
		imageKeyStr = &strval
	}
	row := m.db.QueryRow(
		`INSERT INTO post (author, thread, creationdatetime, text, image) VALUES (
			$1, $2, $3, $4, $5
			) RETURNING key;`,
		newPost.Author,
		newPost.Thread,
		newPost.CreationDateTime,
		newPost.Text,
		imageKeyStr,
	)

	var index model.PostKey

	err := row.Scan(&index)
	if err != nil {
		return 0, err
	}
	return index, nil
}

// ImageDAC is a image table DAC
type ImageDAC struct {
	db *sql.DB
}

// NewImageDAC creates ImageDAC instance
func NewImageDAC(db *sql.DB) *ImageDAC {
	return &ImageDAC{db}
}

// IsImageExist checks image existance by key
func (m *ImageDAC) IsImageExist(imageKey model.ImageKey) bool {
	row := m.db.QueryRow(
		`SELECT EXISTS( SELECT 1
			FROM image
			WHERE key = $1
			)`,
		uuid.UUID(imageKey).String(),
	)

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

// PutImage creates a new image
func (m *ImageDAC) PutImage(newImage *model.Image) error {
	_, err := m.db.Exec(
		`INSERT INTO image (key, filepath) VALUES (
			$1, $2
			)`,
		uuid.UUID(newImage.Key).String(),
		newImage.FilePath,
	)
	return err
}

// AuthorDAC is a author table DAC
type AuthorDAC struct {
	db *sql.DB
}

// NewAuthorDAC creates AuthorDAC instance
func NewAuthorDAC(db *sql.DB) *AuthorDAC {
	return &AuthorDAC{db}
}

// GetAuthor returns author info
func (m *AuthorDAC) GetAuthor(authorKey model.AuthorKey) (*model.Author, error) {
	row := m.db.QueryRow(
		`SELECT Key
				FROM author
				WHERE key = $1`,
		authorKey,
	)
	authorItem := &model.Author{}
	err := row.Scan(
		&authorItem.Key,
	)
	if err != nil {
		return nil, err
	}
	return authorItem, nil
}
