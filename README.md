# Slot Game API 1.0

Slot Game API represents simple REST API that mimics the core functionalities of a slot game system.
The API allows users to do the following:

- Register and log in.
- Deposit and withdraw credits.
- Spin the slot machine.
- Check game history.
- Retrieve the current credit balance.

## Features

### 1. User Management

- **Registration**: Register a user with login and password.
- **Login**: Authenticate user by login and password, return JWT.
- **Get User Profile**: Return basic user info and credit balance.

### 2. Wallet Management

- **Deposit Credits**: Add credits to a user's balance.  
  **Outputs**: resulting balance.
- **Withdraw Credits**: Subtract credits from a user’s balance.  
  **Outputs**: resulting balance.

### 3. Game Logic: Slot Machine Spin

- **Spin**: A user can spin the slot machine by betting a specific amount of credits. If they win, they are awarded credits based on a simple algorithm.  
  The spin generates 3 symbols, each represented by a random number from 1 to 9. Example combinations: `[1,2,3]`, `[7,7,7]`, `[3,5,9]`.

  **Inputs**:
    - Bet amount (must be within the user’s available balance).

  **Outputs**:
    - Updated credit balance.
    - Spin result.

- **Payout Calculation**: Payout is based on the following logic:
    - Three identical symbols: 10x bet (e.g., `[7,7,7]`).
    - Two identical symbols: 2x bet (e.g., `[7,7,3]`).
    - No identical symbols: loss of bet (e.g., `[7,8,9]`).

### 4. Game History

- **List Game History**: Shows a list of all spins, including the result, bet, and any winnings for the user.

## Endpoints Overview

### 1. User Management

- `POST /api/register`: Register a new user.
- `POST /api/login`: Log in and receive an authentication token.
- `GET /api/profile`: Retrieve the user profile and credit balance (authorization required).

### 2. Wallet Management

- `POST /api/wallet/deposit`: Deposit credits to the user's balance (authorization required).
- `POST /api/wallet/withdraw`: Withdraw credits from the user's balance (authorization required).

### 3. Game Logic

- `POST /api/slot/spin`: Spin the slot machine, place a bet, and get the result (authorization required).

### 4. Game History

- `GET /api/slot/history`: Retrieve a list of the user's past spins (authorization required).

## Installation and Use

In the project root folder, type:

```bash
docker compose up -d --build
```

Swagger is accessible by the following URL:
```
http://localhost:8080/swagger/index.html
```

## Known Issues and Limitations

1. **Payout Calculation**: The payout calculation is currently hardcoded. A possible improvement would be to implement it as a table to allow dynamic changes to the payout calculation.
2. **Testing**: Tests are sparse. Additional tests should be written, utilizing mocks for services and potentially including a few end-to-end (E2E) tests with DevContainers.
3. **Database Schema Management**: The database schema is applied through a workaround. A dedicated database migration system (e.g., `golang-migrate`) should be used, and the approach to launching migrations should be adapted to the deployment environment (e.g., in Kubernetes, a sidecar container or a job could handle migrations).
