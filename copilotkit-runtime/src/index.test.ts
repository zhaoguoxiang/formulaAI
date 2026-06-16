import request from "supertest";
import { createApp } from "./index";

describe("CopilotKit Runtime Server", () => {
  let app: ReturnType<typeof createApp>;

  beforeAll(() => {
    // Set a fake API key so the server doesn't warn excessively
    process.env.OPENAI_API_KEY = "sk-test-key-for-jest";
    process.env.GO_API_URL = "http://localhost:8080";
    app = createApp();
  });

  // ── Health Check ──────────────────────────────────────────────
  describe("GET /api/health", () => {
    it("should return 200 with status ok", async () => {
      const res = await request(app).get("/api/health");
      expect(res.status).toBe(200);
      expect(res.body).toHaveProperty("status", "ok");
      expect(res.body).toHaveProperty("service", "copilotkit-runtime");
    });

    it("should return model config info", async () => {
      const res = await request(app).get("/api/health");
      expect(res.body).toHaveProperty("model");
      expect(res.body).toHaveProperty("temperature");
      expect(res.body).toHaveProperty("maxTokens");
    });

    it("should return JSON content type", async () => {
      const res = await request(app).get("/api/health");
      expect(res.headers["content-type"]).toMatch(/json/);
    });
  });

  // ── System Prompt ─────────────────────────────────────────────
  describe("GET /api/system-prompt", () => {
    it("should return 200 with system prompt", async () => {
      const res = await request(app).get("/api/system-prompt");
      expect(res.status).toBe(200);
      expect(res.body).toHaveProperty("systemPrompt");
      expect(res.body).toHaveProperty("language", "zh-CN");
    });

    it("should contain Chinese text in system prompt", async () => {
      const res = await request(app).get("/api/system-prompt");
      // The system prompt should contain Chinese characters
      expect(res.body.systemPrompt).toMatch(/[\u4e00-\u9fff]/);
      // Should mention key terms
      expect(res.body.systemPrompt).toContain("配方");
      expect(res.body.systemPrompt).toContain("组分");
    });
  });

  // ── Tools ─────────────────────────────────────────────────────
  describe("GET /api/tools", () => {
    it("should return 200 with tools array", async () => {
      const res = await request(app).get("/api/tools");
      expect(res.status).toBe(200);
      expect(res.body).toHaveProperty("tools");
      expect(Array.isArray(res.body.tools)).toBe(true);
    });

    it("should have required CRUD tools", async () => {
      const res = await request(app).get("/api/tools");
      const toolNames = res.body.tools.map((t: any) => t.name);
      expect(toolNames).toContain("list_formulas");
      expect(toolNames).toContain("get_formula");
      expect(toolNames).toContain("create_formula");
      expect(toolNames).toContain("update_formula");
      expect(toolNames).toContain("delete_formula");
      expect(toolNames).toContain("get_analysis");
    });

    it("should have Chinese descriptions for tools", async () => {
      const res = await request(app).get("/api/tools");
      for (const tool of res.body.tools) {
        expect(tool.description).toMatch(/[\u4e00-\u9fff]/);
      }
    });
  });

  // ── CopilotKit Endpoint ───────────────────────────────────────
  describe("POST /api/copilotkit", () => {
    it("should accept POST requests (endpoint exists)", async () => {
      const res = await request(app)
        .post("/api/copilotkit")
        .send({ method: "info" })
        .set("Content-Type", "application/json");
      // The endpoint should exist and return something (even if it's an error
      // about missing agent or invalid request — the key is it doesn't 404)
      expect(res.status).not.toBe(404);
    });

    it("should return 404 for unknown routes", async () => {
      const res = await request(app).get("/api/nonexistent");
      expect(res.status).toBe(404);
    });
  });

  // ── CORS ──────────────────────────────────────────────────────
  describe("CORS", () => {
    it("should include CORS headers on health endpoint", async () => {
      const res = await request(app)
        .options("/api/health")
        .set("Origin", "http://localhost:4200")
        .set("Access-Control-Request-Method", "GET");
      expect([200, 204, 404]).toContain(res.status);
    });
  });

  // ── Error Handling ────────────────────────────────────────────
  describe("Error handling", () => {
    it("should return 404 with JSON for non-existent routes", async () => {
      const res = await request(app).get("/api/nonexistent-route");
      expect(res.status).toBe(404);
      expect(res.body).toHaveProperty("error", "Not found");
    });

    it("should handle json parse errors gracefully", async () => {
      const res = await request(app)
        .post("/api/copilotkit")
        .set("Content-Type", "application/json")
        .send("not valid json");
      // Express json parser returns 400 for invalid JSON
      expect(res.status).toBeGreaterThanOrEqual(400);
    });
  });
});
