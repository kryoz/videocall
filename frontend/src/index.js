import React from "react";
import { createRoot } from "react-dom/client";
import App from "./App";
import { AuthProvider } from "./AuthContext";
import "bootswatch/dist/vapor/bootstrap.min.css";
import "./css/VideoRoom.css";
import "./css/Auth.css";

createRoot(document.getElementById("root")).render(
    <React.StrictMode>
        <AuthProvider>
            <App />
        </AuthProvider>
    </React.StrictMode>
);
