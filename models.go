package gochan

import (
	"database/sql"
	"fmt"
	"sync"

	"github.com/ilyakaznacheev/gochan/config"
	"github.com/ilyakaznacheev/gochan/db"
	"github.com/ilyakaznacheev/gochan/model"

	_ "github.com/lib/pq"
)

type modelContext struct {
	repoConnection *model.RepoHandler
	boardModel     *model.BoardModel
	threadModel    *model.ThreadModel
	postModel      *model.PostModel
	authorModel    *model.AuthorModel
	imageModel     *model.ImageModel
}

var mctx *modelContext
var contextSingleton sync.Once

func getmodelContext(config *config.ConfigData) *modelContext {
	contextSingleton.Do(func() {
		repoHnd := model.NewRepoHandler(config)

		// pg, err := model.NewPGClient(config)
		// if err != nil {
		// 	log.Fatal(err)
		// }

		connStr := fmt.Sprintf(
			"user=%s dbname=%s password=%s host=%s sslmode=%s",
			config.Database.User,
			config.Database.Name,
			config.Database.Pass,
			config.Database.Address,
			config.Database.SSL,
		)

		dbConn, _ := sql.Open("postgres", connStr)

		mctx = &modelContext{
			repoConnection: repoHnd,
			boardModel:     model.NewBoardModel(repoHnd, db.NewBoardDAC(dbConn)),
			threadModel:    model.NewThreadModel(repoHnd, db.NewThreadDAC(dbConn)),
			postModel:      model.NewPostModel(repoHnd, db.NewPostDAC(dbConn)),
			authorModel:    model.NewAuthorModel(repoHnd, db.NewAuthorDAC(dbConn)),
			imageModel:     model.NewImageModel(repoHnd, db.NewImageDAC(dbConn)),
		}
	})

	return mctx
}
