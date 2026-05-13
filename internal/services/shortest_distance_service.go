package services

import (
	"database/sql"
	"delivery-system/internal/models"
	"fmt"
	"math"

	"github.com/google/uuid"
)

func (s *CourierService) CourierAssignmentService(id uuid.UUID) (models.Courier, error) {
	const op = "services.shortest_distance_service.CourierAssignmentService"

	type preferredCourier struct {
		courier models.Courier
		score   float64
	}

	var prefCouriers []preferredCourier

	order, err := s.GetOrderForCourier(id)
	if err != nil {
		var prefC preferredCourier
		return prefC.courier, fmt.Errorf("no available order path: %s | err: %v", op, err)
	}

	orderLat, orderLon := order.CurrentLat, order.CurrentLon

	couriers, err := s.AvailableCourier()
	if err != nil {
		var prefC preferredCourier
		return prefC.courier, fmt.Errorf("no available courier path: %s | err: %v", op, err)
	}

	for _, courier := range couriers {
		distance, err := s.DistanceCounter(courier, orderLat, orderLon)
		if err != nil {
			var prefC preferredCourier
			return prefC.courier, fmt.Errorf("cannot count distance %s | err: %v", op, err)
		}

		rating, err := s.RatingCourier(courier)
		if err != nil {
			var prefC preferredCourier
			return prefC.courier, fmt.Errorf("cannot count rating %s | err: %v", op, err)
		}

		distanceScore := distance * 0.75
		ratingScore := rating * 0.25

		totalScore := distanceScore + ratingScore

		prefCouriers = append(prefCouriers, preferredCourier{
			courier: courier,
			score:   totalScore,
		})
	}
	var bestCourier models.Courier
	var bestScore float64

	for _, prefCourier := range prefCouriers {
		if prefCourier.score > bestScore {
			bestScore = prefCourier.score
			bestCourier = prefCourier.courier
		}
	}

	return bestCourier, nil
}

func (s *CourierService) RatingCourier(courier models.Courier) (float64, error) {
	return courier.Rating, nil
}

func (s *CourierService) DistanceCounter(courier models.Courier, currOrderLat float64, currOrderLon float64) (float64, error) {
	const op = "services.shortest_distance_service.DistanceCounter"

	const R = 6371.0 // Радиус Земли в километрах

	// Координаты курьера из структуры
	lat1 := courier.CurrentLat
	lon1 := courier.CurrentLon

	// Координаты заказа
	lat2 := currOrderLat
	lon2 := currOrderLon

	// Переводим градусы в радианы
	dLat := (lat2 - lat1) * (math.Pi / 180.0)
	dLon := (lon2 - lon1) * (math.Pi / 180.0)

	lat1Rad := lat1 * (math.Pi / 180.0)
	lat2Rad := lat2 * (math.Pi / 180.0)

	// Формула гаверсинусов
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Sin(dLon/2)*math.Sin(dLon/2)*math.Cos(lat1Rad)*math.Cos(lat2Rad)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	distance := R * c

	// Если расстояние получилось аномальным (например, из-за битых координат в БД)
	if math.IsNaN(distance) {
		return 0, fmt.Errorf("%s: calculated distance is NaN", op)
	}

	return distance, nil
}

func (s *CourierService) AvailableCourier() ([]models.Courier, error) {
	const op = "services.shortest_distance_service.AvailableCourier"

	var couriers []models.Courier

	query := `
SELECT 
    id, 
    name, 
    phone, 
    status, 
    current_lat, 
    current_lon, 
    created_at, 
    updated_at, 
    last_seen_at, 
    rating, 
    total_reviews
FROM couriers
WHERE status = 'available'
`
	rows, err := s.db.Query(query)
	if err != nil {
		if err == sql.ErrNoRows {
			s.log.Errorf("No courier with status available path: %s | err: %v", op, err)
			return nil, err
		}
		s.log.Errorf("Error querying courier path: %s | err: %v", op, err)
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var courier models.Courier
		if err := rows.Scan(
			&courier.ID,
			&courier.Name,
			&courier.Phone,
			&courier.Status,
			&courier.CurrentLat,
			&courier.CurrentLon,
			&courier.CreatedAt,
			&courier.UpdatedAt,
			&courier.LastSeenAt,
			&courier.Rating,
			&courier.TotalReviews,
		); err != nil {
			s.log.Errorf("Error scanning courier path:%s | err: %v", op, err)
			return couriers, err
		}
		couriers = append(couriers, courier)
	}
	if err := rows.Err(); err != nil {
		s.log.Errorf("Error scanning courier path:%s | err: %v", op, err)
		return nil, err
	}
	return couriers, nil
}
func (s *CourierService) GetOrderForCourier(orderID uuid.UUID) (*models.Order, error) {
	order := &models.Order{}

	query := `
		SELECT id, customer_name, customer_phone, delivery_address, total_amount, 
		       status, courier_id, created_at, updated_at, delivered_at, current_lat, current_lon
		FROM orders 
		WHERE id = $1
	`

	err := s.db.QueryRow(query, orderID).Scan(
		&order.ID,
		&order.CustomerName,
		&order.CustomerPhone,
		&order.DeliveryAddress,
		&order.TotalAmount,
		&order.Status,
		&order.CourierID,
		&order.CreatedAt,
		&order.UpdatedAt,
		&order.DeliveredAt,
		&order.CurrentLat,
		&order.CurrentLon,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("order not found")
		}
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	// Получение товаров заказа
	itemsQuery := `
		SELECT id, order_id, name, quantity, price
		FROM order_items
		WHERE order_id = $1
	`

	rows, err := s.db.Query(itemsQuery, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get order items: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var item models.OrderItem
		if err := rows.Scan(&item.ID, &item.OrderID, &item.Name, &item.Quantity, &item.Price); err != nil {
			return nil, fmt.Errorf("failed to scan order item: %w", err)
		}
		order.Items = append(order.Items, item)
	}

	return order, nil
}
