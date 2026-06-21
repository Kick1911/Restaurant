Build a Go RESTful project for a resturant

Brief:
Frogo Baggins, a hobbit from the Shire, has a great idea. He wants to build a
restaurant that serves traditional dishes from the world of Middle Earth. The
restaurant will be called "The Orc Shack" and will have a cozy atmosphere.
Frogo has hired you to build the website for his restaurant. As payment, he has
offered you either a chest of gold or a ring. Choose wisely.

Requirements:
- An API user must be able to: - Create, View, List, Update, and Delete dishes. - Dishes must have a name, description, price, and image.
- Customers must be able to take the following actions: - Search, View, and Rate dishes
- Multi-tenant SaaS support
- Add user, permission, and authentication support.
- Users must be able to register and login.
- All functionality of the API must require a logged in user (except Registration)
- At a minimum, the system should support password based authentication.
- Users must have a name and email address and password.
- Add validation to the data entities in the API.
- An Evil Orc is attempting to brute force passwords for known email addresses. Add functionality to defend against this. (You can use any methodology that you deem suitable)
- To prevent abuse, add rate-limiting per logged in customer.

- The Core HTTP & Routing Layer: github.com/go-chi/chi
- Database & Data Persistence: github.com/jackc/pgx/v5, github.com/sqlc-dev/sqlc, github.com/jmoiron/sqlx, golang-migrate
- Configuration & Validation: github.com/ilyakaznacheev/cleanenv, github.com/go-playground/validator/v10
- Observability (Logging, Metrics & Tracing): slog, github.com/prometheus/client_golang
- Security & Authentication: golang-jwt/jwt/v5, golang.org/x/crypto/bcrypt, golang.org/x/time/rate
- Directory Structure: golang-standards/project-layout

