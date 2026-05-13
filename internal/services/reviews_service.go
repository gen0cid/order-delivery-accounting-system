package services

import (
	"delivery-system/internal/models"
	"fmt"
	"time"

	"github.com/google/uuid"
)

func (s *CourierService) SetRating(feedbackRequest models.FeedbackRequest, courierID, orderID uuid.UUID) (*models.FeedbackResponse, error) {
	// начинаем транзакцию

	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// создаем ответ пользователю

	feedback := models.FeedbackResponse{
		Id:        uuid.New(),
		OrderId:   orderID,
		CourierId: courierID,
		Rating:    feedbackRequest.Rating,
		Comment:   feedbackRequest.Comment,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// вставляем значения в таблицу reviews

	query := `
		INSERT INTO reviews (id, order_id, courier_id, rating, comment, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
`
	_, err = tx.Exec(query, feedback.Id, orderID, courierID, feedback.Rating, feedback.Comment, feedback.CreatedAt, feedback.UpdatedAt)
	if err != nil {
		s.log.Error("failed to save set rating: %w", err)
		return nil, fmt.Errorf("failed to set rating: %w", err)
	}

	// 	для того чтобы обновить данные у курьера, надо из таблицы взять его рейтинг и количество отзывов
	var currentRating float64
	var currentNumberOfReviews int

	query = `
		SELECT rating, total_reviews
		FROM couriers
		WHERE id = $1
`

	// берем рейтинг и количество отзывов
	if err := tx.QueryRow(query, courierID).Scan(&currentRating, &currentNumberOfReviews); err != nil {
		return nil, fmt.Errorf("failed to get rating, or get NumberOfReviews: %w", err)
	}

	// рассчитываем новый рейтинг на основе того что получили выше
	newRating := ((currentRating * float64(currentNumberOfReviews)) + float64(feedbackRequest.Rating)) / (float64(currentNumberOfReviews) + 1)

	// вставляем значения в таблицу couriers
	query = `
		UPDATE couriers 
		SET rating = $1, 
			total_reviews = total_reviews + 1
`
	if _, err := tx.Exec(query, newRating); err != nil {
		return nil, fmt.Errorf("failed to set rating: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// возвращаем ответ
	return &feedback, nil
}
