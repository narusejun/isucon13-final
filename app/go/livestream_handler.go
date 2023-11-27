package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
)

type ReserveLivestreamRequest struct {
	Tags         []int64 `json:"tags"`
	Title        string  `json:"title"`
	Description  string  `json:"description"`
	PlaylistUrl  string  `json:"playlist_url"`
	ThumbnailUrl string  `json:"thumbnail_url"`
	StartAt      int64   `json:"start_at"`
	EndAt        int64   `json:"end_at"`
}

type LivestreamViewerModel struct {
	UserID       int64 `db:"user_id" json:"user_id"`
	LivestreamID int64 `db:"livestream_id" json:"livestream_id"`
	CreatedAt    int64 `db:"created_at" json:"created_at"`
}

type LivestreamModel struct {
	ID           int64  `db:"id" json:"id"`
	UserID       int64  `db:"user_id" json:"user_id"`
	Title        string `db:"title" json:"title"`
	Description  string `db:"description" json:"description"`
	PlaylistUrl  string `db:"playlist_url" json:"playlist_url"`
	ThumbnailUrl string `db:"thumbnail_url" json:"thumbnail_url"`
	StartAt      int64  `db:"start_at" json:"start_at"`
	EndAt        int64  `db:"end_at" json:"end_at"`
}

type Livestream struct {
	ID           int64  `json:"id"`
	Owner        User   `json:"owner"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	PlaylistUrl  string `json:"playlist_url"`
	ThumbnailUrl string `json:"thumbnail_url"`
	Tags         []Tag  `json:"tags"`
	StartAt      int64  `json:"start_at"`
	EndAt        int64  `json:"end_at"`
}

type LivestreamTagModel struct {
	ID           int64 `db:"id" json:"id"`
	LivestreamID int64 `db:"livestream_id" json:"livestream_id"`
	TagID        int64 `db:"tag_id" json:"tag_id"`
}

type ReservationSlotModel struct {
	ID      int64 `db:"id" json:"id"`
	Slot    int64 `db:"slot" json:"slot"`
	StartAt int64 `db:"start_at" json:"start_at"`
	EndAt   int64 `db:"end_at" json:"end_at"`
}

var (
	livestreamCache = sync.Map{}
)

func getLivestream(ctx context.Context, tx *sqlx.Tx, livestreamID int) (LivestreamModel, error) {
	if livestream, ok := livestreamCache.Load(livestreamID); ok {
		return livestream.(LivestreamModel), nil
	}

	livestream := LivestreamModel{}
	if err := tx.GetContext(ctx, &livestream, "SELECT * FROM livestreams WHERE id = ?", livestreamID); err != nil {
		return LivestreamModel{}, err
	}
	livestreamCache.Store(livestreamID, livestream)

	return livestream, nil
}

func reserveLivestreamHandler(c *fiber.Ctx) error {
	ctx := c.Context()

	if err := verifyUserSession(c); err != nil {
		// fiber.NewErrorが返っているのでそのまま出力
		return err
	}

	// error already checked
	sess, _ := getSession(c)
	// existence already checked
	userID := sess.Values[defaultUserIDKey].(int64)

	var req *ReserveLivestreamRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(http.StatusBadRequest, "failed to decode the request body as json")
	}

	tx, err := dbConn.BeginTxx(ctx, nil)
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, "failed to begin transaction: "+err.Error())
	}
	defer tx.Rollback()

	// 2023/11/25 10:00からの１年間の期間内であるかチェック
	var (
		termStartAt    = time.Date(2023, 11, 25, 1, 0, 0, 0, time.UTC)
		termEndAt      = time.Date(2024, 11, 25, 1, 0, 0, 0, time.UTC)
		reserveStartAt = time.Unix(req.StartAt, 0)
		reserveEndAt   = time.Unix(req.EndAt, 0)
	)
	if (reserveStartAt.Equal(termEndAt) || reserveStartAt.After(termEndAt)) || (reserveEndAt.Equal(termStartAt) || reserveEndAt.Before(termStartAt)) {
		return fiber.NewError(http.StatusBadRequest, "bad reservation time range")
	}

	// 予約枠をみて、予約が可能か調べる
	// NOTE: 並列な予約のoverbooking防止にFOR UPDATEが必要
	var slots []*ReservationSlotModel
	if err := tx.SelectContext(ctx, &slots, "SELECT * FROM reservation_slots WHERE start_at >= ? AND end_at <= ? FOR UPDATE", req.StartAt, req.EndAt); err != nil {
		//c.Logger().Warnf("予約枠一覧取得でエラー発生: %+v", err)
		return fiber.NewError(http.StatusInternalServerError, "failed to get reservation_slots: "+err.Error())
	}
	for _, slot := range slots {
		var count int
		if err := tx.GetContext(ctx, &count, "SELECT slot FROM reservation_slots WHERE start_at = ? AND end_at = ?", slot.StartAt, slot.EndAt); err != nil {
			return fiber.NewError(http.StatusInternalServerError, "failed to get reservation_slots: "+err.Error())
		}
		//c.Logger().Infof("%d ~ %d予約枠の残数 = %d\n", slot.StartAt, slot.EndAt, slot.Slot)
		if count < 1 {
			return fiber.NewError(http.StatusBadRequest, fmt.Sprintf("予約期間 %d ~ %dに対して、予約区間 %d ~ %dが予約できません", termStartAt.Unix(), termEndAt.Unix(), req.StartAt, req.EndAt))
		}
	}

	var (
		livestreamModel = &LivestreamModel{
			UserID:       int64(userID),
			Title:        req.Title,
			Description:  req.Description,
			PlaylistUrl:  req.PlaylistUrl,
			ThumbnailUrl: req.ThumbnailUrl,
			StartAt:      req.StartAt,
			EndAt:        req.EndAt,
		}
	)

	if _, err := tx.ExecContext(ctx, "UPDATE reservation_slots SET slot = slot - 1 WHERE start_at >= ? AND end_at <= ?", req.StartAt, req.EndAt); err != nil {
		return fiber.NewError(http.StatusInternalServerError, "failed to update reservation_slot: "+err.Error())
	}

	rs, err := tx.NamedExecContext(ctx, "INSERT INTO livestreams (user_id, title, description, playlist_url, thumbnail_url, start_at, end_at) VALUES(:user_id, :title, :description, :playlist_url, :thumbnail_url, :start_at, :end_at)", livestreamModel)
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, "failed to insert livestream: "+err.Error())
	}

	livestreamID, err := rs.LastInsertId()
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, "failed to get last inserted livestream id: "+err.Error())
	}
	livestreamModel.ID = livestreamID

	// タグ追加
	for _, tagID := range req.Tags {
		if _, err := tx.NamedExecContext(ctx, "INSERT INTO livestream_tags (livestream_id, tag_id) VALUES (:livestream_id, :tag_id)", &LivestreamTagModel{
			LivestreamID: livestreamID,
			TagID:        tagID,
		}); err != nil {
			return fiber.NewError(http.StatusInternalServerError, "failed to insert livestream tag: "+err.Error())
		}
		livestreamTagsCache.Delete(livestreamID)
	}

	livestream, err := fillLivestreamResponse(ctx, tx, *livestreamModel)
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, "failed to fill livestream: "+err.Error())
	}

	if err := tx.Commit(); err != nil {
		return fiber.NewError(http.StatusInternalServerError, "failed to commit: "+err.Error())
	}

	return c.Status(http.StatusCreated).JSON(livestream)
}

func searchLivestreamsHandler(c *fiber.Ctx) error {
	ctx := c.Context()
	keyTagName := c.Query("tag")

	tx, err := dbConn.BeginTxx(ctx, nil)
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, "failed to begin transaction: "+err.Error())
	}
	defer tx.Rollback()

	var livestreamModels []*LivestreamModel
	if c.Query("tag") != "" {
		// タグによる取得
		//var tagIDList []int
		//if err := tx.SelectContext(ctx, &tagIDList, "SELECT id FROM tags WHERE name = ?", keyTagName); err != nil {
		//	return fiber.NewError(http.StatusInternalServerError, "failed to get tags: "+err.Error())
		//}
		tag, err := getTagByName(keyTagName)
		if err != nil {
			return fiber.NewError(http.StatusInternalServerError, "failed to get tag: "+err.Error())
		}

		// クエリを実行して結果を取得
		if err := tx.SelectContext(ctx, &livestreamModels, `
			SELECT ls.* FROM livestreams ls
			INNER JOIN livestream_tags lt ON ls.id = lt.livestream_id AND lt.tag_id = ?
			`, tag.ID); err != nil {
			return fiber.NewError(http.StatusInternalServerError, "failed to get livestreams: "+err.Error())
		}
	} else {
		// 検索条件なし
		query := `SELECT * FROM livestreams ORDER BY id DESC`
		if c.Query("limit") != "" {
			limit, err := strconv.Atoi(c.Query("limit"))
			if err != nil {
				return fiber.NewError(http.StatusBadRequest, "limit query parameter must be integer")
			}
			query += fmt.Sprintf(" LIMIT %d", limit)
		}

		if err := tx.SelectContext(ctx, &livestreamModels, query); err != nil {
			return fiber.NewError(http.StatusInternalServerError, "failed to get livestreams: "+err.Error())
		}
	}

	livestreams := make([]Livestream, len(livestreamModels))
	for i := range livestreamModels {
		livestream, err := fillLivestreamResponse(ctx, tx, *livestreamModels[i])
		if err != nil {
			return fiber.NewError(http.StatusInternalServerError, "failed to fill livestream: "+err.Error())
		}
		livestreams[i] = livestream
	}

	if err := tx.Commit(); err != nil {
		return fiber.NewError(http.StatusInternalServerError, "failed to commit: "+err.Error())
	}

	return c.Status(http.StatusOK).JSON(livestreams)
}

func getMyLivestreamsHandler(c *fiber.Ctx) error {
	ctx := c.Context()
	if err := verifyUserSession(c); err != nil {
		return err
	}

	tx, err := dbConn.BeginTxx(ctx, nil)
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, "failed to begin transaction: "+err.Error())
	}
	defer tx.Rollback()

	// error already checked
	sess, _ := getSession(c)
	// existence already checked
	userID := sess.Values[defaultUserIDKey].(int64)

	var livestreamModels []*LivestreamModel
	if err := tx.SelectContext(ctx, &livestreamModels, "SELECT * FROM livestreams WHERE user_id = ?", userID); err != nil {
		return fiber.NewError(http.StatusInternalServerError, "failed to get livestreams: "+err.Error())
	}
	livestreams := make([]Livestream, len(livestreamModels))
	for i := range livestreamModels {
		livestream, err := fillLivestreamResponse(ctx, tx, *livestreamModels[i])
		if err != nil {
			return fiber.NewError(http.StatusInternalServerError, "failed to fill livestream: "+err.Error())
		}
		livestreams[i] = livestream
	}

	if err := tx.Commit(); err != nil {
		return fiber.NewError(http.StatusInternalServerError, "failed to commit: "+err.Error())
	}

	return c.Status(http.StatusOK).JSON(livestreams)
}

func getUserLivestreamsHandler(c *fiber.Ctx) error {
	ctx := c.Context()
	if err := verifyUserSession(c); err != nil {
		return err
	}

	username := c.Params("username")

	tx, err := dbConn.BeginTxx(ctx, nil)
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, "failed to begin transaction: "+err.Error())
	}
	defer tx.Rollback()

	var user UserModel
	if err := tx.GetContext(ctx, &user, "SELECT * FROM users WHERE name = ?", username); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fiber.NewError(http.StatusNotFound, "user not found")
		} else {
			return fiber.NewError(http.StatusInternalServerError, "failed to get user: "+err.Error())
		}
	}

	var livestreamModels []*LivestreamModel
	if err := tx.SelectContext(ctx, &livestreamModels, "SELECT * FROM livestreams WHERE user_id = ?", user.ID); err != nil {
		return fiber.NewError(http.StatusInternalServerError, "failed to get livestreams: "+err.Error())
	}
	livestreams := make([]Livestream, len(livestreamModels))
	for i := range livestreamModels {
		livestream, err := fillLivestreamResponse(ctx, tx, *livestreamModels[i])
		if err != nil {
			return fiber.NewError(http.StatusInternalServerError, "failed to fill livestream: "+err.Error())
		}
		livestreams[i] = livestream
	}

	if err := tx.Commit(); err != nil {
		return fiber.NewError(http.StatusInternalServerError, "failed to commit: "+err.Error())
	}

	return c.Status(http.StatusOK).JSON(livestreams)
}

// viewerテーブルの廃止
func enterLivestreamHandler(c *fiber.Ctx) error {
	ctx := c.Context()
	if err := verifyUserSession(c); err != nil {
		// fiber.NewErrorが返っているのでそのまま出力
		return err
	}

	// error already checked
	sess, _ := getSession(c)
	// existence already checked
	userID := sess.Values[defaultUserIDKey].(int64)

	livestreamID, err := strconv.Atoi(c.Params("livestream_id"))
	if err != nil {
		return fiber.NewError(http.StatusBadRequest, "livestream_id must be integer")
	}

	tx, err := dbConn.BeginTxx(ctx, nil)
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, "failed to begin transaction: "+err.Error())
	}
	defer tx.Rollback()

	viewer := LivestreamViewerModel{
		UserID:       int64(userID),
		LivestreamID: int64(livestreamID),
		CreatedAt:    time.Now().Unix(),
	}

	if _, err := tx.NamedExecContext(ctx, "INSERT INTO livestream_viewers_history (user_id, livestream_id, created_at) VALUES(:user_id, :livestream_id, :created_at)", viewer); err != nil {
		return fiber.NewError(http.StatusInternalServerError, "failed to insert livestream_view_history: "+err.Error())
	}

	if err := tx.Commit(); err != nil {
		return fiber.NewError(http.StatusInternalServerError, "failed to commit: "+err.Error())
	}

	return c.SendStatus(http.StatusOK)
}

func exitLivestreamHandler(c *fiber.Ctx) error {
	ctx := c.Context()
	if err := verifyUserSession(c); err != nil {
		// fiber.NewErrorが返っているのでそのまま出力
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

	if _, err := tx.ExecContext(ctx, "DELETE FROM livestream_viewers_history WHERE user_id = ? AND livestream_id = ?", userID, livestreamID); err != nil {
		return fiber.NewError(http.StatusInternalServerError, "failed to delete livestream_view_history: "+err.Error())
	}

	if err := tx.Commit(); err != nil {
		return fiber.NewError(http.StatusInternalServerError, "failed to commit: "+err.Error())
	}

	return c.SendStatus(http.StatusOK)
}

func getLivestreamHandler(c *fiber.Ctx) error {
	ctx := c.Context()

	if err := verifyUserSession(c); err != nil {
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

	livestreamModel, err := getLivestream(ctx, tx, livestreamID)
	if errors.Is(err, sql.ErrNoRows) {
		return fiber.NewError(http.StatusNotFound, "not found livestream that has the given id")
	}
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, "failed to get livestream: "+err.Error())
	}

	livestream, err := fillLivestreamResponse(ctx, tx, livestreamModel)
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, "failed to fill livestream: "+err.Error())
	}

	if err := tx.Commit(); err != nil {
		return fiber.NewError(http.StatusInternalServerError, "failed to commit: "+err.Error())
	}

	return c.Status(http.StatusOK).JSON(livestream)
}

func getLivecommentReportsHandler(c *fiber.Ctx) error {
	ctx := c.Context()

	if err := verifyUserSession(c); err != nil {
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

	livestreamModel, err := getLivestream(ctx, tx, livestreamID)
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, "failed to get livestream: "+err.Error())
	}

	// error already check
	sess, _ := getSession(c)
	// existence already check
	userID := sess.Values[defaultUserIDKey].(int64)

	if livestreamModel.UserID != userID {
		return fiber.NewError(http.StatusForbidden, "can't get other streamer's livecomment reports")
	}

	var reportModels []*LivecommentReportModel
	if err := tx.SelectContext(ctx, &reportModels, "SELECT * FROM livecomment_reports WHERE livestream_id = ?", livestreamID); err != nil {
		return fiber.NewError(http.StatusInternalServerError, "failed to get livecomment reports: "+err.Error())
	}

	reports := make([]LivecommentReport, len(reportModels))
	for i := range reportModels {
		report, err := fillLivecommentReportResponse(ctx, tx, *reportModels[i])
		if err != nil {
			return fiber.NewError(http.StatusInternalServerError, "failed to fill livecomment report: "+err.Error())
		}
		reports[i] = report
	}

	if err := tx.Commit(); err != nil {
		return fiber.NewError(http.StatusInternalServerError, "failed to commit: "+err.Error())
	}

	return c.Status(http.StatusOK).JSON(reports)
}

var (
	livestreamTagsCache = sync.Map{}
)

func fillLivestreamResponse(ctx context.Context, tx *sqlx.Tx, livestreamModel LivestreamModel) (Livestream, error) {
	ownerModel, err := getUser(ctx, tx, livestreamModel.UserID)
	if err != nil {
		return Livestream{}, err
	}
	owner, err := fillUserResponse(ctx, tx, ownerModel)
	if err != nil {
		return Livestream{}, err
	}

	tags := []Tag{}
	if t, ok := livestreamTagsCache.Load(livestreamModel.ID); ok {
		tags = t.([]Tag)
	} else {
		if err := tx.SelectContext(ctx, &tags, "SELECT * FROM tags WHERE id IN (SELECT tag_id FROM livestream_tags WHERE livestream_id = ?)", livestreamModel.ID); err != nil {
			return Livestream{}, err
		}
		livestreamTagsCache.Store(livestreamModel.ID, tags)
	}

	livestream := Livestream{
		ID:           livestreamModel.ID,
		Owner:        owner,
		Title:        livestreamModel.Title,
		Tags:         tags,
		Description:  livestreamModel.Description,
		PlaylistUrl:  livestreamModel.PlaylistUrl,
		ThumbnailUrl: livestreamModel.ThumbnailUrl,
		StartAt:      livestreamModel.StartAt,
		EndAt:        livestreamModel.EndAt,
	}
	return livestream, nil
}
