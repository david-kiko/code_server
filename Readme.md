docker-compose build
docker save container-platform-backend:1.0.0 | gzip > /d/work/image/container-platform-backend-1.0.0.tgz
docker save container-platform-frontend:1.0.0 | gzip > /d/work/image/container-platform-frontend-1.0.0.tgz