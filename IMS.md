# Inventory Management System - SBA Report

Name: Hussain Sameer
Class: 6B
CSNO: 6217

---

## 1. Introduction

### Content Page

1. [Introduction](#1-introduction)
   - [Content Page](#content-page)
   - [Data Dictionary](#data-dictionary)
   - [Overview of Functions](#overview-of-functions)
2. [System Flowchart](#2-system-flowchart)
3. [Basic Features Implementation](#3-basic-features-implementation)
   - [User Authentication](#user-authentication)
   - [Inventory Management](#inventory-management)
   - [Transaction Processing](#transaction-processing)
4. [Other Special Features](#4-other-special-features)
   - [GUI Implementation](#gui-implementation)
   - [Database Design](#database-design)
   - [Reusability of Code](#reusability-of-code)
   - [Error Handling](#error-handling)
5. [Comparisons to Other Alternatives](#5-comparisons-to-other-alternatives)
   - [Go vs Python](#go-vs-python)
   - [Fyne vs QT](#fyne-vs-qt)
   - [SQLite vs Simpler Text File/JSON Database](#sqlite-vs-simpler-text-filejson-database)

---

### Data Dictionary

#### Database Tables

**users**
- `id` (INTEGER): Primary key, unique identifier for each user
- `username` (TEXT): Unique username for login, cannot be null
- `password_hash` (TEXT): Bcrypt-encrypted password, stored as string
- `is_root_admin` (INTEGER): Boolean flag (0 or 1) indicating if user has full system access
- `can_read` (INTEGER): Boolean flag (0 or 1) for permission to view inventory
- `can_transaction` (INTEGER): Boolean flag (0 or 1) for permission to process sales
- `can_revenue` (INTEGER): Boolean flag (0 or 1) for permission to view revenue reports
- `created_at` (DATETIME): Timestamp of when the user account was created

**items**
- `id` (INTEGER): Primary key, unique identifier for each item
- `name` (TEXT): Display name of the item, cannot be null
- `code` (TEXT): Unique barcode or QR code identifier, cannot be null
- `description` (TEXT): Optional detailed description of the item, can be null
- `price` (REAL): Selling price per unit, stored as float (REAL is the equivalent datatype in SQLite)
- `cost` (REAL): Cost price per unit for profit calculation, stored as float, default 0
- `quantity` (INTEGER): Current stock quantity, default 0
- `in_stock_date` (DATETIME): Date when item was first added to inventory
- `expiry_date` (DATETIME): Optional expiration date for perishable items, can be null
- `created_at` (DATETIME): Timestamp of item creation
- `updated_at` (DATETIME): Timestamp of last modification

**item_stock**
- `id` (INTEGER): Primary key, auto-incremented unique identifier for each stock batch
- `item_id` (INTEGER): Foreign key referencing items.id, links batch to parent item
- `quantity` (INTEGER): Quantity in this specific batch, cannot be null
- `in_stock_date` (DATETIME): Date when this batch was received
- `expiry_date` (DATETIME): Optional expiration date for this batch, can be null

**transactions**
- `id` (INTEGER): Primary key, auto-incremented unique transaction identifier
- `user_id` (INTEGER): Foreign key referencing users.id, identifies who processed the transaction
- `total_amount` (REAL): Total sale amount calculated from all items, stored as float64
- `created_at` (DATETIME): Timestamp of when the transaction occurred

**transaction_items**
- `id` (INTEGER): Primary key, auto-incremented unique identifier for transaction item
- `transaction_id` (INTEGER): Foreign key referencing transactions.id, links item to transaction
- `item_id` (INTEGER): Foreign key referencing items.id, identifies which item was sold
- `quantity` (INTEGER): Quantity sold in this transaction, cannot be null
- `price` (REAL): Price per unit at time of sale, stored as float64 for historical accuracy

#### Go Data Structures

Models are essentially `structs` for any SQL package in go, with the benefit of having more robust error handling for data provided by database (i.e., any error that may occur when database is updated)

**models.User**
- `ID` (int): Database primary key
- `Username` (string): Login username
- `IsRootAdmin` (bool): Full system access flag
- `CanRead` (bool): Read-only access permission
- `CanTransaction` (bool): Transaction mode permission
- `CanRevenue` (bool): Revenue reporting permission
- `CreatedAt` (time.Time): Account creation timestamp

**models.Item**
- `ID` (int): Database primary key
- `Name` (string): Display name
- `Code` (string): Barcode/QR code
- `Description` (string): Optional description
- `Price` (float64): Selling price
- `Cost` (float64): Cost price for profit calculation
- `Quantity` (int): Current stock level
- `InStockDate` (time.Time): First stock date
- `ExpiryDate` (*time.Time): Optional expiration date (pointer allows nil)
- `CreatedAt` (time.Time): Creation timestamp
- `UpdatedAt` (time.Time): Last modification timestamp

**models.ItemStock**
- `ID` (int): Batch identifier
- `ItemID` (int): Parent item reference
- `Quantity` (int): Quantity in this batch
- `InStockDate` (time.Time): Batch receipt date
- `ExpiryDate` (*time.Time): Optional batch expiration date

**models.Transaction**
- `ID` (int): Transaction identifier
- `UserID` (int): Processing user ID
- `TotalAmount` (float64): Total sale amount
- `CreatedAt` (time.Time): Transaction timestamp
- `Items` ([]TransactionItem): Slice of transaction items

**models.TransactionItem**
- `ID` (int): Transaction item identifier
- `TransactionID` (int): Parent transaction reference
- `ItemID` (int): Item reference
- `ItemName` (string): Denormalized item name for display
- `Quantity` (int): Quantity sold
- `Price` (float64): Price at time of sale

#### Important Variables

**AppState** (struct in auth package)
- `db` (interface): Database connection with GetDB() and ResetDatabase() methods
- `user` (*models.User): Currently authenticated user, nil if not logged in

**Database** (struct in database package)
- `db` (*sql.DB): SQLite database connection

**Database Interface** (used in packages that handle logic)
- `GetDB() *sql.DB`: Method to access underlying database connection

---

### Overview of Functions

#### Database Package (database/database.go)

- NewDatabase() (*Database, error)
  - Purpose: Initializes database connection and creates schema
  - Returns: Database instance or error
  - Called: Once at application startup from main.go

- **Close() error**
  - Purpose: Closes database connection gracefully
  - Returns: Error if close fails
  - Called: On application shutdown

- GetDB() *sql.DB
  - Purpose: Exposes database connection for other packages
  - Returns: SQL database connection
  - Called: By business logic packages to execute queries

- **ResetDatabase() error**
  - Purpose: Deletes all data and reinitializes database
  - Returns: Error if reset fails
  - Called: By root admin through GUI

#### Authentication Package (auth/auth.go)

- **HashPassword(password string) (string, error)**
  - Purpose: Securely hashes password using bcrypt
  - Parameters: Plaintext password string
  - Returns: Bcrypt hash string or error
  - Called: When creating users or updating passwords

- **CheckPasswordHash(password, hash string) bool**
  - Purpose: Verifies password against bcrypt hash
  - Parameters: Plaintext password and stored hash
  - Returns: True if password matches, false otherwise
  - Called: During user authentication

- **NewAppState(db interface) *AppState**
  - Purpose: Creates new application state instance
  - Parameters: Database interface
  - Returns: AppState instance
  - Called: Once at application startup

- **Authenticate(username, password string) (*models.User, error)**
  - Purpose: Validates user credentials and creates session
  - Parameters: Username and password strings
  - Returns: User object with permissions or error
  - Called: When user attempts to log in

- **GetCurrentUser() *models.User**
  - Purpose: Returns currently logged in user
  - Returns: User object or nil
  - Called: Throughout application for permission checks

#### Inventory Package (inventory/inventory.go, inventory/stock.go)

- **CreateItem(db Database, name, code, description string, price, cost float64, quantity int) (*models.Item, error)**
  - Purpose: Creates new inventory item and initial stock batch
  - Parameters: Item details and initial quantity
  - Returns: Created item object or error
  - Called: When adding new items through GUI

- **GetItemByID(db Database, id int) (*models.Item, error)**
  - Purpose: Retrieves item by database ID
  - Parameters: Item ID
  - Returns: Item object or error
  - Called: After creating or updating items

- **GetItemByCode(db Database, code string) (*models.Item, error)**
  - Purpose: Retrieves item by barcode/QR code
  - Parameters: Barcode code string
  - Returns: Item object or error
  - Called: When scanning barcodes in transaction mode

- **SearchItems(db Database, query string) ([]models.Item, error)**
  - Purpose: Searches items by name or code using LIKE pattern
  - Parameters: Search query string
  - Returns: Slice of matching items or error
  - Called: When user searches for items

- **GetAllItems(db Database) ([]models.Item, error)**
  - Purpose: Retrieves all items in inventory
  - Returns: Slice of all items or error
  - Called: To display complete inventory list

- **UpdateItem(db Database, id int, name, code, description string, price, cost float64, quantity int) error**
  - Purpose: Updates existing item information
  - Parameters: Item ID and new values
  - Returns: Error if update fails
  - Called: When editing items through GUI

- **RestockItem(db Database, itemID int, quantity int, expiryDate *time.Time) error**
  - Purpose: Adds stock to existing item and creates new batch
  - Parameters: Item ID, quantity to add, optional expiry date
  - Returns: Error if restock fails
  - Called: When restocking items through GUI

- **GetItemStockBatches(db Database, itemID int) ([]models.ItemStock, error)**
  - Purpose: Retrieves all stock batches for an item
  - Parameters: Item ID
  - Returns: Slice of stock batches or error
  - Called: When selecting batch during transaction

#### Users Package (users/users.go)

- **CreateUser(db Database, username, password string, canRead, canTransaction, canRevenue bool) (*models.User, error)**
  - Purpose: Creates new user account with permissions
  - Parameters: Username, password, permission flags
  - Returns: Created user object or error
  - Called: When root admin creates new user

- **GetAllUsers(db Database) ([]models.User, error)**
  - Purpose: Retrieves all users in system
  - Returns: Slice of all users or error
  - Called: To display user list in management tab

- **UpdateUserPermissions(db Database, id int, canRead, canTransaction, canRevenue bool) error**
  - Purpose: Updates user permissions
  - Parameters: User ID and new permission flags
  - Returns: Error if update fails
  - Called: When root admin modifies user permissions

- **DeleteUser(db Database, id int) error**
  - Purpose: Deletes user account (prevents root admin deletion)
  - Parameters: User ID
  - Returns: Error if deletion fails or if trying to delete root admin
  - Called: When root admin deletes user

#### Transactions Package (transactions/transactions.go)

- **CreateTransaction(db Database, userID int, items []models.TransactionItem) (*models.Transaction, error)**
  - Purpose: Processes sale and updates inventory
  - Parameters: User ID and list of items being sold
  - Returns: Complete transaction object or error
  - Called: When completing transaction in transaction mode

- **GetTransactionByID(db Database, id int) (*models.Transaction, error)**
  - Purpose: Retrieves complete transaction with all items
  - Parameters: Transaction ID
  - Returns: Transaction object with items or error
  - Called: When viewing transaction details

- **GetRecentTransactions(db Database, limit int) ([]models.Transaction, error)**
  - Purpose: Retrieves recent transactions ordered by date
  - Parameters: Maximum number of transactions to return
  - Returns: Slice of transactions or error
  - Called: To display transaction log

#### Barcode Package (barcode/barcode.go)

- **DecodeBarcodeFromImage(filePath string) (string, error)**
  - Purpose: Extracts barcode or QR code from image file
  - Parameters: Path to image file
  - Returns: Decoded barcode string or error
  - Called: When user uploads barcode image in transaction mode

#### GUI Package (gui/*.go)

- **ShowLoginWindow(appState *auth.AppState)**
  - Purpose: Displays login window and handles authentication
  - Parameters: Application state
  - Called: At application startup and after logout

- **ShowMainWindow(myApp fyne.App, appState *auth.AppState, user *models.User)**
  - Purpose: Displays main application window with tabs
  - Parameters: Fyne app, application state, logged in user
  - Called: After successful login

- **createInventoryTab(parent fyne.Window, appState *auth.AppState, user *models.User) *container.Scroll**
  - Purpose: Creates inventory management tab
  - Returns: Scrollable container with inventory interface
  - Called: When main window is created for users with read permission

- **createTransactionTab(parent fyne.Window, appState *auth.AppState, user *models.User) *container.Scroll**
  - Purpose: Creates transaction/counter mode tab
  - Returns: Scrollable container with transaction interface
  - Called: When main window is created for users with transaction permission

- **createRevenueTab(parent fyne.Window, appState *auth.AppState, user *models.User) *container.Scroll**
  - Purpose: Creates revenue reporting tab
  - Returns: Scrollable container with revenue interface
  - Called: When main window is created for users with revenue permission

---

## 2. System Flowchart

This section explains how the different parts of the program tie into each other, showing what calls what and the general function of each file.

### Overall System Architecture

[system.png]

### File Dependencies and Call Flow

**main.go** is the entry point that
- Calls `database.NewDatabase()` to set up the database
- Calls `auth.NewAppState()` to create application state
- Calls `gui.ShowLoginWindow()` to start the user interface

**database/database.go** provides
- Database connection management
- Schema initialization
- Root admin creation
- Database reset functionality
- Accessed by all logic packages through `GetDB()` method

**auth/auth.go** provides:
- Password hashing and verification
- User authentication
- Session management through AppState
- Called by `gui/login.go` for login
- Accessed by GUI for permission checks

**inventory/inventory.go** and **inventory/stock.go** provides:
- Stock batch management
- Called by `gui/gui.go` for inventory management
- Called by `gui/gui_transaction.go` for item lookups
- Called by `gui/gui_restock.go` for restocking

**users/users.go** provides
- User account management
- Permission updates
- Called by `gui/gui_users.go` for user management (root admin only)

**transactions/transactions.go** provides
- Transaction creation and processing
- Transaction history retrieval
- Called by `gui/gui_transaction.go` when completing sales
- Called by `gui/gui_transactions.go` for transaction log

**barcode/barcode.go** provides:
- Barcode/QR code decoding from images
- Called by `gui/gui_transaction.go` when user uploads barcode image

**gui/** package contains all user interface code:
- `login.go`: Login window
- `gui.go`: Main window and inventory tab
- `gui_transaction.go`: Transaction mode interface
- `gui_transactions.go`: Transaction log viewer
- `gui_users.go`: User management interface
- `gui_revenue.go`: Revenue reporting
- `gui_restock.go`: Restock dialog
- `gui_stock_selection.go`: Batch selection dialog
- `dialog_helpers.go`: Styled dialog utilities

All GUI files use the business logic packages to perform operations and the database package (through AppState) to access data.

---

## 3. Basic Features Implementation

This section covers the implementation of the three core features: user authentication, inventory management, and transaction processing. Each includes a flowchart, program code with explanation, and description of program output.

### User Authentication

#### Flowchart

**Flowchart Explanation:**

The login system follows this flow:

1. **Application Start**: `main.go` calls `ShowLoginWindow()` which creates the login interface
2. **Window Creation**: Login window is created with username field, password field (masked), and login button
3. **User Input**: User enters credentials and clicks login button
4. **Validation**: System checks if both fields are filled
5. **Database Query**: `Authenticate()` function queries database for user by username
6. **User Lookup**: If user not found, returns "invalid credentials" (prevents username enumeration)
7. **Password Verification**: Uses bcrypt to compare entered password with stored hash
8. **User Creation**: If password matches, creates User struct with all permissions from database
9. **Session Storage**: Stores authenticated user in AppState for session management
10. **Success/Error**: Either shows main window (success) or displays error message (failure)

**Key Security Features:**
- Passwords are never stored in plain text (bcrypt hashing)
- Same error message for invalid username or password (prevents username enumeration)
- User session stored in AppState after successful authentication
- All database errors are properly handled and returned

#### Program Code + Explanation

**auth/auth.go - HashPassword Function**

```go
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}
```

**Explanation:**
- **Process:** Uses bcrypt library to generate password hash with default cost (10 rounds)
- **Output:** Returns hashed password as string and error if hashing fails
- **Called:** When creating new users or updating passwords

**auth/auth.go - CheckPasswordHash() Function**

```go
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
```

**Explanation:**
- **Process:** Uses bcrypt to compare password with stored hash
- **Called:** During user authentication

**auth/auth.go - Authenticate Function**

```go
func (a *AppState) Authenticate(username, password string) (*models.User, error) {
	var id int
	var usernameDB, passwordHash string
	var isRootAdmin, canRead, canTransaction, canRevenue int
	var createdAt time.Time

	err := a.db.GetDB().QueryRow(
		"SELECT id, username, password_hash, is_root_admin, can_read, can_transaction, can_revenue, created_at FROM users WHERE username = ?",
		username,
	).Scan(&id, &usernameDB, &passwordHash, &isRootAdmin, &canRead, &canTransaction, &canRevenue, &createdAt)

	if err == sql.ErrNoRows {
		return nil, errors.New("invalid credentials")
	}
	if err != nil {
		return nil, err
	}

	if !CheckPasswordHash(password, passwordHash) {
		return nil, errors.New("invalid credentials")
	}

	user := &models.User{
		ID:             id,
		Username:       usernameDB,
		IsRootAdmin:    isRootAdmin == 1,
		CanRead:        canRead == 1,
		CanTransaction: canTransaction == 1,
		CanRevenue:     canRevenue == 1,
		CreatedAt:      createdAt,
	}

	a.user = user
	return user, nil
}
```

**Explanation:**
- **Process:**
  1. Queries database for user by username
  2. If user not found, returns "invalid credentials" error (prevents username enumeration)
  3. Retrieves password hash and all permission flags
  4. Compares entered password with stored hash using CheckPasswordHash
  5. If password matches, creates User struct with all permissions
  6. Stores user in AppState for session management
- **Output:** Returns User object with permissions or error
- **Called:** When user attempts to log in from login window

**gui/login.go - Login Window**

```go
func ShowLoginWindow(appState *auth.AppState) {
	myApp := app.New()
	loginWindow := myApp.NewWindow("Login - Inventory Management System")
	loginWindow.Resize(fyne.NewSize(400, 300))
	loginWindow.CenterOnScreen()

	usernameEntry := widget.NewEntry()
	usernameEntry.SetPlaceHolder("Username")
	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("Password")

	statusLabel := widget.NewLabel("")
	statusLabel.Wrapping = fyne.TextWrapWord

	loginBtn := widget.NewButton("Login", func() {
		username := usernameEntry.Text
		password := passwordEntry.Text

		if username == "" || password == "" {
			statusLabel.SetText("Please enter both username and password")
			return
		}

		user, err := appState.Authenticate(username, password)
		if err != nil {
			statusLabel.SetText("Invalid credentials")
			return
		}

		// Login successful, show main window
		loginWindow.Hide()
		ShowMainWindow(myApp, appState, user)
	})

	content := container.NewVBox(
		widget.NewLabel("Inventory Management System"),
		widget.NewSeparator(),
		usernameEntry,
		passwordEntry,
		loginBtn,
		statusLabel,
	)

	loginWindow.SetContent(container.NewCenter(content))
	loginWindow.ShowAndRun()
}
```

**Explanation:**
- **Input:** `appState` (Type: *auth.AppState) - Application state for authentication
- **Process:**
  1. Creates Fyne application and login window
  2. Creates username and password entry fields
  3. Creates login button with handler that:
     - Validates that both fields are filled
     - Calls AppState.Authenticate() with entered credentials
     - Shows error message if authentication fails
     - Hides login window and shows main window if successful
- **Output:** Displays login window
- **Called:** At application startup or if user decides to log out

#### Program Output

When the application starts, a login window appears. Users can enter credentials to login.
One of the following will happen.

**Successful Login:**
- Login window disappears
- Main window appears with tabs based on user permissions
- User session stored in AppState

**Failed Login:**
- Status label displays "Invalid credentials"

**Default Credentials:**
- Username: "admin"
- Password: "admin"
- Created automatically when app is first open or if no admin exists

---

### Inventory Management

#### Flowchart

[ims.png]

#### Program Code + Explanation

**inventory/inventory.go - CreateItem Function**

```go
func CreateItem(db Database, name, code, description string, price, cost float64, quantity int) (*models.Item, error) {
	now := time.Now()

  // Question marks refer to the respective values. Similar in function to "%s" in Printf
	result, err := db.GetDB().Exec(
		"INSERT INTO items (name, code, description, price, cost, quantity, in_stock_date, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		name, code, description, price, cost, quantity, now, now, now,
	) 
  
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	// Create stock entry
	_, err = db.GetDB().Exec(
		"INSERT INTO item_stock (item_id, quantity, in_stock_date) VALUES (?, ?, ?)",
		id, quantity, now,
	)
	if err != nil {
		return nil, err
	}

	return GetItemByID(db, int(id))
}
```

**Explanation:**
- **Input:**
  - `db` (Database) - Database interface
  - `name` (string) - Item name
  - `code` (string) - Unique barcode/QR code
  - `description` (string) - Optional description
  - `price` (float) - Selling price
  - `cost` (float) - Cost price
  - `quantity` (int) - Initial stock quantity
- **Process:**
  1. Gets current timestamp for dates
  2. Inserts new item into items table with all provided information
  3. Gets the auto-generated ID of the new item
  4. Creates initial stock entry in item_stock table
  5. Retrieves and returns complete item object
- **Output:** Returns created Item object with all fields populated or error
- **Called:** When user clicks "Add" button in inventory tab after filling form
- **Error Handling:** Returns error if code already exists (unique constraint) or database operation fails

**inventory/inventory.go - SearchItems Function**

```go
func SearchItems(db Database, query string) ([]models.Item, error) {
	rows, err := db.GetDB().Query(
		"SELECT id, name, code, description, price, cost, quantity, in_stock_date, expiry_date, created_at, updated_at FROM items WHERE name LIKE ? OR code LIKE ? ORDER BY name",
		"%"+query+"%", "%"+query+"%",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.Item
	for rows.Next() {
		var item models.Item
		var createdAt, updatedAt, inStockDate time.Time
		var expiryDate sql.NullTime

		err := rows.Scan(&item.ID, &item.Name, &item.Code, &item.Description, &item.Price, &item.Cost, &item.Quantity, &inStockDate, &expiryDate, &createdAt, &updatedAt)
		if err != nil {
			return nil, err
		}

		item.InStockDate = inStockDate
		if expiryDate.Valid {
			item.ExpiryDate = &expiryDate.Time
		}
		item.CreatedAt = createdAt
		item.UpdatedAt = updatedAt
		items = append(items, item)
	}

	return items, nil
}
```

**Explanation:**
- **Input:** `db` (Database) - Database interface, `query` (string) - Search term
- **Process:**
  1. Uses SQL LIKE operator with wildcards (%query%) to search both name and code fields
  2. Orders results by name alphabetically
  3. Loops through all matching rows
  4. Scans each row into Item struct, handling nullable expiry_date with sql.NullTime
  5. Converts sql.NullTime to *time.Time pointer if valid
  6. Appends each item to results slice
- **Output:** Returns slice of matching Item objects or error
- **Called:** When user types in search box in inventory or transaction tab
- **Search Behavior:** Partial matches work (e.g., "milk" finds "Whole Milk", "Skim Milk")

**inventory/inventory.go - GetItemByCode Function**

```go
func GetItemByCode(db Database, code string) (*models.Item, error) {
	var item models.Item
	var createdAt, updatedAt, inStockDate time.Time
	var expiryDate sql.NullTime

	err := db.GetDB().QueryRow(
		"SELECT id, name, code, description, price, cost, quantity, in_stock_date, expiry_date, created_at, updated_at FROM items WHERE code = ?",
		code,
	).Scan(&item.ID, &item.Name, &item.Code, &item.Description, &item.Price, &item.Cost, &item.Quantity, &inStockDate, &expiryDate, &createdAt, &updatedAt)

	if err == sql.ErrNoRows {
		return nil, errors.New("item not found")
	}
	if err != nil {
		return nil, err
	}

	item.InStockDate = inStockDate
	if expiryDate.Valid {
		item.ExpiryDate = &expiryDate.Time
	}
	item.CreatedAt = createdAt
	item.UpdatedAt = updatedAt
	return &item, nil
}
```

**Explanation:**
- **Input:** `db` (Database) - Database interface, `code` (string) - Barcode/QR code
- **Process:**
  1. Queries items table for exact code match
  2. Scans result into Item struct
  3. Handles nullable expiry_date field
  4. Returns error if item not found or database error occurs
- **Output:** Returns Item object or error
- **Called:** When barcode is scanned in transaction mode
- **Critical:** Code field must be unique, enforced by database constraint

**inventory/stock.go - RestockItem Function**

```go
func RestockItem(db Database, itemID int, quantity int, expiryDate *time.Time) error {
	now := time.Now()
	// Update item quantity
	_, err := db.GetDB().Exec(
		"UPDATE items SET quantity = quantity + ?, in_stock_date = ?, updated_at = ? WHERE id = ?",
		quantity, now, now, itemID,
	)
	if err != nil {
		return err
	}

	// Create stock entry
	_, err = db.GetDB().Exec(
		"INSERT INTO item_stock (item_id, quantity, in_stock_date, expiry_date) VALUES (?, ?, ?, ?)",
		itemID, quantity, now, expiryDate,
	)
	return err
}
```

**Explanation:**
- **Input:**
  - `db` (Database) - Database interface
  - `itemID` (int) - ID of item to restock
  - `quantity` (int) - Quantity to add
  - `expiryDate` (*time.Time) - Optional expiry date for this batch
- **Process:**
  1. Updates item's total quantity by adding new quantity (quantity = quantity + ?)
  2. Updates item's in_stock_date to current time
  3. Creates new entry in item_stock table for this batch
  4. Stores batch-specific expiry date if provided
- **Output:** Returns error if operation fails
- **Called:** When user clicks "Restock" button in inventory tab
- **Purpose:** Tracks multiple batches separately for FIFO inventory management

#### Program Output

**Inventory Tab Display:**
- List of all items with columns: Name, Code, Price, Cost, Quantity, In Stock Date, Expiry Date
- Search box at top for filtering items
- Buttons: "Add Item", "Edit Item", "Delete Item", "Restock", "Refresh"

**Add Item Dialog:**
- Form fields: Name, Code, Description, Price, Cost, Quantity, Expiry Date (optional)
- Buttons: "Add" (creates item), "Done" (closes dialog)
- On success: Item appears in list, dialog closes
- On error: Error message displayed, dialog remains open

**Search Functionality:**
- As user types, list filters to show matching items
- Searches both name and code fields
- Partial matches work (e.g., "mil" finds "Milk", "Milk Chocolate")

**Restock Dialog:**
- Fields: Quantity to add, Expiry Date (optional)
- Shows current quantity and calculated new total
- On success: Item quantity gets updated, new batch gets created, and dialog closes

---

### Transaction Processing

#### Flowchart

[transactions.png]

#### Program Code + Explanation

**transactions/transactions.go - CreateTransaction Function**

```go
func CreateTransaction(db Database, userID int, items []models.TransactionItem) (*models.Transaction, error) {
	// Calculate total
	var totalAmount float64
	for _, item := range items {
		totalAmount += item.Price * float64(item.Quantity)
	}

	// Create transaction
	result, err := db.GetDB().Exec(
		"INSERT INTO transactions (user_id, total_amount, created_at) VALUES (?, ?, ?)",
		userID, totalAmount, time.Now(),
	)
	if err != nil {
		return nil, err
	}

	transactionID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	// Create transaction items and update inventory
	for _, item := range items {
		_, err := db.GetDB().Exec(
			"INSERT INTO transaction_items (transaction_id, item_id, quantity, price) VALUES (?, ?, ?, ?)",
			transactionID, item.ItemID, item.Quantity, item.Price,
		)
		if err != nil {
			return nil, err
		}

		// Update item quantity
		var currentQuantity int
		err = db.GetDB().QueryRow("SELECT quantity FROM items WHERE id = ?", item.ItemID).Scan(&currentQuantity)
		if err != nil {
			return nil, err
		}

		newQuantity := currentQuantity - item.Quantity
		if newQuantity < 0 {
			newQuantity = 0
		}

		_, err = db.GetDB().Exec("UPDATE items SET quantity = ?, updated_at = ? WHERE id = ?", newQuantity, time.Now(), item.ItemID)
		if err != nil {
			return nil, err
		}
	}

	return GetTransactionByID(db, int(transactionID))
}
```

**Explanation:**
- **Input:**
  - `db` (Database) - Database interface
  - `userID` (int) - ID of user processing transaction
  - `items` ([]models.TransactionItem) - Slice of items being sold with quantities and prices
- **Process:**
  1. Calculates total amount by summing (price × quantity) for all items
  2. Creates transaction record in transactions table with user ID and total
  3. Gets auto-generated transaction ID
  4. For each item:
     - Creates entry in transaction_items table with price at time of sale
     - Retrieves current item quantity from database
     - Calculates new quantity (current - sold)
     - Prevents negative quantities (sets to 0 if would be negative)
     - Updates item quantity in database
  5. Retrieves and returns complete transaction with all items
- **Output:** Returns Transaction object with all items or error
- **Called:** When user clicks "Complete Transaction" button
- **Important:** Price stored at time of sale preserves historical accuracy even if item price changes later

**barcode/barcode.go - DecodeBarcodeFromImage Function**

```go
func DecodeBarcodeFromImage(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return "", err
	}

	// Convert image to binary bitmap
	bmp, err := gozxing.NewBinaryBitmapFromImage(img)
	if err != nil {
		return "", fmt.Errorf("failed to create binary bitmap: %v", err)
	}

	// Try QR code first
	qrReader := qrcode.NewQRCodeReader()
	result, err := qrReader.Decode(bmp, nil)
	if err == nil && result != nil {
		return result.String(), nil
	}

	// Try Code128
	reader := oned.NewCode128Reader()
	result, err = reader.Decode(bmp, nil)
	if err == nil && result != nil {
		return result.String(), nil
	}

	// Try EAN13, EAN8, UPC-A...
	// Returns first successful decode or error if all fail
}
```

**Explanation:**
- **Input:** `filePath` (string) - Path to image file (JPEG or PNG)
- **Process:**
  1. Opens image file
  2. Decodes image using Go's image package
  3. Converts to gozxing.BinaryBitmap format required by library
  4. Tries multiple barcode formats in order: QR Code, Code128, EAN13, EAN8, UPC-A
  5. Returns first successful decode
- **Output:** Returns decoded barcode string or error if decoding fails
- **Called:** When user uploads barcode image in transaction mode
- **Supported Formats:** QR Code, Code128, EAN13, EAN8, UPC-A for maximum compatibility

#### Program Output

**Transaction Tab Display:**
- Left side: List of all items (searchable)
- Right side: Current transaction items with quantities and prices
- Bottom: Total amount (updates automatically)
- Buttons: "Upload Barcode", "Complete Transaction", "Cancel Transaction"

**Adding Items:**
- User can search and click items, or upload barcode image
- Item appears in transaction list with quantity 1
- Quantity field is editable
- Total updates as items are added or quantities changed

**Batch Selection (if applicable):**
- If item has multiple batches with different expiry dates, dialog appears
- Shows list of batches with quantity and expiry date
- User selects which batch to sell from
- Dialog closes and item added to transaction

**Completing Transaction:**
- On "Complete Transaction" click:
  - Transaction saved to database
  - Inventory quantities updated
  - Transaction list cleared
  - Success message displayed
- Transaction appears in Transaction Log tab

---

## 4. Other Special Features

### GUI Implementation

The GUI is built using the Fyne framework, which provides cross-platform support for Linux and Windows. The interface uses a tab-based design where different tabs are shown based on user permissions.

**Key GUI Features:**

**Permission-Based Tab Display:**
- Inventory tab: Visible if user has CanRead or IsRootAdmin
- Transaction tab: Visible if user has CanTransaction or IsRootAdmin
- Revenue tab: Visible if user has CanRevenue or IsRootAdmin
- Transaction Log: Visible if user has CanTransaction or IsRootAdmin
- User Management: Visible only if IsRootAdmin

**Styled Dialogs**

**Transaction Mode Features:**
- Real-time search filtering
- Click-to-add items from search results
- Automatic total calculation
- Barcode image upload support

### Database Design

**SQLite Choice:**
- Easy backup
- More consistent to handle data through the use of queries 

**Schema Design:**
- Foreign key relationships ensure data consistency (Database is stored in 1NF)
- Indexes on frequently searched fields (barcode id, name) for performance
- Ability to add features such as ALLOW EMPTY values and timestamps

**Migration System**

**Data Integrity**

### Reusability of Code

**Code Structure:**
- Clear separation of concerns

**Helper Functions:**
- `dialog_helpers.go` provides reusable dialog creation functions
- Styled patterns make them visually clearer due to the use of symbols, etc

**Code Organization:**
- Functions are focused and do one thing well
- Easy to add new features without modifying existing code

### Error Handling

**Database Errors:**
- All database operations return errors
- Errors that show up in GUI through pop-ups helps keep users infoormed 
- User-friendly error messages displayed in dialogs

**Validation:**
- Input validation at GUI level (e.g. required fields), and such constraints help catch invalid data
- Uneditable values such as for root admin

**Error Display:**
- `dialog.ShowError()` for database and validation errors 
- Prevents crashes from unexpected conditions

**Specific Error Cases:**
- Duplicate username/code
- Item not found (Includes a pop-up to add item which reduces the overhead of creating a new item)
- Database connection failure: Error shown at startup

---

## 5. Comparisons to Other Alternatives

### Go vs Python

**Why I Chose Go:**

**Type Safety:**
- Go's static typing catches errors at compile time
- Prevents runtime errors from type mismatches
- Makes code easier to debug, as program points out unhandled err variables

**Deployment:**
- Go compiles to single executable file

**Why I Didn't Choose Python:**

**Deployment Complexity:**
- Different Python versions cause compatibility issues
- Go has capabilities to create executables, which ensure smoother user experience

**GUI Framework Issues:**
- Tkinter (built-in) looks outdated, and PYQT global config affects for how all programs look
- Cross-platform deployment is annoying to test out for the look and feel

**Example of Go's Advantages:**
```go
// Go - compile-time type checking
func CreateItem(name string, price float64) (*Item, error) {
    // Compiler catches if wrong types passed
}

// Python - runtime type checking
def create_item(name, price):
    # No type checking until runtime
    # Could pass wrong types and not know until it crashes
    #$ Very important when checking for OKs or other error returns
```

### Fyne vs QT

**Why I Chose Fyne:**

**Simplicity:**
- Fyne has a cleaner, more intuitive API
- Less boilerplate code needed aside from error handling

**Modern Design:**
- Consistent appearance across both Linux and Windows
- Good default styling

**Why I Didn't Choose QT:**

**Error Handling Problems:**
- QT functions often return an "ok" boolean along with the actual result
- This leads to unpredictable behavior and bugs
- Example: `value, ok := qtFunction()` - if you forget to check `ok`, you get wrong results
- This is a huge pain that I would not subject my worst enemy to

**Complexity:**
- More code needed for simple operations

**Language Interdependency:**
- QT is C++ based, requires bindings for other languages
- Potential for bugs in bindings
- Less natural Go code

**Example of QT's Problematic Error Handling:**
```python
// QT - easy to forget to check 'ok'
value, ok = qtWidget.GetValue()
if not ok 
    // Easy to forget this check
    // Leads to bugs when value is invalid
}
// If ok is false, value might be garbage.
// But code continues anyway if check forgotten
// As everyone knows. Garbage in, garbage out.
```

```go
```
// Fyne
value, err := fyneWidget.GetValue()
if err != nil {
    // This is normal in GO, so much harder to forget this check
    // Go compiler also forces error handling, so much easier to debug
    return err
}
```

**Real Problems with QT's "ok" Pattern:**
- You often forget to check the `ok` value. However, code continues executing with invalid data. Bugs are hard to track down, leading to unpredictable program behavior. This is absolute pain. I repeat. It is pain that wastes time and rewards you with nothing. 

**Fyne's Way Better Approach:**
- Uses Go's standard error handling pattern, so errors are explicit and must be handled by compiler.
- More predictable program behavior
- Easier to debug when things go wrong, which they definetely do

### SQLite vs Simpler Text File/JSON Database

**Why I Chose SQLite:**

**Data Integrity:** 
- It ensures data consistency
- Foreign key constraints prevent invalid data
- Prevents data corruption when I tried doing things I shouldn't

**Query Capabilities:**
- SQL queries for complex data retrieval
- JOINs for related data
- Aggregations make it easier for reporting data (SUM, COUNT, etc.)
- Get to use INDEXES and practice database related commands

**Scalability:**
- Professional database features which I'm not qualified to explain

