# Private VideoCalls

A full-stack project built with Golang and ReactJS, implementing video/audio calls via WebRTC.

### How It Works
- The **initiator** creates a room with a password. At this point, a room entity is created through the REST API.
- The initiator (or guest) joins the room — a WebSocket connection with the backend is established. This participant is referred to as the **Initiator**.
- The second participant joins the same room (the **Responder**). A WebSocket connection is also opened, and the **hello** command is sent. 
- The **Initiator** receives the hello and requests a list of TURN/STUN servers from the backend to start the WebRTC protocol initialization.
- The **Initiator** then sends an **answer** command with metadata required to establish a WebRTC connection to the **Responder**.
- The peers exchange WebRTC ICE candidates over the WebSocket (this process is called signaling) to negotiate the optimal communication path. 
- Once negotiation is complete, a duplex media stream between the participants is established.


### Architectural Notes
- `caddy` is used for convenient local deployment. For production, `nginx` is recommended (see an [example](caddy/nginx.server.example)).
- You can use `coturn` as your own TURN server, but make sure to read the configuration docs carefully — pay attention to SSL, external-ip, and firewall settings.
- The project exposes a REST API under `/api/`, secured with JWT.
- The WebSocket endpoint `/api/signal` is also JWT-protected.
- Room data is stored in-memory by default. For persistence or horizontal scaling you can switch to env `STORAGE_TYPE=mariadb`.


## Quick Start (Local)
- Copy `backend/.env.example` to `backend/.env` and set `TURN_HOST` to your local machine’s real IP.
- Create an empty env file `touch frontend/.env` (or configure it based on `env.example`).
- `make build-frontend`
- `make build-backend`
- `make run`
- If everything builds successfully, the service should be available at `http://localhost`.