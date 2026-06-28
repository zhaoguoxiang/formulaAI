FROM postgres:16-alpine

# ── Custom PostgreSQL configuration ──
COPY postgresql.conf /etc/postgresql/postgresql.conf

# Apply custom config on startup
CMD ["postgres", "-c", "config_file=/etc/postgresql/postgresql.conf"]
