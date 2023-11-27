package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
)

type PostLivecommentRequest struct {
	Comment string `json:"comment"`
	Tip     int64  `json:"tip"`
}

type LivecommentModel struct {
	ID           int64  `db:"id"`
	UserID       int64  `db:"user_id"`
	LivestreamID int64  `db:"livestream_id"`
	Comment      string `db:"comment"`
	Tip          int64  `db:"tip"`
	CreatedAt    int64  `db:"created_at"`
}

type Livecomment struct {
	ID         int64       `json:"id"`
	User       User        `json:"user"`
	Livestream *Livestream `json:"livestream"`
	Comment    string      `json:"comment"`
	Tip        int64       `json:"tip"`
	CreatedAt  int64       `json:"created_at"`
}

type LivecommentReport struct {
	ID          int64       `json:"id"`
	Reporter    User        `json:"reporter"`
	Livecomment Livecomment `json:"livecomment"`
	CreatedAt   int64       `json:"created_at"`
}

type LivecommentReportModel struct {
	ID            int64 `db:"id"`
	UserID        int64 `db:"user_id"`
	LivestreamID  int64 `db:"livestream_id"`
	LivecommentID int64 `db:"livecomment_id"`
	CreatedAt     int64 `db:"created_at"`
}

type ModerateRequest struct {
	NGWord string `json:"ng_word"`
}

type NGWord struct {
	ID           int64  `json:"id" db:"id"`
	UserID       int64  `json:"user_id" db:"user_id"`
	LivestreamID int64  `json:"livestream_id" db:"livestream_id"`
	Word         string `json:"word" db:"word"`
	CreatedAt    int64  `json:"created_at" db:"created_at"`
}

func getLivecommentsHandler(c *fiber.Ctx) error {
	ctx := c.Context()

	if err := verifyUserSession(c); err != nil {
		// fiber.NewErrorが返っているのでそのまま出力
		return err
	}

	livestreamID, err := strconv.Atoi(c.Params("livestream_id"))
	if err != nil {
		return fiber.NewError(http.StatusBadRequest, "livestream_id in path must be integer")
	}

	tx, err := dbConn.BeginTxx(ctx, nil)
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, "failed to begin transaction: "+err.Error())
	}
	defer tx.Rollback()

	query := "SELECT * FROM livecomments WHERE livestream_id = ? ORDER BY created_at DESC"
	if c.Query("limit") != "" {
		limit, err := strconv.Atoi(c.Query("limit"))
		if err != nil {
			return fiber.NewError(http.StatusBadRequest, "limit query parameter must be integer")
		}
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	livecommentModels := []LivecommentModel{}
	err = tx.SelectContext(ctx, &livecommentModels, query, livestreamID)
	if errors.Is(err, sql.ErrNoRows) {
		return c.Status(http.StatusOK).JSON([]*Livecomment{})
	}
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, "failed to get livecomments: "+err.Error())
	}

	livestreamModel, err := getLivestream(ctx, tx, livestreamID)
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, "failed to fil livecomments: "+err.Error())
	}
	livestream, err := fillLivestreamResponse(ctx, tx, livestreamModel)
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, "failed to fil livecomments: "+err.Error())
	}

	livecomments := make([]Livecomment, len(livecommentModels))
	for i := range livecommentModels {
		livecomment, err := fillLivecommentResponse(ctx, tx, livecommentModels[i], &livestream)
		if err != nil {
			return fiber.NewError(http.StatusInternalServerError, "failed to fil livecomments: "+err.Error())
		}

		livecomments[i] = livecomment
	}

	if err := tx.Commit(); err != nil {
		return fiber.NewError(http.StatusInternalServerError, "failed to commit: "+err.Error())
	}

	return c.Status(http.StatusOK).JSON(livecomments)
}

func getNgwords(c *fiber.Ctx) error {
	ctx := c.Context()

	if err := verifyUserSession(c); err != nil {
		return err
	}

	// error already checked
	sess, _ := getSession(c)
	// existence already checked
	userID := sess.Values[defaultUserIDKey].(int64)

	livestreamID, err := strconv.Atoi(c.Params("livestream_id"))
	if err != nil {
		return fiber.NewError(http.StatusBadRequest, "livestream_id in path must be integer")
	}

	tx, err := dbConn.BeginTxx(ctx, nil)
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, "failed to begin transaction: "+err.Error())
	}
	defer tx.Rollback()

	var ngWords []*NGWord
	if err := tx.SelectContext(ctx, &ngWords, "SELECT * FROM ng_words WHERE user_id = ? AND livestream_id = ? ORDER BY created_at DESC", userID, livestreamID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.Status(http.StatusOK).JSON([]*NGWord{})
		} else {
			return fiber.NewError(http.StatusInternalServerError, "failed to get NG words: "+err.Error())
		}
	}

	if err := tx.Commit(); err != nil {
		return fiber.NewError(http.StatusInternalServerError, "failed to commit: "+err.Error())
	}

	return c.Status(http.StatusOK).JSON(ngWords)
}

func postLivecommentHandler(c *fiber.Ctx) error {
	ctx := c.Context()

	if err := verifyUserSession(c); err != nil {
		return err
	}

	livestreamID, err := strconv.Atoi(c.Params("livestream_id"))
	if err != nil {
		return fiber.NewError(http.StatusBadRequest, "livestream_id in path must be integer")
	}

	// error already checked
	sess, _ := getSession(c)
	// existence already checked
	userID := sess.Values[defaultUserIDKey].(int64)

	var req *PostLivecommentRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(http.StatusBadRequest, "failed to decode the request body as json")
	}

	tx, err := dbConn.BeginTxx(ctx, nil)
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, "failed to begin transaction: "+err.Error())
	}
	defer tx.Rollback()

	livestreamModel, err := getLivestream(ctx, tx, livestreamID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fiber.NewError(http.StatusNotFound, "livestream not found")
		} else {
			return fiber.NewError(http.StatusInternalServerError, "failed to get livestream: "+err.Error())
		}
	}

	// スパム判定
	var ngwords []*NGWord
	if cached, ok := ngwordsCache.Load(livestreamID); ok {
		ngwords = cached.([]*NGWord)
	} else {
		if err := tx.SelectContext(ctx, &ngwords, "SELECT id, user_id, livestream_id, word FROM ng_words WHERE livestream_id = ?", livestreamModel.ID); err != nil && !errors.Is(err, sql.ErrNoRows) {
			return fiber.NewError(http.StatusInternalServerError, "failed to get NG words: "+err.Error())
		}
		ngwordsCache.Store(livestreamID, ngwords)
	}

	//var hitSpam int
	//for _, ngword := range ngwords {
	//	query := `
	//	SELECT COUNT(*)
	//	FROM
	//	(SELECT ? AS text) AS texts
	//	INNER JOIN
	//	(SELECT concat('%', ?, '%')	AS pattern) AS patterns
	//	ON texts.text LIKE patterns.pattern;
	//	`
	//	if err := tx.GetContext(ctx, &hitSpam, query, req.Comment, ngword.Word); err != nil {
	//		return fiber.NewError(http.StatusInternalServerError, "failed to get hitspam: "+err.Error())
	//	}
	//	c.Logger().Infof("[hitSpam=%d] comment = %s", hitSpam, req.Comment)
	//	if hitSpam >= 1 {
	//		return fiber.NewError(http.StatusBadRequest, "このコメントがスパム判定されました")
	//	}
	//}

	for _, ngword := range ngwords {
		if strings.Contains(req.Comment, ngword.Word) {
			return fiber.NewError(http.StatusBadRequest, "このコメントがスパム判定されました")
		}
	}

	now := time.Now().Unix()
	livecommentModel := LivecommentModel{
		UserID:       userID,
		LivestreamID: int64(livestreamID),
		Comment:      req.Comment,
		Tip:          req.Tip,
		CreatedAt:    now,
	}

	rs, err := tx.NamedExecContext(ctx, "INSERT INTO livecomments (user_id, livestream_id, comment, tip, created_at) VALUES (:user_id, :livestream_id, :comment, :tip, :created_at)", livecommentModel)
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, "failed to insert livecomment: "+err.Error())
	}

	livecommentID, err := rs.LastInsertId()
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, "failed to get last inserted livecomment id: "+err.Error())
	}
	livecommentModel.ID = livecommentID

	livestream, err := fillLivestreamResponse(ctx, tx, livestreamModel)
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, "failed to fil livecomments: "+err.Error())
	}

	livecomment, err := fillLivecommentResponse(ctx, tx, livecommentModel, &livestream)
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, "failed to fill livecomment: "+err.Error())
	}

	if err := tx.Commit(); err != nil {
		return fiber.NewError(http.StatusInternalServerError, "failed to commit: "+err.Error())
	}

	return c.Status(http.StatusCreated).JSON(livecomment)
}

func reportLivecommentHandler(c *fiber.Ctx) error {
	ctx := c.Context()

	if err := verifyUserSession(c); err != nil {
		return err
	}

	livestreamID, err := strconv.Atoi(c.Params("livestream_id"))
	if err != nil {
		return fiber.NewError(http.StatusBadRequest, "livestream_id in path must be integer")
	}

	livecommentID, err := strconv.Atoi(c.Params("livecomment_id"))
	if err != nil {
		return fiber.NewError(http.StatusBadRequest, "livecomment_id in path must be integer")
	}

	// error already checked
	sess, _ := getSession(c)
	// existence already checked
	userID := sess.Values[defaultUserIDKey].(int64)

	tx, err := dbConn.BeginTxx(ctx, nil)
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, "failed to begin transaction: "+err.Error())
	}
	defer tx.Rollback()

	if _, err = getLivestream(ctx, tx, livestreamID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fiber.NewError(http.StatusNotFound, "livestream not found")
		} else {
			return fiber.NewError(http.StatusInternalServerError, "failed to get livestream: "+err.Error())
		}
	}

	var livecommentModel LivecommentModel
	if err := tx.GetContext(ctx, &livecommentModel, "SELECT * FROM livecomments WHERE id = ?", livecommentID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fiber.NewError(http.StatusNotFound, "livecomment not found")
		} else {
			return fiber.NewError(http.StatusInternalServerError, "failed to get livecomment: "+err.Error())
		}
	}

	now := time.Now().Unix()
	reportModel := LivecommentReportModel{
		UserID:        int64(userID),
		LivestreamID:  int64(livestreamID),
		LivecommentID: int64(livecommentID),
		CreatedAt:     now,
	}
	rs, err := tx.NamedExecContext(ctx, "INSERT INTO livecomment_reports(user_id, livestream_id, livecomment_id, created_at) VALUES (:user_id, :livestream_id, :livecomment_id, :created_at)", &reportModel)
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, "failed to insert livecomment report: "+err.Error())
	}
	reportID, err := rs.LastInsertId()
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, "failed to get last inserted livecomment report id: "+err.Error())
	}
	reportModel.ID = reportID

	report, err := fillLivecommentReportResponse(ctx, tx, reportModel)
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, "failed to fill livecomment report: "+err.Error())
	}
	if err := tx.Commit(); err != nil {
		return fiber.NewError(http.StatusInternalServerError, "failed to commit: "+err.Error())
	}

	return c.Status(http.StatusCreated).JSON(report)
}

var (
	ngwordsCache = sync.Map{}
)

// NGワードを登録
func moderateHandler(c *fiber.Ctx) error {
	ctx := c.Context()

	if err := verifyUserSession(c); err != nil {
		return err
	}

	livestreamID, err := strconv.Atoi(c.Params("livestream_id"))
	if err != nil {
		return fiber.NewError(http.StatusBadRequest, "livestream_id in path must be integer")
	}

	// error already checked
	sess, _ := getSession(c)
	// existence already checked
	userID := sess.Values[defaultUserIDKey].(int64)

	var req *ModerateRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(http.StatusBadRequest, "failed to decode the request body as json")
	}

	tx, err := dbConn.BeginTxx(ctx, nil)
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, "failed to begin transaction: "+err.Error())
	}
	defer tx.Rollback()

	// 配信者自身の配信に対するmoderateなのかを検証
	var ownedLivestreams []LivestreamModel
	if err := tx.SelectContext(ctx, &ownedLivestreams, "SELECT * FROM livestreams WHERE id = ? AND user_id = ?", livestreamID, userID); err != nil {
		return fiber.NewError(http.StatusInternalServerError, "failed to get livestreams: "+err.Error())
	}
	if len(ownedLivestreams) == 0 {
		return fiber.NewError(http.StatusBadRequest, "A streamer can't moderate livestreams that other streamers own")
	}

	rs, err := tx.NamedExecContext(ctx, "INSERT INTO ng_words(user_id, livestream_id, word, created_at) VALUES (:user_id, :livestream_id, :word, :created_at)", &NGWord{
		UserID:       int64(userID),
		LivestreamID: int64(livestreamID),
		Word:         req.NGWord,
		CreatedAt:    time.Now().Unix(),
	})
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, "failed to insert new NG word: "+err.Error())
	}

	wordID, err := rs.LastInsertId()
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, "failed to get last inserted NG word id: "+err.Error())
	}

	// 新規に追加したNGワードにヒットする過去の投稿も全削除する
	// ライブコメント一覧取得
	query := `
			DELETE FROM livecomments
			WHERE
			livestream_id = ? AND
			comment LIKE ?
			`
	if _, err := tx.ExecContext(ctx, query, livestreamID, "%"+req.NGWord+"%"); err != nil {
		return fiber.NewError(http.StatusInternalServerError, "failed to delete old livecomments that hit spams: "+err.Error())
	}

	if err := tx.Commit(); err != nil {
		return fiber.NewError(http.StatusInternalServerError, "failed to commit: "+err.Error())
	}
	time.Sleep(1000 * time.Millisecond)
	ngwordsCache.Delete(livestreamID)

	return c.Status(http.StatusCreated).JSON(map[string]interface{}{
		"word_id": wordID,
	})
}

func fillLivecommentResponse(ctx context.Context, tx *sqlx.Tx, livecommentModel LivecommentModel, livestream *Livestream) (Livecomment, error) {
	commentOwnerModel, err := getUser(ctx, tx, livecommentModel.UserID)
	if err != nil {
		return Livecomment{}, err
	}
	commentOwner, err := fillUserResponse(ctx, tx, commentOwnerModel)
	if err != nil {
		return Livecomment{}, err
	}

	livecomment := Livecomment{
		ID:         livecommentModel.ID,
		User:       commentOwner,
		Livestream: livestream,
		Comment:    livecommentModel.Comment,
		Tip:        livecommentModel.Tip,
		CreatedAt:  livecommentModel.CreatedAt,
	}

	return livecomment, nil
}

func fillLivecommentReportResponse(ctx context.Context, tx *sqlx.Tx, reportModel LivecommentReportModel) (LivecommentReport, error) {
	reporterModel, err := getUser(ctx, tx, reportModel.UserID)
	if err != nil {
		return LivecommentReport{}, err
	}
	reporter, err := fillUserResponse(ctx, tx, reporterModel)
	if err != nil {
		return LivecommentReport{}, err
	}

	livecommentModel := LivecommentModel{}
	if err := tx.GetContext(ctx, &livecommentModel, "SELECT * FROM livecomments WHERE id = ?", reportModel.LivecommentID); err != nil {
		return LivecommentReport{}, err
	}

	livestreamModel, err := getLivestream(ctx, tx, int(livecommentModel.LivestreamID))
	if err != nil {
		return LivecommentReport{}, err
	}
	livestream, err := fillLivestreamResponse(ctx, tx, livestreamModel)
	if err != nil {
		return LivecommentReport{}, err
	}

	livecomment, err := fillLivecommentResponse(ctx, tx, livecommentModel, &livestream)
	if err != nil {
		return LivecommentReport{}, err
	}

	report := LivecommentReport{
		ID:          reportModel.ID,
		Reporter:    reporter,
		Livecomment: livecomment,
		CreatedAt:   reportModel.CreatedAt,
	}
	return report, nil
}
