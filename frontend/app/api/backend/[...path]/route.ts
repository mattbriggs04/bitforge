import { NextRequest } from "next/server";

const BACKEND_API_URL =
  process.env.BACKEND_API_URL ?? process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080";

export const dynamic = "force-dynamic";

type RouteContext = {
  params: Promise<{ path: string[] }>;
};

async function sleep(ms: number): Promise<void> {
  await new Promise((resolve) => setTimeout(resolve, ms));
}

async function fetchUpstreamWithRetry(url: string, init: RequestInit): Promise<Response> {
  const attempts = 3;
  let lastError: unknown;

  for (let attempt = 1; attempt <= attempts; attempt += 1) {
    try {
      const response = await fetch(url, init);
      if (response.status >= 500 && attempt < attempts) {
        await sleep(100 * attempt);
        continue;
      }
      return response;
    } catch (error) {
      lastError = error;
      if (attempt < attempts) {
        await sleep(100 * attempt);
      }
    }
  }

  throw lastError instanceof Error ? lastError : new Error("upstream fetch failed");
}

async function proxy(request: NextRequest, context: RouteContext): Promise<Response> {
  const { path } = await context.params;
  const endpointPath = path.join("/");
  const targetURL = new URL(`/api/${endpointPath}`, BACKEND_API_URL);

  request.nextUrl.searchParams.forEach((value, key) => {
    targetURL.searchParams.append(key, value);
  });

  const headers = new Headers();
  const contentType = request.headers.get("content-type");
  const userHandle = request.headers.get("x-user-handle");
  const userKey = request.headers.get("x-user-key");
  if (contentType) headers.set("content-type", contentType);
  if (userHandle) headers.set("x-user-handle", userHandle);
  if (userKey) headers.set("x-user-key", userKey);

  const method = request.method.toUpperCase();
  const body = method === "GET" || method === "HEAD" ? undefined : await request.text();

  let upstream: Response;
  try {
    upstream = await fetchUpstreamWithRetry(targetURL.toString(), {
      method,
      headers,
      body,
      cache: "no-store",
    });
  } catch {
    return Response.json({ error: "backend unavailable, retry shortly" }, { status: 503 });
  }

  return new Response(upstream.body, {
    status: upstream.status,
    headers: {
      "content-type": upstream.headers.get("content-type") ?? "application/json",
    },
  });
}

export async function GET(request: NextRequest, context: RouteContext): Promise<Response> {
  return proxy(request, context);
}

export async function POST(request: NextRequest, context: RouteContext): Promise<Response> {
  return proxy(request, context);
}

export async function PUT(request: NextRequest, context: RouteContext): Promise<Response> {
  return proxy(request, context);
}

export async function DELETE(request: NextRequest, context: RouteContext): Promise<Response> {
  return proxy(request, context);
}

export async function OPTIONS(request: NextRequest, context: RouteContext): Promise<Response> {
  return proxy(request, context);
}
