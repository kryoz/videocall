import React, { useEffect, useState } from "react";
import { Card, Button, Form, Container, Alert, InputGroup, Modal } from "react-bootstrap";
import {useLocation, useNavigate} from "react-router-dom";
import { useAuth } from "./contexts/AuthContext.js";
import { inviteUserToRoom } from "./services/pushNotifications";
import { useAuthFetch } from "./hooks/useAuthFetch";
import { useBackButtonHandler } from './hooks/useBackButtonHandler';

const BASE_PATH = process.env.REACT_APP_BASE_PATH || "";

export default function CreateRoom() {
    const { username, userId, jwt, setAuth } = useAuth();
    const [shareLink, setShareLink] = useState(null);
    const [joinLink, setJoinLink] = useState(null);
    const [roomId, setRoomId] = useState(null);
    const [copied, setCopied] = useState(false);
    const navigate = useNavigate();
    const location = useLocation();
    const [error, setError] = useState("");
    const authFetch = useAuthFetch();

    // Invite modal state
    const [showInviteModal, setShowInviteModal] = useState(false);
    const [inviteUsername, setInviteUsername] = useState("");
    const [inviteLoading, setInviteLoading] = useState(false);
    const [inviteError, setInviteError] = useState("");

    const fadeError = (msg) => {
        setError(msg);
        setTimeout(() => setError(null), 3000);
    };

    // Redirect to auth if no user
    useEffect(() => {
        if (!userId || !username) {
            goToAuth('/auth');
        }
    }, [userId, username, navigate]);

    useBackButtonHandler({
        isExitPage: true,
        confirmMessage: 'Хотите закрыть приложение?'
    });

    async function createRoom(ev) {
        ev.preventDefault();

        if (!jwt) {
            fadeError("Please authenticate first");
            goToAuth();
            return;
        }

        const res = await authFetch(`${BASE_PATH}/api/rooms`, {
            method: "POST",
        });

        if (!res.ok) {
            fadeError("Ошибка создания комнаты. Что-то на сервере.");
            return;
        }

        const data = await res.json();
        setShareLink(BASE_PATH + data.join_url);
        setJoinLink(`/room/${data.room_id}`);
        setRoomId(data.room_id);
        setAuth(data.jwt, userId, username);
    }

    const copyToClipboard = async () => {
        if (!shareLink) return;
        try {
            await navigator.clipboard.writeText(window.location.origin + shareLink);
            setCopied(true);
            setTimeout(() => setCopied(false), 2000);
        } catch (err) {
            console.error("Copy error:", err);
        }
    };

    const handleInvite = async () => {
        if (!inviteUsername || !roomId || !jwt) return;

        setInviteLoading(true);
        setInviteError("");

        try {
            await inviteUserToRoom(jwt, roomId, inviteUsername);
            setShowInviteModal(false);
            setInviteUsername("");
            fadeError("Приглашение отправлено! Можно заходить в комнату.");
        } catch (err) {
            setInviteError(err.message);
        } finally {
            setInviteLoading(false);
        }
    };

    const goToAuth = () => {
        navigate('/auth', {
            replace: true,
            state: { from: location.pathname + location.search }
        });
    };

    const gotoJoin = () => {
        if (!joinLink) return;
        navigate(joinLink, {replace: true});
    };

    return (
        <Container className="py-5" style={{ maxWidth: 480 }}>
            <div className="d-flex justify-content-between align-items-center mb-4">
                <h2>Создать комнату</h2>
                <div className="d-flex align-items-center gap-2">
                    <span className="text-muted small">
                        {username} {!jwt && "(Guest)"}
                    </span>
                    {!jwt && (
                        <Button variant="outline-secondary" size="sm" onClick={goToAuth}>
                            Login
                        </Button>
                    )}
                </div>
            </div>

            <Card className="p-4">
                {!shareLink && (
                    <Form onSubmit={createRoom}>
                        {error && <Alert variant="danger">{error}</Alert>}
                        <Button variant="secondary" className="w-100" onClick={createRoom}>
                            Создать комнату
                        </Button>
                    </Form>
                )}
                {shareLink && (
                    <Alert variant="primary" className="mt-4">
                        <h5>Комната создана!</h5>
                        <Form.Label className="mt-2 mb-1">Пригласительная ссылка:</Form.Label>

                        <InputGroup>
                            <Form.Control readOnly value={window.location.origin + shareLink} />
                            <Button
                                variant={copied ? "outline-success" : "secondary"}
                                onClick={copyToClipboard}
                            >
                                {copied ? "✓ Скопировано" : "Копировать"}
                            </Button>
                            <Button onClick={gotoJoin} variant="outline-secondary">
                                Войти
                            </Button>
                        </InputGroup>

                        {jwt && (
                            <Button
                                variant="outline-primary"
                                className="w-100 mt-3"
                                onClick={() => setShowInviteModal(true)}
                            >
                                Позвать пользователя
                            </Button>
                        )}
                    </Alert>
                )}
            </Card>

            <Modal show={showInviteModal} onHide={() => setShowInviteModal(false)}>
                <Modal.Header closeButton>
                    <Modal.Title>Пригласить пользователя</Modal.Title>
                </Modal.Header>
                <Modal.Body>
                    {inviteError && <Alert variant="danger">{inviteError}</Alert>}
                    <Form.Group>
                        <Form.Label>Кого пригласить?</Form.Label>
                        <Form.Control
                            value={inviteUsername}
                            onChange={(e) => setInviteUsername(e.target.value)}
                            placeholder="Введите имя пользователя"
                        />
                    </Form.Group>
                </Modal.Body>
                <Modal.Footer>
                    <Button
                        variant="secondary"
                        onClick={handleInvite}
                        disabled={inviteLoading || !inviteUsername}
                    >
                        {inviteLoading ? "Отправляем..." : "Отправить приглашение"}
                    </Button>
                    <Button variant="primary" onClick={() => setShowInviteModal(false)}>
                        Отмена
                    </Button>
                </Modal.Footer>
            </Modal>
        </Container>
    );
}