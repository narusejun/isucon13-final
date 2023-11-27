package main

import (
	"database/sql"
	"errors"
	"github.com/gofiber/fiber/v2"
	"net/http"
	"slices"
)

type Tag struct {
	ID   int64  `json:"id" db:"id"`
	Name string `json:"name" db:"name"`
}

type TagsResponse struct {
	Tags []*Tag `json:"tags"`
}

func getTagHandler(c *fiber.Ctx) error {
	tags := make([]*Tag, 0, tagIDCache.Count())
	for _, tag := range tagIDCache.Items() {
		tags = append(tags, &Tag{
			ID:   tag.ID,
			Name: tag.Name,
		})
	}

	slices.SortFunc(tags, func(a, b *Tag) int {
		return int(a.ID - b.ID)
	})

	return c.Status(http.StatusOK).JSON(&TagsResponse{
		Tags: tags,
	})
}

// 配信者のテーマ取得API
// GET /api/user/:username/theme
func getStreamerThemeHandler(c *fiber.Ctx) error {
	ctx := c.Context()

	if err := verifyUserSession(c); err != nil {
		// fiber.NewErrorが返っているのでそのまま出力
		//c.Logger().Printf("verifyUserSession: %+v\n", err)
		return err
	}

	username := c.Params("username")

	tx, err := dbConn.BeginTxx(ctx, nil)
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, "failed to begin transaction: "+err.Error())
	}
	defer tx.Rollback()

	userModel := UserModel{}
	err = tx.GetContext(ctx, &userModel, "SELECT id FROM users WHERE name = ?", username)
	if errors.Is(err, sql.ErrNoRows) {
		return fiber.NewError(http.StatusNotFound, "not found user that has the given username")
	}
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, "failed to get user: "+err.Error())
	}

	themeModel := ThemeModel{}
	if err := tx.GetContext(ctx, &themeModel, "SELECT * FROM themes WHERE user_id = ?", userModel.ID); err != nil {
		return fiber.NewError(http.StatusInternalServerError, "failed to get user theme: "+err.Error())
	}

	if err := tx.Commit(); err != nil {
		return fiber.NewError(http.StatusInternalServerError, "failed to commit: "+err.Error())
	}

	theme := Theme{
		ID:       themeModel.ID,
		DarkMode: themeModel.DarkMode,
	}

	return c.Status(http.StatusOK).JSON(theme)
}
