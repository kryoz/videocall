import React, { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { useAuth } from "./contexts/AuthContext";
import { subscribeToPushNotifications } from "./services/pushNotifications";
import "./css/Auth.css";

const BASE_PATH = process.env.REACT_APP_BASE_PATH || "";

export default function Auth() {
    const [mode, setMode] = useState("login");
    const [username, setUsername] = useState("");
    const [password, setPassword] = useState("");
    const [error, setError] = useState("");
    const [loading, setLoading] = useState(false);
    const { setAuth, setRefreshAuth, jwt, isInitializing } = useAuth();
    const navigate = useNavigate();

    useEffect(() => {
        // Wait for initialization to complete before redirecting
        if (!isInitializing && jwt) {
            navigate('/', {replace: true});
        }
    }, [jwt, isInitializing, navigate]);

    const fadeError = (msg) => {
        setError(msg);
        setTimeout(() => setError(""), 5000);
    };

    const handleSubmit = async (e) => {
        e.preventDefault();
        setLoading(true);
        setError("");

        try {
            const endpoint = mode === "login"
                ? `${BASE_PATH}/api/auth/login`
                : `${BASE_PATH}/api/auth/register`;

            const response = await fetch(endpoint, {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ username, password })
            });

            if (!response.ok) {
                const errorText = await response.text();
                throw new Error(errorText || 'Ошибка авторизации');
            }

            const data = await response.json();
            setAuth(data.jwt, data.user_id, data.username);
            setRefreshAuth(data.token, data.expires);

            // Ask for notification permission after successful auth
            const notifEnabled = await subscribeToPushNotifications(data.jwt);
            if (notifEnabled) {
                console.log("✅ Push notifications enabled");
            }

            navigate('/', {replace: true});
        } catch (err) {
            fadeError(err.message);
        } finally {
            setLoading(false);
        }
    };

    const continueAsGuest = async () => {
        const guestUsername = `Guest_${Math.random().toString(36).substr(2, 6)}`;

        try {
            const response = await fetch(`${BASE_PATH}/api/auth/guest`, {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ username: guestUsername })
            });

            if (!response.ok) {
                throw new Error('Ошибка создания гостя: '+response.text);
            }

            const data = await response.json();
            setAuth(data.jwt, data.user_id, data.username);
            setRefreshAuth(data.token, data.expires);

            navigate('/', {replace: true});
        } catch (err) {
            fadeError(err.message);
        }
    };

    // Show loading while checking for existing session
    if (isInitializing) {
        return (
            <div className="auth-container">
                <div className="auth-card">
                    <div className="text-center">
                        <div className="spinner-border text-primary mb-3" role="status">
                            <span className="visually-hidden">Loading...</span>
                        </div>
                        <p>Checking session...</p>
                    </div>
                </div>
            </div>
        );
    }

    return (
        <div className="auth-container">
            <div className="auth-card">
                <h2 className="auth-title">
                    {mode === "login" ? "Войти" : "Зарегистрироваться"}
                </h2>

                <div className="auth-tabs">
                    <button
                        className={mode === "login" ? "active" : ""}
                        onClick={() => setMode("login")}
                    >
                        Войти
                    </button>
                    <button
                        className={mode === "register" ? "active" : ""}
                        onClick={() => setMode("register")}
                    >
                        Зарегистрироваться
                    </button>
                </div>

                {error && <div className="alert alert-error">{error}</div>}

                <form onSubmit={handleSubmit} className="auth-form">
                    <div className="form-group">
                        <label>{mode === "register" ? "Придумайте логин" : "Логин"} пользователя</label>
                        <input
                            type="text"
                            value={username}
                            onChange={(e) => setUsername(e.target.value)}
                            placeholder="Введите логин"
                            required
                            className="form-control"
                        />
                    </div>

                    <div className="form-group">
                        <label>Пароль</label>
                        <input
                            type="password"
                            value={password}
                            onChange={(e) => setPassword(e.target.value)}
                            placeholder="Введите пароль"
                            required
                            className="form-control"
                        />
                    </div>

                    <button
                        type="submit"
                        className="btn btn-primary"
                        disabled={loading}
                    >
                        {loading ? "Подождите..." : (mode === "login" ? "Войти" : "Зарегистрироваться")}
                    </button>
                </form>

                {mode === "register" && (
                    <div className="auth-info">
                        <small>Регистрация позволит звонить через push</small>
                    </div>
                )}

                <hr />

                <button
                    className="btn btn-outline"
                    onClick={continueAsGuest}
                >
                    Продолжить как гость
                </button>
            </div>
        </div>
    );
}