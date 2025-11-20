Images :
- https://hub.docker.com/r/nicolaspaillard/bdoapi
- https://hub.docker.com/r/nicolaspaillard/bdobackend
- https://hub.docker.com/r/nicolaspaillard/bdobackend-dev



Install using docker compose :

```
services:
  database:
    image: postgres:17-alpine
    restart: always
    environment:
      POSTGRES_DB: ${DB_NAME}
      POSTGRES_PASSWORD: ${DB_PASSWORD}
      POSTGRES_USER: ${DB_USER}
    expose:
      - 5432
    volumes:
      - database:/var/lib/postgresql/data

  backend:
    image: nicolaspaillard/bdobackend:latest
    restart: always
    environment:
      DB_HOST: ${DB_HOST}
      DB_PORT: ${DB_PORT}
      DB_NAME: ${DB_NAME}
      DB_PASSWORD: ${DB_PASSWORD}
      DB_USER: ${DB_USER}
      PORT: 8080
    expose:
      - 8080
    depends_on:
      - database

  backend-dev:
    image: nicolaspaillard/bdobackend-dev:latest
    restart: always
    environment:
      DB_HOST: ${DB_HOST}
      DB_PORT: ${DB_PORT}
      DB_NAME: ${DB_NAME}
      DB_PASSWORD: ${DB_PASSWORD}
      DB_USER: ${DB_USER}
      PORT: 8081
    expose:
      - 8081
    volumes:
      - backend-dev:/app
    depends_on:
      - database

  api:
    image: nicolaspaillard/bdoapi:latest
    restart: always
    environment:
      DB_HOST: ${DB_HOST}
      DB_PORT: ${DB_PORT}
      DB_NAME: ${DB_NAME}
      DB_PASSWORD: ${DB_PASSWORD}
      DB_USER: ${DB_USER}
      PORT: 8082
      NODE_ENV: production
    expose:
      - 8082
    depends_on:
      - database

volumes:
  database:
  backend-dev:
```
