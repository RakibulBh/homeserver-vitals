# Home Server Vitals

A modern, real-time dashboard for monitoring your home server's vital statistics. This application uses a Go backend to collect system metrics and a Next.js frontend to display them in a responsive, professional dashboard.

![Dashboard Screenshot](docs/dashboard-screenshot.png)

## Features

- **Real-time monitoring** via Server-Sent Events (SSE)
- **Responsive UI** that works on desktop and mobile devices
- **Professional dark theme** with clean visualization
- **Low resource usage** on the host system
- **Cross-platform** compatibility (Linux, macOS, Windows)

## Architecture

- **Backend**: Go server with [gopsutil](https://github.com/shirou/gopsutil) for system metrics collection
- **Frontend**: Next.js with React and Tailwind CSS for responsive UI

## System Metrics

The dashboard displays the following metrics in real-time:

- **CPU Usage**: Overall usage percentage with historical chart
- **Memory**: Total, used, and usage percentage
- **Disk**: Storage usage for root partition
- **Network**: Upload and download statistics
- **System Load**: 1, 5, and 15-minute load averages
- **Temperature**: System temperature sensors
- **System Info**: Uptime, processes count, hostname, platform details
- **Go Runtime**: Goroutines and memory allocation metrics

## Installation

### Prerequisites

- Go 1.18 or higher
- Node.js 18.x or higher
- npm 8.x or higher

### Backend Setup

1. Clone the repository:

   ```bash
   git clone https://github.com/yourusername/homeserver-vitals.git
   cd homeserver-vitals
   ```

2. Install Go dependencies:

   ```bash
   go mod tidy
   ```

3. Build and run the backend:

   ```bash
   go run cmd/api/*.go
   ```

   The server will start on port 2000 by default.

### Frontend Setup

1. Navigate to the web directory:

   ```bash
   cd web
   ```

2. Install dependencies:

   ```bash
   npm install
   ```

3. Run the development server:

   ```bash
   npm run dev
   ```

   The frontend will be available at http://localhost:3000.

## Production Deployment

### Backend

1. Build the Go binary:

   ```bash
   go build -o homeserver-vitals ./cmd/api
   ```

2. Run the binary:
   ```bash
   ./homeserver-vitals
   ```

### Frontend

1. Build the production version:

   ```bash
   cd web
   npm run build
   ```

2. Start the production server:

   ```bash
   npm start
   ```

   Alternatively, you can use a process manager like PM2:

   ```bash
   npm install -g pm2
   pm2 start npm --name "homeserver-vitals" -- start
   ```

## Configuration

### Backend Configuration

The backend can be configured through environment variables:

- `PORT`: Server port (default: 2000)
- `ENV`: Environment ("dev" or "prod", default: "dev")
- `FRONTEND_URL`: Allowed CORS origin (default: "http://localhost:3000")

Example:

```bash
PORT=8080 ENV=prod FRONTEND_URL=https://yourdomain.com ./homeserver-vitals
```

### Frontend Configuration

The frontend API URL can be modified in `.env.local`:

```
NEXT_PUBLIC_API_URL=http://your-server-ip:2000
```

## API Endpoints

- `GET /health`: Health check endpoint
- `GET /sse`: Server-Sent Events stream for real-time metrics
- `GET /vitals`: Current system vitals (single request)

## Running as a Service

### Systemd (Linux)

Create a systemd service file:

```bash
sudo nano /etc/systemd/system/homeserver-vitals.service
```

Add the following content:

```
[Unit]
Description=Home Server Vitals Monitoring
After=network.target

[Service]
ExecStart=/path/to/homeserver-vitals
WorkingDirectory=/path/to/homeserver-vitals
User=yourusername
Restart=always
RestartSec=5
Environment=PORT=2000
Environment=ENV=prod
Environment=FRONTEND_URL=http://localhost:3000

[Install]
WantedBy=multi-user.target
```

Enable and start the service:

```bash
sudo systemctl enable homeserver-vitals
sudo systemctl start homeserver-vitals
```

## Troubleshooting

### Common Issues

1. **Connection refused errors**: Ensure the backend is running and the port is accessible.
2. **CORS errors**: Check that the FRONTEND_URL environment variable matches your frontend origin.
3. **Missing temperature data**: Some systems may not expose temperature sensors.
4. **Permissions errors**: The application may need elevated permissions to access certain system metrics.

### Logs

- Backend logs are printed to stdout/stderr.
- Check system logs with `journalctl -u homeserver-vitals` if running as a systemd service.

## License

MIT License

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
