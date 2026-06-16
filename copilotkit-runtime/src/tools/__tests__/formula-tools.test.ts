import {
  listFormulas,
  getFormula,
  createFormula,
  updateFormula,
  deleteFormula,
  getAnalysis,
  formulaTools,
} from "../formula-tools";

const GO_API_URL = "http://localhost:8080";

// ── Helpers ──────────────────────────────────────────────────────────────────

function mockFetch(
  responseData: unknown,
  status = 200,
  headers: Record<string, string> = { "content-type": "application/json" }
) {
  return jest.fn().mockResolvedValue({
    ok: status >= 200 && status < 300,
    status,
    headers: {
      get: (name: string) => headers[name.toLowerCase()] || null,
    },
    json: jest.fn().mockResolvedValue(responseData),
    text: jest.fn().mockResolvedValue(JSON.stringify(responseData)),
  });
}

function lastFetchCall(fetchMock: jest.Mock): { url: string; init?: RequestInit } {
  const calls = fetchMock.mock.calls;
  const last = calls[calls.length - 1];
  return { url: last?.[0] as string, init: last?.[1] as RequestInit | undefined };
}

beforeEach(() => {
  jest.restoreAllMocks();
});

// ── Tool Collection ─────────────────────────────────────────────────────────

describe("formulaTools", () => {
  it("should export exactly 6 tools", () => {
    expect(formulaTools).toHaveLength(6);
  });

  it("each tool should have a name, description, and handler", () => {
    for (const tool of formulaTools) {
      expect(tool).toHaveProperty("name");
      expect(typeof tool.name).toBe("string");
      expect(tool.name.length).toBeGreaterThan(0);

      expect(tool).toHaveProperty("description");
      expect(typeof tool.description).toBe("string");
      expect(tool.description!.length).toBeGreaterThan(0);

      expect(tool).toHaveProperty("handler");
      expect(typeof tool.handler).toBe("function");
    }
  });

  it("tool names should match expected set", () => {
    const names = formulaTools.map((t) => t.name).sort();
    expect(names).toEqual([
      "create_formula",
      "delete_formula",
      "get_analysis",
      "get_formula",
      "list_formulas",
      "update_formula",
    ]);
  });

  it("debug_hint is not included", () => {
    // 验证没有额外的debug_hint等工具
    const names = formulaTools.map((t) => t.name);
    expect(names).not.toContain("debug_hint");
  });
});

// ── list_formulas ───────────────────────────────────────────────────────────

describe("list_formulas", () => {
  it("should call GET /api/formulas and return list", async () => {
    const mockFormulas = [
      { id: "11111111-1111-1111-1111-111111111111", name: "Formula 1" },
      { id: "22222222-2222-2222-2222-222222222222", name: "Formula 2" },
    ];
    const fetchSpy = mockFetch(mockFormulas);
    jest.spyOn(global, "fetch" as any).mockImplementation(fetchSpy);

    const result = await listFormulas.handler!();

    expect(fetchSpy).toHaveBeenCalledTimes(1);
    const { url, init } = lastFetchCall(fetchSpy);
    expect(url).toBe(`${GO_API_URL}/api/formulas`);
    expect(init?.method).toBeUndefined(); // GET is default
    expect(result).toEqual(mockFormulas);
  });

  it("should throw when Go API returns a 500 error", async () => {
    const fetchSpy = mockFetch({ error: "database down" }, 500);
    jest.spyOn(global, "fetch" as any).mockImplementation(fetchSpy);

    await expect(listFormulas.handler!()).rejects.toThrow("database down");
  });
});

// ── get_formula ─────────────────────────────────────────────────────────────

describe("get_formula", () => {
  it("should call GET /api/formulas/:id and return formula", async () => {
    const mockFormula = {
      id: "11111111-1111-1111-1111-111111111111",
      name: "Test Formula",
      code: "TF-001",
    };
    const fetchSpy = mockFetch(mockFormula);
    jest.spyOn(global, "fetch" as any).mockImplementation(fetchSpy);

    const result = await getFormula.handler!({
      id: "11111111-1111-1111-1111-111111111111",
    });

    expect(fetchSpy).toHaveBeenCalledTimes(1);
    const { url, init } = lastFetchCall(fetchSpy);
    expect(url).toBe(
      `${GO_API_URL}/api/formulas/11111111-1111-1111-1111-111111111111`
    );
    expect(init?.method).toBeUndefined();
    expect(result).toEqual(mockFormula);
  });

  it("should throw when id is missing", async () => {
    await expect(getFormula.handler!({} as any)).rejects.toThrow(
      "缺少必填参数 id"
    );
  });

  it("should throw when formula is not found (404)", async () => {
    const fetchSpy = mockFetch({ error: "formula not found" }, 404);
    jest.spyOn(global, "fetch" as any).mockImplementation(fetchSpy);

    await expect(
      getFormula.handler!({ id: "nonexistent-id" })
    ).rejects.toThrow("formula not found");
  });
});

// ── create_formula ──────────────────────────────────────────────────────────

describe("create_formula", () => {
  it("should POST the correct body to /api/formulas", async () => {
    const mockCreated = {
      id: "33333333-3333-3333-3333-333333333333",
      name: "New Formula",
      code: "NF-001",
      component_mode: "double",
      status: "draft",
    };
    const fetchSpy = mockFetch(mockCreated, 201);
    jest.spyOn(global, "fetch" as any).mockImplementation(fetchSpy);

    const result = await createFormula.handler!({
      name: "New Formula",
      code: "NF-001",
      component_mode: "double",
    });

    expect(fetchSpy).toHaveBeenCalledTimes(1);
    const { url, init } = lastFetchCall(fetchSpy);
    expect(url).toBe(`${GO_API_URL}/api/formulas`);
    expect(init?.method).toBe("POST");
    expect(init?.headers).toEqual({ "Content-Type": "application/json" });

    const body = JSON.parse(init!.body as string);
    expect(body.name).toBe("New Formula");
    expect(body.code).toBe("NF-001");
    expect(body.component_mode).toBe("double");
    expect(result).toEqual(mockCreated);
  });

  it("should include parts and steps when provided", async () => {
    const fetchSpy = mockFetch({ id: "4444" }, 201);
    jest.spyOn(global, "fetch" as any).mockImplementation(fetchSpy);

    const handler = createFormula.handler as Function;
    await handler({
      name: "With Parts",
      code: "WP-001",
      component_mode: "double",
      parts: [
        {
          name: "PartA",
          ingredients: [{ material: "Resin", percentage: 60 }],
        },
      ],
      steps: [{ step_no: 1, name: "Mix" }],
    });

    const { init } = lastFetchCall(fetchSpy);
    const body = JSON.parse(init!.body as string);
    expect(body.parts).toHaveLength(1);
    expect(body.steps).toHaveLength(1);
  });

  it("should throw when required fields (name, code, component_mode) are missing", async () => {
    await expect(
      createFormula.handler!({ name: "Only Name" })
    ).rejects.toThrow("缺少必填参数");
  });

  it("should throw on validation error (422)", async () => {
    const fetchSpy = mockFetch(
      { error: "validation failed", details: "invalid data" },
      422
    );
    jest.spyOn(global, "fetch" as any).mockImplementation(fetchSpy);

    await expect(
      createFormula.handler!({
        name: "Bad",
        code: "B-001",
        component_mode: "single",
      })
    ).rejects.toThrow("validation failed");
  });
});

// ── update_formula ──────────────────────────────────────────────────────────

describe("update_formula", () => {
  it("should PUT to /api/formulas/:id with only provided fields", async () => {
    const mockUpdated = {
      id: "11111111-1111-1111-1111-111111111111",
      name: "Updated Name",
      code: "OLD-001",
    };
    const fetchSpy = mockFetch(mockUpdated);
    jest.spyOn(global, "fetch" as any).mockImplementation(fetchSpy);

    const result = await updateFormula.handler!({
      id: "11111111-1111-1111-1111-111111111111",
      name: "Updated Name",
    });

    expect(fetchSpy).toHaveBeenCalledTimes(1);
    const { url, init } = lastFetchCall(fetchSpy);
    expect(url).toBe(
      `${GO_API_URL}/api/formulas/11111111-1111-1111-1111-111111111111`
    );
    expect(init?.method).toBe("PUT");

    const body = JSON.parse(init!.body as string);
    expect(body.name).toBe("Updated Name");
    expect(body.code).toBeUndefined(); // not provided
  });

  it("should throw when id is missing", async () => {
    await expect(updateFormula.handler!({ name: "No ID" })).rejects.toThrow(
      "缺少必填参数 id"
    );
  });

  it("should throw when no updatable fields provided", async () => {
    await expect(
      updateFormula.handler!({
        id: "11111111-1111-1111-1111-111111111111",
      })
    ).rejects.toThrow("至少需要提供一个要修改的字段");
  });

  it("should throw when formula not found (404)", async () => {
    const fetchSpy = mockFetch({ error: "formula not found" }, 404);
    jest.spyOn(global, "fetch" as any).mockImplementation(fetchSpy);

    await expect(
      updateFormula.handler!({
        id: "nonexistent-id",
        name: "Ghost",
      })
    ).rejects.toThrow("formula not found");
  });
});

// ── delete_formula ──────────────────────────────────────────────────────────

describe("delete_formula", () => {
  it("should DELETE /api/formulas/:id and return success", async () => {
    const fetchSpy = mockFetch(null, 204);
    jest.spyOn(global, "fetch" as any).mockImplementation(fetchSpy);

    const result = await deleteFormula.handler!({
      id: "11111111-1111-1111-1111-111111111111",
    });

    expect(fetchSpy).toHaveBeenCalledTimes(1);
    const { url, init } = lastFetchCall(fetchSpy);
    expect(url).toBe(
      `${GO_API_URL}/api/formulas/11111111-1111-1111-1111-111111111111`
    );
    expect(init?.method).toBe("DELETE");
    expect(result).toEqual({ success: true });
  });

  it("should throw when id is missing", async () => {
    await expect(deleteFormula.handler!({} as any)).rejects.toThrow(
      "缺少必填参数 id"
    );
  });

  it("should throw when formula not found (404)", async () => {
    const fetchSpy = mockFetch({ error: "formula not found" }, 404);
    jest.spyOn(global, "fetch" as any).mockImplementation(fetchSpy);

    await expect(
      deleteFormula.handler!({ id: "nonexistent-id" })
    ).rejects.toThrow("formula not found");
  });
});

// ── get_analysis ────────────────────────────────────────────────────────────

describe("get_analysis", () => {
  it("should call GET /api/analysis/ingredient-distribution", async () => {
    const mockData = [
      { material: "Resin", count: 10, avg_percentage: 55.5 },
      { material: "Hardener", count: 8, avg_percentage: 42.0 },
    ];
    const fetchSpy = mockFetch(mockData);
    jest.spyOn(global, "fetch" as any).mockImplementation(fetchSpy);

    const result = await getAnalysis.handler!({
      endpoint: "ingredient-distribution",
    });

    expect(fetchSpy).toHaveBeenCalledTimes(1);
    const { url, init } = lastFetchCall(fetchSpy);
    expect(url).toBe(`${GO_API_URL}/api/analysis/ingredient-distribution`);
    expect(init?.method).toBeUndefined();
    expect(result).toEqual(mockData);
  });

  it("should throw when endpoint is missing", async () => {
    await expect(getAnalysis.handler!({} as any)).rejects.toThrow(
      "缺少必填参数 endpoint"
    );
  });

  it("should throw on server error (500)", async () => {
    const fetchSpy = mockFetch({ error: "query failed" }, 500);
    jest.spyOn(global, "fetch" as any).mockImplementation(fetchSpy);

    await expect(
      getAnalysis.handler!({ endpoint: "component-mode-ratio" })
    ).rejects.toThrow("query failed");
  });
});

// ── Network error handling ──────────────────────────────────────────────────

describe("Network errors", () => {
  it("should throw a user-friendly error on connection refused", async () => {
    const fetchSpy = jest.fn().mockRejectedValue(new Error("connect ECONNREFUSED"));
    jest.spyOn(global, "fetch" as any).mockImplementation(fetchSpy);

    await expect(listFormulas.handler!()).rejects.toThrow(
      "无法连接到配方服务"
    );
  });

  it("should handle non-JSON responses gracefully", async () => {
    const fetchSpy = jest.fn().mockResolvedValue({
      ok: false,
      status: 502,
      headers: {
        get: () => "text/html",
      },
      json: jest.fn().mockRejectedValue(new Error("Not JSON")),
      text: jest.fn().mockResolvedValue("<html>Bad Gateway</html>"),
    });
    jest.spyOn(global, "fetch" as any).mockImplementation(fetchSpy);

    await expect(listFormulas.handler!()).rejects.toThrow("API 返回错误状态码 502");
  });
});
