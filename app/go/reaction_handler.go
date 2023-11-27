package main

import (
	"context"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"net/http"
	"strconv"
	"time"

	"github.com/jmoiron/sqlx"
)

type ReactionModel struct {
	ID           int64  `db:"id"`
	EmojiName    string `db:"emoji_name"`
	UserID       int64  `db:"user_id"`
	LivestreamID int64  `db:"livestream_id"`
	CreatedAt    int64  `db:"created_at"`
}

type Reaction struct {
	ID         int64       `json:"id"`
	EmojiName  string      `json:"emoji_name"`
	User       User        `json:"user"`
	Livestream *Livestream `json:"livestream"`
	CreatedAt  int64       `json:"created_at"`
}

type PostReactionRequest struct {
	EmojiName string `json:"emoji_name"`
}

func getReactionsHandler(c *fiber.Ctx) error {
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

	livestreamModel, err := getLivestream(ctx, tx, livestreamID)
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, "failed to fill reaction: "+err.Error())
	}
	livestream, err := fillLivestreamResponse(ctx, tx, livestreamModel)
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, "failed to fill reaction: "+err.Error())
	}

	query := "SELECT * FROM reactions WHERE livestream_id = ? ORDER BY created_at DESC"
	if c.Query("limit") != "" {
		limit, err := strconv.Atoi(c.Query("limit"))
		if err != nil {
			return fiber.NewError(http.StatusBadRequest, "limit query parameter must be integer")
		}
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	reactionModels := []ReactionModel{}
	if err := tx.SelectContext(ctx, &reactionModels, query, livestreamID); err != nil {
		return fiber.NewError(http.StatusNotFound, "failed to get reactions")
	}

	reactions := make([]Reaction, len(reactionModels))
	for i := range reactionModels {
		reaction, err := fillReactionResponse(ctx, tx, reactionModels[i], &livestream)
		if err != nil {
			return fiber.NewError(http.StatusInternalServerError, "failed to fill reaction: "+err.Error())
		}

		reactions[i] = reaction
	}

	if err := tx.Commit(); err != nil {
		return fiber.NewError(http.StatusInternalServerError, "failed to commit: "+err.Error())
	}

	return c.Status(http.StatusOK).JSON(reactions)
}

func postReactionHandler(c *fiber.Ctx) error {
	ctx := c.Context()
	livestreamID, err := strconv.Atoi(c.Params("livestream_id"))
	if err != nil {
		return fiber.NewError(http.StatusBadRequest, "livestream_id in path must be integer")
	}

	if err := verifyUserSession(c); err != nil {
		// fiber.NewErrorが返っているのでそのまま出力
		return err
	}

	// error already checked
	sess, _ := getSession(c)
	// existence already checked
	userID := sess.Values[defaultUserIDKey].(int64)

	var req *PostReactionRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(http.StatusBadRequest, "failed to decode the request body as json")
	}

	tx, err := dbConn.BeginTxx(ctx, nil)
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, "failed to begin transaction: "+err.Error())
	}
	defer tx.Rollback()

	reactionModel := ReactionModel{
		UserID:       int64(userID),
		LivestreamID: int64(livestreamID),
		EmojiName:    req.EmojiName,
		CreatedAt:    time.Now().Unix(),
	}

	result, err := tx.NamedExecContext(ctx, "INSERT INTO reactions (user_id, livestream_id, emoji_name, created_at) VALUES (:user_id, :livestream_id, :emoji_name, :created_at)", reactionModel)
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, "failed to insert reaction: "+err.Error())
	}

	reactionID, err := result.LastInsertId()
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, "failed to get last inserted reaction id: "+err.Error())
	}
	reactionModel.ID = reactionID

	livestreamModel, err := getLivestream(ctx, tx, livestreamID)
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, "failed to fill reaction: "+err.Error())
	}
	livestream, err := fillLivestreamResponse(ctx, tx, livestreamModel)
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, "failed to fill reaction: "+err.Error())
	}

	reaction, err := fillReactionResponse(ctx, tx, reactionModel, &livestream)
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, "failed to fill reaction: "+err.Error())
	}

	if err := tx.Commit(); err != nil {
		return fiber.NewError(http.StatusInternalServerError, "failed to commit: "+err.Error())
	}

	return c.Status(http.StatusCreated).JSON(reaction)
}

func fillReactionResponse(ctx context.Context, tx *sqlx.Tx, reactionModel ReactionModel, livestream *Livestream) (Reaction, error) {
	userModel, err := getUser(ctx, tx, reactionModel.UserID)
	if err != nil {
		return Reaction{}, err
	}
	user, err := fillUserResponse(ctx, tx, userModel)
	if err != nil {
		return Reaction{}, err
	}

	reaction := Reaction{
		ID:         reactionModel.ID,
		EmojiName:  reactionModel.EmojiName,
		User:       user,
		Livestream: livestream,
		CreatedAt:  reactionModel.CreatedAt,
	}

	return reaction, nil
}
