import type { Action, Parameter } from "@copilotkit/shared";

// ── Helpers ──────────────────────────────────────────────────────────────────

function getGoApiUrl(): string {
  return process.env.GO_API_URL || "http://localhost:8080";
}

async function callGoApi(
  path: string,
  options?: RequestInit
): Promise<unknown> {
  const url = `${getGoApiUrl()}${path}`;

  let res: Response;
  try {
    res = await fetch(url, {
      ...options,
      signal: AbortSignal.timeout(15_000),
      headers: {
        "Content-Type": "application/json",
        ...options?.headers,
      },
    });
  } catch (err) {
    const msg = err instanceof Error ? err.message : String(err);
    throw new Error(`无法连接到配方服务 (${getGoApiUrl()}): ${msg}`);
  }

  // 204 No Content — success with no body (e.g. delete)
  if (res.status === 204) {
    return { success: true };
  }

  // Try to parse JSON body
  const contentType = res.headers.get("content-type") || "";
  let body: unknown;
  let isJson = false;
  if (contentType.includes("application/json")) {
    isJson = true;
    try {
      body = await res.json();
    } catch {
      body = { error: "failed to parse API response" };
    }
  } else {
    const text = await res.text();
    body = { details: text };
  }

  if (!res.ok) {
    const errBody = body as Record<string, unknown>;
    const message = isJson
      ? ((errBody?.error as string) || (errBody?.details as string) || `API 返回错误状态码 ${res.status}`)
      : `API 返回错误状态码 ${res.status}`;
    throw new Error(message);
  }

  return body;
}

// ── Shared parameter definitions ─────────────────────────────────────────────

const ingredientParams: Parameter[] = [
  {
    name: "material",
    type: "string",
    description: "配料/原材料名称，例如 Resin、Hardener、Pigment",
    required: true,
  },
  {
    name: "percentage",
    type: "number",
    description: "配料百分比 (0-100)，组分内所有配料百分比总和应为100",
    required: true,
  },
  {
    name: "sort_order",
    type: "number",
    description: "排序序号（可选）",
    required: false,
  },
];

const partParams: Parameter[] = [
  {
    name: "name",
    type: "string",
    description: "组分名称",
    required: true,
    enum: ["PartA", "PartB", "PartMain"],
  },
  {
    name: "mix_ratio",
    type: "number",
    description: "混合比例（双组分模式下A:B的比例，可选）",
    required: false,
  },
  {
    name: "sort_order",
    type: "number",
    description: "排序序号（可选）",
    required: false,
  },
  {
    name: "ingredients",
    type: "object[]",
    description: "该组分包含的配料列表",
    required: false,
    attributes: ingredientParams,
  },
];

const stepParams: Parameter[] = [
  {
    name: "step_no",
    type: "number",
    description: "步骤序号",
    required: true,
  },
  {
    name: "name",
    type: "string",
    description: "步骤名称，例如 Mix、Heat、Cool",
    required: true,
  },
  {
    name: "temperature",
    type: "string",
    description: "温度条件，例如 25°C 或 100°C",
    required: false,
  },
  {
    name: "duration",
    type: "string",
    description: "持续时长，例如 10min 或 2h",
    required: false,
  },
];

// ── Tool Definitions ─────────────────────────────────────────────────────────

/**
 * list_formulas — 查询所有配方列表
 * GET /api/formulas
 */
export const listFormulas: Action = {
  name: "list_formulas",
  description:
    "查询所有配方列表。当你需要浏览现有配方、查找特定配方、或了解系统中有哪些配方时使用此工具。返回配方的完整信息，包括组分、配料、步骤和投料动作。",
  parameters: [],
  handler: async () => {
    return callGoApi("/api/formulas");
  },
};

/**
 * get_formula — 查询单个配方详情
 * GET /api/formulas/:id
 */
export const getFormula: Action<Parameter[]> = {
  name: "get_formula",
  description:
    "查询单个配方的详细信息。当用户提到特定配方的名称或ID，需要查看该配方的完整内容（组分、配料、步骤等）时使用此工具。",
  parameters: [
    {
      name: "id",
      type: "string",
      description: "配方的唯一标识符 (UUID 格式，例如 550e8400-e29b-41d4-a716-446655440000)",
      required: true,
    },
  ],
  handler: async (args) => {
    if (!args || typeof args !== "object" || !("id" in args)) {
      throw new Error("缺少必填参数 id");
    }
    const id = String((args as Record<string, unknown>).id);
    return callGoApi(`/api/formulas/${encodeURIComponent(id)}`);
  },
};

/**
 * create_formula — 创建新配方
 * POST /api/formulas
 */
export const createFormula: Action<Parameter[]> = {
  name: "create_formula",
  description:
    "创建一个新的配方。当用户要求新建配方、录入配方数据时使用此工具。需要提供配方名称、编码、组分模式，以及可选的组分列表和工艺步骤。",
  parameters: [
    {
      name: "name",
      type: "string",
      description: "配方名称，例如 高硬度硅胶配方",
      required: true,
    },
    {
      name: "code",
      type: "string",
      description: "配方编码，例如 SG-001",
      required: true,
    },
    {
      name: "component_mode",
      type: "string",
      description:
        "组分模式：single（单组分，只有一个PartMain）或 double（双组分，有PartA和PartB）",
      required: true,
      enum: ["single", "double"],
    },
    {
      name: "status",
      type: "string",
      description: "配方状态：draft（草稿）、active（启用）、archived（归档），默认 draft",
      required: false,
      enum: ["draft", "active", "archived"],
    },
    {
      name: "parts",
      type: "object[]",
      description: "组分列表。单组分模式只填一个 PartMain；双组分模式填 PartA 和 PartB",
      required: false,
      attributes: partParams,
    },
    {
      name: "steps",
      type: "object[]",
      description: "工艺步骤列表，按 step_no 排序",
      required: false,
      attributes: stepParams,
    },
  ],
  handler: async (args) => {
    if (!args || typeof args !== "object") {
      throw new Error("缺少必填参数");
    }
    const a = args as Record<string, unknown>;
    if (!a.name || !a.code || !a.component_mode) {
      throw new Error("缺少必填参数：name, code, component_mode 均为必填项");
    }

    const body: Record<string, unknown> = {
      name: a.name,
      code: a.code,
      component_mode: a.component_mode,
    };
    if (a.status) body.status = a.status;

    const parts = a.parts;
    if (Array.isArray(parts) && parts.length > 0) {
      body.parts = parts;
    }

    const steps = a.steps;
    if (Array.isArray(steps) && steps.length > 0) {
      body.steps = steps;
    }

    return callGoApi("/api/formulas", {
      method: "POST",
      body: JSON.stringify(body),
    });
  },
};

/**
 * update_formula — 修改配方
 * PUT /api/formulas/:id
 */
export const updateFormula: Action<Parameter[]> = {
  name: "update_formula",
  description:
    "修改已有配方的信息。当用户要求修改配方名称、编码、组分模式、组分内容或工艺步骤时使用此工具。只需提供要修改的字段，未提供的字段保持不变。注意：修改组分或配料时，Go API 会执行全量替换（需提供完整的 parts/ingredients 列表）。",
  parameters: [
    {
      name: "id",
      type: "string",
      description: "要修改的配方 UUID",
      required: true,
    },
    {
      name: "name",
      type: "string",
      description: "新的配方名称（可选）",
      required: false,
    },
    {
      name: "code",
      type: "string",
      description: "新的配方编码（可选）",
      required: false,
    },
    {
      name: "component_mode",
      type: "string",
      description: "新的组分模式（可选）",
      required: false,
      enum: ["single", "double"],
    },
    {
      name: "status",
      type: "string",
      description: "新的配方状态（可选）",
      required: false,
      enum: ["draft", "active", "archived"],
    },
    {
      name: "parts",
      type: "object[]",
      description: "新的组分列表（可选，提供时会全量替换现有组分）",
      required: false,
      attributes: partParams,
    },
    {
      name: "steps",
      type: "object[]",
      description: "新的工艺步骤列表（可选，提供时会全量替换现有步骤）",
      required: false,
      attributes: stepParams,
    },
  ],
  handler: async (args) => {
    if (!args || typeof args !== "object" || !("id" in args)) {
      throw new Error("缺少必填参数 id");
    }
    const a = args as Record<string, unknown>;
    const id = String(a.id);

    // Build body from provided fields only
    const body: Record<string, unknown> = {};
    const updatableFields = [
      "name",
      "code",
      "component_mode",
      "status",
      "parts",
      "steps",
    ] as const;
    for (const field of updatableFields) {
      if (field in a && a[field] !== undefined && a[field] !== null) {
        body[field] = a[field];
      }
    }

    if (Object.keys(body).length === 0) {
      throw new Error("至少需要提供一个要修改的字段");
    }

    return callGoApi(`/api/formulas/${encodeURIComponent(id)}`, {
      method: "PUT",
      body: JSON.stringify(body),
    });
  },
};

/**
 * delete_formula — 删除配方
 * DELETE /api/formulas/:id
 */
export const deleteFormula: Action<Parameter[]> = {
  name: "delete_formula",
  description:
    "删除指定配方及其所有关联数据（组分、配料、步骤、投料动作）。此操作不可逆，执行前务必向用户确认。",
  parameters: [
    {
      name: "id",
      type: "string",
      description: "要删除的配方 UUID",
      required: true,
    },
  ],
  handler: async (args) => {
    if (!args || typeof args !== "object" || !("id" in args)) {
      throw new Error("缺少必填参数 id");
    }
    const id = String((args as Record<string, unknown>).id);
    return callGoApi(`/api/formulas/${encodeURIComponent(id)}`, {
      method: "DELETE",
    });
  },
};

/**
 * get_analysis — 获取统计数据
 * GET /api/analysis/:endpoint
 */
export const getAnalysis: Action<Parameter[]> = {
  name: "get_analysis",
  description:
    "获取配方数据的统计分析结果。支持四种统计维度：配料分布统计、组分模式比例、步骤数量分布、投料方式统计。当用户询问配方数据的统计信息、趋势或占比时使用此工具。",
  parameters: [
    {
      name: "endpoint",
      type: "string",
      description:
        "统计端点类型：ingredient-distribution（配料分布）、component-mode-ratio（组分模式比例）、step-count-distribution（步骤数量分布）、dosing-method-stats（投料方式统计）",
      required: true,
      enum: [
        "ingredient-distribution",
        "component-mode-ratio",
        "step-count-distribution",
        "dosing-method-stats",
      ],
    },
  ],
  handler: async (args) => {
    if (!args || typeof args !== "object" || !("endpoint" in args)) {
      throw new Error("缺少必填参数 endpoint");
    }
    const endpoint = String((args as Record<string, unknown>).endpoint);
    return callGoApi(
      `/api/analysis/${encodeURIComponent(endpoint)}`
    );
  },
};

// ── Exported tool collection ─────────────────────────────────────────────────

/**
 * All formula CRUD tool definitions for registration with CopilotKit runtime.
 */
export const formulaTools: Action[] = [
  listFormulas,
  getFormula,
  createFormula,
  updateFormula,
  deleteFormula,
  getAnalysis,
] as Action[];
