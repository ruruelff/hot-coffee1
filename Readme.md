# ☕ Hot-Coffee: Coffee Shop Management System

A backend application written in Go to simulate a simple coffee shop system. It supports order management, menu updates, inventory tracking, and sales reporting using RESTful APIs and local JSON file storage.

---

### 📁 Project Structure
<pre>
hot-coffee/
├── cmd/
│ └── main.go
├── internal/
│ ├── handler/
│ │ ├── order_handler.go
│ │ ├── menu_handler.go
│ │ └── inventory_handler.go
│ ├── service/
│ │ ├── order_service.go
│ │ ├── menu_service.go
│ │ └── inventory_service.go
│ └── dal/
│ ├── order_repository.go
│ ├── menu_repository.go
│ └── inventory_repository.go
├── models/
│ ├── order.go
│ ├── menu_item.go
│ └── inventory_item.go
├── data/
│ ├── orders.json
│ ├── menu_items.json
│ └── inventory.json
├── go.mod
├── go.sum
└── README.md
</pre>

---

### 🧱 Architecture

#### 1. Presentation Layer (Handlers)
- Manages HTTP requests/responses
- Parses input and formats output
- Calls service methods

#### 2. Business Logic Layer (Services)
- Implements core logic
- Processes data and manages workflows
- Calls repository layer

#### 3. Data Access Layer (Repositories)
- Interacts with local JSON files
- Ensures persistence and consistency
- Provides interfaces for data operations

---

### 🚀 API Endpoints

#### 🧾 Orders
- `GET /orders`
- `GET /orders/{id}`
- `PUT /orders/{id}`
- `DELETE /orders/{id}`
- `POST /orders/{id}/close`

#### 🍽️ Menu
- `POST /menu`
- `GET /menu`
- `GET /menu/{id}`
- `PUT /menu/{id}`
- `DELETE /menu/{id}`

#### 📦 Inventory
- `POST /inventory`
- `GET /inventory`
- `GET /inventory/{id}`
- `PUT /inventory/{id}`
- `DELETE /inventory/{id}`

#### 📊 Reports
- `GET /reports/total-sales`
- `GET /reports/popular-items`

---

### 💾 Data Storage

Data is saved in local JSON files under `data/`:
- `orders.json`
- `menu_items.json`
- `inventory.json`

Example:
```json
{
  "order_id": "order001",
  "customer_name": "John Doe",
  "items": [
    { "product_id": "latte", "quantity": 2 }
  ],
  "status": "open",
  "created_at": "2023-10-01T09:00:00Z"
}
```
🛠️ Getting Started

Build and run the application:
<pre> go build -o hot-coffee ./cmd 
./hot-coffee --port 8080 --dir data </pre>


Test Cases for Orders and Menu Items

1. Test Case: Create a New Order
Request:

    Method: POST

    URL: http://localhost:8080/orders

    Body (Raw, JSON):
```bash
{
  "customer_name": "John Doe",
  "items": [
    {
      "product_id": "latte",
      "quantity": 2
    },
    {
      "product_id": "croissant",
      "quantity": 1
    }
  ]
}

Expected Response:

    Status Code: 201 Created

    Body:

{
  "order_id": "order125",
  "customer_name": "John Doe",
  "items": [
    {
      "product_id": "latte",
      "quantity": 2
    },
    {
      "product_id": "croissant",
      "quantity": 1
    }
  ],
  "status": "open",
  "created_at": "2023-10-02T09:30:00Z"
}
```
2. Test Case: Retrieve All Orders
Request:

    Method: GET

    URL: http://localhost:8080/orders

```bash
[
  {
    "order_id": "order125",
    "customer_name": "John Doe",
    "items": [
      {
        "product_id": "latte",
        "quantity": 2
      },
      {
        "product_id": "croissant",
        "quantity": 1
      }
    ],
    "status": "open",
    "created_at": "2023-10-02T09:30:00Z"
  }
]
```
3. Test Case: Retrieve a Specific Order by ID
Request:

    Method: GET

    URL: http://localhost:8080/orders/order125

<pre>
{
  "order_id": "order125",
  "customer_name": "John Doe",
  "items": [
    {
      "product_id": "latte",
      "quantity": 2
    },
    {
      "product_id": "croissant",
      "quantity": 1
    }
  ],
  "status": "open",
  "created_at": "2023-10-02T09:30:00Z"
}
</pre>

4. Test Case: Update an Order
Request:

    Method: PUT

    URL: http://localhost:8080/orders/order125

    Body (Raw, JSON):
<pre>
{
  "customer_name": "John Doe",
  "items": [
    {
      "product_id": "latte",
      "quantity": 3  // Updated quantity
    },
    {
      "product_id": "croissant",
      "quantity": 1
    }
  ],
  "status": "open",
  "created_at": "2023-10-02T09:30:00Z"
}

{
  "order_id": "order125",
  "customer_name": "John Doe",
  "items": [
    {
      "product_id": "latte",
      "quantity": 3
    },
    {
      "product_id": "croissant",
      "quantity": 1
    }
  ],
  "status": "open",
  "created_at": "2023-10-02T09:30:00Z"
}
</pre>

5. Test Case: Delete an Order
Request:

    Method: DELETE

    URL: http://localhost:8080/orders/order125

6. Test Case: Create a New Menu Item
Request:

    Method: POST

    URL: http://localhost:8080/menu

    Body (Raw, JSON):
```bash
{
  "product_id": "latte",
  "name": "Caffe Latte",
  "description": "Espresso with steamed milk",
  "price": 3.50,
  "ingredients": [
    {
      "ingredient_id": "espresso_shot",
      "quantity": 1
    },
    {
      "ingredient_id": "milk",
      "quantity": 200
    }
  ]
}

```

7. Test Case: Retrieve All Menu Items
Request:

    Method: GET

    URL: http://localhost:8080/menu


8. Test Case: Retrieve a Specific Menu Item by ID
Request:

    Method: GET

    URL: http://localhost:8080/menu/latte


9. Test Case: Update a Menu Item
Request:

    Method: PUT

    URL: http://localhost:8080/menu/latte
```bash
    Body (Raw, JSON):

{
  "product_id": "latte",
  "name": "Caffe Latte",
  "description": "Espresso with steamed milk and a touch of vanilla",
  "price": 3.75,
  "ingredients": [
    {
      "ingredient_id": "espresso_shot",
      "quantity": 1
    },
    {
      "ingredient_id": "milk",
      "quantity": 200
    },
    {
      "ingredient_id": "vanilla_extract",
      "quantity": 10
    }
  ]
}
```

10. Test Case: Delete a Menu Item
Request:

    Method: DELETE

    URL: http://localhost:8080/menu/latte

### Authors
***zhkuandyk- Zhanel Kuandyk***

***aradilkha- Aruzhan Adilkhassymkyzy***