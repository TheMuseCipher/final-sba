# Inventory Management System - Testing Report

Name: Hussain Sameer
Class: 6B
CSNO: 6217



## Table of Contents

1. [Introduction](#1-introduction)
2. [Unit Testing](#2-unit-testing)
   - [Test Setup](#test-setup)
   - [Authentication Tests](#authentication-tests)
   - [Inventory Tests](#inventory-tests)
   - [Transaction Tests](#transaction-tests)
   - [User Management Tests](#user-management-tests)
   - [Test Results Summary](#test-results-summary)
3. [System Testing](#3-system-testing)
   - [Test Cases](#test-cases)
   - [Test Data](#test-data)
   - [System Test Results](#system-test-results)
4. [User Acceptance Testing](#4-user-acceptance-testing)
   - [Test Scenarios](#test-scenarios)
   - [User Feedback](#user-feedback)
5. [Debugging](#5-debugging)
   - [Issues Found](#issues-found)
   - [How Issues Were Fixed](#how-issues-were-fixed)
6. [Program Evaluation](#6-program-evaluation)
   - [Pros of the Program](#pros-of-the-program)
   - [Cons of the Program](#cons-of-the-program)
7. [Scope Extension](#7-scope-extension)
   - [Implemented Improvement](#implemented-improvement)
   - [Other Suggested Improvements](#other-suggested-improvements)



## 1. Introduction

This report documents the testing done on the Inventory Management System. Testing is important because it helps find bugs before the program is used in real life. I tested the program using three types of tests:

1. **Unit Tests** - Testing small pieces of code by themselves
2. **System Tests** - Testing the whole program works together
3. **User Acceptance Tests** - Getting real users to try the program

I also looked at the pros and cons of my program and described ways it could be improved in the future.

**Discalaimer:** Some of the Unit Tests were written by AI. AI was only used for the purpose of speeding up testing and for the inclusion of comments to clear up readability of the codebase. It was NOT used for generating parts of the code.

## 2. Unit Testing

Unit tests check that each function works correctly on its own. I used Go's built-in testing package to write these tests. Each test creates a fake database in memory so the real database is not affected.

### Test Setup

Each test file creates a mock database that looks like this:

```go
type MockDB struct {
    db *sql.DB
}

func (m *MockDB) GetDB() *sql.DB {
    return m.db
}

func setupTestDB(t *testing.T) *MockDB {
    db, err := sql.Open("sqlite", ":memory:")
    if err != nil {
        t.Fatalf("Failed to open test database: %v", err)
    }
    return &MockDB{db: db}
}
```

**Explanation:**
- `":memory:"` creates a database that only exists in RAM
- This means each test starts fresh with no old data
- Somehow less lines of code

### Authentication Tests

**File:** `auth/auth_test.go`

| Test Name | What It Tests | Expected Result |
|-----------|---------------|-----------------|
| TestHashPassword | Hash function creates a hash | Hash is not empty and not equal to original password |
| TestCheckPasswordHash_Correct | Correct password matches hash | Returns true |
| TestCheckPasswordHash_Wrong | Wrong password does not match hash | Returns false |
| TestHashPassword_DifferentHashes | Same password gives different hashes each time | Two hashes are different (due to salt) |

**Example Test Code:**

```go
func TestCheckPasswordHash_Wrong(t *testing.T) {
    password := "correctpassword"
    wrongPassword := "wrongpassword"

    hash, err := HashPassword(password)
    if err != nil {
        t.Fatalf("HashPassword failed: %v", err)
    }

    result := CheckPasswordHash(wrongPassword, hash)
    if result {
        t.Error("CheckPasswordHash should return false for wrong password")
    }
}
```

**Why These Tests Matter:**
- Passwords must be secure
- Wrong passwords should never work, for obvious reasons
- If hashes were the same, hackers could guess passwords

### Inventory Tests

**File:** `inventory/inventory_test.go`

| Test Name | What It Tests | Expected Result |
|-----------|---------------|-----------------|
| TestCreateItem | Creating a new item | Item is created with correct values |
| TestGetItemByID | Finding item by ID | Correct item is returned |
| TestGetItemByCode | Finding item by barcode | Correct item is returned |
| TestSearchItems | Searching items by name | Only matching items are returned |
| TestUpdateItemQuantity | Changing item quantity | Quantity is updated correctly |
| TestDeleteItem | Deleting an item | Item cannot be found after deletion |
| TestGetLowStockItems | Finding items with low stock | Only items below threshold are returned |

**Example Test Code:**

```go
func TestSearchItems(t *testing.T) {
    mockDB := setupTestDB(t)
    defer mockDB.db.Close()

    CreateItem(mockDB, "Apple Juice", "AJ001", "Fresh apple juice", 3.50, 2.00, 20)
    CreateItem(mockDB, "Orange Juice", "OJ001", "Fresh orange juice", 3.50, 2.00, 15)
    CreateItem(mockDB, "Milk", "MK001", "Fresh milk", 2.00, 1.00, 30)

    results, err := SearchItems(mockDB, "Juice")
    if err != nil {
        t.Fatalf("SearchItems failed: %v", err)
    }

    if len(results) != 2 {
        t.Errorf("Expected 2 results for 'Juice', got %d", len(results))
    }
}
```

**Why These Tests Matter:**
- Inventory is the core of the system
- Wrong quantities can cause problems for the shop, again for obvious reasons
- Search must work so users can find items quickly

### Transaction Tests

**File:** `transactions/transactions_test.go`

| Test Name | What It Tests | Expected Result |
|-----------|---------------|-----------------|
| TestCreateTransaction | Creating a new sale | Transaction is created with correct total |
| TestCreateTransaction_ReducesStock | Sale reduces item quantity | Quantity goes down by amount sold |
| TestGetTransactionByID | Finding a transaction by ID | Correct transaction is returned with all items |

**Example Test Code:**

```go
func TestCreateTransaction_ReducesStock(t *testing.T) {
    mockDB := setupTestDB(t)
    defer mockDB.db.Close()

    items := []models.TransactionItem{
        {ItemID: 1, ItemName: "Apple", Quantity: 10, Price: 1.50},
    }

    _, err := CreateTransaction(mockDB, 1, items)
    if err != nil {
        t.Fatalf("CreateTransaction failed: %v", err)
    }

    var newQuantity int
    err = mockDB.db.QueryRow("SELECT quantity FROM items WHERE id = 1").Scan(&newQuantity)
    if err != nil {
        t.Fatalf("Failed to get item quantity: %v", err)
    }

    if newQuantity != 90 {
        t.Errorf("Expected quantity 90, got %d", newQuantity)
    }
}
```

**Why These Tests Matter:**
- Wrong totals mean wrong money collected
- If stock does not reduce, inventory becomes incorrect
- Transaction history is needed for reports

### User Management Tests

**File:** `users/users_test.go`

| Test Name | What It Tests | Expected Result |
|-----------|---------------|-----------------|
| TestCreateUser | Creating a new user | User is created with correct permissions |
| TestCreateUser_DuplicateUsername | Creating user with existing username | Returns an error |
| TestGetUserByID | Finding user by ID | Correct user is returned |
| TestUpdateUserPermissions | Changing user permissions | Permissions are updated correctly |
| TestDeleteUser | Deleting a user | User cannot be found after deletion |
| TestDeleteUser_CannotDeleteRootAdmin | Trying to delete admin | Returns an error (not allowed) |

**Example Test Code:**

```go
func TestDeleteUser_CannotDeleteRootAdmin(t *testing.T) {
    mockDB := setupTestDB(t)
    defer mockDB.db.Close()

    _, err := mockDB.db.Exec(
        `INSERT INTO users (username, password_hash, is_root_admin)
         VALUES ('admin', 'hash', 1)`,
    )
    if err != nil {
        t.Fatalf("Failed to insert root admin: %v", err)
    }

    err = DeleteUser(mockDB, 1)
    if err == nil {
        t.Error("Should not be able to delete root admin")
    }
}
```

**Why These Tests Matter:**
- Permissions control who can do what
- Duplicate usernames would cause confusion
- Deleting admin would lock everyone out

### Test Results Summary

```
=== RUN   TestHashPassword
--- PASS: TestHashPassword (0.06s)
=== RUN   TestCheckPasswordHash_Correct
--- PASS: TestCheckPasswordHash_Correct (0.12s)
=== RUN   TestCheckPasswordHash_Wrong
--- PASS: TestCheckPasswordHash_Wrong (0.12s)
=== RUN   TestHashPassword_DifferentHashes
--- PASS: TestHashPassword_DifferentHashes (0.24s)
PASS
ok      ims-go/auth     0.544s

=== RUN   TestCreateItem
--- PASS: TestCreateItem (0.00s)
=== RUN   TestGetItemByID
--- PASS: TestGetItemByID (0.00s)
=== RUN   TestGetItemByCode
--- PASS: TestGetItemByCode (0.00s)
=== RUN   TestSearchItems
--- PASS: TestSearchItems (0.00s)
=== RUN   TestUpdateItemQuantity
--- PASS: TestUpdateItemQuantity (0.00s)
=== RUN   TestDeleteItem
--- PASS: TestDeleteItem (0.00s)
=== RUN   TestGetLowStockItems
--- PASS: TestGetLowStockItems (0.00s)
PASS
ok      ims-go/inventory        0.012s

=== RUN   TestCreateTransaction
--- PASS: TestCreateTransaction (0.00s)
=== RUN   TestCreateTransaction_ReducesStock
--- PASS: TestCreateTransaction_ReducesStock (0.00s)
=== RUN   TestGetTransactionByID
--- PASS: TestGetTransactionByID (0.00s)
PASS
ok      ims-go/transactions     0.010s

=== RUN   TestCreateUser
--- PASS: TestCreateUser (0.06s)
=== RUN   TestCreateUser_DuplicateUsername
--- PASS: TestCreateUser_DuplicateUsername (0.06s)
=== RUN   TestGetUserByID
--- PASS: TestGetUserByID (0.06s)
=== RUN   TestUpdateUserPermissions
--- PASS: TestUpdateUserPermissions (0.06s)
=== RUN   TestDeleteUser
--- PASS: TestDeleteUser (0.06s)
=== RUN   TestDeleteUser_CannotDeleteRootAdmin
--- PASS: TestDeleteUser_CannotDeleteRootAdmin (0.00s)
PASS
ok      ims-go/users    0.312s
```

**Summary Table:**

| Package | Tests Run | Tests Passed | Tests Failed |
|---------|-----------|--------------|--------------|
| auth | 4 | 4 | 0 |
| inventory | 7 | 7 | 0 |
| transactions | 3 | 3 | 0 |
| users | 6 | 6 | 0 |
| **Total** | **20** | **20** | **0** |

All 20 unit tests passed successfully.



## 3. System Testing

System testing checks that all parts of the program work together correctly. Unlike unit tests which test one function, system tests follow a whole process from start to finish.

### Test Cases

**Test Case 1: Login Process**

| Step | Action | Expected Result | Actual Result | Pass/Fail |
|------|--------|-----------------|---------------|-----------|
| 1 | Open the program | Login window appears | Login window appears | PASS |
| 2 | Enter wrong password | Error message shows | "Invalid credentials" shown | PASS |
| 3 | Enter correct password | Main window opens | Main window opens | PASS |
| 4 | Check tabs shown | Tabs match user permissions | Correct tabs visible | PASS |

**Test Case 2: Add Item Process**

| Step | Action | Expected Result | Actual Result | Pass/Fail |
|------|--------|-----------------|---------------|-----------|
| 1 | Click "Add Item" | Add item dialog opens | Dialog opens | PASS |
| 2 | Enter item details | Fields accept input | Input accepted | PASS |
| 3 | Click "Add" | Item saved to database | Item appears in list | PASS |
| 4 | Search for new item | Item found | Item found in search | PASS |

**Test Case 3: Complete Transaction Process**

| Step | Action | Expected Result | Actual Result | Pass/Fail |
|------|--------|-----------------|---------------|-----------|
| 1 | Go to Transaction tab | Transaction interface shows | Interface visible | PASS |
| 2 | Search for item | Item appears in list | Item found | PASS |
| 3 | Click item to add | Item added to cart | Item in cart with qty 1 | PASS |
| 4 | Click "Complete Transaction" | Transaction saved | Success message shown | PASS |
| 5 | Check item quantity | Quantity reduced | Quantity reduced by 1 | PASS |

**Test Case 4: User Permission Control**

| Step | Action | Expected Result | Actual Result | Pass/Fail |
|------|--------|-----------------|---------------|-----------|
| 1 | Create user with only Read permission | User created | User created successfully | PASS |
| 2 | Login as new user | Only Inventory tab visible | Only Inventory tab shown | PASS |
| 3 | Update user to have Transaction permission | Permission updated | Permission saved | PASS |
| 4 | Login again as user | Inventory and Transaction tabs visible | Both tabs visible | PASS |

### Test Data

I used the following test data for system testing:

**Test Users:**

| Username | Password | IsRootAdmin | CanRead | CanTransaction | CanRevenue |
|----------|----------|-------------|---------|----------------|------------|
| admin | admin | Yes | Yes | Yes | Yes |
| cashier1 | test123 | No | Yes | Yes | No |
| viewer | test123 | No | Yes | No | No |

**Test Items:**

| Name | Code | Price | Cost | Quantity |
|------|------|-------|------|----------|
| Coca Cola 330ml | CC330 | 8.50 | 5.00 | 50 |
| Bread Loaf | BRD001 | 12.00 | 8.00 | 20 |
| Milk 1L | MLK001 | 15.00 | 10.00 | 30 |
| Chips Original | CHP001 | 6.00 | 3.50 | 100 |
| Water Bottle | WTR001 | 4.00 | 2.00 | 200 |

**Test Transactions:**

| Transaction | Items | Total | Expected Stock Change |
|-------------|-------|-------|----------------------|
| T1 | 2x Coca Cola, 1x Chips | 23.00 | CC: 50 to 48, CHP: 100 to 99 |
| T2 | 3x Milk, 2x Bread | 69.00 | MLK: 30 to 27, BRD: 20 to 18 |
| T3 | 5x Water | 20.00 | WTR: 200 to 195 |

### System Test Results

All system tests passed. The program correctly:

- Authenticates users with correct/incorrect passwords
- Shows correct tabs based on permissions
- Creates, updates, and deletes items
- Processes transactions and updates stock
- Prevents deletion of root admin
- Handles errors gracefully with pop-up messages (Shockingly)

## 4. User Acceptance Testing

User acceptance testing (UAT) involves getting real people to use the program and give feedback. I asked 3 people to test the program:

1. **Tester A**: My classmate (knows no programming)
2. **Tester B**: Family member (no programming knowledge, runs an actual shop)
3. **Tester C**: A friend (knows programming)

### Test Scenarios

Each tester was asked to complete these tasks:

1. Login to the system
2. Add a new item to inventory
3. Search for an item
4. Complete a sale transaction
5. View the transaction history
6. (Admin only) Create a new user

### User Feedback

**Tester A Feedback:**

| Task | Completed? | Comments |
|------|------------|----------|
| Login | Yes | "Easy to use, standard login" |
| Add item | Yes | "Form is clear, all fields make sense" |
| Search | Yes | "Search works fast" |
| Transaction | Yes | "Simple to add items and complete sale" |
| View history | Yes | "Shows all transactions clearly" |

Overall: "Program works well, no major problems found. Nice that it has both light and dark mode support"

**Tester B Feedback:**

| Task | Completed? | Comments |
|------|------------|----------|
| Login | Yes | "Straightforward" |
| Add item | Yes | "Would be nice to have bigger buttons" |
| Search | Yes | "Good that it searches while typing" |
| Transaction | Yes | "Easy to understand" |
| View history | Yes | "Clear layout" |

Overall: "Easy to use even without computer experience. Buttons could be bigger."

**Tester C Feedback:**

| Task | Completed? | Comments |
|------|------------|----------|
| Login | Yes | "Works as expected" |
| Add item | Yes | "Good validation for required fields" |
| Search | Yes | "Nice partial matching" |
| Transaction | Yes | "Total updates automatically, which is a nice feature" |
| View history | Yes | "Would be useful to filter by date" |

Overall: "Good program. Could add date filtering for transactions. Also moree keyboard shortcuts would be nice"

### Summary of User Feedback

**What Users Liked:**
- Simple and easy to use interface
- Search works while typing (real-time)
- Transaction total updates automatically
- Clear error messages
- Good for users without technical knowledge

**Suggestions for Improvement:**
- Make buttons bigger for easier clicking
- Add date filtering for transaction history
- Add keyboard shortcuts for faster use

## 5. Debugging

During development and testing, I found and fixed several issues.

### Issues Found

**Issue 1: Negative Stock Quantity**

- **Problem:** If a transaction sold more items than available, quantity became negative
- **How Found:** During system testing with edge cases
- **Severity:** High - Would cause wrong inventory numbers

**Issue 2: Empty Search Crash**

- **Problem:** Searching with special characters could crash the search
- **How Found:** Tester C tried searching with symbols
- **Severity:** Medium - Program would need restart

**Issue 3: Transaction Total Rounding**

- **Problem:** Total showed too many decimal places (e.g., 23.456789)
- **How Found:** During user acceptance testing
- **Severity:** Low - Visual issue only

### How Issues Were Fixed

**Fix for Issue 1: Negative Stock Quantity**

Added a check in the CreateTransaction function:

```go
// Before fix:
newQuantity := currentQuantity - item.Quantity

// After fix:
newQuantity := currentQuantity - item.Quantity
if newQuantity < 0 {
    newQuantity = 0
}
```

**Explanation:** This ensures quantity never goes below zero. The transaction still goes through, but stock cannot be negative.

**Fix for Issue 2: Empty Search Crash**

The SQL LIKE query handles special characters correctly because I use parameterized queries:

```go
rows, err := db.GetDB().Query(
    "SELECT ... FROM items WHERE name LIKE ? OR code LIKE ?",
    "%"+query+"%", "%"+query+"%",
)
```

**Explanation:** Using `?` placeholders prevents SQL injection and handles special characters safely.

**Fix for Issue 3: Transaction Total Rounding**

The GUI displays money values with 2 decimal places:

```go
priceText := fmt.Sprintf("$%.2f", item.Price)
totalText := fmt.Sprintf("Total: $%.2f", totalAmount)
```

**Explanation:** `%.2f` formats the number to exactly 2 decimal places.

### Debugging Techniques Used

1. **Print Statements** - Added `log.Println()` statements to track variable values
2. **Unit Tests** - Wrote tests to isolate problems
4. **Test with Edge Cases** - Tried unusual inputs like 0, negative numbers, empty strings

## 6. Program Evaluation

### Pros of the Program

**1. Security**
- Passwords are hashed with bcrypt (industry standard)
- Permission system controls who can do what
- Root admin cannot be deleted

**2. Data Integrity**
- SQLite database prevents data corruption
- Foreign keys link related data correctly
- Transactions update all related tables together

**3. User-Friendly**
- Simple interface that anyone can use
- Real-time search saves time
- Automatic calculations reduce errors

**4. Maintainable Code**
- Separate packages for different features
- Functions do one thing each
- Easy to add new features

**5. Cross-Platform**
- Works on Windows and Linux
- Single executable file, no installation needed
- Uses Fyne GUI which looks the same on all systems

### Cons of the Program

**1. No Network Support**
- Only works on one computer
- Cannot share data between computers
- Real shops would need network access (Tester B feedback)

**2. Basic Reporting**
- Revenue report is simple
- No graphs or charts
- Cannot export data to Excel (Tester B feedback)

**3. No Barcode Scanner Support**
- Only supports uploading barcode images
- Could add barcode scanner hardware support if it was commercial software

**4. No Backup System**
- If ims.db is deleted, all data is lost
- Should have automatic backup features (or at least export/import)

**5. Limited Search**
- Cannot search by date range
- Cannot filter items by category
- No advanced search options

## 7. Scope Extension

This section describes how the program could be improved in the future, and shows one major improvement that was actually implemented.

### Implemented Improvement: Functioning Low Stock Alerts

While teesting and evaluvating the code, I realized that the low stock alert feature was created but yet never called. This was leftover from the CLI version of the code.

**What Was Added:**

1. A new function `GetLowStockItems()` in the inventory package
2. Visual warning indicator ("[!] LOW STOCK") in the inventory list
3. A new "Status" column in the inventory table
4. Unit test for the new function

**New Code - inventory/inventory.go:**

```go
func GetLowStockItems(db Database, threshold int) ([]models.Item, error) {
    rows, err := db.GetDB().Query(
        "SELECT id, name, code, description, price, cost, quantity, in_stock_date, expiry_date, created_at, updated_at FROM items WHERE quantity < ? ORDER BY quantity ASC",
        threshold,
    )
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var items []models.Item
    for rows.Next() {
        // scan items...
    }
    return items, nil
}
```

**New Code - gui/gui.go (inventory display):**

```go
const lowStockThreshold = 10

// In the list item display:
if item.Quantity < lowStockThreshold {
    warningLabel.SetText("[!] LOW STOCK")
    warningLabel.TextStyle = fyne.TextStyle{Bold: true}
} else {
    warningLabel.SetText("")
}
```

**Unit Test Added - inventory/inventory_test.go:**

```go
func TestGetLowStockItems(t *testing.T) {
    mockDB := setupTestDB(t)
    defer mockDB.db.Close()

    CreateItem(mockDB, "Low Stock Item", "LOW001", "Only 5 left", 10.00, 5.00, 5)
    CreateItem(mockDB, "Medium Stock Item", "MED001", "Has 15", 10.00, 5.00, 15)
    CreateItem(mockDB, "High Stock Item", "HIGH001", "Has 100", 10.00, 5.00, 100)
    CreateItem(mockDB, "Very Low Stock", "VLOW001", "Only 2 left", 10.00, 5.00, 2)

    lowStockItems, err := GetLowStockItems(mockDB, 10)
    if err != nil {
        t.Fatalf("GetLowStockItems failed: %v", err)
    }

    if len(lowStockItems) != 2 {
        t.Errorf("Expected 2 low stock items, got %d", len(lowStockItems))
    }
}
```

**Test Result:**

```
=== RUN   TestGetLowStockItems
--- PASS: TestGetLowStockItems (0.00s)
```

**How It Works:**

| Item Quantity | What User Sees |
|---------------|----------------|
| 10 or more | Normal display, no warning |
| Less than 10 | Shows "[!] LOW STOCK" in Status column |

**Benefits of This Improvement:**
- Shop owners can quickly see which items need restocking
- Prevents items from running out unexpectedly
- Simple visual indicator is easy to understand
- Does not require any extra user action

### Other Suggested Improvements

**1. Item Categories**

Items could be organized into categories (Food, Drinks, Household, etc.). This makes it easier to find items and create reports.

**How It Could Be Done:**
- Create new `categories` table
- Add `category_id` field to items table
- Add dropdown to select category when adding items
- Add filter by category in inventory view

**2. Barcode Scanner Support**

Real barcode scanners act like keyboards, scanning them and then "typing" the barcode number very fast. The program could detect this fast input and automatically search for the item.

**How It Could Be Done:**
- Detect fast input in the search box
- If input arrives faster than human typing (< 50ms between characters)
- Treat it as barcode scan and search automatically
- Add item to transaction automatically if found

OR

- Just Query and search for item by barcode

**3. Data Export**

Users should be able to export data to CSV or Excel files for further analysis.

**How It Could Be Done:**
- Add "Export" button to inventory and transaction views
- Use Go's `encoding/csv` package to create CSV files
- Let user choose location to save file

```go
func ExportItemsToCSV(items []models.Item, filename string) error {
    file, err := os.Create(filename)
    if err != nil {
        return err
    }
    defer file.Close()

    writer := csv.NewWriter(file)
    defer writer.Flush()

    writer.Write([]string{"Name", "Code", "Price", "Quantity"})

    for _, item := range items {
        writer.Write([]string{
            item.Name,
            item.Code,
            fmt.Sprintf("%.2f", item.Price),
            fmt.Sprintf("%d", item.Quantity),
        })
    }

    return nil
}
```

**4. Automatic Backup**

The program should automatically backup the database to prevent data loss.

**How It Could Be Done:**
- Copy `ims.db` to backup folder on startup
- Keep last 7 days of backups
- Delete old backups automatically (Or users can declare how many backups to keep)
- Add option to restore from backup

### Priority of Remaining Improvements

| Improvement | Difficulty | Impact | Priority |
|-------------|------------|--------|----------|
| Low Stock Alerts | Easy | High | DONE |
| Data Export | Easy | Medium | 1 |
| Automatic Backup | Easy | High | 2 |
| Item Categories | Medium | Medium | 3 |
| Barcode Scanner | Medium | High | 4 |

## Conclusion

The Inventory Management System has been tested thoroughly using unit tests, system tests, and user acceptance tests. All 20 unit tests pass, and system testing shows the program works correctly for all main features.

User feedback was positive overall, with testers finding the program easy to use. Some suggestions for improvement were noted, including bigger buttons and date filtering, as well as additional improvements that would be crucial for a real shop. Therefore, although the program has many strengths including security, data integrity, and a user-friendly interface, there are areas for improvement such as network support and better reporting. These improvements would make the program more useful for real shops.
