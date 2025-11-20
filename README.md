# Bdoapi

Ce dépôt contient l'API front (Node/Nest) et un backend en Go pour la gestion des "pearl items".

## Images Docker publiées

- `bdoapi` : https://hub.docker.com/r/nicolaspaillard/bdoapi
- `bdobackend` : https://hub.docker.com/r/nicolaspaillard/bdobackend
- `bdobackend-dev` : https://hub.docker.com/r/nicolaspaillard/bdobackend-dev

## Résumé

Le projet est composé de deux parties principales :

- `api/` : API en Node.js (NestJS). Contient l'API exposée via Docker (image `bdoapi`).
- `backend/` : Backend en Go (service métier, accès DB). Image : `bdobackend` (et `bdobackend-dev` pour dev).

## Structure importante

- `api/` : code NestJS (Dockerfile, package.json, src/...).
- `backend/` : application Go, migrations SQL et configuration `sqlc`.

## Configuration requise

Le projet attend une base PostgreSQL. Vous pouvez démarrer l'ensemble avec Docker Compose. Les variables d'environnement principales sont :

- `DB_HOST` : hôte PostgreSQL
- `DB_PORT` : port (ex : 5432)
- `DB_NAME` : nom de la base
- `DB_USER` : utilisateur
- `DB_PASSWORD` : mot de passe

Exemple de fichier `.env` (ne pas committer les secrets) :

```
DB_HOST=database
DB_PORT=5432
DB_NAME=bdo
DB_USER=bdo_user
DB_PASSWORD=secret
```

## Démarrage rapide (Docker Compose)

Le dépôt contient un `docker-compose.yml` prêt à l'emploi. Pour lancer les services en arrière-plan (PowerShell) :

```powershell
docker compose up -d
```

Vérifier les logs (exemple pour l'API) :

```powershell
docker compose logs -f api
```

## Services fournis par le compose

- `database` : PostgreSQL 17 (volume persistant pour les données).
- `backend` : backend Go (image `nicolaspaillard/bdobackend:latest`).
- `backend-dev` : version de développement (mount du code local pour hot-reload), image `nicolaspaillard/bdobackend-dev:latest`.
- `api` : front/API NestJS (image `nicolaspaillard/bdoapi:latest`).

## Ports exposés (par défaut dans le compose)

- backend : 8080
- backend-dev : 8081
- api : 8082

## Développement local

- API (NestJS) : se placer dans le dossier `api/`, installer les dépendances (`npm install`) puis lancer en dev.
- Backend (Go) : se placer dans `backend/`, construire et lancer l'application localement (`go run .` ou `go build`).

## Construire les images localement

Exemples (depuis la racine) :

```powershell
docker build -t nicolaspaillard/bdobackend:local -f backend/Dockerfile backend
docker build -t nicolaspaillard/bdoapi:local -f api/Dockerfile api
```

## Notes & bonnes pratiques

- Utiliser un fichier `.env` pour stocker les variables sensibles.
- Les migrations SQL sont dans `backend/migrations/` et les requêtes `sqlc` dans `backend/queries`.

## Contribuer

Les contributions sont les bienvenues : ouvrez une issue ou un pull request. Précisez l'environnement et la version de Docker utilisée.

## Licence

Ce projet n'a pas de licence renseignée dans le dépôt. Ajoutez-en une si vous souhaitez autoriser la réutilisation.
