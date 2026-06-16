import express from "express";
import cors from "cors";
import dotenv from "dotenv";
import {
  CopilotRuntime,
  OpenAIAdapter,
  copilotRuntimeNodeExpressEndpoint,
} from "@copilotkit/runtime";
import OpenAI from "openai";
import { getSystemPrompt } from "./prompts/system";
import { formulaTools } from "./tools/formula-tools";

dotenv.config();

// ── Constants ────────────────────────────────────────────────────
const DEFAULT_MODEL = process.env.OPENAI_MODEL || "gpt-4o";
const DEFAULT_TEMPERATURE = 0.7;
const DEFAULT_MAX_TOKENS = 4096;

export function createApp(): express.Express {
  const app = express();

  // ── Health check endpoint ──────────────────────────────────────
  app.get("/api/health", (_req, res) => {
    res.json({
      status: "ok",
      service: "copilotkit-runtime",
      model: DEFAULT_MODEL,
      temperature: DEFAULT_TEMPERATURE,
      maxTokens: DEFAULT_MAX_TOKENS,
    });
  });

  // ── Validate required environment variables ────────────────────
  const openaiApiKey = process.env.OPENAI_API_KEY;
  const goApiUrl = process.env.GO_API_URL || "http://localhost:8080";

  if (!openaiApiKey) {
    console.error(
      "ERROR: OPENAI_API_KEY is not set. " +
        "The CopilotKit endpoint will not be able to call OpenAI. " +
        "Set OPENAI_API_KEY in your environment or .env file."
    );
  }

  console.log(`GO_API_URL: ${goApiUrl}`);
  console.log(`Model: ${DEFAULT_MODEL}`);
  console.log(`Temperature: ${DEFAULT_TEMPERATURE}`);
  console.log(`Max Tokens: ${DEFAULT_MAX_TOKENS}`);

  // ── OpenAI client and service adapter ──────────────────────────
  const openai = new OpenAI({
    apiKey: openaiApiKey || "placeholder",
  });

  const serviceAdapter = new OpenAIAdapter({
    openai: openai as any,
    model: DEFAULT_MODEL,
  });

  // ── System prompt ──────────────────────────────────────────────
  const systemPrompt = getSystemPrompt();

  // ── CopilotKit Runtime with formula tools ──────────────────────
  const runtime = new CopilotRuntime({
    remoteEndpoints: [
      {
        url: goApiUrl,
      },
    ],
    actions: formulaTools as any,
  });

  // ── CORS ───────────────────────────────────────────────────────
  app.use(
    cors({
      origin: "*",
      methods: ["GET", "POST", "OPTIONS", "PUT", "DELETE", "PATCH"],
      allowedHeaders: ["*"],
    })
  );

  app.use(express.json());

  // ── CopilotKit endpoint ────────────────────────────────────────
  const copilotHandler = copilotRuntimeNodeExpressEndpoint({
    endpoint: "/",
    runtime,
    serviceAdapter,
    properties: {
      systemPrompt,
      model: DEFAULT_MODEL,
      temperature: String(DEFAULT_TEMPERATURE),
      maxTokens: String(DEFAULT_MAX_TOKENS),
    },
  });

  app.use("/api/copilotkit", copilotHandler);

  // ── System prompt info endpoint ────────────────────────────────
  app.get("/api/system-prompt", (_req, res) => {
    res.json({
      systemPrompt,
      language: "zh-CN",
      description: "Formula AI Assistant system prompt in Chinese",
    });
  });

  // ── Tool definitions endpoint ──────────────────────────────────
  app.get("/api/tools", (_req, res) => {
    res.json({
      tools: formulaTools.map((tool) => ({
        name: tool.name,
        description: tool.description,
        parameters: tool.parameters,
      })),
    });
  });

  // ── 404 handler ────────────────────────────────────────────────
  app.use((_req, res) => {
    res.status(404).json({ error: "Not found" });
  });

  // ── Global error handler ───────────────────────────────────────
  app.use(
    (
      err: Error,
      _req: express.Request,
      res: express.Response,
      _next: express.NextFunction
    ) => {
      console.error("Unhandled error:", err.message);

      // Detect token overflow / context length errors
      if (
        err.message.includes("context_length_exceeded") ||
        err.message.includes("maximum context length") ||
        (err.message.includes("token") &&
          err.message.includes("reduce"))
      ) {
        res.status(400).json({
          error: "token_overflow",
          message:
            "对话内容过长，超出了模型上下文限制。" +
            "请简化您的问题或减少对话历史后重试。",
          detail: err.message,
        });
        return;
      }

      // Detect rate limit errors
      if (
        err.message.includes("rate_limit") ||
        err.message.includes("429") ||
        err.message.includes("too many requests")
      ) {
        res.status(429).json({
          error: "rate_limited",
          message:
            "请求频率过高，已被限流。请等待片刻后重试。",
          detail: err.message,
        });
        return;
      }

      res.status(500).json({
        error: "internal_server_error",
        message: "服务器内部错误，请稍后重试。",
      });
    }
  );

  return app;
}

// ── Start server (only when run directly, not during tests) ─────
const port = parseInt(process.env.PORT || "3001", 10);

if (process.env.NODE_ENV !== "test") {
  const app = createApp();
  app.listen(port, () => {
    console.log(`CopilotKit runtime listening at http://localhost:${port}`);
    console.log(`  Health check:    http://localhost:${port}/api/health`);
    console.log(`  CopilotKit:      POST http://localhost:${port}/api/copilotkit`);
    console.log(`  System prompt:   GET  http://localhost:${port}/api/system-prompt`);
    console.log(`  Tools:           GET  http://localhost:${port}/api/tools`);
  });
}
