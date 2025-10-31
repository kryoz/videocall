import React, {useEffect, useRef, useState} from "react";
import {Container, Card, Button } from "react-bootstrap";
import { useNavigate } from "react-router-dom";
import { useAuth } from "./AuthContext";
import { Signaling } from "./Signaling";
import "./css/VideoRoom.css";
import { FaMicrophone, FaMicrophoneSlash, FaVideo, FaVideoSlash, FaPhoneSlash } from "react-icons/fa";

export default function VideoRoom() {
    const navigate = useNavigate();
    const { token } = useAuth();
    const [remoteUser, setRemoteUser] = useState(null);
    const [remoteStream, setRemoteStream] = useState(null);

    const localRef = useRef(null);
    const remoteRef = useRef(null);

    const {
        pcRef,
        localStreamRef,
        closeCall,
    } = Signaling({
        onRemoteUser: (name) => setRemoteUser(name),
        onRemoteStream: (stream) => setRemoteStream(stream),
        onPendingOffer: async () => await fetchOfferWithLocalMedia(),
    });

    const [micEnabled, setMicEnabled] = useState(true);
    const [camEnabled, setCamEnabled] = useState(true);
    const [speakerEnabled, setSpeakerEnabled] = useState(false);

    let mounted = true;

    // Initialize local media stream
    useEffect(() => {
        (async () => {
            try {
                const stream = await navigator.mediaDevices.getUserMedia({
                    video: {
                        frameRate: 15,
                        facingMode: 'user',
                        width: {min: 640, max: 1920},
                        height: {max: 1080},
                    },
                    audio: true
                });

                if (!mounted) {
                    stream.getTracks().forEach((t) => t.stop());
                    return;
                }

                // Store stream in ref for Signaling
                localStreamRef.current = stream;

                // Set local video
                if (localRef.current) {
                    localRef.current.srcObject = stream;
                }

                console.log("✅ Local media initialized");
            } catch (err) {
                console.error("🎥 Media error:", err);
                alert("Failed to access camera/microphone: " + err.message);
            }
        })()

        return () => {
            mounted = false;
            if (localStreamRef.current) {
                localStreamRef.current.getTracks().forEach((t) => t.stop());
                localStreamRef.current = null;
            }
        };
    }, []); // Only run once on mount

    function waitForStream() {
        return new Promise((resolve) => {
            if (localStreamRef.current) return resolve(localStreamRef.current);

            const check = setInterval(() => {
                if (localStreamRef.current) {
                    clearInterval(check);
                    resolve(localStreamRef.current);
                }
            }, 100);
        });
    }

    // Update remote video
    useEffect(() => {
        if (remoteRef.current && remoteStream) {
            remoteRef.current.srcObject = remoteStream;
            console.log("✅ Remote stream set");
        }
    }, [remoteStream]);

    // Добавляем local tracks когда подключились в роли ведомого (handleOffer)
    const fetchOfferWithLocalMedia = async () => {
        if (!localStreamRef.current) {
            console.log("⚠️ No local stream available when handling offer, waiting");
            await waitForStream()
        }

        const senders = pcRef.current.getSenders();

        localStreamRef.current.getTracks().forEach((track) => {
            const existingSender = senders.find(s => s.track === track);
            if (!existingSender) {
                pcRef.current.addTrack(track, localStreamRef.current);
                console.log(`➕ Added ${track.kind} track to peer connection`);
            }
        });
        console.log('Finished fetchOfferWithLocalMedia');
    }

    const enableSpeakerMode = async () => {
        const videoElement = remoteRef.current;
        if (!videoElement) return;

        if (!videoElement.setSinkId) {
            console.error("Browser doesn't support audio output selection");
            alert("Speaker mode not supported in this browser");
            return;
        }

        try {
            const devices = await navigator.mediaDevices.enumerateDevices();
            const speakers = devices.filter((d) => d.kind === "audiooutput");
            const speaker = speakers.find((d) => /speaker|default/i.test(d.label)) || speakers[0];

            if (speaker) {
                await videoElement.setSinkId(speaker.deviceId);
                console.log("🔊 Speaker mode enabled:", speaker.label);
                setSpeakerEnabled(true);
            } else {
                console.error("No audio output device found");
            }
        } catch (err) {
            console.error("Error selecting audio output:", err);
        }
    };

    const toggleMic = () => {
        if (!localStreamRef.current) {
            console.warn("No local stream available");
            return;
        }

        const audioTrack = localStreamRef.current
            .getTracks()
            .find((t) => t.kind === "audio");

        if (audioTrack) {
            audioTrack.enabled = !audioTrack.enabled;
            setMicEnabled(audioTrack.enabled);
            console.log(`🎤 Microphone ${audioTrack.enabled ? 'enabled' : 'disabled'}`);
        }
    };

    const toggleCam = () => {
        if (!localStreamRef.current) {
            console.warn("No local stream available");
            return;
        }

        const videoTrack = localStreamRef.current
            .getTracks()
            .find((t) => t.kind === "video");

        if (videoTrack) {
            videoTrack.enabled = !videoTrack.enabled;
            setCamEnabled(videoTrack.enabled);
            console.log(`📹 Camera ${videoTrack.enabled ? 'enabled' : 'disabled'}`);
        }
    };

    const leaveRoom = () => {
        if (localStreamRef.current) {
            localStreamRef.current.getTracks().forEach((t) => t.stop());
            localStreamRef.current = null;
        }
        closeCall();
        navigate("/");
    };

    if (!token) {
        return (
            <Container className="py-4 text-center">
                <h3 className="mb-4">Эта ссылка не работает</h3>
                <a href="/">В начало</a>
            </Container>
        );
    }

    return (
        <Container fluid className="vh-100 position-relative p-0 bg-dark text-white">
            {/* Main video */}
            <video
                ref={remoteRef}
                autoPlay
                playsInline
                className="video-fullscreen"
            />

            {/* Spinner overlay */}
            {!remoteStream && (
                <div
                    className="position-absolute top-50 start-50 translate-middle d-flex flex-column align-items-center"
                    style={{
                        zIndex: 10,
                        backgroundColor: "rgba(0, 0, 0, 0.5)",
                        padding: "2rem",
                        borderRadius: "1rem",
                    }}
                >
                    <div
                        className="spinner-border text-light mb-3"
                        style={{ width: "4rem", height: "4rem" }}
                        role="status"
                    />
                    <span className="text-light fs-5">Подключаем собеседника...</span>
                </div>
            )}

            {/* Mini video */}
            <Card
                className="video-thumbnail shadow-sm border-light bg-dark bg-opacity-75"
                style={{ cursor: 'pointer' }}
            >
                <video
                    ref={localRef}
                    autoPlay
                    playsInline
                    muted
                    className="w-100 rounded"
                />
            </Card>

            {/* Remote user name */}
            <div className="position-absolute top-0 start-0 p-3">
                <h5 className="text-light mb-0">
                    {remoteUser ? `👤 ${remoteUser}` : "Waiting for connection..."}
                </h5>
            </div>

            {/* Control panel */}
            <div className="control-bar bg-dark bg-opacity-75 p-3 rounded-pill shadow-lg">
                <Button
                    variant="light"
                    className="control-btn"
                    onClick={enableSpeakerMode}
                    title="Enable speaker mode"
                >
                    {speakerEnabled ? "🔊" : "🔇"}
                </Button>

                <Button
                    variant={micEnabled ? "light" : "danger"}
                    className="control-btn"
                    onClick={toggleMic}
                    title={micEnabled ? "Mute microphone" : "Unmute microphone"}
                >
                    {micEnabled ? <FaMicrophone /> : <FaMicrophoneSlash />}
                </Button>

                <Button
                    variant={camEnabled ? "light" : "danger"}
                    className="control-btn"
                    onClick={toggleCam}
                    title={camEnabled ? "Turn off camera" : "Turn on camera"}
                >
                    {camEnabled ? <FaVideo /> : <FaVideoSlash />}
                </Button>

                <Button
                    variant="danger"
                    className="control-btn"
                    onClick={leaveRoom}
                    title="Leave call"
                >
                    <FaPhoneSlash />
                </Button>
            </div>
        </Container>
    );
}