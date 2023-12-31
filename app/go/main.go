package main

// ISUCON的な参考: https://github.com/isucon/isucon12-qualify/blob/main/webapp/go/isuports.go#L336
// sqlx的な参考: https://jmoiron.github.io/sqlx/

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"sync"

	"github.com/go-json-experiment/json"
	"github.com/go-sql-driver/mysql"
	"github.com/gorilla/sessions"
	"github.com/jmoiron/sqlx"
	echoInt "github.com/kaz/pprotein/integration/echov4"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echolog "github.com/labstack/gommon/log"
	cmap "github.com/orcaman/concurrent-map/v2"
	"golang.org/x/sync/errgroup"
)

const (
	listenPort                     = 8080
	powerDNSSubdomainAddressEnvKey = "ISUCON13_POWERDNS_SUBDOMAIN_ADDRESS"
)

var (
	powerDNSSubdomainAddress string
	dbConn                   *sqlx.DB
	secret                   = []byte("isucon13_session_cookiestore_defaultsecret")
)

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	if secretKey, ok := os.LookupEnv("ISUCON13_SESSION_SECRETKEY"); ok {
		secret = []byte(secretKey)
	}
}

type InitializeResponse struct {
	Language string `json:"language"`
}

func connectDB(logger echo.Logger) (*sqlx.DB, error) {
	const (
		networkTypeEnvKey = "ISUCON13_MYSQL_DIALCONFIG_NET"
		addrEnvKey        = "ISUCON13_MYSQL_DIALCONFIG_ADDRESS"
		portEnvKey        = "ISUCON13_MYSQL_DIALCONFIG_PORT"
		userEnvKey        = "ISUCON13_MYSQL_DIALCONFIG_USER"
		passwordEnvKey    = "ISUCON13_MYSQL_DIALCONFIG_PASSWORD"
		dbNameEnvKey      = "ISUCON13_MYSQL_DIALCONFIG_DATABASE"
		parseTimeEnvKey   = "ISUCON13_MYSQL_DIALCONFIG_PARSETIME"
	)

	conf := mysql.NewConfig()

	// 環境変数がセットされていなかった場合でも一旦動かせるように、デフォルト値を入れておく
	// この挙動を変更して、エラーを出すようにしてもいいかもしれない
	conf.Net = "tcp"
	conf.Addr = net.JoinHostPort("127.0.0.1", "3306")
	conf.User = "isucon"
	conf.Passwd = "isucon"
	conf.DBName = "isupipe"
	conf.ParseTime = true
	conf.InterpolateParams = true

	if v, ok := os.LookupEnv(networkTypeEnvKey); ok {
		conf.Net = v
	}
	if addr, ok := os.LookupEnv(addrEnvKey); ok {
		if port, ok2 := os.LookupEnv(portEnvKey); ok2 {
			conf.Addr = net.JoinHostPort(addr, port)
		} else {
			conf.Addr = net.JoinHostPort(addr, "3306")
		}
	}
	if v, ok := os.LookupEnv(userEnvKey); ok {
		conf.User = v
	}
	if v, ok := os.LookupEnv(passwordEnvKey); ok {
		conf.Passwd = v
	}
	if v, ok := os.LookupEnv(dbNameEnvKey); ok {
		conf.DBName = v
	}
	if v, ok := os.LookupEnv(parseTimeEnvKey); ok {
		parseTime, err := strconv.ParseBool(v)
		if err != nil {
			return nil, fmt.Errorf("failed to parse environment variable '%s' as bool: %+v", parseTimeEnvKey, err)
		}
		conf.ParseTime = parseTime
	}

	db, err := sqlx.Open("mysql", conf.FormatDSN())
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(64)
	db.SetMaxIdleConns(64)

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

var (
	tagIDCache   = cmap.New[*Tag]()
	tagNameCache = cmap.New[*Tag]()
)

func resetTagCache(ctx context.Context) error {
	tx, err := dbConn.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var tagModels []*Tag
	if err := tx.SelectContext(ctx, &tagModels, "SELECT * FROM tags"); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	tagIDCache.Clear()
	tagNameCache.Clear()
	for i := range tagModels {
		tagID := strconv.FormatInt(tagModels[i].ID, 10)
		tagIDCache.Set(tagID, &Tag{
			ID:   tagModels[i].ID,
			Name: tagModels[i].Name,
		})
		tagNameCache.Set(tagModels[i].Name, &Tag{
			ID:   tagModels[i].ID,
			Name: tagModels[i].Name,
		})
	}
	return nil
}

func getTagByID(id int64) (*Tag, error) {
	if tag, ok := tagIDCache.Get(strconv.FormatInt(id, 10)); ok {
		return tag, nil
	}

	tag := &Tag{}
	if err := dbConn.Get(tag, "SELECT * FROM tags WHERE id = ?", id); err != nil {
		return nil, err
	}

	tagIDCache.Set(strconv.FormatInt(id, 10), tag)
	tagNameCache.Set(tag.Name, tag)
	return tag, nil
}

func getTagByName(name string) (*Tag, error) {
	if tag, ok := tagNameCache.Get(name); ok {
		return tag, nil
	}

	tag := &Tag{}
	if err := dbConn.Get(tag, "SELECT * FROM tags WHERE name = ?", name); err != nil {
		return nil, err
	}

	tagIDCache.Set(strconv.FormatInt(tag.ID, 10), tag)
	tagNameCache.Set(tag.Name, tag)
	return tag, nil
}

func initializeHandler(c echo.Context) error {
	if out, err := exec.Command("../sql/init.sh").CombinedOutput(); err != nil {
		c.Logger().Warnf("init.sh failed with err=%s", string(out))
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to initialize: "+err.Error())
	}

	eg := errgroup.Group{}
	eg.Go(func() error {
		if _, err := http.Post("http://192.168.0.11:8080/api/initialize/slave", "application/json", nil); err != nil {
			return err
		}
		return nil
	})
	eg.Go(func() error {
		if _, err := http.Post("http://192.168.0.12:8080/api/initialize/slave", "application/json", nil); err != nil {
			return err
		}
		return nil
	})
	eg.Go(func() error {
		if _, err := http.Post("http://192.168.0.13:8080/api/initialize/slave", "application/json", nil); err != nil {
			return err
		}
		return nil
	})

	if err := eg.Wait(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to initialize: "+err.Error())
	}

	go func() {
		if _, err := http.Get("https://pprotein.tokyoscience.jp/api/group/collect"); err != nil {
			log.Printf("failed to communicate with pprotein: %v", err)
		}
	}()

	c.Request().Header.Add("Content-Type", "application/json;charset=utf-8")
	return c.JSON(http.StatusOK, InitializeResponse{
		Language: "golang",
	})
}
func initializeSlaveHandler(c echo.Context) error {
	resetSubdomains()

	cacheLock.Lock()
	rrCache = sync.Map{}
	userCache = sync.Map{}
	ngwordsCache = sync.Map{}
	userFullCache = sync.Map{}
	livestreamCache = sync.Map{}
	userNameIconCache = sync.Map{}
	livestreamTagsCache = sync.Map{}
	cacheLock.Unlock()

	ctx := c.Request().Context()
	if err := resetTagCache(ctx); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to reset tag cache: "+err.Error())
	}

	return c.NoContent(http.StatusNoContent)
}

//type JSONSerializer interface {
//	Serialize(c Context, i interface{}, indent string) error
//	Deserialize(c Context, i interface{}) error
//}

type v2JSONSerializer struct {
}

func (s *v2JSONSerializer) Serialize(c echo.Context, i interface{}, indent string) error {
	return json.MarshalWrite(c.Response(), i)
}

func (s *v2JSONSerializer) Deserialize(c echo.Context, i interface{}) error {
	return json.UnmarshalRead(c.Request().Body, i)
}

func main() {
	go startDNS()

	e := echo.New()

	e.Debug = true
	e.Logger.SetLevel(echolog.DEBUG)
	e.Use(middleware.Logger())
	cookieStore := sessions.NewCookieStore(secret)
	cookieStore.Options.Domain = "*.u.isucon.dev"
	e.Use(session.Middleware(cookieStore))
	// e.Use(middleware.Recover())

	e.JSONSerializer = &v2JSONSerializer{}

	// pprotein
	echoInt.Integrate(e)

	// 初期化
	e.POST("/api/initialize", initializeHandler)
	e.POST("/api/initialize/slave", initializeSlaveHandler)

	// top
	e.GET("/api/tag", getTagHandler)
	e.GET("/api/user/:username/theme", getStreamerThemeHandler)

	// livestream
	// reserve livestream
	e.POST("/api/livestream/reservation", reserveLivestreamHandler)
	// list livestream
	e.GET("/api/livestream/search", searchLivestreamsHandler)
	e.GET("/api/livestream", getMyLivestreamsHandler)
	e.GET("/api/user/:username/livestream", getUserLivestreamsHandler)
	// get livestream
	e.GET("/api/livestream/:livestream_id", getLivestreamHandler)
	// get polling livecomment timeline
	e.GET("/api/livestream/:livestream_id/livecomment", getLivecommentsHandler)
	// ライブコメント投稿
	e.POST("/api/livestream/:livestream_id/livecomment", postLivecommentHandler)
	e.POST("/api/livestream/:livestream_id/reaction", postReactionHandler)
	e.GET("/api/livestream/:livestream_id/reaction", getReactionsHandler)

	// (配信者向け)ライブコメントの報告一覧取得API
	e.GET("/api/livestream/:livestream_id/report", getLivecommentReportsHandler)
	e.GET("/api/livestream/:livestream_id/ngwords", getNgwords)
	// ライブコメント報告
	e.POST("/api/livestream/:livestream_id/livecomment/:livecomment_id/report", reportLivecommentHandler)
	// 配信者によるモデレーション (NGワード登録)
	e.POST("/api/livestream/:livestream_id/moderate", moderateHandler)

	// livestream_viewersにINSERTするため必要
	// ユーザ視聴開始 (viewer)
	e.POST("/api/livestream/:livestream_id/enter", enterLivestreamHandler)
	// ユーザ視聴終了 (viewer)
	e.DELETE("/api/livestream/:livestream_id/exit", exitLivestreamHandler)

	// user
	e.POST("/api/register", registerHandler)
	e.POST("/api/login", loginHandler)
	e.GET("/api/user/me", getMeHandler)
	// フロントエンドで、配信予約のコラボレーターを指定する際に必要
	e.GET("/api/user/:username", getUserHandler)
	e.GET("/api/user/:username/statistics", getUserStatisticsHandler)
	e.GET("/api/user/:username/icon", getIconHandler)
	e.POST("/api/icon", postIconHandler)
	e.POST("/api/internal/icon", postInternalIconHandler)

	// stats
	// ライブ配信統計情報
	e.GET("/api/livestream/:livestream_id/statistics", getLivestreamStatisticsHandler)

	// 課金情報
	e.GET("/api/payment", GetPaymentResult)

	e.HTTPErrorHandler = errorResponseHandler

	// DB接続
	conn, err := connectDB(e.Logger)
	if err != nil {
		e.Logger.Errorf("failed to connect db: %v", err)
		os.Exit(1)
	}
	defer conn.Close()
	dbConn = conn

	// キャッシュの初期化
	if err := resetTagCache(context.Background()); err != nil {
		e.Logger.Errorf("failed to reset tag cache: %v", err)
		os.Exit(1)
	}

	subdomainAddr, ok := os.LookupEnv(powerDNSSubdomainAddressEnvKey)
	if !ok {
		e.Logger.Errorf("environ %s must be provided", powerDNSSubdomainAddressEnvKey)
		os.Exit(1)
	}
	powerDNSSubdomainAddress = subdomainAddr

	// HTTPサーバ起動
	listenAddr := net.JoinHostPort("", strconv.Itoa(listenPort))
	if err := e.Start(listenAddr); err != nil {
		e.Logger.Errorf("failed to start HTTP server: %v", err)
		os.Exit(1)
	}
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func errorResponseHandler(err error, c echo.Context) {
	c.Logger().Errorf("error at %s: %+v", c.Path(), err)
	if he, ok := err.(*echo.HTTPError); ok {
		if e := c.JSON(he.Code, &ErrorResponse{Error: err.Error()}); e != nil {
			c.Logger().Errorf("%+v", e)
		}
		return
	}

	if e := c.JSON(http.StatusInternalServerError, &ErrorResponse{Error: err.Error()}); e != nil {
		c.Logger().Errorf("%+v", e)
	}
}
