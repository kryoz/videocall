FROM node:22

WORKDIR /app
RUN set -ex

COPY package*.json ./
RUN --mount=type=cache,target=/app/node_modules \
    npm install --no-audit --prefer-offline
COPY .. .
RUN npm run build
