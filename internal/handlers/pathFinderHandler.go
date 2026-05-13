package handlers

import (
	"delivery-system/internal/kafka"
	"delivery-system/internal/logger"
	"delivery-system/internal/models"
	"delivery-system/internal/redis"
	"delivery-system/internal/services"
	"encoding/json"
	"net/http"
)

type PathFinderHandler struct {
	PathFinder  *services.PathFinder
	producer    *kafka.Producer
	redisClient *redis.Client
	log         *logger.Logger
}

func NewPathFinderHandler(PathFinder *services.PathFinder, producer *kafka.Producer, redisClient *redis.Client, log *logger.Logger) *PathFinderHandler {
	return &PathFinderHandler{
		PathFinder:  PathFinder,
		producer:    producer,
		redisClient: redisClient,
		log:         log,
	}
}

func (p *PathFinderHandler) CalculateTheCostHandler(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.PathFinderHandler.CalculateTheCostHandler"

	if r.Method != http.MethodGet {
		p.log.Errorf("Method not allowedpath: %s | err: %v", op, r.Method)
		writeErrorResponse(w, http.StatusMethodNotAllowed, "Method Not Allowed")
		return
	}

	var order models.Order

	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		p.log.Errorf("Error decoding body path: %s | err: %v", op, err)
		writeErrorResponse(w, http.StatusBadRequest, "Error decoding body")
		return
	}

	cost, err := p.PathFinder.CalculateTheCost(order.PickupAddress, order.DeliveryAddress)
	if err != nil {
		p.log.Errorf("Error calculating the cost for path: %s | err: %v", op, err)
		writeErrorResponse(w, http.StatusInternalServerError, "Error calculating the cost")
		return
	}

	writeJSONResponse(w, http.StatusOK, cost)
	return
}
