# Stage 1: Build Angular app
FROM node:24-alpine AS builder

WORKDIR /app

COPY package.json package-lock.json* ./
RUN npm ci

COPY . .

RUN npx ng build --configuration=production

# Stage 2: Serve with Nginx
FROM nginxinc/nginx-unprivileged:1.27-alpine

USER root
RUN apk add --no-cache curl
USER nginx

# Remove default nginx config
RUN rm /etc/nginx/conf.d/default.conf

# Copy our custom nginx config (server block for conf.d)
COPY --chown=nginx:nginx nginx.conf /etc/nginx/conf.d/default.conf

# Copy built files from build stage
COPY --from=builder --chown=nginx:nginx /app/dist/frontend/browser /usr/share/nginx/html

EXPOSE 8080

HEALTHCHECK --interval=10s --timeout=5s --retries=5 --start-period=10s \
  CMD curl -sf http://localhost:8080/health | grep -q ok

CMD ["nginx", "-g", "daemon off;"]
