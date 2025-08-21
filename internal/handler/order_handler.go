package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"hot-coffee/internal/service"
	"hot-coffee/models"
	"log/slog"
	"net/http"
	"strconv"
)

var OrderService = service.NewOrderService()

func OrderEndpoints(mux *http.ServeMux) {
	mux.HandleFunc("POST /orders", PostOrderHandler)
	mux.HandleFunc("POST /orders/", PostOrderHandler)

	mux.HandleFunc("GET /orders", GetAllOrdersHandler)
	mux.HandleFunc("GET /orders/", GetAllOrdersHandler)

	mux.HandleFunc("GET /orders/{id}", GetOrderByIDHandler)
	mux.HandleFunc("GET /orders/{id}/", GetOrderByIDHandler)

	mux.HandleFunc("PUT /orders/{id}", PutOrderHandler)
	mux.HandleFunc("PUT /orders/{id}/", PutOrderHandler)

	mux.HandleFunc("DELETE /orders/{id}", DeleteOrderByIDHandler)
	mux.HandleFunc("DELETE /orders/{id}/", DeleteOrderByIDHandler)

	mux.HandleFunc("POST /orders/{id}/close", PostOrderCloserHandler)
	mux.HandleFunc("POST /orders/{id}/close/", PostOrderCloserHandler)
}

func GetAllOrdersHandler(w http.ResponseWriter, r *http.Request) {
	orders, err := OrderService.GetAllOrders()
	if err != nil {
		ErrorResponse(w, "Could not retrieve orders data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	jsonData, err := json.MarshalIndent(orders, "", "    ")
	if err != nil {
		ErrorResponse(w, "Failed to encode inventory items", http.StatusInternalServerError)
		return
	}

	if _, err = w.Write(jsonData); err != nil {
		ErrorResponse(w, "Failed to write response", http.StatusInternalServerError)
		return
	}
	slog.Info("Retrieved all orders")
}

func GetOrderByIDHandler(w http.ResponseWriter, r *http.Request) {
	idString := r.PathValue("id")

	order, err := OrderService.GetOrderByID(idString) // ID передаем как string
	if errors.Is(err, service.ErrOrderNotRead) {
		ErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	} else if err != nil {
		ErrorResponse(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	jsonData, err := json.MarshalIndent(order, "", "    ")
	if err != nil {
		ErrorResponse(w, "Failed to encode order", http.StatusInternalServerError)
		return
	}

	if _, err = w.Write(jsonData); err != nil {
		ErrorResponse(w, "Failed to write response", http.StatusInternalServerError)
		return
	}

	slog.Info("Retrieved order", "ID", order.ID)
}

func PostOrderHandler(w http.ResponseWriter, r *http.Request) {
	order, err := parseOrder(r)
	if errors.Is(err, ErrUnsupportedContentType) {
		ErrorResponse(w, err.Error(), http.StatusUnsupportedMediaType)
	} else if err != nil {
		ErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err = OrderService.AddNewOrder(order); errors.Is(err, service.ErrConflict) {
		ErrorResponse(w, err.Error(), http.StatusConflict)
		return
	} else if err != nil {
		ErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	if _, err = w.Write([]byte("Order added successfully")); err != nil {
		ErrorResponse(w, "Failed to write response", http.StatusInternalServerError)
		return
	}

	slog.Info("Added new order", "ID", order.ID)
}

func PostOrderCloserHandler(w http.ResponseWriter, r *http.Request) {
	idString := r.PathValue("id") // id как строка

	if err := OrderService.CloseOrder(idString); errors.Is(err, service.ErrNotExists) {
		ErrorResponse(w, err.Error(), http.StatusNotFound)
		return
	} else if err != nil {
		ErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	if _, err := w.Write([]byte("Order closed successfully")); err != nil {
		ErrorResponse(w, "Failed to write response", http.StatusInternalServerError)
		return
	}
	slog.Info("Closed order", "ID", idString)
}

func PutOrderHandler(w http.ResponseWriter, r *http.Request) {
	order, err := parseOrder(r)
	if errors.Is(err, ErrUnsupportedContentType) {
		ErrorResponse(w, err.Error(), http.StatusUnsupportedMediaType)
		return
	} else if err != nil {
		ErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	idString := r.PathValue("id") // id как строка

	if err = OrderService.ModifyOrder(order, idString); errors.Is(err, service.ErrConflict) {
		ErrorResponse(w, err.Error(), http.StatusConflict)
		return
	} else if err != nil {
		ErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)

	if _, err = w.Write([]byte("Order is updated successfully")); err != nil {
		ErrorResponse(w, "Failed to write response", http.StatusInternalServerError)
		return
	}

	slog.Info("Updated order", "ID", order.ID)
}

func DeleteOrderByIDHandler(w http.ResponseWriter, r *http.Request) {
	idString := r.PathValue("id") // id как строка

	err := OrderService.DeleteOrder(idString) // ID передаем как string
	if errors.Is(err, service.ErrOrderNotRead) {
		ErrorResponse(w, err.Error(), http.StatusNotFound)
		return
	} else if err != nil {
		ErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)

	slog.Info("Deleted order", "ID", idString)
}

func parseOrder(r *http.Request) (models.Order, error) {
	var order models.Order
	contentType := r.Header.Get("Content-Type")

	if contentType == "application/json" {
		if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
			return order, fmt.Errorf("invalid JSON payload")
		}
	} else if contentType == "application/x-www-form-urlencoded" {
		if err := r.ParseForm(); err != nil {
			return order, fmt.Errorf("invalid form data")
		}

		ID, err := strconv.Atoi(r.FormValue("order_id"))
		if err != nil {
			return order, fmt.Errorf("ID is not an integer")
		}

		var items []models.OrderItem
		itemsJson := r.FormValue("items")
		if err := json.Unmarshal([]byte(itemsJson), &items); err != nil {
			return order, fmt.Errorf("error parsing ingredients: %v", err)
		}

		// Используем строковый ID
		order = models.Order{
			ID:           fmt.Sprintf("%d", ID), // Преобразуем ID в строку
			CustomerName: r.FormValue("customer_name"),
			Items:        items,
			Status:       r.FormValue("status"),
			CreatedAt:    r.FormValue("created_at"),
		}
	} else {
		return order, fmt.Errorf("unsupported content type")
	}

	return order, nil
}
