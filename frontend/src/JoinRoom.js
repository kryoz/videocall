import React, {useEffect, useState} from "react";
import { useNavigate, useParams } from "react-router-dom";
import { Container, Card, Form, Button, Alert } from "react-bootstrap";
import { useAuth } from "./AuthContext";

const BASE_PATH = process.env.REACT_APP_BASE_PATH || "";

export default function JoinRoom() {
    const { room_id } = useParams();
    const { username, setAuth } = useAuth();
    const [protectWithPassword, setProtectWithPassword] = useState(false);
    const navigate = useNavigate();
    const [user, setUser] = useState(localStorage.getItem("user") || "");
    const [password, setPassword] = useState("");
    const [error, setError] = useState("");
    const fadeError = (msg) => {
        setError(msg);
        setTimeout(() => setError(null), 5000);
    }
    
    async function join(ev) {
        ev.preventDefault();

        const res = await fetch(`${BASE_PATH}/api/rooms/${room_id}/join`, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ username: user, password }),
        });

        if (res.ok) {
            const data = await res.json();
            setAuth(data.token, user);
            navigate(`/room/${room_id}`);
        } else if (res.status === 406) {
            fadeError("Комната уже заполнена")
        } else if (res.status === 404) {
            fadeError("Комната не найдена")
        } else if (res.status === 401) {
            fadeError("Неверный пароль")
        } else if (res.status === 400) {
            fadeError("Имя обязательно")
        } else {
            fadeError("Серверная ошибка, попробуйте позднее")
        }
    }

    useEffect(() => {
        if (username) setUser(username);
    }, [username]);

    useEffect(() => {
        if (user) localStorage.setItem("user", user);
    }, [user]);

    useEffect(() => {
        (async () => {
            const res = await fetch(`${BASE_PATH}/api/rooms/${room_id}/fetch`, {
                method: "POST",
                headers: { "Content-Type": "application/json" },
            });

            if (res.ok) {
                const data = await res.json();
                setProtectWithPassword(data.is_protected === "1")
            } else if (res.status === 404) {
                fadeError("Комната не найдена")
            } else {
                fadeError("Серверная ошибка, попробуйте позднее")
            }
        })()

    }, []);

    return (
        <Container className="py-5" style={{ maxWidth: 480 }}>
            <h2 className="text-center mb-4">Подключиться к комнате</h2>
            <Card className="p-4">
                {error && <Alert variant="danger">{error}</Alert>}
                <Form onSubmit={join}>
                    <Form.Group className="mb-3">
                        <Form.Label>Ваше имя</Form.Label>
                        <Form.Control
                            value={user}
                            onChange={(e) => setUser(e.target.value)}
                            placeholder="Введите имя"
                        />
                    </Form.Group>

                    {protectWithPassword && (
                        <Form.Group className="mb-3">
                            <Form.Label>Пароль комнаты</Form.Label>
                            <Form.Control
                                type="password"
                                value={password}
                                onChange={(e) => setPassword(e.target.value)}
                                placeholder="Введите пароль"
                            />
                        </Form.Group>
                    )}

                    <Button variant="secondary" className="w-100" onClick={join}>
                        Присоединиться
                    </Button>
                </Form>
            </Card>
        </Container>
    );
}
