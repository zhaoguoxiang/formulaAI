"""
LangChain tools that call the Go backend analysis & formula APIs.

Each tool fetches real data from the backend and returns JSON for the LLM to reason over.
All calls include an X-Project-Id header to scope data to the current workspace.
"""

from __future__ import annotations

import json
import os
from typing import Any

import httpx
from langchain_core.tools import tool

BACKEND_URL = os.getenv("GO_BACKEND_URL", "http://backend:8080")
REQUEST_TIMEOUT = 15.0


# ── helper ────────────────────────────────────────────────────────────────────

def _get(path: str, project_id: str) -> list[dict[str, Any]]:
    """GET a JSON array from the Go backend with X-Project-Id header."""
    url = f"{BACKEND_URL}{path}"
    headers = {"X-Project-Id": project_id}
    try:
        resp = httpx.get(url, headers=headers, timeout=REQUEST_TIMEOUT)
        resp.raise_for_status()
        return resp.json()
    except httpx.HTTPError as exc:
        return [{"error": f"backend request failed: {exc}"}]


# ── analysis tools ────────────────────────────────────────────────────────────

@tool
def get_ingredient_distribution(project_id: str) -> str:
    """获取当前工作空间中所有配方里每种物料的使用频率和平均用量百分比。"""
    data = _get("/api/analysis/ingredient-distribution", project_id)
    return json.dumps(data, ensure_ascii=False, indent=2)


@tool
def get_component_mode_ratio(project_id: str) -> str:
    """获取当前工作空间中单组分与双组分配方的数量及占比。"""
    data = _get("/api/analysis/component-mode-ratio", project_id)
    return json.dumps(data, ensure_ascii=False, indent=2)


@tool
def get_step_count_distribution(project_id: str) -> str:
    """获取当前工作空间中配方工艺步骤数量的分布（分桶: 1-2, 3-5, 5+）。"""
    data = _get("/api/analysis/step-count-distribution", project_id)
    return json.dumps(data, ensure_ascii=False, indent=2)


@tool
def get_dosing_method_stats(project_id: str) -> str:
    """获取当前工作空间中各投料类别（如填料、树脂、助剂等）下物料数量的统计。"""
    data = _get("/api/analysis/dosing-method-stats", project_id)
    return json.dumps(data, ensure_ascii=False, indent=2)


# ── formula query tools ───────────────────────────────────────────────────────

@tool
def search_formulas(query: str, project_id: str, formula_type: str = "") -> str:
    """搜索/列出当前工作空间的配方。可按名称匹配（query）或按类型筛选（formula_type: formula/material）。"""
    params: dict[str, str] = {}
    if query:
        params["name"] = query
    if formula_type:
        params["formula_type"] = formula_type
    headers = {"X-Project-Id": project_id}

    url = f"{BACKEND_URL}/api/formulas/list"
    try:
        resp = httpx.get(url, params=params, headers=headers, timeout=REQUEST_TIMEOUT)
        resp.raise_for_status()
        data = resp.json()
        compact = [
            {
                "id": f["id"],
                "name": f["name"],
                "code": f.get("code", ""),
                "component_mode": f.get("component_mode", ""),
                "formula_type": f.get("formula_type", "formula"),
                "status": f.get("status", ""),
            }
            for f in data.get("formulas", [])
        ]
        return json.dumps(compact, ensure_ascii=False, indent=2)
    except httpx.HTTPError as exc:
        return json.dumps([{"error": f"backend request failed: {exc}"}], ensure_ascii=False)


@tool
def get_formula_detail(formula_id: str, project_id: str) -> str:
    """获取当前工作空间中单个配方的完整详情，包括部件、步骤、物料和参数。"""
    headers = {"X-Project-Id": project_id}
    try:
        resp = httpx.get(
            f"{BACKEND_URL}/api/formulas/{formula_id}",
            headers=headers,
            timeout=REQUEST_TIMEOUT,
        )
        resp.raise_for_status()
        return json.dumps(resp.json(), ensure_ascii=False, indent=2)
    except httpx.HTTPError as exc:
        return json.dumps({"error": f"backend request failed: {exc}"}, ensure_ascii=False)


# ── tool collection ───────────────────────────────────────────────────────────

ALL_TOOLS = [
    get_ingredient_distribution,
    get_component_mode_ratio,
    get_step_count_distribution,
    get_dosing_method_stats,
    search_formulas,
    get_formula_detail,
]
