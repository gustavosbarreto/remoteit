FROM node:9-alpine
RUN npm install -g json-server
ADD . /src
WORKDIR /src
ENTRYPOINT ["/src/entrypoint.sh"]
