# Bdoapi

[![CI status](https://github.com/nicolaspaillard/Bdoapi/actions/workflows/docker-publish.yml/badge.svg?branch=main)](https://github.com/nicolaspaillard/Bdoapi/actions/workflows/docker-publish.yml)
[![Docker Pulls - bdoapi](https://img.shields.io/docker/pulls/nicolaspaillard/bdoapi?style=flat-square)](https://hub.docker.com/r/nicolaspaillard/bdoapi)
[![Docker Pulls - bdobackend](https://img.shields.io/docker/pulls/nicolaspaillard/bdobackend?style=flat-square)](https://hub.docker.com/r/nicolaspaillard/bdobackend)
[![Docker Pulls - bdobackend-dev](https://img.shields.io/docker/pulls/nicolaspaillard/bdobackend-dev?style=flat-square)](https://hub.docker.com/r/nicolaspaillard/bdobackend-dev)

Bdoapi est un projet composé de deux services principaux :

- **api/** : Un backend Node.js utilisant NestJS pour exposer des endpoints REST.
- **backend/** : Un service Go pour la gestion de la base de données et la logique métier associée.

### Images Docker utilisées

Ce projet publie et utilise les images Docker suivantes :

- [nicolaspaillard/bdobackend:latest](https://hub.docker.com/r/nicolaspaillard/bdobackend) — backend (Go).
- [nicolaspaillard/bdobackend-dev:latest](https://hub.docker.com/r/nicolaspaillard/bdobackend-dev) — image de développement pour le backend.
- [nicolaspaillard/bdoapi:latest](https://hub.docker.com/r/nicolaspaillard/bdoapi) — API (NestJS).
- [postgres:17-alpine](https://hub.docker.com/_/postgres) — image officielle PostgreSQL utilisée pour la base de données.

## Intégration continue (CI)

Ce projet utilise **GitHub Actions** pour l'intégration continue. À chaque push ou pull request sur la branche principale, le workflow défini dans `.github/workflows/docker-publish.yml` est déclenché. Il automatise :

- La construction des images Docker pour les services `api` (NestJS) et `backend` (Go)
- Les tests de build
- La publication des images sur Docker Hub si le build est réussi

L'état de la CI est visible via le badge en haut de ce fichier. Pour plus de détails, consultez le fichier de workflow ou la page [Actions du dépôt](https://github.com/nicolaspaillard/Bdoapi/actions/workflows/docker-publish.yml).

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

## Liens de test

Utilisez ces URLs pour tester rapidement les endpoints de l'API :

- https://bdoapi.nicolaspaillard.fr/pearlitems?date=2025-11-20%2002:00:00
- https://bdoapi-nest.nicolaspaillard.fr/pearl_items?date=2025-11-20%2002:00:00

## Licence

Voir le fichier `LICENSE`.
