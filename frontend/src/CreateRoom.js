import React, {useEffect, useState} from "react";
import {Card, Button, Form, Container, Alert, InputGroup} from "react-bootstrap";
import { useNavigate } from "react-router-dom";
import { useAuth } from "./AuthContext";

const BASE_PATH = process.env.REACT_APP_BASE_PATH || "";

export default function CreateRoom() {
    const [user, setUser] = useState("");
    const { username } = useAuth();
    const [password, setPassword] = useState("");
    const [shareLink, setShareLink] = useState(null);
    const [joinLink, setJoinLink] = useState(null);
    const [copied, setCopied] = useState(false);
    const { setAuth } = useAuth();
    const navigate = useNavigate();
    const [error, setError] = useState("");
    const fadeError = (msg) => {
        setError(msg);
        setTimeout(() => setError(null), 2000);
    }

    useEffect(() => {
        if (username) {
            setUser(username);
        }
    }, [useAuth]);

    async function createRoom() {
        if (!user) {
            fadeError("Имя пользователя обязательно");
            return
        }

        const res = await fetch(`${BASE_PATH}/api/rooms`, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ username: user, password }),
        });
        const data = await res.json();
        setShareLink(BASE_PATH+data.join_url);
        setJoinLink(`/room/${data.room_id}`)
        setAuth(data.token, user);
    }

    const copyToClipboard = async () => {
        if (!shareLink) return;
        try {
            await navigator.clipboard.writeText(window.location.origin+shareLink);
            setCopied(true);
            setTimeout(() => setCopied(false), 2000);
        } catch (err) {
            console.error("Ошибка копирования:", err);
        }
    };

    const gotoJoin = async () => {
        if (!joinLink) return;
        navigate(joinLink);
    }

    return (
        <Container className="py-5" style={{ maxWidth: 480 }}>
            <h2 className="text-center mb-4">Создать комнату</h2>
            <Card className="shadow-sm p-4">
                <Form onSubmit={createRoom}>
                    <Form.Group className="mb-3">
                        <Form.Label>Ваше имя</Form.Label>
                        <Form.Control
                            value={user}
                            onChange={(e) => setUser(e.target.value)}
                            placeholder="Например, Александр"
                        />
                    </Form.Group>

                    <Form.Group className="mb-3">
                        <Form.Label>Введите пароль новой комнаты</Form.Label>
                        <Form.Control
                            type="password"
                            value={password}
                            onChange={(e) => setPassword(e.target.value)}
                            placeholder="Введите пароль"
                        />
                    </Form.Group>

                    {error && <Alert variant="danger">{error}</Alert>}
                    <Button variant="primary" className="w-100" onClick={createRoom}>
                        Создать комнату
                    </Button>
                </Form>

                {shareLink && (
                    <Alert variant="success" className="mt-4">
                        <strong>Комната создана!</strong>
                        <Form.Label className="mt-2 mb-1">Ссылка для подключения:</Form.Label>

                        <InputGroup>
                            <Form.Control readOnly value={window.location.origin+shareLink} />
                            <Button
                                variant={copied ? "success" : "outline-secondary"}
                                onClick={copyToClipboard}
                            >
                                {copied ? "✓ Скопировано" : "Копировать"}
                            </Button>
                            <Button onClick={gotoJoin}>Войти</Button>

                        </InputGroup>
                    </Alert>
                )}
            </Card>
        </Container>
    );
}
