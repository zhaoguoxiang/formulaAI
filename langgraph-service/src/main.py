"""
FastAPI entry point for the LangGraph Agent service.

Provides:
  - POST /api/chat     SSE streaming endpoint (compatible with Angular chat-panel)
  - GET  /api/health   Health check
"""

from __future__ import annotations

import logging
import os

from fastapi import FastAPI, Request
from fastapi.responses import JSONResponse
from sse_starlette.sse import EventSourceResponse

from .agent import run_agent_stream
from .sse_protocol import format_sse_stream

logging.basicConfig(level=logging.INFO, format="%(asctime)s [%(levelname)s] %(message)s")
logger = logging.getLogger(__name__)

app = FastAPI(title="FormulAI LangGraph Agent", version="0.1.0")


@app.get("/api/health")
async def health():
    return {"status": "ok", "service": "langgraph-agent"}


@app.post("/api/chat")
async def chat(request: Request):
    """SSE streaming chat endpoint.

    Request body (JSON):
        {
            "messages": [{"role":"user","content":"分析填料对硬度的影响"}],
            "project_id": "uuid-of-current-workspace"
        }

    Response: SSE stream with events matching the Angular chat-panel protocol.
    """
    try:
        body = await request.json()
    except Exception:
        return JSONResponse(
            status_code=400,
            content={"error": "invalid JSON body"},
        )

    messages = body.get("messages", [])
    if not messages:
        return JSONResponse(
            status_code=400,
            content={"error": "messages array is required"},
        )

    project_id = body.get("project_id", "")

    logger.info("chat request received, %d messages, project=%s", len(messages), project_id or "none")

    async def event_generator():
        agent_stream = run_agent_stream(messages, project_id=project_id)
        async for sse_line in format_sse_stream(agent_stream):
            yield sse_line

    return EventSourceResponse(event_generator())


if __name__ == "__main__":
    import uvicorn

    port = int(os.getenv("PORT", "5050"))
    uvicorn.run(app, host="0.0.0.0", port=port)
