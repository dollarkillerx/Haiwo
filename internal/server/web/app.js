const state = {
  projects: [],
  pipelines: [],
  runs: [],
  agents: [],
  settings: {},
  selectedRunId: "",
  selectedRunPipelineId: "",
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
    "nav.settings": "Settings",
    "nav.settingsHint": "System",
    "sidebar.rpc": "JSON-RPC over WebSocket",
    "top.eyebrow": "Haiwo MVP Console",
    "hero.eyebrow": "Release Operations",
    "hero.title": "Ship from git events, comments, schedules, or manual runs.",
    "hero.body": "Agent labels route work to build, deploy, staging, or production machines while the server keeps the orchestration record.",
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
    "pipeline.createHint": "Choose a trigger and add ordered Agent command steps.",
    "pipeline.listHint": "Reusable workflows for push, tag, manual runs, and ordered Agent commands.",
    "run.start": "Start Run",
    "run.startHint": "Launch a pipeline with a ref or comment command.",
    "run.historyFilter": "Run History",
    "run.historyFilterHint": "Select a pipeline to inspect previous executions.",
    "run.allPipelines": "All pipelines",
    "run.listHint": "Execution records are polled every five seconds. Click Details to inspect logs.",
    "agent.listHint": "Workers connected through JSON-RPC over WebSocket.",
    "agent.create": "Add Agent",
    "agent.createHint": "Create an agent identity and deployment script.",
    "agent.deployScript": "Deployment Script",
    "agent.deployScriptHint": "Run this on the target machine after placing the agent binary there.",
    "field.name": "Name",
    "field.provider": "Provider",
    "field.repoURL": "Repository URL",
    "field.defaultBranch": "Default Branch",
    "field.project": "Project",
    "field.pipelineJSON": "Pipeline JSON",
    "field.pipelineName": "Pipeline Name",
    "field.stageName": "Stage Name",
    "field.jobName": "Job Name",
    "field.triggerType": "Trigger",
    "field.triggerBranch": "Branch",
    "field.triggerCommit": "Commit message contains",
    "field.triggerTag": "Tag name",
    "field.steps": "Execution steps",
    "field.stepAgent": "Agent",
    "field.stepCommands": "Commands",
    "field.jobType": "Job Type",
    "field.agentMode": "Agent Mode",
    "field.agentLabels": "Agent Labels",
    "field.commands": "Commands",
    "field.timeout": "Timeout Seconds",
    "field.requiredJob": "Required job",
    "field.pipeline": "Pipeline",
    "field.ref": "Ref",
    "field.comment": "Comment",
    "field.labels": "Labels",
    "field.maxRunning": "Concurrency",
    "field.sshUser": "SSH User",
    "field.sshHost": "SSH Host",
    "field.sshPort": "SSH Port",
    "field.reverseSSH": "Reverse SSH WS URL",
    "field.serverBaseURL": "Server URL",
    "help.jobType.title": "Job Type",
    "help.jobType.command": "Command",
    "help.jobType.commandDesc": "Run one or more shell commands on the selected agent.",
    "help.jobType.git": "Git Checkout",
    "help.jobType.gitDesc": "Clone or update a repository and checkout the selected ref.",
    "help.agentMode.title": "Agent Mode",
    "help.agentMode.single": "Single",
    "help.agentMode.singleDesc": "Choose one matching online agent and run the job once.",
    "help.agentMode.parallel": "Parallel",
    "help.agentMode.parallelDesc": "Run the same job on all matching online agents at the same time.",
    "help.agentMode.chain": "Chain",
    "help.agentMode.chainDesc": "Run on matching agents one by one. The next agent starts after the previous one succeeds.",
    "view.overview": "Overview",
    "view.projects": "Projects",
    "view.pipelines": "Pipelines",
    "view.runs": "Runs",
    "view.agents": "Agents",
    "view.settings": "Settings",
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
    "notice.agentRequired": "Add at least one agent before creating a pipeline step.",
    "notice.stepCommandsRequired": "Each step needs command content.",
    "notice.runStarted": "Run {id} started.",
    "notice.agentCreated": "Agent {id} created. Deployment script is ready.",
    "notice.agentDeleted": "Agent deleted.",
    "notice.agentDuplicate": "Agent name already exists.",
    "notice.scriptCopied": "Deployment command copied.",
    "notice.sshCommand": "SSH command: {command}",
    "notice.settingsSaved": "Settings saved.",
    "empty.projects": "Create a project to connect a repository.",
    "empty.pipelines": "Create a pipeline to define trigger rules and ordered Agent command steps.",
    "empty.runs": "Runs appear here after a manual, webhook, or scheduled trigger starts a pipeline.",
    "empty.agents": "Start an agent to make execution capacity available.",
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
    "th.ssh": "SSH",
    "th.actions": "Actions",
    "action.script": "Script",
    "action.ssh": "SSH",
    "action.run": "Run",
    "action.details": "Details",
    "action.copy": "Copy",
    "action.add": "Add",
    "action.addStep": "Add step",
    "action.cancel": "Cancel",
    "action.delete": "Delete",
    "action.remove": "Remove",
    "confirm.deleteAgent": "Delete agent {name}?",
    "trigger.manual": "Manual only",
    "trigger.push": "Push branch",
    "trigger.tag": "Tag name",
    "settings.title": "Haiwo Settings",
    "settings.hint": "Set the public http/https base URL. Haiwo appends fixed paths automatically.",
    "settings.save": "Save Settings",
    "settings.current": "Current Settings",
    "settings.currentHint": "Agents created after saving will inherit the generated reverse SSH endpoint.",
    "settings.reverseSSHEmpty": "Server URL is not configured.",
    "run.details": "Run Details",
    "run.error": "Error",
    "run.logs": "Logs",
    "run.noLogs": "No logs recorded yet.",
    "run.missingFailureLogs": "This run failed before error details were saved. Trigger it again to capture the exact failure after the logging fix.",
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
    "nav.settings": "设置",
    "nav.settingsHint": "系统设置",
    "sidebar.rpc": "基于 WebSocket 的 JSON-RPC",
    "top.eyebrow": "Haiwo MVP 控制台",
    "hero.eyebrow": "发布运维",
    "hero.title": "通过 Git 事件、评论、定时任务或手动操作发布。",
    "hero.body": "Agent 标签会把任务路由到构建、部署、预发或生产机器，Server 负责保存完整编排记录。",
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
    "pipeline.createHint": "选择触发方式，然后添加按顺序执行的 Agent 命令步骤。",
    "pipeline.listHint": "用于 push、tag、手动运行和顺序 Agent 命令的可复用工作流。",
    "run.start": "启动运行",
    "run.startHint": "使用 ref 或评论命令启动流水线。",
    "run.historyFilter": "运行历史",
    "run.historyFilterHint": "选择流水线查看历史执行情况。",
    "run.allPipelines": "全部流水线",
    "run.listHint": "执行记录每五秒自动刷新。点击详情查看日志。",
    "agent.listHint": "通过 WebSocket JSON-RPC 连接的工作节点。",
    "agent.create": "添加 Agent",
    "agent.createHint": "创建 Agent 身份并生成部署脚本。",
    "agent.deployScript": "部署脚本",
    "agent.deployScriptHint": "把 agent 二进制放到目标机器后，运行这段脚本。",
    "field.name": "名称",
    "field.provider": "代码平台",
    "field.repoURL": "仓库地址",
    "field.defaultBranch": "默认分支",
    "field.project": "项目",
    "field.pipelineJSON": "流水线 JSON",
    "field.pipelineName": "流水线名称",
    "field.stageName": "阶段名称",
    "field.jobName": "任务名称",
    "field.triggerType": "触发方式",
    "field.triggerBranch": "分支",
    "field.triggerCommit": "Commit 信息包含",
    "field.triggerTag": "Tag 名称",
    "field.steps": "执行步骤",
    "field.stepAgent": "Agent",
    "field.stepCommands": "命令内容",
    "field.jobType": "任务类型",
    "field.agentMode": "Agent 模式",
    "field.agentLabels": "Agent 标签",
    "field.commands": "命令",
    "field.timeout": "超时秒数",
    "field.requiredJob": "必要任务",
    "field.pipeline": "流水线",
    "field.ref": "Ref",
    "field.comment": "评论命令",
    "field.labels": "标签",
    "field.maxRunning": "并发数",
    "field.sshUser": "SSH 用户",
    "field.sshHost": "SSH 主机",
    "field.sshPort": "SSH 端口",
    "field.reverseSSH": "反向 SSH WS 地址",
    "field.serverBaseURL": "Server 域名",
    "help.jobType.title": "任务类型",
    "help.jobType.command": "Command",
    "help.jobType.commandDesc": "在选中的 Agent 上执行一条或多条 shell 命令。",
    "help.jobType.git": "Git Checkout",
    "help.jobType.gitDesc": "拉取或更新仓库，并切换到指定 ref。",
    "help.agentMode.title": "Agent 模式",
    "help.agentMode.single": "Single",
    "help.agentMode.singleDesc": "从匹配的在线 Agent 中选择一个，只执行一次任务。",
    "help.agentMode.parallel": "Parallel",
    "help.agentMode.parallelDesc": "在所有匹配的在线 Agent 上同时执行同一个任务。",
    "help.agentMode.chain": "Chain",
    "help.agentMode.chainDesc": "按顺序在匹配的 Agent 上执行；前一个成功后再执行下一个。",
    "view.overview": "概览",
    "view.projects": "项目",
    "view.pipelines": "流水线",
    "view.runs": "运行",
    "view.agents": "Agent",
    "view.settings": "设置",
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
    "notice.agentRequired": "请先添加至少一个 Agent，再创建流水线步骤。",
    "notice.stepCommandsRequired": "每个步骤都需要填写命令内容。",
    "notice.runStarted": "运行 {id} 已启动。",
    "notice.agentCreated": "Agent {id} 已创建，部署脚本已生成。",
    "notice.agentDeleted": "Agent 已删除。",
    "notice.agentDuplicate": "Agent 名称已存在。",
    "notice.scriptCopied": "部署命令已复制。",
    "notice.sshCommand": "SSH 命令：{command}",
    "notice.settingsSaved": "设置已保存。",
    "empty.projects": "创建项目后即可连接代码仓库。",
    "empty.pipelines": "创建流水线后即可定义触发规则和顺序 Agent 命令步骤。",
    "empty.runs": "通过手动、webhook 或定时触发启动流水线后，运行记录会出现在这里。",
    "empty.agents": "启动 agent 后即可提供任务执行容量。",
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
    "th.ssh": "SSH",
    "th.actions": "操作",
    "action.script": "脚本",
    "action.ssh": "SSH",
    "action.run": "运行",
    "action.details": "详情",
    "action.copy": "复制",
    "action.add": "添加",
    "action.addStep": "添加步骤",
    "action.cancel": "取消",
    "action.delete": "删除",
    "action.remove": "删除",
    "confirm.deleteAgent": "删除 Agent {name}？",
    "trigger.manual": "只手动触发",
    "trigger.push": "Push 分支",
    "trigger.tag": "Tag 名称",
    "settings.title": "Haiwo 系统设置",
    "settings.hint": "只填写公开 http/https 基础地址，Haiwo 会自动拼接固定路径。",
    "settings.save": "保存设置",
    "settings.current": "当前设置",
    "settings.currentHint": "保存后新创建的 Agent 会继承自动生成的反向 SSH 地址。",
    "settings.reverseSSHEmpty": "尚未配置 Server 域名。",
    "run.details": "运行详情",
    "run.error": "错误",
    "run.logs": "日志",
    "run.noLogs": "暂无日志记录。",
    "run.missingFailureLogs": "这次运行失败时没有保存错误详情。修复后重新触发一次即可记录准确失败原因。",
  },
  ja: {
    "brand.subtitle": "デリバリー制御プレーン",
    "nav.overview": "概要",
    "nav.overviewHint": "実行状況",
    "nav.projects": "プロジェクト",
    "nav.projectsHint": "リポジトリ",
    "nav.pipelines": "パイプライン",
    "nav.pipelinesHint": "オーケストレーション",
    "nav.runs": "実行履歴",
    "nav.runsHint": "履歴",
    "nav.agents": "Agent",
    "nav.agentsHint": "実行ノード",
    "nav.settings": "設定",
    "nav.settingsHint": "システム",
    "sidebar.rpc": "WebSocket JSON-RPC",
    "top.eyebrow": "Haiwo MVP コンソール",
    "hero.eyebrow": "リリース運用",
    "hero.title": "Git イベント、コメント、スケジュール、手動実行からデリバリー。",
    "hero.body": "Agent ラベルでビルド、デプロイ、ステージング、本番マシンへ作業をルーティングし、Server が実行記録を保持します。",
    "metric.projects": "プロジェクト",
    "metric.projectsHint": "登録済みリポジトリ",
    "metric.pipelines": "パイプライン",
    "metric.pipelinesHint": "デリバリーワークフロー",
    "metric.runs": "実行",
    "metric.agents": "オンライン Agent",
    "metric.noRuns": "実行履歴なし",
    "metric.allQuiet": "安定稼働中",
    "metric.active": "{count} 件が実行中",
    "metric.review": "{count} 件の確認が必要",
    "metric.capacity": "容量あり",
    "metric.noCapacity": "オンライン容量なし",
    "overview.recentRuns": "最近の実行",
    "overview.recentRunsHint": "最新のパイプライン実行とロールバック試行。",
    "overview.agentCapacity": "Agent 容量",
    "overview.agentCapacityHint": "スケジュールジョブを実行できるオンラインノード。",
    "project.create": "プロジェクト作成",
    "project.createHint": "リポジトリを Haiwo に紐付けます。",
    "project.listHint": "制御プレーンに登録されたソースリポジトリ。",
    "pipeline.create": "パイプライン作成",
    "pipeline.createHint": "トリガーを選び、順番に実行する Agent コマンドステップを追加します。",
    "pipeline.listHint": "push、tag、手動実行、順序付き Agent コマンド用の再利用可能なワークフロー。",
    "run.start": "実行開始",
    "run.startHint": "ref またはコメントコマンドでパイプラインを起動します。",
    "run.historyFilter": "実行履歴",
    "run.historyFilterHint": "パイプラインを選択して過去の実行を確認します。",
    "run.allPipelines": "すべてのパイプライン",
    "run.listHint": "実行履歴は 5 秒ごとに更新されます。詳細でログを確認できます。",
    "agent.listHint": "WebSocket JSON-RPC で接続された実行ノード。",
    "agent.create": "Agent 追加",
    "agent.createHint": "Agent ID とデプロイスクリプトを作成します。",
    "agent.deployScript": "デプロイスクリプト",
    "agent.deployScriptHint": "対象マシンに agent バイナリを配置してから実行してください。",
    "field.name": "名前",
    "field.provider": "プロバイダー",
    "field.repoURL": "リポジトリ URL",
    "field.defaultBranch": "デフォルトブランチ",
    "field.project": "プロジェクト",
    "field.pipelineJSON": "パイプライン JSON",
    "field.pipelineName": "パイプライン名",
    "field.stageName": "Stage 名",
    "field.jobName": "Job 名",
    "field.triggerType": "トリガー",
    "field.triggerBranch": "ブランチ",
    "field.triggerCommit": "Commit メッセージに含む",
    "field.triggerTag": "Tag 名",
    "field.steps": "実行ステップ",
    "field.stepAgent": "Agent",
    "field.stepCommands": "コマンド",
    "field.jobType": "Job タイプ",
    "field.agentMode": "Agent モード",
    "field.agentLabels": "Agent ラベル",
    "field.commands": "コマンド",
    "field.timeout": "タイムアウト秒数",
    "field.requiredJob": "必須 Job",
    "field.pipeline": "パイプライン",
    "field.ref": "Ref",
    "field.comment": "コメントコマンド",
    "field.labels": "ラベル",
    "field.maxRunning": "同時実行数",
    "field.sshUser": "SSH ユーザー",
    "field.sshHost": "SSH ホスト",
    "field.sshPort": "SSH ポート",
    "field.reverseSSH": "リバース SSH WS URL",
    "field.serverBaseURL": "Server URL",
    "help.jobType.title": "Job タイプ",
    "help.jobType.command": "Command",
    "help.jobType.commandDesc": "選択した Agent 上で 1 つ以上の shell コマンドを実行します。",
    "help.jobType.git": "Git Checkout",
    "help.jobType.gitDesc": "リポジトリを clone または更新し、指定した ref に checkout します。",
    "help.agentMode.title": "Agent モード",
    "help.agentMode.single": "Single",
    "help.agentMode.singleDesc": "一致するオンライン Agent から 1 台を選び、Job を 1 回実行します。",
    "help.agentMode.parallel": "Parallel",
    "help.agentMode.parallelDesc": "一致するすべてのオンライン Agent で同じ Job を同時に実行します。",
    "help.agentMode.chain": "Chain",
    "help.agentMode.chainDesc": "一致する Agent で順番に実行します。前の Agent が成功した後に次へ進みます。",
    "view.overview": "概要",
    "view.projects": "プロジェクト",
    "view.pipelines": "パイプライン",
    "view.runs": "実行履歴",
    "view.agents": "Agent",
    "view.settings": "設定",
    "hero.noAgents": "オンライン Agent なし",
    "hero.noAgentsHint": "Agent を起動するとパイプラインタスクを受け付けられます。",
    "hero.failed": "失敗した実行を確認",
    "hero.failedHint": "{count} 件の実行またはロールバックが失敗しました。",
    "hero.running": "デリバリー実行中",
    "hero.runningHint": "{count} 件がキューまたは実行中です。",
    "hero.ready": "デプロイ可能",
    "hero.readyHint": "{count} 件の agent が待機中です。",
    "sync.never": "未同期",
    "sync.now": "{time} に同期",
    "sidebar.onlineAgents": "{count} 件のオンライン agent",
    "notice.projectCreated": "プロジェクトを作成しました。",
    "notice.projectRequired": "パイプラインを追加する前にプロジェクトを作成してください。",
    "notice.pipelineInvalid": "パイプライン JSON が無効です: {message}",
    "notice.pipelineCreated": "パイプラインを作成しました。",
    "notice.pipelineRequired": "実行を開始する前にパイプラインを作成してください。",
    "notice.agentRequired": "パイプラインステップを作成する前に Agent を追加してください。",
    "notice.stepCommandsRequired": "各ステップにコマンドを入力してください。",
    "notice.runStarted": "実行 {id} を開始しました。",
    "notice.agentCreated": "Agent {id} を作成しました。デプロイスクリプトを生成しました。",
    "notice.agentDeleted": "Agent を削除しました。",
    "notice.agentDuplicate": "Agent 名はすでに存在します。",
    "notice.scriptCopied": "デプロイコマンドをコピーしました。",
    "notice.sshCommand": "SSH コマンド: {command}",
    "notice.settingsSaved": "設定を保存しました。",
    "empty.projects": "プロジェクトを作成するとリポジトリを接続できます。",
    "empty.pipelines": "パイプラインを作成するとトリガールールと順序付き Agent コマンドステップを定義できます。",
    "empty.runs": "手動、webhook、スケジュールでパイプラインを起動すると実行履歴がここに表示されます。",
    "empty.agents": "Agent を起動すると実行容量を利用できます。",
    "empty.capacity": "Agent はまだ登録されていません。起動コマンド:",
    "th.name": "名前",
    "th.provider": "プロバイダー",
    "th.repository": "リポジトリ",
    "th.defaultBranch": "デフォルトブランチ",
    "th.id": "ID",
    "th.project": "プロジェクト",
    "th.stages": "Stage",
    "th.jobs": "Job",
    "th.run": "実行",
    "th.pipeline": "パイプライン",
    "th.status": "ステータス",
    "th.source": "ソース",
    "th.ref": "Ref",
    "th.updated": "更新日時",
    "th.labels": "ラベル",
    "th.load": "負荷",
    "th.version": "バージョン",
    "th.lastSeen": "最終確認",
    "th.ssh": "SSH",
    "th.actions": "操作",
    "action.script": "スクリプト",
    "action.ssh": "SSH",
    "action.run": "実行",
    "action.details": "詳細",
    "action.copy": "コピー",
    "action.add": "追加",
    "action.addStep": "ステップ追加",
    "action.cancel": "キャンセル",
    "action.delete": "削除",
    "action.remove": "削除",
    "confirm.deleteAgent": "Agent {name} を削除しますか？",
    "trigger.manual": "手動のみ",
    "trigger.push": "Push ブランチ",
    "trigger.tag": "Tag 名",
    "settings.title": "Haiwo システム設定",
    "settings.hint": "公開 http/https ベース URL だけを設定します。固定パスは Haiwo が自動で追加します。",
    "settings.save": "設定を保存",
    "settings.current": "現在の設定",
    "settings.currentHint": "保存後に作成される Agent は生成されたリバース SSH エンドポイントを継承します。",
    "settings.reverseSSHEmpty": "Server URL は未設定です。",
    "run.details": "実行詳細",
    "run.error": "エラー",
    "run.logs": "ログ",
    "run.noLogs": "ログはまだ記録されていません。",
    "run.missingFailureLogs": "この実行はエラー詳細が保存される前に失敗しました。修正後に再実行すると正確な原因を記録できます。",
  },
};

let currentLang = initialLanguage();
let lastSyncTime = "";
let pipelineSteps = [{ id: crypto.randomUUID(), agent_id: "", commands: "echo deploy staging", timeout_seconds: 300 }];
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
document.getElementById("pipeline-project").addEventListener("change", syncPipelineRef);
document.getElementById("pipeline-trigger-type").addEventListener("change", renderTriggerFields);
document.getElementById("add-pipeline-step").addEventListener("click", addPipelineStep);
document.getElementById("pipeline-steps").addEventListener("click", pipelineStepAction);
document.getElementById("pipeline-steps").addEventListener("input", updatePipelineStepsFromDOM);
document.getElementById("pipeline-steps").addEventListener("change", updatePipelineStepsFromDOM);
document.getElementById("run-pipeline-filter").addEventListener("change", (event) => {
  state.selectedRunPipelineId = event.currentTarget.value;
  state.selectedRunId = "";
  renderRuns();
});
document.getElementById("agent-form").addEventListener("submit", createAgent);
document.getElementById("settings-form").addEventListener("submit", saveSettings);
document.getElementById("agents-table").addEventListener("click", agentTableAction);
document.getElementById("pipelines-table").addEventListener("click", pipelineTableAction);
document.getElementById("runs-table").addEventListener("click", runTableAction);
document.getElementById("recent-runs").addEventListener("click", runTableAction);
document.querySelectorAll("[data-create-target]").forEach((button) => {
  button.addEventListener("click", () => showCreatePanel(button.dataset.createTarget));
});
document.querySelectorAll("[data-create-cancel]").forEach((button) => {
  button.addEventListener("click", () => closeCreatePanel(button.dataset.createCancel));
});

resetProjectFormDefaults();
resetPipelineFormDefaults();
resetAgentFormDefaults();
refresh();
setInterval(refresh, 5000);

function showView(name) {
  document.querySelectorAll(".view").forEach((view) => view.classList.add("hidden"));
  document.getElementById(`${name}-view`).classList.remove("hidden");
  document.getElementById("view-title").textContent = t(`view.${name}`);
  if (name === "agents") {
    resetAgentFormDefaults();
  }
}

async function refresh() {
  try {
    const [projects, pipelines, runs, agents, settings] = await Promise.all([
      api("/api/projects"),
      api("/api/pipelines"),
      api("/api/runs"),
      api("/api/agents"),
      api("/api/settings"),
    ]);
    Object.assign(state, { projects, pipelines, runs, agents, settings });
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
  notify(t("notice.projectCreated"));
  await refresh();
  resetProjectFormDefaults();
  closeCreatePanel("projects");
}

async function createPipeline(event) {
  event.preventDefault();
  const form = new FormData(event.currentTarget);
  const projectID = form.get("project_id");
  if (!projectID) {
    notify(t("notice.projectRequired"), true);
    return;
  }
  updatePipelineStepsFromDOM();
  if (!state.agents.length) {
    notify(t("notice.agentRequired"), true);
    return;
  }
  const project = state.projects.find((item) => item.id === projectID);
  const validSteps = pipelineSteps
    .map((step) => ({ ...step, commands: String(step.commands || "").trim() }))
    .filter((step) => step.agent_id && step.commands);
  if (validSteps.length !== pipelineSteps.length || !validSteps.length) {
    notify(t("notice.stepCommandsRequired"), true);
    return;
  }
  const definition = {
    name: form.get("pipeline_name"),
    stages: validSteps.map((step, index) => ({
      name: `step-${index + 1}`,
      jobs: [
        {
          id: slug(`step-${index + 1}-${step.agent_id}`),
          name: stepName(step, index),
          job_type: "command",
          agent_mode: "single",
          agent_ids: [step.agent_id],
          commands: splitLines(step.commands),
          timeout_seconds: Number(step.timeout_seconds || 300),
          required: true,
          repo: project?.repo_url ? { url: project.repo_url, ref: project.default_branch || "main" } : undefined,
        },
      ],
    })),
  };
  const pipeline = await api(`/api/projects/${projectID}/pipelines`, { method: "POST", body: definition });
  await createPipelineTrigger(projectID, pipeline.id, form, project);
  notify(t("notice.pipelineCreated"));
  await refresh();
  resetPipelineFormDefaults();
  closeCreatePanel("pipelines");
}

async function createPipelineTrigger(projectID, pipelineID, form, project) {
  const triggerType = form.get("trigger_type");
  if (triggerType === "manual") {
    return;
  }
  const body = {
    pipeline_id: pipelineID,
    type: triggerType,
  };
  if (triggerType === "push") {
    body.branch_pattern = String(form.get("trigger_branch") || project?.default_branch || "main").trim();
    body.commit_pattern = String(form.get("trigger_commit") || "").trim();
  }
  if (triggerType === "tag") {
    body.tag_pattern = String(form.get("trigger_tag") || "").trim();
  }
  await api(`/api/projects/${projectID}/triggers`, { method: "POST", body });
}

function addPipelineStep() {
  updatePipelineStepsFromDOM();
  const firstAgentID = state.agents[0]?.id || "";
  pipelineSteps.push({ id: crypto.randomUUID(), agent_id: firstAgentID, commands: "", timeout_seconds: 300 });
  renderPipelineSteps();
}

function pipelineStepAction(event) {
  const button = event.target.closest("[data-step-action]");
  if (!button) return;
  updatePipelineStepsFromDOM();
  const stepID = button.dataset.stepId;
  if (button.dataset.stepAction === "remove" && pipelineSteps.length > 1) {
    pipelineSteps = pipelineSteps.filter((step) => step.id !== stepID);
    renderPipelineSteps();
  }
}

function updatePipelineStepsFromDOM() {
  const rows = [...document.querySelectorAll("[data-pipeline-step]")];
  if (!rows.length) return;
  pipelineSteps = rows.map((row) => ({
    id: row.dataset.pipelineStep,
    agent_id: row.querySelector('[name="step_agent"]')?.value || "",
    commands: row.querySelector('[name="step_commands"]')?.value || "",
    timeout_seconds: Number(row.querySelector('[name="step_timeout"]')?.value || 300),
  }));
}

async function createAgent(event) {
  event.preventDefault();
  const formElement = event.currentTarget;
  const form = new FormData(formElement);
  const payload = {
    name: form.get("name"),
    labels: checkedValues("agent-labels"),
    max_running: Number(form.get("max_running") || 1),
    ssh_enabled: true,
  };
  try {
    const result = await api("/api/agents", { method: "POST", body: payload });
    formElement.reset();
    resetAgentFormDefaults();
    closeCreatePanel("agents");
    notify(t("notice.agentCreated", { id: result.agent.id }));
    await refresh();
  } catch (error) {
    const message = error.message.includes("agent name already exists") ? t("notice.agentDuplicate") : error.message;
    notify(message, true);
  }
}

async function saveSettings(event) {
  event.preventDefault();
  const form = new FormData(event.currentTarget);
  const settings = await api("/api/settings", { method: "PUT", body: { server_base_url: form.get("server_base_url") } });
  state.settings = settings;
  renderSettings();
  notify(t("notice.settingsSaved"));
}

async function agentTableAction(event) {
  const button = event.target.closest("[data-agent-action]");
  if (!button) return;
  const agentID = button.dataset.agentId;
  if (button.dataset.agentAction === "script") {
    const result = await api(`/api/agents/${agentID}/deploy-script`);
    await copyText(result.command || result.url);
    notify(t("notice.scriptCopied"));
  }
  if (button.dataset.agentAction === "ssh") {
    window.open(`/ssh/agents/${agentID}`, "_blank", "noopener");
  }
  if (button.dataset.agentAction === "delete") {
    const agent = state.agents.find((item) => item.id === agentID);
    if (!confirm(t("confirm.deleteAgent", { name: agent?.name || agentID }))) {
      return;
    }
    await api(`/api/agents/${agentID}`, { method: "DELETE" });
    notify(t("notice.agentDeleted"));
    await refresh();
  }
}

async function pipelineTableAction(event) {
  const button = event.target.closest("[data-pipeline-action]");
  if (!button) return;
  const pipelineID = button.dataset.pipelineId;
  const pipeline = state.pipelines.find((item) => item.id === pipelineID);
  const project = pipeline ? state.projects.find((item) => item.id === pipeline.project_id) : null;
  const ref = project?.default_branch || "main";
  const run = await api(`/api/pipelines/${pipelineID}/runs`, { method: "POST", body: { type: "manual", ref } });
  notify(t("notice.runStarted", { id: run.id }));
  await refresh();
}

function runTableAction(event) {
  const button = event.target.closest("[data-run-action]");
  if (!button) return;
  if (button.dataset.runAction === "details") {
    state.selectedRunId = state.selectedRunId === button.dataset.runId ? "" : button.dataset.runId;
    renderRuns();
  }
}

function toggleHelp(event) {
  const button = event.target.closest("[data-help]");
  if (!button) return;
  const target = document.getElementById(`${button.dataset.help}-help`);
  if (!target) return;
  target.classList.toggle("hidden");
  button.classList.toggle("active", !target.classList.contains("hidden"));
}

function showCreatePanel(name) {
  if (name === "projects") resetProjectFormDefaults();
  if (name === "pipelines") resetPipelineFormDefaults();
  if (name === "agents") {
    resetAgentFormDefaults();
  }
  document.querySelector(`[data-create-panel="${name}"]`)?.classList.remove("hidden");
}

function closeCreatePanel(name) {
  document.querySelector(`[data-create-panel="${name}"]`)?.classList.add("hidden");
  if (name === "projects") resetProjectFormDefaults();
  if (name === "pipelines") resetPipelineFormDefaults();
  if (name === "agents") resetAgentFormDefaults();
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
  renderRunPipelineFilter();
  syncPipelineRef(false);
  updatePipelineStepsFromDOM();
  renderTriggerFields();
  renderPipelineSteps();

  renderProjects();
  renderPipelines();
  renderRuns();
  renderAgents();
  renderSettings();
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
    [t("th.name"), t("th.project"), t("th.stages"), t("th.jobs"), t("th.id"), t("th.actions")],
    state.pipelines.map((p) => [
      strong(p.name),
      shortName(state.projects.find((x) => x.id === p.project_id)),
      p.stages?.length || 0,
      (p.stages || []).reduce((sum, stage) => sum + (stage.jobs?.length || 0), 0),
      code(p.id),
      pipelineActions(p),
    ]),
    t("empty.pipelines")
  );
}

function renderRuns() {
  const runs = state.runs
    .filter((run) => !state.selectedRunPipelineId || run.pipeline_id === state.selectedRunPipelineId)
    .reverse();
  if (state.selectedRunId && !runs.some((run) => run.id === state.selectedRunId)) {
    state.selectedRunId = "";
  }
  document.getElementById("runs-table").innerHTML = runTable(runs) + runDetails();
}

function renderRunPipelineFilter() {
  const select = document.getElementById("run-pipeline-filter");
  const selected = state.pipelines.some((pipeline) => pipeline.id === state.selectedRunPipelineId) ? state.selectedRunPipelineId : "";
  state.selectedRunPipelineId = selected;
  select.innerHTML = [
    `<option value="">${escapeHTML(t("run.allPipelines"))}</option>`,
    ...state.pipelines.map((pipeline) => `<option value="${escapeHTML(pipeline.id)}">${escapeHTML(`${pipeline.name} (${pipeline.id})`)}</option>`),
  ].join("");
  select.value = selected;
}

function renderAgents() {
  document.getElementById("agents-table").innerHTML = table(
    [t("th.name"), t("th.status"), t("th.labels"), t("th.load"), t("th.ssh"), t("th.version"), t("th.lastSeen"), t("th.actions")],
    state.agents.map((a) => [
      strong(a.name),
      status(a.status),
      labels(a.labels || []),
      `${a.current_run}/${a.max_running}`,
      sshState(a),
      escapeHTML(a.version),
      date(a.last_seen_at),
      agentActions(a),
    ]),
    t("empty.agents")
  );
}

function renderSettings() {
  const form = document.getElementById("settings-form");
  const baseURL = state.settings?.server_base_url || "";
  const reverseURL = state.settings?.reverse_ssh_url || "";
  if (document.activeElement !== form.server_base_url) {
    form.server_base_url.value = baseURL;
  }
  document.getElementById("settings-summary").innerHTML = baseURL
    ? `<div class="setting-item"><span>${escapeHTML(t("field.serverBaseURL"))}</span>${code(baseURL)}<span>${escapeHTML(t("field.reverseSSH"))}</span>${code(reverseURL)}</div>`
    : `<div class="empty">${escapeHTML(t("settings.reverseSSHEmpty"))}</div>`;
}

function renderTriggerFields() {
  const form = document.getElementById("pipeline-form");
  const target = document.getElementById("pipeline-trigger-fields");
  if (!form || !target) return;
  const triggerType = form.trigger_type.value || "manual";
  const project = state.projects.find((item) => item.id === form.project_id.value);
  const oldBranch = target.querySelector('[name="trigger_branch"]')?.value;
  const oldCommit = target.querySelector('[name="trigger_commit"]')?.value;
  const oldTag = target.querySelector('[name="trigger_tag"]')?.value;
  if (triggerType === "push") {
    target.innerHTML = `
      <label><span>${escapeHTML(t("field.triggerBranch"))}</span><input name="trigger_branch" value="${escapeHTML(oldBranch || project?.default_branch || "main")}" /></label>
      <label><span>${escapeHTML(t("field.triggerCommit"))}</span><input name="trigger_commit" value="${escapeHTML(oldCommit || "")}" placeholder="deploy" /></label>
    `;
    return;
  }
  if (triggerType === "tag") {
    target.innerHTML = `<label><span>${escapeHTML(t("field.triggerTag"))}</span><input name="trigger_tag" value="${escapeHTML(oldTag || "")}" placeholder="dev" /></label>`;
    return;
  }
  target.innerHTML = "";
}

function renderPipelineSteps() {
  const target = document.getElementById("pipeline-steps");
  if (!target) return;
  if (!pipelineSteps.length) {
    pipelineSteps = [{ id: crypto.randomUUID(), agent_id: state.agents[0]?.id || "", commands: "", timeout_seconds: 300 }];
  }
  pipelineSteps = pipelineSteps.map((step) => ({ ...step, agent_id: step.agent_id || state.agents[0]?.id || "" }));
  target.innerHTML = pipelineSteps.map((step, index) => stepCard(step, index)).join("");
}

function stepCard(step, index) {
  const remove = pipelineSteps.length > 1 ? `<button class="table-action" type="button" data-step-action="remove" data-step-id="${escapeHTML(step.id)}">${escapeHTML(t("action.remove"))}</button>` : "";
  return `
    <div class="step-card" data-pipeline-step="${escapeHTML(step.id)}">
      <div class="step-card-head">
        <strong>${index + 1}</strong>
        ${remove}
      </div>
      <label><span>${escapeHTML(t("field.stepAgent"))}</span><select name="step_agent">${agentOptions(step.agent_id)}</select></label>
      <label><span>${escapeHTML(t("field.stepCommands"))}</span><textarea name="step_commands" rows="4" spellcheck="false">${escapeHTML(step.commands || "")}</textarea></label>
      <label><span>${escapeHTML(t("field.timeout"))}</span><input name="step_timeout" type="number" min="0" value="${Number(step.timeout_seconds || 300)}" /></label>
    </div>
  `;
}

function agentOptions(selectedID) {
  if (!state.agents.length) {
    return `<option value="">${escapeHTML(t("empty.agents"))}</option>`;
  }
  return state.agents
    .map((agent) => `<option value="${escapeHTML(agent.id)}" ${agent.id === selectedID ? "selected" : ""}>${escapeHTML(agent.name)} (${escapeHTML(agent.id)})</option>`)
    .join("");
}

function stepName(step, index) {
  const agent = state.agents.find((item) => item.id === step.agent_id);
  return `${index + 1}. ${agent?.name || step.agent_id}`;
}

function renderHelpPanels() {
  document.getElementById("job-type-help").innerHTML = helpPanel("help.jobType.title", [
    ["help.jobType.command", "help.jobType.commandDesc"],
    ["help.jobType.git", "help.jobType.gitDesc"],
  ]);
  document.getElementById("agent-mode-help").innerHTML = helpPanel("help.agentMode.title", [
    ["help.agentMode.single", "help.agentMode.singleDesc"],
    ["help.agentMode.parallel", "help.agentMode.parallelDesc"],
    ["help.agentMode.chain", "help.agentMode.chainDesc"],
  ]);
}

function helpPanel(titleKey, rows) {
  return `<h4>${escapeHTML(t(titleKey))}</h4><dl>${rows
    .map(([label, description]) => `<div><dt>${escapeHTML(t(label))}</dt><dd>${escapeHTML(t(description))}</dd></div>`)
    .join("")}</dl>`;
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
    [t("th.run"), t("th.project"), t("th.pipeline"), t("th.status"), t("th.source"), t("th.ref"), t("th.updated"), t("th.actions")],
    runs.map((r) => [
      code(r.id),
      shortName(state.projects.find((p) => p.id === r.project_id)),
      shortName(state.pipelines.find((p) => p.id === r.pipeline_id)),
      status(r.status),
      escapeHTML(r.source),
      code(r.ref || "-"),
      date(r.updated_at),
      runActions(r),
    ]),
    t("empty.runs")
  );
}

function runActions(run) {
  return `<button class="table-action" type="button" data-run-action="details" data-run-id="${escapeHTML(run.id)}">${escapeHTML(t("action.details"))}</button>`;
}

function runDetails() {
  if (!state.selectedRunId) return "";
  const run = state.runs.find((item) => item.id === state.selectedRunId);
  if (!run) return "";
  const metadata = run.metadata || {};
  const logs = metadata.logs || "";
  const error = metadata.error || "";
  const logText = logs || (isFailedRun(run) ? t("run.missingFailureLogs") : t("run.noLogs"));
  return `<div class="run-details">
    <div class="panel-header">
      <div>
        <h3>${escapeHTML(t("run.details"))}</h3>
        <p>${escapeHTML(run.id)}</p>
      </div>
    </div>
    ${error ? `<div class="run-error"><strong>${escapeHTML(t("run.error"))}</strong><span>${escapeHTML(error)}</span></div>` : ""}
    <div class="run-log-title">${escapeHTML(t("run.logs"))}</div>
    <pre class="run-log">${escapeHTML(logText)}</pre>
  </div>`;
}

function isFailedRun(run) {
  return run.status === "failed" || run.status === "rollback_failed";
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

function syncPipelineRef(force = true) {
  const select = document.getElementById("pipeline-project");
  const input = document.querySelector('#pipeline-form [name="job_ref"]');
  const project = state.projects.find((item) => item.id === select.value);
  if (!input || !project?.default_branch) return;
  if (force || !input.value.trim()) {
    input.value = project.default_branch;
  }
}

function notify(message, error = false) {
  const notice = document.getElementById("notice");
  notice.textContent = message;
  notice.style.borderLeftColor = error ? "var(--red)" : "var(--accent)";
  notice.classList.remove("hidden");
  setTimeout(() => notice.classList.add("hidden"), 3200);
}

async function copyText(value) {
  if (navigator.clipboard?.writeText) {
    try {
      await navigator.clipboard.writeText(value);
      return;
    } catch (_) {
      // Fall back to the legacy selection API when clipboard permission is unavailable.
    }
  }
  const textarea = document.createElement("textarea");
  textarea.value = value;
  textarea.setAttribute("readonly", "");
  textarea.style.position = "fixed";
  textarea.style.left = "-9999px";
  document.body.appendChild(textarea);
  textarea.select();
  document.execCommand("copy");
  textarea.remove();
}

function applyI18n() {
  document.documentElement.lang = currentLang === "zh" ? "zh-CN" : currentLang === "ja" ? "ja" : "en";
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

function sshState(agent) {
  if (agent.ssh_enabled && agent.ssh_host) {
    const user = agent.ssh_user || "root";
    const port = agent.ssh_port || 22;
    return code(`${user}@${agent.ssh_host}:${port}`);
  }
  if (agent.reverse_ssh_url) {
    return badge("reverse-ws");
  }
  return "-";
}

function agentActions(agent) {
  const script = `<button class="table-action" type="button" data-agent-action="script" data-agent-id="${escapeHTML(agent.id)}">${escapeHTML(t("action.script"))}</button>`;
  const ssh = `<button class="table-action" type="button" data-agent-action="ssh" data-agent-id="${escapeHTML(agent.id)}">${escapeHTML(t("action.ssh"))}</button>`;
  const remove = `<button class="table-action danger" type="button" data-agent-action="delete" data-agent-id="${escapeHTML(agent.id)}">${escapeHTML(t("action.delete"))}</button>`;
  return `<div class="row-actions">${script}${ssh}${remove}</div>`;
}

function pipelineActions(pipeline) {
  return `<button class="table-action" type="button" data-pipeline-action="run" data-pipeline-id="${escapeHTML(pipeline.id)}">${escapeHTML(t("action.run"))}</button>`;
}

function splitList(value) {
  return String(value || "")
    .split(",")
    .map((part) => part.trim())
    .filter(Boolean);
}

function splitLines(value) {
  return String(value || "")
    .split(/\r?\n/)
    .map((part) => part.trim())
    .filter(Boolean);
}

function checkedValues(group) {
  return [...document.querySelectorAll(`[data-checkbox-group="${group}"] input:checked`)].map((input) => input.value);
}

function setCheckedValues(group, values) {
  const wanted = new Set(values);
  document.querySelectorAll(`[data-checkbox-group="${group}"] input`).forEach((input) => {
    input.checked = wanted.has(input.value);
  });
}

function resetProjectFormDefaults() {
  const form = document.getElementById("project-form");
  if (!form) return;
  form.reset();
  form.name.value = "";
  form.provider.value = "github";
  form.repo_url.value = "";
  form.default_branch.value = "main";
}

function resetPipelineFormDefaults() {
  const form = document.getElementById("pipeline-form");
  if (!form) return;
  form.pipeline_name.value = "staging deploy";
  form.trigger_type.value = "manual";
  pipelineSteps = [{ id: crypto.randomUUID(), agent_id: state.agents[0]?.id || "", commands: "echo deploy staging", timeout_seconds: 300 }];
  syncPipelineRef(true);
  renderTriggerFields();
  renderPipelineSteps();
}

function resetAgentFormDefaults() {
  const form = document.getElementById("agent-form");
  if (!form) return;
  form.reset();
  form.name.value = "";
  form.max_running.value = "1";
  setCheckedValues("agent-labels", ["build", "deploy"]);
}

function slug(value) {
  const normalized = String(value || "job")
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-+|-+$/g, "");
  return normalized || "job";
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
  if (saved === "zh" || saved === "ja" || saved === "en") {
    return saved;
  }
  const browserLang = (navigator.language || navigator.userLanguage || "").toLowerCase();
  if (browserLang.startsWith("zh")) return "zh";
  if (browserLang.startsWith("ja")) return "ja";
  return "en";
}

function t(key, vars = {}) {
  let value = translations[currentLang][key] || translations.en[key] || key;
  for (const [name, replacement] of Object.entries(vars)) {
    value = value.replaceAll(`{${name}}`, replacement);
  }
  return value;
}
