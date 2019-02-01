FROM node:11.9.0-alpine

RUN npm install -g http-server python2

WORKDIR /app

COPY package*.json ./

RUN npm install

COPY . .

RUN npm run build

EXPOSE 8080
CMD [ "http-server", "-c-1", "dist" ]
