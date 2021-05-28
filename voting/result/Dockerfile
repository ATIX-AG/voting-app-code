FROM node:14-slim

WORKDIR /app

RUN npm install -g nodemon

COPY package*.json ./

RUN npm ci \
 && npm cache clean --force \
 && mv /app/node_modules /node_modules

COPY . .

ENV PORT 8080

EXPOSE 8080

COPY ./options ./options

CMD ["node", "server.js"]
