import { useEffect, useRef } from "react";
import { useAuth } from "./AuthContext";

export function Signaling({onRemoteUser, onRemoteStream, onPendingOffer}) {
    const BASE_PATH = process.env.REACT_APP_BASE_PATH || "";
    const { token, username } = useAuth();
    const pcRef = useRef(null);
    const wsRef = useRef(null);
    const iceServers = useRef(null);
    const localStreamRef = useRef(null);
    const remoteUser = useRef(null);
    const pendingCandidates = useRef([]);
    const isRemoteDescriptionSet = useRef(false);
    const isInitiator = useRef(false);
    const reconnectTimeoutRef = useRef(null);
    const isClosingRef = useRef(false);

    const sendSignal = (obj) => {
        const ws = wsRef.current;
        if (!ws || ws.readyState !== WebSocket.OPEN) {
            console.warn("⚠️ WS not open, can't send", obj.type);
            return false;
        }
        try {
            ws.send(JSON.stringify(obj));
            console.log("* ws =>", obj.type, obj.from ? `from ${obj.from}` : "");
            return true;
        } catch (err) {
            console.error("❌ Error sending signal:", err);
            return false;
        }
    };

    async function getIceServers() {
        try {
            const res = await fetch(`${BASE_PATH}/api/turn`, {
                headers: { Authorization: `Bearer ${token}` },
            });

            if (res.ok) {
                const creds = await res.json();
                console.log("✅ TURN credentials received:");

                // Собираем ICE серверы
                const iceServers = [];

                // Добавляем TURN/STUN серверы с credentials
                if (creds.uris && creds.username && creds.password) {
                    // Разделяем TURN и STUN URIs
                    const turnUris = creds.uris.filter(uri => uri.startsWith('turn:'));
                    const stunUris = creds.uris.filter(uri => uri.startsWith('stun:'));

                    // TURN серверы с аутентификацией
                    if (turnUris.length > 0) {
                        iceServers.push({
                            urls: turnUris,
                            username: creds.username,
                            credential: creds.password,
                            credentialType: 'password'
                        });
                        console.log("  ✓ Added", turnUris.length, "TURN servers");
                    }

                    // STUN серверы (без аутентификации)
                    if (stunUris.length > 0) {
                        iceServers.push({
                            urls: stunUris
                        });
                        console.log("  ✓ Added", stunUris.length, "STUN servers");
                    }
                }

                return iceServers;
            } else {
                console.warn("⚠️ TURN fetch failed:", res.status, res.statusText);
            }
        } catch (err) {
            console.error("❌ Error fetching TURN credentials:", err);
        }

        // Fallback: только публичные STUN серверы
        console.warn("⚠️ Using fallback STUN servers only");
        return [
            { urls: "stun:stun.l.google.com:19302" },
            { urls: "stun:stun1.l.google.com:19302" }
        ];
    }

    async function fetchTurnAndStart() {
        if (!token) {
            console.error("Token not issued - please authorize first");
            return;
        }

        isInitiator.current = true;
        await startCall();
    }

    const startCall = async (iceServers = []) => {
        try {
            if (!iceServers.current) {
                iceServers.current = await getIceServers();
            }

            if (!pcRef.current) {
                pcRef.current = initPeerConnection(iceServers.current);
            }

            // Add local tracks if available
            if (localStreamRef.current) {
                const existingSenders = pcRef.current.getSenders();
                localStreamRef.current.getTracks().forEach((track) => {
                    // Check if track is already added
                    const alreadyAdded = existingSenders.some(sender => sender.track === track);
                    if (!alreadyAdded) {
                        pcRef.current.addTrack(track, localStreamRef.current);
                        console.log(`➕ Added ${track.kind} track to peer connection (startCall)`);
                    }
                });
            } else {
                console.warn("⚠️ No local stream available when creating offer");
            }

            console.log("📤 Creating and sending offer...");
            const offer = await pcRef.current.createOffer();
            await pcRef.current.setLocalDescription(offer);
            sendSignal({ type: "offer", offer, from: username });
            console.log("✅ Initiated call. Sending offer to "+remoteUser.current);
        } catch (err) {
            console.error("❌ Error starting call:", err);
        }
    };

    const handleOffer = async (msg) => {
        try {
            console.log("📥 Received offer from", msg.from || "peer");

            if (!remoteUser.current) {
                remoteUser.current = msg.from;
                if (onRemoteUser) onRemoteUser(remoteUser.current);
            }

            // Get ICE servers first
            if (!iceServers.current) {
                iceServers.current = await getIceServers();
            }

            // Create peer connection (для ведомого)
            if (!pcRef.current) {
                console.log("Creating peer connection to handle offer");
                pcRef.current = initPeerConnection(iceServers.current);

                await onPendingOffer(pcRef.current)
            }

            await pcRef.current.setRemoteDescription(new RTCSessionDescription(msg.offer));
            isRemoteDescriptionSet.current = true;
            console.log("✅ Remote offer set (RemoteDescription)");

            // Add all buffered candidates
            if (pendingCandidates.current.length > 0) {
                console.log(`📦 Adding ${pendingCandidates.current.length} buffered candidates`);
                for (const c of pendingCandidates.current) {
                    try {
                        await pcRef.current.addIceCandidate(new RTCIceCandidate(c));
                    } catch (err) {
                        console.error("❌ Error adding buffered candidate:", err);
                    }
                }
                pendingCandidates.current = [];
            }

            console.log("📤 Creating and sending answer...");
            const answer = await pcRef.current.createAnswer();
            await pcRef.current.setLocalDescription(answer);
            sendSignal({ type: "answer", answer });
            console.log("✅ Answer sent");
        } catch (err) {
            console.error("❌ Error handling offer:", err);
        }
    };

    const handleAnswer = async (msg) => {
        try {
            if (!pcRef.current) {
                console.error("❌ Received answer but pc not initialized - this should not happen!");
                return;
            }

            console.log("📥 Received answer from peer");
            await pcRef.current.setRemoteDescription(new RTCSessionDescription(msg.answer));
            isRemoteDescriptionSet.current = true;
            console.log("✅ Remote description (answer) set");

            // Add buffered candidates
            for (const c of pendingCandidates.current) {
                try {
                    await pcRef.current.addIceCandidate(new RTCIceCandidate(c));
                    console.log("✅ Buffered ICE candidate added");
                } catch (err) {
                    console.error("❌ Error adding buffered candidate:", err);
                }
            }
            pendingCandidates.current = [];
        } catch (err) {
            console.error("❌ Error handling answer:", err);
        }
    };

    const handleCandidate = async (msg) => {
        if (!msg.candidate) return;

        if (isRemoteDescriptionSet.current && pcRef.current) {
            try {
                await pcRef.current.addIceCandidate(new RTCIceCandidate(msg.candidate));
                console.log("✅ Added ICE candidate:", msg.candidate.candidate);
            } catch (err) {
                console.error("❌ Error adding ICE candidate:", err);
            }
        } else {
            console.log("🕐 Candidate buffered (waiting for remote SDP)");
            pendingCandidates.current.push(msg.candidate);
        }
    };

    const initPeerConnection = (iceServers = []) => {
        // Don't recreate if already exists
        if (pcRef.current) {
            console.log("⚠️ Peer connection already exists, reusing it");
            return pcRef.current;
        }

        console.log("🔧 Initializing peer connection with ICE servers:",
            iceServers.map(s => ({ urls: s.urls, hasAuth: !!s.username })));

        const peer = new RTCPeerConnection({
            iceServers,
            iceTransportPolicy: "all", // Try both STUN and TURN
            iceCandidatePoolSize: 10
        });

        peer.onicecandidate = (e) => {
            if (e.candidate) {
                const type = e.candidate.type;
                const protocol = e.candidate.protocol;
                const address = e.candidate.address || 'unknown';
                console.log(`🧊 ICE candidate generated: ${type} ${protocol} ${address}`);
                sendSignal({ type: "candidate", candidate: e.candidate });
            } else {
                console.log("✅ All ICE candidates have been sent");
            }
        };

        peer.onicegatheringstatechange = () => {
            console.log("ICE gathering state:", peer.iceGatheringState);
        };

        peer.oniceconnectionstatechange = () => {
            console.log("🔌 ICE connection state:", peer.iceConnectionState);

            if (peer.iceConnectionState === "failed") {
                console.error("❌ ICE connection failed, restarting...");
                peer.restartIce();
            } else if (peer.iceConnectionState === "connected") {
                console.log("✅ ICE connection established");
            } else if (peer.iceConnectionState === "disconnected") {
                console.warn("⚠️ ICE connection disconnected");
            } else if (peer.iceConnectionState === "closed") {
                console.log("ICE connection closed");
            }
        };

        peer.onconnectionstatechange = () => {
            console.log("🔗 Peer connection state:", peer.connectionState);

            if (peer.connectionState === "connected") {
                console.log("✅ Peer connection established successfully!");
            } else if (peer.connectionState === "failed") {
                console.error("❌ Peer connection failed");
            } else if (peer.connectionState === "disconnected") {
                console.warn("⚠️ Peer connection disconnected");
            } else if (peer.connectionState === "closed") {
                console.log("Peer connection closed");
            }
        };

        peer.ontrack = (e) => {
            console.log("🎥 Remote track received:", e.track.kind);
            if (e.streams && e.streams[0]) {
                onRemoteStream?.(e.streams[0]);
            }
        };

        return peer;
    };

    const closeCall = () => {
        isClosingRef.current = true;

        if (pcRef.current) {
            pcRef.current.close();
            pcRef.current = null;
        }

        if (wsRef.current) {
            wsRef.current.close();
            wsRef.current = null;
        }

        if (reconnectTimeoutRef.current) {
            clearTimeout(reconnectTimeoutRef.current);
            reconnectTimeoutRef.current = null;
        }

        remoteUser.current = null;
        isRemoteDescriptionSet.current = false;
        pendingCandidates.current = [];
    };

    useEffect(() => {
        if (!token) {
            console.error("JWT token not issued!");
            return () => {};
        }

        isClosingRef.current = false;

        const connectWebSocket = () => {
            const origin = window.location.origin.replace(/^http/, "ws");
            const url = `${origin}${BASE_PATH}/api/signal?token=${encodeURIComponent(token)}`;

            console.log("Connecting to WebSocket:", url);
            const ws = new WebSocket(url);
            wsRef.current = ws;

            ws.onopen = () => {
                console.log("✅ WS connected");
                sendSignal({ type: "hello", from: username });
            };

            ws.onmessage = async (ev) => {
                let msg;
                try {
                    msg = JSON.parse(ev.data);
                } catch (e) {
                    console.error("WS parse error:", e);
                    return;
                }

                if (!msg.type) return;

                switch (msg.type) {
                    case "offer":
                        await handleOffer(msg);
                        break;

                    case "answer":
                        await handleAnswer(msg);
                        break;

                    case "candidate":
                        await handleCandidate(msg);
                        break;

                    case "hello":
                        if (!remoteUser.current) {
                            remoteUser.current = msg.from;
                            if (onRemoteUser) onRemoteUser(remoteUser.current);

                            console.log(`📞 Got hello from ${msg.from}, I am ${username}`);
                            console.log("📞 I will initiate the call (send offer)");
                            await fetchTurnAndStart();
                        }
                        break;
                    case "ping":
                        break;
                    default:
                        console.warn("Unknown message:", msg);
                }
            };

            ws.onclose = (event) => {
                console.log("❌ WS closed:", event.code, event.reason);
                remoteUser.current = null;
                onRemoteUser?.(null);

                // Don't reconnect if we're intentionally closing
                if (!isClosingRef.current && token) {
                    console.log("Attempting to reconnect in 3s...");
                    reconnectTimeoutRef.current = setTimeout(() => {
                        if (!isClosingRef.current) {
                            connectWebSocket();
                        }
                    }, 3000);
                }
            };

            ws.onerror = (e) => {
                console.error("❌ WS error:", e);
            };
        };

        connectWebSocket();

        // Ping interval
        const interval = setInterval(() => {
            if (wsRef.current?.readyState === WebSocket.OPEN) {
                wsRef.current.send(JSON.stringify({ type: "ping" }));
            }
        }, 20000);

        return () => {
            isClosingRef.current = true;

            if (reconnectTimeoutRef.current) {
                clearTimeout(reconnectTimeoutRef.current);
            }

            try {
                if (wsRef.current) {
                    wsRef.current.close();
                }
            } catch (err) {
                console.error("Error closing WS:", err);
            }

            wsRef.current = null;
            remoteUser.current = null;
            clearInterval(interval);
        };
    }, [token, username]);

    return {
        pcRef,
        localStreamRef,
        closeCall,
    };
}