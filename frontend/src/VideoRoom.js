import React, {useEffect, useRef, useState} from "react";
import {Container, Card, Button } from "react-bootstrap";
import {useNavigate, useParams} from "react-router-dom";
import { FaMicrophone, FaMicrophoneSlash, FaVideo, FaVideoSlash, FaPhoneSlash, FaWifi } from "react-icons/fa";
import { FaCameraRotate, FaDisplay } from "react-icons/fa6";
import { CircularProgressbarWithChildren, buildStyles } from "react-circular-progressbar";
import "react-circular-progressbar/dist/styles.css";

import { useAuth } from "./contexts/AuthContext";
import { Signaling } from "./Signaling";
import { useBackButtonHandler } from './hooks/useBackButtonHandler';

export default function VideoRoom() {
    const navigate = useNavigate();
    const { jwt } = useAuth();
    const { room_id } = useParams();
    const [remoteUser, setRemoteUser] = useState(null);
    const [hasRemoteStream, setHasRemoteStream] = useState(false);
    const [hasRemoteVideo, setHasRemoteVideo] = useState(false);
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
            setHasRemoteVideo(stream?.getVideoTracks().length > 0);
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

    // Drag state for mini video
    const [isDragging, setIsDragging] = useState(false);
    const [position, setPosition] = useState({ x: window.innerWidth - 200, y: 5 });
    const [dragOffset, setDragOffset] = useState({ x: 0, y: 0 });
    const miniVideoRef = useRef(null);

    const [isAudioOnly, setIsAudioOnly] = useState(false)

    let mounted = true;

    useBackButtonHandler({
        preventBack: true,
        onBackPress: () => {
            const confirmed = window.confirm(
                'Действительно хотите завершить звонок?'
            );
            if (confirmed) {
                leaveRoom();
            }
        }
    });

    // Initialize local media stream
    useEffect(() => {
        if (!jwt) {
            navigate(`/join/${room_id}`, {replace: true})
            return
        }

        (async () => {
            let stream
            try {
                stream = await initCamera()
            } catch (err) {
                console.error("🎥 Media error:", err);
                stream = await initAudioOnly()
            }

            if (!mounted) {
                stream.getTracks().forEach((t) => t.stop());
                return;
            }

            // Store stream in ref for signaling
            localStreamRef.current = stream;

            // Set local video
            if (localRef.current) {
                localRef.current.srcObject = stream;
            }

            console.log("✅ Local media initialized");
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
            audio: audioSettings(),
        });
    }

    const initAudioOnly = async () => {
        setIsAudioOnly(true)
        return await navigator.mediaDevices.getUserMedia({
            audio: audioSettings(),
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
            if (localRef.current) {
                localRef.current.srcObject = newStream;
            }
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

        const audioTracks = localStreamRef.current.getAudioTracks();
        if (audioTracks.length > 0) {
            const audioTrack = audioTracks[0];
            
            // Create a new reference to the track and enable/disable it
            const newEnabledState = !audioTrack.enabled;
            audioTrack.enabled = newEnabledState;
            
            // Force update state - this ensures UI sync regardless of browser behavior
            setMicEnabled(newEnabledState);
            
            // Debug for browser-specific issues
            console.log(`🎤 Microphone ${newEnabledState ? 'enabled' : 'disabled'}`, {
                trackEnabled: audioTrack.enabled,
                trackId: audioTrack.id,
                trackKind: audioTrack.kind,
                userAgent: navigator.userAgent
            });
            
            // Additional verification for problematic browsers like Vivaldi
            if (navigator.userAgent.includes('Vivaldi')) {
                setTimeout(() => {
                    const refreshedTracks = localStreamRef.current?.getAudioTracks();
                    if (refreshedTracks && refreshedTracks.length > 0) {
                        const refreshedTrack = refreshedTracks[0];
                        console.log('🔄 Vivaldi track verification:', {
                            refreshEnabled: refreshedTrack.enabled,
                            originalState: newEnabledState
                        });
                    }
                }, 100);
            }
        } else {
            console.warn("No audio tracks found in local stream");
        }
    };

    const toggleCam = () => {
        if (!localStreamRef.current) {
            console.warn("No local stream available");
            return;
        }

        const videoTracks = localStreamRef.current.getVideoTracks();
        if (videoTracks.length > 0) {
            const videoTrack = videoTracks[0];
            // Ensure we're accessing the actual track from the stream each time
            const track = localStreamRef.current.getVideoTracks().find(t => t === videoTrack);
            if (track) {
                track.enabled = !track.enabled;
                setCamEnabled(track.enabled);
                console.log(`📹 Camera ${track.enabled ? 'enabled' : 'disabled'}`);
            }
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
            if (localRef.current) {
                localRef.current.srcObject = screenStream;
            }
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
        try {
            // 1. Остановить экранный стрим
            if (localRef.current?.srcObject) {
                localRef.current.srcObject.getTracks().forEach((t) => t.stop());
            }

            // 2. Получить текущую камеру из localStreamRef
            const cameraTrack = localStreamRef.current
                ?.getVideoTracks()
                ?.find((t) => t.readyState === "live");

            if (cameraTrack && pcRef.current) {
                const sender = pcRef.current
                    .getSenders()
                    .find((s) => s.track?.kind === "video");

                if (sender) {
                    await sender.replaceTrack(cameraTrack);
                }
            }

            // 3. Вернуть локальное отображение камеры
            if (localRef.current) {
                localRef.current.srcObject = localStreamRef.current;
            }
            setIsScreenSharing(false);
        } catch (err) {
            console.error("Error stopping screen share:", err);
        }
    };

    const leaveRoom = () => {
        if (localStreamRef.current) {
            localStreamRef.current.getTracks().forEach((t) => t.stop());
            localStreamRef.current = null;
        }

        endCall();

        if (localRef.current) {
            localRef.current.srcObject = null;
        }

        if (pcRef.current) {
            pcRef.current.close();
            pcRef.current = null;
        }

        mounted = false;
        if (remoteRef.current) {
            remoteRef.current.srcObject = null;
        }

        navigate('/', { replace: true });
    };

    const videoSettings = (cameraMode) => {
        return {
            facingMode: cameraMode,
            width: { min: 640, ideal: 1280, max: 1920 },
            height: { min: 400, ideal: 720, max: 1080 },
            frameRate: { min: 10, ideal: 24, max: 30 },
        }
    }

    const audioSettings = () => {
        return {
            echoCancellation: true,
            noiseSuppression: true,
            autoGainControl: true,
        }
    }

    const displaySettings = () => {
        return {
            cursor: "always",
            width: { ideal: 1920, max: 3840 },
            height: { ideal: 1080, max: 2160 },
            frameRate: { ideal: 15, max: 30 },
        }
    }

    // Drag handlers for mini video
    const handleDragStart = (e) => {
        e.preventDefault();
        
        const clientX = e.type.includes('mouse') ? e.clientX : e.touches[0].clientX;
        const clientY = e.type.includes('mouse') ? e.clientY : e.touches[0].clientY;
        
        if (miniVideoRef.current) {
            const rect = miniVideoRef.current.getBoundingClientRect();
            setDragOffset({
                x: clientX - rect.left,
                y: clientY - rect.top
            });
        }
        
        setIsDragging(true);
    };

    const handleDragMove = (e) => {
        if (!isDragging) return;
        
        e.preventDefault();
        
        const clientX = e.type.includes('mouse') ? e.clientX : e.touches[0].clientX;
        const clientY = e.type.includes('mouse') ? e.clientY : e.touches[0].clientY;
        
        const videoWidth = miniVideoRef.current?.offsetWidth || 180;
        const videoHeight = miniVideoRef.current?.offsetHeight || 135;
        
        let newX = clientX - dragOffset.x;
        let newY = clientY - dragOffset.y;
        
        // Проверка границ области видимости при движении карточки
        const maxX = window.innerWidth - videoWidth;
        const maxY = window.innerHeight - videoHeight;
        
        newX = Math.max(0, Math.min(newX, maxX));
        newY = Math.max(0, Math.min(newY, maxY));
        
        setPosition({ x: newX, y: newY });
    };

    const handleDragEnd = () => {
        setIsDragging(false);
    };

    useEffect(() => {
        if (isDragging) {
            const moveHandler = (e) => handleDragMove(e);
            const endHandler = () => handleDragEnd();
            
            // Mouse events
            window.addEventListener('mousemove', moveHandler);
            window.addEventListener('mouseup', endHandler);
            
            // Touch events
            window.addEventListener('touchmove', moveHandler, { passive: false });
            window.addEventListener('touchend', endHandler);
            
            return () => {
                window.removeEventListener('mousemove', moveHandler);
                window.removeEventListener('mouseup', endHandler);
                window.removeEventListener('touchmove', moveHandler);
                window.removeEventListener('touchend', endHandler);
            };
        }
    }, [isDragging, dragOffset]);

    function ConnectionQualityIndicator({ quality = 100}) {
        const color =
            quality > 80 ? "#4caf50" : quality > 40 ? "#ffb300" : "#f44336";

        const label =
            quality > 80 ? "Отлично" : quality > 40 ? "Средне" : "Плохо";

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
            {!isAudioOnly && (
                <Card
                    ref={miniVideoRef}
                    className={`video-thumbnail shadow-sm border-light bg-dark bg-opacity-75 ${isDragging ? 'dragging' : ''}`}
                    style={{
                        left: `${position.x}px`,
                        top: `${position.y}px`,
                        cursor: isDragging ? 'grabbing' : 'grab',
                        userSelect: 'none',
                        touchAction: 'none'
                    }}
                    onMouseDown={handleDragStart}
                    onTouchStart={handleDragStart}
                >
                    <video
                        ref={localRef}
                        autoPlay
                        playsInline
                        muted
                        className="w-100 rounded"
                        style={{ pointerEvents: 'none' }}
                    />
                </Card>
            )}

            {/* Audio-only indicator */}
            {hasRemoteStream && !hasRemoteVideo && (
                <div
                    className="position-absolute top-50 start-50 translate-middle d-flex flex-column align-items-center"
                    style={{
                        backgroundColor: "rgba(0, 0, 0, 0.7)",
                        padding: "2rem",
                        borderRadius: "1rem",
                    }}
                >
                    <FaMicrophone size={80} className="text-light mb-3" />
                    <span className="text-light fs-5">Только аудио</span>
                </div>
            )}

            {/* Remote user name */}
            <div className="position-absolute top-0 start-0 p-3">
                <h5 className="text-light mb-0">
                    {remoteUser ? `${remoteUser}` : ""}
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
                        className="control-btn"
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
                    title="Покинуть"
                >
                    <FaPhoneSlash />
                </Button>
            </div>
        </Container>
    );
}