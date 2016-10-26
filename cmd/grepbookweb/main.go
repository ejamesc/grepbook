package main

import (
	"fmt"
	"html/template"
	"log"
	"path"

	"github.com/boltdb/bolt"
	"github.com/ejamesc/grepbook"
	"github.com/gorilla/context"
	"github.com/gorilla/sessions"
	"github.com/justinas/alice"
	"github.com/kardianos/osext"
	"github.com/microcosm-cc/bluemonday"
	"github.com/spf13/viper"
	"github.com/unrolled/render"
)

// App is the main app.
type App struct {
	rndr   *render.Render
	router *Router
	store  *sessions.CookieStore
	gp     globalPresenter
	bm     *bluemonday.Policy
	logr   appLogger
}

// globalPresenter contains the fields necessary for presenting in all templates
type globalPresenter struct {
	SiteName    string
	Description string
	SiteURL     string
}

// localPresenter contains the fields necessary for specific pages.
type localPresenter struct {
	PageTitle string
	PageURL   string
	User      *grepbook.User
	globalPresenter
}

func SetupApp(r *Router, cookieSecretKey []byte, directoryPath string) *App {
	rndr := render.New(render.Options{
		Directory:  path.Join(directoryPath, "templates"),
		Extensions: []string{".html"},
		Layout:     "base",
		Funcs: []template.FuncMap{
			template.FuncMap{
				"dateFmt": dateFmt,
			}},
	})

	gp := globalPresenter{
		SiteName:    "Grepbook",
		Description: "Grepbook is for reviewing books.",
		SiteURL:     "book.elijames.org",
	}

	bm := bluemonday.UGCPolicy()
	ml := newLogger()

	return &App{
		rndr:   rndr,
		router: r,
		gp:     gp,
		store:  sessions.NewCookieStore(cookieSecretKey),
		bm:     bm,
		logr:   ml,
	}
}

func main() {
	pwd, err := osext.ExecutableFolder()
	if err != nil {
		log.Fatalf("cannot retrieve present working directory: %s", err)
	}

	boltdb, err := bolt.Open(path.Join(pwd, "grepbook.db"), 0600, nil)
	if err != nil {
		log.Fatal("unable to open bolt db: %s", err)
	}
	db := &grepbook.DB{boltdb}
	err = db.CreateAllBuckets()
	if err != nil {
		log.Fatal("unable to create all buckets: %s", err)
	}

	// Load configuration
	viper.SetConfigName("config")
	viper.AddConfigPath(pwd)
	viper.AddConfigPath("/Users/cedric/Projects/gocode/src/github.com/ejamesc/grepbook")
	err = viper.ReadInConfig() // Find and read the config file
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}

	r := NewRouter()
	cookieSecretKey := viper.GetString("cookie-secret")
	a := SetupApp(r, []byte(cookieSecretKey), pwd)

	common := alice.New(context.ClearHandler, a.loggingHandler, a.recoverHandler, a.userMiddlewareGenerator(db))

	r.Get("/", common.Then(a.Wrap(a.IndexHandler(db))))
}
