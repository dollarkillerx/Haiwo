const state = {
  projects: [],
  pipelines: [],
  runs: [],
  agents: [],
  backups: [],
};

const translations = {
  en: {
    "brand.subtitle": "Delivery control plane",
    "nav.overview": "Overview",
    "nav.overviewHint": "Run overview",
    "nav.projects": "Projects",
    "nav.projectsHint": "Repositories",
    "nav.pipelines": "Pipelines",
    "nav.pipelinesHint": "Orchestration",
    "nav.runs": "Runs",
    "nav.runsHint": "History",
    "nav.agents": "Agents",
    "nav.agentsHint": "Workers",
    "nav.backups": "Backups",
    "nav.backupsHint": "Database",
    "sidebar.rpc": "JSON-RPC over WebSocket",
    "top.eyebrow": "Haiwo MVP Console",
    "hero.eyebrow": "Release Operations",
    "hero.title": "Ship from git events, comments, schedules, or manual runs.",
    "hero.body": "Agent labels route work to build, deploy, database, staging, or production machines while the server keeps the orchestration record.",
    "metric.projects": "Projects",
    "metric.projectsHint": "registered repositories",
    "metric.pipelines": "Pipelines",
    "metric.pipelinesHint": "delivery workflows",
    "metric.runs": "Runs",
    "metric.agents": "Online Agents",
    "metric.noRuns": "no executions yet",
    "metric.allQuiet": "all quiet",
    "metric.active": "{count} active",
    "metric.review": "{count} need review",
    "metric.capacity": "capacity available",
    "metric.noCapacity": "no capacity online",
    "overview.recentRuns": "Recent Runs",
    "overview.recentRunsHint": "Latest pipeline executions and rollback attempts.",
    "overview.agentCapacity": "Agent Capacity",
    "overview.agentCapacityHint": "Live workers available for scheduled jobs.",
    "project.create": "Create Project",
    "project.createHint": "Bind a repository to Haiwo.",
    "project.listHint": "Source repositories registered in the control plane.",
    "pipeline.create": "Create Pipeline",
    "pipeline.createHint": "Paste a stage and job definition.",
    "pipeline.listHint": "Reusable workflows for builds, deploys, backups, and rollbacks.",
    "run.start": "Start Run",
    "run.startHint": "Launch a pipeline with a ref or comment command.",
    "run.listHint": "Execution records are polled every five seconds.",
    "agent.listHint": "Workers connected through JSON-RPC over WebSocket.",
    "backup.listHint": "Database backup artifacts created by agent jobs.",
    "field.name": "Name",
    "field.provider": "Provider",
    "field.repoURL": "Repository URL",
    "field.defaultBranch": "Default Branch",
    "field.project": "Project",
    "field.pipelineJSON": "Pipeline JSON",
    "field.pipeline": "Pipeline",
    "field.ref": "Ref",
    "field.comment": "Comment",
    "view.overview": "Overview",
    "view.projects": "Projects",
    "view.pipelines": "Pipelines",
    "view.runs": "Runs",
    "view.agents": "Agents",
    "view.backups": "Backups",
    "hero.noAgents": "No agents online",
    "hero.noAgentsHint": "Start an agent to accept pipeline tasks.",
    "hero.failed": "Review failed runs",
    "hero.failedHint": "{count} run{plural} failed or rollback failed.",
    "hero.running": "Delivery in progress",
    "hero.runningHint": "{count} run{plural} queued or running.",
    "hero.ready": "Ready to deploy",
    "hero.readyHint": "{count} agent{plural} online and waiting.",
    "sync.never": "Not synced",
    "sync.now": "Synced {time}",
    "sidebar.onlineAgents": "{count} online agent{plural}",
    "notice.projectCreated": "Project created.",
    "notice.projectRequired": "Create a project before adding a pipeline.",
    "notice.pipelineInvalid": "Pipeline JSON is invalid: {message}",
    "notice.pipelineCreated": "Pipeline created.",
    "notice.pipelineRequired": "Create a pipeline before starting a run.",
    "notice.runStarted": "Run {id} started.",
    "empty.projects": "Create a project to connect a repository.",
    "empty.pipelines": "Create a pipeline to define build, deploy, backup, or rollback work.",
    "empty.runs": "Runs appear here after a manual, webhook, or scheduled trigger starts a pipeline.",
    "empty.agents": "Start an agent to make execution capacity available.",
    "empty.backups": "Backup records appear after db_backup jobs complete.",
    "empty.capacity": "No agents have registered yet. Start one with",
    "th.name": "Name",
    "th.provider": "Provider",
    "th.repository": "Repository",
    "th.defaultBranch": "Default Branch",
    "th.id": "ID",
    "th.project": "Project",
    "th.stages": "Stages",
    "th.jobs": "Jobs",
    "th.run": "Run",
    "th.pipeline": "Pipeline",
    "th.status": "Status",
    "th.source": "Source",
    "th.ref": "Ref",
    "th.updated": "Updated",
    "th.labels": "Labels",
    "th.load": "Load",
    "th.version": "Version",
    "th.lastSeen": "Last Seen",
    "th.database": "Database",
    "th.type": "Type",
    "th.environment": "Environment",
    "th.path": "Path",
    "th.checksum": "Checksum",
    "th.created": "Created",
  },
  zh: {
    "brand.subtitle": "交付控制平面",
    "nav.overview": "概览",
    "nav.overviewHint": "运行概览",
    "nav.projects": "项目",
    "nav.projectsHint": "仓库项目",
    "nav.pipelines": "流水线",
    "nav.pipelinesHint": "任务编排",
    "nav.runs": "运行",
    "nav.runsHint": "执行历史",
    "nav.agents": "Agent",
    "nav.agentsHint": "执行节点",
    "nav.backups": "备份",
    "nav.backupsHint": "数据库",
    "sidebar.rpc": "基于 WebSocket 的 JSON-RPC",
    "top.eyebrow": "Haiwo MVP 控制台",
    "hero.eyebrow": "发布运维",
    "hero.title": "通过 Git 事件、评论、定时任务或手动操作发布。",
    "hero.body": "Agent 标签会把任务路由到构建、部署、数据库、预发或生产机器，Server 负责保存完整编排记录。",
    "metric.projects": "项目",
    "metric.projectsHint": "已注册仓库",
    "metric.pipelines": "流水线",
    "metric.pipelinesHint": "交付工作流",
    "metric.runs": "运行",
    "metric.agents": "在线 Agent",
    "metric.noRuns": "暂无执行记录",
    "metric.allQuiet": "当前稳定",
    "metric.active": "{count} 个运行中",
    "metric.review": "{count} 个需检查",
    "metric.capacity": "有可用容量",
    "metric.noCapacity": "暂无在线容量",
    "overview.recentRuns": "最近运行",
    "overview.recentRunsHint": "最新流水线执行和回滚记录。",
    "overview.agentCapacity": "Agent 容量",
    "overview.agentCapacityHint": "可执行定时任务的在线工作节点。",
    "project.create": "创建项目",
    "project.createHint": "把代码仓库绑定到 Haiwo。",
    "project.listHint": "已注册到控制平面的源代码仓库。",
    "pipeline.create": "创建流水线",
    "pipeline.createHint": "粘贴 stage 和 job 定义。",
    "pipeline.listHint": "用于构建、部署、备份和回滚的可复用工作流。",
    "run.start": "启动运行",
    "run.startHint": "使用 ref 或评论命令启动流水线。",
    "run.listHint": "执行记录每五秒自动刷新。",
    "agent.listHint": "通过 WebSocket JSON-RPC 连接的工作节点。",
    "backup.listHint": "由 agent 数据库任务创建的备份产物。",
    "field.name": "名称",
    "field.provider": "代码平台",
    "field.repoURL": "仓库地址",
    "field.defaultBranch": "默认分支",
    "field.project": "项目",
    "field.pipelineJSON": "流水线 JSON",
    "field.pipeline": "流水线",
    "field.ref": "Ref",
    "field.comment": "评论命令",
    "view.overview": "概览",
    "view.projects": "项目",
    "view.pipelines": "流水线",
    "view.runs": "运行",
    "view.agents": "Agent",
    "view.backups": "备份",
    "hero.noAgents": "暂无在线 Agent",
    "hero.noAgentsHint": "启动一个 agent 后即可接收流水线任务。",
    "hero.failed": "请检查失败运行",
    "hero.failedHint": "{count} 个运行失败或回滚失败。",
    "hero.running": "交付执行中",
    "hero.runningHint": "{count} 个运行正在排队或执行。",
    "hero.ready": "可以发布",
    "hero.readyHint": "{count} 个 agent 在线等待任务。",
    "sync.never": "尚未同步",
    "sync.now": "已同步 {time}",
    "sidebar.onlineAgents": "{count} 个在线 agent",
    "notice.projectCreated": "项目已创建。",
    "notice.projectRequired": "请先创建项目，再添加流水线。",
    "notice.pipelineInvalid": "流水线 JSON 无效：{message}",
    "notice.pipelineCreated": "流水线已创建。",
    "notice.pipelineRequired": "请先创建流水线，再启动运行。",
    "notice.runStarted": "运行 {id} 已启动。",
    "empty.projects": "创建项目后即可连接代码仓库。",
    "empty.pipelines": "创建流水线后即可定义构建、部署、备份或回滚任务。",
    "empty.runs": "通过手动、webhook 或定时触发启动流水线后，运行记录会出现在这里。",
    "empty.agents": "启动 agent 后即可提供任务执行容量。",
    "empty.backups": "db_backup 任务完成后，备份记录会出现在这里。",
    "empty.capacity": "尚无 agent 注册。启动命令：",
    "th.name": "名称",
    "th.provider": "平台",
    "th.repository": "仓库",
    "th.defaultBranch": "默认分支",
    "th.id": "ID",
    "th.project": "项目",
    "th.stages": "阶段",
    "th.jobs": "任务",
    "th.run": "运行",
    "th.pipeline": "流水线",
    "th.status": "状态",
    "th.source": "来源",
    "th.ref": "Ref",
    "th.updated": "更新时间",
    "th.labels": "标签",
    "th.load": "负载",
    "th.version": "版本",
    "th.lastSeen": "最后在线",
    "th.database": "数据库",
    "th.type": "类型",
    "th.environment": "环境",
    "th.path": "路径",
    "th.checksum": "校验值",
    "th.created": "创建时间",
  },
};

let currentLang = initialLanguage();
let lastSyncTime = "";

const samplePipeline = {
  name: "staging deploy",
  stages: [
    {
      name: "checkout",
      jobs: [
        {
          id: "checkout",
          name: "Checkout code",
          job_type: "git_checkout",
          agent_mode: "single",
          agent_labels: ["build"],
          repo: {
            url: "https://github.com/example/app.git",
            ref: "main",
          },
          required: true,
        },
      ],
    },
    {
      name: "deploy",
      jobs: [
        {
          id: "deploy",
          name: "Deploy service",
          job_type: "command",
          agent_mode: "chain",
          agent_labels: ["deploy", "staging"],
          commands: ["echo deploy staging"],
          timeout_seconds: 300,
          required: true,
        },
      ],
    },
  ],
};

document.querySelector('[name="definition"]').value = JSON.stringify(samplePipeline, null, 2);
document.getElementById("last-refresh").textContent = t("sync.never");

document.querySelectorAll(".nav-item").forEach((button) => {
  button.addEventListener("click", () => {
    document.querySelectorAll(".nav-item").forEach((item) => item.classList.remove("active"));
    button.classList.add("active");
    showView(button.dataset.view);
  });
});

document.querySelectorAll("[data-lang]").forEach((button) => {
  button.addEventListener("click", () => {
    currentLang = button.dataset.lang;
    localStorage.setItem("haiwo_lang", currentLang);
    applyI18n();
    render();
  });
});

document.getElementById("refresh").addEventListener("click", refresh);
document.getElementById("project-form").addEventListener("submit", createProject);
document.getElementById("pipeline-form").addEventListener("submit", createPipeline);
document.getElementById("run-form").addEventListener("submit", startRun);

refresh();
setInterval(refresh, 5000);

function showView(name) {
  document.querySelectorAll(".view").forEach((view) => view.classList.add("hidden"));
  document.getElementById(`${name}-view`).classList.remove("hidden");
  document.getElementById("view-title").textContent = t(`view.${name}`);
}

async function refresh() {
  try {
    const [projects, pipelines, runs, agents, backups] = await Promise.all([
      api("/api/projects"),
      api("/api/pipelines"),
      api("/api/runs"),
      api("/api/agents"),
      api("/api/backups"),
    ]);
    Object.assign(state, { projects, pipelines, runs, agents, backups });
    lastSyncTime = new Date().toLocaleTimeString();
    render();
    document.getElementById("last-refresh").textContent = t("sync.now", { time: lastSyncTime });
  } catch (error) {
    notify(error.message, true);
  }
}

async function createProject(event) {
  event.preventDefault();
  const form = new FormData(event.currentTarget);
  const payload = Object.fromEntries(form.entries());
  await api("/api/projects", { method: "POST", body: payload });
  event.currentTarget.reset();
  event.currentTarget.default_branch.value = "main";
  notify(t("notice.projectCreated"));
  await refresh();
}

async function createPipeline(event) {
  event.preventDefault();
  const form = new FormData(event.currentTarget);
  const projectID = form.get("project_id");
  if (!projectID) {
    notify(t("notice.projectRequired"), true);
    return;
  }
  let definition;
  try {
    definition = JSON.parse(form.get("definition"));
  } catch (error) {
    notify(t("notice.pipelineInvalid", { message: error.message }), true);
    return;
  }
  await api(`/api/projects/${projectID}/pipelines`, { method: "POST", body: definition });
  notify(t("notice.pipelineCreated"));
  await refresh();
}

async function startRun(event) {
  event.preventDefault();
  const form = new FormData(event.currentTarget);
  const pipelineID = form.get("pipeline_id");
  if (!pipelineID) {
    notify(t("notice.pipelineRequired"), true);
    return;
  }
  const payload = { type: "manual", ref: form.get("ref"), comment: form.get("comment") };
  const run = await api(`/api/pipelines/${pipelineID}/runs`, { method: "POST", body: payload });
  notify(t("notice.runStarted", { id: run.id }));
  await refresh();
}

async function api(path, options = {}) {
  const init = { ...options, headers: { "content-type": "application/json", ...(options.headers || {}) } };
  if (init.body && typeof init.body !== "string") {
    init.body = JSON.stringify(init.body);
  }
  const response = await fetch(path, init);
  const text = await response.text();
  if (!response.ok) {
    throw new Error(text || response.statusText);
  }
  return text ? JSON.parse(text) : null;
}

function render() {
  applyI18n();
  const onlineAgents = state.agents.filter((agent) => agent.status === "online");
  const failedRuns = state.runs.filter((run) => run.status === "failed" || run.status === "rollback_failed").length;
  const runningRuns = state.runs.filter((run) => run.status === "running" || run.status === "rollback_running" || run.status === "queued").length;

  document.getElementById("metric-projects").textContent = state.projects.length;
  document.getElementById("metric-pipelines").textContent = state.pipelines.length;
  document.getElementById("metric-runs").textContent = state.runs.length;
  document.getElementById("metric-agents").textContent = onlineAgents.length;
  document.getElementById("metric-run-note").textContent = runningRuns
    ? t("metric.active", { count: runningRuns })
    : failedRuns
      ? t("metric.review", { count: failedRuns })
      : t("metric.allQuiet");
  document.getElementById("metric-agent-note").textContent = onlineAgents.length ? t("metric.capacity") : t("metric.noCapacity");
  document.getElementById("sidebar-agent-count").textContent = t("sidebar.onlineAgents", {
    count: onlineAgents.length,
    plural: onlineAgents.length === 1 ? "" : "s",
  });

  renderHero(onlineAgents.length, runningRuns, failedRuns);
  renderSelect("pipeline-project", state.projects, "id", (project) => `${project.name} (${project.provider})`);
  renderSelect("run-pipeline", state.pipelines, "id", (pipeline) => `${pipeline.name} (${pipeline.id})`);

  renderProjects();
  renderPipelines();
  renderRuns();
  renderAgents();
  renderBackups();
  renderCapacity();
  document.getElementById("recent-runs").innerHTML = runTable(state.runs.slice(-8).reverse());
}

function renderHero(onlineAgents, runningRuns, failedRuns) {
  const title = document.getElementById("hero-status");
  const subtitle = document.getElementById("hero-subtitle");
  if (!onlineAgents) {
    title.textContent = t("hero.noAgents");
    subtitle.textContent = t("hero.noAgentsHint");
    return;
  }
  if (failedRuns) {
    title.textContent = t("hero.failed");
    subtitle.textContent = t("hero.failedHint", { count: failedRuns, plural: failedRuns === 1 ? "" : "s" });
    return;
  }
  if (runningRuns) {
    title.textContent = t("hero.running");
    subtitle.textContent = t("hero.runningHint", { count: runningRuns, plural: runningRuns === 1 ? "" : "s" });
    return;
  }
  title.textContent = t("hero.ready");
  subtitle.textContent = t("hero.readyHint", { count: onlineAgents, plural: onlineAgents === 1 ? "" : "s" });
}

function renderProjects() {
  document.getElementById("projects-table").innerHTML = table(
    [t("th.name"), t("th.provider"), t("th.repository"), t("th.defaultBranch"), t("th.id")],
    state.projects.map((p) => [strong(p.name), badge(p.provider), repo(p.repo_url), code(p.default_branch), code(p.id)]),
    t("empty.projects")
  );
}

function renderPipelines() {
  document.getElementById("pipelines-table").innerHTML = table(
    [t("th.name"), t("th.project"), t("th.stages"), t("th.jobs"), t("th.id")],
    state.pipelines.map((p) => [
      strong(p.name),
      shortName(state.projects.find((x) => x.id === p.project_id)),
      p.stages?.length || 0,
      (p.stages || []).reduce((sum, stage) => sum + (stage.jobs?.length || 0), 0),
      code(p.id),
    ]),
    t("empty.pipelines")
  );
}

function renderRuns() {
  document.getElementById("runs-table").innerHTML = runTable([...state.runs].reverse());
}

function renderAgents() {
  document.getElementById("agents-table").innerHTML = table(
    [t("th.name"), t("th.status"), t("th.labels"), t("th.load"), t("th.version"), t("th.lastSeen")],
    state.agents.map((a) => [
      strong(a.name),
      status(a.status),
      labels(a.labels || []),
      `${a.current_run}/${a.max_running}`,
      escapeHTML(a.version),
      date(a.last_seen_at),
    ]),
    t("empty.agents")
  );
}

function renderBackups() {
  document.getElementById("backups-table").innerHTML = table(
    [t("th.id"), t("th.database"), t("th.type"), t("th.environment"), t("th.path"), t("th.checksum"), t("th.created")],
    state.backups.map((b) => [
      code(b.id),
      escapeHTML(b.database || "-"),
      badge(b.type || "-"),
      escapeHTML(b.environment || "-"),
      path(b.path || "-"),
      code(b.checksum || "-"),
      date(b.created_at),
    ]),
    t("empty.backups")
  );
}

function renderCapacity() {
  const target = document.getElementById("agent-capacity");
  if (!state.agents.length) {
    target.innerHTML = `<div class="empty">${escapeHTML(t("empty.capacity"))} <span class="code">go run ./cmd/agent</span>.</div>`;
    return;
  }
  target.innerHTML = state.agents
    .map((agent) => {
      const max = Math.max(agent.max_running || 1, 1);
      const current = agent.current_run || 0;
      const pct = Math.min(100, Math.round((current / max) * 100));
      return `<div class="capacity-item">
        <div>
          <strong>${escapeHTML(agent.name)}</strong>
          <span>${labels(agent.labels || [])}</span>
        </div>
        <div class="load">
          <div class="load-track"><div class="load-fill" style="width:${pct}%"></div></div>
          <span>${current}/${max}</span>
        </div>
      </div>`;
    })
    .join("");
}

function runTable(runs) {
  return table(
    [t("th.run"), t("th.project"), t("th.pipeline"), t("th.status"), t("th.source"), t("th.ref"), t("th.updated")],
    runs.map((r) => [
      code(r.id),
      shortName(state.projects.find((p) => p.id === r.project_id)),
      shortName(state.pipelines.find((p) => p.id === r.pipeline_id)),
      status(r.status),
      escapeHTML(r.source),
      code(r.ref || "-"),
      date(r.updated_at),
    ]),
    t("empty.runs")
  );
}

function table(headers, rows, emptyText) {
  if (!rows.length) {
    return `<div class="empty">${escapeHTML(emptyText || "No records yet.")}</div>`;
  }
  return `<table><thead><tr>${headers.map((h) => `<th>${escapeHTML(h)}</th>`).join("")}</tr></thead><tbody>${rows
    .map((row) => `<tr>${row.map((cell) => `<td>${cell}</td>`).join("")}</tr>`)
    .join("")}</tbody></table>`;
}

function renderSelect(id, items, valueKey, labelFn) {
  const select = document.getElementById(id);
  const selected = select.value;
  select.innerHTML = items.map((item) => `<option value="${escapeHTML(item[valueKey])}">${escapeHTML(labelFn(item))}</option>`).join("");
  if (items.some((item) => item[valueKey] === selected)) {
    select.value = selected;
  }
}

function notify(message, error = false) {
  const notice = document.getElementById("notice");
  notice.textContent = message;
  notice.style.borderLeftColor = error ? "var(--red)" : "var(--accent)";
  notice.classList.remove("hidden");
  setTimeout(() => notice.classList.add("hidden"), 3200);
}

function applyI18n() {
  document.documentElement.lang = currentLang === "zh" ? "zh-CN" : "en";
  document.querySelectorAll("[data-lang]").forEach((button) => {
    button.classList.toggle("active", button.dataset.lang === currentLang);
  });
  document.querySelectorAll("[data-i18n]").forEach((node) => {
    node.textContent = t(node.dataset.i18n);
  });
  document.getElementById("last-refresh").textContent = lastSyncTime ? t("sync.now", { time: lastSyncTime }) : t("sync.never");
  const active = document.querySelector(".nav-item.active");
  if (active) {
    document.getElementById("view-title").textContent = t(`view.${active.dataset.view}`);
  }
}

function shortName(item) {
  return item ? strong(item.name) : "-";
}

function strong(value) {
  return `<strong>${escapeHTML(value || "-")}</strong>`;
}

function status(value) {
  const label = translations[currentLang][`status.${value}`] || value || "-";
  return `<span class="status ${escapeHTML(value || "")}">${escapeHTML(label)}</span>`;
}

function badge(value) {
  return `<span class="status">${escapeHTML(value || "-")}</span>`;
}

function labels(values) {
  if (!values.length) return "-";
  return values.map((value) => `<span class="status">${escapeHTML(value)}</span>`).join(" ");
}

function code(value) {
  return `<span class="code">${escapeHTML(value || "-")}</span>`;
}

function repo(value) {
  return `<span title="${escapeHTML(value || "-")}">${escapeHTML(value || "-")}</span>`;
}

function path(value) {
  return `<span class="code" title="${escapeHTML(value)}">${escapeHTML(value)}</span>`;
}

function date(value) {
  if (!value) return "-";
  return escapeHTML(new Date(value).toLocaleString());
}

function escapeHTML(value) {
  return String(value)
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#039;");
}

function initialLanguage() {
  const saved = localStorage.getItem("haiwo_lang");
  if (saved === "zh" || saved === "en") {
    return saved;
  }
  const browserLang = (navigator.language || navigator.userLanguage || "").toLowerCase();
  return browserLang.startsWith("zh") ? "zh" : "en";
}

function t(key, vars = {}) {
  let value = translations[currentLang][key] || translations.en[key] || key;
  for (const [name, replacement] of Object.entries(vars)) {
    value = value.replaceAll(`{${name}}`, replacement);
  }
  return value;
}
