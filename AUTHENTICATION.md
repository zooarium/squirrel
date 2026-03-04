# Authentication & SSO Guide

This guide details the authentication architecture, shared library usage, and integration steps for the **Keeper** (Identity Provider) and **Squirrel** (Resource Service) microservices.

## Architecture

The system uses a **Centralized JWT Authentication** model.

1.  **Keeper (Identity Provider)**:
    *   Manages user registration and login.
    *   Issues JWT (JSON Web Tokens) signed with a `JWT_SECRET`.
    *   Port: `:8080`

2.  **Squirrel (Resource Service)**:
    *   Manages expense categories and transactions.
    *   **Does not** manage users.
    *   Validates the JWT issued by Keeper using the same shared `JWT_SECRET`.
    *   Extracts the `user_id` from the token context to enforce data isolation (Multi-tenancy).
    *   Port: `:8081`

## Shared Library (`pkg/auth`)

Keeper exposes its authentication logic as a public Go package: `keeper/pkg/auth`.

*   **`JWTManager`**: Handles token generation (Keeper) and validation (Squirrel).
*   **`Middleware`**: HTTP middleware that intercepts requests, validates the Bearer token, and injects `UserClaims` into the request context.
*   **`GetClaimsFromContext`**: Helper function to retrieve the authenticated `user_id` inside handlers.

## Configuration

Both services must share the **exact same** `JWT_SECRET` and `JWT_EXPIRY` configuration.

### Keeper (`keeper/config/config.yaml`)
```yaml
SERVER:
  ADDR: ":8080"
AUTH:
  JWT_SECRET: "your-shared-secret-key-change-me"
  JWT_EXPIRY: 24h
```

### Squirrel (`squirrel/config/config.yaml`)
```yaml
SERVER:
  ADDR: ":8081"
AUTH:
  JWT_SECRET: "your-shared-secret-key-change-me"
  JWT_EXPIRY: 24h
```

> **Security Note**: In production, `JWT_SECRET` should be a long, random string injected via environment variables (e.g., `SQUIRREL_AUTH_JWT_SECRET`).

## Testing Flow (CURL)

Follow these steps to test the entire authentication flow.

### Prerequisites
Ensure both services are running:
```bash
# Terminal 1
cd keeper && make up

# Terminal 2
cd squirrel && make up
```

### 1. Register a User (Keeper)
Create a new user account.

```bash
curl -X POST http://localhost:8080/users 
  -H "Content-Type: application/json" 
  -d '{
    "firstname": "John",
    "lastname": "Doe",
    "email": "john.doe@example.com",
    "password": "securepassword123"
  }'
```

### 2. Login (Keeper) -> Get Token
Authenticate to receive a JWT access token.

```bash
# Login and store the token in a variable
TOKEN=$(curl -s -X POST http://localhost:8080/users/auth 
  -H "Content-Type: application/json" 
  -d '{
    "email": "john.doe@example.com",
    "password": "securepassword123"
  }' | jq -r '.data.token')

echo "Token: $TOKEN"
```

### 3. Access Protected Resource (Squirrel)
Use the token to create a category in Squirrel. The `user_id` is automatically extracted from the token.

```bash
curl -X POST http://localhost:8081/categories 
  -H "Authorization: Bearer $TOKEN" 
  -H "Content-Type: application/json" 
  -d '{
    "name": "Groceries",
    "status": 1
  }'
```

### 4. Verify Data Isolation (Squirrel)
List categories. You should only see the ones created by your user.

```bash
curl -X GET http://localhost:8081/categories 
  -H "Authorization: Bearer $TOKEN"
```

### 5. Test Unauthorized Access
Try to access Squirrel without a token or with an invalid one.

```bash
# No Token
curl -v http://localhost:8081/categories

# Invalid Token
curl -v -H "Authorization: Bearer invalid-token" http://localhost:8081/categories
```

## Integration Guide

### Adding a New Go Service
1.  **Update `go.mod`**:
    ```go
    require keeper v0.0.0
    replace keeper => ../keeper
    ```
2.  **Import Package**:
    ```go
    import "keeper/pkg/auth"
    ```
3.  **Initialize & Middleware**:
    ```go
    // In main.go
    jwtManager := auth.NewJWTManager(cfg.Auth.JWTSecret, cfg.Auth.JWTExpiry)
    
    // In router setup
    r.Use(auth.Middleware(jwtManager))
    ```
4.  **Access User ID**:
    ```go
    claims, ok := auth.GetClaimsFromContext(r.Context())
    userID := claims.UserID
    ```

### Adding a Non-Go Service (e.g., Node.js)
1.  **Install JWT Library**: `npm install jsonwebtoken`
2.  **Configure Secret**: Set `JWT_SECRET` to match Keeper's secret.
3.  **Middleware Example**:
    ```javascript
    const jwt = require('jsonwebtoken');

    const authenticate = (req, res, next) => {
      const authHeader = req.headers['authorization'];
      const token = authHeader && authHeader.split(' ')[1];

      if (!token) return res.sendStatus(401);

      jwt.verify(token, process.env.JWT_SECRET, (err, user) => {
        if (err) return res.sendStatus(403);
        req.user = user; // user.user_id is available here
        next();
      });
    };
    ```
