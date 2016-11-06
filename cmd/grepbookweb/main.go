package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
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

const isDevelopment = true

// App is the main app.
type App struct {
	rndr   *render.Render
	router *Router
	store  *sessions.CookieStore
	gp     globalPresenter
	bm     *bluemonday.Policy
	logr   appLogger
}

func (a *App) GetStore() *sessions.CookieStore {
	return a.store
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

func SetupApp(r *Router, logger appLogger, cookieSecretKey []byte, templateDirectoryPath string) *App {
	rndr := render.New(render.Options{
		Directory:  templateDirectoryPath,
		Extensions: []string{".html"},
		Layout:     "base",
		Funcs: []template.FuncMap{
			template.FuncMap{
				"datefmt": dateFmt,
			}},
	})

	gp := globalPresenter{
		SiteName:    "Grepbook",
		Description: "Grepbook is for reviewing books.",
		SiteURL:     "book.elijames.org",
	}

	bm := bluemonday.UGCPolicy()

	return &App{
		rndr:   rndr,
		router: r,
		gp:     gp,
		store:  sessions.NewCookieStore(cookieSecretKey),
		bm:     bm,
		logr:   logger,
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
	defer db.Close()
	err = db.CreateAllBuckets()
	if err != nil {
		log.Fatal("unable to create all buckets: %s", err)
	}

	// Load configuration
	err = LoadConfiguration(pwd)
	if err != nil && viper.GetBool("isProduction") {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	staticFilePath := path.Join(viper.GetString("path"), "static")
	templateFolderPath := path.Join(viper.GetString("path"), "templates")

	r := NewRouter()
	cookieSecretKey := viper.GetString("cookieSecret")
	logr := newLogger()
	a := SetupApp(r, logr, []byte(cookieSecretKey), templateFolderPath)

	common := alice.New(context.ClearHandler, a.loggingHandler, a.recoverHandler, a.userMiddlewareGenerator(db))
	auth := common.Append(a.authMiddleware)

	r.Get("/", common.Then(a.Wrap(a.IndexHandler(db))))
	r.Get("/about", common.Then(a.Wrap(a.AboutHandler())))

	r.Post("/summaries", auth.Then(a.Wrap(a.CreateBookReviewHandler(db))))
	r.Get("/summaries/:id", common.Then(a.Wrap(a.ReadHandler(db))))

	r.Get("/login", common.Then(a.Wrap(a.LoginPageHandler())))
	r.Post("/login", common.Then(a.Wrap(a.LoginPostHandler(db))))

	r.Post("/logout", common.Then(a.Wrap(a.LogoutHandler())))

	r.Get("/signup", common.Then(a.Wrap(a.SignupPageHandler(db))))
	r.Post("/signup", common.Then(a.Wrap(a.SignupPostHandler(db))))

	r.ServeFiles("/static/*filepath", http.Dir(staticFilePath))

	http.ListenAndServe(":3000", r)
}

func LoadConfiguration(pwd string) error {
	viper.SetConfigName("grepbook-config")
	viper.AddConfigPath(pwd)
	devPath := "/Users/cedric/Projects/gocode/src/github.com/ejamesc/grepbook/cmd/grepbookweb"
	viper.AddConfigPath(devPath)

	viper.SetDefault("path", devPath)
	viper.SetDefault("cookieSecret", "@%3V?#ay!ONfzV7N&3|{?[YT6-gDHgZIhP_;qaw5e7i3t`SAT)w&+GO*>w2EX+[5")
	viper.SetDefault("isProduction", true)
	return viper.ReadInConfig() // Find and read the config file
}
