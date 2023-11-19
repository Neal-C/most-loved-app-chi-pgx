 # A Go Playground
 ## Calm down, this is not meant for production. it is not "prod ready"
 
 ### live

- the environment variable CHI_PGX_ENV must be set to a non-empty string when live

### Technical & design choices

Every design and tech decision were taken based in order of importance by: what I want to experiment/discover, what I like, what's been recommended to me by the Go community

### Choices

to gain time, and give a minimum viable product, I overlooked several things:

- database initialization
- security (errors returned to the client, etc...)
- proper typing 
- error and exception handling
- clean architecture patterns
- validating and/or escaping user input

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
