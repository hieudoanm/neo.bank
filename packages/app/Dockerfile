FROM node:26.13.1-alpine AS build
WORKDIR /app
# Playwright is a dev-only dependency for the E2E demo; the browser binaries
# are installed on the host, not inside the image.
ENV PLAYWRIGHT_SKIP_BROWSER_DOWNLOAD=1
# Optional absolute API origin baked into the bundle. Empty means same-origin,
# which the bundled nginx proxies to the api-gateway service.
ARG VITE_API_BASE=""
ENV VITE_API_BASE=${VITE_API_BASE}
COPY package.json pnpm-lock.yaml* ./
RUN corepack enable && pnpm install --no-frozen-lockfile
COPY . .
RUN pnpm build

FROM nginx:alpine AS runtime
COPY --from=build /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
