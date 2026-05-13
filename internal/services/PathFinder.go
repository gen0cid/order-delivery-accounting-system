package services

import (
	"delivery-system/internal/database"
	"delivery-system/internal/logger"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strconv"
)

const costPerKm = 50

type PathFinder struct {
	db  *database.DB
	log *logger.Logger
}

type CurrentCityInfoLATLON struct {
	city string
	lat  float64
	lon  float64
}
type Point struct {
	Lat float64
	Lon float64
}

func NewPathFinder(db *database.DB, log *logger.Logger) *PathFinder {
	return &PathFinder{
		db:  db,
		log: log,
	}
}

func (p *PathFinder) CalculateTheCost(pickup_address, delivery_address string) (float64, error) {
	const op = "services.PathFinder.CalculateTheCost"
	lat1, lon1, err := p.FindCoords(pickup_address)
	if err != nil {
		return 0, fmt.Errorf("cannot find coordinates path: %s | err: %v", op, err)
	}

	lat2, lon2, err := p.FindCoords(delivery_address)
	if err != nil {
		return 0, fmt.Errorf("cannot find coordinates path: %s | err: %v", op, err)
	}
	point1 := Point{
		Lat: lat1,
		Lon: lon1,
	}

	point2 := Point{
		Lat: lat2,
		Lon: lon2,
	}

	distance := CalculateDistance(point1, point2)

	return distance * costPerKm, nil
}

func (p *PathFinder) FindCoords(city string) (float64, float64, error) {
	const op = "handlers.FindCoords"

	escapedCity := url.QueryEscape(city)
	apiURL := fmt.Sprintf("https://nominatim.openstreetmap.org/search?q=%s&format=json&limit=1", escapedCity)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return 0, 0, err
	}

	req.Header.Set("User-Agent", "order-delivery-accounting-system/1.0")

	// отправил запрос
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, 0, err
	}

	if resp == nil || resp.Body == nil {
		return 0, 0, fmt.Errorf("cannot find coordinates")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, 0, fmt.Errorf("cannot find coordinates: %s", resp.Status)
	}

	var results []struct {
		Lat string `json:"lat"`
		Lon string `json:"lon"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return 0, 0, fmt.Errorf("cannot decode coordinates path: %s | err: %v", op, err)
	}

	if len(results) == 0 {
		return 0, 0, fmt.Errorf("%s: no results found for city: %s", op, city)
	}

	// Конвертируем строки в float64
	lat, _ := strconv.ParseFloat(results[0].Lat, 64)
	lon, _ := strconv.ParseFloat(results[0].Lon, 64)

	return lat, lon, nil
}

func CalculateDistance(p1, p2 Point) float64 {
	const earthRadius = 6371.0 // Радиус Земли в км

	// Переводим градусы в радианы
	lat1 := p1.Lat * math.Pi / 180
	lon1 := p1.Lon * math.Pi / 180
	lat2 := p2.Lat * math.Pi / 180
	lon2 := p2.Lon * math.Pi / 180

	// Разница координат
	dLat := lat2 - lat1
	dLon := lon2 - lon1

	// Формула гаверсинусов
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1)*math.Cos(lat2)*
			math.Sin(dLon/2)*math.Sin(dLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadius * c
}
