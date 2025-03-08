# Real-Time Video Conference Platform

![App Preview](./previews/1.png)
![App Preview](./previews/2.png)
![App Preview](./previews/3.png)
![App Preview](./previews/4.png)

A high-performance video conferencing solution with integrated ephemeral messaging, built using modern technologies for optimal performance and scalability.

## Key Features ✨

**🎥 Core Video Features**

- WebRTC-based peer-to-peer video/audio
- Screen sharing with resolution control
- Dynamic bandwidth adaptation
- Room-based access via unique IDs
- Participant pinning (spotlight)

**💬 Real-Time Messaging**

- Redis Pub/Sub for instant messaging
- Ephemeral message storage (call duration only)
- Emoji reactions & message formatting
- Message history per session
- Typing indicators

**🛠️ Participant Controls**

- Individual audio/video mute
- Connection quality monitoring
- Participant role management
- Temporary ban capabilities
- Volume controls per participant

**🔒 Room Management**

- Password-protected rooms
- Auto-expiring sessions
- Participant capacity limits
- Persistent room configurations (PostgreSQL)
- Admin dashboard for moderation

## Technology Stack ⚙️

**Frontend**

- React + TypeScript
- WebRTC (Pion for Go integration)
- Tailwind CSS + Headless UI
- Redux Toolkit + React Query

**Backend (Go)**

- fiber framework
- PostgreSQL (Room metadata)
- Redis 7+ (Pub/Sub messaging)
- gRPC for internal services
- WebSocket (Gorilla)

## Architecture Overview 🏗️

```mermaid
graph TD
    A[Client] -->|WebSocket| B[Go Service]
    B -->|gRPC| C[Room Service]
    B -->|Redis Pub/Sub| D[Messaging]
    C -->|PostgreSQL| E[(Room DB)]
    D -->|Redis| F[(Message Store)]
    A -->|WebRTC| G[Peer Connections]
```
