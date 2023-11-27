package main

// ISUCON的な参考: https://github.com/isucon/isucon12-qualify/blob/main/webapp/go/isuports.go#L336
// sqlx的な参考: https://jmoiron.github.io/sqlx/

import (
	"context"
	"fmt"
	"github.com/gorilla/sessions"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"sync"

	"github.com/go-json-experiment/json"
	"github.com/go-sql-driver/mysql"
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"

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
	store                    sessions.Store
)

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	if secretKey, ok := os.LookupEnv("ISUCON13_SESSION_SECRETKEY"); ok {
		secret = []byte(secretKey)
	}
	store = sessions.NewCookieStore(secret)
}

type InitializeResponse struct {
	Language string `json:"language"`
}

func connectDB() (*sqlx.DB, error) {
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

func initializeHandler(c *fiber.Ctx) error {
	if _, err := exec.Command("../sql/init.sh").CombinedOutput(); err != nil {
		//c.Logger().Warnf("init.sh failed with err=%s", string(out))
		return fiber.NewError(http.StatusInternalServerError, "failed to initialize: "+err.Error())
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
		return fiber.NewError(http.StatusInternalServerError, "failed to initialize: "+err.Error())
	}

	//go func() {
	//	if _, err := http.Get("https://pprotein.tokyoscience.jp/api/group/collect"); err != nil {
	//		log.Printf("failed to communicate with pprotein: %v", err)
	//	}
	//}()

	c.Request().Header.Add("Content-Type", "application/json;charset=utf-8")
	return c.Status(http.StatusOK).JSON(InitializeResponse{
		Language: "golang",
	})
}
func initializeSlaveHandler(c *fiber.Ctx) error {
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

	if err := resetTagCache(c.Context()); err != nil {
		return fiber.NewError(http.StatusInternalServerError, "failed to reset tag cache: "+err.Error())
	}

	return c.SendStatus(http.StatusNoContent)
}

//type JSONSerializer interface {
//	Serialize(c Context, i interface{}, indent string) error
//	Deserialize(c Context, i interface{}) error
//}

func main() {
	go startDNS()

	e := fiber.New(fiber.Config{
		JSONEncoder: func(v interface{}) ([]byte, error) {
			return json.Marshal(v)
		},
		JSONDecoder: func(data []byte, v interface{}) error {
			return json.Unmarshal(data, v)
		},
		ErrorHandler: errorResponseHandler,
	})

	//e.Debug = true
	//e.Logger.SetLevel(echolog.OFF)
	//e.Use(middleware.Logger())
	// e.Use(middleware.Recover())

	// 初期化
	e.Post("/api/initialize", initializeHandler)
	e.Post("/api/initialize/slave", initializeSlaveHandler)

	// top
	e.Get("/api/tag", getTagHandler)
	e.Get("/api/user/:username/theme", getStreamerThemeHandler)

	// livestream
	// reserve livestream
	e.Post("/api/livestream/reservation", reserveLivestreamHandler)
	// list livestream
	e.Get("/api/livestream/search", searchLivestreamsHandler)
	e.Get("/api/livestream", getMyLivestreamsHandler)
	e.Get("/api/user/:username/livestream", getUserLivestreamsHandler)
	// get livestream
	e.Get("/api/livestream/:livestream_id", getLivestreamHandler)
	// get polling livecomment timeline
	e.Get("/api/livestream/:livestream_id/livecomment", getLivecommentsHandler)
	// ライブコメント投稿
	e.Post("/api/livestream/:livestream_id/livecomment", postLivecommentHandler)
	e.Post("/api/livestream/:livestream_id/reaction", postReactionHandler)
	e.Get("/api/livestream/:livestream_id/reaction", getReactionsHandler)

	// (配信者向け)ライブコメントの報告一覧取得API
	e.Get("/api/livestream/:livestream_id/report", getLivecommentReportsHandler)
	e.Get("/api/livestream/:livestream_id/ngwords", getNgwords)
	// ライブコメント報告
	e.Post("/api/livestream/:livestream_id/livecomment/:livecomment_id/report", reportLivecommentHandler)
	// 配信者によるモデレーション (NGワード登録)
	e.Post("/api/livestream/:livestream_id/moderate", moderateHandler)

	// livestream_viewersにINSERTするため必要
	// ユーザ視聴開始 (viewer)
	e.Post("/api/livestream/:livestream_id/enter", enterLivestreamHandler)
	// ユーザ視聴終了 (viewer)
	e.Delete("/api/livestream/:livestream_id/exit", exitLivestreamHandler)

	// user
	e.Post("/api/register", registerHandler)
	e.Post("/api/login", loginHandler)
	e.Get("/api/user/me", getMeHandler)
	// フロントエンドで、配信予約のコラボレーターを指定する際に必要
	e.Get("/api/user/:username", getUserHandler)
	e.Get("/api/user/:username/statistics", getUserStatisticsHandler)
	e.Get("/api/user/:username/icon", getIconHandler)
	e.Post("/api/icon", postIconHandler)
	e.Post("/api/internal/icon", postInternalIconHandler)

	// stats
	// ライブ配信統計情報
	e.Get("/api/livestream/:livestream_id/statistics", getLivestreamStatisticsHandler)

	// 課金情報
	e.Get("/api/payment", GetPaymentResult)

	// DB接続
	conn, err := connectDB()
	if err != nil {
		//e.Logger.Errorf("failed to connect db: %v", err)
		os.Exit(1)
	}
	defer conn.Close()
	dbConn = conn

	// キャッシュの初期化
	if err := resetTagCache(context.Background()); err != nil {
		//e.Logger.Errorf("failed to reset tag cache: %v", err)
		os.Exit(1)
	}

	subdomainAddr, ok := os.LookupEnv(powerDNSSubdomainAddressEnvKey)
	if !ok {
		//e.Logger.Errorf("environ %s must be provided", powerDNSSubdomainAddressEnvKey)
		os.Exit(1)
	}
	powerDNSSubdomainAddress = subdomainAddr

	// HTTPサーバ起動
	listenAddr := net.JoinHostPort("", strconv.Itoa(listenPort))
	if err := e.Listen(listenAddr); err != nil {
		//e.Logger.Errorf("failed to start HTTP server: %v", err)
		os.Exit(1)
	}
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func errorResponseHandler(c *fiber.Ctx, err error) error {
	//c.Logger().Errorf("error at %s: %+v", c.Path(), err)
	log.Printf("error at %s: %+v", c.Path(), err)
	if he, ok := err.(*fiber.Error); ok {
		return c.Status(he.Code).JSON(&ErrorResponse{Error: err.Error()})
	}

	return c.Status(http.StatusInternalServerError).JSON(&ErrorResponse{Error: err.Error()})
}
