Project Context: Money Planning Web App
1. Project Overview
This project is a financial planning application that converts static spreadsheet logic into a dynamic web platform. It allows users to project their financial future (12-24 months) based on income, expenses, debts, and savings goals. Goal: Provide a dashboard where users can simulate financial scenarios, track debt repayment (Paylater/Loans), and calculate automated savings to reach specific goals.

2. Tech Stack
Language: Golang (Go 1.20+)

Framework: Fiber v2 (github.com/gofiber/fiber/v2)

Database: MongoDB (Official Driver go.mongodb.org/mongo-driver)

Architecture: Clean Architecture (Domain-Driven Design approach)

3. Directory Structure & Responsibilities
The project follows a strict Clean Architecture pattern:

cmd/web/main.go: Entry point. Initializes DB, Repositories, Usecases, and starts the Fiber server.

config/: Configuration (Env variables, MongoDB connection setup).

internal/model/: Struct definitions for Request/Response and MongoDB Documents (Data Layer).

internal/delivery/http/: Handlers/Controllers and Routing logic. No business logic here.

internal/usecase/: Core Business Logic. Contains the calculation engine (Projection algorithms).

internal/repository/: Database interactions (CRUD operations only).

4. Key Business Logic (The "Calculation Engine")
The core feature is the Monthly Projection Loop. When a user requests a calculation:

Iterate Months: Loop from Start Month to N months ahead.

Income: Fixed monthly Salary + Side Income.

Expenses (Categories):

Fixed/Basic: Recurring expenses (e.g., Food, Transport).

Lifestyle: Optional recurring (e.g., Gym).

Debts (Complex Logic):

Flat Loan: Fixed monthly deduction until debt is 0.

Percentage Loan (Paylater): Calculate interest (RemainingBalance * AnnualRate / 12). Add interest to expense, deduct payment from principal.

Zakat (Optional): If enabled, calculate 2.5% of Income or Savings.

Disposable Income: Income - (Expenses + Debts + Zakat).

Savings Logic:

Remaining Disposable Income is added to Savings Balance.

Track if Savings Balance meets the Target Goal.

Carry Over: The Savings Balance and Remaining Debt of Month i become the starting values for Month i+1.

5. Database Schema (MongoDB)
We use a document-based approach for flexibility.

Collection plans: Stores the entire projection result.

Contains an array projections where each item represents a specific month's data (snapshot).

This avoids re-calculating everything on read; we calculate on write/update.

Collection categories: Master data for user's expense types (Name, Type: DEBT/EXPENSE, InterestRate, IsFixed).

6. Coding Conventions
Error Handling: Return explicit errors from Repository -> Usecase -> Delivery. Use JSON response for errors.

Dependency Injection: Inject Repository into Usecase, and Usecase into Handler in main.go.

Routing: All routes are defined in internal/delivery/http/route.go.