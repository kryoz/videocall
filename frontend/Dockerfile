FROM node:22

WORKDIR /app
COPY package*.json ./
RUN --mount=type=cache,target=/app/node_modules \
    npm install
COPY . .
RUN npm run build
