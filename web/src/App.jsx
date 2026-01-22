import React, { useEffect, useRef, useState } from "react";
import {
  Link,
  Navigate,
  Route,
  Routes,
  useLocation,
  useNavigate,
  useParams,
} from "react-router-dom";

const API_BASE =
  import.meta.env.VITE_API_URL || `${window.location.origin}/api`;
const ICE_SERVERS = [
  { urls: "stun:stun.l.google.com:19302" },
  ...(import.meta.env.VITE_TURN_URL
    ? [
        {
          urls: import.meta.env.VITE_TURN_URL,
          username: import.meta.env.VITE_TURN_USERNAME,
          credential: import.meta.env.VITE_TURN_CREDENTIAL,
        },
      ]
    : []),
];
const AUTH_TOKEN_KEY = "authToken";
const AUTH_USER_KEY = "authUser";
const DISPLAY_NAME_KEY = "displayName";

function randomId() {
  return Math.random().toString(36).slice(2, 10);
}

const DEVICE_ID_KEY = "deviceId";

function getDeviceId() {
  let deviceId = localStorage.getItem(DEVICE_ID_KEY);
  if (!deviceId) {
    deviceId = randomId() + randomId();
    localStorage.setItem(DEVICE_ID_KEY, deviceId);
  }
  return deviceId;
}

function buildWsUrl(serverUrl, roomId, token) {
  const wsBase = serverUrl.replace(/^http/, "ws");
  const url = `${wsBase}/ws/${roomId}`;
  return token ? `${url}?token=${token}` : url;
}

function readAuth() {
  return {
    token: localStorage.getItem(AUTH_TOKEN_KEY) || "",
    user: localStorage.getItem(AUTH_USER_KEY) || "",
  };
}

function saveAuth(token, user) {
  localStorage.setItem(AUTH_TOKEN_KEY, token);
  localStorage.setItem(AUTH_USER_KEY, user);
}

function clearAuth() {
  localStorage.removeItem(AUTH_TOKEN_KEY);
  localStorage.removeItem(AUTH_USER_KEY);
}

async function logout() {
  try {
    await fetch(`${API_BASE}/auth/logout`, {
      method: "POST",
      credentials: "include",
    });
  } catch (err) {
    // ignore
  }
  clearAuth();
}

function readDisplayName() {
  return localStorage.getItem(DISPLAY_NAME_KEY) || "";
}

function saveDisplayName(name) {
  localStorage.setItem(DISPLAY_NAME_KEY, name);
}

async function tryRefreshToken() {
  try {
    const response = await fetch(`${API_BASE}/auth/refresh`, {
      method: "POST",
      credentials: "include",
    });
    if (response.ok) {
      const data = await response.json();
      saveAuth(data.token, data.username);
      return data;
    }
  } catch (err) {
    // ignore
  }
  return null;
}

function VideoTile({ stream, muted, label }) {
  const ref = useRef(null);

  useEffect(() => {
    if (ref.current) {
      ref.current.srcObject = stream || null;
    }
  }, [stream]);

  return (
    <div className="video-tile">
      <video ref={ref} autoPlay playsInline muted={muted} />
      <div className="video-label">{label}</div>
    </div>
  );
}

function RegisterPage() {
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [authError, setAuthError] = useState("");
  const navigate = useNavigate();
  const location = useLocation();

  useEffect(() => {
    async function check() {
      if (location.state?.loggedOut) {
        return;
      }
      const { token } = readAuth();
      if (token) {
        navigate("/dashboard", { replace: true });
        return;
      }
      const data = await tryRefreshToken();
      if (data) {
        navigate("/dashboard", { replace: true });
      }
    }
    check();
  }, [navigate, location]);

  async function register() {
    setAuthError("");
    const response = await fetch(`${API_BASE}/auth/register`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "X-Device-ID": getDeviceId(),
      },
      body: JSON.stringify({ username, password }),
      credentials: "include",
    });
    if (!response.ok) {
      setAuthError("Registration failed");
      return;
    }
    const data = await response.json();
    saveAuth(data.token, data.username);
    navigate("/dashboard");
  }

  return (
    <div className="auth-page">
      <div className="auth-card">
        <h1>Register</h1>
        <div className="row">
          <div className="field grow">
            <label>Username</label>
            <input
              value={username}
              onChange={(event) => setUsername(event.target.value)}
              placeholder="username"
            />
          </div>
          <div className="field grow">
            <label>Password</label>
            <input
              type="password"
              value={password}
              onChange={(event) => setPassword(event.target.value)}
              placeholder="password"
            />
          </div>
        </div>
        <div className="row auth-actions">
          <button onClick={register} disabled={!username || !password}>
            Create account
          </button>
          <Link className="ghost-link" to="/login">
            Уже есть аккаунт
          </Link>
        </div>
        {authError && <div className="error">{authError}</div>}
      </div>
    </div>
  );
}

function LoginPage() {
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [authError, setAuthError] = useState("");
  const navigate = useNavigate();
  const location = useLocation();

  useEffect(() => {
    async function check() {
      if (location.state?.loggedOut) {
        return;
      }
      const { token } = readAuth();
      if (token) {
        navigate("/dashboard", { replace: true });
        return;
      }
      const data = await tryRefreshToken();
      if (data) {
        navigate("/dashboard", { replace: true });
      }
    }
    check();
  }, [navigate, location]);

  async function login() {
    setAuthError("");
    const response = await fetch(`${API_BASE}/auth/login`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "X-Device-ID": getDeviceId(),
      },
      body: JSON.stringify({ username, password }),
      credentials: "include",
    });
    if (!response.ok) {
      setAuthError("Login failed");
      return;
    }
    const data = await response.json();
    saveAuth(data.token, data.username);
    navigate("/dashboard");
  }

  return (
    <div className="auth-page">
      <div className="auth-card">
        <h1>Login</h1>
        <div className="row">
          <div className="field grow">
            <label>Username</label>
            <input
              value={username}
              onChange={(event) => setUsername(event.target.value)}
              placeholder="username"
            />
          </div>
          <div className="field grow">
            <label>Password</label>
            <input
              type="password"
              value={password}
              onChange={(event) => setPassword(event.target.value)}
              placeholder="password"
            />
          </div>
        </div>
        <div className="row auth-actions">
          <button onClick={login} disabled={!username || !password}>
            Login
          </button>
          <Link className="ghost-link" to="/register">
            Создать аккаунт
          </Link>
        </div>
        {authError && <div className="error">{authError}</div>}
      </div>
    </div>
  );
}

function DashboardPage() {
  const [roomId, setRoomId] = useState("");
  const [wsUrl, setWsUrl] = useState("");
  const [serverUrl, setServerUrl] = useState(API_BASE);
  const [displayName, setDisplayName] = useState("");
  const [authError, setAuthError] = useState("");
  const [authUser, setAuthUser] = useState("");
  const [authToken, setAuthToken] = useState("");
  const [sessions, setSessions] = useState([]);
  const [sessionsError, setSessionsError] = useState("");
  const [sessionsLoading, setSessionsLoading] = useState(false);
  const navigate = useNavigate();
  const deviceId = getDeviceId();

  useEffect(() => {
    async function init() {
      let { token, user } = readAuth();
      if (!token) {
        const data = await tryRefreshToken();
        if (data) {
          token = data.token;
          user = data.username;
        }
      }

      if (token) {
        setAuthToken(token);
        setAuthUser(user);
      } else {
        navigate("/login", { replace: true });
      }
      setDisplayName(readDisplayName());
    }
    init();
  }, [navigate]);

  async function loadSessions() {
    if (!authToken) {
      return;
    }
    setSessionsLoading(true);
    setSessionsError("");
    try {
      const response = await fetch(`${serverUrl}/auth/sessions`, {
        headers: {
          Authorization: authToken ? `Bearer ${authToken}` : "",
        },
      });
      if (!response.ok) {
        setSessions([]);
        setSessionsError("Failed to load sessions");
        return;
      }
      const data = await response.json();
      setSessions(Array.isArray(data) ? data : []);
    } catch (err) {
      setSessions([]);
      setSessionsError("Failed to load sessions");
    } finally {
      setSessionsLoading(false);
    }
  }

  useEffect(() => {
    loadSessions();
  }, [authToken, serverUrl]);

  function formatSessionTime(value) {
    if (!value) {
      return "unknown";
    }
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) {
      return "unknown";
    }
    return date.toLocaleString();
  }

  async function createRoom() {
    const response = await fetch(`${serverUrl}/rooms`, {
      method: "POST",
      headers: {
        Authorization: authToken ? `Bearer ${authToken}` : "",
      },
    });
    if (!response.ok) {
      setAuthError("Create room requires login");
      navigate("/login");
      return;
    }
    const data = await response.json();
    setRoomId(data.roomId);
    setWsUrl(data.wsUrl);
  }

  function joinById() {
    if (!roomId.trim()) {
      return;
    }
    if (!displayName.trim()) {
      setAuthError("Enter display name before joining");
      return;
    }
    saveDisplayName(displayName.trim());
    const encoded = encodeURIComponent(roomId.trim());
    navigate(
      `/room/${encoded}?server=${encodeURIComponent(serverUrl)}&name=${encodeURIComponent(
        displayName.trim()
      )}`
    );
  }

  function joinByUrl() {
    if (!wsUrl.trim()) {
      return;
    }
    if (!displayName.trim()) {
      setAuthError("Enter display name before joining");
      return;
    }
    saveDisplayName(displayName.trim());
    navigate(
      `/room/custom?ws=${encodeURIComponent(wsUrl.trim())}&name=${encodeURIComponent(
        displayName.trim()
      )}`
    );
  }

  return (
    <div className="app">
      <header className="header">
        <h1>Chatter</h1>
        <div className="status status-disconnected">home</div>
      </header>

      <section className="panel">
        <div className="auth-row">
          <div className="auth-user">Logged in as {authUser}</div>
          <button
            className="ghost"
            onClick={() => {
              logout().then(() => navigate("/login", { state: { loggedOut: true } }));
            }}
          >
            Logout
          </button>
        </div>
        {authError && <div className="error">{authError}</div>}

        <div className="sessions">
          <div className="sessions-header">
            <div className="sessions-title">Active sessions</div>
            <button
              className="ghost small"
              onClick={loadSessions}
              disabled={sessionsLoading || !authToken}
            >
              Refresh
            </button>
          </div>
          {sessionsLoading && <div className="muted">Loading sessions...</div>}
          {sessionsError && <div className="error">{sessionsError}</div>}
          {!sessionsLoading && !sessionsError && sessions.length === 0 && (
            <div className="empty">No active sessions</div>
          )}
          {sessions.length > 0 && (
            <div className="sessions-list">
              {sessions.map((session) => {
                const isCurrent = session.deviceId === deviceId;
                return (
                  <div key={session.id} className="session-item">
                    <div className="session-main">
                      <div className="session-device">
                        {session.deviceId || "unknown device"}
                        {isCurrent && <span className="badge-current">Current device</span>}
                      </div>
                      <div className="session-meta">
                        <span>Last seen: {formatSessionTime(session.lastSeen)}</span>
                        <span>Expires: {formatSessionTime(session.expiresAt)}</span>
                      </div>
                    </div>
                  </div>
                );
              })}
            </div>
          )}
        </div>

        <div className="field">
          <label>Server URL</label>
          <input
            value={serverUrl}
            onChange={(event) => setServerUrl(event.target.value)}
            placeholder="http://localhost:8080"
          />
        </div>

        <div className="field">
          <label>Display name</label>
          <input
            value={displayName}
            onChange={(event) => {
              setDisplayName(event.target.value);
              saveDisplayName(event.target.value);
            }}
            placeholder="your name"
          />
        </div>

        <div className="row">
          <button onClick={createRoom}>Create room</button>
          <div className="field grow">
            <label>Room ID</label>
            <input
              value={roomId}
              onChange={(event) => setRoomId(event.target.value)}
              placeholder="room id"
            />
          </div>
          <button onClick={joinById} disabled={!roomId}>
            Join by ID
          </button>
        </div>

        <div className="row">
          <div className="field grow">
            <label>WS URL</label>
            <input
              value={wsUrl}
              onChange={(event) => setWsUrl(event.target.value)}
              placeholder="ws://localhost:8080/ws/..."
            />
          </div>
          <button onClick={joinByUrl} disabled={!wsUrl}>
            Join by WS URL
          </button>
        </div>
      </section>
    </div>
  );
}

function IndexRedirect() {
  const navigate = useNavigate();
  useEffect(() => {
    async function check() {
      const { token } = readAuth();
      if (token) {
        navigate("/dashboard", { replace: true });
        return;
      }
      const data = await tryRefreshToken();
      if (data) {
        navigate("/dashboard", { replace: true });
      } else {
        navigate("/register", { replace: true });
      }
    }
    check();
  }, [navigate]);
  return null;
}

function RoomPage() {
  const { roomId } = useParams();
  const [status, setStatus] = useState("disconnected");
  const [messages, setMessages] = useState([]);
  const [participants, setParticipants] = useState([]);
  const [displayNames, setDisplayNames] = useState({});
  const [text, setText] = useState("");
  const [localStream, setLocalStream] = useState(null);
  const [remoteStreams, setRemoteStreams] = useState([]);
  const [mediaError, setMediaError] = useState("");
  const [micEnabled, setMicEnabled] = useState(false);
  const [camEnabled, setCamEnabled] = useState(false);
  const [peerStatus, setPeerStatus] = useState([]);
  const socketRef = useRef(null);
  const peersRef = useRef(new Map());
  const pendingOffersRef = useRef(new Map());
  const [clientId, setClientId] = useState("");
  const [localName, setLocalName] = useState("");
  const [activeTab, setActiveTab] = useState("video");
  const navigate = useNavigate();

  const search = new URLSearchParams(window.location.search);
  const serverUrl = search.get("server") || API_BASE;
  const customWsUrl = search.get("ws");
  const nameParam = search.get("name");
  // If we have a custom WS URL, use it. Otherwise build one with roomId and auth token.
  // We need to read the token from localStorage (readAuth) because authToken state might not be set yet on initial render.
  const { token } = readAuth();
  let wsUrl = customWsUrl || buildWsUrl(serverUrl, roomId || "", token);
  if (customWsUrl && token) {
    // Append token to custom URL if not present
    if (!wsUrl.includes("token=")) {
      wsUrl += wsUrl.includes("?") ? `&token=${token}` : `?token=${token}`;
    }
  }

  function upsertParticipant(id) {
    setParticipants((prev) => {
      const exists = prev.some((p) => p.id === id);
      if (exists) {
        return prev.map((p) =>
          p.id === id ? { ...p, lastSeen: Date.now() } : p
        );
      }
      return [...prev, { id, lastSeen: Date.now() }];
    });
  }

  function removeParticipant(id) {
    setParticipants((prev) => prev.filter((p) => p.id !== id));
  }

  function setDisplayNameFor(id, name) {
    if (!id || !name) {
      return;
    }
    setDisplayNames((prev) => ({ ...prev, [id]: name }));
  }

  function displayNameFor(id) {
    if (id === clientId && localName) {
      return localName;
    }
    return displayNames[id] || "Anonymous";
  }

  function sendSignal(payload) {
    if (!socketRef.current || socketRef.current.readyState !== WebSocket.OPEN) {
      return;
    }
    socketRef.current.send(JSON.stringify(payload));
  }

  function removeRemoteStream(id) {
    setRemoteStreams((prev) => prev.filter((item) => item.id !== id));
  }

  function updatePeerStatus(id, status) {
    setPeerStatus((prev) => {
      const exists = prev.some((item) => item.id === id);
      if (exists) {
        return prev.map((item) => (item.id === id ? { id, status } : item));
      }
      return [...prev, { id, status }];
    });
  }

  function getPeerStatus(id) {
    return peerStatus.find((item) => item.id === id)?.status;
  }

  function attachLocalTracks(pc) {
    if (!localStream) {
      return;
    }
    const existing = pc.getSenders().map((sender) => sender.track?.id);
    localStream.getTracks().forEach((track) => {
      if (!existing.includes(track.id)) {
        pc.addTrack(track, localStream);
      }
    });
  }

  function createPeer(remoteId) {
    if (peersRef.current.has(remoteId)) {
      return peersRef.current.get(remoteId);
    }

    const pc = new RTCPeerConnection({ iceServers: ICE_SERVERS });
    peersRef.current.set(remoteId, pc);

    attachLocalTracks(pc);

    pc.onicecandidate = (event) => {
      if (event.candidate) {
        sendSignal({
          type: "webrtc",
          action: "ice",
          from: clientId,
          to: remoteId,
          candidate: event.candidate,
        });
      }
    };

    pc.ontrack = (event) => {
      const [stream] = event.streams;
      if (!stream) {
        return;
      }
      setRemoteStreams((prev) => {
        const exists = prev.some((item) => item.id === remoteId);
        if (exists) {
          return prev.map((item) => (item.id === remoteId ? { id: remoteId, stream } : item));
        }
        return [...prev, { id: remoteId, stream }];
      });
    };

    pc.onconnectionstatechange = () => {
      updatePeerStatus(remoteId, pc.connectionState);
      if (pc.connectionState === "failed" || pc.connectionState === "closed") {
        removeRemoteStream(remoteId);
      }
    };

    return pc;
  }

  async function sendOffer(remoteId) {
    const pc = createPeer(remoteId);
    const offer = await pc.createOffer();
    await pc.setLocalDescription(offer);
    sendSignal({
      type: "webrtc",
      action: "offer",
      from: clientId,
      to: remoteId,
      sdp: offer,
    });
  }

  function shouldInitiate(remoteId) {
    return clientId < remoteId;
  }

  function handleOffer(fromId, sdp) {
    const pc = createPeer(fromId);
    pc.setRemoteDescription(sdp)
      .then(() => pc.createAnswer())
      .then((answer) => pc.setLocalDescription(answer).then(() => answer))
      .then((answer) => {
        sendSignal({
          type: "webrtc",
          action: "answer",
          from: clientId,
          to: fromId,
          sdp: answer,
        });
      })
      .catch(() => {});
  }

  useEffect(() => {
    let isMounted = true;
    navigator.mediaDevices
      .getUserMedia({ video: true, audio: true })
      .then((stream) => {
        if (!isMounted) {
          return;
        }
        setLocalStream(stream);
        setMicEnabled(false);
        setCamEnabled(false);
      })
      .catch((err) => {
        if (!isMounted) {
          return;
        }
        setMediaError(err?.message || "Failed to access camera/microphone");
      });

    return () => {
      isMounted = false;
    };
  }, []);

  useEffect(() => {
    if (!localStream) {
      return;
    }
    const pending = pendingOffersRef.current;
    if (pending.size === 0) {
      return;
    }
    for (const [fromId, sdp] of pending.entries()) {
      handleOffer(fromId, sdp);
      pending.delete(fromId);
    }
  }, [localStream]);

  useEffect(() => {
    if (!localStream) {
      return;
    }
    localStream.getAudioTracks().forEach((track) => {
      track.enabled = micEnabled;
    });
  }, [localStream, micEnabled]);

  useEffect(() => {
    if (!localStream) {
      return;
    }
    localStream.getVideoTracks().forEach((track) => {
      track.enabled = camEnabled;
    });
  }, [localStream, camEnabled]);

  useEffect(() => {
    return () => {
      if (localStream) {
        localStream.getTracks().forEach((track) => track.stop());
      }
    };
  }, [localStream]);

  useEffect(() => {
    const name = (nameParam || readDisplayName()).trim();
    if (name) {
      setLocalName(name);
      saveDisplayName(name);
    }
  }, [nameParam]);

  useEffect(() => {
    if (!wsUrl) {
      return;
    }
    setStatus("connecting");
    const socket = new WebSocket(wsUrl);
    socketRef.current = socket;

    const sendProfile = () => {
      if (!localName) {
        return;
      }
      sendSignal({
        type: "profile",
        displayName: localName,
      });
    };

    socket.onopen = () => {
      setStatus("connected");
      sendProfile();
    };
    socket.onclose = () => setStatus("disconnected");
    socket.onerror = () => setStatus("error");
    socket.onmessage = (event) => {
      try {
        const payload = JSON.parse(event.data);
        if (payload?.type === "welcome" && payload.clientId) {
          setClientId(payload.clientId);
          upsertParticipant(payload.clientId);
          if (localName) {
            setDisplayNameFor(payload.clientId, localName);
            sendProfile();
          }
          return;
        }
        if (payload?.type === "participants" && Array.isArray(payload.participants)) {
          setParticipants(
            payload.participants.map((entry) => ({
              id: entry.id,
              lastSeen: Date.now(),
            }))
          );
          const mapped = {};
          payload.participants.forEach((entry) => {
            if (entry.displayName) {
              mapped[entry.id] = entry.displayName;
            }
          });
          if (Object.keys(mapped).length > 0) {
            setDisplayNames((prev) => ({ ...prev, ...mapped }));
          }
          sendProfile();
          return;
        }
        if (payload?.type === "profile" && payload.clientId && payload.displayName) {
          setDisplayNameFor(payload.clientId, payload.displayName);
          return;
        }
        if (payload?.type === "webrtc" && payload.from) {
          if (payload.to && clientId && payload.to !== clientId) {
            return;
          }
          if (payload.action === "offer" && payload.sdp) {
            if (!localStream) {
              pendingOffersRef.current.set(payload.from, payload.sdp);
              return;
            }
            handleOffer(payload.from, payload.sdp);
            return;
          }
          if (payload.action === "answer" && payload.sdp) {
            const pc = peersRef.current.get(payload.from);
            if (pc) {
              pc.setRemoteDescription(payload.sdp).catch(() => {});
            }
            return;
          }
          if (payload.action === "ice" && payload.candidate) {
            const pc = peersRef.current.get(payload.from);
            if (pc) {
              pc.addIceCandidate(payload.candidate).catch(() => {});
            }
            return;
          }
        }
        if (payload?.type === "presence") {
          if (payload.action === "join") {
            upsertParticipant(payload.clientId);
            sendProfile();
          } else if (payload.action === "leave") {
            removeParticipant(payload.clientId);
          }
        }
        if (payload?.displayName && payload?.clientId) {
          setDisplayNameFor(payload.clientId, payload.displayName);
        }
        if (payload?.clientId) {
          upsertParticipant(payload.clientId);
        }
        setMessages((prev) => [...prev, payload]);
      } catch {
        setMessages((prev) => [...prev, { type: "raw", text: event.data }]);
      }
    };

    return () => {
      socket.close();
    };
  }, [wsUrl]);

  useEffect(() => {
    if (!localStream || status !== "connected") {
      return;
    }

    for (const pc of peersRef.current.values()) {
      attachLocalTracks(pc);
    }

    const ids = participants
      .map((p) => p.id)
      .filter((id) => id && id !== clientId && clientId);
    ids.forEach((id) => {
      if (!peersRef.current.has(id) && shouldInitiate(id)) {
        sendOffer(id).catch(() => {});
      }
    });

    for (const [id, pc] of peersRef.current.entries()) {
      if (!ids.includes(id)) {
        pc.close();
        peersRef.current.delete(id);
        removeRemoteStream(id);
      }
    }
  }, [participants, clientId, localStream, status]);

  function sendMessage() {
    if (!socketRef.current || socketRef.current.readyState !== WebSocket.OPEN) {
      return;
    }
    if (!text.trim()) {
      return;
    }

    const payload = {
      type: "chat",
      clientId,
      displayName: localName || displayNameFor(clientId),
      text: text.trim(),
      ts: new Date().toISOString(),
    };
    upsertParticipant(clientId);
    socketRef.current.send(JSON.stringify(payload));
    setMessages((prev) => [...prev, payload]);
    setText("");
  }

  function leaveRoom() {
    if (socketRef.current) {
      socketRef.current.close();
    }
    for (const pc of peersRef.current.values()) {
      pc.close();
    }
    peersRef.current.clear();
    if (localStream) {
      localStream.getTracks().forEach((track) => track.stop());
    }
    setRemoteStreams([]);
    navigate("/");
  }

  if (!roomId && !customWsUrl) {
    return <Navigate to="/" replace />;
  }

  return (
    <div className={`room-layout tab-${activeTab}`}>
      {/* Mobile Tab Navigation */}
      <div className="mobile-nav">
        <button
          className={`nav-btn ${activeTab === "participants" ? "active" : ""}`}
          onClick={() => setActiveTab("participants")}
        >
          People
        </button>
        <button
          className={`nav-btn ${activeTab === "video" ? "active" : ""}`}
          onClick={() => setActiveTab("video")}
        >
          Video
        </button>
        <button
          className={`nav-btn ${activeTab === "chat" ? "active" : ""}`}
          onClick={() => setActiveTab("chat")}
        >
          Chat
        </button>
      </div>

      <aside className="sidebar">
        <div className="sidebar-header">
          <div className="sidebar-title">Room</div>
          <div className="sidebar-subtitle">{roomId || "custom"}</div>
          <div className={`status status-${status}`}>{status}</div>
        </div>

        <div className="participants">
          <div className="participants-title">Participants ({participants.length})</div>
          <div className="participants-list">
            {participants.length === 0 && (
              <div className="empty">No participants yet</div>
            )}
            {participants.map((participant) => (
              <div key={participant.id} className="participant">
                <span className="participant-name">{displayNameFor(participant.id)}</span>
                {participant.id === clientId && <span className="badge-me"> (You)</span>}
              </div>
            ))}
          </div>
        </div>

        <div className="sidebar-controls">
          <button
            className={`control-btn ${micEnabled ? "active" : "muted"}`}
            onClick={() => setMicEnabled((prev) => !prev)}
          >
            {micEnabled ? "Mute Mic" : "Unmute Mic"}
          </button>
          <button
            className={`control-btn ${camEnabled ? "active" : "muted"}`}
            onClick={() => setCamEnabled((prev) => !prev)}
          >
            {camEnabled ? "Stop Cam" : "Start Cam"}
          </button>
        </div>

        <div className="sidebar-footer">
          <button className="btn-leave" onClick={leaveRoom}>
            Leave Room
          </button>
        </div>
      </aside>

      <main className="stage">
        <header className="stage-header">
          <div className="field grow">
            <label>WebSocket URL</label>
            <input value={wsUrl} readOnly className="url-input" />
          </div>
        </header>

        {mediaError && <div className="error-banner">{mediaError}</div>}
        <div className="video-grid">
          <div className="video-wrapper local">
            <VideoTile
              stream={localStream}
              muted
              label={`${localName || displayNameFor(clientId)} (You)`}
            />
          </div>
          {remoteStreams.map((remote) => (
            <div key={remote.id} className="video-wrapper remote">
              <VideoTile
                key={remote.id}
                stream={remote.stream}
                label={
                  getPeerStatus(remote.id)
                    ? `${displayNameFor(remote.id)} · ${getPeerStatus(remote.id)}`
                    : displayNameFor(remote.id)
                }
              />
            </div>
          ))}
        </div>
      </main>

      <aside className="chat-panel">
        <div className="chat-header">Chat</div>
        <div className="messages">
          {messages.length === 0 && (
            <div className="empty-chat">No messages yet. Say hi!</div>
          )}
          {messages.map((msg, idx) => (
            <div
              key={`${msg.ts || "raw"}-${idx}`}
              className={`message ${
                msg.clientId === clientId ? "message-self" : "message-other"
              }`}
            >
              {msg.type === "chat" ? (
                <>
                  <div className="message-header">
                    <span className="sender-name">
                      {msg.displayName || displayNameFor(msg.clientId)}
                    </span>
                  </div>
                  <div className="message-content">{msg.text}</div>
                </>
              ) : (
                <div className="system-message">{msg.text || JSON.stringify(msg)}</div>
              )}
            </div>
          ))}
        </div>
        <div className="chat-input-area">
          <input
            className="chat-input"
            value={text}
            onChange={(event) => setText(event.target.value)}
            placeholder="Type a message..."
            onKeyDown={(event) => {
              if (event.key === "Enter") {
                sendMessage();
              }
            }}
          />
          <button className="btn-send" onClick={sendMessage} disabled={!text.trim()}>
            Send
          </button>
        </div>
      </aside>
    </div>
  );
}

export default function App() {
  return (
    <Routes>
      <Route path="/" element={<IndexRedirect />} />
      <Route path="/register" element={<RegisterPage />} />
      <Route path="/login" element={<LoginPage />} />
      <Route path="/dashboard" element={<DashboardPage />} />
      <Route path="/room/:roomId" element={<RoomPage />} />
      <Route path="/room/custom" element={<RoomPage />} />
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  );
}
