"use client";

import Link from "next/link";
import { useCallback, useEffect, useMemo, useState } from "react";
import { CompetitionRoom } from "@/lib/types";
import {
  loadOrCreateUserKey,
  loadStoredUsername,
  sanitizeUsername,
  saveStoredUsername,
  validateUsername,
} from "@/lib/username";

function normalizeCode(input: string): string {
  return input.toUpperCase().replace(/[^A-Z0-9]/g, "");
}

function formatMode(mode: string): string {
  if (mode === "time_based") return "Time-Based";
  if (mode === "questions_complete") return "Questions Complete";
  if (mode === "code_golf") return "Code Golf";
  return mode;
}

function formatDifficultyPolicy(policy: string): string {
  if (policy === "progressive") return "Progressive (Easy -> Medium -> Hard)";
  return policy.charAt(0).toUpperCase() + policy.slice(1);
}

async function readErrorMessage(response: Response): Promise<string> {
  try {
    const payload = (await response.json()) as { error?: string };
    if (payload.error) {
      return payload.error;
    }
  } catch {
    // fall through
  }
  return `Request failed (${response.status})`;
}

type Props = {
  roomCode: string;
};

export function CompetitionRoomView({ roomCode }: Props) {
  const normalizedCode = useMemo(() => normalizeCode(roomCode), [roomCode]);
  const [userKey, setUserKey] = useState("");
  const [room, setRoom] = useState<CompetitionRoom | null>(null);
  const [username, setUsername] = useState("");
  const [usernameDraft, setUsernameDraft] = useState("");
  const [usernameModalOpen, setUsernameModalOpen] = useState(false);
  const [usernameError, setUsernameError] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(true);
  const [isJoining, setIsJoining] = useState(false);
  const [copiedCode, setCopiedCode] = useState(false);

  const activeHandle = useMemo(() => sanitizeUsername(username), [username]);
  const activeUserKey = useMemo(() => userKey.trim(), [userKey]);

  const loadRoom = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const response = await fetch(`/api/backend/v1/competitions/rooms/${normalizedCode}`, {
        cache: "no-store",
      });
      if (!response.ok) {
        throw new Error(await readErrorMessage(response));
      }
      const payload = (await response.json()) as CompetitionRoom;
      setRoom(payload);
    } catch (loadErr) {
      setError(loadErr instanceof Error ? loadErr.message : "Failed to load room");
    } finally {
      setLoading(false);
    }
  }, [normalizedCode]);

  useEffect(() => {
    setUserKey(loadOrCreateUserKey());
    const saved = loadStoredUsername();
    if (saved) {
      setUsername(saved);
      setUsernameDraft(saved);
    } else {
      setUsernameModalOpen(true);
    }
  }, []);

  useEffect(() => {
    if (!usernameModalOpen) {
      return;
    }
    const onKeyDown = (event: KeyboardEvent): void => {
      if (event.key === "Escape" && activeHandle) {
        setUsernameModalOpen(false);
      }
    };
    window.addEventListener("keydown", onKeyDown);
    return () => window.removeEventListener("keydown", onKeyDown);
  }, [activeHandle, usernameModalOpen]);

  useEffect(() => {
    void loadRoom();
    const timer = window.setInterval(() => {
      void loadRoom();
    }, 3000);
    return () => window.clearInterval(timer);
  }, [loadRoom]);

  const joined = useMemo(() => {
    if (!room) return false;
    return room.members.some((member) => member.handle.toLowerCase() === activeHandle.toLowerCase());
  }, [room, activeHandle]);

  const handleJoin = async (): Promise<void> => {
    if (!activeHandle || !activeUserKey) {
      setUsernameModalOpen(true);
      return;
    }
    setIsJoining(true);
    setError("");
    try {
      const response = await fetch("/api/backend/v1/competitions/rooms/join", {
        method: "POST",
        headers: {
          "content-type": "application/json",
          "x-user-handle": activeHandle,
          "x-user-key": activeUserKey,
        },
        body: JSON.stringify({ code: normalizedCode }),
      });
      if (!response.ok) {
        throw new Error(await readErrorMessage(response));
      }
      const payload = (await response.json()) as CompetitionRoom;
      setRoom(payload);
    } catch (joinErr) {
      setError(joinErr instanceof Error ? joinErr.message : "Failed to join room");
    } finally {
      setIsJoining(false);
    }
  };

  const handleCopyCode = async (): Promise<void> => {
    try {
      await navigator.clipboard.writeText(normalizedCode);
      setCopiedCode(true);
      window.setTimeout(() => setCopiedCode(false), 1200);
    } catch {
      setError("Failed to copy room code");
    }
  };

  const openUsernameModal = (): void => {
    setUsernameDraft(activeHandle || "");
    setUsernameError("");
    setUsernameModalOpen(true);
  };

  const saveUsername = (): void => {
    const cleaned = sanitizeUsername(usernameDraft);
    const validation = validateUsername(cleaned);
    if (validation) {
      setUsernameError(validation);
      return;
    }
    const stored = saveStoredUsername(cleaned);
    setUsername(stored);
    setUsernameDraft(stored);
    setUsernameError("");
    setUsernameModalOpen(false);
  };

  if (loading && !room) {
    return (
      <section className="competition-card">
        <p>Loading room...</p>
      </section>
    );
  }

  return (
    <div className="competition-layout">
      <section className="competition-card">
        <div className="competition-card-head">
          <div>
            <p className="eyebrow">Room</p>
            <h2>{room?.name ?? normalizedCode}</h2>
          </div>
          <button className="btn btn-muted competition-room-copy" type="button" onClick={() => void handleCopyCode()}>
            {copiedCode ? "Copied" : "Copy Code"}
          </button>
        </div>

        <div className="competition-code-hero">
          <span>Room Code</span>
          <code>{normalizedCode}</code>
        </div>

        <p>
          Host: <strong>{room?.hostHandle ?? "unknown"}</strong> · Status:{" "}
          <strong>{room?.status ?? "unknown"}</strong>
        </p>
        <p>
          {room ? (
            <>
              {formatMode(room.mode)} · {room.questionCount} problems ·{" "}
              {formatDifficultyPolicy(room.difficultyPolicy)}
            </>
          ) : null}
        </p>
        <p>
          Host model: participants join the same room code and race under host-selected rules. Live state updates
          are polled every few seconds in this MVP.
        </p>

        <div className="competition-identity">
          <span className="competition-identity-label">Username</span>
          <button className="competition-user-chip competition-user-chip-button" type="button" onClick={openUsernameModal}>
            {activeHandle || "Set Username"}
          </button>
        </div>
        <div className="competition-actions">
          <button className="btn btn-primary" type="button" disabled={joined || isJoining || !activeHandle} onClick={() => void handleJoin()}>
            {joined ? "Joined" : isJoining ? "Joining..." : "Join This Room"}
          </button>
          <button className="btn btn-muted" type="button" onClick={() => void loadRoom()}>
            Refresh
          </button>
          <Link className="btn btn-muted" href="/competitions">
            Back to Competitions
          </Link>
        </div>

        {error ? <p className="competition-error">{error}</p> : null}
      </section>

      <section className="competition-card">
        <h3>Participants ({room?.members.length ?? 0})</h3>
        <div className="competition-member-list">
          {room?.members.map((member) => (
            <article key={`${member.userId}-${member.handle}`} className="competition-member-item">
              <div>
                <strong>{member.handle}</strong>
                {member.isHost ? <span className="competition-host-pill">Host</span> : null}
              </div>
              <small>joined {new Date(member.joinedAt).toLocaleString()}</small>
            </article>
          ))}
        </div>
      </section>

      {usernameModalOpen ? (
        <div
          className="competition-modal-backdrop"
          onClick={() => {
            if (activeHandle) {
              setUsernameModalOpen(false);
            }
          }}
          role="presentation"
        >
          <section
            className="competition-modal"
            role="dialog"
            aria-modal="true"
            aria-labelledby="room-username-modal-title"
            onClick={(event) => event.stopPropagation()}
          >
            <h3 id="room-username-modal-title">{activeHandle ? "Change Username" : "Set Username"}</h3>
            <p>Use one username for all rooms. You can change it later.</p>
            <label className="competition-modal-field">
              Username
              <input
                value={usernameDraft}
                onChange={(event) => setUsernameDraft(event.target.value)}
                placeholder="your_username"
                autoFocus
                maxLength={24}
              />
            </label>
            {usernameError ? <p className="competition-modal-error">{usernameError}</p> : null}
            <div className="competition-modal-actions">
              {activeHandle ? (
                <button className="btn btn-muted" type="button" onClick={() => setUsernameModalOpen(false)}>
                  Cancel
                </button>
              ) : null}
              <button className="btn btn-primary" type="button" onClick={saveUsername}>
                Save Username
              </button>
            </div>
          </section>
        </div>
      ) : null}
    </div>
  );
}
