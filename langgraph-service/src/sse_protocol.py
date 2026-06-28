"""
SSE protocol helpers compatible with the Angular chat-panel component.

The frontend expects Server-Sent Events with JSON data in this format:

    data: {"type":"tool_call","name":"get_formula_detail"}

    data: {"type":"text","content":"根据分析..."}

    data: {"type":"done","analysis_id":"abc123"}

    data: {"type":"error","message":"something went wrong"}

See: frontend/src/app/components/chat-panel/chat-panel.component.ts
"""

from __future__ import annotations

import json
import uuid
from collections.abc import AsyncGenerator
from typing import Any


def sse_event(event_type: str, **kwargs: Any) -> str:
    """Format a single SSE data line with JSON payload."""
    payload = {"type": event_type, **kwargs}
    return f"data: {json.dumps(payload, ensure_ascii=False)}\n\n"


async def format_sse_stream(
    agent_stream: AsyncGenerator[dict[str, Any], None],
    *,
    tool_timeout: float = 30.0,
) -> AsyncGenerator[str, None]:
    """Wrap a LangGraph agent stream into SSE events the chat-panel understands.

    Expected agent stream items:
        {"type": "tool_call", "name": "get_xxx", "args": {...}}
        {"type": "text", "content": "..."}         # chunked LLM output
        {"type": "done", "analysis_id": "..."}     # optional
        {"type": "error", "message": "..."}

    Every chunk is forwarded as an SSE data line.
    """

    analysis_id: str | None = None

    async for chunk in agent_stream:
        chunk_type = chunk.get("type", "")

        if chunk_type == "tool_call":
            yield sse_event("tool_call", name=chunk["name"])

        elif chunk_type == "text":
            yield sse_event("text", content=chunk["content"])

        elif chunk_type == "done":
            analysis_id = chunk.get("analysis_id")
            yield sse_event("done", analysis_id=analysis_id or "")

        elif chunk_type == "error":
            yield sse_event("error", message=chunk.get("message", "Unknown error"))
            return  # stop stream on error

        else:
            # unknown chunk type – treat as text for resilience
            text = chunk.get("content", "")
            if text:
                yield sse_event("text", content=str(text))

    # Ensure a final 'done' event if the agent didn't send one
    if analysis_id is None:
        yield sse_event("done", analysis_id=str(uuid.uuid4()))
