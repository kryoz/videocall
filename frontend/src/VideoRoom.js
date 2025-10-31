import React, {useEffect, useRef, useState} from "react";
import {Container, Card, Button } from "react-bootstrap";
import {useNavigate, useParams} from "react-router-dom";
import { FaMicrophone, FaMicrophoneSlash, FaVideo, FaVideoSlash, FaPhoneSlash, FaWifi } from "react-icons/fa";
import { FaCameraRotate, FaDisplay } from "react-icons/fa6";
import { CircularProgressbarWithChildren, buildStyles } from "react-circular-progressbar";
import "react-circular-progressbar/dist/styles.css";

import { useAuth } from "./AuthContext";
import { Signaling } from "./Signaling";


export default function VideoRoom() {
    const navigate = useNavigate();
    const { token } = useAuth();
    const { room_id } = useParams();
    const [remoteUser, setRemoteUser] = useState(null);
    const [hasRemoteStream, setHasRemoteStream] = useState(false);

    const localRef = useRef(null);
    const remoteRef = useRef(null);

    const {
        pcRef,
        localStreamRef,
        endCall,
    } = Signaling({
        onRemoteUser: (name) => setRemoteUser(name),
        onRemoteStream: (stream) => {
            setHasRemoteStream(!!stream)
            if (remoteRef.current) {
                remoteRef.current.srcObject = stream;
                if (stream) {
                    console.log("✅ Remote stream is set");
                } else {
                    console.log("❌ Remote stream is removed");
                }
            }
        },
        onPendingOffer: async () => await fetchOfferWithLocalMedia(),
        onStatsUpdate: (quality) => setVideoQuality(quality),
    });

    const [micEnabled, setMicEnabled] = useState(true);
    const [camEnabled, setCamEnabled] = useState(true);
    const [cameraFacingMode, setCameraFacingMode] = useState("user");
    const [videoQuality, setVideoQuality] = useState(100);
    const [isScreenSharing, setIsScreenSharing] = useState(false);

    let mounted = true;

    // Initialize local media stream
    useEffect(() => {
        if (!token) {
            navigate(`/join/${room_id}`)
            return
        }

        (async () => {
            try {
                const stream = await initCamera()

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
            if (localRef.current) {
                localRef.current.getTracks().forEach((t) => t.stop());
                localRef.current = null;
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

    const initCamera = async () => {
        return await navigator.mediaDevices.getUserMedia({
            video: videoSettings('user'),
            audio: {
                echoCancellation: true,
                noiseSuppression: true,
                autoGainControl: true,
            },
        });
    }

    const cameraSwitch = async () => {
        try {
            if (!localStreamRef.current) {
                console.warn("No local stream to switch camera");
                return;
            }

            // Останавливаем старую камеру
            localStreamRef.current.getVideoTracks().forEach((t) => t.stop());

            // Выбираем новый режим
            const newMode = cameraFacingMode === "user" ? "environment" : "user";

            // Запрашиваем новый стрим
            const newStream = await navigator.mediaDevices.getUserMedia({
                video: videoSettings(newMode),
                audio: false,
            });

            // Меняем трек в PeerConnection
            const videoSender = pcRef.current
                ?.getSenders()
                .find((s) => s.track?.kind === "video");
            if (videoSender) {
                await videoSender.replaceTrack(newStream.getVideoTracks()[0]);
            }

            // Обновляем локальное видео
            newStream.getVideoTracks().forEach((track) =>
                localStreamRef.current.addTrack(track)
            );
            localRef.current.srcObject = newStream;
            localStreamRef.current = newStream;

            setCameraFacingMode(newMode);
            console.log(`🔁 Switched to ${newMode} camera`);
        } catch (err) {
            console.error("Camera switch error:", err);
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

    const startScreenShare = async () => {
        if (!pcRef.current) {
            return
        }

        try {
            // 1. Запрос экрана
            const screenStream = await navigator.mediaDevices.getDisplayMedia({
                video: displaySettings(),
                audio: false,
            });
            localRef.current = screenStream;

            // 2. Берём первый видеотрек (экран)
            const screenTrack = screenStream.getVideoTracks()[0];

            // 3. Находим текущий видеосендер и заменяем трек
            try {
                const sender = pcRef.current
                    .getSenders()
                    .find((s) => s.track?.kind === "video");
                if (sender) {
                    await sender.replaceTrack(screenTrack);
                }
            } catch (err) {
                console.error(err)
            }

            // 4. Отображаем локально
            localRef.current.srcObject = screenStream;
            setIsScreenSharing(true);

            // 5. Когда пользователь остановит демонстрацию (через UI браузера)
            screenTrack.onended = () => {
                stopScreenShare();
            };
        } catch (err) {
            console.error("Ошибка при демонстрации экрана:", err);
        }
    };

    const stopScreenShare = async () => {
        // 1. Остановить экранный стрим
        localRef.current?.getTracks().forEach((t) => t.stop());

        // 2. Вернуть камеру обратно
        let cameraTrack = null
        try {
            cameraTrack = localStreamRef.current
                ?.getVideoTracks()
                ?.find((t) => t.readyState === "live");
            if (!cameraTrack) return;
        } catch (err) {
            console.error(err)
            return;
        }

        if (pcRef.current) {
            const sender = pcRef.current
                .getSenders()
                .find((s) => s.track?.kind === "video");

            if (sender) {
                await sender.replaceTrack(cameraTrack);
            }
        }

        // 3. Вернуть локальное отображение камеры
        localRef.current.srcObject = localStreamRef.current;
        setIsScreenSharing(false);
    };

    const leaveRoom = () => {
        endCall();

        mounted = false;
        if (localStreamRef.current) {
            localStreamRef.current.getTracks().forEach((t) => t.stop());
            localStreamRef.current = null;
        }

        localRef.current.srcObject = null;
        remoteRef.current.srcObject = null;

        navigate("/");
    };

    const videoSettings = (cameraMode) => {
        return {
            facingMode: cameraMode,
            width: { min: 640, ideal: 1280, max: 1920 },
            height: { min: 400, ideal: 720, max: 1080 },
            frameRate: { min: 10, ideal: 24, max: 30 },
        }
    }

    const displaySettings = () => {
        return {
            cursor: "always",
            width: { ideal: 1920, max: 2560 },
            height: { ideal: 1080, max: 1440 },
            frameRate: { ideal: 15, max: 30 },
        }
    }

    function ConnectionQualityIndicator({ quality = 100}) {
        const color =
            quality > 75 ? "#4caf50" : quality > 40 ? "#ffb300" : "#f44336";

        // Текстовая подсказка
        const label =
            quality > 75 ? "Отлично" : quality > 40 ? "Средне" : "Плохо";

        return (
            <div
                className="position-absolute top-0 end-0 m-3 quality-monitor"
                title={`Качество соединения: ${quality}% (${label})`}
            >
                <CircularProgressbarWithChildren
                    value={quality}
                    strokeWidth={10}
                    styles={buildStyles({
                        pathColor: color,
                        trailColor: "rgba(255,255,255,0.15)",
                        pathTransitionDuration: 0.5,
                    })}
                >
                    <FaWifi
                        size={20}
                        color={color}
                        style={{ marginBottom: "4px", transition: "color 0.3s" }}
                    />
                    <div
                        style={{
                            fontSize: "0.5rem",
                            color,
                            fontWeight: "600",
                            lineHeight: "1rem",
                        }}
                    >
                        {quality}%
                    </div>
                </CircularProgressbarWithChildren>
            </div>
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

            {/* Connection quality */}
            <ConnectionQualityIndicator quality={videoQuality} />

            {/* Spinner overlay */}
            {!hasRemoteStream && (
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
                    <span className="text-light fs-5">Ожидаем собеседника...</span>
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
                    {remoteUser ? `👤 ${remoteUser}` : "Ожидание подключения..."}
                </h5>
            </div>

            {/* Control panel */}
            <div className="control-bar nav-pills p-3">
                <Button
                    variant="light"
                    className="control-btn"
                    onClick={cameraSwitch}
                    title="Переключить камеру (фронт/тыл)"
                >
                    <FaCameraRotate/>
                </Button>

                {!isScreenSharing ? (
                    <Button
                        variant="light"
                        onClick={startScreenShare}
                        className="control-btn"
                        title="Поделиться экраном"
                    >
                        <FaDisplay />
                    </Button>
                ) : (
                    <Button
                        variant="danger"
                        onClick={stopScreenShare}
                        className="bg-red-600 text-white px-4 py-2 rounded"
                        title="Остановить демонстрацию"
                    >
                        <FaDisplay />
                    </Button>
                )}

                <Button
                    variant={micEnabled ? "light" : "danger"}
                    className="control-btn"
                    onClick={toggleMic}
                    title={micEnabled ? "Выкл. микрофон" : "Вкл. микрофон"}
                >
                    {micEnabled ? <FaMicrophone /> : <FaMicrophoneSlash />}
                </Button>

                <Button
                    variant={camEnabled ? "light" : "danger"}
                    className="control-btn"
                    onClick={toggleCam}
                    title={camEnabled ? "Выкл. камеру" : "Вкл. камеру"}
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