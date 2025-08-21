package service

import (
	"errors"
	"fmt"
	"hot-coffee/internal/dal"
	"hot-coffee/models"
	"strconv"
	"strings"
	"time"
)

type Order struct {
	cacheOrders   map[string]models.Order
	takenIDOrders map[string]int
}

type OrderService interface {
	GetAllOrders() ([]models.Order, error)
	GetOrderByID(ID string) (models.Order, error)
	AddNewOrder(order models.Order) error
	CloseOrder(ID string) error
	DeleteOrder(ID string) error
	ModifyOrder(order models.Order, ID string) error
	LoadOrdersCache() error
}

func NewOrderService() OrderService {
	return &Order{
		cacheOrders:   make(map[string]models.Order),
		takenIDOrders: make(map[string]int),
	}
}

func (o *Order) findOrderIndexByID(ID string) (int, error) {
	_, exists := o.takenIDOrders[ID]
	if !exists {
		return -1, fmt.Errorf("order with ID %s not found", ID)
	}
	return 0, nil
}

func (o *Order) LoadOrdersCache() error {
	orders, err := dal.NewOrderRepository().ReadOrder()
	if err != nil {
		return errors.Join(ErrOrderNotRead, err)
	}

	o.cacheOrders = make(map[string]models.Order)
	o.takenIDOrders = make(map[string]int)

	if err = validateOrders(orders); err != nil {
		return err
	}

	for _, val := range orders {
		o.cacheOrders[val.ID] = val
		o.takenIDOrders[val.ID] = 1
	}

	return nil
}

func (o *Order) GetAllOrders() ([]models.Order, error) {
	err := o.LoadOrdersCache()
	if err != nil {
		return nil, err
	}

	if len(o.cacheOrders) == 0 {
		return nil, errors.New("no orders in orders storage")
	}

	ordersSlice := make([]models.Order, 0, len(o.cacheOrders))
	for _, order := range o.cacheOrders {
		ordersSlice = append(ordersSlice, order)
	}

	return ordersSlice, nil
}

func (o *Order) GetOrderByID(ID string) (models.Order, error) {
	if err := o.LoadOrdersCache(); err != nil {
		return models.Order{}, err
	}

	order, exists := o.cacheOrders[ID]
	if !exists {
		return models.Order{}, fmt.Errorf("order with ID %s not found", ID)
	}

	return order, nil
}

func (o *Order) AddNewOrder(order models.Order) error {
	if err := o.LoadOrdersCache(); err != nil {
		return err
	}

	var lastID int
	if len(o.cacheOrders) > 0 {
		for _, ord := range o.cacheOrders {
			idStr := ord.ID[5:]
			idNum, err := strconv.Atoi(idStr)
			if err != nil {
				return fmt.Errorf("invalid ID format: %v", err)
			}
			if idNum > lastID {
				lastID = idNum
			}
		}
	}

	newID := fmt.Sprintf("order%d", lastID+1)
	order.ID = newID

	// ✅ ДОБАВЛЯЕМ ЭТУ ПРОВЕРКУ:
	for _, item := range order.Items {
		if err := validateDeductCheckIngredients(item.ProductID, float64(item.Quantity)); err != nil {
			return err
		}
	}

	if err := validateOrder(order); err != nil {
		return err
	}

	order.Status = "open"
	order.CreatedAt = time.Now().Format(time.DateTime)

	if _, exists := o.takenIDOrders[order.ID]; exists {
		return ErrConflict
	}

	o.cacheOrders[order.ID] = order

	ordersSlice := make([]models.Order, 0, len(o.cacheOrders))
	for _, v := range o.cacheOrders {
		ordersSlice = append(ordersSlice, v)
	}

	if err := dal.NewOrderRepository().WriteOrder(ordersSlice); err != nil {
		return err
	}

	return nil
}

func (o *Order) CloseOrder(ID string) error {
	m := NewMenuService()

	// ✅ Сначала загружаем кэш
	if err := o.LoadOrdersCache(); err != nil {
		return err
	}

	// ✅ Потом ищем заказ
	order, exists := o.cacheOrders[ID]
	if !exists {
		return fmt.Errorf("order with ID %s not found", ID)
	}

	// ✅ Проверка без учёта регистра
	if strings.ToLower(order.Status) == "open" {
		order.Status = "closed"
	} else {
		return fmt.Errorf("order is already closed")
	}

	if err := o.LoadOrdersCache(); err != nil {
		return err
	}

	if err := validateOrder(order); err != nil {
		return err
	}

	if err := validateCloseOrder(order); err != nil {
		return err
	}

	for _, product := range order.Items {
		if err := validateDeductCheckIngredients(product.ProductID, float64(product.Quantity)); err != nil {
			return err
		}
		if err := m.DeductMenuProduct(product.ProductID, float64(product.Quantity)); err != nil {
			return err
		}
	}

	o.cacheOrders[ID] = order

	ordersSlice := make([]models.Order, 0, len(o.cacheOrders))
	for _, v := range o.cacheOrders {
		ordersSlice = append(ordersSlice, v)
	}

	if err := dal.NewOrderRepository().WriteOrder(ordersSlice); err != nil {
		return err
	}

	return nil
}

func (o *Order) DeleteOrder(ID string) error {
	if err := o.LoadOrdersCache(); err != nil {
		return err
	}

	// Проверяем наличие заказа в карте
	_, exists := o.cacheOrders[ID]
	if !exists {
		return fmt.Errorf("order with ID %s not found", ID)
	}

	// Удаляем заказ из карты
	delete(o.cacheOrders, ID)

	// Преобразуем карту в слайс
	ordersSlice := make([]models.Order, 0, len(o.cacheOrders))
	for _, v := range o.cacheOrders {
		ordersSlice = append(ordersSlice, v)
	}

	// Записываем в репозиторий
	if err := dal.NewOrderRepository().WriteOrder(ordersSlice); err != nil {
		return err
	}

	return nil
}

func (o *Order) ModifyOrder(order models.Order, ID string) error {
	if err := o.LoadOrdersCache(); err != nil {
		return err
	}

	// Проверяем наличие заказа в карте
	existingOrder, exists := o.cacheOrders[ID]
	if !exists {
		return fmt.Errorf("order with ID %s not found", ID)
	}

	order = orderInit(order, existingOrder)
	if err := validateModifying(order, existingOrder); err != nil {
		return err
	}

	if err := validateOrder(order); err != nil {
		return err
	}

	o.cacheOrders[ID] = order

	// Преобразуем карту в слайс
	ordersSlice := make([]models.Order, 0, len(o.cacheOrders))
	for _, v := range o.cacheOrders {
		ordersSlice = append(ordersSlice, v)
	}

	// Записываем в репозиторий
	if err := dal.NewOrderRepository().WriteOrder(ordersSlice); err != nil {
		return errors.New("failed to modify order")
	}

	return nil
}

func orderInit(modifiedOrder, originalOrder models.Order) models.Order {
	if modifiedOrder.CreatedAt == "" {
		modifiedOrder.CreatedAt = originalOrder.CreatedAt
	}

	if modifiedOrder.CustomerName == "" {
		modifiedOrder.CustomerName = originalOrder.CustomerName
	}

	if modifiedOrder.Items == nil {
		modifiedOrder.Items = originalOrder.Items
	}

	if modifiedOrder.Status == "" {
		modifiedOrder.Status = originalOrder.Status
	}

	return modifiedOrder
}

func validateOrders(Orders []models.Order) error {
	takenIdOrder := make(map[string]int)
	for i, val := range Orders {
		if _, exists := takenIdOrder[val.ID]; exists {
			return errors.New("duplicated order id")
		}
		takenIdOrder[val.ID] = i

		if _, exists := takenIdOrder[val.ID]; !exists {
			return fmt.Errorf("item with order ID %s does not exist", val.ID)
		}
		for _, items := range val.Items {
			if items.Quantity < 1 {
				return fmt.Errorf("item with quantity %v is less than 1", items.Quantity)
			}
		}
	}
	return nil
}

func validateOrder(order models.Order) error {
	varTakenIdOrder := make(map[string]int)
	m := NewMenuService()
	if order.ID < "0" {
		return errors.New("order ID cannot be negative")
	} else if len(order.Items) == 0 {
		return errors.New("empty order")
	} else if order.CustomerName == "" {
		return errors.New("customer name cannot be empty")
	} else if order.Items == nil {
		return errors.New("empty order")
	}
	for i, item := range order.Items {
		product, err := m.GetMenuByID(item.ProductID)
		if err != nil {
			return err
		}
		if _, exists := varTakenIdOrder[item.ProductID]; exists {
			return errors.New("duplicated products in order")
		}
		varTakenIdOrder[item.ProductID] = i
		if item.Quantity <= 0 {
			return fmt.Errorf("item with quantity %v is less than or equal to 0", item.Quantity)
		}
		if err := validatePostMenu(product); err != nil {
			return err
		}
	}
	return nil
}

func validateCloseOrder(order models.Order) error {
	if order.ID < "0" {
		return errors.New("order ID cannot be negative")
	}
	if order.CustomerName == "" {
		return errors.New("customer name cannot be empty")
	}
	if order.Items == nil {
		return errors.New("items cannot be null")
	}
	if order.Status == "Closed" {
		return errors.New("order is already closed")
	}
	return nil
}

func validateDeductCheckIngredients(productID string, quantity float64) error {
	m := NewMenuService()

	item, err := m.GetMenuByID(productID)
	if err != nil {
		return err
	}
	for _, ingredient := range item.Ingredients {
		requiredQuantity := ingredient.Quantity * quantity
		if err := CheckInventoryAvailability(ingredient.IngredientID, requiredQuantity); err != nil {
			return fmt.Errorf("not enough %s (required: %.2f)", ingredient.IngredientID, requiredQuantity)
		}
	}
	return nil
}

func validateAggregation(order models.Order) error {
	if order.Status == "" {
		return errors.New("order status cannot be empty")
	}
	return nil
}

func CheckInventoryAvailability(ingredientID string, requiredQuantity float64) error {
	i := NewInventoryService()
	item, err := i.GetInventoryByID(ingredientID)
	if err != nil {
		return err
	}

	if item.Quantity < requiredQuantity {
		return fmt.Errorf("not enough quantity for ingredient %s", ingredientID)
	}

	return nil
}

func validateModifying(modifiedOrder, originalOrder models.Order) error {
	if modifiedOrder.ID != originalOrder.ID {
		return errors.New("order with id does not match")
	}

	// if originalOrder.Status != modifiedOrder.Status {
	// 	return errors.New("modifying status is not permitted")
	// }

	if modifiedOrder.Status != "Open" && modifiedOrder.Status != "Closed" {
		return errors.New("wrong order status (should be \"Closed\" or \"Open\")")
	}

	if originalOrder.CreatedAt != modifiedOrder.CreatedAt {
		return errors.New("modifying created time is not permitted")
	}

	if originalOrder.ID == modifiedOrder.ID &&
		originalOrder.CustomerName == modifiedOrder.CustomerName &&
		areOrderItemsEqual(originalOrder.Items, modifiedOrder.Items) &&
		originalOrder.Status == modifiedOrder.Status &&
		originalOrder.CreatedAt == modifiedOrder.CreatedAt {

		return ErrNothingToModify
	}

	return nil
}

func areOrderItemsEqual(a, b []models.OrderItem) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i].ProductID != b[i].ProductID || a[i].Quantity != b[i].Quantity {
			return false
		}
	}
	return true
}
