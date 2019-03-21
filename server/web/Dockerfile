FROM node:11.12.0-alpine as builder
WORKDIR /app
COPY package*.json ./
RUN npm install
COPY . .
RUN npm run build

FROM builder as development
RUN npm run serve

FROM nginx:1.13.12-alpine as production
COPY --from=builder /app/dist /usr/share/nginx/html
CMD ["nginx", "-g", "daemon off;"]
