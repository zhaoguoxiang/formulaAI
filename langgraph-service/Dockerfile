# ── build stage: install dependencies ──
FROM python:3.12-alpine AS builder

# Use China mirror for faster builds in mainland China
ENV PIP_INDEX_URL=https://pypi.tuna.tsinghua.edu.cn/simple

WORKDIR /app

COPY pyproject.toml .

# Install dependencies into a virtualenv
RUN python -m venv /opt/venv && \
    /opt/venv/bin/pip install --no-cache-dir .

# ── run stage ──
FROM python:3.12-alpine

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

COPY --from=builder /opt/venv /opt/venv

WORKDIR /app

COPY src/ ./src/

ENV PATH="/opt/venv/bin:$PATH" \
    PYTHONUNBUFFERED=1 \
    PORT=5050

USER appuser

HEALTHCHECK --interval=15s --timeout=5s --retries=3 \
    CMD curl -sf http://localhost:${PORT}/api/health | grep -q '"status":"ok"'

EXPOSE 5050

CMD ["python", "-m", "src.main"]
