FROM node:26.12.1-alpine AS build
WORKDIR /app
# Playwright is a dev-only dependency for the E2E demo; the browser binaries
# are installed on the host, not inside the image.
ENV PLAYWRIGHT_SKIP_BROWSER_DOWNLOAD=1
COPY package.json pnpm-lock.yaml* ./
RUN corepack enable && pnpm install --no-frozen-lockfile
COPY . .
RUN pnpm build

FROM nginx:alpine AS runtime
COPY --from=build /app/dist /usr/share/nginx/html
EXPOSE 80
