package main

import (
	"github.com/gofiber/fiber/v2"
	"net/http"
)

type PaymentResult struct {
	TotalTip int64 `json:"total_tip"`
}

func GetPaymentResult(c *fiber.Ctx) error {
	ctx := c.Context()

	tx, err := dbConn.BeginTxx(ctx, nil)
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, "failed to begin transaction: "+err.Error())
	}
	defer tx.Rollback()

	var totalTip int64
	if err := tx.GetContext(ctx, &totalTip, "SELECT ifnull(SUM(tip), 0) FROM livecomments"); err != nil {
		return fiber.NewError(http.StatusInternalServerError, "failed to count total tip: "+err.Error())
	}

	if err := tx.Commit(); err != nil {
		return fiber.NewError(http.StatusInternalServerError, "failed to commit: "+err.Error())
	}

	return c.Status(http.StatusOK).JSON(&PaymentResult{
		TotalTip: totalTip,
	})
}
