package Controller

import (
	"encoding/json"
	"net/http"
	"prevent-race-condition/domain"
	"prevent-race-condition/helper"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Order struct {
	db *gorm.DB
}

func NewOrderController(db *gorm.DB) Order {
	return Order{db}
}
func (d Order) CreateRaceCondition(w http.ResponseWriter, r *http.Request) {
	var request domain.Order
	var stock domain.Stock
	decoder := json.NewDecoder(r.Body)

	err := decoder.Decode(&request)
	if err != nil {
		helper.MaptoHttpResponse(w, http.StatusBadRequest, domain.HttpResponse{Message: "Invalid request Payload"})
		return
	}

	err = d.db.First(&stock, request.StockId).Error
	if err != nil {

		helper.MaptoHttpResponse(w, http.StatusInternalServerError, domain.HttpResponse{Message: "Failed to create Order"})
		return
	}
	if stock.Stock <= 0 {

		helper.MaptoHttpResponse(w, http.StatusBadRequest, domain.HttpResponse{Message: "Failed to create Order,because stock is less than 0 "})
		return
	}

	if err := d.db.Create(request).Error; err != nil {
		helper.MaptoHttpResponse(w, http.StatusInternalServerError, domain.HttpResponse{Message: "Failed to create Order"})
		return
	}
	stock.Stock = stock.Stock - 1
	err = d.db.Save(&stock).Error
	if err != nil {
		helper.MaptoHttpResponse(w, http.StatusInternalServerError, domain.HttpResponse{Message: "Failed to create Order"})
		return
	}
	helper.MaptoHttpResponse(w, http.StatusCreated, domain.HttpResponse{Message: "Success Booking Order"})

}

func (d Order) CreateNoRaceCondition(w http.ResponseWriter, r *http.Request) {
	var request domain.Order
	var stock domain.Stock
	decoder := json.NewDecoder(r.Body)

	err := decoder.Decode(&request)
	if err != nil {
		helper.MaptoHttpResponse(w, http.StatusBadRequest, domain.HttpResponse{Message: "Invalid request Payload"})
		return
	}
	tx := d.db.Begin()

	err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&stock, request.StockId).Error
	if err != nil {
		defer tx.Rollback()
		helper.MaptoHttpResponse(w, http.StatusInternalServerError, domain.HttpResponse{Message: "Failed to create Order"})
		return
	}
	if stock.Stock <= 0 {
		defer tx.Rollback()
		helper.MaptoHttpResponse(w, http.StatusBadRequest, domain.HttpResponse{Message: "Failed to create Order,because stock is less than 0 "})

		return
	}

	if err := tx.Create(request).Error; err != nil {
		defer tx.Rollback()

		helper.MaptoHttpResponse(w, http.StatusInternalServerError, domain.HttpResponse{Message: "Failed to create Order"})
		return
	}
	stock.Stock--
	err = tx.Save(&stock).Error
	if err != nil {
		defer tx.Rollback()
		helper.MaptoHttpResponse(w, http.StatusInternalServerError, domain.HttpResponse{Message: "Failed to create Order"})
		return
	}
	defer tx.Commit()
	helper.MaptoHttpResponse(w, http.StatusCreated, domain.HttpResponse{Message: "Success Booking Order"})
}
