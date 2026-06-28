"""
LangChain tools that call the Go backend analysis & formula APIs.

Each tool fetches real data from the backend and returns JSON for the LLM to reason over.
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

def _get(path: str) -> list[dict[str, Any]]:
    """GET a JSON array from the Go backend. Raises on non-2xx."""
    url = f"{BACKEND_URL}{path}"
    try:
        resp = httpx.get(url, timeout=REQUEST_TIMEOUT)
        resp.raise_for_status()
        return resp.json()
    except httpx.HTTPError as exc:
        return [{"error": f"backend request failed: {exc}"}]


# ── analysis tools ────────────────────────────────────────────────────────────

@tool
def get_ingredient_distribution() -> str:
    """获取所有配方中每种物料的使用频率和平均用量百分比。
    返回物料名、出现次数、平均百分比的列表。"""
    data = _get("/api/analysis/ingredient-distribution")
    return json.dumps(data, ensure_ascii=False, indent=2)


@tool
def get_component_mode_ratio() -> str:
    """获取单组分与双组分配方的数量及占比。
    返回 mode（single/double）、数量和百分比。"""
    data = _get("/api/analysis/component-mode-ratio")
    return json.dumps(data, ensure_ascii=False, indent=2)


@tool
def get_step_count_distribution() -> str:
    """获取配方中工艺步骤数量的分布（分桶: 1-2, 3-5, 5+）。
    返回每个区间的配方数量。"""
    data = _get("/api/analysis/step-count-distribution")
    return json.dumps(data, ensure_ascii=False, indent=2)


@tool
def get_dosing_method_stats() -> str:
    """获取各投料类别（如填料、树脂、助剂等）下物料数量的统计。
    返回类别名和对应物料数量。"""
    data = _get("/api/analysis/dosing-method-stats")
    return json.dumps(data, ensure_ascii=False, indent=2)


# ── formula query tools ───────────────────────────────────────────────────────

@tool
def search_formulas(query: str = "", formula_type: str = "") -> str:
    """搜索/列出配方。可按名称模糊匹配（query）或按类型筛选（formula_type: formula/material）。
    返回配方的 id、name、code、component_mode、status、formula_type 等。"""
    params = {}
    if query:
        params["name"] = query
    if formula_type:
        params["formula_type"] = formula_type

    url = f"{BACKEND_URL}/api/formulas/list"
    try:
        resp = httpx.get(url, params=params, timeout=REQUEST_TIMEOUT)
        resp.raise_for_status()
        data = resp.json()
        # return a compact view for the LLM
        compact = [
            {
                "id": f["id"],
                "name": f["name"],
                "code": f.get("code", ""),
                "component_mode": f.get("component_mode", ""),
                "formula_type": f.get("formula_type", "formula"),
                "status": f.get("status", ""),
            }
            for f in data
        ]
        return json.dumps(compact, ensure_ascii=False, indent=2)
    except httpx.HTTPError as exc:
        return json.dumps([{"error": f"backend request failed: {exc}"}], ensure_ascii=False)


@tool
def get_formula_detail(formula_id: str) -> str:
    """获取单个配方的完整详情，包括部件、步骤、物料和参数。
    参数 formula_id 是配方的 UUID。"""
    try:
        resp = httpx.get(
            f"{BACKEND_URL}/api/formulas/{formula_id}", timeout=REQUEST_TIMEOUT
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
