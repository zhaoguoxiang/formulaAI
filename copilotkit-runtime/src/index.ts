import express from "express";
import cors from "cors";
import helmet from "helmet";
import rateLimit from "express-rate-limit";
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

const DEFAULT_MODEL = process.env.OPENAI_MODEL || "gpt-4o";
const DEFAULT_TEMPERATURE = 0.7;
const DEFAULT_MAX_TOKENS = 4096;

function createApp(): express.Express {
  const app = express();

  // ── CORS ───────────────────────────────────────────────────────
  const corsOrigin = process.env.CORS_ORIGIN || "http://localhost:4200";
  app.use(
    cors({
      origin: corsOrigin,
      methods: ["GET", "POST", "OPTIONS"],
      allowedHeaders: ["Content-Type", "Accept"],
    })
  );

  // ── Security headers ───────────────────────────────────────────
  app.use(helmet());

  // ── Rate limiting (100 req/min per IP on CopilotKit endpoint) ──
  app.use(
    "/api/copilotkit",
    rateLimit({
      windowMs: 60_000,
      max: 100,
      standardHeaders: true,
      legacyHeaders: false,
      message: { error: "rate_limited", message: "请求频率过高，请稍后重试" },
    })
  );

  // ── Body parsing ───────────────────────────────────────────────
  app.use(express.json({ limit: "1mb" }));

  // ── Health check ───────────────────────────────────────────────
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
  if (!openaiApiKey) {
    console.error(
      "WARNING: OPENAI_API_KEY is not set. The AI assistant will not function. " +
        "Set OPENAI_API_KEY in your environment or .env file."
    );
  }

  const goApiUrl = process.env.GO_API_URL || "http://localhost:8080";
  console.log(`GO_API_URL: ${goApiUrl}`);
  console.log(`Model: ${DEFAULT_MODEL}`);

  // ── OpenAI client and service adapter ──────────────────────────
  const openai = openaiApiKey
    ? new OpenAI({ apiKey: openaiApiKey })
    : undefined;

  const serviceAdapter = openaiApiKey
    ? new OpenAIAdapter({
        openai: openai as unknown as OpenAI,
        model: DEFAULT_MODEL,
      })
    : undefined;

  // ── System prompt ──────────────────────────────────────────────
  const systemPrompt = getSystemPrompt();

  // ── CopilotKit Runtime with formula tools ──────────────────────
  const runtime = new CopilotRuntime({
    remoteEndpoints: [{ url: goApiUrl }],
    actions: formulaTools,
  });

  // ── CopilotKit endpoint ────────────────────────────────────────
  const copilotHandler = copilotRuntimeNodeExpressEndpoint({
    endpoint: "/",
    runtime,
    serviceAdapter: serviceAdapter as any,
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

  // ── Global error handler (no internals leaked) ─────────────────
  app.use(
    (
      err: unknown,
      _req: express.Request,
      res: express.Response,
      _next: express.NextFunction,
    ) => {
      const message = err instanceof Error ? err.message : String(err);
      console.error("Unhandled error:", message);

      if (message.includes("context_length_exceeded") || message.includes("maximum context length")) {
        res.status(400).json({
          error: "token_overflow",
          message: "对话内容过长，超出了模型上下文限制。请简化您的问题或减少对话历史后重试。",
        });
        return;
      }

      if (message.includes("rate_limit") || message.includes("429") || message.includes("too many requests")) {
        res.status(429).json({
          error: "rate_limited",
          message: "请求频率过高，已被限流。请等待片刻后重试。",
        });
        return;
      }

      res.status(500).json({
        error: "internal_server_error",
        message: "服务器内部错误，请稍后重试。",
      });
    },
  );

  return app;
}

// ── Start server ────────────────────────────────────────────────
const port = parseInt(process.env.PORT || "3001", 10);

if (process.env.NODE_ENV !== "test") {
  const app = createApp();
  app.listen(port, () => {
    console.log(`CopilotKit runtime listening at http://localhost:${port}`);
    console.log(`  Health check:  http://localhost:${port}/api/health`);
    console.log(`  CopilotKit:    POST http://localhost:${port}/api/copilotkit`);
    console.log(`  System prompt: GET  http://localhost:${port}/api/system-prompt`);
    console.log(`  Tools:         GET  http://localhost:${port}/api/tools`);
  });
}

export { createApp };
