# Bdoapi

Bdoapi est un projet composé de deux services principaux :

- **api/** : Un backend Node.js utilisant NestJS pour exposer des endpoints REST.
- **backend/** : Un service Go pour la gestion de la base de données et la logique métier associée.

### Images Docker utilisées

- [nicolaspaillard/bdobackend:latest](https://hub.docker.com/r/nicolaspaillard/bdobackend)
  (backend)
- [nicolaspaillard/bdobackend-dev:latest](https://hub.docker.com/r/nicolaspaillard/bdobackend-dev)
  (backend-dev)
- [nicolaspaillard/bdoapi:latest](https://hub.docker.com/r/nicolaspaillard/bdoapi)
  (api)
- [postgres:17-alpine](https://hub.docker.com/_/postgres) (database)

## Prérequis

- Docker & Docker Compose
- Node.js (pour développement sur `api/`)
- Go (pour développement sur `backend/`)

## Démarrage rapide

## Utilisation avec Docker et Docker Compose

1. **Configurer les variables d'environnement**

   - Copiez le fichier d'exemple :
     ```powershell
     cp example.env .env
     ```
   - Modifiez `.env` selon vos besoins (voir les variables utilisées dans `docker-compose.yml`).

2. **Lancer tous les services**

   ```powershell
   docker-compose up --build
   ```

   Cela démarre :

   - `database` (PostgreSQL, port 5432)
   - `backend` (Go, port 8080)
   - `backend-dev` (Go dev, port 8081, volume monté)
   - `api` (NestJS, port 8082)

3. **Arrêter les services**
   ```powershell
   docker-compose down
   ```

### Accès aux services

> Les ports sont exposés en interne par défaut. Pour accéder à un service depuis l'extérieur, adaptez le `docker-compose.yml` (ajoutez `ports:` si besoin).

## Démarrage manuel

### API (NestJS)

```powershell
cd api
npm install
npm run start:dev
```

### Backend (Go)

```powershell
cd backend
go run main.go
```

## Migration & Base de données

- Les migrations SQL sont dans `backend/sqlc/migrations/`.
- Les requêtes SQL sont dans `backend/sqlc/queries/`.

## Licence

Voir le fichier `LICENSE`.
