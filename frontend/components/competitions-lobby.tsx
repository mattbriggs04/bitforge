"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useCallback, useEffect, useMemo, useState } from "react";
import { CompetitionRoom } from "@/lib/types";
import {
  loadOrCreateUserKey,
  loadStoredUsername,
  sanitizeUsername,
  saveStoredUsername,
  validateUsername,
} from "@/lib/username";

type RoomMode = "time_based" | "questions_complete" | "code_golf";
type DifficultyPolicy = "easy" | "medium" | "hard" | "random" | "progressive";

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
    // Ignore parse errors and fallback to status text.
  }
  return `Request failed (${response.status})`;
}

export function CompetitionsLobby() {
  const router = useRouter();

  const [userKey, setUserKey] = useState("");
  const [username, setUsername] = useState("");
  const [usernameDraft, setUsernameDraft] = useState("");
  const [usernameModalOpen, setUsernameModalOpen] = useState(false);
  const [usernameError, setUsernameError] = useState("");
  const [roomName, setRoomName] = useState("");
  const [joinCode, setJoinCode] = useState("");
  const [mode, setMode] = useState<RoomMode>("time_based");
  const [questionCount, setQuestionCount] = useState(5);
  const [difficultyPolicy, setDifficultyPolicy] = useState<DifficultyPolicy>("random");

  const [rooms, setRooms] = useState<CompetitionRoom[]>([]);
  const [loadingRooms, setLoadingRooms] = useState(true);
  const [isCreating, setIsCreating] = useState(false);
  const [isJoining, setIsJoining] = useState(false);
  const [deletingCode, setDeletingCode] = useState("");
  const [deleteModalRoom, setDeleteModalRoom] = useState<CompetitionRoom | null>(null);
  const [copiedCode, setCopiedCode] = useState("");
  const [error, setError] = useState("");

  const activeHandle = useMemo(() => sanitizeUsername(username), [username]);
  const activeUserKey = useMemo(() => userKey.trim(), [userKey]);

  const fetchMyRooms = useCallback(async () => {
    if (!activeHandle || !activeUserKey) {
      setRooms([]);
      setLoadingRooms(false);
      return;
    }
    setLoadingRooms(true);
    setError("");
    try {
      const response = await fetch("/api/backend/v1/competitions/rooms", {
        cache: "no-store",
        headers: {
          "x-user-handle": activeHandle,
          "x-user-key": activeUserKey,
        },
      });
      if (!response.ok) {
        throw new Error(await readErrorMessage(response));
      }
      const payload = (await response.json()) as { items?: CompetitionRoom[] };
      setRooms(Array.isArray(payload.items) ? payload.items : []);
    } catch (fetchErr) {
      setError(fetchErr instanceof Error ? fetchErr.message : "Failed to load rooms");
    } finally {
      setLoadingRooms(false);
    }
  }, [activeHandle, activeUserKey]);

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
    void fetchMyRooms();
  }, [fetchMyRooms]);

  useEffect(() => {
    if (!deleteModalRoom && !usernameModalOpen) {
      return;
    }
    const onKeyDown = (event: KeyboardEvent): void => {
      if (event.key === "Escape") {
        if (deleteModalRoom) {
          setDeleteModalRoom(null);
        } else if (usernameModalOpen && activeHandle) {
          setUsernameModalOpen(false);
        }
      }
    };
    window.addEventListener("keydown", onKeyDown);
    return () => window.removeEventListener("keydown", onKeyDown);
  }, [activeHandle, deleteModalRoom, usernameModalOpen]);

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

  const handleCreateRoom = async (): Promise<void> => {
    if (!activeHandle || !activeUserKey) {
      openUsernameModal();
      return;
    }
    setIsCreating(true);
    setError("");
    try {
      const response = await fetch("/api/backend/v1/competitions/rooms", {
        method: "POST",
        headers: {
          "content-type": "application/json",
          "x-user-handle": activeHandle,
          "x-user-key": activeUserKey,
        },
        body: JSON.stringify({
          name: roomName,
          mode,
          questionCount,
          difficultyPolicy,
        }),
      });
      if (!response.ok) {
        throw new Error(await readErrorMessage(response));
      }
      const room = (await response.json()) as CompetitionRoom;
      router.push(`/competitions/${room.code}`);
    } catch (createErr) {
      setError(createErr instanceof Error ? createErr.message : "Failed to create room");
    } finally {
      setIsCreating(false);
    }
  };

  const handleJoinRoom = async (): Promise<void> => {
    if (!activeHandle || !activeUserKey) {
      openUsernameModal();
      return;
    }
    const code = normalizeCode(joinCode);
    if (!code) {
      setError("Enter a room code");
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
        body: JSON.stringify({ code }),
      });
      if (!response.ok) {
        throw new Error(await readErrorMessage(response));
      }
      const room = (await response.json()) as CompetitionRoom;
      router.push(`/competitions/${room.code}`);
    } catch (joinErr) {
      setError(joinErr instanceof Error ? joinErr.message : "Failed to join room");
    } finally {
      setIsJoining(false);
    }
  };

  const handleDeleteRoom = async (): Promise<void> => {
    if (!deleteModalRoom) {
      return;
    }
    if (!activeHandle || !activeUserKey) {
      openUsernameModal();
      return;
    }

    const code = deleteModalRoom.code;
    setDeletingCode(code);
    setError("");
    try {
      let response = await fetch(`/api/backend/v1/competitions/rooms/${code}`, {
        method: "DELETE",
        headers: {
          "x-user-handle": activeHandle,
          "x-user-key": activeUserKey,
        },
      });
      if (response.status === 405) {
        response = await fetch(`/api/backend/v1/competitions/rooms/${code}/delete`, {
          method: "POST",
          headers: {
            "x-user-handle": activeHandle,
            "x-user-key": activeUserKey,
          },
        });
      }
      if (!response.ok) {
        throw new Error(await readErrorMessage(response));
      }
      setRooms((previous) => previous.filter((room) => room.code !== code));
      setDeleteModalRoom(null);
    } catch (deleteErr) {
      setError(deleteErr instanceof Error ? deleteErr.message : "Failed to delete room");
    } finally {
      setDeletingCode("");
    }
  };

  const handleCopyCode = async (code: string): Promise<void> => {
    try {
      await navigator.clipboard.writeText(code);
      setCopiedCode(code);
      window.setTimeout(() => {
        setCopiedCode((current) => (current === code ? "" : current));
      }, 1200);
    } catch {
      setError("Failed to copy room code");
    }
  };

  return (
    <div className="competition-layout">
      <section className="competition-card">
        <h2>Competitions with Friends</h2>
        <p>
          Create a room, share the code, and race through systems problems. The host configures mode,
          question count, and difficulty policy.
        </p>
        <div className="competition-identity">
          <span className="competition-identity-label">Username</span>
          <button className="competition-user-chip competition-user-chip-button" type="button" onClick={openUsernameModal}>
            {activeHandle || "Set Username"}
          </button>
        </div>
      </section>

      <section className="competition-grid">
        <article className="competition-card">
          <h3>Create Room</h3>
          <label>
            Room Name (optional)
            <input
              value={roomName}
              onChange={(event) => setRoomName(event.target.value)}
              placeholder="Friday Firmware Sprint"
            />
          </label>

          <label>
            Mode
            <select value={mode} onChange={(event) => setMode(event.target.value as RoomMode)}>
              <option value="time_based">Time-Based</option>
              <option value="questions_complete">Questions Complete</option>
              <option value="code_golf">Code Golf</option>
            </select>
          </label>

          <label>
            Number of Questions
            <input
              type="number"
              min={1}
              max={100}
              value={questionCount}
              onChange={(event) => setQuestionCount(Number(event.target.value) || 1)}
            />
          </label>

          <label>
            Difficulty Policy
            <select
              value={difficultyPolicy}
              onChange={(event) => setDifficultyPolicy(event.target.value as DifficultyPolicy)}
            >
              <option value="easy">Easy</option>
              <option value="medium">Medium</option>
              <option value="hard">Hard</option>
              <option value="random">Random</option>
              <option value="progressive">Progressive (Easy -&gt; Medium -&gt; Hard)</option>
            </select>
          </label>

          <button
            className="btn btn-primary"
            type="button"
            disabled={isCreating || isJoining}
            onClick={() => void handleCreateRoom()}
          >
            {isCreating ? "Creating..." : "Create Room"}
          </button>
        </article>

        <article className="competition-card">
          <h3>Join Room</h3>
          <label>
            Room Code
            <input
              className="room-code-input"
              value={joinCode}
              onChange={(event) => setJoinCode(normalizeCode(event.target.value))}
              autoCapitalize="characters"
              autoCorrect="off"
              spellCheck={false}
              maxLength={8}
              placeholder="XXXXXX"
            />
          </label>
          <button
            className="btn btn-muted"
            type="button"
            disabled={isJoining || isCreating}
            onClick={() => void handleJoinRoom()}
          >
            {isJoining ? "Joining..." : "Join Room"}
          </button>
        </article>
      </section>

      {error ? <p className="competition-error">{error}</p> : null}

      <section className="competition-card">
        <div className="competition-card-head">
          <h3>Your Rooms</h3>
          <button className="btn btn-muted" type="button" onClick={() => void fetchMyRooms()} disabled={loadingRooms}>
            Refresh
          </button>
        </div>

        {loadingRooms ? <p>Loading rooms...</p> : null}
        {!loadingRooms && rooms.length === 0 ? <p>No rooms yet. Create one or join by code.</p> : null}

        {!loadingRooms && rooms.length > 0 ? (
          <div className="competition-room-list">
            {rooms.map((room) => (
              <article key={room.id} className="competition-room-card">
                <div className="competition-room-head">
                  <h4>{room.name}</h4>
                </div>
                <p>
                  Host: <strong>{room.hostHandle}</strong>
                </p>
                <p>
                  {formatMode(room.mode)} · {room.questionCount} problems ·{" "}
                  {formatDifficultyPolicy(room.difficultyPolicy)}
                </p>
                <div className="competition-room-code-wrap">
                  <span className="competition-room-code-label">Room Code</span>
                  <code className="competition-room-code">{room.code}</code>
                  <button
                    className="btn btn-muted competition-room-copy"
                    type="button"
                    onClick={() => void handleCopyCode(room.code)}
                  >
                    {copiedCode === room.code ? "Copied" : "Copy"}
                  </button>
                </div>
                <div className="competition-room-actions">
                  <Link href={`/competitions/${room.code}`} className="btn btn-primary">
                    Open Room
                  </Link>
                  {activeHandle && room.hostHandle.toLowerCase() === activeHandle.toLowerCase() ? (
                    <button
                      className="btn btn-muted competition-room-delete"
                      type="button"
                      disabled={deletingCode === room.code}
                      onClick={() => setDeleteModalRoom(room)}
                    >
                      {deletingCode === room.code ? "Deleting..." : "Delete Room"}
                    </button>
                  ) : null}
                </div>
              </article>
            ))}
          </div>
        ) : null}
      </section>

      {deleteModalRoom ? (
        <div className="competition-modal-backdrop" onClick={() => setDeleteModalRoom(null)} role="presentation">
          <section
            className="competition-modal"
            role="dialog"
            aria-modal="true"
            aria-labelledby="delete-room-title"
            aria-describedby="delete-room-copy"
            onClick={(event) => event.stopPropagation()}
          >
            <h3 id="delete-room-title">Delete Room</h3>
            <p id="delete-room-copy">
              Delete <strong>{deleteModalRoom.name}</strong> ({deleteModalRoom.code}) for all participants?
              This action cannot be undone.
            </p>
            <div className="competition-modal-actions">
              <button className="btn btn-muted" type="button" onClick={() => setDeleteModalRoom(null)}>
                Cancel
              </button>
              <button
                className="btn btn-muted competition-room-delete"
                type="button"
                disabled={deletingCode === deleteModalRoom.code}
                onClick={() => void handleDeleteRoom()}
              >
                {deletingCode === deleteModalRoom.code ? "Deleting..." : "Delete Room"}
              </button>
            </div>
          </section>
        </div>
      ) : null}

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
            aria-labelledby="username-modal-title"
            onClick={(event) => event.stopPropagation()}
          >
            <h3 id="username-modal-title">{activeHandle ? "Change Username" : "Set Username"}</h3>
            <p>Use one username for all rooms. You can update it later from this screen.</p>
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
