"use client";

import { useRouter } from "next/navigation";
import { useState } from "react";

export function DevResetButton() {
  const router = useRouter();
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");

  const reset = async (): Promise<void> => {
    setBusy(true);
    setError("");
    try {
      const response = await fetch("/api/backend/v1/dev/reset", {
        method: "POST",
      });
      if (!response.ok) {
        throw new Error(`Reset failed (${response.status})`);
      }
      router.refresh();
    } catch (resetErr) {
      setError(resetErr instanceof Error ? resetErr.message : "Reset failed");
    } finally {
      setBusy(false);
    }
  };

  return (
    <div className="dev-reset">
      <button className="btn btn-muted" type="button" disabled={busy} onClick={() => void reset()}>
        {busy ? "Resetting..." : "Reset DB"}
      </button>
      {error ? <span>{error}</span> : null}
    </div>
  );
}
