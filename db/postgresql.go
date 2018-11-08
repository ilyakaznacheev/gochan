package db

import (
	"database/sql"
	"errors"
	"log"

	"github.com/google/uuid"
	_ "github.com/lib/pq" // use Postgres driver

	"github.com/ilyakaznacheev/gochan"
)

// BoardModel is a board table DAC
type BoardModel struct {
	db *sql.DB
}

// NewBoardModel creates BoardModel instance
func NewBoardModel(db *sql.DB) *BoardModel {
	return &BoardModel{db}
}

// GetBoard returns board data
func (m *BoardModel) GetBoard(key gochan.BoardKey) (*gochan.Board, error) {
	row := m.db.QueryRow(
		`SELECT key, name 
			FROM board
			WHERE key = $1`,
		key,
	)

	boardItem := &gochan.Board{}
	err := row.Scan(
		&boardItem.Key,
		&boardItem.Name,
	)

	return boardItem, err
}

// ThreadModel is a thread table DAC
type ThreadModel struct {
	db *sql.DB
}

// NewThreadModel creates ThreadModel instance
func NewThreadModel(db *sql.DB) *ThreadModel {
	return &ThreadModel{db}
}

// GetTheadsByBoard returns threads of certain board
func (m *ThreadModel) GetTheadsByBoard(boardName gochan.BoardKey) ([]*gochan.Thread, error) {
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

	threadList := make([]*gochan.Thread, 0)
	for rows.Next() {
		threadItem := &gochan.Thread{}
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
func (m *ThreadModel) GetThreadsByAuthor(authorKey gochan.AuthorKey) ([]*gochan.Thread, error) {
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

	threadList := make([]*gochan.Thread, 0)
	for rows.Next() {
		threadItem := &gochan.Thread{}
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
func (m *ThreadModel) GetThread(threadKey gochan.ThreadKey) (*gochan.Thread, error) {
	row := m.db.QueryRow(
		`SELECT thread.key, thread.title, thread.authorid, thread.boardname, thread.creationdatetime, image.filepath
			FROM thread
				LEFT OUTER JOIN image ON
				(thread.image = image.key)
			WHERE thread.key = $1`,
		threadKey,
	)
	threadItem := &gochan.Thread{}
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
func (m *ThreadModel) PutThread(newThread gochan.Thread) (gochan.ThreadKey, error) {
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

	var index gochan.ThreadKey

	err := row.Scan(&index)
	if err != nil {
		return 0, err
	}
	return index, nil
}

// PostModel is a post table DAC
type PostModel struct {
	db *sql.DB
}

// NewPostModel creates PostModel instance
func NewPostModel(db *sql.DB) *PostModel {
	return &PostModel{db}
}

// GetPostsByThread returns posts of certain thread
func (m *PostModel) GetPostsByThread(threadKey gochan.ThreadKey) ([]*gochan.Post, error) {
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

	postList := make([]*gochan.Post, 0)
	for rows.Next() {
		postItem := &gochan.Post{}
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
func (m *PostModel) GetPostsByAuthor(authorKey gochan.AuthorKey) ([]*gochan.Post, error) {
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

	postList := make([]*gochan.Post, 0)
	for rows.Next() {
		postItem := &gochan.Post{}
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
func (m *PostModel) GetPost(postKey gochan.PostKey) (*gochan.Post, error) {
	row := m.db.QueryRow(
		`SELECT post.key, post.author, post.thread, post.creationdatetime, post.text, image.filepath
			FROM post
				LEFT OUTER JOIN image ON
				(post.image = image.key)
			WHERE post.key = $1`,
		postKey,
	)
	postItem := &gochan.Post{}
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
func (m *PostModel) PutPost(newPost gochan.Post) (gochan.PostKey, error) {
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

	var index gochan.PostKey

	err := row.Scan(&index)
	if err != nil {
		return 0, err
	}
	return index, nil
}

// ImageModel is a image table DAC
type ImageModel struct {
	db *sql.DB
}

// NewImageModel creates ImageModel instance
func NewImageModel(db *sql.DB) *ImageModel {
	return &ImageModel{db}
}

// IsImageExist checks image existance by key
func (m *ImageModel) IsImageExist(imageKey gochan.ImageKey) bool {
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
func (m *ImageModel) PutImage(newImage *gochan.Image) error {
	_, err := m.db.Exec(
		`INSERT INTO image (key, filepath) VALUES (
			$1, $2
			)`,
		uuid.UUID(newImage.Key).String(),
		newImage.FilePath,
	)
	return err
}

// AuthorModel is a author table DAC
type AuthorModel struct {
	db *sql.DB
}

// NewAuthorModel creates AuthorModel instance
func NewAuthorModel(db *sql.DB) *AuthorModel {
	return &AuthorModel{db}
}

// GetAuthor returns author info
func (m *AuthorModel) GetAuthor(authorKey gochan.AuthorKey) (*gochan.Author, error) {
	row := m.db.QueryRow(
		`SELECT Key
				FROM author
				WHERE key = $1`,
		authorKey,
	)
	authorItem := &gochan.Author{}
	err := row.Scan(
		&authorItem.Key,
	)
	if err != nil {
		return nil, err
	}
	return authorItem, nil
}
