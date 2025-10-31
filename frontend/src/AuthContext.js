// src/AuthContext.js
import React, { createContext, useContext, useState, useCallback } from "react";

const AuthContext = createContext(null);

export function AuthProvider({ children }) {
    // token и username хранятся в памяти
    const [token, setToken] = useState(null);
    const [username, setUsername] = useState(null);

    const setAuth = useCallback((newToken, user) => {
        setToken(newToken);
        setUsername(user);
    }, []);

    const clearAuth = useCallback(() => {
        setToken(null);
        setUsername(null);
    }, []);

    return (
        <AuthContext.Provider value={{ token, username, setAuth, clearAuth }}>
            {children}
        </AuthContext.Provider>
    );
}

export function useAuth() {
    const ctx = useContext(AuthContext);
    if (!ctx) throw new Error("useAuth must be used inside AuthProvider");
    return ctx;
}
