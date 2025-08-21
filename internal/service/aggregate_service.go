package service

import (
	"errors"
	"hot-coffee1/models"
	"sort"
	"strings"
)

var (
	ErrNotFoundID             = errors.New("id was not found")
	ErrUnsupportedContentType = errors.New("unsupported content type")
)

func GetTotalSales() (models.TotalSales, error) {
	m := NewMenuService()
	totalSales := models.TotalSales{}
	ordersStruct := Order{}

	// Загружаем кеш заказов
	err := ordersStruct.LoadOrdersCache()
	if err != nil {
		return totalSales, err
	}

	// Загружаем кеш меню
	err = m.LoadMenuCache()
	if err != nil {
		return totalSales, err
	}

	// Проверка: если нет заказов — ошибка
	if len(ordersStruct.cacheOrders) == 0 {
		return totalSales, ErrOrderNotRead
	}

	// Обработка каждого заказа
	for _, order := range ordersStruct.cacheOrders {
		status := strings.ToLower(order.Status) // Приводим к нижнему регистру

		if status == "closed" {
			for _, product := range order.Items {
				// Валидация
				if err = validateAggregation(order); err != nil {
					return totalSales, err
				}

				// Получаем данные продукта из меню
				menu, errMenu := m.GetMenuByID(product.ProductID)
				if errMenu != nil {
					return models.TotalSales{}, errMenu
				}

				// Увеличиваем итоговую сумму
				totalSales.Amount += float64(product.Quantity) * menu.Price
			}
		} else if status != "open" && status != "closed" {
			return models.TotalSales{}, errors.New("order is not closed")
		}
	}

	return totalSales, nil
}

func GetPopularItems() ([]models.PopularItem, error) {
	o := NewOrderService()
	allOrders, err := o.GetAllOrders()
	if err != nil {
		return nil, err
	}
	if len(allOrders) == 0 {
		return nil, ErrOrderNotRead
	}

	sumProdID := map[string]int{}

	for _, order := range allOrders {
		status := strings.ToLower(order.Status) // ✅ Учитываем регистр

		if status == "closed" {
			for _, product := range order.Items {
				if err := validateAggregation(order); err != nil {
					return nil, err
				}
				if product.Quantity <= 0 {
					return nil, errors.New("quantity is <= 0")
				}
				sumProdID[product.ProductID] += product.Quantity
			}
		} else if status != "open" && status != "closed" {
			return nil, errors.New("order has unknown status")
		}
	}

	return GetTopItemsByQuantity(sumProdID, 3), nil
}

func GetTopItemsByQuantity(productQuantities map[string]int, topN int) []models.PopularItem {
	m := NewMenuService()
	var quantities []models.OrderItem
	for id, quantity := range productQuantities {
		quantities = append(quantities, models.OrderItem{id, quantity})
	}

	sort.Slice(quantities, func(i, j int) bool {
		return quantities[i].Quantity > quantities[j].Quantity
	})

	var topItems []models.PopularItem
	for i := 0; i < len(quantities) && i < topN; i++ {
		menu, menuErr := m.GetMenuByID(quantities[i].ProductID)
		if menuErr != nil {
			return []models.PopularItem{}
		}
		topItems = append(topItems, models.PopularItem{
			Quantity:    quantities[i].Quantity,
			ID:          menu.ID,
			Name:        menu.Name,
			Description: menu.Description,
			Price:       menu.Price,
			Ingredients: menu.Ingredients,
		})
	}

	return topItems
}
