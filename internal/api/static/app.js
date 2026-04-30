const TEMPLATE_OPTION_GROUPS = [
  {
    label: "SubConv Next",
    options: [
      {
        key: "none",
        label: "跟随当前服务模板",
        description:
          "模板模式下跟随后端 service.template 的 lite / standard / full 预设。",
        ruleProfile: "跟随服务模板",
        groupProfile: "自动继承",
      },
      {
        key: "subconv_lite",
        label: "SubConv Next Lite",
        description: "只保留基础代理组，不展开业务分流规则。",
        ruleProfile: "基础代理组",
        groupProfile: "紧凑代理组",
      },
      {
        key: "subconv_standard",
        label: "SubConv Next Standard",
        description: "标准日常分流模板，覆盖常用国际服务和基础业务流量。",
        ruleProfile: "标准分流",
        groupProfile: "紧凑代理组",
      },
      {
        key: "subconv_full",
        label: "SubConv Next Full",
        description: "完整规则覆盖，并为业务组提供完整节点选择。",
        ruleProfile: "全量分流",
        groupProfile: "全量代理组",
      },
    ],
  },
  {
    label: "CM_Online 兼容",
    options: [
      {
        key: "cm_online",
        label: "CM_Online 默认版",
        description:
          "偏向常用国际访问场景，覆盖 Google、Telegram、微软、苹果等常见业务。",
        ruleProfile: "通用国际分流",
        groupProfile: "紧凑代理组",
      },
      {
        key: "cm_online_game",
        label: "CM_Online_Game",
        description: "在通用分流基础上增加游戏平台和主流流媒体分流。",
        ruleProfile: "游戏 + 流媒体",
        groupProfile: "紧凑代理组",
      },
      {
        key: "cm_online_multi_country",
        label: "CM_Online_MultiCountry",
        description: "启用地区组，适合需要按国家快速切换节点的场景。",
        ruleProfile: "多地区通用分流",
        groupProfile: "地区代理组",
      },
      {
        key: "cm_online_multi_country_cf",
        label: "CM_Online_MultiCountry_CF",
        description:
          "多地区模板，额外补充云服务分流，适合 Worker / CDN 相关节点。",
        ruleProfile: "多地区 + 云服务",
        groupProfile: "地区代理组",
      },
      {
        key: "cm_online_full",
        label: "CM_Online_Full",
        description: "完整规则覆盖和完整代理组展开。",
        ruleProfile: "完整分流",
        groupProfile: "全量代理组",
      },
      {
        key: "cm_online_full_cf",
        label: "CM_Online_Full_CF",
        description: "完整规则覆盖，额外补充云服务和 CDN 相关分流。",
        ruleProfile: "完整分流 + 云服务",
        groupProfile: "全量代理组",
      },
    ],
  },
  {
    label: "ACL4SSR 兼容",
    options: [
      {
        key: "acl4ssr_online_mini",
        label: "ACL4SSR_Online_Mini",
        description: "精简国际访问模板，只保留基础常用业务。",
        ruleProfile: "精简分流",
        groupProfile: "紧凑代理组",
      },
      {
        key: "acl4ssr_online",
        label: "ACL4SSR_Online",
        description: "兼顾广告拦截、常用国际服务和流媒体的通用模板。",
        ruleProfile: "通用分流 + 广告",
        groupProfile: "紧凑代理组",
      },
      {
        key: "acl4ssr_online_adblock",
        label: "ACL4SSR_Online_AdblockPlus",
        description: "在通用模板基础上扩展广告和更多业务分流。",
        ruleProfile: "广告增强分流",
        groupProfile: "紧凑代理组",
      },
      {
        key: "acl4ssr_online_full",
        label: "ACL4SSR_Online_Full",
        description: "完整业务分流模板，适合偏向大而全的配置风格。",
        ruleProfile: "完整分流",
        groupProfile: "全量代理组",
      },
    ],
  },
  {
    label: "BlackMatrix7 风格",
    options: [
      {
        key: "blackmatrix7_basic",
        label: "BlackMatrix7 Basic",
        description: "偏向开发、云服务和基础国际业务分流。",
        ruleProfile: "开发 / 云服务",
        groupProfile: "紧凑代理组",
      },
      {
        key: "blackmatrix7_streaming",
        label: "BlackMatrix7 Streaming",
        description: "以视频、音乐和国际访问为主的模板。",
        ruleProfile: "流媒体优先",
        groupProfile: "地区代理组",
      },
      {
        key: "blackmatrix7_global",
        label: "BlackMatrix7 Global",
        description: "全量规则和全量节点选择，适合重度分流场景。",
        ruleProfile: "全量分流",
        groupProfile: "全量代理组",
      },
    ],
  },
  {
    label: "兼容保留",
    options: [
      {
        key: "custom_url",
        label: "自定义外部模板 URL",
        description:
          "仅保留兼容字段。当前仍使用内置 renderer 生成，不直接替换本地模板。",
        ruleProfile: "兼容保留",
        groupProfile: "回退内置模板",
      },
    ],
  },
];

const TEMPLATE_OPTIONS = TEMPLATE_OPTION_GROUPS.flatMap((group) =>
  group.options.map((item) => ({ ...item, group: group.label })),
);

const BUILTIN_RULES = [
  { key: "adblock", label: "广告拦截" },
  { key: "microsoft", label: "微软服务" },
  { key: "ai", label: "AI 服务" },
  { key: "apple", label: "苹果服务" },
  { key: "bilibili", label: "哔哩哔哩" },
  { key: "social", label: "社交媒体" },
  { key: "youtube", label: "油管视频" },
  { key: "streaming", label: "流媒体" },
  { key: "google", label: "谷歌服务" },
  { key: "gaming", label: "游戏平台" },
  { key: "private", label: "私有网络" },
  { key: "education", label: "教育资源" },
  { key: "domestic", label: "国内服务" },
  { key: "finance", label: "金融服务" },
  { key: "telegram", label: "电报消息" },
  { key: "cloud", label: "云服务" },
  { key: "github", label: "Github" },
  { key: "non_cn", label: "非中国" },
];

const RULE_PRESETS = {
  minimal: ["private", "domestic", "non_cn"],
  balanced: [
    "private",
    "domestic",
    "microsoft",
    "apple",
    "google",
    "github",
    "ai",
    "telegram",
    "streaming",
    "non_cn",
  ],
  full: BUILTIN_RULES.map((item) => item.key),
  custom: [],
};

const OUTPUT_OPTIONS = [
  {
    key: "emoji",
    id: "opt-emoji",
    title: "地区 Emoji（兼容）",
    desc: "V1 默认保持节点原名，不再自动添加地区 Emoji",
    default: false,
    icon: "spark",
  },
  {
    key: "show_node_type",
    id: "opt-show-type",
    title: "显示节点类型（兼容）",
    desc: "V1 默认保持节点原名，不再自动追加协议类型",
    default: false,
    icon: "tag",
  },
  {
    key: "source_prefix",
    id: "opt-source-prefix",
    title: "显示订阅源标识",
    desc: "在节点名前添加 Emoji + 订阅源名称，便于区分来源",
    default: true,
    icon: "link",
  },
  {
    key: "include_info_node",
    id: "opt-info-node",
    title: "包含信息节点",
    desc: "保留订阅中的信息节点",
    default: true,
    icon: "info",
  },
  {
    key: "skip_tls_verify",
    id: "opt-skip-tls",
    title: "跳过 TLS 验证",
    desc: "忽略证书校验，仅在特殊情况下使用",
    default: false,
    icon: "shield",
    badge: "不推荐",
    badgeClass: "badge-red",
  },
  {
    key: "udp",
    id: "opt-udp",
    title: "启用 UDP",
    desc: "启用 UDP 转发支持",
    default: true,
    icon: "zap",
  },
  {
    key: "node_list",
    id: "opt-node-list",
    title: "节点列表",
    desc: "在配置末尾附加节点列表",
    default: false,
    icon: "list",
  },
  {
    key: "sort_nodes",
    id: "opt-sort-nodes",
    title: "节点二次排序",
    desc: "对节点顺序进行优化",
    default: false,
    icon: "sort",
  },
  {
    key: "filter_illegal",
    id: "opt-filter-illegal",
    title: "过滤非法节点",
    desc: "自动跳过不可用节点",
    default: true,
    icon: "filter",
    badge: "推荐",
    badgeClass: "badge-green",
  },
  {
    key: "insert_url",
    id: "opt-insert-url",
    title: "插入源链接",
    desc: "在配置中保留源地址信息",
    default: false,
    icon: "code",
  },
];

const OUTPUT_OPTION_GROUPS = [
  {
    title: "名称显示",
    desc: "默认保持上游节点原名，仅添加订阅源标识前缀。",
    items: ["source_prefix"],
  },
  {
    title: "连接与兼容",
    desc: "影响兼容性与连接行为的开关项。",
    items: ["udp", "skip_tls_verify"],
  },
  {
    title: "输出内容",
    desc: "决定最终导出的配置中包含哪些附加信息。",
    items: ["include_info_node", "node_list", "insert_url"],
  },
  {
    title: "节点处理",
    desc: "在生成前对节点进行过滤与排序。",
    items: ["sort_nodes", "filter_illegal"],
  },
];

const svgIcon = (body) =>
  `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">${body}</svg>`;

const ICONS = {
  settings: svgIcon(
    '<circle cx="12" cy="12" r="3"></circle><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 1 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 1 1-4 0v-.09a1.65 1.65 0 0 0-1-1.51 1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 1 1-2.83-2.83l.06-.06A1.65 1.65 0 0 0 4.6 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 1 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 1 1 2.83-2.83l.06.06A1.65 1.65 0 0 0 8.92 4.6H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 1 1 4 0v.09A1.65 1.65 0 0 0 15 4.6a1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 1 1 2.83 2.83l-.06.06A1.65 1.65 0 0 0 19.4 9c.36.58.56 1.25.6 1.93.04.68-.12 1.37-.6 2.07Z"></path>',
  ),
  sliders: svgIcon(
    '<line x1="4" y1="21" x2="4" y2="14"></line><line x1="4" y1="10" x2="4" y2="3"></line><line x1="12" y1="21" x2="12" y2="12"></line><line x1="12" y1="8" x2="12" y2="3"></line><line x1="20" y1="21" x2="20" y2="16"></line><line x1="20" y1="12" x2="20" y2="3"></line><line x1="2" y1="14" x2="6" y2="14"></line><line x1="10" y1="8" x2="14" y2="8"></line><line x1="18" y1="16" x2="22" y2="16"></line>',
  ),
  link: svgIcon(
    '<path d="M10 13a5 5 0 0 0 7.07 0l2.12-2.12a5 5 0 1 0-7.07-7.07L11 5"></path><path d="M14 11a5 5 0 0 0-7.07 0L4.81 13.12a5 5 0 1 0 7.07 7.07L13 19"></path>',
  ),
  target: svgIcon(
    '<circle cx="12" cy="12" r="8"></circle><circle cx="12" cy="12" r="3"></circle><line x1="12" y1="2" x2="12" y2="5"></line><line x1="12" y1="19" x2="12" y2="22"></line><line x1="2" y1="12" x2="5" y2="12"></line><line x1="19" y1="12" x2="22" y2="12"></line>',
  ),
  server: svgIcon(
    '<rect x="3" y="4" width="18" height="7" rx="2"></rect><rect x="3" y="13" width="18" height="7" rx="2"></rect><line x1="7" y1="8" x2="7.01" y2="8"></line><line x1="7" y1="17" x2="7.01" y2="17"></line><line x1="11" y1="8" x2="17" y2="8"></line><line x1="11" y1="17" x2="17" y2="17"></line>',
  ),
  file: svgIcon(
    '<path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"></path><polyline points="14 2 14 8 20 8"></polyline>',
  ),
  filter: svgIcon(
    '<polygon points="3 4 21 4 14 12 14 19 10 21 10 12 3 4"></polygon>',
  ),
  shield: svgIcon(
    '<path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10Z"></path>',
  ),
  zap: svgIcon(
    '<polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"></polygon>',
  ),
  list: svgIcon(
    '<line x1="8" y1="6" x2="21" y2="6"></line><line x1="8" y1="12" x2="21" y2="12"></line><line x1="8" y1="18" x2="21" y2="18"></line><line x1="3" y1="6" x2="3.01" y2="6"></line><line x1="3" y1="12" x2="3.01" y2="12"></line><line x1="3" y1="18" x2="3.01" y2="18"></line>',
  ),
  sort: svgIcon(
    '<path d="M11 5h10"></path><path d="M11 9h7"></path><path d="M11 13h4"></path><path d="M3 17V5"></path><path d="m6 8-3-3-3 3"></path><path d="m0 14 3 3 3-3"></path>',
  ),
  globe: svgIcon(
    '<circle cx="12" cy="12" r="10"></circle><path d="M2 12h20"></path><path d="M12 2a15 15 0 0 1 0 20"></path><path d="M12 2a15 15 0 0 0 0 20"></path>',
  ),
  code: svgIcon(
    '<polyline points="16 18 22 12 16 6"></polyline><polyline points="8 6 2 12 8 18"></polyline>',
  ),
  check: svgIcon('<path d="M20 6 9 17l-5-5"></path>'),
  copy: svgIcon(
    '<rect x="9" y="9" width="13" height="13" rx="2"></rect><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"></path>',
  ),
  external: svgIcon(
    '<path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"></path><polyline points="15 3 21 3 21 9"></polyline><line x1="10" y1="14" x2="21" y2="3"></line>',
  ),
  import: svgIcon(
    '<path d="M12 3v12"></path><path d="m7 10 5 5 5-5"></path><path d="M5 21h14"></path>',
  ),
  database: svgIcon(
    '<ellipse cx="12" cy="5" rx="8" ry="3"></ellipse><path d="M4 5v6c0 1.7 3.6 3 8 3s8-1.3 8-3V5"></path><path d="M4 11v6c0 1.7 3.6 3 8 3s8-1.3 8-3v-6"></path>',
  ),
  alert: svgIcon(
    '<path d="M10.29 3.86 1.82 18A2 2 0 0 0 3.53 21h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0Z"></path><line x1="12" y1="9" x2="12" y2="13"></line><line x1="12" y1="17" x2="12.01" y2="17"></line>',
  ),
  trash: svgIcon(
    '<path d="M3 6h18"></path><path d="M8 6V4a1 1 0 0 1 1-1h6a1 1 0 0 1 1 1v2"></path><path d="M19 6l-1 14a2 2 0 0 1-2 2H8a2 2 0 0 1-2-2L5 6"></path><path d="M10 11v6"></path><path d="M14 11v6"></path>',
  ),
  terminal: svgIcon(
    '<polyline points="4 17 10 11 4 5"></polyline><line x1="12" y1="19" x2="20" y2="19"></line>',
  ),
  nodes: svgIcon(
    '<circle cx="6" cy="12" r="2"></circle><circle cx="18" cy="6" r="2"></circle><circle cx="18" cy="18" r="2"></circle><path d="M8 12h8"></path><path d="M16.5 7.5 8 11"></path><path d="M16.5 16.5 8 13"></path>',
  ),
  rule: svgIcon(
    '<path d="M4 4h16v4H4z"></path><path d="M4 10h10v4H4z"></path><path d="M4 16h16v4H4z"></path>',
  ),
  template: svgIcon(
    '<path d="M3 7h18"></path><path d="M7 3v18"></path><rect x="3" y="3" width="18" height="18" rx="2"></rect>',
  ),
  spark: svgIcon(
    '<path d="m12 3 1.8 5.2L19 10l-5.2 1.8L12 17l-1.8-5.2L5 10l5.2-1.8L12 3Z"></path>',
  ),
  info: svgIcon(
    '<circle cx="12" cy="12" r="10"></circle><path d="M12 16v-4"></path><path d="M12 8h.01"></path>',
  ),
  search: svgIcon(
    '<circle cx="11" cy="11" r="7"></circle><path d="m21 21-4.35-4.35"></path>',
  ),
  lock: svgIcon(
    '<rect x="3" y="11" width="18" height="10" rx="2"></rect><path d="M7 11V7a5 5 0 0 1 10 0v4"></path>',
  ),
  home: svgIcon(
    '<path d="M3 10.5 12 3l9 7.5"></path><path d="M5 10v10h14V10"></path>',
  ),
  cloud: svgIcon(
    '<path d="M17.5 19a4.5 4.5 0 1 0-.7-8.94A6 6 0 1 0 6 18h11.5Z"></path>',
  ),
  book: svgIcon(
    '<path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20"></path><path d="M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2Z"></path>',
  ),
  wallet: svgIcon(
    '<path d="M20 7H4a2 2 0 0 0-2 2v9a2 2 0 0 0 2 2h16a2 2 0 0 0 2-2V9a2 2 0 0 0-2-2Z"></path><path d="M16 13h.01"></path><path d="M6 7V5a2 2 0 0 1 2-2h10"></path>',
  ),
  send: svgIcon(
    '<path d="m22 2-7 20-4-9-9-4Z"></path><path d="M22 2 11 13"></path>',
  ),
  play: svgIcon('<polygon points="5 3 19 12 5 21 5 3"></polygon>'),
  film: svgIcon(
    '<rect x="2" y="3" width="20" height="18" rx="2"></rect><path d="M7 3v18"></path><path d="M17 3v18"></path><path d="M2 8h20"></path><path d="M2 16h20"></path>',
  ),
  gamepad: svgIcon(
    '<path d="M6 12h4"></path><path d="M8 10v4"></path><path d="M15 13h.01"></path><path d="M18 11h.01"></path><path d="M6.8 20h10.4a3 3 0 0 0 2.88-3.84l-1.3-4.55A5 5 0 0 0 13.97 8H10.03a5 5 0 0 0-4.81 3.61l-1.3 4.55A3 3 0 0 0 6.8 20Z"></path>',
  ),
  refresh: svgIcon(
    '<path d="M21 12a9 9 0 1 1-2.64-6.36"></path><polyline points="21 3 21 9 15 9"></polyline>',
  ),
  save: svgIcon(
    '<path d="M19 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11l5 5v11a2 2 0 0 1-2 2Z"></path><polyline points="17 21 17 13 7 13 7 21"></polyline><polyline points="7 3 7 8 15 8"></polyline>',
  ),
  tag: svgIcon(
    '<path d="M20.59 13.41 11 3H4v7l9.59 9.59a2 2 0 0 0 2.82 0l4.18-4.18a2 2 0 0 0 0-2.82Z"></path><line x1="7" y1="7" x2="7.01" y2="7"></line>',
  ),
};

const RULE_ICON_MAP = {
  adblock: "shield",
  microsoft: "server",
  ai: "spark",
  apple: "target",
  bilibili: "film",
  social: "globe",
  youtube: "play",
  streaming: "film",
  google: "search",
  gaming: "gamepad",
  private: "lock",
  education: "book",
  domestic: "home",
  finance: "wallet",
  telegram: "send",
  cloud: "cloud",
  github: "code",
  non_cn: "globe",
};

function icon(name, className = "") {
  return `<span class="icon ${className}" aria-hidden="true">${ICONS[name] || ""}</span>`;
}

const DEFAULT_SERVICE = {
  enabled: true,
  listen_addr: "127.0.0.1",
  listen_port: 9876,
  log_level: "info",
  template: "standard",
  output_path: "/data/mihomo.yaml",
  cache_dir: "/data/cache",
  state_path: "/data/state.json",
  refresh_interval: 3600,
  refresh_on_request: true,
  stale_if_error: true,
  access_token: "",
  subscription_token: "",
  max_subscription_bytes: 5242880,
  fetch_timeout_seconds: 15,
  allow_lan: false,
};

const DEFAULT_RENDER = {
  mixed_port: 7890,
  allow_lan: false,
  mode: "rule",
  log_level: "info",
  ipv6: false,
  dns_enabled: true,
  enhanced_mode: "fake-ip",
  emoji: false,
  show_node_type: false,
  include_info_node: true,
  skip_tls_verify: false,
  udp: true,
  node_list: false,
  sort_nodes: false,
  filter_illegal: true,
  insert_url: false,
  source_prefix: true,
  name_options: {
    keep_raw_name: true,
    source_prefix_mode: "emoji_name",
    source_prefix_separator: "｜",
    dedupe_suffix_style: "#n",
  },
  include_keywords: "",
  exclude_keywords: "",
  output_filename: "mihomo.yaml",
  source_prefix_format: "{emoji} {name}",
  source_prefix_separator: "｜",
  dedupe_scope: "global",
  source_mode: "rules",
  template_rule_mode: "rules",
  external_config: {
    template_key: "none",
    template_label: "跟随当前服务模板",
    custom_url: "",
  },
  rule_mode: "full",
  enabled_rules: RULE_PRESETS.full,
  custom_rules: [],
  subscription_info: {
    enabled: true,
    expose_header: true,
    show_per_source: true,
    merge_strategy: "sum",
    expire_strategy: "earliest",
  },
};

const MASKED_SECRET = "********";
const YAML_PREVIEW_LINE_LIMIT = 500;
const DEFAULT_SOURCE_USER_AGENT = "clash.meta";
const LOCAL_DRAFT_STORAGE_KEY = "SUBCONV_LOCAL_DRAFT";
const SOURCE_EMOJI_OPTIONS = [
  "🦉",
  "🎶",
  "🤖",
  "🚀",
  "⚡",
  "🔥",
  "🐱",
  "🌐",
  "☁️",
  "🛡️",
  "🐶",
  "🐼",
];
const SOURCE_NAME_PREVIEW = "[anytls]JP Osaka Oracle";
const SOURCE_PREFIX_SEPARATOR = "｜";
const SOURCE_PREFIX_MODES = [
  { value: "emoji_name", label: "Emoji + 名称，推荐" },
  { value: "emoji", label: "仅 Emoji" },
  { value: "name", label: "仅名称" },
  { value: "none", label: "不显示" },
];
const CHINA_TIME_ZONE = "Asia/Shanghai";

const state = {
  config: null,
  activeWorkspace: "config",
  activeSourceMode: "rules",
  ruleMode: "full",
  enabledRules: new Set(RULE_PRESETS.full),
  customRules: [],
  externalConfig: { ...DEFAULT_RENDER.external_config },
  generatedUrl: "",
  workspaceId: "",
  workspaceExpiresAt: "",
  draftMode: "privacy",
  hasLocalDraft: false,
  localDraftMeta: null,
  localDraftPayload: null,
  generateStatus: "idle",
  refreshStatus: "idle",
  refreshStage: "",
  generateProgressTimer: null,
  lastError: "",
  publishNotice: "",
  resultNodeCount: 0,
  resultRuleCount: 0,
  yamlPreview: "",
  yamlSearch: "",
  yamlWrap: true,
  lastGeneratedAt: "",
  backendOnline: true,
  statusPayload: null,
  auditPayload: null,
  published: null,
  logsText: "",
  logsDisplay: "",
  subscriptionMeta: {
    aggregate: null,
    sources: [],
  },
  siteLogos: {},
  siteLogoRequests: {},
  subscriptions: [],
  inlineEntries: [],
  manualNodesEnabled: true,
  allNodes: [],
  filteredNodes: [],
  nodes: [],
  nodeSummary: { total: 0, enabled: 0, disabled: 0, modified: 0, warnings: 0 },
  nodeSourceOptions: [],
  selectedNodeIds: new Set(),
  activeNodeDeletePopoverId: "",
  confirmDialog: null,
  nodeFilters: {
    q: "",
    type: "all",
    region: "ALL",
    status: "all",
    source: "",
  },
  nodePagination: {
    page: 1,
    pageSize: 25,
    total: 0,
  },
  editingNode: null,
  nodeValidationWarnings: [],
  addNodeMode: "uri",
  ruleChipsExpanded: false,
  activeRuleSubtab: "builtin",
  activeNodeDialogTab: "basic",
  activeEmojiPopover: null,
};

document.addEventListener("DOMContentLoaded", () => {
  decorateStaticIcons();
  decorateButtons();
  renderOutputTiles();
  applyDefaultState();
  bindEvents();
  init();
});

function decorateStaticIcons() {
  document.querySelectorAll("[data-icon-name]").forEach((element) => {
    element.innerHTML = icon(element.dataset.iconName);
  });
}

function decorateButtons() {
  const mappings = [
    ["workspace-tab-config", "settings"],
    ["workspace-tab-nodes", "nodes"],
    ["workspace-tab-yaml", "code"],
    ["workspace-tab-diagnostics", "terminal"],
    ["add-subscription-btn", "link"],
    ["open-batch-import-btn", "import"],
    ["add-inline-btn", "nodes"],
    ["toggle-rule-chips-btn", "list"],
    ["open-node-editor-btn", "nodes"],
    ["generate-btn", "link"],
    ["import-btn", "import"],
    ["preview-btn", "code"],
    ["summary-refresh-btn", "refresh"],
    ["copy-generated-url-btn", "copy"],
    ["open-generated-url-btn", "external"],
    ["view-generated-link-btn", "link"],
    ["rotate-token-btn", "refresh"],
    ["delete-published-btn", "trash"],
    ["refresh-nodes-btn", "refresh"],
    ["refresh-yaml-btn", "refresh"],
    ["copy-yaml-btn", "copy"],
    ["refresh-diagnostics-btn", "refresh"],
    ["refresh-logs-btn", "refresh"],
    ["copy-logs-btn", "copy"],
    ["clear-logs-display-btn", "filter"],
    ["save-custom-rule-btn", "save"],
    ["reset-custom-rule-form-btn", "refresh"],
    ["save-node-btn", "save"],
  ];

  mappings.forEach(([id, iconName]) => {
    const button = document.getElementById(id);
    if (!button || button.dataset.iconized === "1") return;
    const text = button.textContent.trim();
    button.dataset.buttonIcon = iconName;
    button.innerHTML = `${icon(iconName)}<span>${escapeHtml(text)}</span>`;
    button.dataset.iconized = "1";
  });
}

function setButtonIconText(target, text) {
  const button =
    typeof target === "string" ? document.getElementById(target) : target;
  if (!button) return;
  const iconName = button.dataset.buttonIcon;
  if (!iconName) {
    button.textContent = text;
    return;
  }
  button.innerHTML = `${icon(iconName)}<span>${escapeHtml(text)}</span>`;
}

async function init() {
  setValue("backend-origin", window.location.origin);
  updateGeneratedUrlPlaceholder();
  renderResult(false);
  renderYamlViewer();
  renderDiagnostics();
  await initializeWorkspaceSession();
}

function bindEvents() {
  document.querySelectorAll("[data-workspace]").forEach((button) => {
    button.addEventListener("click", () =>
      switchWorkspace(button.dataset.workspace),
    );
  });

  document
    .getElementById("generate-btn")
    .addEventListener("click", generateSubscription);
  document.getElementById("import-btn").addEventListener("click", importClash);
  document
    .getElementById("preview-btn")
    .addEventListener("click", previewYAMLWorkspace);
  document
    .getElementById("summary-refresh-btn")
    .addEventListener("click", generateSubscription);
  document
    .getElementById("copy-generated-url-btn")
    .addEventListener("click", copyGeneratedUrl);
  document
    .getElementById("open-generated-url-btn")
    .addEventListener("click", () => openUrl(state.generatedUrl));
  document
    .getElementById("view-generated-link-btn")
    .addEventListener("click", focusGeneratedLink);
  document
    .getElementById("rotate-token-btn")
    .addEventListener("click", rotatePublishedToken);
  document
    .getElementById("delete-published-btn")
    .addEventListener("click", deletePublishedLink);

  document
    .getElementById("add-subscription-btn")
    .addEventListener("click", () => {
      state.subscriptions.push(createSubscriptionEntry());
      renderSubscriptionManager();
    });
  document
    .getElementById("open-batch-import-btn")
    .addEventListener("click", openBatchImportDialog);
  document
    .getElementById("apply-batch-import-btn")
    .addEventListener("click", applyBatchImport);
  document.getElementById("add-inline-btn").addEventListener("click", addInlineEntry);
  document
    .getElementById("manual-nodes-enabled")
    .addEventListener("change", () => {
      state.manualNodesEnabled = getChecked("manual-nodes-enabled");
      renderInlineManager();
      updateSubscriptionSummary();
    });
  document
    .getElementById("source-mode-rules")
    .addEventListener("click", () => switchSourceMode("rules"));
  document
    .getElementById("source-mode-template")
    .addEventListener("click", () => switchSourceMode("template"));
  document
    .getElementById("template-select")
    .addEventListener("change", (event) => selectTemplate(event.target.value));
  document
    .getElementById("rule-preset-control")
    .addEventListener("click", (event) => {
      const button = event.target.closest("[data-rule-mode]");
      if (!button) return;
      applyRulePreset(button.dataset.ruleMode);
    });
  document
    .getElementById("toggle-rule-chips-btn")
    .addEventListener("click", toggleRuleChipExpansion);
  document
    .getElementById("rule-subtab-builtin")
    .addEventListener("click", () => switchRuleSubtab("builtin"));
  document
    .getElementById("rule-subtab-custom")
    .addEventListener("click", () => switchRuleSubtab("custom"));
  document
    .getElementById("save-custom-rule-btn")
    .addEventListener("click", saveCustomRuleFromDialog);
  document
    .getElementById("reset-custom-rule-form-btn")
    .addEventListener("click", resetCustomRuleForm);
  document
    .getElementById("custom-rule-key")
    .addEventListener("input", renderCustomRulePreview);
  document
    .getElementById("custom-rule-label")
    .addEventListener("input", renderCustomRulePreview);
  document
    .getElementById("custom-rule-emoji")
    .addEventListener("input", renderCustomRulePreview);
  document
    .getElementById("custom-rule-target-mode")
    .addEventListener("change", () => {
      syncCustomRuleFormVisibility();
      renderCustomRulePreview();
    });
  document
    .getElementById("custom-rule-target-group")
    .addEventListener("input", renderCustomRulePreview);
  document
    .getElementById("custom-rule-source-type")
    .addEventListener("change", () => {
      syncCustomRuleFormVisibility();
      renderCustomRulePreview();
    });
  document
    .getElementById("custom-rule-behavior")
    .addEventListener("change", () => {
      syncCustomRuleFormVisibility();
      renderCustomRulePreview();
    });
  document
    .getElementById("custom-rule-format")
    .addEventListener("change", () => {
      syncCustomRuleFormVisibility();
      renderCustomRulePreview();
    });
  document
    .getElementById("custom-rule-url")
    .addEventListener("input", renderCustomRulePreview);
  document
    .getElementById("custom-rule-path")
    .addEventListener("input", renderCustomRulePreview);
  document
    .getElementById("custom-rule-interval")
    .addEventListener("input", renderCustomRulePreview);
  document
    .getElementById("custom-rule-insert-position")
    .addEventListener("change", renderCustomRulePreview);
  document
    .getElementById("custom-rule-enabled")
    .addEventListener("change", renderCustomRulePreview);
  document
    .getElementById("custom-rule-no-resolve")
    .addEventListener("change", renderCustomRulePreview);
  document
    .getElementById("custom-rule-payload")
    .addEventListener("input", renderCustomRulePreview);

  document
    .getElementById("open-node-editor-btn")
    .addEventListener("click", () => switchWorkspace("nodes"));
  document
    .getElementById("refresh-nodes-btn")
    .addEventListener("click", loadNodes);
  document
    .getElementById("node-type-filter")
    .addEventListener("change", handleNodeFilterChange);
  document
    .getElementById("node-region-filter")
    .addEventListener("change", handleNodeFilterChange);
  document
    .getElementById("node-status-filter")
    .addEventListener("change", handleNodeFilterChange);
  document
    .getElementById("node-source-filter")
    .addEventListener("change", handleNodeFilterChange);
  document
    .getElementById("node-page-size")
    .addEventListener("change", handleNodePageSizeChange);
  document
    .getElementById("node-prev-page-btn")
    .addEventListener("click", () => changeNodePage(-1));
  document
    .getElementById("node-next-page-btn")
    .addEventListener("click", () => changeNodePage(1));

  document
    .getElementById("refresh-yaml-btn")
    .addEventListener("click", loadYamlPreview);
  document.getElementById("copy-yaml-btn").addEventListener("click", copyYAML);
  document.getElementById("yaml-search").addEventListener("input", () => {
    state.yamlSearch = getValue("yaml-search").trim();
    renderYamlViewer();
  });
  document.getElementById("yaml-wrap-toggle").addEventListener("change", () => {
    state.yamlWrap = getChecked("yaml-wrap-toggle");
    renderYamlViewer();
  });

  document
    .getElementById("refresh-diagnostics-btn")
    .addEventListener("click", refreshDiagnostics);
  document
    .getElementById("refresh-logs-btn")
    .addEventListener("click", loadLogs);
  document.getElementById("copy-logs-btn").addEventListener("click", copyLogs);
  document
    .getElementById("clear-logs-display-btn")
    .addEventListener("click", clearLogsDisplay);

  document
    .getElementById("close-node-dialog-btn")
    .addEventListener("click", closeNodeDialog);
  document
    .getElementById("save-node-btn")
    .addEventListener("click", () => saveNodeOverride(false));
  document
    .getElementById("delete-node-btn")
    .addEventListener("click", handleDeleteCurrentNode);
  document
    .getElementById("close-danger-dialog-btn")
    .addEventListener("click", closeDangerDialog);
  document
    .getElementById("danger-dialog-cancel-btn")
    .addEventListener("click", closeDangerDialog);
  document
    .getElementById("danger-dialog-confirm-btn")
    .addEventListener("click", handleDangerDialogConfirm);
  document.getElementById("danger-dialog").addEventListener("close", () => {
    state.confirmDialog = null;
  });

  document.querySelectorAll("[data-close-dialog]").forEach((button) => {
    button.addEventListener("click", () =>
      closeDialog(button.dataset.closeDialog),
    );
  });
  document.querySelectorAll("[data-toggle-secret]").forEach((button) => {
    button.addEventListener("click", () =>
      toggleSecretField(button.dataset.toggleSecret, button),
    );
  });

  document.addEventListener("input", handleLiveFieldUpdates, true);
  document.addEventListener("change", handleLiveFieldUpdates, true);
  document.addEventListener("click", (event) => {
    document.querySelectorAll(".row-menu[open]").forEach((menu) => {
      if (menu.contains(event.target)) return;
      menu.removeAttribute("open");
    });
    if (
      state.activeEmojiPopover &&
      !event.target.closest("[data-emoji-popover]") &&
      !event.target.closest("[data-emoji-trigger]")
    ) {
      closeEmojiPopover();
    }
    if (
      state.activeNodeDeletePopoverId &&
      !event.target.closest("[data-popconfirm-root]") &&
      !event.target.closest("[data-node-delete-trigger]")
    ) {
      state.activeNodeDeletePopoverId = "";
      renderNodeTable();
    }
  });
  document.addEventListener("keydown", (event) => {
    if (event.key !== "Escape") return;
    if (state.activeEmojiPopover) {
      closeEmojiPopover();
      return;
    }
    if (document.getElementById("danger-dialog")?.open) {
      closeDangerDialog();
      return;
    }
    if (state.activeNodeDeletePopoverId) {
      state.activeNodeDeletePopoverId = "";
      renderNodeTable();
    }
  });
  window.addEventListener("resize", () => {
    if (!state.activeEmojiPopover) return;
    closeEmojiPopover();
  });
  window.addEventListener(
    "scroll",
    () => {
      if (!state.activeEmojiPopover) return;
      closeEmojiPopover();
    },
    true,
  );
}

function handleLiveFieldUpdates(event) {
  const target = event.target;
  if (!target) return;

  if (target.id === "backend-origin") {
    if (state.generateStatus === "success") {
      state.generatedUrl = buildSubscriptionURL();
      renderResult();
    }
    updateGeneratedUrlPlaceholder();
  }
  if (target.id === "opt-source-prefix") {
    const modeSelect = document.getElementById("source-prefix-mode");
    if (modeSelect) {
      if (!target.checked) modeSelect.value = "none";
      else if (modeSelect.value === "none")
        modeSelect.value = DEFAULT_RENDER.name_options.source_prefix_mode;
    }
    syncSourcePrefixModeControl();
    renderSubscriptionManager();
  }
  if (target.id === "source-prefix-mode") {
    const mode = normalizeSourcePrefixMode(target.value);
    setChecked("opt-source-prefix", mode !== "none");
    syncSourcePrefixModeControl();
    renderSubscriptionManager();
  }
  if (target.id === "node-search") {
    window.clearTimeout(handleLiveFieldUpdates.nodeSearchTimer);
    handleLiveFieldUpdates.nodeSearchTimer = window.setTimeout(() => {
      state.nodeFilters.q = getValue("node-search").trim();
      state.nodePagination.page = 1;
      applyNodeFiltersAndPagination();
    }, 220);
  }
  if (target.id === "manual-node-type") {
    setValue("manual-node-name", getValue("manual-node-name"));
  }
  updateSummary();
}

function switchWorkspace(workspace) {
  state.activeWorkspace = workspace;
  document.querySelectorAll("[data-workspace]").forEach((button) => {
    button.classList.toggle("active", button.dataset.workspace === workspace);
  });
  ["config", "nodes", "yaml", "diagnostics"].forEach((name) => {
    document
      .getElementById(`workspace-${name}`)
      .classList.toggle("hidden", name !== workspace);
  });

  if (workspace === "yaml") {
    loadYamlPreview();
  } else if (workspace === "nodes") {
    if (!state.allNodes.length) loadNodes();
  } else if (workspace === "diagnostics") {
    refreshDiagnostics();
  }
}

function applyDefaultState() {
  setValue("client-type", "mihomo");
  setValue("output-filename", DEFAULT_RENDER.output_filename);
  setValue("include-keywords", "");
  setValue("exclude-keywords", "");
  setChecked("opt-source-prefix", DEFAULT_RENDER.source_prefix);
  setValue(
    "source-prefix-mode",
    DEFAULT_RENDER.name_options.source_prefix_mode,
  );
  syncSourcePrefixModeControl();
  setValue("template-select", DEFAULT_RENDER.external_config.template_key);
  setValue("custom-template-url", "");
  setValue("batch-import-textarea", "");
  setChecked("yaml-wrap-toggle", true);
  setValue("yaml-search", "");
  setValue("node-type-filter", "all");
  setValue("node-region-filter", "ALL");
  setValue("node-status-filter", "all");
  setValue("node-source-filter", "");
  setValue("node-page-size", "25");
  state.activeWorkspace = "config";
  state.activeSourceMode = "rules";
  state.ruleMode = "full";
  state.enabledRules = new Set(RULE_PRESETS.full);
  state.customRules = [];
  state.externalConfig = { ...DEFAULT_RENDER.external_config };
  state.generatedUrl = "";
  state.workspaceId = "";
  state.workspaceExpiresAt = "";
  state.draftMode = "privacy";
  state.hasLocalDraft = false;
  state.localDraftMeta = null;
  state.localDraftPayload = null;
  state.generateStatus = "idle";
  state.refreshStatus = "idle";
  state.lastError = "";
  state.publishNotice = "";
  state.lastGeneratedAt = "";
  state.resultNodeCount = 0;
  state.resultRuleCount = 0;
  state.published = null;
  state.ruleChipsExpanded = false;
  state.activeRuleSubtab = "builtin";
  state.subscriptions = [
    createSubscriptionEntry({
      name: "source-1",
      enabled: true,
      user_agent: DEFAULT_SOURCE_USER_AGENT,
    }),
  ];
  state.siteLogos = {};
  state.siteLogoRequests = {};
  state.inlineEntries = [];
  state.manualNodesEnabled = true;
  state.allNodes = [];
  state.filteredNodes = [];
  state.subscriptionMeta = { aggregate: null, sources: [] };
  state.nodeFilters = {
    q: "",
    type: "all",
    region: "ALL",
    status: "all",
    source: "",
  };
  state.nodePagination = { page: 1, pageSize: 25, total: 0 };
  state.selectedNodeIds = new Set();
  state.activeNodeDeletePopoverId = "";
  state.confirmDialog = null;
  renderTemplateOptions();
  renderRulePresetControl();
  renderRuleChips();
  renderSourceMode();
  renderTemplateDetail();
  renderSubscriptionManager();
  renderInlineManager();
  renderResult();
  renderSubscriptionMeta();
  renderNodeEditor();
  renderConfigNodeSummary();
  renderDiagnostics();
  renderSessionBanner();
  switchWorkspace("config");
}

async function initializeWorkspaceSession() {
  await createNewWorkspace({ preserveDraftState: false });
  const storedDraft = loadLocalDraft();
  if (storedDraft) {
    applyDraftState("draft_detected", storedDraft);
  } else {
    applyDraftState("privacy", null);
  }
  renderSessionBanner();
  await reloadWorkspaceData();
}

async function createNewWorkspace(options = {}) {
  const { preserveDraftState = true } = options;
  const draftSnapshot = preserveDraftState ? snapshotDraftState() : null;
  const response = await fetchJSON("/api/workspaces", { method: "POST" });
  if (!response?.ok) {
    showToast(readAPIError(response) || "创建隐私会话失败。", true);
    applyDefaultState();
    return false;
  }
  applyDefaultState();
  state.workspaceId = response.workspace_id || "";
  state.workspaceExpiresAt = response.expires_at || "";
  if (draftSnapshot) {
    restoreDraftState(draftSnapshot);
  }
  renderSessionBanner();
  return true;
}

function renderSessionBanner() {
  const banner = document.getElementById("session-banner");
  if (!banner) return;

  const mode = state.draftMode || "privacy";
  const meta = state.localDraftMeta;
  let title = "当前模式：隐私会话";
  let description = "刷新页面后不会自动保留配置。";
  let metaLine = "默认不会自动保存或自动恢复配置。";
  let buttons = `
    <button id="save-local-draft-btn" class="ghost-button small-button" type="button">保存为本机草稿</button>
    <button id="clear-session-btn" class="danger-ghost-button small-button" type="button">清空当前会话</button>
  `;

  if (mode === "draft_detected") {
    title = "发现本机草稿";
    description = "这是之前手动保存到本机浏览器的配置，是否恢复？";
    metaLine = formatDraftMetaLine(meta);
    buttons = `
      <button id="restore-draft-btn" class="secondary-button small-button" type="button">恢复草稿</button>
      <button id="discard-draft-btn" class="ghost-button small-button" type="button">丢弃草稿</button>
    `;
  } else if (mode === "local_draft") {
    title = "当前模式：本机草稿";
    description = "此配置已保存在当前浏览器中，请勿在公共设备上使用。";
    metaLine = formatDraftMetaLine(meta);
    buttons = `
      <button id="update-local-draft-btn" class="secondary-button small-button" type="button">更新本机草稿</button>
      <button id="exit-draft-mode-btn" class="ghost-button small-button" type="button">退出草稿模式</button>
      <button id="clear-session-btn" class="danger-ghost-button small-button" type="button">清空当前会话</button>
    `;
  }

  banner.className = `panel side-card session-banner mode-${mode.replace(/_/g, "-")}`;
  banner.innerHTML = `
    <div class="session-banner-main">
      <strong class="session-banner-title">${escapeHtml(title)}</strong>
      <span class="session-banner-text">${escapeHtml(description)}</span>
      <span class="session-banner-meta">${escapeHtml(metaLine)}</span>
    </div>
    <div class="session-banner-actions">
      ${buttons}
    </div>
  `;

  document
    .getElementById("save-local-draft-btn")
    ?.addEventListener("click", saveLocalDraft);
  document
    .getElementById("update-local-draft-btn")
    ?.addEventListener("click", updateLocalDraft);
  document
    .getElementById("clear-session-btn")
    ?.addEventListener("click", clearCurrentSession);
  document
    .getElementById("restore-draft-btn")
    ?.addEventListener("click", restoreLocalDraft);
  document
    .getElementById("discard-draft-btn")
    ?.addEventListener("click", discardLocalDraft);
  document
    .getElementById("exit-draft-mode-btn")
    ?.addEventListener("click", exitDraftMode);
}

function formatDraftMetaLine(meta) {
  if (!meta) {
    return "本机草稿会保存订阅地址，请仅在私人设备上使用。";
  }
  const savedAt = formatDraftDateTime(meta.saved_at);
  const updatedAt = formatDraftDateTime(meta.updated_at);
  const sourceCount = Number.parseInt(meta.source_count, 10) || 0;
  if (savedAt && updatedAt && savedAt !== updatedAt) {
    return `首次保存 ${savedAt} · 最近更新 ${updatedAt} · 来源 ${sourceCount} 个`;
  }
  if (updatedAt) {
    return `最近更新 ${updatedAt} · 来源 ${sourceCount} 个`;
  }
  return `本机草稿会保存订阅地址，请仅在私人设备上使用。`;
}

function formatDraftDateTime(value) {
  return formatChinaDateTime(value, { seconds: false, fallback: "" });
}

function formatChinaDateTime(value, options = {}) {
  const fallback = options.fallback ?? "-";
  if (!value) return fallback;
  const date = value instanceof Date ? value : new Date(value);
  if (Number.isNaN(date.getTime())) return fallback;
  const parts = new Intl.DateTimeFormat("zh-CN", {
    timeZone: CHINA_TIME_ZONE,
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
    second: options.seconds === false ? undefined : "2-digit",
    hour12: false,
    hourCycle: "h23",
  })
    .formatToParts(date)
    .reduce((acc, part) => {
      if (part.type !== "literal") acc[part.type] = part.value;
      return acc;
    }, {});
  const datePart = `${parts.year}-${parts.month}-${parts.day}`;
  const timePart = `${parts.hour}:${parts.minute}`;
  return options.seconds === false
    ? `${datePart} ${timePart}`
    : `${datePart} ${timePart}:${parts.second}`;
}

function snapshotDraftState() {
  return {
    draftMode: state.draftMode,
    hasLocalDraft: state.hasLocalDraft,
    localDraftMeta: state.localDraftMeta
      ? deepClone(state.localDraftMeta)
      : null,
    localDraftPayload: state.localDraftPayload
      ? deepClone(state.localDraftPayload)
      : null,
  };
}

function restoreDraftState(snapshot) {
  if (!snapshot) return;
  state.draftMode = snapshot.draftMode || "privacy";
  state.hasLocalDraft = Boolean(snapshot.hasLocalDraft);
  state.localDraftMeta = snapshot.localDraftMeta
    ? deepClone(snapshot.localDraftMeta)
    : null;
  state.localDraftPayload = snapshot.localDraftPayload
    ? deepClone(snapshot.localDraftPayload)
    : null;
}

function applyDraftState(mode, payload) {
  state.draftMode = mode;
  if (payload) {
    state.hasLocalDraft = true;
    state.localDraftPayload = deepClone(payload);
    state.localDraftMeta = {
      saved_at: payload.saved_at || "",
      updated_at: payload.updated_at || payload.saved_at || "",
      source_count: Number(
        payload.source_count || countDraftSources(payload.config) || 0,
      ),
    };
    return;
  }
  state.hasLocalDraft = false;
  state.localDraftPayload = null;
  state.localDraftMeta = null;
}

function draftPublishRefFromPublished(published) {
  const publishID = String(published?.publish_id || "").trim();
  if (!publishID) return null;
  return {
    publish_id: publishID,
    token_hint: String(published?.token_hint || "").trim(),
    updated_at: String(published?.updated_at || "").trim(),
  };
}

function normalizePublishedStatus(response) {
  if (!response?.ok || !response.publish_id) return null;
  const url = response.subscription_url || response.url || "";
  return {
    ...response,
    url,
    subscription_url: url,
  };
}

function applyPublishedStatus(response) {
  const published = normalizePublishedStatus(response);
  if (!published) {
    state.published = null;
    state.generatedUrl = "";
    if (state.generateStatus === "success") {
      state.generateStatus = "idle";
    }
    renderResult();
    return false;
  }
  state.publishNotice = "";
  state.published = published;
  state.generatedUrl = published.url || "";
  state.generateStatus = state.generatedUrl ? "success" : state.generateStatus;
  if (published.updated_at) {
    state.lastGeneratedAt = formatChinaDateTime(published.updated_at, {
      fallback: published.updated_at,
    });
  }
  renderResult();
  return true;
}

async function restoreLocalDraft() {
  const draft = loadLocalDraft();
  if (!draft?.config) {
    applyDraftState("privacy", null);
    renderSessionBanner();
    showToast("未找到可恢复的本机草稿。", true);
    return;
  }
  await deleteCurrentWorkspace();
  const created = await createNewWorkspace({ preserveDraftState: false });
  if (!created) return;
  const restoredConfig = sanitizeLocalDraftConfig(draft.config);
  const backendConfig = configForBackend(restoredConfig);
  const restoredNodeState = sanitizeLocalDraftNodeState(draft.node_state);
  const restoreResult = await restoreDraftToWorkspace(
    backendConfig,
    draft.publish_ref,
    restoredNodeState,
  );
  if (!restoreResult?.ok) return;
  state.config = backendConfig;
  fillFormFromConfig(restoredConfig);
  applyDraftState("local_draft", draft);
  renderSessionBanner();
  const restoredPublished = applyRestoreDraftPublish(restoreResult.publish);
  const missingPublished = Boolean(
    draft.publish_ref?.publish_id && !restoreResult.publish?.exists,
  );
  if (missingPublished) {
    state.publishNotice = "原订阅发布已不存在，需要重新生成订阅链接。";
    renderResult();
  }
  await Promise.all([loadStatus(), loadSubscriptionMeta(), loadAudit()]);
  if (!restoredPublished && !missingPublished) {
    await loadPublishedStatus();
  }
  if (restoredNodeState) {
    void loadNodes();
  }
  updateSummary();
  renderDiagnostics();
  showToast(
    missingPublished
      ? "已恢复本机草稿。原订阅发布已不存在，需要重新生成订阅链接。"
      : "已恢复本机草稿。",
    missingPublished,
  );
}

async function restoreDraftToWorkspace(config, publishRef, nodeState) {
  if (!state.workspaceId) return null;
  const body = {
    config: configForBackend(config),
    publish_ref: {
      publish_id: String(publishRef?.publish_id || "").trim(),
    },
  };
  if (nodeState) {
    body.node_state = nodeState;
  }
  const response = await fetchJSON(
    `/api/workspaces/${encodeURIComponent(state.workspaceId)}/restore-draft`,
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body),
    },
  );
  if (!response?.ok) {
    showToast(readAPIError(response) || "恢复本机草稿失败。", true);
    return null;
  }
  return response;
}

function applyRestoreDraftPublish(publish) {
  if (!publish?.exists) {
    state.published = null;
    state.generatedUrl = "";
    state.generateStatus = "idle";
    state.lastGeneratedAt = "";
    return false;
  }
  return applyPublishedStatus({
    ok: true,
    publish_id: publish.publish_id,
    url: publish.url || publish.subscription_url || "",
    subscription_url: publish.subscription_url || publish.url || "",
    token_hint: publish.token_hint || "",
    created_at: publish.created_at || "",
    updated_at: publish.updated_at || "",
    last_access_at: publish.last_access_at || "",
    access_count: publish.access_count || 0,
    status: publish.status || "active",
  });
}

async function discardLocalDraft() {
  clearLocalDraft();
  await deleteCurrentWorkspace();
  const created = await createNewWorkspace({ preserveDraftState: false });
  if (!created) return;
  applyDraftState("privacy", null);
  renderSessionBanner();
  await reloadWorkspaceData();
  showToast("已丢弃本机草稿。");
}

function saveLocalDraft() {
  void persistLocalDraft(false);
}

function updateLocalDraft() {
  void persistLocalDraft(true);
}

async function persistLocalDraft(isUpdate, options = {}) {
  if (!state.workspaceId) {
    showToast("当前没有可保存的会话。", true);
    return;
  }
  const existing = loadLocalDraft();
  const now = new Date().toISOString();
  const config = sanitizeLocalDraftConfig(buildConfigFromForm());
  config.client_type = getValue("client-type") || "mihomo";
  config.backend_url = normalizeBackendOrigin(
    getValue("backend-origin") || window.location.origin,
  );
  const nodeState = await loadNodeStateDraft();
  const payload = {
    version: 2,
    saved_at: existing?.saved_at || now,
    updated_at: now,
    source_count: countDraftSources(config),
    config,
  };
  if (nodeState) {
    payload.node_state = nodeState;
  }
  const latestPublished = await syncPublishedStatusForDraft();
  const publishRef = draftPublishRefFromPublished(latestPublished);
  if (publishRef) {
    payload.publish_ref = publishRef;
  }
  localStorage.setItem(LOCAL_DRAFT_STORAGE_KEY, JSON.stringify(payload));
  applyDraftState("local_draft", payload);
  renderSessionBanner();
  if (!options.silent) {
    showToast(
      isUpdate
        ? "本机草稿已更新。"
        : "已保存为本机草稿。草稿会保存当前配置，但不会保存完整私密订阅 token。",
    );
  }
}

async function exitDraftMode() {
  let keepDraft = true;
  if (state.hasLocalDraft) {
    const shouldDeleteDraft = window.confirm(
      "是否删除本机草稿？\n选择“确定”将删除，选择“取消”则保留。",
    );
    if (shouldDeleteDraft) {
      clearLocalDraft();
      keepDraft = false;
    }
  }
  if (keepDraft) {
    const payload = loadLocalDraft();
    applyDraftState("privacy", payload);
  } else {
    applyDraftState("privacy", null);
  }
  renderSessionBanner();
  showToast(
    keepDraft
      ? "已退出草稿模式，本机草稿已保留。"
      : "已退出草稿模式，并删除本机草稿。",
  );
}

async function clearCurrentSession() {
  const currentMode = state.draftMode;
  let nextMode =
    currentMode === "draft_detected" ? "draft_detected" : "privacy";
  let nextDraftPayload = loadLocalDraft();
  if (currentMode === "local_draft" && state.hasLocalDraft) {
    const shouldDeleteDraft = window.confirm(
      "是否同时删除本机草稿？\n选择“确定”将删除草稿，选择“取消”则仅清空当前会话。",
    );
    if (shouldDeleteDraft) {
      clearLocalDraft();
      nextDraftPayload = null;
      nextMode = "privacy";
    } else {
      nextMode = nextDraftPayload ? "draft_detected" : "privacy";
    }
  }
  await deleteCurrentWorkspace();
  sessionStorage.clear();
  const created = await createNewWorkspace({ preserveDraftState: false });
  if (!created) return;
  applyDraftState(nextMode, nextDraftPayload);
  renderSessionBanner();
  await reloadWorkspaceData();
  showToast("当前会话已清空。");
}

async function deleteCurrentWorkspace() {
  if (!state.workspaceId) return;
  await fetchJSON(`/api/workspaces/${encodeURIComponent(state.workspaceId)}`, {
    method: "DELETE",
  });
}

async function reloadWorkspaceData() {
  await Promise.all([
    loadConfig(),
    loadStatus(),
    loadSubscriptionMeta(),
    loadAudit(),
    loadPublishedStatus(),
  ]);
  updateSummary();
  renderDiagnostics();
}

function loadLocalDraft() {
  const raw = localStorage.getItem(LOCAL_DRAFT_STORAGE_KEY);
  if (!raw) return null;
  try {
    const parsed = JSON.parse(raw);
    if (
      !parsed ||
      typeof parsed !== "object" ||
      !parsed.config ||
      typeof parsed.config !== "object"
    ) {
      throw new Error("invalid local draft");
    }
    return parsed;
  } catch (error) {
    localStorage.removeItem(LOCAL_DRAFT_STORAGE_KEY);
    return null;
  }
}

function clearLocalDraft() {
  localStorage.removeItem(LOCAL_DRAFT_STORAGE_KEY);
}

function removeLocalDraftPublishRef(publishID) {
  const draft = loadLocalDraft();
  if (!draft?.publish_ref) return;
  if (
    publishID &&
    String(draft.publish_ref.publish_id || "").trim() !== String(publishID).trim()
  ) {
    return;
  }
  delete draft.publish_ref;
  draft.updated_at = new Date().toISOString();
  localStorage.setItem(LOCAL_DRAFT_STORAGE_KEY, JSON.stringify(draft));
  if (state.localDraftPayload) {
    delete state.localDraftPayload.publish_ref;
  }
}

function sanitizeLocalDraftConfig(config) {
  const draftConfig = deepClone(config || {});
  draftConfig.service = {
    ...(draftConfig.service || {}),
    access_token: "",
    subscription_token: "",
  };
  return draftConfig;
}

function configForBackend(config) {
  const backendConfig = sanitizeLocalDraftConfig(config);
  const service = backendConfig.service || {};
  const outputPath = String(service.output_path || "").trim();
  const statePath = String(service.state_path || "").trim();
  const cacheDir = String(service.cache_dir || "").trim();
  backendConfig.service = {
    ...service,
    output_path: outputPath.startsWith("/") ? outputPath : DEFAULT_SERVICE.output_path,
    state_path: statePath.startsWith("/") ? statePath : DEFAULT_SERVICE.state_path,
    cache_dir: cacheDir.startsWith("/") ? cacheDir : DEFAULT_SERVICE.cache_dir,
    listen_addr: String(service.listen_addr || DEFAULT_SERVICE.listen_addr),
    listen_port: Number.parseInt(service.listen_port, 10) || DEFAULT_SERVICE.listen_port,
    refresh_interval:
      Number.parseInt(service.refresh_interval, 10) ||
      DEFAULT_SERVICE.refresh_interval,
    max_subscription_bytes:
      Number.parseInt(service.max_subscription_bytes, 10) ||
      DEFAULT_SERVICE.max_subscription_bytes,
    fetch_timeout_seconds:
      Number.parseInt(service.fetch_timeout_seconds, 10) ||
      DEFAULT_SERVICE.fetch_timeout_seconds,
  };
  delete backendConfig.client_type;
  delete backendConfig.backend_url;
  delete backendConfig.publish_ref;
  delete backendConfig.node_state;
  return backendConfig;
}

function sanitizeLocalDraftNodeState(nodeState) {
  if (!nodeState || typeof nodeState !== "object") return null;
  const sanitized = {
    node_overrides:
      nodeState.node_overrides && typeof nodeState.node_overrides === "object"
        ? deepClone(nodeState.node_overrides)
        : {},
    disabled_nodes: Array.isArray(nodeState.disabled_nodes)
      ? [...nodeState.disabled_nodes]
      : [],
    deleted_nodes: Array.isArray(nodeState.deleted_nodes)
      ? [...nodeState.deleted_nodes]
      : [],
    custom_nodes: Array.isArray(nodeState.custom_nodes)
      ? deepClone(nodeState.custom_nodes)
      : [],
  };
  return sanitized;
}

async function loadNodeStateDraft() {
  const response = await fetchJSON("/api/nodes/state");
  if (!response?.ok) return null;
  return sanitizeLocalDraftNodeState(response.state);
}

async function saveNodeStateDraft(nodeState) {
  const sanitized = sanitizeLocalDraftNodeState(nodeState);
  if (!sanitized) return true;
  const response = await fetchJSON("/api/nodes/state", {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ state: sanitized }),
  });
  if (!response?.ok) {
    showToast(readAPIError(response) || "恢复节点编辑记录失败。", true);
    return false;
  }
  return true;
}

function countDraftSources(config) {
  const subscriptions = Array.isArray(config?.subscriptions)
    ? config.subscriptions.filter((item) => String(item?.url || "").trim())
    : [];
  const inlineEntries = Array.isArray(config?.inline)
    ? config.inline.filter((item) => String(item?.content || "").trim())
    : [];
  return subscriptions.length + inlineEntries.length;
}

async function loadConfig() {
  const response = await fetchJSON("/api/config");
  if (!response?.ok) {
    showToast(readAPIError(response) || "加载配置失败。", true);
    const draftSnapshot = snapshotDraftState();
    const workspaceId = state.workspaceId;
    const workspaceExpiresAt = state.workspaceExpiresAt;
    applyDefaultState();
    state.workspaceId = workspaceId;
    state.workspaceExpiresAt = workspaceExpiresAt;
    restoreDraftState(draftSnapshot);
    renderSessionBanner();
    return;
  }

  state.config = response.config;
  fillFormFromConfig(response.config);
}

function fillFormFromConfig(config) {
  const render = config?.render || {};
  setValue("client-type", config?.client_type || "mihomo");
  setValue("backend-origin", config?.backend_url || window.location.origin);
  setValue(
    "output-filename",
    render.output_filename || DEFAULT_RENDER.output_filename,
  );
  setValue("include-keywords", render.include_keywords || "");
  setValue("exclude-keywords", render.exclude_keywords || "");

  for (const option of OUTPUT_OPTIONS) {
    setChecked(option.id, render[option.key] ?? option.default);
  }
  const prefixMode = normalizeSourcePrefixMode(
    render.name_options?.source_prefix_mode ||
      (render.source_prefix === false
        ? "none"
        : DEFAULT_RENDER.name_options.source_prefix_mode),
  );
  setValue("source-prefix-mode", prefixMode);
  setChecked(
    "opt-source-prefix",
    prefixMode !== "none" && render.source_prefix !== false,
  );
  syncSourcePrefixModeControl();

  state.activeSourceMode =
    (render.source_mode || render.template_rule_mode) === "template"
      ? "template"
      : "rules";
  state.ruleMode = render.rule_mode || "full";
  state.enabledRules = new Set(
    Array.isArray(render.enabled_rules) && render.enabled_rules.length
      ? render.enabled_rules
      : RULE_PRESETS.full,
  );
  state.customRules = Array.isArray(render.custom_rules)
    ? render.custom_rules
    : [];
  state.externalConfig = {
    ...DEFAULT_RENDER.external_config,
    ...(render.external_config || {}),
  };
  state.subscriptions = normalizeSubscriptionEntries(
    config?.subscriptions || [],
  );
  state.inlineEntries = normalizeInlineEntries(config?.inline || []);
  state.manualNodesEnabled = config?.manual_nodes_enabled !== false;
  setChecked("manual-nodes-enabled", state.manualNodesEnabled);
  if (!state.subscriptions.length) {
    state.subscriptions = [
      createSubscriptionEntry({
        name: "source-1",
        enabled: true,
        user_agent: DEFAULT_SOURCE_USER_AGENT,
      }),
    ];
  }

  setValue(
    "template-select",
    state.externalConfig.template_key ||
      DEFAULT_RENDER.external_config.template_key,
  );
  setValue("custom-template-url", state.externalConfig.custom_url || "");
  renderRulePresetControl();
  renderRuleChips();
  renderRuleSubtab();
  renderTemplateOptions();
  renderTemplateDetail();
  renderSourceMode();
  updateGeneratedUrlPlaceholder();
  renderResult();
  renderSubscriptionManager();
  renderInlineManager();
  renderSubscriptionMeta();
}

function renderOutputTiles() {
  const grid = document.getElementById("options-grid");
  grid.innerHTML = "";

  const optionMap = Object.fromEntries(
    OUTPUT_OPTIONS.map((item) => [item.key, item]),
  );

  OUTPUT_OPTION_GROUPS.forEach((group) => {
    const card = document.createElement("section");
    card.className = "option-group-card";
    card.innerHTML = `
      <div class="option-group-head">
        <div class="option-group-title">${escapeHtml(group.title)}</div>
        <div class="option-group-desc">${escapeHtml(group.desc)}</div>
      </div>
      <div class="option-group-list">
        ${group.items
          .map((key) => optionMap[key])
          .filter(Boolean)
          .map(
            (item) => `
            <label class="option-item" title="${escapeHtml(item.desc)}">
              <div class="option-icon">${icon(item.icon)}</div>
              <div class="setting-body">
                <div class="option-title-row">
                  <div class="option-title">${escapeHtml(item.title)}</div>
                  ${item.badge ? `<span class="badge ${item.badgeClass || ""}">${escapeHtml(item.badge)}</span>` : ""}
                </div>
                <div class="option-desc">${escapeHtml(item.desc)}</div>
              </div>
              <span class="switch">
                <input id="${item.id}" type="checkbox" ${item.default ? "checked" : ""} />
                <span class="switch-track"></span>
              </span>
            </label>
          `,
          )
          .join("")}
      </div>
      ${group.items.includes("source_prefix") ? renderSourcePrefixModeControl() : ""}
    `;
    grid.appendChild(card);
  });
}

function renderSourcePrefixModeControl() {
  return `
    <div class="source-prefix-mode-control">
      <label for="source-prefix-mode">订阅源标识样式</label>
      <select id="source-prefix-mode">
        ${SOURCE_PREFIX_MODES.map((mode) => `<option value="${escapeHtml(mode.value)}">${escapeHtml(mode.label)}</option>`).join("")}
      </select>
    </div>
  `;
}

function syncSourcePrefixModeControl() {
  const modeSelect = document.getElementById("source-prefix-mode");
  const prefixToggle = document.getElementById("opt-source-prefix");
  if (!modeSelect || !prefixToggle) return;
  modeSelect.disabled = !prefixToggle.checked;
}

function renderRulePresetControl() {
  document
    .querySelectorAll("#rule-preset-control [data-rule-mode]")
    .forEach((button) => {
      button.classList.toggle(
        "active",
        button.dataset.ruleMode === state.ruleMode,
      );
    });
}

function renderRuleChips() {
  const primary = document.getElementById("rule-chip-grid");
  const extraWrap = document.getElementById("rule-chip-extra-wrap");
  const extra = document.getElementById("rule-chip-grid-extra");
  const toggleButton = document.getElementById("toggle-rule-chips-btn");
  primary.innerHTML = "";
  extra.innerHTML = "";

  const rules = [...BUILTIN_RULES, ...state.customRules];
  const visibleRules = rules.slice(0, 12);
  const extraRules = rules.slice(12);

  visibleRules.forEach((rule) => primary.appendChild(createRuleChip(rule)));
  extraRules.forEach((rule) => extra.appendChild(createRuleChip(rule)));

  extraWrap.classList.toggle(
    "hidden",
    !extraRules.length || !state.ruleChipsExpanded,
  );
  toggleButton.classList.toggle("hidden", !extraRules.length);
  setButtonIconText(
    toggleButton,
    state.ruleChipsExpanded ? "收起" : "展开全部",
  );
}

function switchRuleSubtab(tab) {
  state.activeRuleSubtab = tab;
  renderRuleSubtab();
}

function renderRuleSubtab() {
  const isBuiltin = state.activeRuleSubtab !== "custom";
  document
    .getElementById("rule-subtab-builtin")
    .classList.toggle("active", isBuiltin);
  document
    .getElementById("rule-subtab-custom")
    .classList.toggle("active", !isBuiltin);
  document
    .getElementById("rule-builtin-panel")
    .classList.toggle("hidden", !isBuiltin);
  document
    .getElementById("rule-custom-panel")
    .classList.toggle("hidden", isBuiltin);
  if (!isBuiltin) {
    renderCustomRuleList();
    renderCustomRulePreview();
  }
}

function createRuleChip(rule) {
  const button = document.createElement("button");
  button.type = "button";
  button.className = `rule-chip ${state.enabledRules.has(rule.key) ? "selected" : ""}`;
  const iconName = RULE_ICON_MAP[rule.key] || "rule";
  button.innerHTML = `
    <span class="rule-chip-main">
      ${icon(iconName, "rule-icon")}
      <span class="rule-chip-label">${escapeHtml(rule.label)}</span>
    </span>
    <span class="rule-chip-check">${state.enabledRules.has(rule.key) ? icon("check", "chip-check-icon") : ""}</span>
  `;
  button.addEventListener("click", () => {
    toggleRule(rule.key);
  });
  return button;
}

function toggleRuleChipExpansion() {
  state.ruleChipsExpanded = !state.ruleChipsExpanded;
  renderRuleChips();
}

function toggleRule(ruleKey) {
  if (state.activeSourceMode !== "rules") {
    switchSourceMode("rules");
  }
  if (state.enabledRules.has(ruleKey)) {
    state.enabledRules.delete(ruleKey);
  } else {
    state.enabledRules.add(ruleKey);
  }
  state.ruleMode = "custom";
  renderRulePresetControl();
  renderRuleChips();
  updateSummary();
}

function applyRulePreset(mode) {
  state.ruleMode = mode;
  if (mode === "minimal") {
    state.enabledRules = new Set(RULE_PRESETS.minimal);
  } else if (mode === "balanced") {
    state.enabledRules = new Set(RULE_PRESETS.balanced);
  } else if (mode === "full") {
    state.enabledRules = new Set(RULE_PRESETS.full);
  }
  renderRulePresetControl();
  renderRuleChips();
  updateSummary();
}

function switchSourceMode(mode) {
  const nextMode = mode === "template" ? "template" : "rules";
  if (state.activeSourceMode === nextMode) return;
  state.activeSourceMode = nextMode;
  renderSourceMode();
  renderRuleChips();
  renderTemplateDetail();
  updateSummary();
}

function renderSourceMode() {
  const isRules = state.activeSourceMode === "rules";
  const rulesTab = document.getElementById("source-mode-rules");
  const templateTab = document.getElementById("source-mode-template");
  rulesTab.classList.toggle("active", isRules);
  rulesTab.setAttribute("aria-selected", String(isRules));
  templateTab.classList.toggle("active", !isRules);
  templateTab.setAttribute("aria-selected", String(!isRules));
  document
    .getElementById("rules-mode-panel")
    .classList.toggle("hidden", !isRules);
  document
    .getElementById("template-mode-panel")
    .classList.toggle("hidden", isRules);
}

function renderTemplateOptions() {
  const select = document.getElementById("template-select");
  select.innerHTML = TEMPLATE_OPTION_GROUPS.map(
    (group) => `
    <optgroup label="${escapeHtml(group.label)}">
      ${group.options
        .map((item) => {
          const selected =
            state.externalConfig.template_key === item.key ? " selected" : "";
          return `<option value="${escapeHtml(item.key)}"${selected}>${escapeHtml(templateOptionText(item))}</option>`;
        })
        .join("")}
    </optgroup>
  `,
  ).join("");
}

function templateOptionText(item) {
  const brief = (item.ruleProfile || item.description || "").trim();
  return brief ? `${item.label}｜${brief}` : item.label;
}

function currentTemplateOption() {
  return (
    TEMPLATE_OPTIONS.find(
      (item) => item.key === state.externalConfig.template_key,
    ) || TEMPLATE_OPTIONS[0]
  );
}

function selectTemplate(templateKey) {
  const selected =
    TEMPLATE_OPTIONS.find((item) => item.key === templateKey) ||
    TEMPLATE_OPTIONS[0];
  state.externalConfig.template_key = selected.key;
  state.externalConfig.template_label = selected.label;
  if (selected.key !== "custom_url") {
    state.externalConfig.custom_url = "";
    setValue("custom-template-url", "");
  }
  setValue("template-select", selected.key);
  renderTemplateOptions();
  renderTemplateDetail();
  updateSummary();
}

function renderTemplateDetail() {
  const selected = currentTemplateOption();
  const serviceTemplate = (
    (state.config && state.config.service && state.config.service.template) ||
    DEFAULT_SERVICE.template ||
    "standard"
  ).toLowerCase();
  const note =
    state.activeSourceMode === "template"
      ? selected.key === "custom_url"
        ? "模板模式已启用。当前自定义 URL 仅保留兼容字段，渲染时会回退到内置模板。"
        : "模板模式已启用，规则分类选择将被停用。"
      : "切换到模板模式后，该模板将接管规则来源。";
  document.getElementById("template-detail-card").innerHTML = `
    <div class="template-detail-title">${escapeHtml(selected.label)}</div>
    <div class="template-detail-desc">${escapeHtml(selected.description)}</div>
    <div class="template-detail-meta">
      <div class="template-detail-pill">
        <span class="template-detail-pill-label">模板族</span>
        <strong>${escapeHtml(selected.group)}</strong>
      </div>
      <div class="template-detail-pill">
        <span class="template-detail-pill-label">规则范围</span>
        <strong>${escapeHtml(selected.ruleProfile || "-")}</strong>
      </div>
      <div class="template-detail-pill">
        <span class="template-detail-pill-label">代理组模式</span>
        <strong>${escapeHtml(selected.groupProfile || "-")}</strong>
      </div>
    </div>
    <div class="template-detail-note">${escapeHtml(note)}</div>
    <div class="template-detail-note">${selected.key === "none" ? `当前将跟随 service.template=${escapeHtml(serviceTemplate)}。` : ""}</div>
  `;
  document
    .getElementById("custom-template-url-wrap")
    .classList.toggle("hidden", selected.key !== "custom_url");
}

function addCustomRule() {
  openCustomRuleDialog();
}

function openCustomRuleDialog() {
  switchRuleSubtab("custom");
  renderCustomRuleList();
  resetCustomRuleForm();
}

function renderCustomRuleList() {
  const container = document.getElementById("custom-rule-list");
  if (!state.customRules.length) {
    container.innerHTML = `<div class="empty-state compact-empty-state">还没有自定义规则。你可以创建内联规则、远程规则集或仅创建代理组。</div>`;
    return;
  }

  container.innerHTML = state.customRules
    .map(
      (rule, index) => `
      <div class="custom-rule-item">
        <div class="custom-rule-item-main">
          <div class="custom-rule-item-title">${escapeHtml(rule.icon || rule.emoji || "")} ${escapeHtml(rule.label || rule.key)}</div>
          <div class="meta-text">${escapeHtml(rule.key)}</div>
          <div class="meta-text">${escapeHtml(rule.source_type || "inline")} · ${escapeHtml(rule.behavior || "domain")} · ${escapeHtml(resolveCustomRulePreviewTarget(rule))}</div>
        </div>
        <div class="custom-rule-item-actions">
          <label class="checkbox-line compact-check"><input type="checkbox" data-custom-rule-toggle="${index}" ${rule.enabled !== false ? "checked" : ""} /> 启用</label>
          <button class="tiny-button primary-text" type="button" data-custom-rule-action="edit" data-custom-rule-index="${index}">编辑</button>
          <button class="tiny-button warn-text" type="button" data-custom-rule-action="delete" data-custom-rule-index="${index}">删除</button>
        </div>
      </div>
    `,
    )
    .join("");

  container
    .querySelectorAll("[data-custom-rule-toggle]")
    .forEach((checkbox) => {
      checkbox.addEventListener("change", () => {
        const index = Number.parseInt(checkbox.dataset.customRuleToggle, 10);
        state.customRules[index].enabled = checkbox.checked;
        state.ruleMode = "custom";
        updateSummary();
      });
    });
  container.querySelectorAll("[data-custom-rule-action]").forEach((button) => {
    button.addEventListener("click", () => {
      const index = Number.parseInt(button.dataset.customRuleIndex, 10);
      if (button.dataset.customRuleAction === "edit") {
        fillCustomRuleForm(index);
      } else if (button.dataset.customRuleAction === "delete") {
        deleteCustomRule(index);
      }
    });
  });
}

function fillCustomRuleForm(index) {
  const rule = state.customRules[index];
  if (!rule) return;
  setValue("custom-rule-edit-index", index);
  setValue("custom-rule-key", rule.key || "");
  setValue("custom-rule-label", rule.label || "");
  setValue("custom-rule-emoji", rule.icon || rule.emoji || "");
  setValue("custom-rule-target-mode", rule.target_mode || "new_group");
  setValue("custom-rule-target-group", rule.target_group || "");
  setValue("custom-rule-source-type", rule.source_type || "inline");
  setValue("custom-rule-behavior", rule.behavior || "domain");
  setValue("custom-rule-format", rule.format || "text");
  setValue("custom-rule-url", rule.url || "");
  setValue("custom-rule-path", rule.path || "");
  setValue("custom-rule-interval", rule.interval || 86400);
  setValue(
    "custom-rule-insert-position",
    rule.insert_position || "before_match",
  );
  setChecked("custom-rule-enabled", rule.enabled !== false);
  setChecked("custom-rule-no-resolve", Boolean(rule.no_resolve));
  setValue(
    "custom-rule-payload",
    Array.isArray(rule.payload) ? rule.payload.join("\n") : "",
  );
  syncCustomRuleFormVisibility();
  renderCustomRulePreview();
}

function resetCustomRuleForm() {
  setValue("custom-rule-edit-index", "");
  setValue("custom-rule-key", "");
  setValue("custom-rule-label", "");
  setValue("custom-rule-emoji", "");
  setValue("custom-rule-target-mode", "new_group");
  setValue("custom-rule-target-group", "");
  setValue("custom-rule-source-type", "inline");
  setValue("custom-rule-behavior", "domain");
  setValue("custom-rule-format", "text");
  setValue("custom-rule-url", "");
  setValue("custom-rule-path", "");
  setValue("custom-rule-interval", "86400");
  setValue("custom-rule-insert-position", "before_match");
  setChecked("custom-rule-enabled", true);
  setChecked("custom-rule-no-resolve", false);
  setValue("custom-rule-payload", "");
  syncCustomRuleFormVisibility();
  renderCustomRulePreview();
}

function renderCustomRulePreview() {
  const draft = buildCustomRuleDraft();
  const preview = document.getElementById("custom-rule-preview");
  if (!draft.key && !draft.label && !draft.icon) {
    preview.textContent = "未填写";
    return;
  }
  const missing = [];
  if (!draft.key) missing.push("缺少规则 Key");
  if (!draft.label) missing.push("缺少显示名称");
  if (draft.source_type === "http" && !draft.url)
    missing.push("缺少远程规则 URL");
  if (draft.source_type === "file" && !draft.path)
    missing.push("缺少本地文件路径");
  if (draft.source_type === "inline" && !draft.payload.length)
    missing.push("缺少内联规则内容");
  if (draft.format === "mrs" && draft.behavior === "classical")
    missing.push("mrs 不能搭配 classical");

  if (missing.length) {
    preview.textContent = missing.join("\n");
    return;
  }

  const lines = [];
  const targetGroup = resolveCustomRulePreviewTarget(draft);
  if (draft.target_mode === "new_group" || draft.source_type === "group_only") {
    lines.push("proxy-groups:");
    lines.push(
      `  - {name: "${targetGroup}", type: select, proxies: ["🚀 节点选择", "⚡ 自动选择", DIRECT, REJECT]}`,
    );
  }
  if (draft.source_type !== "group_only") {
    lines.push("");
    lines.push("rule-providers:");
    if (draft.source_type === "inline") {
      lines.push(
        `  ${draft.key}: {type: inline, behavior: ${draft.behavior}, format: ${draft.format}, payload: [${draft.payload.map((item) => `"${item}"`).join(", ")}]}`,
      );
    } else if (draft.source_type === "http") {
      lines.push(
        `  ${draft.key}: {type: http, behavior: ${draft.behavior}, url: "${draft.url}", path: ${draft.path || "./ruleset/" + draft.key + "." + previewFormatExt(draft.format)}, interval: ${draft.interval || 86400}, format: ${draft.format}}`,
      );
    } else if (draft.source_type === "file") {
      lines.push(
        `  ${draft.key}: {type: file, behavior: ${draft.behavior}, path: ${draft.path}, format: ${draft.format}}`,
      );
    }
    lines.push("");
    lines.push("rules:");
    lines.push(
      `  - RULE-SET,${draft.key},${targetGroup}${draft.no_resolve ? ",no-resolve" : ""}`,
    );
  }
  preview.textContent = lines.join("\n");
}

function saveCustomRuleFromDialog() {
  const indexValue = getValue("custom-rule-edit-index");
  const draft = buildCustomRuleDraft();

  if (!draft.key) {
    showToast("规则 key 不能为空。", true);
    return;
  }
  if (!draft.label) {
    showToast("显示名称不能为空。", true);
    return;
  }
  if (draft.source_type === "http" && !draft.url) {
    showToast("远程规则 URL 不能为空。", true);
    return;
  }
  if (draft.source_type === "file" && !draft.path) {
    showToast("本地文件路径不能为空。", true);
    return;
  }
  if (draft.source_type === "inline" && !draft.payload.length) {
    showToast("内联规则内容至少要有一条。", true);
    return;
  }
  if (draft.format === "mrs" && draft.behavior === "classical") {
    showToast("mrs 不能搭配 classical。", true);
    return;
  }

  const duplicate = [...BUILTIN_RULES, ...state.customRules].some(
    (item, index) => {
      if (indexValue !== "" && index === Number.parseInt(indexValue, 10))
        return false;
      return item.key === draft.key;
    },
  );
  if (duplicate) {
    showToast("规则 key 已存在。", true);
    return;
  }

  const rule = draft;
  if (indexValue === "") {
    state.customRules.push(rule);
  } else {
    state.customRules[Number.parseInt(indexValue, 10)] = rule;
  }

  state.enabledRules.add(draft.key);
  state.ruleMode = "custom";
  renderRulePresetControl();
  renderRuleChips();
  renderCustomRuleList();
  resetCustomRuleForm();
  updateSummary();
  showToast("自定义规则已保存。");
}

function deleteCustomRule(index) {
  const rule = state.customRules[index];
  if (!rule) return;
  state.customRules.splice(index, 1);
  state.enabledRules.delete(rule.key);
  state.ruleMode = "custom";
  renderRulePresetControl();
  renderRuleChips();
  renderCustomRuleList();
  resetCustomRuleForm();
  updateSummary();
}

function syncCustomRuleFormVisibility() {
  const sourceType = getValue("custom-rule-source-type");
  const targetMode = getValue("custom-rule-target-mode");
  document
    .getElementById("custom-rule-target-group-wrap")
    .classList.toggle("hidden", targetMode !== "existing_group");
  document
    .getElementById("custom-rule-url-wrap")
    .classList.toggle("hidden", sourceType !== "http");
  document
    .getElementById("custom-rule-path-wrap")
    .classList.toggle("hidden", sourceType !== "file" && sourceType !== "http");
  document
    .getElementById("custom-rule-interval-wrap")
    .classList.toggle("hidden", sourceType !== "http");
  document
    .getElementById("custom-rule-payload-wrap")
    .classList.toggle("hidden", sourceType !== "inline");
}

function buildCustomRuleDraft() {
  return {
    key: getValue("custom-rule-key").trim().toLowerCase(),
    label: getValue("custom-rule-label").trim(),
    icon: getValue("custom-rule-emoji").trim(),
    enabled: getChecked("custom-rule-enabled"),
    target_mode: getValue("custom-rule-target-mode"),
    target_group: getValue("custom-rule-target-group").trim(),
    source_type: getValue("custom-rule-source-type"),
    behavior: getValue("custom-rule-behavior"),
    format: getValue("custom-rule-format"),
    url: getValue("custom-rule-url").trim(),
    path: getValue("custom-rule-path").trim(),
    interval: Number.parseInt(getValue("custom-rule-interval"), 10) || 86400,
    payload: getValue("custom-rule-payload")
      .split("\n")
      .map((item) => item.trim())
      .filter(Boolean),
    insert_position: getValue("custom-rule-insert-position"),
    no_resolve: getChecked("custom-rule-no-resolve"),
  };
}

function resolveCustomRulePreviewTarget(rule) {
  if (rule.target_mode === "direct") return "DIRECT";
  if (rule.target_mode === "reject") return "REJECT";
  if (rule.target_mode === "existing_group")
    return rule.target_group || "节点选择";
  return `${rule.icon ? `${rule.icon} ` : ""}${rule.label || "未命名规则"}`.trim();
}

function previewFormatExt(format) {
  if (format === "mrs") return "mrs";
  if (format === "text") return "txt";
  return "yaml";
}

function sourceDomainFromUrl(rawUrl) {
  try {
    const parsed = new URL(String(rawUrl || "").trim());
    return parsed.hostname || "";
  } catch {
    return "";
  }
}

function sourceLogoKey(item) {
  const domain = sourceDomainFromUrl(item?.url || "");
  return item?.id || domain || item?.name || "";
}

function hashString(value) {
  let hash = 0;
  const text = String(value || "");
  for (let i = 0; i < text.length; i += 1) {
    hash = (hash * 31 + text.charCodeAt(i)) >>> 0;
  }
  return hash;
}

function fallbackAvatarStyle(seed) {
  const hue = hashString(seed) % 360;
  return `background:hsl(${hue} 70% 96%);color:hsl(${hue} 55% 38%);border-color:hsl(${hue} 45% 84%)`;
}

function fallbackAvatarLabel(item) {
  const source = (
    item?.name ||
    item?.source_name ||
    item?.domain ||
    "?"
  ).trim();
  return source ? source.slice(0, 1).toUpperCase() : "?";
}

function resolveLogoPayload(item) {
  const key = sourceLogoKey(item);
  const cached = state.siteLogos[key];
  if (item?.source_logo) {
    return { kind: "image", src: item.source_logo };
  }
  if (cached?.logoUrl) {
    return { kind: "image", src: cached.logoUrl };
  }
  return {
    kind: "fallback",
    label: fallbackAvatarLabel(item),
    style: fallbackAvatarStyle(key || fallbackAvatarLabel(item)),
  };
}

function renderSourceLogo(item, className = "source-logo") {
  const payload = resolveLogoPayload(item);
  if (payload.kind === "image") {
    return `<span class="${className}"><img src="${escapeHtml(payload.src)}" alt="" loading="lazy" /></span>`;
  }
  return `<span class="${className} source-logo-fallback" style="${payload.style}">${escapeHtml(payload.label)}</span>`;
}

async function ensureSubscriptionLogos() {
  const tasks = state.subscriptions
    .filter((item) => item.url && !item.source_logo)
    .map(async (item) => {
      const domain = sourceDomainFromUrl(item.url);
      const key = sourceLogoKey(item);
      if (!domain || state.siteLogos[key] || state.siteLogoRequests[key])
        return;
      state.siteLogoRequests[key] = true;
      try {
        const response = await fetchJSON(
          `/api/site-logo?url=${encodeURIComponent(item.url)}`,
        );
        if (response?.ok) {
          state.siteLogos[key] = {
            logoUrl: response.logoUrl || "",
            domain: response.domain || domain,
            source: response.source || "fallback",
          };
        } else {
          state.siteLogos[key] = { logoUrl: "", domain, source: "fallback" };
        }
      } finally {
        delete state.siteLogoRequests[key];
        renderSubscriptionManager();
        renderSubscriptionMeta();
        renderNodeEditor();
      }
    });
  await Promise.all(tasks);
}

function createSubscriptionEntry(values = {}) {
  const index = state.subscriptions.length + 1;
  return {
    id: values.id || `source-${index}`,
    name: values.name || `source-${index}`,
    emoji:
      values.emoji !== undefined
        ? normalizeSourceEmoji(values.emoji)
        : defaultSourceEmoji(index - 1),
    source_logo: values.source_logo || "",
    enabled: values.enabled !== false,
    url: values.url || "",
    user_agent: values.user_agent || DEFAULT_SOURCE_USER_AGENT,
  };
}

function defaultSourceEmoji(index) {
  return SOURCE_EMOJI_OPTIONS[index % SOURCE_EMOJI_OPTIONS.length] || "";
}

function normalizeSourceEmoji(value) {
  const text = String(value || "").trim();
  if (!text) return "";
  if (typeof Intl !== "undefined" && Intl.Segmenter) {
    const segments = new Intl.Segmenter(undefined, {
      granularity: "grapheme",
    }).segment(text);
    const first = segments[Symbol.iterator]().next();
    return first.done ? "" : first.value.segment;
  }
  return Array.from(text).slice(0, 2).join("");
}

function normalizeEmojiInput(value) {
  return Array.from(String(value || "").trim())
    .slice(0, 2)
    .join("");
}

function normalizeSourcePrefixMode(value) {
  const mode = String(value || "").trim();
  return SOURCE_PREFIX_MODES.some((item) => item.value === mode)
    ? mode
    : DEFAULT_RENDER.name_options.source_prefix_mode;
}

function currentSourcePrefixMode() {
  if (!getChecked("opt-source-prefix")) return "none";
  return normalizeSourcePrefixMode(
    getValue("source-prefix-mode") ||
      DEFAULT_RENDER.name_options.source_prefix_mode,
  );
}

function buildSourcePrefix(item, mode = currentSourcePrefixMode()) {
  const emoji = normalizeSourceEmoji(item?.emoji);
  const sourceName = String(item?.name || "").trim();
  if (mode === "none") return "";
  if (mode === "emoji") return emoji;
  if (mode === "name") return sourceName;
  if (emoji && sourceName) return `${emoji} ${sourceName}`;
  return emoji || sourceName;
}

function sourceEmojiDuplicateWarning(item, index) {
  const emoji = normalizeSourceEmoji(item.emoji);
  if (!emoji) return "";
  const duplicated = state.subscriptions.some(
    (other, otherIndex) =>
      otherIndex !== index && normalizeSourceEmoji(other.emoji) === emoji,
  );
  return duplicated ? "该 Emoji 已被其他订阅源使用，可能不易区分。" : "";
}

function sourceNamePreview(item) {
  const mode = currentSourcePrefixMode();
  const prefix = buildSourcePrefix(item, mode);
  if (!prefix) return SOURCE_NAME_PREVIEW;
  if (mode === "emoji") return `${prefix} ${SOURCE_NAME_PREVIEW}`;
  return `${prefix}${SOURCE_PREFIX_SEPARATOR}${SOURCE_NAME_PREVIEW}`;
}

function renderSubscriptionRow(item, index) {
  const warning = sourceEmojiDuplicateWarning(item, index);
  return `
    <div class="source-row subscription-item">
      <div class="source-row-main">
        <div class="source-enable source-toggle">
          <span class="source-enable-label">启用</span>
          <label class="switch compact-switch">
            <input type="checkbox" data-sub-field="enabled" data-sub-index="${index}" ${item.enabled ? "checked" : ""} />
            <span class="switch-track"></span>
          </label>
        </div>
        <div class="source-name-combo${state.activeEmojiPopover?.index === index ? " is-emoji-open" : ""}">
          <input class="source-name source-name-inline source-name-input" type="text" data-sub-field="name" data-sub-index="${index}" placeholder="名称，例如 主力机场" value="${escapeHtml(item.name)}" />
          <button class="source-emoji-suffix" type="button" data-emoji-trigger data-sub-index="${index}" title="订阅源标识，用于生成节点名前缀" aria-haspopup="menu" aria-expanded="${state.activeEmojiPopover?.index === index ? "true" : "false"}" aria-label="选择订阅源标识">${escapeHtml(normalizeSourceEmoji(item.emoji) || "＋")}</button>
        </div>
        <input class="source-url source-url-input" type="text" data-sub-field="url" data-sub-index="${index}" placeholder="https://example.com/sub?token=xxx" value="${escapeHtml(item.url)}" />
        <button class="tiny-button danger-ghost source-delete-btn" type="button" data-sub-action="delete" data-sub-index="${index}">删除</button>
      </div>
      <div class="source-row-meta">
        <span class="source-domain">${escapeHtml(sourceDomainFromUrl(item.url) || "未识别域名")}</span>
        <span class="source-meta-separator">·</span>
        <span class="source-preview" data-sub-index="${index}" title="${escapeHtml(sourceNamePreview(item))}">${escapeHtml(sourceNamePreview(item))}</span>
        <span class="source-meta-separator">·</span>
        <span class="source-meta-link">User-Agent</span>
        ${warning ? `<span class="source-meta-separator">·</span><span class="source-warning-inline">${escapeHtml(warning)}</span>` : ""}
      </div>
      <details class="source-advanced inline-advanced">
        <summary>User-Agent</summary>
        <input type="text" data-sub-field="user_agent" data-sub-index="${index}" value="${escapeHtml(item.user_agent || DEFAULT_SOURCE_USER_AGENT)}" />
      </details>
    </div>
  `;
}

function createInlineEntry(values = {}) {
  const index = state.inlineEntries.length + 1;
  return {
    id: values.id || `manual-${index}`,
    name: values.name || `手动节点 ${index}`,
    enabled: values.enabled !== false,
    content: values.content || "",
    status: values.status || "idle",
    result: values.result || "",
    error: values.error || "",
    parsed: values.parsed || null,
  };
}

function normalizeSubscriptionEntries(items) {
  return (Array.isArray(items) ? items : []).map((item, index) =>
    createSubscriptionEntry({
      id: item.id || `source-${index + 1}`,
      name: item.name || `source-${index + 1}`,
      emoji: item.emoji !== undefined ? item.emoji : defaultSourceEmoji(index),
      source_logo: item.source_logo || "",
      enabled: item.enabled !== false,
      url: item.url || "",
      user_agent: item.user_agent || DEFAULT_SOURCE_USER_AGENT,
    }),
  );
}

function normalizeInlineEntries(items) {
  return (Array.isArray(items) ? items : []).map((item, index) =>
    createInlineEntry({
      id: item.id || `manual-${index + 1}`,
      name: item.name || `手动节点 ${index + 1}`,
      enabled: item.enabled !== false,
      content: item.content || "",
      status: item.status || "idle",
      result: item.result || "",
      error: item.error || "",
      parsed: item.parsed || null,
    }),
  );
}

function renderSubscriptionManager() {
  const container = document.getElementById("subscription-list");
  if (!state.subscriptions.length) {
    state.subscriptions.push(createSubscriptionEntry());
  }
  container.innerHTML = state.subscriptions
    .map((item, index) => renderSubscriptionRow(item, index))
    .join("");

  container.querySelectorAll("[data-sub-field]").forEach((element) => {
    const index = Number.parseInt(element.dataset.subIndex, 10);
    const field = element.dataset.subField;
    const isCheckbox = element.type === "checkbox";
    const eventName = isCheckbox ? "change" : "input";
    element.addEventListener(eventName, () => {
      state.subscriptions[index][field] = isCheckbox
        ? element.checked
        : element.value;
      updateSubscriptionSummary();
      renderSubscriptionMeta();
      renderNodeEditor();
      if (field === "url") {
        renderSubscriptionManager();
      } else if (field === "name") {
        updateSourceRowPreview(index);
      }
    });
  });
  container.querySelectorAll("[data-emoji-trigger]").forEach((button) => {
    button.addEventListener("click", () => {
      const index = Number.parseInt(button.dataset.subIndex, 10);
      if (state.activeEmojiPopover?.index === index) {
        closeEmojiPopover();
        return;
      }
      openEmojiPopover(index, button);
    });
  });
  container.querySelectorAll("[data-sub-action='delete']").forEach((button) => {
    button.addEventListener("click", () => {
      const index = Number.parseInt(button.dataset.subIndex, 10);
      state.subscriptions.splice(index, 1);
      if (state.activeEmojiPopover?.index === index) {
        closeEmojiPopover();
      }
      if (!state.subscriptions.length) {
        state.subscriptions.push(createSubscriptionEntry());
      }
      renderSubscriptionManager();
    });
  });
  updateSubscriptionSummary();
}

function updateSourceRowPreview(index) {
  const item = state.subscriptions[index];
  if (!item) return;
  const row = document.querySelector(
    `.source-preview[data-sub-index="${index}"]`,
  );
  if (!row) return;
  const preview = sourceNamePreview(item);
  row.textContent = preview;
  row.title = preview;
}

function renderEmojiPopoverContent(index) {
  const currentEmoji = normalizeSourceEmoji(state.subscriptions[index]?.emoji);
  return `
	    <div class="emoji-popover-title">选择标识</div>
	    <div class="emoji-grid">
	      ${SOURCE_EMOJI_OPTIONS.map(
          (emoji) => `
	          <button
	            class="emoji-option${currentEmoji === emoji ? " active" : ""}"
	            type="button"
	            role="menuitem"
	            data-emoji-option="${escapeHtml(emoji)}"
	            data-sub-index="${index}"
	            aria-label="选择 ${escapeHtml(emoji)}"
	          >${escapeHtml(emoji)}</button>
	        `,
        ).join("")}
	    </div>
	    <div class="emoji-divider"></div>
	    <div class="custom-emoji-row">
      <input
        type="text"
        maxlength="8"
        value="${escapeHtml(currentEmoji)}"
        placeholder="自定义"
        data-emoji-custom-input
        data-sub-index="${index}"
      />
      <button type="button" class="secondary-button" data-emoji-custom-apply data-sub-index="${index}">使用</button>
    </div>
  `;
}

function placeEmojiPopover(trigger, popover) {
  const rect = trigger.getBoundingClientRect();
  const popW = 196;
  const popH = popover.offsetHeight || 174;
  const gap = 6;

  let left = rect.right - popW;
  let top = rect.bottom + gap;

  if (left < 12) left = 12;
  if (left + popW > window.innerWidth - 12) {
    left = window.innerWidth - popW - 12;
  }

  if (top + popH > window.innerHeight - 12) {
    top = rect.top - popH - gap;
  }

  if (top < 12) top = 12;

  popover.style.left = `${left}px`;
  popover.style.top = `${top}px`;
}

function applySubscriptionEmoji(index, emoji) {
  const normalized = normalizeEmojiInput(emoji);
  if (!normalized || !state.subscriptions[index]) return;
  state.subscriptions[index].emoji = normalized;
  updateSubscriptionSummary();
  renderSubscriptionMeta();
  renderNodeEditor();
  closeEmojiPopover();
  renderSubscriptionManager();
}

function bindEmojiPopoverEvents(popover, index) {
  popover.querySelectorAll("[data-emoji-option]").forEach((button) => {
    button.addEventListener("click", () =>
      applySubscriptionEmoji(index, button.dataset.emojiOption),
    );
  });

  const input = popover.querySelector("[data-emoji-custom-input]");
  const applyButton = popover.querySelector("[data-emoji-custom-apply]");
  const applyCustomEmoji = () => {
    if (!input) return;
    applySubscriptionEmoji(index, input.value);
  };

  applyButton?.addEventListener("click", applyCustomEmoji);
  input?.addEventListener("keydown", (event) => {
    if (event.key === "Enter") {
      event.preventDefault();
      applyCustomEmoji();
    }
  });
}

function closeEmojiPopover() {
  const popover = document.getElementById("emoji-popover");
  const trigger = state.activeEmojiPopover?.trigger;
  if (popover) popover.remove();
  if (trigger) trigger.setAttribute("aria-expanded", "false");
  trigger?.closest(".source-name-combo")?.classList.remove("is-emoji-open");
  state.activeEmojiPopover = null;
}

function openEmojiPopover(index, trigger) {
  closeEmojiPopover();
  const popover = document.createElement("div");
  popover.id = "emoji-popover";
  popover.className = "emoji-popover";
  popover.setAttribute("data-emoji-popover", "true");
  popover.setAttribute("role", "menu");
  popover.innerHTML = renderEmojiPopoverContent(index);
  document.body.appendChild(popover);
  bindEmojiPopoverEvents(popover, index);
  trigger.setAttribute("aria-expanded", "true");
  trigger.closest(".source-name-combo")?.classList.add("is-emoji-open");
  placeEmojiPopover(trigger, popover);
  state.activeEmojiPopover = { index, trigger };
}

function renderInlineManager() {
  const container = document.getElementById("inline-list");
  const disabled = !state.manualNodesEnabled;
  document
    .querySelector(".manual-section")
    ?.classList.toggle("is-disabled", disabled);
  document
    .getElementById("add-inline-btn")
    ?.classList.toggle("hidden", state.inlineEntries.length === 0);
  document
    .getElementById("manual-disabled-note")
    ?.classList.toggle("hidden", state.manualNodesEnabled);
  container.classList.toggle("is-disabled", disabled);

  if (state.inlineEntries.length === 0) {
    container.innerHTML = renderManualEmptyState();
    container
      .querySelector("[data-inline-action='add']")
      ?.addEventListener("click", addInlineEntry);
    updateSubscriptionSummary();
    return;
  }

  container.innerHTML = `
    <div class="manual-node-list">
      ${state.inlineEntries
        .map((item, index) => renderInlineCard(item, index, disabled))
        .join("")}
    </div>
  `;

  container.querySelectorAll("[data-inline-field]").forEach((element) => {
    const index = Number.parseInt(element.dataset.inlineIndex, 10);
    const field = element.dataset.inlineField;
    const eventName = element.tagName === "TEXTAREA" ? "input" : "input";
    element.addEventListener(eventName, () => {
      state.inlineEntries[index][field] = element.value;
      if (field === "content") {
        resetInlineParseState(index);
      }
      updateSubscriptionSummary();
    });
  });
  container
    .querySelectorAll("[data-inline-action='preview']")
    .forEach((button) => {
      button.addEventListener("click", () => {
        const index = Number.parseInt(button.dataset.inlineIndex, 10);
        void previewInlineEntry(index);
      });
    });
  container
    .querySelectorAll("[data-inline-action='clear']")
    .forEach((button) => {
      button.addEventListener("click", () => {
        const index = Number.parseInt(button.dataset.inlineIndex, 10);
        state.inlineEntries[index].content = "";
        resetInlineParseState(index);
        renderInlineManager();
        updateSubscriptionSummary();
      });
    });
  container
    .querySelectorAll("[data-inline-action='delete']")
    .forEach((button) => {
      button.addEventListener("click", () => {
        const index = Number.parseInt(button.dataset.inlineIndex, 10);
        requestDeleteInlineEntry(index);
      });
    });
  updateSubscriptionSummary();
}

function addInlineEntry() {
  state.inlineEntries.push(createInlineEntry());
  renderInlineManager();
  const index = state.inlineEntries.length - 1;
  window.requestAnimationFrame(() => {
    const card = document.querySelector(`[data-manual-card-index="${index}"]`);
    card?.scrollIntoView({ behavior: "smooth", block: "nearest" });
    card?.querySelector("[data-inline-field='name']")?.focus();
  });
}

function renderManualEmptyState() {
  return `
    <div class="manual-empty-state">
      <div class="manual-empty-icon">${icon("terminal")}</div>
      <div class="manual-empty-title">暂无手动节点</div>
      <div class="manual-empty-desc">可添加 ss://、vless://、trojan://、anytls://、wireguard:// 或 WireGuard 配置片段。</div>
      <button class="secondary-button small-button" type="button" data-inline-action="add">添加手动节点</button>
    </div>
  `;
}

function renderInlineCard(item, index, disabled) {
  const status = inlineStatus(item, disabled);
  const resultText = inlineResultText(item, disabled);
  return `
    <div class="manual-node-card${disabled ? " disabled" : ""}" data-manual-card-index="${index}">
      <div class="manual-node-card-header">
        <div class="manual-node-title-wrap">
          <strong>${escapeHtml(item.name || `手动节点 ${index + 1}`)}</strong>
          <span class="manual-node-status ${escapeHtml(status.className)}">${escapeHtml(status.label)}</span>
        </div>
        <div class="manual-node-header-actions">
          <button class="manual-node-delete" type="button" data-inline-action="delete" data-inline-index="${index}">删除</button>
        </div>
      </div>
      <div class="manual-node-field">
        <label>名称</label>
        <input class="manual-node-input" type="text" data-inline-field="name" data-inline-index="${index}" placeholder="手动节点 ${index + 1}" value="${escapeHtml(item.name)}" />
      </div>
      <div class="manual-node-field">
        <label>节点内容</label>
        <textarea class="manual-node-textarea" data-inline-field="content" data-inline-index="${index}" placeholder="粘贴 ss://、vless://、anytls://、wireguard:// 或 WireGuard 配置片段">${escapeHtml(item.content)}</textarea>
      </div>
      <div class="manual-node-actions">
        <button class="secondary-button small-button" type="button" data-inline-action="preview" data-inline-index="${index}">解析预览</button>
        <button class="ghost-button small-button" type="button" data-inline-action="clear" data-inline-index="${index}">清空</button>
      </div>
      <div class="manual-node-result ${escapeHtml(status.className)}${resultText ? "" : " hidden"}">
        ${escapeHtml(resultText)}
      </div>
    </div>
  `;
}

function requestDeleteInlineEntry(index) {
  const item = state.inlineEntries[index];
  if (!item) return;
  if (!item.content.trim()) {
    deleteInlineEntry(index);
    return;
  }
  openDangerDialog({
    action: "delete-manual-node",
    manualIndex: index,
    title: "删除手动节点",
    subtitle: `确定删除 “${item.name || `手动节点 ${index + 1}`}” 吗？`,
    confirmText: "删除",
    detailsHtml:
      '<div class="danger-dialog-note">该手动节点已有内容，删除后不会参与后续生成。此操作不会影响订阅源配置。</div>',
  });
}

function deleteInlineEntry(index) {
  if (index < 0 || index >= state.inlineEntries.length) return;
  state.inlineEntries.splice(index, 1);
  renderInlineManager();
  updateSubscriptionSummary();
}

function inlineStatus(item, moduleDisabled) {
  if (moduleDisabled) return { label: "已禁用", className: "disabled" };
  if (item.status === "parsed") {
    const protocol = item.parsed?.protocol || "";
    return {
      label: protocol ? `已解析 · ${protocol}` : "已解析",
      className: "success",
    };
  }
  if (item.status === "error") return { label: "解析失败", className: "error" };
  return { label: "未解析", className: "" };
}

function inlineResultText(item, moduleDisabled) {
  if (moduleDisabled) return "手动节点模块已关闭，当前内容不会参与生成。";
  if (item.status === "parsed") return item.result || "已识别。";
  if (item.status === "error")
    return item.error ? `解析失败：${item.error}` : "解析失败。";
  return "";
}

function resetInlineParseState(index) {
  const item = state.inlineEntries[index];
  if (!item) return;
  item.status = "idle";
  item.result = "";
  item.error = "";
  item.parsed = null;
}

async function previewInlineEntry(index, options = {}) {
  const item = state.inlineEntries[index];
  if (!item) return;
  if (!item.content.trim()) {
    resetInlineParseState(index);
    if (!options.silent) showToast("请先粘贴手动节点内容。", true);
    renderInlineManager();
    return;
  }
  const response = await fetchJSON("/api/parse", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ content: item.content, content_type: "inline" }),
  });
  if (
    !response?.ok ||
    response.errors?.length ||
    !Array.isArray(response.nodes) ||
    !response.nodes.length
  ) {
    item.status = "error";
    item.parsed = null;
    item.result = "";
    item.error =
      response?.errors?.[0]?.message ||
      response?.warnings?.[0] ||
      readAPIError(response) ||
      "未识别到可用节点";
    renderInlineManager();
    if (!options.silent) showToast(item.error, true);
    return;
  }
  const node = response.nodes[0];
  item.status = "parsed";
  item.parsed = {
    protocol: String(node.type || ""),
    name: String(node.name || ""),
  };
  item.result = `已识别：${inlineNodeSummary(node)}`;
  item.error = "";
  renderInlineManager();
}

function inlineNodeSummary(node) {
  const parts = [node.type, node.name].filter(Boolean);
  if (node.tls?.reality) parts.push("reality");
  if (node.transport?.network) parts.push(node.transport.network);
  if (node.type === "wireguard" && node.wireguard?.peers?.length) {
    const peer = node.wireguard.peers[0];
    const endpoint = [peer.server, peer.port].filter(Boolean).join(":");
    if (endpoint) parts.push(`endpoint ${endpoint}`);
  }
  if (!parts.length && node.server) {
    parts.push(node.server);
  }
  return parts.join(" · ");
}

function updateSubscriptionSummary() {
  const subCount = state.subscriptions.filter((item) => item.url.trim()).length;
  const inlineCount = state.inlineEntries.filter((item) =>
    item.content.trim(),
  ).length;
  const inlineLabel = state.manualNodesEnabled
    ? `${inlineCount} 个手动节点`
    : `${inlineCount} 个手动节点（已关闭）`;
  const enabledCount = state.subscriptions.filter(
    (item) => item.enabled && item.url.trim(),
  ).length;
  document.getElementById("subscription-summary").textContent =
    `已配置：${subCount} 个订阅源 · ${inlineLabel} · ${enabledCount} 个启用源`;
}

function openBatchImportDialog() {
  const dialog = document.getElementById("batch-import-dialog");
  if (!dialog.open) dialog.showModal();
}

function applyBatchImport() {
  const parsed = parseBatchImportEntries(getValue("batch-import-textarea"));
  if (!parsed.length) {
    showToast("没有识别到可导入的订阅源。", true);
    return;
  }

  const nextEntries = [...state.subscriptions];
  parsed.forEach((item, index) => {
    nextEntries.push(
      createSubscriptionEntry({
        id: `source-${nextEntries.length + 1}`,
        name: item.name || `source-${nextEntries.length + 1 + index}`,
        emoji: defaultSourceEmoji(nextEntries.length),
        url: item.url,
        enabled: true,
        user_agent: DEFAULT_SOURCE_USER_AGENT,
      }),
    );
  });
  state.subscriptions = normalizeImportedSubscriptionNames(nextEntries);
  renderSubscriptionManager();
  closeDialog("batch-import-dialog");
  setValue("batch-import-textarea", "");
  showToast(`已导入 ${parsed.length} 个订阅源。`);
}

function parseBatchImportEntries(rawText) {
  const lines = String(rawText || "")
    .replace(/\r\n/g, "\n")
    .split("\n")
    .map((line) => line.trim())
    .filter(Boolean);

  const entries = [];
  lines.forEach((line) => {
    if (/^https?:\/\//i.test(line)) {
      entries.push({ name: "", url: line });
      return;
    }
    const pipeIndex = line.indexOf("|");
    const commaIndex = line.indexOf(",");
    let splitIndex = -1;
    if (pipeIndex > 0) splitIndex = pipeIndex;
    else if (commaIndex > 0) splitIndex = commaIndex;
    if (splitIndex > 0) {
      const name = line.slice(0, splitIndex).trim();
      const url = line.slice(splitIndex + 1).trim();
      if (/^https?:\/\//i.test(url)) {
        entries.push({ name, url });
      }
    }
  });

  let autoIndex = 1;
  const named = [];
  const seen = {};
  entries.forEach((item) => {
    let name = item.name || `source-${autoIndex++}`;
    if (!item.name) {
      while (seen[name]) {
        name = `source-${autoIndex++}`;
      }
    } else {
      seen[name] = (seen[name] || 0) + 1;
      if (seen[name] > 1) name = `${name} ${seen[name]}`;
    }
    seen[name] = 1;
    named.push({ name, url: item.url });
  });
  return named;
}

function normalizeImportedSubscriptionNames(items) {
  const seen = {};
  return items.map((item, index) => {
    const baseName =
      (item.name || `source-${index + 1}`).trim() || `source-${index + 1}`;
    seen[baseName] = (seen[baseName] || 0) + 1;
    const name =
      seen[baseName] === 1 ? baseName : `${baseName} ${seen[baseName]}`;
    return {
      ...item,
      id: item.id || `source-${index + 1}`,
      name,
    };
  });
}

function updateSummary() {
  const enabledOptionCount = OUTPUT_OPTIONS.reduce(
    (count, item) => count + (getChecked(item.id) ? 1 : 0),
    0,
  );
  const backend = normalizeBackendOrigin(
    getValue("backend-origin") || window.location.origin,
  );
  const selectedTemplate = currentTemplateOption();
  const ruleModeLabel =
    state.activeSourceMode === "rules"
      ? `规则模式（${ruleModeText(state.ruleMode)}）`
      : `模板模式（${selectedTemplate.label}）`;

  const items = [
    ["客户端类型", clientTypeLabel(getValue("client-type"))],
    ["规则模式", ruleModeLabel],
    [
      state.activeSourceMode === "rules" ? "已选规则数" : "模板配置",
      state.activeSourceMode === "rules"
        ? `${state.enabledRules.size} / ${BUILTIN_RULES.length + state.customRules.length}`
        : `${selectedTemplate.ruleProfile || "模板接管"} · ${selectedTemplate.groupProfile || "默认代理组"}`,
    ],
    ["已启用选项数", `${enabledOptionCount} / ${OUTPUT_OPTIONS.length}`],
    ["后端地址", backend || "-"],
    ["发布链接", state.published?.publish_id ? "已生成随机私密链接" : "未生成"],
    ["后端状态", state.backendOnline ? "在线" : "离线"],
  ];

  document.getElementById("summary-list").innerHTML = items
    .map(
      ([label, value]) => `
      <div class="summary-item">
        <span class="summary-key">${escapeHtml(label)}</span>
        <span class="summary-value">${escapeHtml(String(value))}</span>
      </div>
    `,
    )
    .join("");

  renderConfigNodeSummary();
}

function renderConfigNodeSummary() {
  const sources = new Set(state.nodeSourceOptions);
  const hasNodeCache = state.allNodes.length > 0;
  const items = [
    [
      "总节点数",
      hasNodeCache
        ? state.nodeSummary.total || 0
        : state.statusPayload?.node_count || 0,
    ],
    ["已启用", hasNodeCache ? state.nodeSummary.enabled || 0 : "-"],
    ["已禁用", hasNodeCache ? state.nodeSummary.disabled || 0 : "-"],
    ["已修改", hasNodeCache ? state.nodeSummary.modified || 0 : "-"],
    ["有警告", hasNodeCache ? state.nodeSummary.warnings || 0 : "-"],
    ["来源数量", sources.size],
    ["上次刷新", formatChinaDateTime(state.statusPayload?.last_success_at)],
    ["下次刷新", formatChinaDateTime(state.statusPayload?.next_refresh_at)],
    ["刷新状态", summarizeRefreshState()],
  ];
  document.getElementById("config-node-summary-list").innerHTML = items
    .map(
      ([label, value]) => `
      <div class="summary-item">
        <span class="summary-key">${escapeHtml(label)}</span>
        <span class="summary-value">${escapeHtml(String(value))}</span>
      </div>
    `,
    )
    .join("");
}

async function loadSubscriptionMeta() {
  const response = await fetchJSON("/api/subscription-meta");
  if (!response?.ok) {
    state.subscriptionMeta = { aggregate: null, sources: [] };
    renderSubscriptionMeta();
    return;
  }

  state.subscriptionMeta = {
    aggregate: response.aggregate || null,
    sources: Array.isArray(response.sources) ? response.sources : [],
  };
  renderSubscriptionMeta();
}

function renderSubscriptionMeta() {
  const overview = document.getElementById("subscription-meta-overview");
  const list = document.getElementById("subscription-meta-list");
  if (!overview || !list) return;

  const aggregate = state.subscriptionMeta?.aggregate || null;
  const sources = Array.isArray(state.subscriptionMeta?.sources)
    ? state.subscriptionMeta.sources
    : [];

  if (!sources.length) {
    overview.innerHTML = `
      <div class="subscription-meta-title">暂无订阅信息</div>
      <div class="subscription-meta-line">刷新订阅后会显示各订阅源的流量和到期时间。</div>
    `;
    list.innerHTML = "";
    return;
  }

  overview.innerHTML = `
    <div class="subscription-meta-title">总览</div>
    <div class="subscription-meta-line">${escapeHtml(renderAggregateUsageLine(aggregate))}</div>
    <div class="subscription-meta-line">${escapeHtml(renderAggregateRemainingLine(aggregate))}</div>
    <div class="subscription-meta-line">${escapeHtml(renderAggregateExpireLine(aggregate))}</div>
  `;

  const showPerSource =
    state.config?.render?.subscription_info?.show_per_source !== false;
  if (!showPerSource) {
    list.innerHTML = "";
    return;
  }

  list.innerHTML = sources
    .map((meta) => {
      const severity = subscriptionMetaSeverity(meta);
      const sourceLabel = subscriptionMetaOrigin(meta);
      const subscription = state.subscriptions.find(
        (item) => item.id === meta.source_id || item.name === meta.source_name,
      );
      const displaySourceName =
        subscription?.name ||
        meta.source_name ||
        meta.source_id ||
        "未命名订阅源";
      return `
        <div class="subscription-meta-item ${severity}">
          <div class="subscription-meta-item-header">
            <div class="subscription-meta-item-title">${escapeHtml(displaySourceName)}</div>
            <span class="subscription-meta-pill">${escapeHtml(sourceLabel)}</span>
          </div>
          <div class="source-info-stack">
            <div class="subscription-meta-line">${escapeHtml(renderSourceUsageLine(meta))}</div>
            <div class="subscription-meta-line">${escapeHtml(renderSourceRemainingLine(meta))}</div>
            <div class="subscription-meta-line">${escapeHtml(renderSourceExpireLine(meta))}</div>
          </div>
        </div>
      `;
    })
    .join("");
}

function renderAggregateUsageLine(aggregate) {
  if (!aggregate) return "暂无聚合数据";
  if (aggregate.total > 0) {
    return `已用 ${formatBytes(aggregate.used)} / ${formatBytes(aggregate.total)}`;
  }
  if (aggregate.used > 0) {
    return `已用 ${formatBytes(aggregate.used)} · 无总量信息`;
  }
  return "暂无流量信息";
}

function renderAggregateRemainingLine(aggregate) {
  if (!aggregate || aggregate.total <= 0) return "无剩余流量总览";
  return `剩余 ${formatBytes(aggregate.remaining)}`;
}

function renderAggregateExpireLine(aggregate) {
  if (!aggregate || !aggregate.expire) return "无到期信息";
  const source = aggregate.expire_source_name
    ? `${aggregate.expire_source_name} · `
    : "";
  return `最近到期：${source}${formatDate(aggregate.expire)}`;
}

function renderSourceUsageLine(meta) {
  if (meta.total > 0) {
    return `已用 ${formatBytes(meta.used)} / ${formatBytes(meta.total)}`;
  }
  if (meta.used > 0) {
    return `已用 ${formatBytes(meta.used)} · 无总量信息`;
  }
  return "无流量信息";
}

function renderSourceRemainingLine(meta) {
  if (meta.total > 0 || meta.remaining > 0) {
    return `剩余 ${formatBytes(meta.remaining)}`;
  }
  return "无总量信息";
}

function renderSourceExpireLine(meta) {
  if (!meta.expire) return "无到期信息";
  return `到期 ${formatDate(meta.expire)}`;
}

function subscriptionMetaOrigin(meta) {
  if (meta.from_header && meta.from_info_node) return "Header + 信息节点";
  if (meta.from_header) return "Header";
  if (meta.from_info_node) return "信息节点";
  return "未获取";
}

function subscriptionMetaSeverity(meta) {
  const nowSec = Math.floor(Date.now() / 1000);
  if (meta.expire && meta.expire <= nowSec) return "danger";
  if (meta.used_ratio >= 0.95) return "danger";
  if (meta.expire && meta.expire - nowSec <= 7 * 24 * 3600) return "warning";
  if (meta.used_ratio >= 0.8) return "warning";
  return "";
}

function formatBytes(value) {
  const size = Number(value || 0);
  if (size <= 0) return "0 B";
  const units = ["B", "KB", "MB", "GB", "TB", "PB"];
  let index = 0;
  let current = size;
  while (current >= 1024 && index < units.length - 1) {
    current /= 1024;
    index += 1;
  }
  const decimals = current >= 100 || index === 0 ? 0 : 1;
  return `${current.toFixed(decimals)} ${units[index]}`;
}

function formatDate(unixSeconds) {
  if (!unixSeconds) return "-";
  const date = new Date(Number(unixSeconds) * 1000);
  if (Number.isNaN(date.getTime())) return "-";
  return date.toLocaleDateString("zh-CN", {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
  });
}

function summarizeRefreshState() {
  if (state.statusPayload?.refreshing) return "刷新中";
  return "空闲";
}

function clientTypeLabel(value) {
  if (value === "clash") return "Clash";
  if (value === "singbox") return "Sing-box";
  return "Mihomo / Clash Meta";
}

function ruleModeText(mode) {
  return (
    { minimal: "最小化", balanced: "均衡", full: "全面", custom: "自定义" }[
      mode
    ] || "自定义"
  );
}

function buildConfigFromForm(options = {}) {
  const { activeOnly = false } = options;
  const base = deepClone(
    state.config || {
      service: DEFAULT_SERVICE,
      render: DEFAULT_RENDER,
      subscriptions: [],
      inline: [],
    },
  );
  const prefixMode = currentSourcePrefixMode();
  const activeMode =
    state.activeSourceMode === "template" ? "template" : "rules";
  const templateExternalConfig =
    !activeOnly || activeMode === "template"
      ? {
          template_key: state.externalConfig.template_key,
          template_label: state.externalConfig.template_label,
          custom_url:
            state.externalConfig.template_key === "custom_url"
              ? getValue("custom-template-url").trim()
              : "",
        }
      : { ...DEFAULT_RENDER.external_config };

  base.service = {
    ...DEFAULT_SERVICE,
    ...(base.service || {}),
    listen_addr: "127.0.0.1",
    listen_port: 9876,
    access_token: "",
    subscription_token: "",
  };

  base.render = {
    ...DEFAULT_RENDER,
    ...(base.render || {}),
    emoji: getChecked("opt-emoji"),
    show_node_type: getChecked("opt-show-type"),
    include_info_node: getChecked("opt-info-node"),
    skip_tls_verify: getChecked("opt-skip-tls"),
    udp: getChecked("opt-udp"),
    node_list: getChecked("opt-node-list"),
    sort_nodes: getChecked("opt-sort-nodes"),
    filter_illegal: getChecked("opt-filter-illegal"),
    insert_url: getChecked("opt-insert-url"),
    source_prefix: prefixMode !== "none",
    name_options: {
      keep_raw_name: true,
      source_prefix_mode: prefixMode,
      source_prefix_separator: SOURCE_PREFIX_SEPARATOR,
      dedupe_suffix_style: "#n",
    },
    include_keywords: getValue("include-keywords").trim(),
    exclude_keywords: getValue("exclude-keywords").trim(),
    output_filename: getValue("output-filename").trim() || "mihomo.yaml",
    source_mode: activeMode,
    template_rule_mode: activeMode,
    external_config: templateExternalConfig,
    rule_mode:
      !activeOnly || activeMode === "rules"
        ? state.ruleMode
        : DEFAULT_RENDER.rule_mode,
    enabled_rules:
      !activeOnly || activeMode === "rules" ? Array.from(state.enabledRules) : [],
    custom_rules: !activeOnly || activeMode === "rules" ? state.customRules : [],
  };

  base.subscriptions = state.subscriptions
    .filter((item) => item.url.trim())
    .map((item, index) => ({
      id: item.id || `source-${index + 1}`,
      name: item.name.trim() || `source-${index + 1}`,
      emoji: normalizeSourceEmoji(item.emoji),
      source_logo: item.source_logo?.trim() || "",
      enabled: item.enabled !== false,
      url: item.url.trim(),
      user_agent: item.user_agent?.trim() || DEFAULT_SOURCE_USER_AGENT,
      insecure_skip_verify: getChecked("opt-skip-tls"),
    }));

  base.manual_nodes_enabled = Boolean(state.manualNodesEnabled);
  base.inline = state.inlineEntries
    .filter((item) => item.content.trim())
    .map((item, index) => ({
      id: item.id || `manual-${index + 1}`,
      name: item.name.trim() || `手动节点 ${index + 1}`,
      enabled: item.enabled !== false,
      content: item.content,
    }));

  return base;
}

async function generateSubscription() {
  try {
    const hasSubscription = state.subscriptions.some((item) => item.url.trim());
    const hasInline =
      state.manualNodesEnabled &&
      state.inlineEntries.some((item) => item.content.trim());
    if (!hasSubscription && !hasInline && !state.nodeSummary.total) {
      showToast("请至少添加一个订阅源或手动节点。", true);
      return;
    }
    if (!(await validateManualNodesBeforeGenerate())) {
      return;
    }

    state.generateStatus = "generating";
    state.lastError = "";
    startGenerateProgressPolling();
    renderResult();

    const config = buildConfigFromForm({ activeOnly: true });
    const saveResult = await saveConfig(config);
    if (!saveResult) {
      state.generateStatus = "error";
      state.lastError = "当前会话写入失败";
      stopGenerateProgressPolling();
      renderResult();
      return;
    }

    const refreshResult = await refreshNow();
    if (!refreshResult) {
      state.generateStatus = "error";
      stopGenerateProgressPolling();
      renderResult();
      return;
    }

    state.config = config;
    state.generatedUrl = refreshResult.subscription_url || buildSubscriptionURL();
    state.published = {
      ok: true,
      publish_id: refreshResult.publish_id || state.published?.publish_id || "",
      url: state.generatedUrl,
      subscription_url: state.generatedUrl,
      token_hint: refreshResult.token_hint || state.published?.token_hint || "",
      updated_at: new Date().toISOString(),
      access_count: state.published?.access_count || 0,
      status: "active",
    };
    state.lastGeneratedAt = formatChinaDateTime(new Date());
    state.resultNodeCount =
      Number.parseInt(refreshResult.node_count, 10) || state.resultNodeCount;
    state.refreshStatus = "fresh";
    state.generateStatus = "success";
    state.publishNotice = "";
    stopGenerateProgressPolling();
    renderResult();
    updateSummary();
    showToast("订阅链接已生成。");
    void postGenerateRefresh();
  } catch (error) {
    state.generateStatus = "error";
    state.lastError = error?.message || String(error) || "未知错误";
    stopGenerateProgressPolling();
    renderResult();
    showToast(`生成失败：${state.lastError}`, true);
  }
}

async function validateManualNodesBeforeGenerate() {
  if (!state.manualNodesEnabled) return true;
  for (let index = 0; index < state.inlineEntries.length; index += 1) {
    const item = state.inlineEntries[index];
    if (item.enabled === false || !item.content.trim()) continue;
    if (item.status === "parsed") continue;
    await previewInlineEntry(index, { silent: true });
    if (state.inlineEntries[index]?.status === "error") {
      const name = state.inlineEntries[index]?.name || `手动节点 ${index + 1}`;
      showToast(`${name} 解析失败，请修正或清空后再生成。`, true);
      return false;
    }
  }
  return true;
}

async function saveConfig(config) {
  const backendConfig = configForBackend(config);
  const response = await fetchJSON("/api/config", {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(backendConfig),
  });
  if (!response?.ok) {
    showToast(readAPIError(response) || "写入当前会话失败。", true);
    return false;
  }
  return true;
}

async function refreshNow() {
  const response = await fetchJSON("/api/refresh", {
    method: "POST",
  });
  if (!response?.ok) {
    state.lastError = readAPIError(response) || "刷新失败";
    showToast(state.lastError, true);
    return null;
  }
  state.lastError = "";
  return response;
}

async function postGenerateRefresh() {
  const tasks = [
    loadStatus(),
    loadLogs(),
    loadSubscriptionMeta(),
    loadAudit(),
    loadPublishedStatus(),
  ];
  if (state.activeWorkspace === "yaml") {
    tasks.push(loadYamlPreview());
  }
  if (state.activeWorkspace === "nodes") {
    tasks.push(refreshNodesAfterGenerate());
  } else if (state.activeWorkspace === "diagnostics") {
    tasks.push(validateNodes());
  }
  await Promise.all(tasks);
  renderResult();
}

function startGenerateProgressPolling() {
  stopGenerateProgressPolling();
  state.generateProgressTimer = window.setInterval(() => {
    void loadStatus();
  }, 350);
}

function stopGenerateProgressPolling() {
  if (state.generateProgressTimer) {
    window.clearInterval(state.generateProgressTimer);
    state.generateProgressTimer = null;
  }
}

function renderResult() {
  const panel = document.getElementById("result-panel");
  const summaryList = document.getElementById("result-summary-list");
  const actionStatus = document.getElementById("action-result-status");
  const time = document.getElementById("action-result-time");
  const input = document.getElementById("generated-url");
  const copyButton = document.getElementById("copy-generated-url-btn");
  const openButton = document.getElementById("open-generated-url-btn");
  const viewButton = document.getElementById("view-generated-link-btn");
  const rotateButton = document.getElementById("rotate-token-btn");
  const deleteButton = document.getElementById("delete-published-btn");
  const generateButton = document.getElementById("generate-btn");
  const hasResult = Boolean(state.generatedUrl);
  const published = state.published;

  let statusLabel = "未生成";
  let statusText = "尚未生成订阅链接";
  let refreshStatusText =
    state.refreshStatus && state.refreshStatus !== "idle"
      ? state.refreshStatus
      : "-";
  let statusClass = "result-empty";
  let statusDotClass = "";

  if (state.generateStatus === "generating") {
    statusLabel = "生成中";
    statusText = refreshStageLabel(state.refreshStage);
    statusClass = "result-warning";
    statusDotClass = "warning";
    refreshStatusText = refreshStageLabel(state.refreshStage);
  } else if (state.generateStatus === "success" && hasResult) {
    statusLabel = "已生成";
    statusText = "订阅链接已生成";
    statusClass = "";
    statusDotClass = "success";
  } else if (state.generateStatus === "error") {
    statusLabel = "生成失败";
    statusText = state.lastError || "生成失败";
    statusClass = "result-error";
    statusDotClass = "error";
  } else if (state.publishNotice) {
    statusLabel = "未发布";
    statusText = state.publishNotice;
    statusClass = "result-warning";
    statusDotClass = "warning";
  }

  if (panel) {
    panel.classList.toggle("success-panel", state.generateStatus === "success");
    panel.classList.toggle("error-panel", state.generateStatus === "error");
  }
  if (actionStatus) {
    actionStatus.className = `result-status ${statusClass}`.trim();
    actionStatus.innerHTML = `
      <span class="status-dot ${statusDotClass}"></span>
      <span>${escapeHtml(statusText)}</span>
    `;
  }
  if (time) {
    time.textContent = state.lastGeneratedAt || "";
  }
  if (input) {
    input.disabled = !hasResult;
    input.placeholder = hasResult ? "" : "点击右侧按钮生成订阅链接";
    input.value = hasResult ? state.generatedUrl : "";
  }
  if (copyButton) {
    setButtonIconText(copyButton, "复制");
    copyButton.disabled = !hasResult;
  }
  if (openButton) {
    setButtonIconText(openButton, "打开");
    openButton.disabled = !hasResult;
  }
  if (viewButton) {
    viewButton.disabled = !hasResult && state.generateStatus !== "error";
  }
  if (generateButton) {
    generateButton.disabled = state.generateStatus === "generating";
    if (state.generateStatus === "generating") {
      setButtonIconText(generateButton, "生成中...");
    } else if (state.generateStatus === "success") {
      setButtonIconText(generateButton, "重新生成配置");
    } else {
      setButtonIconText(generateButton, "生成订阅链接");
    }
  }
  if (rotateButton) {
    rotateButton.disabled =
      !published?.publish_id || state.generateStatus === "generating";
  }
  if (deleteButton) {
    deleteButton.disabled =
      !published?.publish_id || state.generateStatus === "generating";
  }

  const items = [
    ["状态", statusLabel],
    ["生成时间", state.lastGeneratedAt || "-"],
    ["发布 ID", published?.publish_id || "-"],
    ["Token 摘要", published?.token_hint || "-"],
    [
      "节点数量",
      hasResult
        ? state.resultNodeCount
          ? String(state.resultNodeCount)
          : "-"
        : "-",
    ],
    [
      "规则数量",
      hasResult
        ? state.resultRuleCount
          ? String(state.resultRuleCount)
          : "-"
        : "-",
    ],
    ["上次访问", formatChinaDateTime(published?.last_access_at)],
    [
      "访问次数",
      published
        ? String(Number.parseInt(published.access_count, 10) || 0)
        : "-",
    ],
    ["刷新状态", hasResult ? refreshStatusText : "-"],
  ];

  if (summaryList) {
    summaryList.innerHTML = items
      .map(
        ([label, value]) => `
        <div class="summary-item">
          <span class="summary-key">${escapeHtml(label)}</span>
          <span class="summary-value">${escapeHtml(String(value))}</span>
        </div>
      `,
      )
      .join("");
  }
}

function refreshStageLabel(stage) {
  switch (String(stage || "").toLowerCase()) {
    case "fetching":
      return "正在抓取订阅";
    case "rendering":
      return "正在渲染配置";
    case "writing":
      return "正在写入输出";
    default:
      return "正在生成订阅链接";
  }
}

async function importClash() {
  if (!state.generatedUrl.trim()) {
    await generateSubscription();
  }
  const url = state.generatedUrl.trim();
  if (!url) return;
  openUrl(`clash://install-config?url=${encodeURIComponent(url)}`);
}

async function loadPublishedStatus() {
  const response = await fetchJSON("/api/published");
  if (!normalizePublishedStatus(response)) {
    state.published = null;
    if (!state.statusPayload?.yaml_exists) {
      state.generatedUrl = "";
    }
    if (!state.publishNotice) {
      state.generateStatus = "idle";
    }
    renderResult();
    return;
  }
  applyPublishedStatus(response);
}

async function syncPublishedStatusForDraft() {
  const response = await fetchJSON("/api/published");
  const published = normalizePublishedStatus(response);
  if (!published) {
    state.published = null;
    return null;
  }
  state.published = published;
  state.generatedUrl = published.url || state.generatedUrl;
  if (state.generatedUrl && state.generateStatus !== "error") {
    state.generateStatus = "success";
  }
  if (published.updated_at) {
    state.lastGeneratedAt = formatChinaDateTime(published.updated_at, {
      fallback: published.updated_at,
    });
  }
  renderResult();
  return published;
}

async function fetchPublishedByID(publishID) {
  const response = await fetchJSON(
    `/api/published/${encodeURIComponent(publishID)}`,
  );
  return normalizePublishedStatus(response);
}

async function bindWorkspacePublished(publishID) {
  if (!state.workspaceId) return false;
  const response = await fetchJSON(
    `/api/workspaces/${encodeURIComponent(state.workspaceId)}/bind-publish`,
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ publish_id: publishID }),
    },
  );
  if (!response?.ok) {
    showToast(readAPIError(response) || "恢复发布绑定失败。", true);
    return false;
  }
  return true;
}

async function rotatePublishedToken() {
  const publishID = state.published?.publish_id;
  if (!publishID) {
    showToast("当前还没有可轮换的私密链接。", true);
    return;
  }
  const response = await fetchJSON(
    `/api/published/${encodeURIComponent(publishID)}/rotate-token`,
    {
      method: "POST",
    },
  );
  if (!response?.ok) {
    showToast(readAPIError(response) || "重新生成私密链接失败。", true);
    return;
  }
  applyPublishedStatus(response);
  if (!response.updated_at) {
    state.lastGeneratedAt = formatChinaDateTime(new Date());
  }
  if (state.draftMode === "local_draft") {
    await persistLocalDraft(true, { silent: true });
  }
  renderResult();
  showToast("私密链接已轮换，旧链接立即失效。");
}

async function deletePublishedLink() {
  const publishID = state.published?.publish_id;
  if (!publishID) {
    showToast("当前没有可删除的发布链接。", true);
    return;
  }
  const response = await fetchJSON(
    `/api/published/${encodeURIComponent(publishID)}`,
    {
      method: "DELETE",
    },
  );
  if (!response?.ok) {
    showToast(readAPIError(response) || "删除发布失败。", true);
    return;
  }
  state.published = null;
  state.generatedUrl = "";
  state.generateStatus = "idle";
  state.publishNotice = "";
  if (state.hasLocalDraft) {
    removeLocalDraftPublishRef(publishID);
  }
  renderResult();
  showToast("当前发布已删除，订阅链接已失效。");
}

async function copyGeneratedUrl() {
  const url = state.generatedUrl.trim();
  if (!url) {
    showToast("请先生成订阅链接。", true);
    return;
  }
  const button = document.getElementById("copy-generated-url-btn");
  await copyText(url);
  setButtonIconText(button, "已复制");
  window.setTimeout(() => setButtonIconText(button, "复制"), 1200);
  showToast("订阅链接已复制。");
}

function focusGeneratedLink() {
  const bar = document.getElementById("sticky-action-bar");
  const input = document.getElementById("generated-url");
  if (!bar || !input) return;
  bar.scrollIntoView({ behavior: "smooth", block: "end" });
  input.classList.add("result-highlight");
  if (!input.disabled) {
    input.focus({ preventScroll: true });
    input.select();
  }
  window.setTimeout(() => input.classList.remove("result-highlight"), 1600);
}

function previewYAMLWorkspace() {
  switchWorkspace("yaml");
  loadYamlPreview();
}

async function loadNodes() {
  const response = await fetchJSON("/api/nodes?all=1");
  if (!response?.ok) {
    document.getElementById("node-table-body").innerHTML =
      `<tr><td colspan="7" class="table-empty">${escapeHtml(readAPIError(response) || "加载节点失败。")}</td></tr>`;
    showToast(readAPIError(response) || "加载节点失败。", true);
    return;
  }

  state.allNodes = Array.isArray(response.nodes) ? response.nodes : [];
  state.nodeSourceOptions = uniqueOrdered(
    state.allNodes.map((node) => node.source?.name).filter(Boolean),
  );
  syncSourceFilterOptions();
  applyNodeFiltersAndPagination();
}

function applyNodeFiltersAndPagination() {
  const q = String(state.nodeFilters.q || "")
    .trim()
    .toLowerCase();
  const typeFilter = String(state.nodeFilters.type || "all").toLowerCase();
  const regionFilter = String(state.nodeFilters.region || "ALL").toUpperCase();
  const statusFilter = String(state.nodeFilters.status || "all").toLowerCase();
  const sourceFilter = String(state.nodeFilters.source || "")
    .trim()
    .toLowerCase();

  state.filteredNodes = state.allNodes.filter((node) => {
    if (q) {
      const haystack = [
        node.name,
        node.server,
        node.type,
        regionLabel(node.region),
        node.region,
        ...(Array.isArray(node.tags) ? node.tags : []),
      ]
        .join(" ")
        .toLowerCase();
      if (!haystack.includes(q)) return false;
    }
    if (
      typeFilter !== "all" &&
      String(node.type || "").toLowerCase() !== typeFilter
    )
      return false;
    if (
      regionFilter !== "ALL" &&
      String(node.region || "").toUpperCase() !== regionFilter
    )
      return false;
    if (
      sourceFilter &&
      !String(node.source?.name || "")
        .toLowerCase()
        .includes(sourceFilter)
    )
      return false;
    if (statusFilter === "enabled" && !node.enabled) return false;
    if (statusFilter === "disabled" && node.enabled) return false;
    if (statusFilter === "modified" && !node.modified) return false;
    if (statusFilter === "warning" && !(node.warnings && node.warnings.length))
      return false;
    return true;
  });

  state.nodeSummary = summarizeNodeItems(state.filteredNodes);
  state.nodePagination.total = state.filteredNodes.length;

  const totalPages = Math.max(
    1,
    Math.ceil(
      (state.nodePagination.total || 0) / state.nodePagination.pageSize,
    ),
  );
  if (state.nodePagination.page > totalPages) {
    state.nodePagination.page = totalPages;
  }

  const start = (state.nodePagination.page - 1) * state.nodePagination.pageSize;
  const end = start + state.nodePagination.pageSize;
  state.nodes = state.filteredNodes.slice(start, end);

  renderNodeEditor();
  renderConfigNodeSummary();
}

function summarizeNodeItems(nodes) {
  const summary = {
    total: nodes.length,
    enabled: 0,
    disabled: 0,
    modified: 0,
    warnings: 0,
  };
  nodes.forEach((node) => {
    if (node.enabled) summary.enabled += 1;
    else summary.disabled += 1;
    if (node.modified) summary.modified += 1;
    if (node.warnings?.length) summary.warnings += 1;
  });
  return summary;
}

function renderNodeEditor() {
  renderNodeStatsStrip();
  renderNodeTable();
  renderNodePagination();
}

function renderNodeStatsStrip() {
  const items = [
    ["共", `${state.nodeSummary.total || 0} 个节点`],
    ["已启用", state.nodeSummary.enabled || 0],
    ["已修改", state.nodeSummary.modified || 0],
    ["已禁用", state.nodeSummary.disabled || 0],
    ["有警告", state.nodeSummary.warnings || 0],
  ];
  document.getElementById("node-stats-strip").innerHTML = items
    .map(
      ([label, value]) =>
        `<span class="stat-pill"><strong>${escapeHtml(String(value))}</strong><span>${escapeHtml(label)}</span></span>`,
    )
    .join("");
}

function syncSourceFilterOptions() {
  const select = document.getElementById("node-source-filter");
  const values = uniqueOrdered(
    [state.nodeFilters.source, ...state.nodeSourceOptions].filter(Boolean),
  );
  const current = state.nodeFilters.source || "";
  select.innerHTML = [`<option value="">全部来源</option>`]
    .concat(
      values.map(
        (value) =>
          `<option value="${escapeHtml(value)}"${value === current ? " selected" : ""}>${escapeHtml(value)}</option>`,
      ),
    )
    .join("");
}

function renderNodeTable() {
  const body = document.getElementById("node-table-body");
  if (!state.nodes.length) {
    body.innerHTML = `<tr><td colspan="6" class="table-empty">当前没有可显示的节点</td></tr>`;
    return;
  }

  body.innerHTML = state.nodes
    .map((node) => {
      const featureBadges = buildNodeFeatureBadges(node).join("");
      const subscription = state.subscriptions.find(
        (item) =>
          item.id === node.source?.id || item.name === node.source?.name,
      );
      const displaySourceName = subscription?.name || node.source?.name || "-";
      const disabledBadge = !node.enabled
        ? '<span class="mini-badge disabled">已禁用</span>'
        : "";
      return `
        <tr>
          <td>
            <div class="node-cell">
              <span class="status-dot ${node.enabled ? "enabled" : "disabled"}"></span>
              <div class="node-cell-copy">
                <div class="node-title">${escapeHtml(node.name)}</div>
                <div class="node-submeta">
                  ${node.modified ? '<span class="mini-badge modified">已修改</span>' : ""}
                  ${node.warnings?.length ? `<span class="mini-badge warning">有警告 ${node.warnings.length}</span>` : ""}
                  ${disabledBadge}
                  <span class="mini-badge">ID ${escapeHtml(shortNodeID(node.id))}</span>
                </div>
              </div>
            </div>
          </td>
          <td>
            <div class="node-badge-list">
              <span class="mini-badge type-badge type-${escapeHtml(node.type)}">${escapeHtml(node.type)}</span>
              <span class="mini-badge region-badge">${escapeHtml(regionLabel(node.region))}</span>
            </div>
          </td>
          <td><span class="address-ellipsis" title="${escapeHtml(`${node.server || "-"}:${node.port || ""}`)}">${escapeHtml(`${node.server || "-"}${node.port ? `:${node.port}` : ""}`)}</span></td>
          <td><div class="node-feature-list">${featureBadges || '<span class="mini-badge">-</span>'}</div></td>
          <td>
            <div class="source-info-stack">
              <div>${escapeHtml(displaySourceName)}</div>
              <div class="meta-text">${escapeHtml(node.source?.kind || "")}</div>
            </div>
          </td>
          <td>
            <div class="node-actions">
              <button class="tiny-button primary-text" type="button" data-node-action="edit" data-node-id="${escapeHtml(node.id)}">编辑</button>
              <div class="node-delete-wrap">
                <button class="tiny-button ${node.enabled === false && node.source?.kind !== "custom" ? "primary-text" : "danger-ghost"}" type="button" data-node-action="delete" data-node-delete-trigger data-node-id="${escapeHtml(node.id)}">${node.enabled === false && node.source?.kind !== "custom" ? "恢复" : "删除"}</button>
                ${state.activeNodeDeletePopoverId === node.id ? renderNodeDeletePopconfirm(node) : ""}
              </div>
            </div>
          </td>
        </tr>
      `;
    })
    .join("");

  body.querySelectorAll("[data-node-action]").forEach((button) => {
    button.addEventListener("click", async (event) => {
      const nodeId = button.dataset.nodeId;
      const action = button.dataset.nodeAction;
      if (action === "edit") await openNodeDialog(nodeId);
      if (action === "delete") {
        event.stopPropagation();
        toggleNodeDeletePopover(nodeId);
      }
    });
  });
  body.querySelectorAll("[data-node-delete-cancel]").forEach((button) => {
    button.addEventListener("click", (event) => {
      event.stopPropagation();
      state.activeNodeDeletePopoverId = "";
      renderNodeTable();
    });
  });
  body.querySelectorAll("[data-node-delete-confirm]").forEach((button) => {
    button.addEventListener("click", async (event) => {
      event.stopPropagation();
      const nodeId = button.dataset.nodeDeleteConfirm;
      const node =
        state.nodes.find((item) => item.id === nodeId) || state.editingNode;
      if (node?.enabled === false && node.source?.kind !== "custom") {
        await performRestoreNode(nodeId);
      } else {
        await performDeleteNode(nodeId);
      }
    });
  });
  if (state.activeNodeDeletePopoverId) {
    window.requestAnimationFrame(() => {
      document
        .querySelector(
          `[data-node-delete-cancel="${CSS.escape(state.activeNodeDeletePopoverId)}"]`,
        )
        ?.focus();
    });
  }
}

function toggleNodeDeletePopover(nodeId) {
  state.activeNodeDeletePopoverId =
    state.activeNodeDeletePopoverId === nodeId ? "" : nodeId;
  renderNodeTable();
}

function renderNodeDeletePopconfirm(node) {
  const semantics = nodeDeleteSemantics(node);
  return `
    <div class="node-popconfirm" data-popconfirm-root>
      <div class="node-popconfirm-title">${escapeHtml(semantics.inlineTitle)}</div>
      <div class="node-popconfirm-desc">${escapeHtml(semantics.inlineDescription)}</div>
      <div class="node-popconfirm-actions">
        <button class="ghost-button small-button" type="button" data-node-delete-cancel="${escapeHtml(node.id)}">取消</button>
        <button class="${semantics.inlineConfirmClass} small-button" type="button" data-node-delete-confirm="${escapeHtml(node.id)}">${escapeHtml(semantics.inlineConfirmText)}</button>
      </div>
    </div>
  `;
}

function nodeDeleteSemantics(node) {
  const isManual = node?.source?.kind === "custom";
  const isDisabled = node?.enabled === false;
  return {
    isManual,
    isDisabled,
    sourceName: node?.source?.name || "-",
    sourceType: isManual ? "手动节点" : "订阅节点",
    inlineTitle: isDisabled && !isManual ? "恢复这个节点？" : "删除这个节点？",
    inlineDescription:
      isDisabled && !isManual
        ? "将取消本地隐藏，节点会重新参与生成。"
        : isManual
          ? "将从当前配置中移除，无法恢复。"
          : "将标记为本地删除，后续订阅更新时不再显示。",
    inlineConfirmText: isDisabled && !isManual ? "恢复" : "删除",
    inlineConfirmClass:
      isDisabled && !isManual ? "secondary-button" : "danger-solid-button",
    modalDescription:
      isDisabled && !isManual
        ? "恢复后将重新显示，并继续参与后续生成。"
        : isManual
          ? "删除后将从当前配置中移除，此操作不可撤销。"
          : "删除后将标记为本地隐藏，订阅刷新后不会恢复显示。",
  };
}

function buildNodeFeatureBadges(node) {
  const badges = [];
  if (node.tls?.enabled) badges.push('<span class="mini-badge">TLS</span>');
  if (node.udp) badges.push('<span class="mini-badge">UDP</span>');
  if (node.has_reality) badges.push('<span class="mini-badge">Reality</span>');
  if (node.transport_network === "xhttp")
    badges.push('<span class="mini-badge">XHTTP</span>');
  if (node.has_ipv6) badges.push('<span class="mini-badge">IPv6</span>');
  if (node.warnings?.length)
    badges.push(
      `<span class="mini-badge warning">警告 ${node.warnings.length}</span>`,
    );
  return badges;
}

function renderNodePagination() {
  const pageSize = state.nodePagination.pageSize;
  const total = state.nodePagination.total || 0;
  const totalPages = Math.max(1, Math.ceil(total / pageSize));
  if (state.nodePagination.page > totalPages) {
    state.nodePagination.page = totalPages;
  }
  const page = state.nodePagination.page;
  const start = total === 0 ? 0 : (page - 1) * pageSize + 1;
  const end = total === 0 ? 0 : Math.min(page * pageSize, total);
  setValue("node-page-size", String(pageSize));
  document.getElementById("node-range-info").textContent =
    total === 0 ? "0 / 0 个节点" : `${start}-${end} / ${total} 个节点`;
  document.getElementById("node-page-info").textContent =
    `${page} / ${totalPages}`;
  document.getElementById("node-prev-page-btn").disabled =
    page <= 1;
  document.getElementById("node-next-page-btn").disabled =
    page >= totalPages;
}

function handleNodeFilterChange() {
  state.nodeFilters.type = getValue("node-type-filter");
  state.nodeFilters.region = getValue("node-region-filter");
  state.nodeFilters.status = getValue("node-status-filter");
  state.nodeFilters.source = getValue("node-source-filter");
  state.nodePagination.page = 1;
  applyNodeFiltersAndPagination();
}

function handleNodePageSizeChange() {
  state.nodePagination.pageSize =
    Number.parseInt(getValue("node-page-size"), 10) || 25;
  state.nodePagination.page = 1;
  applyNodeFiltersAndPagination();
}

function changeNodePage(direction) {
  const totalPages = Math.max(
    1,
    Math.ceil(
      (state.nodePagination.total || 0) / state.nodePagination.pageSize,
    ),
  );
  const next = state.nodePagination.page + direction;
  if (next < 1 || next > totalPages) return;
  state.nodePagination.page = next;
  applyNodeFiltersAndPagination();
}

async function openNodeDialog(nodeId) {
  state.activeNodeDeletePopoverId = "";
  renderNodeTable();
  const dialog = document.getElementById("node-dialog");
  const localNode =
    state.nodes.find((node) => node.id === nodeId) ||
    state.allNodes.find((node) => node.id === nodeId);
  if (localNode) {
    state.editingNode = localNode;
    fillNodeDialog(localNode);
    if (!dialog.open) dialog.showModal();
    return;
  }

  const response = await fetchJSON(`/api/nodes/${encodeURIComponent(nodeId)}`);
  if (!response?.ok) {
    showToast(readAPIError(response) || "加载节点详情失败。", true);
    return;
  }

  state.editingNode = response.node;
  fillNodeDialog(response.node);
  if (!dialog.open) dialog.showModal();
}

function fillNodeDialog(node) {
  setValue("edit-node-id", node.id || "");
  setValue("edit-node-name", node.name || "");
  setValue("edit-node-type", node.type || "");
  setValue(
    "edit-node-source",
    `${node.source?.name || "-"}${node.source?.kind ? ` · ${node.source.kind}` : ""}`,
  );
  setValue("edit-node-stable-short", node.stable_short || shortNodeID(node.id));
  document.getElementById("node-dialog-meta").textContent =
    `${node.name || "-"} · ${node.type || "-"} · ${node.source?.name || "-"}`;
  const semantics = nodeDeleteSemantics(node);
  const actionButton = document.getElementById("delete-node-btn");
  const help = document.getElementById("node-dialog-help");
  if (actionButton) {
    actionButton.textContent =
      semantics.isDisabled && !semantics.isManual ? "恢复节点" : "删除节点";
    actionButton.classList.remove("danger-ghost-button", "secondary-button");
    actionButton.classList.add(
      semantics.isDisabled && !semantics.isManual
        ? "secondary-button"
        : "danger-ghost-button",
    );
  }
  if (help) {
    help.textContent =
      semantics.isDisabled && !semantics.isManual
        ? "该节点当前处于本地隐藏状态，可恢复显示。"
        : "删除操作会进入应用内确认流程，不会直接执行。";
  }
}

function closeNodeDialog() {
  closeDialog("node-dialog");
}

function openDangerDialog(config) {
  state.confirmDialog = config;
  const dialog = document.getElementById("danger-dialog");
  if (!dialog) return;

  document.getElementById("danger-dialog-title").textContent =
    config.title || "危险操作";
  document.getElementById("danger-dialog-subtitle").textContent =
    config.subtitle || "";
  document.getElementById("danger-dialog-details").innerHTML =
    config.detailsHtml || "";
  setButtonIconText("danger-dialog-confirm-btn", config.confirmText || "确认");

  if (!dialog.open) dialog.showModal();
  window.requestAnimationFrame(() => {
    document.getElementById("danger-dialog-cancel-btn")?.focus();
  });
}

function closeDangerDialog() {
  state.confirmDialog = null;
  closeDialog("danger-dialog");
}

async function handleDangerDialogConfirm() {
  const config = state.confirmDialog;
  if (!config) return;
  if (config.action === "delete-node") {
    const ok = await performDeleteNode(config.nodeId);
    if (ok) {
      closeDangerDialog();
      if (config.closeNodeDialog) closeNodeDialog();
    }
    return;
  }
  if (config.action === "clear-overrides") {
    const ok = await performClearNodeOverrides();
    if (ok) {
      closeDangerDialog();
    }
    return;
  }
  if (config.action === "delete-manual-node") {
    deleteInlineEntry(config.manualIndex);
    closeDangerDialog();
  }
}

function closeDialog(id) {
  const dialog = document.getElementById(id);
  if (dialog?.open) dialog.close();
}

function toggleSecretField(inputId, button) {
  const input = document.getElementById(inputId);
  if (!input) return;
  const showing = input.type === "text";
  input.type = showing ? "password" : "text";
  button.textContent = showing ? "显示" : "隐藏";
}

function buildNodeOverridePayload() {
  return {
    enabled: Boolean(state.editingNode?.enabled),
    name: getValue("edit-node-name").trim(),
  };
}

async function saveNodeOverride(regenerate) {
  const nodeId = getValue("edit-node-id");
  if (!nodeId) return;

  let payload;
  try {
    payload = buildNodeOverridePayload();
  } catch (error) {
    showToast(error.message, true);
    return;
  }

  const response = await fetchJSON(
    `/api/nodes/${encodeURIComponent(nodeId)}/override`,
    {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    },
  );
  if (!response?.ok) {
    showToast(readAPIError(response) || "保存节点失败。", true);
    return;
  }

  closeNodeDialog();
  await refreshNodesAfterGenerate();
  showToast("节点名称已保存。");
}

async function handleDeleteCurrentNode() {
  const node = state.editingNode;
  const nodeId = getValue("edit-node-id");
  if (!nodeId || !node) return;
  const semantics = nodeDeleteSemantics(node);
  if (semantics.isDisabled && !semantics.isManual) {
    const ok = await performRestoreNode(nodeId);
    if (ok) closeNodeDialog();
    return;
  }
  openDangerDialog({
    action: "delete-node",
    nodeId,
    closeNodeDialog: true,
    title: "删除节点",
    subtitle: `确定删除 “${node.name || "-"}” 吗？`,
    confirmText: "删除节点",
    detailsHtml: `
      <div class="danger-dialog-info">
        <div class="danger-dialog-row"><span>来源</span><strong>${escapeHtml(semantics.sourceName)}</strong></div>
        <div class="danger-dialog-row"><span>类型</span><strong>${escapeHtml(semantics.sourceType)}</strong></div>
      </div>
      <div class="danger-dialog-note">${escapeHtml(semantics.modalDescription)}</div>
    `,
  });
}

async function resetNodeOverride(nodeId) {
  const response = await fetchJSON(
    `/api/nodes/${encodeURIComponent(nodeId)}/reset`,
    {
      method: "POST",
    },
  );
  if (!response?.ok) {
    showToast(readAPIError(response) || "重置节点失败。", true);
    return;
  }
  await refreshNodesAfterGenerate();
  showToast("节点已恢复原始值。");
}

async function toggleNodeEnabled(nodeId, enabled) {
  const response = await fetchJSON(
    enabled ? "/api/nodes/enable" : "/api/nodes/disable",
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ ids: [nodeId] }),
    },
  );
  if (!response?.ok) {
    showToast(readAPIError(response) || "切换节点状态失败。", true);
    return;
  }
  await refreshNodesAfterGenerate();
}

async function performRestoreNode(nodeId) {
  const response = await fetchJSON("/api/nodes/enable", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ ids: [nodeId] }),
  });
  if (!response?.ok) {
    showToast(readAPIError(response) || "恢复节点失败。", true);
    return false;
  }
  state.activeNodeDeletePopoverId = "";
  await refreshNodesAfterGenerate();
  showToast("节点已恢复显示。");
  return true;
}

async function performDeleteNode(nodeId) {
  const node =
    state.nodes.find((item) => item.id === nodeId) || state.editingNode;
  if (!node) return false;
  state.activeNodeDeletePopoverId = "";
  renderNodeTable();
  if (node.source?.kind === "custom") {
    return performDeleteCustomNode(nodeId);
  }
  const response = await fetchJSON("/api/nodes/delete", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ ids: [nodeId] }),
  });
  if (!response?.ok) {
    showToast(readAPIError(response) || "删除节点失败。", true);
    return false;
  }
  await refreshNodesAfterGenerate();
  showToast("节点已删除。");
  return true;
}

async function copyNodeDetail(nodeId) {
  const response = await fetchJSON(`/api/nodes/${encodeURIComponent(nodeId)}`);
  if (!response?.ok) {
    showToast(readAPIError(response) || "复制节点失败。", true);
    return;
  }
  await copyText(JSON.stringify(response.node, null, 2));
  showToast("节点详情已复制。");
}

async function performDeleteCustomNode(nodeId) {
  const response = await fetchJSON(
    `/api/nodes/custom/${encodeURIComponent(nodeId)}`,
    {
      method: "DELETE",
    },
  );
  if (!response?.ok) {
    showToast(readAPIError(response) || "删除手动节点失败。", true);
    return false;
  }
  state.selectedNodeIds.delete(nodeId);
  await refreshNodesAfterGenerate();
  showToast("手动节点已删除。");
  return true;
}

async function clearNodeOverrides() {
  openDangerDialog({
    action: "clear-overrides",
    title: "清空覆盖规则",
    subtitle: "确定清空全部节点覆盖规则和禁用状态吗？",
    confirmText: "清空规则",
    detailsHtml: `<div class="danger-dialog-note">清空后将恢复节点默认显示状态，已重命名和已禁用记录会一并移除。</div>`,
  });
}

async function performClearNodeOverrides() {
  const response = await fetchJSON("/api/nodes/overrides/clear", {
    method: "POST",
  });
  if (!response?.ok) {
    showToast(readAPIError(response) || "清空覆盖规则失败。", true);
    return false;
  }
  state.selectedNodeIds.clear();
  await refreshNodesAfterGenerate();
  showToast("节点覆盖规则已清空。");
  return true;
}

function openBulkRenameDialog(scope) {
  if (scope) setValue("bulk-rename-scope", scope);
  else
    setValue(
      "bulk-rename-scope",
      state.selectedNodeIds.size ? "selected" : "current_filtered",
    );
  previewBulkRename();
  const dialog = document.getElementById("bulk-rename-dialog");
  if (!dialog.open) dialog.showModal();
}

function previewBulkRename() {
  const preview = document.getElementById("bulk-rename-preview");
  const targets = getBulkRenameTargets();
  if (!targets.length) {
    preview.textContent = "当前没有可预览的目标节点";
    return;
  }
  const lines = targets
    .slice(0, 10)
    .map((node) => `${node.name} -> ${previewRenamedName(node)}`);
  preview.textContent = lines.join("\n");
}

function getBulkRenameTargets() {
  const scope = getValue("bulk-rename-scope");
  if (scope === "selected") {
    return state.nodes.filter((node) => state.selectedNodeIds.has(node.id));
  }
  return state.nodes;
}

function previewRenamedName(node) {
  const mode = getValue("bulk-rename-mode");
  const prefix = getValue("bulk-rename-prefix");
  const suffix = getValue("bulk-rename-suffix");
  const pattern = getValue("bulk-rename-pattern");
  const replacement = getValue("bulk-rename-replacement");

  if (mode === "add_prefix") return `${prefix}${node.name}`.trim();
  if (mode === "add_suffix") return `${node.name}${suffix}`.trim();
  if (mode === "protocol_prefix")
    return `[${String(node.type).toUpperCase()}] ${node.name}`.trim();
  if (mode === "region_emoji")
    return `${regionEmoji(node.region)} ${node.name}`.trim();
  if (mode === "remove_info_text") {
    return node.name
      .replace(
        /(剩余流量[:：]?\s*[^ ]+|到期时间[:：]?\s*[^ ]+|官网[:：]?\s*[^ ]+|套餐[:：]?\s*[^ ]+)/gi,
        "",
      )
      .trim();
  }
  if (mode === "regex_replace") {
    try {
      return node.name.replace(new RegExp(pattern, "g"), replacement).trim();
    } catch {
      return node.name;
    }
  }
  return node.name;
}

async function applyBulkRename() {
  const payload = {
    scope: getValue("bulk-rename-scope"),
    ids:
      getValue("bulk-rename-scope") === "selected"
        ? Array.from(state.selectedNodeIds)
        : [],
    mode: getValue("bulk-rename-mode"),
    prefix: getValue("bulk-rename-prefix"),
    suffix: getValue("bulk-rename-suffix"),
    pattern: getValue("bulk-rename-pattern"),
    replacement: getValue("bulk-rename-replacement"),
    q: state.nodeFilters.q,
    type: state.nodeFilters.type,
    region: state.nodeFilters.region,
    status: state.nodeFilters.status,
    source: state.nodeFilters.source,
  };
  const response = await fetchJSON("/api/nodes/bulk-rename", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });
  if (!response?.ok) {
    showToast(readAPIError(response) || "批量重命名失败。", true);
    return;
  }
  closeDialog("bulk-rename-dialog");
  await refreshNodesAfterGenerate();
  showToast(`批量重命名已应用，共修改 ${response.changed || 0} 个节点。`);
}

function openAddNodeDialog() {
  switchAddNodeMode("uri");
  const dialog = document.getElementById("add-node-dialog");
  if (!dialog.open) dialog.showModal();
}

function switchAddNodeMode(mode) {
  state.addNodeMode = mode;
  document
    .getElementById("add-node-tab-uri")
    .classList.toggle("active", mode === "uri");
  document
    .getElementById("add-node-tab-manual")
    .classList.toggle("active", mode === "manual");
  document
    .getElementById("add-node-uri-panel")
    .classList.toggle("hidden", mode !== "uri");
  document
    .getElementById("add-node-manual-panel")
    .classList.toggle("hidden", mode !== "manual");
}

async function addCustomNode() {
  let payload;
  if (state.addNodeMode === "uri") {
    const content = getValue("custom-node-uri").trim();
    if (!content) {
      showToast("请填写节点 URI。", true);
      return;
    }
    payload = { content, content_type: "uri" };
  } else {
    payload = {
      node: {
        name: getValue("manual-node-name").trim(),
        type: getValue("manual-node-type"),
        server: getValue("manual-node-server").trim(),
        port: Number.parseInt(getValue("manual-node-port"), 10) || 0,
        auth: {
          uuid: getValue("manual-node-uuid").trim(),
          password: getValue("manual-node-password").trim(),
        },
        tls: {
          enabled: Boolean(getValue("manual-node-sni").trim()),
          sni: getValue("manual-node-sni").trim(),
          client_fingerprint: getValue("manual-node-client-fingerprint").trim(),
        },
        transport: {
          network: getValue("manual-node-network").trim(),
          path: getValue("manual-node-path").trim(),
          host: getValue("manual-node-host").trim(),
        },
        udp: getChecked("manual-node-udp"),
      },
    };
  }

  const response = await fetchJSON("/api/nodes/custom", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });
  if (!response?.ok) {
    showToast(readAPIError(response) || "添加手动节点失败。", true);
    return;
  }

  closeDialog("add-node-dialog");
  resetAddNodeForm();
  await refreshNodesAfterGenerate();
  showToast("手动节点已添加。");
}

function resetAddNodeForm() {
  [
    "custom-node-uri",
    "manual-node-name",
    "manual-node-server",
    "manual-node-port",
    "manual-node-uuid",
    "manual-node-password",
    "manual-node-sni",
    "manual-node-network",
    "manual-node-path",
    "manual-node-host",
    "manual-node-client-fingerprint",
  ].forEach((id) => setValue(id, ""));
  setChecked("manual-node-udp", true);
}

async function loadYamlPreview() {
  try {
    const previewUrl = state.generatedUrl
      ? `${state.generatedUrl}${state.generatedUrl.includes("?") ? "&" : "?"}_ts=${Date.now()}`
      : withWorkspace(`/api/preview-yaml?_ts=${Date.now()}`);
    const response = await fetch(previewUrl);
    if (!response.ok) {
      state.yamlPreview = `加载失败：HTTP ${response.status}`;
      state.generateStatus =
        state.generateStatus === "generating" ? "error" : state.generateStatus;
      state.lastError = `YAML 预览加载失败：HTTP ${response.status}`;
      renderYamlViewer();
      renderResult();
      return;
    }
    state.yamlPreview = await response.text();
    const generatedAt = response.headers.get("X-SubConv-Generated-At");
    const nodeCount = response.headers.get("X-SubConv-Node-Count");
    const refreshStatus = response.headers.get("X-SubConv-Refresh-Status");
    if (generatedAt) {
      state.lastGeneratedAt = formatChinaDateTime(generatedAt, {
        fallback: generatedAt,
      });
    }
    if (nodeCount) {
      state.resultNodeCount = Number.parseInt(nodeCount, 10) || 0;
    }
    if (refreshStatus) {
      state.refreshStatus = refreshStatus;
    }
    state.resultRuleCount = countRulesInYAML(state.yamlPreview);
    if (state.generatedUrl) {
      state.generateStatus = "success";
    }
    renderYamlViewer();
    renderResult();
  } catch (error) {
    state.yamlPreview = `加载失败：${error.message}`;
    state.generateStatus =
      state.generateStatus === "generating" ? "error" : state.generateStatus;
    state.lastError = `YAML 预览加载失败：${error.message}`;
    renderYamlViewer();
    renderResult();
  }
}

function renderYamlViewer() {
  const viewer = document.getElementById("yaml-viewer");
  const gutter = document.getElementById("yaml-line-numbers");
  const code = document.getElementById("yaml-code-content");
  const visibleText = limitLines(
    state.yamlPreview || "",
    YAML_PREVIEW_LINE_LIMIT,
  );
  const lines = String(visibleText || "").split("\n");
  gutter.textContent = lines.map((_, index) => String(index + 1)).join("\n");
  viewer.classList.toggle("wrapped", state.yamlWrap);

  const query = state.yamlSearch.trim();
  let matchCount = 0;
  const html = lines
    .map((line) => {
      const highlighted = renderYAMLLine(line, query);
      matchCount += highlighted.count;
      return highlighted.html;
    })
    .join("\n");

  code.innerHTML = html;
  renderYamlPreviewMeta(lines.length, countYAMLLines(state.yamlPreview), {
    query,
    matchCount,
  });
}

function renderYamlPreviewMeta(previewLines, totalLines, options = {}) {
  const meta = document.getElementById("yaml-search-meta");
  if (!meta) return;
  const hasYaml = Boolean(String(state.yamlPreview || "").trim());
  if (!hasYaml) {
    meta.textContent = `未生成 · 当前仅显示前 ${YAML_PREVIEW_LINE_LIMIT} 行，可复制完整 YAML 或通过订阅链接使用。`;
    return;
  }
  const visibleLines = Math.min(previewLines || 0, YAML_PREVIEW_LINE_LIMIT);
  const lineText =
    totalLines > visibleLines
      ? `显示前 ${visibleLines} 行 / 完整 ${totalLines} 行`
      : `显示 ${visibleLines} 行 / 完整 ${totalLines} 行`;
  const nodeCount = Number.parseInt(state.resultNodeCount, 10) || 0;
  const ruleCount = Number.parseInt(state.resultRuleCount, 10) || 0;
  const parts = [
    "已生成",
    lineText,
    `${nodeCount} 个节点`,
    `${ruleCount} 条规则`,
  ];
  if (options.query) {
    parts.push(`匹配 ${options.matchCount || 0} 处`);
  }
  parts.push("可复制完整 YAML 或通过订阅链接使用。");
  meta.textContent = parts.join(" · ");
}

async function copyYAML() {
  const content = state.yamlPreview || "";
  if (!content) {
    showToast("当前没有可复制的 YAML 内容。", true);
    return;
  }
  await copyText(content);
  showToast("已复制完整 YAML。");
}

async function loadStatus() {
  const response = await fetchJSON("/api/status");
  state.backendOnline = Boolean(response && !response.error);
  state.statusPayload = response || {};
  state.refreshStage = state.statusPayload?.refresh_stage || "";
  if (state.statusPayload?.last_error && state.generateStatus !== "success") {
    state.lastError = state.statusPayload.last_error;
  }
  if (state.statusPayload?.node_count && !state.resultNodeCount) {
    state.resultNodeCount = state.statusPayload.node_count;
  }
  if (state.statusPayload?.yaml_exists && !state.generatedUrl) {
    if (state.generatedUrl) state.generateStatus = "success";
    if (state.statusPayload?.yaml_updated_at) {
      state.lastGeneratedAt = formatChinaDateTime(
        state.statusPayload.yaml_updated_at,
        { fallback: state.statusPayload.yaml_updated_at },
      );
    }
  }
  const badge = document.getElementById("backend-badge");
  badge.className = `status-badge ${state.backendOnline ? "online" : "offline"}`;
  badge.innerHTML = `${icon("server", `status-icon ${state.backendOnline ? "success" : "danger"}`)}<span class="dot"></span><span>${state.backendOnline ? "Backend Online" : "Backend Offline"}</span>`;
  updateSummary();
  renderResult();
  renderDiagnostics();
}

async function loadLogs() {
  const response = await fetchJSON("/api/logs?tail=200");
  state.logsText = response?.ok
    ? (response.lines || []).join("\n")
    : JSON.stringify(response || {}, null, 2);
  state.logsDisplay = state.logsText;
  renderLogsDisplay();
  renderDiagnostics();
}

async function loadAudit() {
  const response = await fetchJSON("/api/audit");
  state.auditPayload = response?.ok ? response : null;
  renderDiagnostics();
}

function renderLogsDisplay() {
  document.getElementById("logs-output").textContent =
    state.logsDisplay || "日志显示已清空。";
}

async function validateNodes() {
  const response = await fetchJSON("/api/nodes/validate", { method: "POST" });
  state.nodeValidationWarnings = response?.ok ? response.warnings || [] : [];
  renderDiagnostics();
  return state.nodeValidationWarnings;
}

async function refreshDiagnostics() {
  await Promise.all([loadStatus(), loadLogs(), validateNodes(), loadAudit()]);
}

function renderDiagnostics() {
  const logDiagnostics = extractLogDiagnostics(state.logsText);
  const missingFieldCount = state.nodeValidationWarnings.filter((item) =>
    String(item.message || "").includes("缺少"),
  ).length;
  const excludedNodes = state.auditPayload?.excluded_nodes || [];
  const reasonCount = (reason) =>
    excludedNodes.filter((item) => item.reason === reason).length;
  const leakWarnings = (state.auditPayload?.warnings || []).filter(
    (item) =>
      String(item.code || "").includes("leak") ||
      String(item.message || "").includes("leak"),
  );
  const counters = [
    ["原始节点", state.auditPayload?.raw_count ?? "-"],
    ["最终节点", state.auditPayload?.final_count ?? "-"],
    ["过滤节点", state.auditPayload?.excluded_count ?? "-"],
    ["信息节点", reasonCount("info_node")],
    ["禁用节点", reasonCount("disabled_node")],
    ["删除节点", reasonCount("deleted_node")],
    [
      "关键词过滤",
      reasonCount("exclude_keyword_matched") +
        reasonCount("include_keyword_not_matched"),
    ],
    ["重复节点", reasonCount("duplicate")],
  ];

  document.getElementById("diagnostic-counters").innerHTML = counters
    .map(
      ([label, value]) =>
        `<span class="stat-pill"><strong>${escapeHtml(String(value))}</strong><span>${escapeHtml(label)}</span></span>`,
    )
    .join("");

  document.getElementById("diagnostic-status-summary").innerHTML = `
    <div class="summary-item"><span class="summary-key">运行状态</span><span class="summary-value">${escapeHtml(state.backendOnline ? "在线" : "离线")}</span></div>
    <div class="summary-item"><span class="summary-key">上次刷新</span><span class="summary-value">${escapeHtml(state.statusPayload?.last_refresh_at || "-")}</span></div>
    <div class="summary-item"><span class="summary-key">输出文件</span><span class="summary-value">${escapeHtml(state.statusPayload?.output_path || "-")}</span></div>
    <div class="summary-item"><span class="summary-key">泄露检测</span><span class="summary-value">${escapeHtml(leakWarnings.length ? "发现问题" : "未发现节点泄露")}</span></div>
  `;

  const diagnosticItems = [];
  leakWarnings.forEach((warning) => {
    diagnosticItems.push({
      title: "完整性检查",
      detail: warning.message || "发现节点泄露风险",
    });
  });
  excludedNodes.slice(0, 20).forEach((item) => {
    diagnosticItems.push({
      title: `${item.name || "-"} · ${item.reason || "-"}`,
      detail: `${item.source?.name || "-"} · ${item.source?.kind || "-"}`,
    });
  });
  state.nodeValidationWarnings.slice(0, 40).forEach((warning) => {
    diagnosticItems.push({
      title: warning.node_id
        ? `节点 ${shortNodeID(warning.node_id)}`
        : "节点校验",
      detail: warning.message || "",
    });
  });
  logDiagnostics.items.forEach((item) => diagnosticItems.push(item));

  document.getElementById("diagnostic-warning-list").innerHTML =
    diagnosticItems.length
      ? diagnosticItems
          .map(
            (item) =>
              `<div class="diagnostic-item"><strong>${escapeHtml(item.title)}</strong><div class="meta-text">${escapeHtml(item.detail)}</div></div>`,
          )
          .join("")
      : '<div class="diagnostic-item"><strong>诊断正常</strong><div class="meta-text">当前未发现新的生成诊断问题。</div></div>';
}

function extractLogDiagnostics(logText) {
  const lines = String(logText || "").split("\n");
  const diagnostics = {
    duplicateGroups: 0,
    missingRuleProviders: 0,
    items: [],
  };
  lines.forEach((line) => {
    const lower = line.toLowerCase();
    if (
      lower.includes("duplicate proxy-group name") ||
      lower.includes("duplicate group")
    ) {
      diagnostics.duplicateGroups += 1;
      diagnostics.items.push({ title: "重复代理组", detail: line });
    }
    if (
      lower.includes("rule provider") &&
      (lower.includes("missing") || lower.includes("not found"))
    ) {
      diagnostics.missingRuleProviders += 1;
      diagnostics.items.push({ title: "缺失规则提供器", detail: line });
    }
  });
  diagnostics.items = diagnostics.items.slice(0, 20);
  return diagnostics;
}

async function copyLogs() {
  await copyText(state.logsDisplay || state.logsText || "");
  showToast("日志已复制。");
}

function clearLogsDisplay() {
  state.logsDisplay = "";
  renderLogsDisplay();
}

async function refreshNodesAfterGenerate() {
  await loadNodes();
  await validateNodes();
}

function updateGeneratedUrlPlaceholder() {
  if (state.generateStatus === "success" && state.generatedUrl) {
    setValue("generated-url", state.generatedUrl);
    return;
  }
  setValue("generated-url", "");
  const input = document.getElementById("generated-url");
  if (input) {
    input.placeholder = "点击右侧按钮生成订阅链接";
  }
}

function countRulesInYAML(text) {
  const lines = String(text || "").split("\n");
  let inRules = false;
  let count = 0;
  for (const line of lines) {
    if (/^rules:\s*$/.test(line.trim())) {
      inRules = true;
      continue;
    }
    if (!inRules) continue;
    if (/^[^\s-]/.test(line)) break;
    if (/^\s*-\s+/.test(line)) {
      count += 1;
    } else if (/^\S/.test(line.trim())) {
      break;
    }
  }
  return count;
}

function countYAMLLines(text) {
  if (!text) return 0;
  return String(text).split("\n").length;
}

function buildSubscriptionURL() {
  return state.published?.url || state.generatedUrl || "";
}

function openUrl(url) {
  if (!url) return;
  window.open(url, "_blank", "noopener,noreferrer");
}

function limitLines(text, limit) {
  const lines = String(text || "").split("\n");
  if (lines.length <= limit) return text;
  return `${lines.slice(0, limit).join("\n")}\n\n... 已截断，完整内容请下载使用。`;
}

function renderYAMLLine(line, query) {
  const indent = line.match(/^\s*/)?.[0] || "";
  const rest = line.slice(indent.length);

  if (!rest) {
    return { html: escapeHtml(line), count: 0 };
  }

  if (rest.startsWith("#")) {
    return renderHighlightedToken(indent + rest, "yaml-comment", query);
  }

  if (rest.startsWith("- ")) {
    const marker = renderHighlightedToken(indent, "", query);
    const dash = renderHighlightedToken("- ", "yaml-punc", query);
    const body = renderInlineYAML(rest.slice(2), query);
    return {
      html: `${marker.html}${dash.html}${body.html}`,
      count: marker.count + dash.count + body.count,
    };
  }

  const colonIndex = findTopLevelColon(rest);
  if (colonIndex >= 0) {
    const keyPart = rest.slice(0, colonIndex);
    const valuePart = rest.slice(colonIndex + 1);
    const indentToken = renderHighlightedToken(indent, "", query);
    const keyToken = renderHighlightedToken(keyPart, "yaml-key", query);
    const colonToken = renderHighlightedToken(":", "yaml-punc", query);
    const valueToken = renderInlineYAML(valuePart, query);
    return {
      html: `${indentToken.html}${keyToken.html}${colonToken.html}${valueToken.html}`,
      count:
        indentToken.count +
        keyToken.count +
        colonToken.count +
        valueToken.count,
    };
  }

  return renderInlineYAML(line, query);
}

function findTopLevelColon(text) {
  let quote = "";
  let depth = 0;
  for (let i = 0; i < text.length; i += 1) {
    const char = text[i];
    const prev = i > 0 ? text[i - 1] : "";
    if (quote) {
      if (char === quote && prev !== "\\") {
        quote = "";
      }
      continue;
    }
    if (char === '"' || char === "'") {
      quote = char;
      continue;
    }
    if (char === "[" || char === "{") {
      depth += 1;
      continue;
    }
    if ((char === "]" || char === "}") && depth > 0) {
      depth -= 1;
      continue;
    }
    if (char === ":" && depth === 0) {
      return i;
    }
  }
  return -1;
}

function renderInlineYAML(text, query) {
  let html = "";
  let count = 0;
  let index = 0;

  while (index < text.length) {
    const char = text[index];

    if (char === '"' || char === "'") {
      const end = findQuotedEnd(text, index, char);
      const token = renderHighlightedToken(
        text.slice(index, end),
        "yaml-string",
        query,
      );
      html += token.html;
      count += token.count;
      index = end;
      continue;
    }

    if ("{}[],:".includes(char)) {
      const token = renderHighlightedToken(char, "yaml-punc", query);
      html += token.html;
      count += token.count;
      index += 1;
      continue;
    }

    if (char === "#") {
      const token = renderHighlightedToken(
        text.slice(index),
        "yaml-comment",
        query,
      );
      html += token.html;
      count += token.count;
      break;
    }

    if (/\s/.test(char)) {
      const start = index;
      while (index < text.length && /\s/.test(text[index])) {
        index += 1;
      }
      const token = renderHighlightedToken(text.slice(start, index), "", query);
      html += token.html;
      count += token.count;
      continue;
    }

    const start = index;
    while (index < text.length && !/[\s{}\[\],:#]/.test(text[index])) {
      index += 1;
    }

    const raw = text.slice(start, index);
    const nextNonSpace = nextNonSpaceChar(text, index);
    let className = "";

    if (nextNonSpace === ":") {
      className = "yaml-key";
    } else if (/^(true|false|yes|no|on|off|null|~)$/i.test(raw)) {
      className = "yaml-bool";
    } else if (/^-?\d+(\.\d+)?$/.test(raw)) {
      className = "yaml-number";
    } else if (/^(https?:\/\/|\.\/|\/)/.test(raw)) {
      className = "yaml-string";
    }

    const token = renderHighlightedToken(raw, className, query);
    html += token.html;
    count += token.count;
  }

  return { html, count };
}

function findQuotedEnd(text, start, quoteChar) {
  let index = start + 1;
  while (index < text.length) {
    if (text[index] === quoteChar && text[index - 1] !== "\\") {
      return index + 1;
    }
    index += 1;
  }
  return text.length;
}

function nextNonSpaceChar(text, index) {
  let cursor = index;
  while (cursor < text.length && /\s/.test(text[cursor])) {
    cursor += 1;
  }
  return text[cursor] || "";
}

function renderHighlightedToken(text, className, query) {
  if (!query) {
    const html = escapeHtml(text);
    return {
      html: className ? `<span class="${className}">${html}</span>` : html,
      count: 0,
    };
  }
  const escapedQuery = escapeRegExp(query);
  const regex = new RegExp(`(${escapedQuery})`, "gi");
  const parts = String(text).split(regex);
  let count = 0;
  const html = parts
    .map((part) => {
      if (!part) return "";
      if (part.toLowerCase() === query.toLowerCase()) {
        count += 1;
        return `<mark>${escapeHtml(part)}</mark>`;
      }
      return escapeHtml(part);
    })
    .join("");
  return {
    html: className ? `<span class="${className}">${html}</span>` : html,
    count,
  };
}

async function fetchJSON(url, options) {
  try {
    const response = await fetch(withWorkspace(url), options);
    return await response.json();
  } catch (error) {
    return { ok: false, error: { message: error.message || "request failed" } };
  }
}

function withWorkspace(rawUrl) {
  const url = String(rawUrl || "");
  if (!state.workspaceId) return url;
  if (!url.startsWith("/api/")) return url;
  if (
    url.startsWith("/api/workspaces") ||
    url.startsWith("/api/site-logo") ||
    url.startsWith("/api/parse")
  )
    return url;
  const separator = url.includes("?") ? "&" : "?";
  return `${url}${separator}workspace=${encodeURIComponent(state.workspaceId)}`;
}

function setValue(id, value) {
  const element = document.getElementById(id);
  if (element) element.value = value ?? "";
}

function getValue(id) {
  const element = document.getElementById(id);
  return element ? element.value || "" : "";
}

function setChecked(id, value) {
  const element = document.getElementById(id);
  if (element) element.checked = Boolean(value);
}

function getChecked(id) {
  return Boolean(document.getElementById(id)?.checked);
}

function normalizeBackendOrigin(value) {
  return String(value || "")
    .trim()
    .replace(/\/+$/, "");
}

function deepClone(value) {
  return JSON.parse(JSON.stringify(value));
}

function readAPIError(response) {
  return response?.error?.message || "";
}

async function copyText(value) {
  if (navigator.clipboard && window.isSecureContext) {
    await navigator.clipboard.writeText(value);
    return;
  }
  const textarea = document.createElement("textarea");
  textarea.value = value;
  textarea.setAttribute("readonly", "readonly");
  textarea.style.position = "fixed";
  textarea.style.top = "-9999px";
  document.body.appendChild(textarea);
  textarea.select();
  try {
    document.execCommand("copy");
  } finally {
    document.body.removeChild(textarea);
  }
}

function readFileAsDataURL(file) {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = () => resolve(String(reader.result || ""));
    reader.onerror = () =>
      reject(reader.error || new Error("read file failed"));
    reader.readAsDataURL(file);
  });
}

function escapeHtml(value) {
  return String(value)
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#39;");
}

function showToast(message, isError = false) {
  const toast = document.getElementById("toast");
  toast.textContent = message;
  toast.classList.remove("hidden", "error");
  if (isError) toast.classList.add("error");
  window.clearTimeout(showToast.timer);
  showToast.timer = window.setTimeout(
    () => toast.classList.add("hidden"),
    2600,
  );
}

function parseCSVInput(value) {
  return String(value || "")
    .split(",")
    .map((item) => item.trim())
    .filter(Boolean);
}

function applyRawField(raw, key, value) {
  const trimmed = String(value || "").trim();
  if (trimmed) raw[key] = trimmed;
  else delete raw[key];
}

function regionLabel(code) {
  return (
    {
      HK: "香港",
      JP: "日本",
      US: "美国",
      SG: "新加坡",
      TW: "台湾",
      KR: "韩国",
      GB: "英国",
      DE: "德国",
      NL: "荷兰",
      RU: "俄罗斯",
      OTHER: "其它",
    }[String(code || "OTHER").toUpperCase()] || "其它"
  );
}

function regionEmoji(code) {
  return (
    {
      HK: "🇭🇰",
      JP: "🇯🇵",
      US: "🇺🇸",
      SG: "🇸🇬",
      TW: "🇹🇼",
      KR: "🇰🇷",
      GB: "🇬🇧",
      DE: "🇩🇪",
      NL: "🇳🇱",
      RU: "🇷🇺",
    }[String(code || "").toUpperCase()] || ""
  );
}

function shortNodeID(id) {
  return String(id || "").slice(0, 8);
}

function uniqueOrdered(values) {
  const seen = new Set();
  const out = [];
  values.forEach((value) => {
    if (!value || seen.has(value)) return;
    seen.add(value);
    out.push(value);
  });
  return out;
}

function escapeRegExp(value) {
  return String(value || "").replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
}
