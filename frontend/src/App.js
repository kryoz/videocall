import React from "react";
import { BrowserRouter, Routes, Route } from "react-router-dom";
import CreateRoom from "./CreateRoom";
import JoinRoom from "./JoinRoom";
import VideoRoom from "./VideoRoom";

const BASE_PATH = process.env.REACT_APP_BASE_PATH || "";

export default function App() {
    return (
        <BrowserRouter basename={BASE_PATH}>
            <Routes>
                <Route path="/" element={<CreateRoom />} />
                <Route path="/join/:room_id" element={<JoinRoom />} />
                <Route path="/room/:room_id" element={<VideoRoom />} />
            </Routes>
        </BrowserRouter>
    );
}
