FROM node:18-alpine

WORKDIR /app

COPY package.json ./
RUN npm ci
COPY ./app.js ./

CMD ["npm", "start"]
