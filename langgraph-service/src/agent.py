"""
LangGraph ReAct agent for formula analysis.

The agent has access to the Go backend via tools (analysis endpoints, formula CRUD)
and streams its reasoning back as structured chunks for SSE conversion.
"""

from __future__ import annotations

import os
import uuid
from collections.abc import AsyncGenerator
from typing import Any

from langchain_openai import ChatOpenAI
from langgraph.prebuilt import create_react_agent
from langgraph.checkpoint.memory import MemorySaver

from .tools import ALL_TOOLS

SYSTEM_PROMPT = """你是 FormulAI 配方分析助手，专精于化工/材料配方的数据分析。

## 能力
- 查询配方数据库，获取物料分布、组分模式、工艺步骤等统计信息
- 对比不同配方的组成和工艺差异
- 分析物料使用趋势和配方优化方向

## 规则
1. **必须先查数据再回答**。如果用户问的是配方相关分析问题，先用工具获取真实数据，再基于数据给出结论。
2. 所有回答用中文，专业术语保留英文对照。
3. 数据不足时，如实说明并给出建议方向。
4. 回答简洁、结构化，用 Markdown 格式（表格、列表等）提升可读性。
5. 不要编造任何数据，所有数据必须来自工具调用的结果。
6. 工具调用不超过 5 轮，之后必须基于已有数据给出最佳回答。
"""

MAX_TOOL_ROUNDS = 5


def create_agent() -> Any:
    """Create a LangGraph ReAct agent wired to the Go backend tools."""
    llm = ChatOpenAI(
        model=os.getenv("LLM_MODEL", "gpt-4o"),
        temperature=0.3,
        streaming=True,
        timeout=60.0,
    )

    # MemorySaver keeps conversation state across tool calls within one request
    memory = MemorySaver()

    return create_react_agent(
        model=llm,
        tools=ALL_TOOLS,
        prompt=SYSTEM_PROMPT,
        checkpointer=memory,
    )


async def run_agent_stream(
    messages: list[dict[str, str]],
    project_id: str = "",
) -> AsyncGenerator[dict[str, Any], None]:
    """Run the agent on a conversation and yield structured chunks.

    Args:
        messages: Chat messages from the frontend.
        project_id: Current workspace UUID for scoping tool calls.

    Chunk format:
        {"type": "tool_call", "name": "get_ingredient_distribution"}
        {"type": "text", "content": "..."}
        {"type": "done", "analysis_id": "..."}
        {"type": "error", "message": "..."}
    """
    agent = create_agent()
    config = {"configurable": {"thread_id": str(uuid.uuid4())}}
    analysis_id = str(uuid.uuid4())

    # Build LangChain message format from the chat-panel messages
    user_messages = [m for m in messages if m.get("role") == "user"]
    if not user_messages:
        yield {"type": "error", "message": "没有收到用户消息"}
        return

    last_user = user_messages[-1]["content"]

    # Inject project_id into the user message so tools can use it
    if project_id:
        last_user = (
            f"[当前工作空间ID: {project_id}]\n\n{last_user}"
        )

    tool_rounds = 0
    try:
        async for event in agent.astream_events(
            {"messages": [("user", last_user)]},
            config=config,
            version="v2",
        ):
            kind = event.get("event", "")

            # ── tool call started ──
            if kind == "on_tool_start":
                tool_rounds += 1
                if tool_rounds > MAX_TOOL_ROUNDS:
                    break
                name = event.get("name", "unknown_tool")
                yield {"type": "tool_call", "name": name}

            # ── LLM streaming tokens ──
            elif kind == "on_chat_model_stream":
                chunk_data = event.get("data", {}).get("chunk")
                if chunk_data and hasattr(chunk_data, "content"):
                    content = chunk_data.content
                    if isinstance(content, str) and content:
                        yield {"type": "text", "content": content}

        # Signal completion
        yield {"type": "done", "analysis_id": analysis_id}

    except Exception as exc:
        yield {"type": "error", "message": f"Agent 执行异常: {exc}"}
