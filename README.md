# TextJar

TextJar is a simple, lightweight self-hosted pastebin application built with **Go** and the **notectl** rich-text editor. It supports versioning, image uploads, and clean formatting, making it ideal for internal notes and snippets.

> [!WARNING]
> **No Authentication:** This application does not currently have a login system or authentication. It is designed for use on a **Local Area Network (LAN)** or via a VPN. **Do not expose this application directly to the public internet.**

## Features

- **Rich Text Editor:** Powered by `notectl` with support for formatting, tables, and images.
- **Versioning:** Every edit creates a new version, allowing you to track history.
- **Relative Timestamps:** See exactly when a paste was created or last updated (e.g., "5 minutes ago").
- **Automatic Slugs:** Generates unique, memorable URLs (e.g., `apple-banana-cherry`).
- **Mobile Friendly:** Clean, responsive design that works well on all devices.
- **Pagination:** Handles large lists of pastes efficiently.

## Getting Started (LAN Deployment)

To get TextJar running on your local network:

### Option 1: Using Docker (Recommended)

This is the fastest way to deploy.

1.  **Clone the repository:**
    ```bash
    git clone https://github.com/numbertheory/textjar.git
    cd textjar
    ```

2.  **Build and Run:**
    ```bash
    docker-compose up --build -d
    ```

3.  **Access on LAN:**
    Find your machine's IP address (e.g., `192.168.1.50`) and open it in your browser on port **3636**:
    `http://192.168.1.50:3636`

### Option 2: Running Locally (Bare Metal)

Ensure you have **Go 1.26+** and **Node.js 20+** installed.

1.  **Install dependencies and build the editor:**
    ```bash
    npm install
    npm run build
    ```

2.  **Run the Go application:**
    ```bash
    go run main.go
    ```

3.  **Access:**
    Open `http://localhost:3636` in your browser.

## Browser Compatibility Note

When accessing TextJar via an IP address over HTTP, some modern browser security features (like the Crypto API) are normally disabled. TextJar includes a built-in polyfill to ensure the editor remains fully functional in these "non-secure" local network contexts.

## Technical Stack

- **Backend:** Go with [Gin Gonic](https://github.com/gin-gonic/gin) and [GORM](https://gorm.io/).
- **Database:** SQLite (stored in `db/textjar.db`).
- **Frontend:** Vanilla JS/CSS with [notectl](https://samyssmile.github.io/notectl/).
- **JS Bundler:** [esbuild](https://esbuild.github.io/).
- **Containerization:** Docker & Docker Compose.

## Development

If you make changes to the JavaScript source in `static/js/main.js`, you must rebuild the bundle:
```bash
npm run build
```

Or, if using Docker, rebuild the container:
```bash
docker-compose up --build
```
