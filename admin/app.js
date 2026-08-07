(() => {
  const STORAGE_BASE = "streampass.admin.apiBase";
  const STORAGE_KEY = "streampass.admin.apiKey";

  const el = {
    loginView: document.getElementById("view-login"),
    appView: document.getElementById("view-app"),
    loginBase: document.getElementById("login-base"),
    loginKey: document.getElementById("login-key"),
    loginError: document.getElementById("login-error"),
    loginSubmit: document.getElementById("login-submit"),
    logout: document.getElementById("logout"),
    sessionBase: document.getElementById("session-base"),
    healthCheck: document.getElementById("health-check"),
    healthOut: document.getElementById("health-out"),
    tabs: document.getElementById("tabs"),
    usersBody: document.getElementById("users-body"),
    usersError: document.getElementById("users-error"),
    usersRefresh: document.getElementById("users-refresh"),
    relaysBody: document.getElementById("relays-body"),
    relaysError: document.getElementById("relays-error"),
    relaysRefresh: document.getElementById("relays-refresh"),
    relayForm: document.getElementById("relay-form"),
    rulesJson: document.getElementById("rules-json"),
    rulesMeta: document.getElementById("rules-meta"),
    rulesError: document.getElementById("rules-error"),
    rulesRefresh: document.getElementById("rules-refresh"),
    rulesPublish: document.getElementById("rules-publish"),
    rulesOut: document.getElementById("rules-out"),
    configForm: document.getElementById("config-form"),
    configMeta: document.getElementById("config-meta"),
    configError: document.getElementById("config-error"),
    configRefresh: document.getElementById("config-refresh"),
    configOut: document.getElementById("config-out"),
    diagRefresh: document.getElementById("diag-refresh"),
    diagError: document.getElementById("diag-error"),
    diagLimit: document.getElementById("diag-limit"),
    diagFails: document.getElementById("diag-fails"),
    diagProblems: document.getElementById("diag-problems"),
    diagBody: document.getElementById("diag-body"),
  };

  function defaultApiBase() {
    const origin = window.location.origin;
    if (!origin || origin === "null") return "https://212-43-156-33.nip.io/api/v1";
    return `${origin}/api/v1`;
  }

  function getSession() {
    return {
      base: (sessionStorage.getItem(STORAGE_BASE) || "").replace(/\/+$/, ""),
      key: sessionStorage.getItem(STORAGE_KEY) || "",
    };
  }

  function setSession(base, key) {
    sessionStorage.setItem(STORAGE_BASE, base.replace(/\/+$/, ""));
    sessionStorage.setItem(STORAGE_KEY, key);
  }

  function clearSession() {
    sessionStorage.removeItem(STORAGE_BASE);
    sessionStorage.removeItem(STORAGE_KEY);
  }

  function esc(value) {
    return String(value ?? "")
      .replaceAll("&", "&amp;")
      .replaceAll("<", "&lt;")
      .replaceAll(">", "&gt;")
      .replaceAll('"', "&quot;");
  }

  async function api(path, options = {}) {
    const { base, key } = getSession();
    if (!base || !key) throw new Error("Нет сессии");
    const url = `${base}${path.startsWith("/") ? path : `/${path}`}`;
    const headers = {
      Accept: "application/json",
      "X-Admin-Key": key,
      ...(options.body ? { "Content-Type": "application/json" } : {}),
      ...(options.headers || {}),
    };
    const res = await fetch(url, { ...options, headers });
    const text = await res.text();
    let data = null;
    if (text) {
      try {
        data = JSON.parse(text);
      } catch {
        data = text;
      }
    }
    if (!res.ok) {
      const msg =
        (data && data.error && (data.error.message || data.error.code)) ||
        (typeof data === "string" ? data : res.statusText) ||
        `HTTP ${res.status}`;
      throw new Error(msg);
    }
    return data;
  }

  /** Public GET (no admin key) for /rules and /config. */
  async function apiPublic(path) {
    const { base } = getSession();
    if (!base) throw new Error("Нет сессии");
    const url = `${base}${path.startsWith("/") ? path : `/${path}`}`;
    const res = await fetch(url, { headers: { Accept: "application/json" } });
    const data = await res.json();
    if (!res.ok) {
      throw new Error((data && data.error && data.error.message) || res.statusText);
    }
    return data;
  }

  function sanitizeKey(raw) {
    return String(raw || "")
      .replace(/[\u200B-\u200D\uFEFF]/g, "")
      .trim();
  }

  function showLogin(error, keepKey) {
    el.appView.hidden = true;
    el.loginView.hidden = false;
    if (el.loginBase) {
      el.loginBase.value = getSession().base || defaultApiBase();
    }
    if (!keepKey) {
      el.loginKey.value = "";
    }
    el.loginError.hidden = !error;
    el.loginError.textContent = error || "";
  }

  function showApp() {
    const { base } = getSession();
    el.loginView.hidden = true;
    el.appView.hidden = false;
    el.sessionBase.textContent = base;
    el.healthOut.textContent = "Нажмите «Проверить».";
  }

  function activateTab(name) {
    document.querySelectorAll(".tab").forEach((btn) => {
      btn.classList.toggle("active", btn.dataset.tab === name);
    });
    document.querySelectorAll(".tab-panel").forEach((panel) => {
      panel.hidden = panel.id !== `tab-${name}`;
    });
    if (name === "users") loadUsers();
    if (name === "relays") loadRelays();
    if (name === "rules") loadRules();
    if (name === "config") loadConfig();
    if (name === "diagnostics") loadDiagnostics();
  }

  async function loadDiagnostics() {
    if (!el.diagBody) return;
    el.diagError.hidden = true;
    el.diagBody.innerHTML = `<tr><td colspan="9">Загрузка…</td></tr>`;
    el.diagProblems.textContent = "…";
    try {
      const limit = Number(el.diagLimit?.value || 100);
      let events = await api(`/admin/diag?limit=${encodeURIComponent(limit)}`);
      if (!Array.isArray(events)) events = [];
      if (el.diagFails?.checked) {
        events = events.filter((e) => e.result && e.result !== "ok" && e.result !== "xfer");
      }
      const problemMap = new Map();
      for (const e of events) {
        if (!e.result || e.result === "ok" || e.result === "xfer") continue;
        const key = `${e.site || e.host || e.dest_ip || "?"} | ${e.reason || e.result}`;
        problemMap.set(key, (problemMap.get(key) || 0) + 1);
      }
      const top = [...problemMap.entries()]
        .sort((a, b) => b[1] - a[1])
        .slice(0, 15)
        .map(([k, n]) => `${n}×  ${k}`);
      el.diagProblems.textContent = top.length ? top.join("\n") : "Проблем в выборке нет";
      el.diagBody.innerHTML = events
        .map((e) => {
          const site = esc(e.site || e.host || "");
          const ip = esc(`${e.dest_ip || ""}:${e.dest_port || ""}`);
          const rule = esc([e.rule, e.decision_reason].filter(Boolean).join(" / "));
          return `<tr>
            <td class="mono">${esc(e.recorded_at || "")}</td>
            <td>${site}</td>
            <td class="mono">${ip}</td>
            <td>${esc(e.mode || "")}</td>
            <td>${esc(e.result || "")}${e.slow ? " 🐢" : ""}</td>
            <td class="mono">${esc(e.latency_ms ?? "")}</td>
            <td class="mono">${esc(e.speed_kbps ?? "")}</td>
            <td class="mono">${rule}</td>
            <td>${esc(e.reason || e.error_code || "")}</td>
          </tr>`;
        })
        .join("");
      if (!events.length) {
        el.diagBody.innerHTML = `<tr><td colspan="9">Нет событий</td></tr>`;
      }
    } catch (err) {
      el.diagError.hidden = false;
      el.diagError.textContent = String(err.message || err);
      el.diagBody.innerHTML = "";
      el.diagProblems.textContent = "";
    }
  }

  async function runHealthCheck() {
    el.healthOut.textContent = "…";
    try {
      const origin = getSession().base.replace(/\/api\/v1$/, "");
      const publicHealth = await fetch(`${origin}/health`).then(async (r) => ({
        status: r.status,
        body: await r.text(),
      }));
      const servers = await api("/servers/all");
      el.healthOut.textContent = JSON.stringify(
        {
          ok: true,
          public_health: publicHealth,
          servers_count: Array.isArray(servers) ? servers.length : 0,
          servers,
        },
        null,
        2
      );
    } catch (err) {
      el.healthOut.textContent = JSON.stringify(
        { ok: false, error: String(err.message || err) },
        null,
        2
      );
    }
  }

  async function loadUsers() {
    el.usersError.hidden = true;
    el.usersBody.innerHTML = `<tr><td colspan="5">Загрузка…</td></tr>`;
    try {
      const users = await api("/users");
      if (!Array.isArray(users) || users.length === 0) {
        el.usersBody.innerHTML = `<tr><td colspan="5">Нет пользователей</td></tr>`;
        return;
      }
      el.usersBody.innerHTML = users
        .map((u) => {
          const active = u.subscription_active
            ? `<span class="badge ok">active</span>`
            : `<span class="badge bad">inactive</span>`;
          return `<tr>
            <td>${esc(u.email)}</td>
            <td>${active}</td>
            <td class="mono">${esc(u.subscription_active_until || "—")}</td>
            <td class="mono">${esc(u.created_at || "—")}</td>
            <td class="mono">${esc(u.id)}</td>
          </tr>`;
        })
        .join("");
    } catch (err) {
      el.usersBody.innerHTML = "";
      el.usersError.hidden = false;
      el.usersError.textContent = String(err.message || err);
    }
  }

  async function loadRelays() {
    el.relaysError.hidden = true;
    el.relaysBody.innerHTML = `<tr><td colspan="6">Загрузка…</td></tr>`;
    try {
      const servers = await api("/servers/all");
      if (!Array.isArray(servers) || servers.length === 0) {
        el.relaysBody.innerHTML = `<tr><td colspan="6">Нет relay</td></tr>`;
        return;
      }
      el.relaysBody.innerHTML = servers
        .map((s) => {
          const health = s.healthy
            ? `<span class="badge ok">healthy</span>`
            : `<span class="badge bad">unhealthy</span>`;
          return `<tr data-id="${esc(s.id)}">
            <td class="mono">${esc(s.id)}</td>
            <td>${esc(s.region_name || s.region)}</td>
            <td class="mono">${esc(s.host)}:${esc(s.port)}</td>
            <td>${health}</td>
            <td class="mono">${esc(s.rtt_ms ?? "—")}</td>
            <td><button type="button" class="danger" data-delete-relay="${esc(s.id)}">Delete</button></td>
          </tr>`;
        })
        .join("");
    } catch (err) {
      el.relaysBody.innerHTML = "";
      el.relaysError.hidden = false;
      el.relaysError.textContent = String(err.message || err);
    }
  }

  async function loadRules() {
    el.rulesError.hidden = true;
    el.rulesOut.textContent = "";
    try {
      const set = await apiPublic("/rules");
      el.rulesMeta.textContent = `version=${set.version} created_at=${set.created_at || "—"}`;
      el.rulesJson.value = JSON.stringify(set.rules || [], null, 2);
    } catch (err) {
      el.rulesError.hidden = false;
      el.rulesError.textContent = String(err.message || err);
    }
  }

  async function publishRules() {
    el.rulesError.hidden = true;
    el.rulesOut.textContent = "…";
    try {
      const rules = JSON.parse(el.rulesJson.value);
      if (!Array.isArray(rules)) throw new Error("Ожидается JSON-массив rules");
      const published = await api("/rules", {
        method: "POST",
        body: JSON.stringify({ rules }),
      });
      el.rulesMeta.textContent = `version=${published.version} created_at=${published.created_at || "—"}`;
      el.rulesJson.value = JSON.stringify(published.rules || [], null, 2);
      el.rulesOut.textContent = JSON.stringify(published, null, 2);
    } catch (err) {
      el.rulesError.hidden = false;
      el.rulesError.textContent = String(err.message || err);
      el.rulesOut.textContent = "";
    }
  }

  async function loadConfig() {
    el.configError.hidden = true;
    el.configOut.textContent = "";
    try {
      const cfg = await apiPublic("/config");
      el.configMeta.textContent = `version=${cfg.version} updated_at=${cfg.updated_at || "—"}`;
      el.configForm.min_supported_client_version.value = cfg.min_supported_client_version || "";
      el.configForm.latest_client_version.value = cfg.latest_client_version || "";
      el.configForm.client_download_url.value = cfg.client_download_url || "";
      el.configForm.telemetry_enabled.checked = !!cfg.telemetry_enabled;
      el.configForm.rule_poll_interval_sec.value = cfg.rule_poll_interval_sec ?? 300;
      el.configForm.relay_poll_interval_sec.value = cfg.relay_poll_interval_sec ?? 60;
    } catch (err) {
      el.configError.hidden = false;
      el.configError.textContent = String(err.message || err);
    }
  }

  async function publishConfig(ev) {
    ev.preventDefault();
    el.configError.hidden = true;
    el.configOut.textContent = "…";
    const fd = new FormData(el.configForm);
    const body = {
      min_supported_client_version: String(fd.get("min_supported_client_version") || "").trim(),
      latest_client_version: String(fd.get("latest_client_version") || "").trim(),
      client_download_url: String(fd.get("client_download_url") || "").trim(),
      telemetry_enabled: el.configForm.telemetry_enabled.checked,
      rule_poll_interval_sec: Number(fd.get("rule_poll_interval_sec") || 0),
      relay_poll_interval_sec: Number(fd.get("relay_poll_interval_sec") || 0),
    };
    try {
      const published = await api("/config", {
        method: "POST",
        body: JSON.stringify(body),
      });
      el.configMeta.textContent = `version=${published.version} updated_at=${published.updated_at || "—"}`;
      el.configOut.textContent = JSON.stringify(published, null, 2);
    } catch (err) {
      el.configError.hidden = false;
      el.configError.textContent = String(err.message || err);
      el.configOut.textContent = "";
    }
  }

  async function tryLogin() {
    const base = (el.loginBase.value || defaultApiBase()).trim().replace(/\/+$/, "");
    const key = sanitizeKey(el.loginKey.value);
    if (!base || !key) {
      showLogin("Вставьте Admin Key", true);
      return;
    }
    el.loginSubmit.disabled = true;
    el.loginError.hidden = true;
    setSession(base, key);
    try {
      await api("/servers/all");
      showApp();
      activateTab("health");
      // Health is best-effort: failure must not kick the operator back to login.
      try {
        await runHealthCheck();
      } catch (healthErr) {
        el.healthOut.textContent = JSON.stringify(
          { ok: false, error: String(healthErr.message || healthErr) },
          null,
          2
        );
      }
    } catch (err) {
      clearSession();
      const raw = String(err.message || err);
      const hint =
        /invalid or missing admin key|FORBIDDEN/i.test(raw)
          ? "Неверный Admin Key. Нужен ADMIN_API_KEY из /root/StreamPass/.env (не пароль пользователя)."
          : raw;
      showLogin(hint, true);
    } finally {
      el.loginSubmit.disabled = false;
    }
  }

  el.loginSubmit.addEventListener("click", () => {
    tryLogin();
  });
  el.loginKey.addEventListener("keydown", (ev) => {
    if (ev.key === "Enter") {
      ev.preventDefault();
      tryLogin();
    }
  });

  el.logout.addEventListener("click", () => {
    clearSession();
    showLogin();
  });

  el.healthCheck.addEventListener("click", () => runHealthCheck());
  el.usersRefresh.addEventListener("click", () => loadUsers());
  el.relaysRefresh.addEventListener("click", () => loadRelays());
  el.rulesRefresh.addEventListener("click", () => loadRules());
  el.rulesPublish.addEventListener("click", () => publishRules());
  el.configRefresh.addEventListener("click", () => loadConfig());
  el.configForm.addEventListener("submit", publishConfig);
  el.diagRefresh?.addEventListener("click", () => loadDiagnostics());
  el.diagFails?.addEventListener("change", () => loadDiagnostics());

  el.relaysBody.addEventListener("click", async (ev) => {
    const btn = ev.target.closest("[data-delete-relay]");
    if (!btn) return;
    const id = btn.getAttribute("data-delete-relay");
    if (!id || !confirm(`Удалить relay ${id}?`)) return;
    try {
      await api(`/servers/${encodeURIComponent(id)}`, { method: "DELETE" });
      await loadRelays();
    } catch (err) {
      el.relaysError.hidden = false;
      el.relaysError.textContent = String(err.message || err);
    }
  });

  el.relayForm.addEventListener("submit", async (ev) => {
    ev.preventDefault();
    el.relaysError.hidden = true;
    const fd = new FormData(el.relayForm);
    const body = {
      id: String(fd.get("id") || "").trim(),
      region: String(fd.get("region") || "").trim(),
      host: String(fd.get("host") || "").trim(),
      port: Number(fd.get("port") || 0),
      connection_config: String(fd.get("connection_config") || "").trim(),
    };
    try {
      await api("/servers", { method: "POST", body: JSON.stringify(body) });
      el.relayForm.reset();
      el.relayForm.port.value = "443";
      await loadRelays();
    } catch (err) {
      el.relaysError.hidden = false;
      el.relaysError.textContent = String(err.message || err);
    }
  });

  el.tabs.addEventListener("click", (ev) => {
    const btn = ev.target.closest(".tab");
    if (!btn || btn.disabled) return;
    activateTab(btn.dataset.tab);
  });

  const session = getSession();
  try {
    if (el.loginBase && !el.loginBase.value) {
      el.loginBase.value = session.base || defaultApiBase();
    }
    if (session.base && session.key) {
      showApp();
      activateTab("health");
    } else {
      showLogin();
    }
  } catch (bootErr) {
    console.error(bootErr);
    showLogin(String(bootErr.message || bootErr), true);
  }

  window.StreamPassAdmin = {
    api,
    apiPublic,
    getSession,
    setSession,
    clearSession,
    activateTab,
    loadUsers,
    loadRelays,
    loadRules,
    loadConfig,
    publishRules,
  };
})();
