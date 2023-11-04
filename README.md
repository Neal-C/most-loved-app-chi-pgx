 ### Production

- the environment variable CHI_PGX_ENV must be set to a non-empty string on production

### Tech Stack

- Go (chi)

- PostgreSQL

- Pgx

- Docker & Docker Compose

- joho/godotenv
- google/uuid
- testify/require


```sh
go get -u github.com/go-chi/chi/v5
go get -u github.com/go-chi/cors
go get -u github.com/go-chi/httprate
go get -u github.com/jackc/pgx/v5
go get -u github.com/google/uuid
go get -u  github.com/joho/godotenv
go get -u github.com/stretchr/testify/require
```
