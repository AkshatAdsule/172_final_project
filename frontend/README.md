# B³ Ride Tracking Frontend

A React-based frontend application for tracking and visualizing GPS rides. This application connects to a backend service that processes MQTT GPS updates and provides real-time ride tracking with historical data.

## Features

- **Live GPS Tracking**: Real-time location updates via WebSocket connection
- **Ride Management**: Automatic ride detection, start/end tracking
- **Historical Rides**: Browse and view detailed information about past rides
- **Interactive Maps**: Google Maps integration for route visualization
- **Responsive Design**: Modern UI with sidebar navigation

## Architecture

### Key Components

- **Real-time WebSocket Communication**: Receives live GPS updates and ride status changes
- **REST API Integration**: Fetches historical ride data from the backend
- **Interactive Maps**: Displays GPS tracks using Google Maps
- **Type-safe Development**: Full TypeScript support with proper type definitions

### WebSocket Message Types

The frontend handles these WebSocket message types from the backend:

- `current_location`: Sent on every valid GPS update.
- `ride_started`: A new ride has begun.
- `ride_position_added`: New GPS point added to current ride.
- `ride_ended`: The current ride has ended.

### API Endpoints

- `GET /api/rides`: Fetch ride summaries with pagination and date filtering
- `GET /api/rides/{id}`: Get detailed ride information including GPS track

## Setup

### Prerequisites

- Node.js 18+ or Bun
- Google Maps API key
- Backend service running (see planning.md for backend setup)

### Environment Variables

Create a `.env` file with:

```env
VITE_GOOGLE_MAPS_API_KEY=your_google_maps_api_key_here
VITE_API_BASE_URL=http://localhost:8080/api
```

### Installation

```bash
# Install dependencies
npm install
# or
bun install

# Start development server
npm run dev
# or
bun dev
```

### Build for Production

```bash
npm run build
# or
bun run build
```

## Project Structure

```
src/
├── components/          # Reusable UI components
│   ├── map/            # Map-related components
│   └── sidebar.tsx     # Navigation sidebar
├── hooks/              # Custom React hooks
│   ├── useLatLngList.ts # WebSocket data management
│   ├── useRides.ts     # API data fetching
│   └── websocket.ts    # WebSocket connection
├── routes/             # Page components (TanStack Router)
│   ├── rides/          # Ride-related pages
│   └── index.tsx       # Home page
├── services/           # External service integrations
│   └── api.ts          # REST API client
├── types/              # TypeScript type definitions
└── providers/          # React context providers
```

## Technology Stack

- **React 19** with TypeScript
- **TanStack Router** for routing
- **Google Maps API** for map visualization
- **WebSocket** for real-time communication
- **Vite** for build tooling
- **CSS Modules** for styling

## Development

The application uses TanStack Router for file-based routing. Routes are automatically generated based on the file structure in `src/routes/`.

### Key Routes

- `/` - Home page
- `/rides` - Ride history overview
- `/rides/live` - Live GPS tracking
- `/rides/{id}` - Individual ride details

### WebSocket Integration

The `useLatLngList` hook manages WebSocket connections and handles different message types from the backend, maintaining both live tracking data and ride state information.

### API Integration

The `ApiService` class provides methods for fetching ride data from the backend REST API, with proper error handling and TypeScript support.
