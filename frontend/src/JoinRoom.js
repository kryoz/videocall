import React, { useEffect, useState } from "react";
import {useLocation, useNavigate, useParams} from "react-router-dom";
import { Container, Card, Button, Alert } from "react-bootstrap";
import { useAuth } from "./contexts/AuthContext";
import { useAuthFetch } from "./hooks/useAuthFetch";

const BASE_PATH = process.env.REACT_APP_BASE_PATH || "";

export default function JoinRoom() {
    const { room_id } = useParams();
    const { username, userId, jwt, setAuth, isInitializing, refreshToken } = useAuth();
    const navigate = useNavigate();
    const location = useLocation()
    const [error, setError] = useState("");
    const [loading, setLoading] = useState(true);
    const authFetch = useAuthFetch();

    const fadeError = (msg) => {
        setError(msg);
        setTimeout(() => setError(null), 5000);
    };

    useEffect(() => {
        // Wait for auth initialization to complete
        if (isInitializing) {
            return;
        }

        if (!jwt) {
            // Redirect to auth if no JWT
            setLoading(false);
            return;
        }

        if (jwt) {
            (async () => {
                try {
                    const res = await authFetch(`${BASE_PATH}/api/rooms/${room_id}/join`, {
                        method: "POST",
                    });

                    if (res.ok) {
                        const data = await res.json();
                        // Обновляем jwt поскольку в него добавлен теперь room_id
                        if (data.jwt) {
                            setAuth(data.jwt, userId, username);
                        }

                        // Переходим на страницу звонка VideoRoom
                        navigate(`/room/${room_id}`, {replace: true});
                    } else if (res.status === 406) {
                        fadeError('Комната уже занята!');
                        setLoading(false);
                    } else if (res.status === 404) {
                        fadeError('Комната не найдена!');
                        setLoading(false);
                    } else {
                        throw new Error()
                    }
                } catch (err) {
                    fadeError(err.message || 'Серверная ошибка, попробуйте  позднее');
                    setLoading(false);
                }
            })();
        }
    }, [jwt, refreshToken, isInitializing, room_id]);

    // Show loading while initializing auth
    if (isInitializing || (loading && (jwt || refreshToken))) {
        return (
            <Container className="py-5" style={{ maxWidth: 480 }}>
                <Card className="p-4 text-center">
                    <div className="spinner-border text-primary mb-3" role="status">
                        <span className="visually-hidden">Загрузка...</span>
                    </div>
                    <p>Присоединяемся...</p>
                </Card>
            </Container>
        );
    }

    // Show auth prompt only if no JWT and no refresh token
    if (!jwt && !refreshToken) {
        return (
            <Container className="py-5" style={{maxWidth: 480}}>
                <h2 className="text-center mb-4">Подключение к комнате</h2>
                <Card className="p-4">
                    {error && <Alert variant="danger">{error}</Alert>}

                    <p className="text-center mb-3">
                        Для входа в комнату необходимо идентифицироваться.
                    </p>

                    <Button
                        variant="secondary"
                        className="w-100 mb-2"
                        onClick={() => navigate(
                            '/auth', {
                                replace: true,
                                state: { from: location.pathname + location.search }
                            })
                    }
                    >
                        Авторизоваться
                    </Button>

                    <Button
                        variant="outline-secondary"
                        className="w-100"
                        onClick={() => navigate('/', {replace: true})}
                    >
                        На главную
                    </Button>
                </Card>
            </Container>
        );
    }

    // Default fallback state
    return (
        <Container className="py-5" style={{ maxWidth: 480 }}>
            <Card className="p-4">
                {error && <Alert variant="danger">{error}</Alert>}
                <p className="text-center">Невозможно подключиться. Попробуйте позже или пересоздайте комнату.</p>
                <Button
                    variant="outline-secondary"
                    className="w-100"
                    onClick={() => navigate('/', {replace: true})}
                >
                    На главную
                </Button>
            </Card>
        </Container>
    );
}