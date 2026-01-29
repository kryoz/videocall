import React, { createContext, useContext, useState, useCallback, useEffect, useRef } from "react";
import { Cookies } from 'react-cookie'

const AuthContext = createContext(null);

const BASE_PATH = process.env.REACT_APP_BASE_PATH || "";

export function AuthProvider({ children }) {
    const [jwt, setJwt] = useState(null);
    const [refreshToken, setRefreshToken] = useState(null);
    const [userId, setUserId] = useState(() => localStorage.getItem("user_id") || null);
    const [username, setUsername] = useState(() => localStorage.getItem("username") || null);
    const [isInitializing, setIsInitializing] = useState(true);
    const cookies = new Cookies();
    const refreshingRef = useRef(false);
    const refreshPromiseRef = useRef(null);

    useEffect(() => {
        const token = cookies.get('refresh_token');
        setRefreshToken(token);

        // Try to restore session on mount if we have a refresh token
        if (token) {
            refreshJwt().finally(() => setIsInitializing(false));
        } else {
            setIsInitializing(false);
        }
    }, []);

    const refreshJwt = useCallback(async () => {
        // If already refreshing, return the existing promise
        if (refreshingRef.current && refreshPromiseRef.current) {
            return refreshPromiseRef.current;
        }

        const currentRefreshToken = cookies.get('refresh_token');

        if (!currentRefreshToken) {
            console.log("No refresh token available");
            return null;
        }

        refreshingRef.current = true;

        refreshPromiseRef.current = (async () => {
            try {
                const response = await fetch(`${BASE_PATH}/api/auth/refresh`, {
                    method: "POST",
                    headers: {"Content-Type": "application/json"},
                    body: JSON.stringify({token: currentRefreshToken})
                });

                if (!response.ok) {
                    throw new Error("Token refresh failed");
                }

                const data = await response.json();

                // Update state with new JWT and user info
                setJwt(data.jwt);

                if (data.user_id) {
                    setUserId(data.user_id);
                    localStorage.setItem("user_id", data.user_id);
                }

                if (data.username) {
                    setUsername(data.username);
                    localStorage.setItem("username", data.username);
                }

                // Optionally update refresh token if backend sends a new one
                if (data.token && data.expires) {
                    const expires = new Date(data.expires);
                    cookies.set('refresh_token', data.token, {path: '/', expires});
                    setRefreshToken(data.token);
                }

                return data.jwt;
            } catch (error) {
                console.error("Failed to refresh token:", error);
                // Clear auth on refresh failure
                clearAuth();
                return null;
            } finally {
                refreshingRef.current = false;
                refreshPromiseRef.current = null;
            }
        })();

        return refreshPromiseRef.current;
    }, [cookies]);

    const setAuth = useCallback((newJwt, newUserId, newUsername) => {
        setJwt(newJwt);
        setUserId(newUserId);
        setUsername(newUsername);

        if (newUserId) {
            localStorage.setItem("user_id", newUserId);
        }
        if (newUsername) {
            localStorage.setItem("username", newUsername);
        }
    }, []);

    const setRefreshAuth = useCallback((newToken, newTokenExpires) => {
        const expires = new Date(newTokenExpires);
        cookies.set('refresh_token', newToken, {path: '/', expires});
        setRefreshToken(newToken);
    }, [cookies]);

    const clearAuth = useCallback(() => {
        const currentRefreshToken = refreshToken || cookies.get('refresh_token');

        setJwt(null);
        setUserId(null);
        setUsername(null);
        setRefreshToken(null);

        localStorage.removeItem("user_id");
        localStorage.removeItem("username");
        cookies.remove("refresh_token");

        // Revoke refresh token on backend
        if (currentRefreshToken) {
            fetch(`${BASE_PATH}/api/auth/revoke`, {
                method: "POST",
                headers: {"Content-Type": "application/json"},
                body: JSON.stringify({token: currentRefreshToken})
            }).catch(err => console.error("Failed to revoke token:", err));
        }
    }, [refreshToken, cookies]);

    return (
        <AuthContext.Provider value={{
            jwt,
            userId,
            username,
            refreshToken,
            isInitializing,
            setAuth,
            setRefreshAuth,
            clearAuth,
            refreshJwt
        }}>
            {children}
        </AuthContext.Provider>
    );
}

export function useAuth() {
    const ctx = useContext(AuthContext);
    if (!ctx) throw new Error("useAuth must be used inside AuthProvider");
    return ctx;
}