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
  { key: "onedrive", label: "微软云盘" },
  { key: "microsoft", label: "微软服务" },
  { key: "ai", label: "AI 服务" },
  { key: "ai_cn", label: "国内 AI 直连" },
  { key: "ai_extended", label: "扩展 AI 服务" },
  { key: "aethersailor_direct", label: "上游专项规则" },
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
  { key: "communication", label: "即时通讯" },
  { key: "containers", label: "容器服务" },
  { key: "emby", label: "Emby" },
  { key: "entertainment", label: "国外媒体" },
  { key: "game_direct", label: "游戏 CDN 直连" },
  { key: "game_platforms", label: "游戏平台" },
  { key: "google_fcm", label: "谷歌 FCM" },
  { key: "github", label: "Github" },
  { key: "gfw", label: "GFW 代理" },
  { key: "non_cn", label: "非中国" },
  { key: "nonstandard_port", label: "非标端口" },
  { key: "pt", label: "PT 站点" },
  { key: "social_media", label: "海外社交媒体" },
  { key: "speedtest", label: "测速工具" },
];

const RULE_PRESETS = {
  minimal: ["private", "domestic", "non_cn"],
  balanced: [
    "private",
    "domestic",
    "onedrive",
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
    publicHidden: true,
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
    key: "enable_region_groups",
    id: "opt-region-groups",
    title: "生成国家代理组",
    desc: "按国家生成两层代理组：国家选择组和国家自动测速组",
    default: false,
    icon: "globe",
    groupOption: true,
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
    title: "代理组",
    desc: "控制策略组展开方式和国家分组。",
    items: ["enable_region_groups"],
  },
  {
    title: "节点处理",
    desc: "在生成前对节点进行过滤与排序。",
    items: ["sort_nodes", "filter_illegal"],
  },
];

const svgIcon = (body) =>
  `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round">${body}</svg>`;

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
  logout: svgIcon(
    '<path d="M10 17l5-5-5-5"></path><path d="M15 12H3"></path><path d="M15 4h4a2 2 0 0 1 2 2v12a2 2 0 0 1-2 2h-4"></path>',
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
  onedrive: "cloud",
  microsoft: "server",
  ai: "spark",
  ai_cn: "spark",
  ai_extended: "spark",
  aethersailor_direct: "tag",
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
  communication: "send",
  containers: "server",
  emby: "film",
  entertainment: "film",
  game_direct: "gamepad",
  game_platforms: "gamepad",
  google_fcm: "send",
  github: "code",
  gfw: "shield",
  non_cn: "globe",
  nonstandard_port: "server",
  pt: "tag",
  social_media: "globe",
  speedtest: "refresh",
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
  mixed_port: 7897,
  allow_lan: true,
  mode: "rule",
  log_level: "info",
  ipv6: false,
  dns_enabled: true,
  custom_dns: false,
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
  group_options: {
    enable_region_groups: false,
    rule_group_node_mode: "full",
    include_real_nodes_in_rule_groups: true,
    special_groups_use_compact: true,
  },
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
  unified_delay: true,
  tcp_concurrent: false,
  find_process_mode: "strict",
  global_client_fingerprint: "chrome",
  dns: {
    enable: true,
    listen: "127.0.0.1:5335",
    use_hosts: true,
    use_system_hosts: false,
    respect_rules: false,
    enhanced_mode: "fake-ip",
    fake_ip_range: "198.18.0.1/16",
    default_nameserver: ["223.5.5.5", "119.29.29.29"],
    nameserver: ["223.5.5.5", "119.29.29.29"],
    fallback: ["https://dns.alidns.com/dns-query"],
    fake_ip_filter: [
      "*.lan",
      "*.local",
      "localhost",
      "*.msftconnecttest.com",
      "*.msftncsi.com",
      "*.nintendo.net",
      "*.playstation.net",
      "*.xboxlive.com",
      "stun.*",
      "time.*",
    ],
  },
  profile: {
    store_selected: true,
    store_fake_ip: false,
  },
  sniffer: {
    enable: true,
    parse_pure_ip: false,
    http: {
      ports: ["80", "8080-8880"],
      override_destination: true,
    },
    tls: {
      ports: ["443", "8443"],
    },
  },
  external_config: {
    template_key: "none",
    template_label: "跟随当前服务模板",
    custom_url: "",
  },
  rule_mode: "balanced",
  enabled_rules: RULE_PRESETS.balanced,
  custom_rules: [],
  additional_rules: [],
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
const LOCAL_DRAFTS_STORAGE_KEY = "SUBCONV_LOCAL_DRAFTS";
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
const WORKSPACE_TITLES = {
  config: "配置生成",
  nodes: "节点编辑",
  yaml: "YAML 预览",
  diagnostics: "诊断日志",
};
const WORKSPACE_KEYS = Object.keys(WORKSPACE_TITLES);

const state = {
  authRequired: false,
  publicConverter: false,
  csrfToken: "",
  config: null,
  activeWorkspace: "config",
  activeSourceMode: "rules",
  ruleMode: "balanced",
  enabledRules: new Set(RULE_PRESETS.balanced),
  customRules: [],
  additionalRules: [],
  externalConfig: { ...DEFAULT_RENDER.external_config },
  generatedUrl: "",
  workspaceId: "",
  workspaceExpiresAt: "",
  draftMode: "privacy",
  hasLocalDraft: false,
  activeLocalDraftId: "",
  localDraftMeta: null,
  localDraftPayload: null,
  generateStatus: "idle",
  generationInFlight: false,
  refreshStatus: "idle",
  refreshStage: "",
  generateProgressTimer: null,
  lastError: "",
  publishNotice: "",
  resultNodeCount: 0,
  resultRuleCount: 0,
  yamlPreview: "",
  yamlError: "",
  yamlRequest: null,
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
  nodeById: new Map(),
  selectedNodeCache: new Map(),
  nodeCacheStale: false,
  nodeRequest: null,
  nodeSummary: { total: 0, enabled: 0, disabled: 0, modified: 0, warnings: 0 },
  nodeFetchNotices: [],
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
  updateCheck: {
    checking: false,
    updateAvailable: false,
    releaseUrl: "",
  },
  editingNode: null,
  nodeValidationWarnings: [],
  addNodeMode: "uri",
  ruleChipsExpanded: false,
  activeRuleSubtab: "builtin",
  activeNodeDialogTab: "basic",
  activeEmojiPopover: null,
  shortcutChordPending: false,
  shortcutChordTimer: null,
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
    ["apply-restore-published-btn", "import"],
    ["save-current-draft-copy-btn", "save"],
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
    ["review-error-source-btn", "settings"],
    ["open-error-diagnostics-btn", "terminal"],
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
  if (!(await loadAuthSession())) return;
  setValue("backend-origin", window.location.origin);
  updateGeneratedUrlPlaceholder();
  renderResult(false);
  renderYamlViewer();
  renderDiagnostics();
  void refreshVersionStatus();
  await initializeWorkspaceSession();
}

function bindEvents() {
  document.querySelectorAll("[data-workspace]").forEach((button) => {
    button.addEventListener("click", () =>
      switchWorkspace(button.dataset.workspace),
    );
  });

  document
    .getElementById("config-import-input")
    ?.addEventListener("change", (event) => {
      const input = event.currentTarget;
      const file =
        input instanceof HTMLInputElement ? input.files?.[0] : null;
      importLocalConfigFile(file);
      input.value = "";
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
    .addEventListener("click", () => openUrl(buildBrowserViewURL(state.generatedUrl)));
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
    .getElementById("review-error-source-btn")
    .addEventListener("click", focusSubscriptionSources);
  document
    .getElementById("open-error-diagnostics-btn")
    .addEventListener("click", () => switchWorkspace("diagnostics"));
  document
    .getElementById("action-result-status")
    .addEventListener("click", focusGenerationError);
  document
    .getElementById("action-result-status")
    .addEventListener("keydown", (event) => {
      if (event.key !== "Enter" && event.key !== " ") return;
      event.preventDefault();
      focusGenerationError();
    });
  document.getElementById("version-badge").addEventListener("click", () => {
    if (state.updateCheck.updateAvailable && state.updateCheck.releaseUrl) {
      openUrl(state.updateCheck.releaseUrl);
    }
  });
  document.getElementById("logout-btn").addEventListener("click", logoutManagementSession);

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
  document
    .getElementById("apply-restore-published-btn")
    .addEventListener("click", restoreWorkspaceFromPublishedLink);
  document
    .getElementById("restore-published-url-input")
    .addEventListener("keydown", (event) => {
      if (event.key !== "Enter") return;
      event.preventDefault();
      restoreWorkspaceFromPublishedLink();
    });
  document
    .getElementById("draft-manager-list")
    .addEventListener("click", handleDraftManagerAction);
  document
    .getElementById("save-current-draft-copy-btn")
    .addEventListener("click", () => persistLocalDraft(false));
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
    .getElementById("additional-rules-input")
    .addEventListener("input", updateAdditionalRulesFromInput);
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
    .getElementById("custom-rule-key")
    .addEventListener("blur", normalizeCustomRuleKeyField);
  document
    .getElementById("custom-rule-label")
    .addEventListener("input", handleCustomRuleLabelInput);
  document
    .getElementById("custom-rule-emoji")
    .addEventListener("input", renderCustomRulePreview);
  document
    .getElementById("custom-rule-target-mode")
    .addEventListener("change", () => {
      if (getValue("custom-rule-target-mode") !== "existing_group") {
        setValue("custom-rule-target-group", "");
      }
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
      syncCustomRuleCompatibility();
      renderCustomRulePreview();
    });
  document
    .getElementById("custom-rule-behavior")
    .addEventListener("change", () => {
      syncCustomRuleFormVisibility();
      syncCustomRuleCompatibility();
      renderCustomRulePreview();
    });
  document
    .getElementById("custom-rule-format")
    .addEventListener("change", () => {
      syncCustomRuleFormVisibility();
      syncCustomRuleCompatibility();
      renderCustomRulePreview();
    });
  document
    .getElementById("custom-rule-url")
    .addEventListener("input", () => {
      syncCustomRuleCompatibility();
      renderCustomRulePreview();
    });
  document
    .getElementById("custom-rule-path")
    .addEventListener("input", () => {
      syncCustomRuleCompatibility();
      renderCustomRulePreview();
    });
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
    .addEventListener("input", () => {
      syncCustomRuleCompatibility();
      renderCustomRulePreview();
    });
  document
    .getElementById("normalize-custom-rule-payload-btn")
    .addEventListener("click", normalizeCustomRulePayloadInput);
  document
    .getElementById("clear-custom-rule-payload-btn")
    .addEventListener("click", clearCustomRulePayloadInput);

  document
    .getElementById("open-node-editor-btn")
    .addEventListener("click", () => switchWorkspace("nodes"));
  document
    .getElementById("refresh-nodes-btn")
    .addEventListener("click", loadNodes);
  document
    .getElementById("restore-deleted-nodes-btn")
    .addEventListener("click", openDeletedNodesDialog);
  document
    .getElementById("restore-all-deleted-nodes-btn")
    .addEventListener("click", restoreAllDeletedNodesFromDialog);
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
  document.addEventListener("keydown", handleGlobalKeydown);
  window.addEventListener("resize", () => {
    if (state.activeEmojiPopover) closeEmojiPopover();
    syncYamlGutter();
  });
  window.addEventListener("hashchange", () => {
    const workspace = workspaceFromLocation();
    if (workspace !== state.activeWorkspace) switchWorkspace(workspace, { skipHistory: true });
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

function handleGlobalKeydown(event) {
  if (event.key === "Escape") {
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
      return;
    }
    document.querySelectorAll("dialog[open]").forEach((dialog) => dialog.close());
    return;
  }

  const target = event.target;
  const editing = target instanceof HTMLElement && (
    target.matches("input, textarea, select") || target.isContentEditable
  );

  if ((event.metaKey || event.ctrlKey) && event.key === "Enter") {
    if (state.activeWorkspace === "config" && !state.generationInFlight) {
      event.preventDefault();
      void generateSubscription();
    }
    return;
  }

  if (editing || event.metaKey || event.ctrlKey || event.altKey) return;

  const key = event.key.toLowerCase();
  if (key === "g") {
    state.shortcutChordPending = true;
    window.clearTimeout(state.shortcutChordTimer);
    state.shortcutChordTimer = window.setTimeout(() => {
      state.shortcutChordPending = false;
    }, 900);
    return;
  }

  if (state.shortcutChordPending) {
    const workspace = { c: "config", n: "nodes", y: "yaml", d: "diagnostics" }[key];
    state.shortcutChordPending = false;
    window.clearTimeout(state.shortcutChordTimer);
    if (workspace) {
      event.preventDefault();
      switchWorkspace(workspace);
    }
    return;
  }

  if (key === "/") {
    const search = document.getElementById("node-search");
    if (search && state.activeWorkspace === "nodes") {
      event.preventDefault();
      search.focus();
    }
  }
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
  if (target.id === "custom-dns-enabled") {
    syncCustomDNSFields();
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

function workspaceFromLocation() {
  const hash = String(window.location.hash || "").replace(/^#/, "").trim().toLowerCase();
  if (WORKSPACE_KEYS.includes(hash)) return hash;
  const stored = window.localStorage.getItem("subconv-active-workspace");
  return WORKSPACE_KEYS.includes(stored) ? stored : "config";
}

function renderWorkspaceTabState() {
  document.querySelectorAll(".workspace-tab").forEach((button) => {
    button.setAttribute("aria-current", button.dataset.workspace === state.activeWorkspace ? "page" : "false");
  });
}

function switchWorkspace(workspace, options = {}) {
  if (!WORKSPACE_KEYS.includes(workspace)) workspace = "config";
  state.activeWorkspace = workspace;
  window.localStorage.setItem("subconv-active-workspace", workspace);
  if (!options.skipHistory && window.location.hash !== `#${workspace}`) {
    window.history.replaceState(null, "", `${window.location.pathname}${window.location.search}#${workspace}`);
  }
  document.title = `${WORKSPACE_TITLES[workspace]} · SubConv Next`;
  renderWorkspaceTabState();
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
    if (state.nodeCacheStale || !state.allNodes.length) loadNodes();
  } else if (workspace === "diagnostics") {
    refreshDiagnostics();
  }

  if (window.scrollY > 0) {
    window.scrollTo({ top: 0, behavior: "auto" });
  }
}

function applyDefaultState() {
  cancelWorkspaceViewRequests();
  state.logsText = "";
  state.logsDisplay = "";
  setValue("client-type", "mihomo");
  setValue("output-filename", DEFAULT_RENDER.output_filename);
  setValue("include-keywords", "");
  setValue("exclude-keywords", "");
  setChecked("opt-source-prefix", DEFAULT_RENDER.source_prefix);
  setValue(
    "source-prefix-mode",
    DEFAULT_RENDER.name_options.source_prefix_mode,
  );
  setValue("rule-group-node-mode", DEFAULT_RENDER.group_options.rule_group_node_mode);
  setChecked(
    "opt-region-groups",
    DEFAULT_RENDER.group_options.enable_region_groups,
  );
  syncSourcePrefixModeControl();
  setDNSFormFromConfig(DEFAULT_RENDER);
  setValue("template-select", DEFAULT_RENDER.external_config.template_key);
  setValue("custom-template-url", "");
  setValue("additional-rules-input", "");
  setValue("batch-import-textarea", "");
  setChecked("yaml-wrap-toggle", true);
  setValue("yaml-search", "");
  setValue("node-type-filter", "all");
  setValue("node-region-filter", "ALL");
  setValue("node-status-filter", "all");
  setValue("node-source-filter", "");
  setValue("node-page-size", "25");
  state.activeWorkspace = workspaceFromLocation();
  state.activeSourceMode = "rules";
  state.ruleMode = DEFAULT_RENDER.rule_mode;
  state.enabledRules = new Set(DEFAULT_RENDER.enabled_rules);
  state.customRules = [];
  state.additionalRules = [];
  state.externalConfig = { ...DEFAULT_RENDER.external_config };
  state.generatedUrl = "";
  state.workspaceId = "";
  state.workspaceExpiresAt = "";
  state.draftMode = "privacy";
  state.hasLocalDraft = false;
  state.activeLocalDraftId = "";
  state.localDraftMeta = null;
  state.localDraftPayload = null;
  state.generateStatus = "idle";
  state.generationInFlight = false;
  state.refreshStatus = "idle";
  state.lastError = "";
  state.publishNotice = "";
  state.lastGeneratedAt = "";
  state.resultNodeCount = 0;
  state.resultRuleCount = 0;
  state.yamlPreview = "";
  state.yamlError = "";
  state.yamlSearch = "";
  state.yamlWrap = true;
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
  state.nodes = [];
  state.nodeSummary = { total: 0, enabled: 0, disabled: 0, modified: 0, warnings: 0 };
  state.nodeFetchNotices = [];
  state.nodeSourceOptions = [];
  state.editingNode = null;
  state.nodeValidationWarnings = [];
  state.nodeById = new Map();
  state.selectedNodeCache = new Map();
  state.nodeCacheStale = false;
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
  renderYamlViewer();
  renderConfigNodeSummary();
  renderDiagnostics();
  renderSessionBanner();
  if (state.activeWorkspace !== "config") {
    resetCustomRuleForm();
    switchWorkspace(state.activeWorkspace, { skipHistory: true });
    return;
  }
  resetCustomRuleForm();
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
  const draftCount = loadLocalDrafts().length;
  let title = "隐私会话";
  let description = "配置仅保留在当前页面，刷新后自动清除。";
  let metaLine = "";
  const restorePublishedButton = state.publicConverter
    ? ""
    : '<button id="restore-published-link-btn" class="secondary-button small-button" type="button">从订阅链接恢复</button>';
  let buttons = `
    <button id="save-local-draft-btn" class="ghost-button small-button" type="button">保存草稿</button>
    <button id="export-config-btn" class="ghost-button small-button" type="button">导出配置</button>
    <button id="import-config-btn" class="ghost-button small-button" type="button">导入配置</button>
    ${draftCount ? '<button id="manage-local-drafts-btn" class="ghost-button small-button" type="button">管理草稿</button>' : ""}
    ${restorePublishedButton}
    <button id="clear-session-btn" class="danger-ghost-button small-button" type="button">清空会话</button>
  `;

  if (mode === "draft_detected") {
    title = draftCount > 1 ? `发现 ${draftCount} 个本机草稿` : "发现本机草稿";
    description = draftCount > 1 ? "可恢复最近更新的草稿，或打开管理器选择。" : "这是之前手动保存到本机浏览器的配置，是否恢复？";
    metaLine = formatDraftMetaLine(meta);
    buttons = `
      <button id="restore-draft-btn" class="secondary-button small-button" type="button">${draftCount > 1 ? "恢复最近草稿" : "恢复草稿"}</button>
      <button id="manage-local-drafts-btn" class="ghost-button small-button" type="button">管理草稿</button>
      <button id="import-config-btn" class="ghost-button small-button" type="button">导入配置</button>
      ${state.publicConverter ? "" : '<button id="restore-published-link-btn" class="ghost-button small-button" type="button">从订阅链接恢复</button>'}
      <button id="discard-draft-btn" class="ghost-button small-button" type="button">删除当前草稿</button>
    `;
  } else if (mode === "local_draft") {
    title = "当前模式：本机草稿";
    description = "此配置已保存在当前浏览器中，请勿在公共设备上使用。";
    metaLine = formatDraftMetaLine(meta);
    buttons = `
      <button id="update-local-draft-btn" class="secondary-button small-button" type="button">更新本机草稿</button>
      <button id="save-local-draft-btn" class="ghost-button small-button" type="button">另存为新草稿</button>
      <button id="export-config-btn" class="ghost-button small-button" type="button">导出配置</button>
      <button id="import-config-btn" class="ghost-button small-button" type="button">导入配置</button>
      <button id="manage-local-drafts-btn" class="ghost-button small-button" type="button">管理草稿</button>
      ${state.publicConverter ? "" : '<button id="restore-published-link-btn" class="ghost-button small-button" type="button">从订阅链接恢复</button>'}
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
    .getElementById("export-config-btn")
    ?.addEventListener("click", exportLocalConfig);
  document
    .getElementById("import-config-btn")
    ?.addEventListener("click", () =>
      document.getElementById("config-import-input")?.click(),
    );
  document
    .getElementById("update-local-draft-btn")
    ?.addEventListener("click", updateLocalDraft);
  document
    .getElementById("manage-local-drafts-btn")
    ?.addEventListener("click", openDraftManagerDialog);
  document
    .getElementById("clear-session-btn")
    ?.addEventListener("click", clearCurrentSession);
  document
    .getElementById("restore-draft-btn")
    ?.addEventListener("click", restoreLocalDraft);
  document
    .getElementById("restore-published-link-btn")
    ?.addEventListener("click", openRestorePublishedDialog);
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
  const draftName = String(meta.draft_name || "").trim();
  const draftCount = Number.parseInt(meta.draft_count, 10) || 0;
  const prefix = draftName ? `${draftName} · ` : "";
  const suffix = draftCount > 1 ? ` · 共 ${draftCount} 个草稿` : "";
  if (savedAt && updatedAt && savedAt !== updatedAt) {
    return `${prefix}首次保存 ${savedAt} · 最近更新 ${updatedAt} · 来源 ${sourceCount} 个${suffix}`;
  }
  if (updatedAt) {
    return `${prefix}最近更新 ${updatedAt} · 来源 ${sourceCount} 个${suffix}`;
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
    activeLocalDraftId: state.activeLocalDraftId,
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
  state.activeLocalDraftId = snapshot.activeLocalDraftId || "";
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
    state.activeLocalDraftId = payload.draft_id || state.activeLocalDraftId || "";
    state.localDraftPayload = deepClone(payload);
    state.localDraftMeta = {
      draft_id: payload.draft_id || "",
      draft_name: payload.draft_name || "",
      saved_at: payload.saved_at || "",
      updated_at: payload.updated_at || payload.saved_at || "",
      draft_count: loadLocalDrafts().length,
      source_count: Number(
        payload.source_count || countDraftSources(payload.config) || 0,
      ),
    };
    return;
  }
  state.hasLocalDraft = false;
  state.activeLocalDraftId = "";
  state.localDraftPayload = null;
  state.localDraftMeta = null;
}

function openDraftManagerDialog() {
  renderDraftManager();
  const dialog = document.getElementById("draft-manager-dialog");
  if (!dialog) return;
  if (!dialog.open) dialog.showModal();
}

function renderDraftManager() {
  const list = document.getElementById("draft-manager-list");
  if (!list) return;
  const drafts = loadLocalDrafts();
  if (!drafts.length) {
    list.innerHTML = `
      <div class="draft-empty-state">
        <strong>暂无本机草稿</strong>
        <span>${state.publicConverter ? "可将当前配置另存为本机草稿。" : "可将当前配置另存为新草稿，或从订阅链接恢复。"}</span>
      </div>
    `;
    return;
  }
  list.innerHTML = drafts
    .map((draft) => {
      const isActive = draft.draft_id === state.activeLocalDraftId;
      const sourceCount = Number.parseInt(draft.source_count, 10) || 0;
      const updatedAt = formatDraftDateTime(draft.updated_at);
      const publishHint = draft.publish_ref?.token_hint
        ? ` · 发布 ${draft.publish_ref.token_hint}`
        : "";
      return `
        <article class="draft-item ${isActive ? "active" : ""}">
          <div class="draft-item-main">
            <input class="draft-name-input" data-draft-name-input="${escapeHtml(draft.draft_id)}" value="${escapeHtml(draft.draft_name || "本机草稿")}" />
            <div class="draft-item-meta">
              ${escapeHtml(updatedAt || "未知时间")} · 来源 ${sourceCount} 个${escapeHtml(publishHint)}
            </div>
          </div>
          <div class="draft-item-actions">
            <button class="ghost-button tiny-button" type="button" data-draft-action="rename" data-draft-id="${escapeHtml(draft.draft_id)}">重命名</button>
            <button class="secondary-button tiny-button" type="button" data-draft-action="restore" data-draft-id="${escapeHtml(draft.draft_id)}">${isActive ? "重新恢复" : "恢复"}</button>
            <button class="danger-ghost-button tiny-button" type="button" data-draft-action="delete" data-draft-id="${escapeHtml(draft.draft_id)}">删除</button>
          </div>
        </article>
      `;
    })
    .join("");
}

function handleDraftManagerAction(event) {
  const button = event.target.closest("[data-draft-action]");
  if (!button) return;
  const draftID = String(button.dataset.draftId || "").trim();
  if (!draftID) return;
  const action = button.dataset.draftAction;
  if (action === "restore") {
    void restoreLocalDraft(draftID);
    return;
  }
  if (action === "delete") {
    deleteLocalDraftFromManager(draftID);
    return;
  }
  if (action === "rename") {
    const input = document.querySelector(
      `[data-draft-name-input="${CSS.escape(draftID)}"]`,
    );
    renameLocalDraft(draftID, input?.value || "");
  }
}

function deleteLocalDraftFromManager(draftID) {
  const deletingActive = draftID === state.activeLocalDraftId;
  const nextDraft = clearLocalDraft(draftID);
  if (deletingActive) {
    applyDraftState(nextDraft ? "draft_detected" : "privacy", nextDraft);
  } else if (state.localDraftMeta) {
    state.localDraftMeta.draft_count = loadLocalDrafts().length;
  }
  renderSessionBanner();
  renderDraftManager();
  showToast("本机草稿已删除。");
}

function renameLocalDraft(draftID, name) {
  const draftName = String(name || "").trim();
  if (!draftName) {
    showToast("草稿名称不能为空。", true);
    return;
  }
  const store = loadDraftStore();
  const draft = store.drafts.find((item) => item.draft_id === draftID);
  if (!draft) {
    showToast("未找到要重命名的草稿。", true);
    return;
  }
  draft.draft_name = draftName;
  saveDraftStore(store);
  if (state.activeLocalDraftId === draftID) {
    state.localDraftPayload = { ...(state.localDraftPayload || {}), draft_name: draftName };
    state.localDraftMeta = { ...(state.localDraftMeta || {}), draft_name: draftName };
    renderSessionBanner();
  }
  renderDraftManager();
  showToast("草稿名称已更新。");
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

async function restoreLocalDraft(id = "") {
  const draft = loadLocalDraft(id);
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
  activateLocalDraft(draft.draft_id);
  applyDraftState("local_draft", draft);
  renderSessionBanner();
  closeDialog("draft-manager-dialog");
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

function openRestorePublishedDialog() {
  if (state.publicConverter) {
    showToast("公网模式不支持从发布链接恢复源配置。", true);
    return;
  }
  setValue("restore-published-url-input", "");
  const dialog = document.getElementById("restore-published-dialog");
  if (!dialog) return;
  if (!dialog.open) dialog.showModal();
  window.requestAnimationFrame(() => {
    document.getElementById("restore-published-url-input")?.focus();
  });
}

async function restoreWorkspaceFromPublishedLink() {
  if (state.publicConverter) {
    showToast("公网模式不支持从发布链接恢复源配置。", true);
    return;
  }
  const publishedURL = getValue("restore-published-url-input").trim();
  if (!publishedURL) {
    showToast("请粘贴本工具生成的完整订阅链接。", true);
    return;
  }

  if (!state.workspaceId) {
    const created = await createNewWorkspace({ preserveDraftState: true });
    if (!created) return;
  }

  const button = document.getElementById("apply-restore-published-btn");
  if (button) {
    button.disabled = true;
    setButtonIconText(button, "恢复中");
  }

  try {
    const response = await fetchJSON(
      `/api/workspaces/${encodeURIComponent(state.workspaceId)}/restore-from-published`,
      {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ url: publishedURL }),
      },
    );
    if (!response?.ok || !response.config) {
      showToast(
        readAPIError(response) || "恢复失败，请确认链接是本工具生成的订阅链接。",
        true,
      );
      return;
    }

    const restoredConfig = response.config;
    const restoredNodeState = sanitizeLocalDraftNodeState(response.node_state);
    state.config = restoredConfig;
    fillFormFromConfig(restoredConfig);

    const existingDraft = findLocalDraftByPublishID(response.publish?.publish_id);
    const draft = buildRestoredPublishedDraftPayload(
      restoredConfig,
      restoredNodeState,
      response.publish,
      existingDraft,
    );
    const savedDraft = saveLocalDraftPayload(draft, {
      draftID: existingDraft?.draft_id || draft.draft_id,
    });
    applyDraftState("local_draft", savedDraft || draft);
    renderSessionBanner();
    applyRestoreDraftPublish(response.publish);
    markNodeCacheStale();
    closeDialog("restore-published-dialog");
    setValue("restore-published-url-input", "");

    await Promise.all([loadStatus(), loadSubscriptionMeta(), loadAudit()]);
    await loadNodes();
    updateSummary();
    renderDiagnostics();
    showToast(
      existingDraft
        ? "已从订阅链接更新已有本机草稿。"
        : "已从订阅链接恢复为新的本机草稿。",
    );
  } finally {
    if (button) {
      button.disabled = false;
      setButtonIconText(button, "恢复到草稿");
    }
  }
}

function buildRestoredPublishedDraftPayload(
  config,
  nodeState,
  published,
  existingDraft = null,
) {
  const now = new Date().toISOString();
  const draftConfig = sanitizeLocalDraftConfig(config);
  const cleanedNodeState = pruneLocalDraftDeletedNodes(nodeState, draftConfig);
  draftConfig.client_type = getValue("client-type") || "mihomo";
  draftConfig.backend_url = normalizeBackendOrigin(
    getValue("backend-origin") || window.location.origin,
  );
  const payload = {
    version: 2,
    draft_id: existingDraft?.draft_id || createLocalDraftID(),
    draft_name:
      preservedLocalDraftName(existingDraft) ||
      defaultLocalDraftName(draftConfig, "订阅链接恢复", { time: now }),
    saved_at: existingDraft?.saved_at || now,
    updated_at: now,
    source_count: countDraftSources(draftConfig),
    config: draftConfig,
  };
  if (cleanedNodeState) {
    payload.node_state = cleanedNodeState;
  }
  const publishRef = draftPublishRefFromPublished(published);
  if (publishRef) {
    payload.publish_ref = publishRef;
  }
  return payload;
}

function findLocalDraftByPublishID(publishID) {
  const targetPublishID = String(publishID || "").trim();
  if (!targetPublishID) return null;
  return (
    loadLocalDrafts().find(
      (draft) =>
        String(draft.publish_ref?.publish_id || "").trim() === targetPublishID,
    ) || null
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
  const nextDraft = clearLocalDraft();
  await deleteCurrentWorkspace();
  const created = await createNewWorkspace({ preserveDraftState: false });
  if (!created) return;
  applyDraftState(nextDraft ? "draft_detected" : "privacy", nextDraft);
  renderSessionBanner();
  await reloadWorkspaceData();
  renderDraftManager();
  showToast(nextDraft ? "已删除当前草稿，可继续恢复其它草稿。" : "已删除本机草稿。");
}

async function exportLocalConfig() {
  if (!state.workspaceId) {
    showToast("当前没有可导出的会话。", true);
    return;
  }
  try {
    const { config, nodeState } = await collectLocalDraftSource();
    const payload = {
      app: "subconv-next",
      kind: "subconv-config",
      version: 2,
      exported_at: new Date().toISOString(),
      config,
    };
    if (nodeState) {
      payload.node_state = nodeState;
    }
    const blob = new Blob([JSON.stringify(payload, null, 2)], {
      type: "application/json",
    });
    const url = URL.createObjectURL(blob);
    const link = document.createElement("a");
    link.href = url;
    link.download = `subconv-config-${new Date().toISOString().slice(0, 10)}.json`;
    document.body.appendChild(link);
    link.click();
    link.remove();
    window.setTimeout(() => URL.revokeObjectURL(url), 1000);
    showToast("配置已导出为 JSON 文件。");
  } catch (error) {
    showToast("导出失败：" + (error?.message || "未知错误"), true);
  }
}

function normalizeImportedConfigPayload(parsed) {
  if (!parsed || typeof parsed !== "object" || Array.isArray(parsed)) {
    return null;
  }
  if (parsed.kind && parsed.kind !== "subconv-config") return null;
  const config =
    parsed.config && typeof parsed.config === "object" ? parsed.config : null;
  if (!config) return null;
  const looksLikeConfig =
    Array.isArray(config.subscriptions) ||
    Array.isArray(config.inline) ||
    (config.render && typeof config.render === "object") ||
    (config.service && typeof config.service === "object");
  if (!looksLikeConfig) return null;
  const now = new Date().toISOString();
  return {
    version: 2,
    draft_id: createLocalDraftID(),
    draft_name: defaultLocalDraftName(config, "导入配置", {
      time: parsed.exported_at || now,
    }),
    saved_at: now,
    updated_at: now,
    source_count: countDraftSources(config),
    config,
    ...(parsed.node_state && typeof parsed.node_state === "object"
      ? { node_state: parsed.node_state }
      : {}),
  };
}

function importLocalConfigFile(file) {
  if (!file) return;
  if (file.size > 4 * 1024 * 1024) {
    showToast("配置文件超过 4 MB，请检查文件内容。", true);
    return;
  }
  const reader = new FileReader();
  reader.onload = async () => {
    let parsed = null;
    try {
      parsed = JSON.parse(String(reader.result || ""));
    } catch (error) {
      showToast("导入失败：文件不是有效的 JSON。", true);
      return;
    }
    const payload = normalizeImportedConfigPayload(parsed);
    if (!payload) {
      showToast("导入失败：文件不是本工具导出的配置。", true);
      return;
    }
    const draft = saveLocalDraftPayload(payload, { activate: true });
    if (!draft?.draft_id) {
      showToast("导入失败：配置内容无效。", true);
      return;
    }
    await restoreLocalDraft(draft.draft_id);
  };
  reader.onerror = () => showToast("导入失败：文件读取错误。", true);
  reader.readAsText(file);
}

function saveLocalDraft() {
  void persistLocalDraft(false);
}

function updateLocalDraft() {
  void persistLocalDraft(true);
}

async function collectLocalDraftSource() {
  const config = sanitizeLocalDraftConfig(buildConfigFromForm());
  config.client_type = getValue("client-type") || "mihomo";
  config.backend_url = normalizeBackendOrigin(
    getValue("backend-origin") || window.location.origin,
  );
  const nodeState = pruneLocalDraftDeletedNodes(
    await loadNodeStateDraft(),
    config,
  );
  return { config, nodeState };
}

async function persistLocalDraft(isUpdate, options = {}) {
  if (!state.workspaceId) {
    showToast("当前没有可保存的会话。", true);
    return;
  }
  const existing = isUpdate ? loadLocalDraft() : null;
  const now = new Date().toISOString();
  const { config, nodeState } = await collectLocalDraftSource();
  const payload = {
    version: 2,
    draft_id:
      isUpdate && existing?.draft_id
        ? existing.draft_id
        : createLocalDraftID(),
    draft_name:
      options.name ||
      (isUpdate ? preservedLocalDraftName(existing) : "") ||
      defaultLocalDraftName(config, "本机草稿", { time: now }),
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
  const savedDraft = saveLocalDraftPayload(payload, {
    draftID: isUpdate ? payload.draft_id : "",
  });
  applyDraftState("local_draft", savedDraft || payload);
  renderSessionBanner();
  renderDraftManager();
  if (!options.silent) {
    showToast(
      isUpdate
        ? "本机草稿已更新。"
        : "已另存为新的本机草稿。草稿会保存当前配置，但不会保存完整私密订阅 token。",
    );
  }
}

async function exitDraftMode() {
  const payload = loadLocalDraft();
  applyDraftState("privacy", payload);
  renderSessionBanner();
  showToast("已退出草稿模式，本机草稿已保留。");
}

async function clearCurrentSession() {
  const currentMode = state.draftMode;
  let nextMode =
    currentMode === "draft_detected" ? "draft_detected" : "privacy";
  let nextDraftPayload = loadLocalDraft();
  if (currentMode === "local_draft" && state.hasLocalDraft) {
    nextMode = nextDraftPayload ? "draft_detected" : "privacy";
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
  if (state.activeWorkspace === "nodes") await loadNodes();
  if (state.activeWorkspace === "yaml") await loadYamlPreview();
}

function loadLocalDraft(id = "") {
  const store = loadDraftStore();
  const draftID = String(id || state.activeLocalDraftId || store.active_id || "").trim();
  let draft = draftID
    ? store.drafts.find((item) => item.draft_id === draftID)
    : null;
  if (!draft) {
    draft = latestLocalDraft(store);
  }
  return draft ? deepClone(draft) : null;
}

function loadLocalDrafts() {
  return sortedLocalDrafts(loadDraftStore().drafts).map((item) =>
    deepClone(item),
  );
}

function loadDraftStore() {
  let store = readDraftStore();
  const legacyDraft = loadLegacyLocalDraft();
  if (legacyDraft && !store.drafts.length) {
    const migrated = normalizeLocalDraftPayload({
      ...legacyDraft,
      draft_id: legacyDraft.draft_id || createLocalDraftID(),
      draft_name:
        legacyDraft.draft_name ||
        defaultLocalDraftName(legacyDraft.config, "迁移草稿"),
    });
    if (migrated) {
      store = {
        version: 1,
        active_id: migrated.draft_id,
        drafts: [migrated],
      };
      saveDraftStore(store);
    }
  }
  if (legacyDraft) {
    localStorage.removeItem(LOCAL_DRAFT_STORAGE_KEY);
  }
  return store;
}

function readDraftStore() {
  const fallback = { version: 1, active_id: "", drafts: [] };
  const raw = localStorage.getItem(LOCAL_DRAFTS_STORAGE_KEY);
  if (!raw) return fallback;
  try {
    const parsed = JSON.parse(raw);
    const drafts = (Array.isArray(parsed?.drafts) ? parsed.drafts : [])
      .map((item) => normalizeLocalDraftPayload(item))
      .filter(Boolean);
    const activeID = String(parsed?.active_id || "").trim();
    return {
      version: 1,
      active_id: drafts.some((item) => item.draft_id === activeID)
        ? activeID
        : latestLocalDraft({ drafts })?.draft_id || "",
      drafts,
    };
  } catch (error) {
    localStorage.removeItem(LOCAL_DRAFTS_STORAGE_KEY);
    return fallback;
  }
}

function loadLegacyLocalDraft() {
  const raw = localStorage.getItem(LOCAL_DRAFT_STORAGE_KEY);
  if (!raw) return null;
  try {
    return normalizeLocalDraftPayload(JSON.parse(raw));
  } catch (error) {
    localStorage.removeItem(LOCAL_DRAFT_STORAGE_KEY);
    return null;
  }
}

function saveDraftStore(store) {
  const drafts = (Array.isArray(store?.drafts) ? store.drafts : [])
    .map((item) => normalizeLocalDraftPayload(item))
    .filter(Boolean);
  const activeID = String(store?.active_id || "").trim();
  const nextStore = {
    version: 1,
    active_id: drafts.some((item) => item.draft_id === activeID)
      ? activeID
      : latestLocalDraft({ drafts })?.draft_id || "",
    drafts: sortedLocalDrafts(drafts),
  };
  localStorage.setItem(LOCAL_DRAFTS_STORAGE_KEY, JSON.stringify(nextStore));
  localStorage.removeItem(LOCAL_DRAFT_STORAGE_KEY);
  return nextStore;
}

function saveLocalDraftPayload(payload, options = {}) {
  const store = loadDraftStore();
  const existingID = String(options.draftID || payload?.draft_id || "").trim();
  const draftID = existingID || createLocalDraftID();
  const existing = store.drafts.find((item) => item.draft_id === draftID);
  const draft = normalizeLocalDraftPayload({
    ...payload,
    draft_id: draftID,
    draft_name:
      options.name ||
      preservedLocalDraftName(payload) ||
      preservedLocalDraftName(existing) ||
      defaultLocalDraftName(payload?.config, "本机草稿", {
        time: payload?.updated_at || payload?.saved_at,
      }),
    saved_at: payload?.saved_at || existing?.saved_at || new Date().toISOString(),
  });
  if (!draft) return null;
  const index = store.drafts.findIndex((item) => item.draft_id === draft.draft_id);
  if (index >= 0) {
    store.drafts[index] = draft;
  } else {
    store.drafts.push(draft);
  }
  if (options.activate !== false) {
    store.active_id = draft.draft_id;
  }
  saveDraftStore(store);
  return deepClone(draft);
}

function clearLocalDraft(id = "") {
  const store = loadDraftStore();
  const draftID = String(id || state.activeLocalDraftId || store.active_id || "").trim();
  if (!draftID) return null;
  const nextDrafts = store.drafts.filter((item) => item.draft_id !== draftID);
  const nextStore = saveDraftStore({
    version: 1,
    active_id:
      store.active_id === draftID
        ? latestLocalDraft({ drafts: nextDrafts })?.draft_id || ""
        : store.active_id,
    drafts: nextDrafts,
  });
  return latestLocalDraft(nextStore);
}

function activateLocalDraft(id) {
  const draftID = String(id || "").trim();
  if (!draftID) return null;
  const store = loadDraftStore();
  const draft = store.drafts.find((item) => item.draft_id === draftID);
  if (!draft) return null;
  saveDraftStore({ ...store, active_id: draftID });
  state.activeLocalDraftId = draftID;
  return deepClone(draft);
}

function removeLocalDraftPublishRef(publishID) {
  const targetPublishID = String(publishID || "").trim();
  const store = loadDraftStore();
  let changed = false;
  const drafts = store.drafts.map((draft) => {
    const draftPublishID = String(draft.publish_ref?.publish_id || "").trim();
    if (!draft.publish_ref || (targetPublishID && draftPublishID !== targetPublishID)) {
      return draft;
    }
    changed = true;
    const next = deepClone(draft);
    delete next.publish_ref;
    next.updated_at = new Date().toISOString();
    return next;
  });
  if (!changed) return;
  saveDraftStore({ ...store, drafts });
  if (
    state.localDraftPayload?.publish_ref &&
    (!targetPublishID ||
      String(state.localDraftPayload.publish_ref.publish_id || "").trim() ===
        targetPublishID)
  ) {
    delete state.localDraftPayload.publish_ref;
  }
}

function normalizeLocalDraftPayload(payload) {
  if (
    !payload ||
    typeof payload !== "object" ||
    !payload.config ||
    typeof payload.config !== "object"
  ) {
    return null;
  }
  const now = new Date().toISOString();
  const draft = deepClone(payload);
  draft.version = Number.parseInt(draft.version, 10) || 2;
  draft.draft_id = String(draft.draft_id || "").trim() || createLocalDraftID();
  draft.draft_name =
    preservedLocalDraftName(draft) ||
    defaultLocalDraftName(draft.config, "本机草稿", {
      time: draft.updated_at || draft.saved_at,
    });
  draft.saved_at = String(draft.saved_at || draft.updated_at || now);
  draft.updated_at = String(draft.updated_at || draft.saved_at || now);
  draft.source_count = Number(
    draft.source_count || countDraftSources(draft.config) || 0,
  );
  return draft;
}

function sortedLocalDrafts(drafts) {
  return [...(Array.isArray(drafts) ? drafts : [])].sort((a, b) => {
    const at = new Date(a?.updated_at || a?.saved_at || 0).getTime() || 0;
    const bt = new Date(b?.updated_at || b?.saved_at || 0).getTime() || 0;
    return bt - at;
  });
}

function latestLocalDraft(store) {
  return sortedLocalDrafts(store?.drafts || [])[0] || null;
}

function createLocalDraftID() {
  if (window.crypto?.randomUUID) {
    return `draft_${window.crypto.randomUUID()}`;
  }
  return `draft_${Date.now().toString(36)}_${Math.random()
    .toString(36)
    .slice(2, 10)}`;
}

function defaultLocalDraftName(config, fallback = "本机草稿", options = {}) {
  const baseName = localDraftSourceLabel(config, fallback);
  if (options.includeTime === false) return baseName;
  const timeLabel = formatDraftNameTime(options.time || new Date());
  return timeLabel ? `${baseName} · ${timeLabel}` : baseName;
}

function preservedLocalDraftName(draft) {
  const name = String(draft?.draft_name || "").trim();
  if (!name || isGenericLocalDraftName(name)) return "";
  return name;
}

function isGenericLocalDraftName(name) {
  const value = String(name || "").trim();
  return (
    !value ||
    /^source-\d+$/i.test(value) ||
    /^manual-\d+$/i.test(value) ||
    /^手动节点\s*\d*$/i.test(value) ||
    ["本机草稿", "订阅链接恢复", "迁移草稿"].includes(value)
  );
}

function localDraftSourceLabel(config, fallback) {
  const subscriptions = Array.isArray(config?.subscriptions)
    ? config.subscriptions
    : [];
  const subscriptionLabels = subscriptions
    .map((item) => localDraftSubscriptionLabel(item))
    .filter(Boolean);
  if (subscriptionLabels.length > 1) {
    return `${subscriptionLabels.slice(0, 2).join(" + ")} 等 ${subscriptionLabels.length} 个订阅源`;
  }
  if (subscriptionLabels.length === 1) {
    const inlineCount = countInlineDraftSources(config);
    return inlineCount > 0
      ? `${subscriptionLabels[0]} + ${inlineCount} 个手动节点`
      : subscriptionLabels[0];
  }

  const inlineEntries = Array.isArray(config?.inline) ? config.inline : [];
  const inlineLabels = inlineEntries
    .map((item) => localDraftInlineLabel(item))
    .filter(Boolean);
  if (inlineLabels.length > 1) {
    return `${inlineLabels[0]} 等 ${inlineLabels.length} 个手动节点`;
  }
  if (inlineLabels.length === 1) {
    return inlineLabels[0];
  }
  return fallback;
}

function localDraftSubscriptionLabel(item) {
  const name = meaningfulDraftLabel(item?.name);
  if (name) return name;
  const host = subscriptionURLHostLabel(item?.url);
  if (host) return host;
  return meaningfulDraftLabel(item?.id);
}

function localDraftInlineLabel(item) {
  return (
    meaningfulDraftLabel(item?.name) ||
    meaningfulDraftLabel(item?.id) ||
    "手动节点草稿"
  );
}

function meaningfulDraftLabel(value) {
  const label = String(value || "").trim();
  return isGenericLocalDraftName(label) ? "" : label;
}

function countInlineDraftSources(config) {
  return Array.isArray(config?.inline)
    ? config.inline.filter((item) => String(item?.content || "").trim()).length
    : 0;
}

function formatDraftNameTime(value) {
  const formatted = formatChinaDateTime(value, { seconds: false, fallback: "" });
  return formatted.length > 5 ? formatted.slice(5) : formatted;
}

function subscriptionURLHostLabel(rawURL) {
  try {
    return new URL(String(rawURL || "")).hostname || "";
  } catch {
    return "";
  }
}

function normalizeDNSStringList(values, fallback = []) {
  const source = Array.isArray(values) ? values : fallback;
  return source
    .map((item) => String(item || "").trim())
    .filter(Boolean);
}

function normalizeNameserverPolicy(policy) {
  if (!policy || typeof policy !== "object" || Array.isArray(policy)) return {};
  const next = {};
  Object.entries(policy).forEach(([key, values]) => {
    const policyKey = String(key || "").trim();
    if (!policyKey) return;
    const resolvers = normalizeDNSStringList(
      Array.isArray(values) ? values : [values],
    );
    if (resolvers.length) next[policyKey] = resolvers;
  });
  return next;
}

function normalizeDNSConfigForFrontend(dns) {
  const base = deepClone(DEFAULT_RENDER.dns);
  const next = {
    ...base,
    ...(dns || {}),
  };
  next.enable = true;
  next.listen = String(next.listen || base.listen).trim() || base.listen;
  next.use_hosts = next.use_hosts !== false;
  next.use_system_hosts = next.use_system_hosts === true;
  next.respect_rules = next.respect_rules === true;
  next.enhanced_mode =
    String(next.enhanced_mode || base.enhanced_mode).trim() || base.enhanced_mode;
  next.fake_ip_range =
    String(next.fake_ip_range || base.fake_ip_range).trim() || base.fake_ip_range;
  next.default_nameserver = normalizeDNSStringList(
    next.default_nameserver,
    base.default_nameserver,
  );
  if (!next.default_nameserver.length) {
    next.default_nameserver = [...base.default_nameserver];
  }
  next.nameserver = normalizeDNSStringList(next.nameserver, base.nameserver);
  if (!next.nameserver.length) {
    next.nameserver = [...base.nameserver];
  }
  next.proxy_server_nameserver = normalizeDNSStringList(
    next.proxy_server_nameserver,
  );
  next.direct_nameserver = normalizeDNSStringList(next.direct_nameserver);
  next.direct_nameserver_follow_policy =
    next.direct_nameserver_follow_policy === true;
  next.fallback = normalizeDNSStringList(next.fallback);
  next.fake_ip_filter = normalizeDNSStringList(
    next.fake_ip_filter,
    base.fake_ip_filter,
  );
  if (!next.fake_ip_filter.length) {
    next.fake_ip_filter = [...base.fake_ip_filter];
  }
  next.nameserver_policy = normalizeNameserverPolicy(next.nameserver_policy);
  return next;
}

function stripDNSListValue(value) {
  return String(value || "")
    .trim()
    .replace(/^-\s*/, "")
    .replace(/^['"]|['"]$/g, "")
    .trim();
}

function parseDNSLines(value) {
  const text = String(value || "").trim();
  if (!text) return [];
  const rawLines =
    text.startsWith("[") && text.endsWith("]")
      ? [text.slice(1, -1)]
      : text.split(/\r?\n/);
  const lines = [];
  rawLines.forEach((line) => {
    splitDNSListItems(line).forEach((item) => lines.push(item));
  });
  return lines;
}

function splitDNSListItems(value) {
  const cleaned = String(value || "").trim().replace(/^\[|\]$/g, "");
  if (!cleaned) return [];
  return cleaned
    .split(",")
    .map(stripDNSListValue)
    .filter(Boolean);
}

function formatDNSLines(values) {
  return normalizeDNSStringList(values).join("\n");
}

function splitNameserverPolicyResolvers(value) {
  return splitDNSListItems(value);
}

function parseNameserverPolicyText(value) {
  const text = String(value || "").trim();
  if (!text) return {};
  if (text.startsWith("{")) {
    try {
      return normalizeNameserverPolicy(JSON.parse(text));
    } catch {
      return {};
    }
  }
  const policy = {};
  text.split(/\r?\n/).forEach((line) => {
    const cleaned = stripDNSListValue(line);
    if (!cleaned) return;
    const match = cleaned.match(/^(.+?)(?:\s*=\s*|:\s+)(.+)$/);
    if (!match) return;
    const key = match[1].trim().replace(/:$/, "");
    const resolvers = splitNameserverPolicyResolvers(match[2]);
    if (key && resolvers.length) policy[key] = resolvers;
  });
  return policy;
}

function formatNameserverPolicyText(policy) {
  const normalized = normalizeNameserverPolicy(policy);
  return Object.entries(normalized)
    .map(([key, values]) => `${key}: ${values.join(", ")}`)
    .join("\n");
}

function setDNSFormFromConfig(render) {
  const custom = render?.custom_dns === true;
  const dns = normalizeDNSConfigForFrontend(render?.dns);
  setChecked("custom-dns-enabled", custom);
  setValue("dns-listen", dns.listen);
  setValue("dns-enhanced-mode", dns.enhanced_mode);
  setValue("dns-fake-ip-range", dns.fake_ip_range);
  setChecked("dns-use-hosts", dns.use_hosts);
  setChecked("dns-use-system-hosts", dns.use_system_hosts);
  setChecked("dns-respect-rules", dns.respect_rules);
  setChecked("dns-direct-follow-policy", dns.direct_nameserver_follow_policy);
  setValue("dns-default-nameserver", formatDNSLines(dns.default_nameserver));
  setValue("dns-nameserver", formatDNSLines(dns.nameserver));
  setValue("dns-fallback", formatDNSLines(dns.fallback));
  setValue(
    "dns-proxy-server-nameserver",
    formatDNSLines(dns.proxy_server_nameserver),
  );
  setValue("dns-direct-nameserver", formatDNSLines(dns.direct_nameserver));
  setValue("dns-fake-ip-filter", formatDNSLines(dns.fake_ip_filter));
  setValue("dns-nameserver-policy", formatNameserverPolicyText(dns.nameserver_policy));
  syncCustomDNSFields();
}

function readDNSFormConfig(fallbackDNS) {
  const fallback = normalizeDNSConfigForFrontend(fallbackDNS || DEFAULT_RENDER.dns);
  const dns = {
    ...fallback,
    enable: true,
    listen: getValue("dns-listen").trim() || DEFAULT_RENDER.dns.listen,
    use_hosts: getChecked("dns-use-hosts"),
    use_system_hosts: getChecked("dns-use-system-hosts"),
    respect_rules: getChecked("dns-respect-rules"),
    enhanced_mode:
      getValue("dns-enhanced-mode").trim() || DEFAULT_RENDER.dns.enhanced_mode,
    fake_ip_range:
      getValue("dns-fake-ip-range").trim() || DEFAULT_RENDER.dns.fake_ip_range,
    default_nameserver: parseDNSLines(getValue("dns-default-nameserver")),
    nameserver: parseDNSLines(getValue("dns-nameserver")),
    proxy_server_nameserver: parseDNSLines(getValue("dns-proxy-server-nameserver")),
    direct_nameserver: parseDNSLines(getValue("dns-direct-nameserver")),
    direct_nameserver_follow_policy: getChecked("dns-direct-follow-policy"),
    fallback: parseDNSLines(getValue("dns-fallback")),
    fake_ip_filter: parseDNSLines(getValue("dns-fake-ip-filter")),
    nameserver_policy: parseNameserverPolicyText(
      getValue("dns-nameserver-policy"),
    ),
  };
  if (!dns.default_nameserver.length) {
    dns.default_nameserver = [...DEFAULT_RENDER.dns.default_nameserver];
  }
  if (!dns.nameserver.length) {
    dns.nameserver = [...DEFAULT_RENDER.dns.nameserver];
  }
  return normalizeDNSConfigForFrontend(dns);
}

function syncCustomDNSFields() {
  const enabled = getChecked("custom-dns-enabled");
  const fields = document.getElementById("custom-dns-fields");
  if (!fields) return;
  fields.classList.toggle("hidden", !enabled);
  fields
    .querySelectorAll("input, select, textarea")
    .forEach((control) => {
      control.disabled = !enabled;
    });
}

function normalizeStableRenderConfig(render) {
  const stableDNS = deepClone(DEFAULT_RENDER.dns);
  const stableProfile = deepClone(DEFAULT_RENDER.profile);
  const stableSniffer = deepClone(DEFAULT_RENDER.sniffer);
  const customDNS = render?.custom_dns === true;
  const next = {
    ...(render || {}),
    mixed_port: 7897,
    allow_lan: true,
    mode: "rule",
    log_level: "info",
    ipv6: false,
    dns_enabled: true,
    custom_dns: customDNS,
    enhanced_mode: "fake-ip",
    unified_delay: true,
    tcp_concurrent: false,
    find_process_mode: "strict",
    global_client_fingerprint: "chrome",
    geodata_mode: false,
    geo_auto_update: false,
    geodata_loader: "",
    geo_update_interval: 0,
    dns: customDNS
      ? normalizeDNSConfigForFrontend(render?.dns || stableDNS)
      : stableDNS,
    profile: stableProfile,
    sniffer: stableSniffer,
    additional_rules: normalizeAdditionalRuleLines(render?.additional_rules),
  };
  delete next.geox_url;
  return next;
}

function sanitizeLocalDraftConfig(config) {
  const draftConfig = deepClone(config || {});
  draftConfig.service = {
    ...(draftConfig.service || {}),
    access_token: "",
    subscription_token: "",
  };
  draftConfig.render = normalizeStableRenderConfig(draftConfig.render || {});
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
    deleted_node_sources:
      nodeState.deleted_node_sources &&
      typeof nodeState.deleted_node_sources === "object"
        ? deepClone(nodeState.deleted_node_sources)
        : {},
    custom_nodes: Array.isArray(nodeState.custom_nodes)
      ? deepClone(nodeState.custom_nodes)
      : [],
  };
  return sanitized;
}

function pruneLocalDraftDeletedNodes(nodeState, config) {
  const sanitized = sanitizeLocalDraftNodeState(nodeState);
  if (!sanitized) return null;
  const activeSources = new Set(
    (Array.isArray(config?.subscriptions) ? config.subscriptions : [])
      .map((item) => String(item?.id || "").trim())
      .filter(Boolean),
  );
  const removedNodeIDs = new Set();
  Object.entries(sanitized.deleted_node_sources).forEach(
    ([nodeID, sourceIDs]) => {
      const sources = Array.isArray(sourceIDs)
        ? sourceIDs
            .map((sourceID) => String(sourceID || "").trim())
            .filter(Boolean)
        : [];
      if (
        sources.length &&
        !sources.some((sourceID) => activeSources.has(sourceID))
      ) {
        removedNodeIDs.add(nodeID);
        delete sanitized.deleted_node_sources[nodeID];
      }
    },
  );
  if (removedNodeIDs.size) {
    sanitized.deleted_nodes = sanitized.deleted_nodes.filter(
      (nodeID) => !removedNodeIDs.has(nodeID),
    );
  }
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

  state.config = {
    ...(response.config || {}),
    render: normalizeStableRenderConfig(response.config?.render || {}),
  };
  fillFormFromConfig(state.config);
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
    const value = option.groupOption
      ? render.group_options?.[option.key]
      : render[option.key];
    setChecked(option.id, value ?? option.default);
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
  setValue(
    "rule-group-node-mode",
    normalizeRuleGroupNodeMode(render.group_options?.rule_group_node_mode),
  );
  syncSourcePrefixModeControl();
  setDNSFormFromConfig(render);

  state.activeSourceMode =
    (render.source_mode || render.template_rule_mode) === "template"
      ? "template"
      : "rules";
  state.ruleMode = render.rule_mode || DEFAULT_RENDER.rule_mode;
  state.enabledRules = new Set(
    Array.isArray(render.enabled_rules) && render.enabled_rules.length
      ? render.enabled_rules
      : RULE_PRESETS[state.ruleMode] || RULE_PRESETS.custom,
  );
  state.customRules = Array.isArray(render.custom_rules)
    ? render.custom_rules
    : [];
  state.additionalRules = normalizeAdditionalRuleLines(render.additional_rules);
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
  setValue("additional-rules-input", state.additionalRules.join("\n"));
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
          .filter((item) => Boolean(item) && !(state.publicConverter && item.publicHidden))
          .map(
            (item) => `
            <label class="option-item">
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
      ${group.items.includes("enable_region_groups") ? renderRuleGroupNodeModeControl() : ""}
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

function renderRuleGroupNodeModeControl() {
  return `
    <div class="source-prefix-mode-control">
      <label for="rule-group-node-mode">策略组节点列表</label>
      <select id="rule-group-node-mode">
        <option value="full">完整：分流组可直接选择具体节点</option>
        <option value="compact">简洁：仅显示主策略、自动选择、DIRECT、REJECT</option>
      </select>
      <div class="option-desc">完整模式会让 AI、YouTube、Netflix 等分流组直接显示所有节点；广告拦截、私有网络、国内服务仍保持紧凑。</div>
    </div>
  `;
}

function normalizeRuleGroupNodeMode(value) {
  return value === "compact" ? "compact" : "full";
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
    syncCustomRuleFormVisibility();
    renderAdditionalRules();
    renderCustomRuleList();
    renderCustomRulePreview();
  }
}

function renderAdditionalRules() {
  const input = document.getElementById("additional-rules-input");
  if (!input) return;
  input.value = state.additionalRules.join("\n");
}

function updateAdditionalRulesFromInput() {
  state.additionalRules = normalizeAdditionalRuleLines(
    getValue("additional-rules-input"),
  );
  state.ruleMode = "custom";
  renderRulePresetControl();
  updateSummary();
}

function normalizeAdditionalRuleLines(value) {
  const source = Array.isArray(value) ? value : String(value || "").split(/\r?\n/);
  const rules = [];
  const seen = new Set();
  source.forEach((item) => {
    const rule = normalizeAdditionalRuleLine(item);
    if (!rule || seen.has(rule)) return;
    seen.add(rule);
    rules.push(rule);
  });
  return rules;
}

function normalizeAdditionalRuleLine(value) {
  let line = String(value || "").trim();
  if (!line || line.startsWith("#")) return "";
  line = line.replace(/^-\s*/, "").replace(/^['"]|['"]$/g, "").trim();
  if (!line || line.startsWith("#")) return "";
  const commentIndex = line.indexOf(" #");
  if (commentIndex >= 0) line = line.slice(0, commentIndex).trim();
  return line;
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

function templateUsesRegionGroups(item) {
  return (item?.groupProfile || "").includes("地区");
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
  if (templateUsesRegionGroups(selected)) {
    setChecked("opt-region-groups", true);
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
    "priority",
  );
  setChecked("custom-rule-enabled", rule.enabled !== false);
  setChecked("custom-rule-no-resolve", Boolean(rule.no_resolve));
  setValue(
    "custom-rule-payload",
    Array.isArray(rule.payload) ? rule.payload.join("\n") : "",
  );
  syncCustomRuleFormVisibility();
  syncCustomRuleCompatibility();
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
  setValue("custom-rule-insert-position", "priority");
  setChecked("custom-rule-enabled", true);
  setChecked("custom-rule-no-resolve", false);
  setValue("custom-rule-payload", "");
  syncCustomRuleFormVisibility();
  renderCustomRulePreview();
}

function handleCustomRuleLabelInput() {
  const key = getValue("custom-rule-key").trim();
  if (!key) {
    const generated = customRuleKeyFromLabel(getValue("custom-rule-label"));
    if (generated) setValue("custom-rule-key", generated);
  }
  renderCustomRulePreview();
}

function normalizeCustomRuleKeyField() {
  const current = getValue("custom-rule-key");
  const normalized = customRuleKeyFromLabel(current);
  if (current && current !== normalized) {
    setValue("custom-rule-key", normalized);
  }
  renderCustomRulePreview();
}

function customRuleKeyFromLabel(value) {
  return String(value || "")
    .trim()
    .toLowerCase()
    .normalize("NFKD")
    .replace(/[^a-z0-9]+/g, "_")
    .replace(/_+/g, "_")
    .replace(/^_+|_+$/g, "")
    .slice(0, 64);
}

function normalizeCustomRulePayloadInput() {
  const analysis = analyzeCustomRulePayloadInput(
    getValue("custom-rule-payload"),
  );
  setValue("custom-rule-payload", analysis.lines.join("\n"));
  syncCustomRuleCompatibility();
  renderCustomRulePreview();
  const details = [];
  if (analysis.repaired) details.push(`拆分 ${analysis.repaired} 处粘连`);
  if (analysis.duplicates) details.push(`移除 ${analysis.duplicates} 条重复`);
  showToast(
    analysis.lines.length
      ? `已整理 ${analysis.lines.length} 条规则${details.length ? `，${details.join("，")}` : ""}。`
      : "没有可整理的有效规则。",
    !analysis.lines.length,
  );
}

function clearCustomRulePayloadInput() {
  setValue("custom-rule-payload", "");
  renderCustomRulePreview();
}

function renderCustomRulePreview() {
  const draft = buildCustomRuleDraft();
  const preview = document.getElementById("custom-rule-preview");
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

  renderCustomRuleEditorSummary(draft, missing);
  renderCustomRulePayloadMeta(draft);

  if (!draft.key && !draft.label && !draft.icon) {
    preview.textContent = "未填写";
    return;
  }

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

function renderCustomRuleEditorSummary(draft, missing = []) {
  const summary = document.getElementById("custom-rule-editor-summary");
  if (!summary) return;
  const editIndex = getValue("custom-rule-edit-index");
  const mode = editIndex === "" ? "新建" : `编辑 #${Number.parseInt(editIndex, 10) + 1}`;
  const status = missing.length
    ? `待补全 ${missing.length} 项`
    : draft.enabled
      ? "启用"
      : "停用";
  const sourceLabel = customRuleSourceTypeLabel(draft.source_type);
  const ruleCount =
    draft.source_type === "inline"
      ? `${draft.payload.length} 条`
      : draft.source_type === "group_only"
        ? "不生成"
        : "-";
  const items = [
    ["状态", `${mode} · ${status}`],
    ["目标", resolveCustomRulePreviewTarget(draft)],
    ["来源", sourceLabel],
    ["行为", draft.source_type === "group_only" ? "-" : draft.behavior],
    ["规则", ruleCount],
    [
      "位置",
      draft.source_type === "group_only"
        ? "-"
        : customRuleInsertPositionLabel(draft.insert_position),
    ],
  ];
  summary.innerHTML = items
    .map(
      ([label, value]) =>
        `<span class="custom-rule-summary-pill">${escapeHtml(label)}<strong>${escapeHtml(value)}</strong></span>`,
    )
    .join("");
}

function renderCustomRulePayloadMeta(draft) {
  const meta = document.getElementById("custom-rule-payload-meta");
  if (!meta) return;
  if (draft.source_type !== "inline") {
    meta.textContent = "";
    return;
  }
  const analysis = analyzeCustomRulePayloadInput(
    getValue("custom-rule-payload"),
  );
  const parts = [`${analysis.lines.length} 条有效规则`];
  if (analysis.entered) parts.push(`${analysis.entered} 行输入`);
  if (analysis.repaired) parts.push(`${analysis.repaired} 处粘连已拆分`);
  if (analysis.ignored) parts.push(`${analysis.ignored} 行无效或注释`);
  if (analysis.duplicates) parts.push(`${analysis.duplicates} 条重复`);
  if (analysis.classical) parts.push("已识别 classical");
  meta.textContent = parts.join(" · ");
}

function customRuleSourceTypeLabel(value) {
  return (
    {
      inline: "内联",
      http: "远程",
      file: "本地",
      group_only: "仅代理组",
    }[value] || value || "-"
  );
}

function customRuleInsertPositionLabel(value) {
  return (
    {
      priority: "优先匹配",
      before_match: "优先匹配",
      after_adblock: "优先匹配",
      before_domestic: "优先匹配",
      before_non_cn: "优先匹配",
    }[value] || "优先匹配"
  );
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

  const editIndex =
    indexValue === "" ? -1 : Number.parseInt(indexValue, 10);
  const duplicateBuiltin = BUILTIN_RULES.some((item) => item.key === draft.key);
  const duplicateCustom = state.customRules.some((item, index) => {
    if (index === editIndex) return false;
    return item.key === draft.key;
  });
  const duplicate = duplicateBuiltin || duplicateCustom;
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
  document
    .getElementById("custom-rule-behavior-wrap")
    .classList.toggle("hidden", sourceType === "group_only");
  document
    .getElementById("custom-rule-format-wrap")
    .classList.toggle("hidden", sourceType === "group_only");
  document
    .getElementById("custom-rule-insert-position-wrap")
    .classList.toggle("hidden", sourceType === "group_only");
  document
    .getElementById("custom-rule-no-resolve-wrap")
    .classList.toggle("hidden", sourceType === "group_only");
}

function syncCustomRuleCompatibility() {
  const sourceType = getValue("custom-rule-source-type");
  let behavior = getValue("custom-rule-behavior");
  let format = getValue("custom-rule-format");
  if (sourceType === "group_only") return;

  const payload = normalizeCustomRulePayloadLines(
    getValue("custom-rule-payload"),
  );
  const hasClassicalPayload =
    sourceType === "inline" && hasClassicalCustomRulePayload(payload);

  if (hasClassicalPayload && behavior !== "classical") {
    behavior = "classical";
    setValue("custom-rule-behavior", behavior);
  }
  if (!hasClassicalPayload && behavior === "classical" && format === "mrs") {
    behavior = "domain";
    setValue("custom-rule-behavior", behavior);
  }
  if (hasClassicalPayload && format === "mrs") {
    format = "text";
    setValue("custom-rule-format", format);
  }
  if (sourceType === "http") {
    const remoteShape = normalizeRemoteRuleProviderShape(
      behavior,
      format,
      getValue("custom-rule-url"),
      getValue("custom-rule-path"),
    );
    if (remoteShape.behavior !== behavior) {
      behavior = remoteShape.behavior;
      setValue("custom-rule-behavior", behavior);
    }
    if (remoteShape.format !== format) {
      format = remoteShape.format;
      setValue("custom-rule-format", format);
    }
  }
}

function buildCustomRuleDraft() {
  const sourceType = getValue("custom-rule-source-type");
  const payload = normalizeCustomRulePayloadLines(getValue("custom-rule-payload"));
  let behavior = getValue("custom-rule-behavior");
  let format = getValue("custom-rule-format");
  const targetMode = getValue("custom-rule-target-mode");
  const targetGroup = targetMode === "existing_group"
    ? normalizePolicyTargetAlias(getValue("custom-rule-target-group"))
    : "";
  if (sourceType === "inline" && hasClassicalCustomRulePayload(payload)) {
    behavior = "classical";
    if (format === "mrs") format = "text";
  }
  if (sourceType === "http") {
    const remoteShape = normalizeRemoteRuleProviderShape(
      behavior,
      format,
      getValue("custom-rule-url"),
      getValue("custom-rule-path"),
    );
    behavior = remoteShape.behavior;
    format = remoteShape.format;
  }
  return {
    key: getValue("custom-rule-key").trim().toLowerCase(),
    label: getValue("custom-rule-label").trim(),
    icon: getValue("custom-rule-emoji").trim(),
    enabled: getChecked("custom-rule-enabled"),
    target_mode: targetMode,
    target_group: targetGroup,
    source_type: sourceType,
    behavior,
    format,
    url: getValue("custom-rule-url").trim(),
    path: getValue("custom-rule-path").trim(),
    interval: Number.parseInt(getValue("custom-rule-interval"), 10) || 86400,
    payload,
    insert_position: "before_match",
    no_resolve: getChecked("custom-rule-no-resolve"),
  };
}

function normalizeCustomRulePayloadLines(value) {
  return analyzeCustomRulePayloadInput(value).lines;
}

function normalizeCustomRulePayloadLine(value) {
  let line = prepareCustomRulePayloadLine(value);
  if (!line) return "";
  const parts = line.split(",");
  if (parts.length >= 3) {
    let targetIndex = parts.length - 1;
    if (parts[targetIndex].trim().toLowerCase() === "no-resolve" && parts.length >= 4) {
      targetIndex -= 1;
    }
    if (isPolicyTargetAlias(parts[targetIndex])) {
      parts.splice(targetIndex, 1);
      line = parts.join(",");
    }
  }
  return line;
}

function prepareCustomRulePayloadLine(value) {
  let line = String(value || "").trim();
  if (!line || line.startsWith("#")) return "";
  line = line.replace(/^-\s*/, "").replace(/^['"]|['"]$/g, "").trim();
  if (!line || line.startsWith("#")) return "";
  const commentIndex = line.indexOf(" #");
  if (commentIndex >= 0) line = line.slice(0, commentIndex).trim();
  return line;
}

const CUSTOM_RULE_PAYLOAD_PREFIXES = [
  "PROCESS-PATH-REGEX,", "PROCESS-NAME-REGEX,", "DOMAIN-SUFFIX,",
  "DOMAIN-KEYWORD,", "DOMAIN-WILDCARD,", "DOMAIN-REGEX,", "SRC-IP-CIDR,",
  "SRC-IP-SUFFIX,", "SRC-IP-ASN,", "PROCESS-NAME,", "PROCESS-PATH,",
  "IP-CIDR6,", "IP-CIDR,", "IP-SUFFIX,", "IP-ASN,", "SRC-GEOIP,",
  "DST-PORT,", "SRC-PORT,", "RULE-SET,", "SUB-RULE,", "GEOSITE,",
  "GEOIP,", "DOMAIN,", "IN-TYPE,", "IN-USER,", "IN-NAME,", "NETWORK,",
  "UID,", "AND,", "OR,", "NOT,",
];

const CUSTOM_RULE_PAYLOAD_NO_SPLIT_PREFIXES = [
  "DOMAIN-REGEX,", "PROCESS-PATH-REGEX,", "PROCESS-NAME-REGEX,",
  "AND,", "OR,", "NOT,",
];

function splitConcatenatedCustomRulePayloadLine(value) {
  const line = prepareCustomRulePayloadLine(value);
  if (!line) return [];
  const upper = line.toUpperCase();
  if (CUSTOM_RULE_PAYLOAD_NO_SPLIT_PREFIXES.some((prefix) => upper.startsWith(prefix))) {
    return [line];
  }
  const boundaries = [0];
  for (let index = 1; index < line.length; index += 1) {
    if (!/[A-Za-z0-9._/-]/.test(line[index - 1])) continue;
    if (CUSTOM_RULE_PAYLOAD_PREFIXES.some((prefix) => upper.startsWith(prefix, index))) {
      boundaries.push(index);
    }
  }
  boundaries.push(line.length);
  return boundaries
    .slice(0, -1)
    .map((start, index) => line.slice(start, boundaries[index + 1]).trim())
    .filter(Boolean);
}

function analyzeCustomRulePayloadInput(value) {
  const lines = [];
  const seen = new Set();
  let entered = 0;
  let ignored = 0;
  let duplicates = 0;
  let repaired = 0;
  String(value || "")
    .split(/\r?\n/)
    .forEach((item) => {
      if (!String(item || "").trim()) return;
      entered += 1;
      const segments = splitConcatenatedCustomRulePayloadLine(item);
      if (!segments.length) {
        ignored += 1;
        return;
      }
      if (segments.length > 1) repaired += segments.length - 1;
      segments.forEach((segment) => {
        const line = normalizeCustomRulePayloadLine(segment);
        if (!line) {
          ignored += 1;
          return;
        }
        if (seen.has(line)) {
          duplicates += 1;
          return;
        }
        seen.add(line);
        lines.push(line);
      });
    });
  return {
    entered,
    ignored,
    duplicates,
    repaired,
    lines,
    classical: lines.filter((line) => isClassicalCustomRulePayloadLine(line)).length,
  };
}

function hasClassicalCustomRulePayload(payload) {
  return payload.some((line) => isClassicalCustomRulePayloadLine(line));
}

function isClassicalCustomRulePayloadLine(line) {
  const parts = String(line || "").trim().split(",");
  if (parts.length < 2) return false;
  return new Set([
    "DOMAIN", "DOMAIN-SUFFIX", "DOMAIN-KEYWORD", "DOMAIN-WILDCARD", "DOMAIN-REGEX",
    "GEOSITE", "IP-CIDR", "IP-CIDR6", "IP-SUFFIX", "IP-ASN", "GEOIP", "SRC-GEOIP",
    "SRC-IP-ASN", "SRC-IP-CIDR", "SRC-IP-SUFFIX", "DST-PORT", "SRC-PORT",
    "PROCESS-NAME", "PROCESS-PATH", "PROCESS-PATH-REGEX", "PROCESS-NAME-REGEX",
    "IN-TYPE", "IN-USER", "IN-NAME", "NETWORK", "UID", "SUB-RULE", "RULE-SET", "AND", "OR", "NOT",
  ]).has(parts[0].trim().toUpperCase());
}

function resolveCustomRulePreviewTarget(rule) {
  if (rule.target_mode === "direct") return "DIRECT";
  if (rule.target_mode === "reject") return "REJECT";
  if (rule.target_mode === "existing_group")
    return normalizePolicyTargetAlias(rule.target_group) || "节点选择";
  return `${rule.icon ? `${rule.icon} ` : ""}${rule.label || "未命名规则"}`.trim();
}

function normalizePolicyTargetAlias(value) {
  const target = String(value || "").trim();
  return target.toUpperCase() === "PROXY" ? "🚀 节点选择" : target;
}

function isPolicyTargetAlias(value) {
  const target = String(value || "").trim();
  return ["PROXY", "DIRECT", "REJECT", "🚀 节点选择", "⚡ 自动选择", "🐟 漏网之鱼"].includes(target);
}

function previewFormatExt(format) {
  if (format === "mrs") return "mrs";
  if (format === "text") return "txt";
  return "yaml";
}

function normalizeRemoteRuleProviderShape(behavior, format, url, path) {
  let nextBehavior = String(behavior || "").trim().toLowerCase() || "domain";
  let nextFormat =
    String(format || "").trim().toLowerCase() ||
    inferRuleProviderFormatFromSource(url, path) ||
    "text";
  if (isClassicalYAMLRuleProviderSource(url, path)) {
    nextBehavior = "classical";
    nextFormat = "yaml";
  }
  if (nextBehavior === "classical" && nextFormat === "mrs") {
    nextFormat = "yaml";
  }
  return { behavior: nextBehavior, format: nextFormat };
}

function inferRuleProviderFormatFromSource(url, path) {
  const source = String(path || url || "").trim().toLowerCase();
  if (source.endsWith(".mrs")) return "mrs";
  if (source.endsWith(".yaml") || source.endsWith(".yml")) return "yaml";
  if (source.endsWith(".txt") || source.endsWith(".text")) return "text";
  return "";
}

function isClassicalYAMLRuleProviderSource(url, path) {
  const source = `${url || ""} ${path || ""}`.trim().toLowerCase();
  return (
    source.includes("/rule/clash/") ||
    source.includes("blackmatrix7/ios_rule_script")
  );
}

function sourceDomainFromUrl(rawUrl) {
  try {
    const parsed = new URL(String(rawUrl || "").trim());
    return parsed.hostname || "";
  } catch {
    return "";
  }
}

function isPrivateSubscriptionURL(rawUrl) {
  const host = sourceDomainFromUrl(rawUrl).toLowerCase();
  if (!host) return false;
  if (host === "localhost" || host.endsWith(".local")) return true;
  const parts = host.split(".").map((part) => Number.parseInt(part, 10));
  if (parts.length !== 4 || parts.some((part) => Number.isNaN(part))) {
    return false;
  }
  if (
    parts[0] === 10 ||
    parts[0] === 127 ||
    (parts[0] === 169 && parts[1] === 254)
  ) {
    return true;
  }
  if (parts[0] === 172 && parts[1] >= 16 && parts[1] <= 31) return true;
  return parts[0] === 192 && parts[1] === 168;
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

function normalizeSourceID(value) {
  return String(value || "").trim();
}

function nextSubscriptionIDFromSet(used, start = 1) {
  let index = Math.max(1, start);
  while (used.has(`source-${index}`)) {
    index += 1;
  }
  return `source-${index}`;
}

function nextSubscriptionID(existing = state.subscriptions) {
  const used = new Set(
    (Array.isArray(existing) ? existing : [])
      .map((item) => normalizeSourceID(item?.id))
      .filter(Boolean),
  );
  return nextSubscriptionIDFromSet(used);
}

function uniqueSubscriptionID(value, used, start = 1) {
  const current = normalizeSourceID(value);
  if (current && !used.has(current)) {
    used.add(current);
    return current;
  }
  const next = nextSubscriptionIDFromSet(used, start);
  used.add(next);
  return next;
}

function createSubscriptionEntry(values = {}) {
  const id = normalizeSourceID(values.id) || nextSubscriptionID();
  const index = state.subscriptions.length + 1;
  const url = values.url || "";
  return {
    id,
    name: values.name || id,
    emoji:
      values.emoji !== undefined
        ? normalizeSourceEmoji(values.emoji)
        : defaultSourceEmoji(index - 1),
    source_logo: values.source_logo || "",
    enabled: values.enabled !== false,
    url,
    user_agent: values.user_agent || DEFAULT_SOURCE_USER_AGENT,
    allow_lan: !state.publicConverter && values.allow_lan === true,
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

function sourceLANWarning(item) {
  if (state.publicConverter && isPrivateSubscriptionURL(item?.url)) {
    return "公开服务不支持局域网订阅地址。";
  }
  if (!isPrivateSubscriptionURL(item?.url) || item?.allow_lan === true) {
    return "";
  }
  return "局域网订阅地址需在高级中开启允许局域网。";
}

function sourceNamePreview(item) {
  const mode = currentSourcePrefixMode();
  const prefix = buildSourcePrefix(item, mode);
  if (!prefix) return SOURCE_NAME_PREVIEW;
  if (mode === "emoji") return `${prefix} ${SOURCE_NAME_PREVIEW}`;
  return `${prefix}${SOURCE_PREFIX_SEPARATOR}${SOURCE_NAME_PREVIEW}`;
}

function sourceNamePreviewLabel(item) {
  return `命名预览：${sourceNamePreview(item)}`;
}

function renderSubscriptionRow(item, index) {
  return `
    <div class="source-row subscription-item" data-sub-row-index="${index}">
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
      <div class="source-row-meta" data-source-meta-index="${index}">${renderSourceRowMeta(item, index)}</div>
      <details class="source-advanced inline-advanced">
        <summary>高级</summary>
        ${state.publicConverter ? "" : `<label class="checkbox-line compact-check">
          <input type="checkbox" data-sub-field="allow_lan" data-sub-index="${index}" ${item.allow_lan ? "checked" : ""} />
          允许局域网订阅地址
        </label>`}
        <input type="text" data-sub-field="user_agent" data-sub-index="${index}" value="${escapeHtml(item.user_agent || DEFAULT_SOURCE_USER_AGENT)}" />
      </details>
    </div>
  `;
}

function renderSourceRowMeta(item, index) {
  const warnings = [
    sourceEmojiDuplicateWarning(item, index),
    sourceLANWarning(item),
  ].filter(Boolean);
  const lanBadge = item.allow_lan && !state.publicConverter
    ? '<span class="source-meta-separator">·</span><span class="source-meta-link">LAN</span>'
    : "";
  const preview = sourceNamePreviewLabel(item);
  return `
    <span class="source-domain">${escapeHtml(sourceDomainFromUrl(item.url) || "未识别域名")}</span>
    <span class="source-meta-separator">·</span>
    <span class="source-preview" data-sub-index="${index}" title="${escapeHtml(preview)}">${escapeHtml(preview)}</span>
    <span class="source-meta-separator">·</span>
    <span class="source-meta-link">User-Agent</span>
    ${lanBadge}
    ${warnings.length ? `<span class="source-meta-separator">·</span><span class="source-warning-inline">${escapeHtml(warnings.join(" "))}</span>` : ""}
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
  const usedIDs = new Set();
  return (Array.isArray(items) ? items : []).map((item, index) => {
    const source = item || {};
    const id = uniqueSubscriptionID(source.id, usedIDs, index + 1);
    return createSubscriptionEntry({
      id,
      name: source.name || id,
      emoji:
        source.emoji !== undefined ? source.emoji : defaultSourceEmoji(index),
      source_logo: source.source_logo || "",
      enabled: source.enabled !== false,
      url: source.url || "",
      user_agent: source.user_agent || DEFAULT_SOURCE_USER_AGENT,
      allow_lan: !state.publicConverter && source.allow_lan === true,
    });
  });
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
        if (!state.publicConverter && isPrivateSubscriptionURL(element.value)) {
          state.subscriptions[index].allow_lan = true;
        }
        updateSubscriptionRowMeta(index);
      } else if (field === "name") {
        updateSourceRowPreview(index);
      } else if (field === "allow_lan") {
        renderSubscriptionManager();
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
  const preview = sourceNamePreviewLabel(item);
  row.textContent = preview;
  row.title = preview;
}

function updateSubscriptionRowMeta(index) {
  const item = state.subscriptions[index];
  const meta = document.querySelector(`[data-source-meta-index="${index}"]`);
  if (!item || !meta) return;
  meta.innerHTML = renderSourceRowMeta(item, index);
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
      <div class="manual-empty-desc">可添加 ss://、vless://、trojan://、anytls://、mieru://、wireguard:// 或 WireGuard 配置片段。</div>
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
        <textarea class="manual-node-textarea" data-inline-field="content" data-inline-index="${index}" placeholder="粘贴 ss://、vless://、anytls://、mieru://、wireguard:// 或 WireGuard 配置片段">${escapeHtml(item.content)}</textarea>
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
  const usedIDs = new Set(
    nextEntries.map((item) => normalizeSourceID(item?.id)).filter(Boolean),
  );
  parsed.forEach((item) => {
    const id = uniqueSubscriptionID("", usedIDs, nextEntries.length + 1);
    nextEntries.push(
      createSubscriptionEntry({
        id,
        name: item.name || id,
        emoji: defaultSourceEmoji(nextEntries.length),
        url: item.url,
        enabled: true,
        user_agent: DEFAULT_SOURCE_USER_AGENT,
        allow_lan: !state.publicConverter && isPrivateSubscriptionURL(item.url),
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
  const usedIDs = new Set();
  return items.map((item, index) => {
    const baseName =
      (item.name || `source-${index + 1}`).trim() || `source-${index + 1}`;
    seen[baseName] = (seen[baseName] || 0) + 1;
    const name =
      seen[baseName] === 1 ? baseName : `${baseName} ${seen[baseName]}`;
    return {
      ...item,
      id: uniqueSubscriptionID(item.id, usedIDs, index + 1),
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
  const hasNodeCache = state.allNodes.length > 0 && !state.nodeCacheStale;
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
    ["来源数量", hasNodeCache ? sources.size : "-"],
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

function markNodeCacheStale() {
  state.nodeRequest?.controller.abort();
  state.nodeRequest = null;
  setNodeLoading(false);
  state.nodeCacheStale = true;
  renderConfigNodeSummary();
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
        (item) =>
          normalizeSourceID(item.id) === normalizeSourceID(meta.source_id),
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
  const ruleGroupNodeMode = normalizeRuleGroupNodeMode(
    getValue("rule-group-node-mode") ||
      DEFAULT_RENDER.group_options.rule_group_node_mode,
  );
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
  const customDNS = getChecked("custom-dns-enabled");
  const dnsConfig = customDNS
    ? readDNSFormConfig(base.render?.dns || DEFAULT_RENDER.dns)
    : deepClone(DEFAULT_RENDER.dns);

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
    custom_dns: customDNS,
    dns: dnsConfig,
    source_prefix: prefixMode !== "none",
    group_options: {
      ...(base.render?.group_options || DEFAULT_RENDER.group_options),
      enable_region_groups: getChecked("opt-region-groups"),
      rule_group_node_mode: ruleGroupNodeMode,
      include_real_nodes_in_rule_groups: ruleGroupNodeMode === "full",
      special_groups_use_compact: true,
    },
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
    additional_rules:
      !activeOnly || activeMode === "rules"
        ? normalizeAdditionalRuleLines(getValue("additional-rules-input"))
        : [],
  };
  base.render = normalizeStableRenderConfig(base.render);

  const usedSubscriptionIDs = new Set();
  base.subscriptions = state.subscriptions
    .filter((item) => item.url.trim())
    .map((item, index) => {
      const id = uniqueSubscriptionID(item.id, usedSubscriptionIDs, index + 1);
      return {
        id,
        name: item.name.trim() || id,
        emoji: normalizeSourceEmoji(item.emoji),
        source_logo: item.source_logo?.trim() || "",
        enabled: item.enabled !== false,
        url: item.url.trim(),
        user_agent: item.user_agent?.trim() || DEFAULT_SOURCE_USER_AGENT,
        insecure_skip_verify: getChecked("opt-skip-tls"),
        allow_lan: item.allow_lan === true && !state.publicConverter,
      };
    });

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
  if (state.generationInFlight) {
    showToast("生成任务正在进行，请稍候。");
    return;
  }
  state.generationInFlight = true;
  renderResult();
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
      state.lastError = state.lastError || "当前会话写入失败";
      stopGenerateProgressPolling();
      renderResult();
      return;
    }
    markNodeCacheStale();

    const refreshResult = await refreshNow();
    if (!refreshResult) {
      state.generateStatus = "error";
      stopGenerateProgressPolling();
      renderResult();
      return;
    }

    state.yamlRequest?.controller.abort();
    state.yamlRequest = null;
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
    showToast(describeGenerationError(state.lastError).title, true);
  } finally {
    state.generationInFlight = false;
    renderResult();
  }
}

function describeGenerationError(rawError) {
  const raw = String(rawError || "").trim();
  const lower = raw.toLowerCase();

  if (lower.includes("rate_limited") || lower.includes("too many requests")) {
    return {
      title: "请求过于频繁",
      description: "生成请求触发了安全限流，请等待一分钟后再试。当前配置已经保存，不需要重新填写。",
    };
  }
  if (lower.includes("refresh_in_progress")) {
    return {
      title: "已有生成任务正在运行",
      description: "请等待当前任务结束后再试，不需要重复点击生成按钮。",
    };
  }
  if (lower.includes("cross_origin_request")) {
    return {
      title: "页面来源校验失败",
      description: "请从当前服务地址直接打开页面，不要通过第三方网页或嵌入页面操作。",
    };
  }
  if (lower.includes("workspace_not_found") || lower.includes("workspace_required")) {
    return {
      title: "当前隐私会话已失效",
      description: "请先保存本机草稿，再重新打开页面恢复草稿后生成。",
    };
  }

  if (lower.includes("no allowed ips") || lower.includes("blocked host")) {
    return {
      title: "订阅地址无法访问",
      description:
        "该域名解析到了受限网络地址。请检查本机 DNS 或代理设置，或改用可公网访问的订阅地址。",
    };
  }
  if (lower.includes("resolve host") || lower.includes("no such host")) {
    return {
      title: "订阅域名无法解析",
      description: "请确认订阅地址拼写正确，并检查当前网络的 DNS 是否可用。",
    };
  }
  if (lower.includes("timeout") || lower.includes("deadline exceeded")) {
    return {
      title: "订阅服务器响应超时",
      description: "订阅服务器暂时没有响应。请稍后重试，或确认该地址能从当前网络访问。",
    };
  }
  const statusMatch = raw.match(/status code\s+(\d{3})/i);
  if (statusMatch) {
    const statusCode = statusMatch[1];
    return {
      title: `订阅服务器返回 HTTP ${statusCode}`,
      description:
        statusCode === "401" || statusCode === "403"
          ? "订阅链接可能已失效或无权访问，请重新获取有效链接。"
          : "订阅服务器当前返回异常，请检查链接或稍后重试。",
    };
  }
  if (lower.includes("no nodes available") || lower.includes("fetch_failed")) {
    return {
      title: "没有获取到可用节点",
      description: "请检查订阅地址是否有效，并确认订阅内容中包含受支持的节点。",
    };
  }
  if (lower.includes("too large") || lower.includes("size limit")) {
    return {
      title: "订阅文件超过大小限制",
      description: "请减少订阅内容，或使用体积更小的订阅文件后重试。",
    };
  }
  return {
    title: "生成失败",
    description: "当前配置未能生成。请检查订阅源后重试，详细原因可在诊断日志中查看。",
  };
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
    const code = String(response?.error?.code || "").trim();
    const message = readAPIError(response) || "写入当前会话失败。";
    state.lastError = code ? `[${code}] ${message}` : message;
    showToast(describeGenerationError(state.lastError).title, true);
    return false;
  }
  return true;
}

function wait(milliseconds) {
  return new Promise((resolve) => window.setTimeout(resolve, milliseconds));
}

async function recoverRunningRefresh(timeoutMilliseconds = 30000) {
  const deadline = Date.now() + timeoutMilliseconds;
  while (Date.now() < deadline) {
    await wait(800);
    const status = await fetchJSON("/api/status");
    if (status?.error) return null;
    if (status?.refreshing) continue;

    const published = await fetchJSON("/api/published");
    if (!published?.ok || !published?.publish_id) return null;
    return {
      ...published,
      ok: true,
      node_count: Number.parseInt(status?.node_count, 10) || 0,
      subscription_url: published.subscription_url || published.url || "",
    };
  }
  return null;
}

async function refreshNow() {
  let response = await fetchJSON("/api/refresh", { method: "POST" });
  if (!response?.ok && response?._networkError) {
    await wait(800);
    response = await fetchJSON("/api/refresh", { method: "POST" });
  }
  if (!response?.ok && response?.error?.code === "REFRESH_IN_PROGRESS") {
    const recovered = await recoverRunningRefresh();
    if (recovered) return recovered;
  }
  if (!response?.ok) {
    const code = String(response?.error?.code || "").trim();
    const message = readAPIError(response) || "刷新失败";
    state.lastError = code ? `[${code}] ${message}` : message;
    showToast(describeGenerationError(state.lastError).title, true);
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
  }, 800);
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
  const summaryRefreshButton = document.getElementById("summary-refresh-btn");
  const errorCard = document.getElementById("result-error-card");
  const errorTitle = document.getElementById("result-error-title");
  const errorDescription = document.getElementById("result-error-description");
  const hasResult = Boolean(state.generatedUrl);
  const published = state.published;
  const errorState = describeGenerationError(state.lastError);

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
    statusText = errorState.title;
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
  if (errorCard) {
    const showError = state.generateStatus === "error";
    errorCard.classList.toggle("hidden", !showError);
    if (showError) {
      errorTitle.textContent = errorState.title;
      errorDescription.textContent = errorState.description;
    }
  }
  if (actionStatus) {
    actionStatus.className = `result-status ${statusClass}`.trim();
    actionStatus.innerHTML = `
      <span class="status-dot ${statusDotClass}"></span>
      <span>${escapeHtml(statusText)}</span>
      ${state.generateStatus === "error" ? '<span class="result-status-hint">查看原因</span>' : ""}
    `;
    if (state.generateStatus === "error") {
      actionStatus.setAttribute("role", "button");
      actionStatus.setAttribute("tabindex", "0");
      actionStatus.setAttribute("title", "查看失败原因");
    } else {
      actionStatus.removeAttribute("role");
      actionStatus.removeAttribute("tabindex");
      actionStatus.removeAttribute("title");
    }
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
    viewButton.disabled = !hasResult;
  }
  if (generateButton) {
    generateButton.disabled = state.generationInFlight;
    if (state.generateStatus === "generating") {
      setButtonIconText(generateButton, "生成中...");
    } else if (state.generateStatus === "success") {
      setButtonIconText(generateButton, "重新生成配置");
    } else {
      setButtonIconText(generateButton, "生成订阅链接");
    }
  }
  if (summaryRefreshButton) {
    summaryRefreshButton.disabled = state.generationInFlight;
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
    [
      "已返回订阅信息",
      published?.has_subscription_userinfo ? "是" : "否",
    ],
    ["Header", published?.has_subscription_userinfo ? "subscription-userinfo" : "-"],
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

function focusSubscriptionSources() {
  switchWorkspace("config");
  const list = document.getElementById("subscription-list");
  const firstSource = list?.querySelector(".source-url-input");
  list?.scrollIntoView({ behavior: "smooth", block: "center" });
  window.setTimeout(() => firstSource?.focus({ preventScroll: true }), 260);
}

function focusGenerationError() {
  if (state.generateStatus !== "error") return;
  switchWorkspace("config");
  window.setTimeout(() => {
    document
      .getElementById("result-panel")
      ?.scrollIntoView({ behavior: "smooth", block: "center" });
  }, 80);
}

function previewYAMLWorkspace() {
  switchWorkspace("yaml");
}

function cancelWorkspaceViewRequests() {
  state.nodeRequest?.controller.abort();
  state.yamlRequest?.controller.abort();
  state.nodeRequest = null;
  state.yamlRequest = null;
  setNodeLoading(false);
}

function setNodeLoading(loading) {
  document.getElementById("refresh-nodes-btn").disabled = loading;
  const body = document.getElementById("node-table-body");
  body.setAttribute("aria-busy", String(loading));
  if (loading && !state.nodes.length) {
    body.innerHTML = '<tr><td colspan="7" class="table-empty">正在加载节点...</td></tr>';
  }
}

function loadNodes() {
  if (!state.workspaceId) return Promise.resolve();
  const url = `/api/nodes?${buildNodeQueryParams()}`;
  if (state.nodeRequest?.url === url && state.nodeRequest.workspaceId === state.workspaceId) {
    return state.nodeRequest.promise;
  }
  state.nodeRequest?.controller.abort();
  const request = { url, workspaceId: state.workspaceId, controller: new AbortController() };
  state.nodeRequest = request;
  setNodeLoading(true);
  request.promise = loadNodePage(request).finally(() => {
    if (state.nodeRequest !== request) return;
    state.nodeRequest = null;
    setNodeLoading(false);
  });
  return request.promise;
}

async function loadNodePage(request) {
  const response = await fetchJSON(request.url, { signal: request.controller.signal });
  if (state.nodeRequest !== request) return;
  if (!response?.ok) {
    state.nodeCacheStale = true;
    document.getElementById("node-table-body").innerHTML =
      `<tr><td colspan="7" class="table-empty">${escapeHtml(readAPIError(response) || "加载节点失败。")}</td></tr>`;
    showToast(readAPIError(response) || "加载节点失败。", true);
    return;
  }

  state.allNodes = Array.isArray(response.nodes) ? response.nodes : [];
  state.filteredNodes = state.allNodes;
  state.nodes = state.allNodes;
  state.nodeSummary = response.summary || {
    total: state.nodes.length,
    enabled: 0,
    disabled: 0,
    modified: 0,
    warnings: 0,
  };
  state.nodePagination.total = Number.parseInt(state.nodeSummary.total, 10) || 0;
  const totalPages = Math.max(
    1,
    Math.ceil(state.nodePagination.total / state.nodePagination.pageSize),
  );
  if (state.nodePagination.page > totalPages) {
    state.nodePagination.page = totalPages;
    await loadNodes();
    return;
  }
  state.nodeFetchNotices = [
    ...(Array.isArray(response.warnings) ? response.warnings : []),
    ...(Array.isArray(response.errors)
      ? response.errors.map((item) => item?.message || String(item))
      : []),
  ].filter(Boolean);
  buildNodeIndexes();
  state.nodeCacheStale = false;
  state.nodeSourceOptions = Array.isArray(response.source_options)
    ? response.source_options
    : state.nodeSourceOptions;
  syncSourceFilterOptions();
  renderNodeEditor();
  renderConfigNodeSummary();
}

function buildNodeQueryParams(options = {}) {
  const params = new URLSearchParams();
  if (options.all) {
    params.set("all", "1");
  } else {
    params.set("page", String(state.nodePagination.page || 1));
    params.set("page_size", String(state.nodePagination.pageSize || 25));
  }
  const q = String(state.nodeFilters.q || "").trim();
  if (q) params.set("q", q);
  if (state.nodeFilters.type && state.nodeFilters.type !== "all")
    params.set("type", state.nodeFilters.type);
  if (state.nodeFilters.region && state.nodeFilters.region !== "ALL")
    params.set("region", state.nodeFilters.region);
  if (state.nodeFilters.status && state.nodeFilters.status !== "all")
    params.set("status", state.nodeFilters.status);
  if (state.nodeFilters.source) params.set("source", state.nodeFilters.source);
  return params.toString();
}

async function loadAllFilteredNodes() {
  const response = await fetchJSON(`/api/nodes?${buildNodeQueryParams({ all: true })}`);
  if (!response?.ok) {
    showToast(readAPIError(response) || "加载筛选节点失败。", true);
    return [];
  }
  return Array.isArray(response.nodes) ? response.nodes : [];
}

function buildNodeIndexes() {
  state.nodeById = new Map();
  state.allNodes.forEach((node) => {
    if (!node?.id) return;
    node._searchText = buildNodeSearchText(node);
    node._sourceText = String(node.source?.name || "").toLowerCase();
    state.nodeById.set(node.id, node);
  });
}

function buildNodeSearchText(node) {
  return [
    node.name,
    node.server,
    node.type,
    regionLabel(node.region),
    node.region,
    node.source?.name,
    ...(Array.isArray(node.tags) ? node.tags : []),
  ]
    .join(" ")
    .toLowerCase();
}

function buildSubscriptionNameIndex() {
  const index = new Map();
  state.subscriptions.forEach((item) => {
    if (item.id) index.set(`id:${item.id}`, item.name || "");
    if (item.name) index.set(`name:${item.name}`, item.name || "");
  });
  return index;
}

function nodeDisplaySourceName(node, sourceNameIndex) {
  return (
    sourceNameIndex.get(`id:${node.source?.id || ""}`) ||
    sourceNameIndex.get(`name:${node.source?.name || ""}`) ||
    node.source?.name ||
    "-"
  );
}

function applyNodeFiltersAndPagination() {
  void loadNodes();
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
  renderNodeFetchNotices();
  renderNodeBulkActionBar();
  renderNodeTable();
  renderNodePagination();
}

function renderNodeFetchNotices() {
  const box = document.getElementById("node-fetch-notices");
  if (!box) return;
  const notices = Array.isArray(state.nodeFetchNotices)
    ? state.nodeFetchNotices.slice(0, 5)
    : [];
  box.classList.toggle("hidden", !notices.length);
  if (!notices.length) {
    box.innerHTML = "";
    return;
  }
  box.innerHTML = notices
    .map(
      (notice) =>
        `<div class="node-fetch-notice">${escapeHtml(String(notice))}</div>`,
    )
    .join("");
}

function renderNodeStatsStrip() {
  const items = [
    ["节点", state.nodeSummary.total || 0],
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

function renderNodeBulkActionBar() {
  const bar = document.getElementById("node-bulk-action-bar");
  if (!bar) return;
  const selectedCount = state.selectedNodeIds.size;
  const currentPageCount = state.nodes.length;
  const filteredCount = state.nodePagination.total || state.filteredNodes.length;
  const hasNodes = currentPageCount > 0 || filteredCount > 0 || selectedCount > 0;
  bar.classList.toggle("hidden", !hasNodes);
  if (!hasNodes) {
    bar.innerHTML = "";
    return;
  }
  const selectedText = selectedCount
    ? `已选 ${selectedCount} 个节点`
    : `当前筛选 ${filteredCount} 个节点`;
  bar.innerHTML = `
    <div class="bulk-action-summary">
      <strong>${escapeHtml(selectedText)}</strong>
      <span>当前页 ${currentPageCount} 个</span>
    </div>
    <div class="bulk-action-controls">
      <button id="select-current-page-nodes-btn" class="ghost-button small-button" type="button" ${currentPageCount ? "" : "disabled"}>选择当前页</button>
      <button id="clear-selected-nodes-btn" class="ghost-button small-button" type="button" ${selectedCount ? "" : "disabled"}>清空选择</button>
      <button id="bulk-delete-current-filter-btn" class="danger-ghost-button small-button" type="button" ${filteredCount ? "" : "disabled"}>删除当前筛选</button>
      <button id="bulk-delete-selected-nodes-btn" class="danger-solid-button small-button" type="button" ${selectedCount ? "" : "disabled"}>删除已选</button>
    </div>
  `;
  document
    .getElementById("select-current-page-nodes-btn")
    ?.addEventListener("click", selectCurrentPageNodes);
  document
    .getElementById("clear-selected-nodes-btn")
    ?.addEventListener("click", clearSelectedNodes);
  document
    .getElementById("bulk-delete-current-filter-btn")
    ?.addEventListener("click", () => openBulkDeleteDialog("current_filtered"));
  document
    .getElementById("bulk-delete-selected-nodes-btn")
    ?.addEventListener("click", () => openBulkDeleteDialog("selected"));
}

function renderNodeTable() {
  const body = document.getElementById("node-table-body");
  if (!state.nodes.length) {
    const hasActiveFilters =
      String(state.nodeFilters.q || "").trim() ||
      (state.nodeFilters.type && state.nodeFilters.type !== "all") ||
      (state.nodeFilters.region && state.nodeFilters.region !== "ALL") ||
      (state.nodeFilters.status && state.nodeFilters.status !== "all") ||
      state.nodeFilters.source;
    const emptyTitle = hasActiveFilters ? "没有匹配的节点" : "暂无节点";
    const emptyDescription = hasActiveFilters
      ? "调整搜索或筛选条件后再试"
      : "添加订阅或手动节点后会显示在这里";
    body.innerHTML = `
      <tr class="node-empty-row">
        <td colspan="7" class="table-empty">
          <div class="node-empty-state">
            <strong>${emptyTitle}</strong>
            <span>${emptyDescription}</span>
          </div>
        </td>
      </tr>
    `;
    updateNodeSelectPageCheckbox();
    return;
  }

  const sourceNameIndex = buildSubscriptionNameIndex();
  body.innerHTML = state.nodes
    .map((node) => {
      const featureBadges = buildNodeFeatureBadges(node).join("");
      const displaySourceName = nodeDisplaySourceName(node, sourceNameIndex);
      const disabledBadge = !node.enabled
        ? '<span class="mini-badge disabled">已禁用</span>'
        : "";
      return `
        <tr data-node-row-id="${escapeHtml(node.id)}" class="${state.selectedNodeIds.has(node.id) ? "selected-row" : ""}">
          <td class="node-select-col">
            <input class="node-select-checkbox" type="checkbox" aria-label="选择 ${escapeHtml(node.name)}" data-node-select-id="${escapeHtml(node.id)}" ${state.selectedNodeIds.has(node.id) ? "checked" : ""} />
          </td>
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

  bindNodeTableDelegates(body);
  updateNodeSelectPageCheckbox();
  if (state.activeNodeDeletePopoverId) {
    window.requestAnimationFrame(() => {
      document
        .querySelector(
          `[data-node-delete-cancel="${CSS.escape(state.activeNodeDeletePopoverId)}"]`,
        )
        ?.focus({ preventScroll: true });
    });
  }
}

function bindNodeTableDelegates(body) {
  if (body.dataset.delegatesBound === "1") return;
  body.dataset.delegatesBound = "1";
  body.addEventListener("change", (event) => {
    const checkbox = event.target.closest("[data-node-select-id]");
    if (!checkbox) return;
    setNodeSelected(checkbox.dataset.nodeSelectId, checkbox.checked);
  });
  body.addEventListener("click", async (event) => {
    const cancelButton = event.target.closest("[data-node-delete-cancel]");
    if (cancelButton) {
      event.stopPropagation();
      state.activeNodeDeletePopoverId = "";
      renderNodeTable();
      return;
    }

    const confirmButton = event.target.closest("[data-node-delete-confirm]");
    if (confirmButton) {
      event.stopPropagation();
      const nodeId = confirmButton.dataset.nodeDeleteConfirm;
      const node = state.nodeById.get(nodeId) || state.editingNode;
      if (node?.enabled === false && node.source?.kind !== "custom") {
        await performRestoreNode(nodeId);
      } else {
        await performDeleteNode(nodeId);
      }
      return;
    }

    const actionButton = event.target.closest("[data-node-action]");
    if (!actionButton) return;
    const nodeId = actionButton.dataset.nodeId;
    const action = actionButton.dataset.nodeAction;
    if (action === "edit") {
      await openNodeDialog(nodeId);
      return;
    }
    if (action === "delete") {
      event.stopPropagation();
      toggleNodeDeletePopover(nodeId);
    }
  });
}

function updateNodeSelectPageCheckbox() {
  const checkbox = document.getElementById("node-select-page-checkbox");
  if (!checkbox) return;
  const visibleIds = state.nodes.map((node) => node.id);
  const selectedVisibleCount = visibleIds.filter((id) =>
    state.selectedNodeIds.has(id),
  ).length;
  checkbox.checked = visibleIds.length > 0 && selectedVisibleCount === visibleIds.length;
  checkbox.indeterminate =
    selectedVisibleCount > 0 && selectedVisibleCount < visibleIds.length;
  checkbox.disabled = visibleIds.length === 0;
  checkbox.onchange = () => {
    state.nodes.forEach((node) => {
      setNodeSelected(node.id, checkbox.checked, false);
    });
    refreshNodeSelectionUI();
  };
}

function setNodeSelected(nodeId, selected, render = true) {
  if (!nodeId) return;
  if (selected) state.selectedNodeIds.add(nodeId);
  else state.selectedNodeIds.delete(nodeId);
  if (selected) {
    const node = state.nodeById.get(nodeId);
    if (node) state.selectedNodeCache.set(nodeId, node);
  } else {
    state.selectedNodeCache.delete(nodeId);
  }
  if (render) refreshNodeSelectionUI(nodeId);
}

function selectCurrentPageNodes() {
  state.nodes.forEach((node) => {
    state.selectedNodeIds.add(node.id);
    state.selectedNodeCache.set(node.id, node);
  });
  refreshNodeSelectionUI();
}

function clearSelectedNodes() {
  state.selectedNodeIds.clear();
  state.selectedNodeCache.clear();
  refreshNodeSelectionUI();
}

function refreshNodeSelectionUI(nodeId = "") {
  if (nodeId) {
    const escapedID = CSS.escape(nodeId);
    document
      .querySelector(`[data-node-row-id="${escapedID}"]`)
      ?.classList.toggle("selected-row", state.selectedNodeIds.has(nodeId));
    const checkbox = document.querySelector(
      `[data-node-select-id="${escapedID}"]`,
    );
    if (checkbox) checkbox.checked = state.selectedNodeIds.has(nodeId);
  } else {
    document.querySelectorAll("[data-node-row-id]").forEach((row) => {
      const nodeId = row.dataset.nodeRowId;
      row.classList.toggle("selected-row", state.selectedNodeIds.has(nodeId));
    });
    document.querySelectorAll("[data-node-select-id]").forEach((checkbox) => {
      checkbox.checked = state.selectedNodeIds.has(checkbox.dataset.nodeSelectId);
    });
  }
  renderNodeBulkActionBar();
  updateNodeSelectPageCheckbox();
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
  if (state.activeNodeDeletePopoverId) {
    state.activeNodeDeletePopoverId = "";
    renderNodeTable();
  }
  const dialog = document.getElementById("node-dialog");
  const localNode = state.nodeById.get(nodeId);
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
  if (config.action === "bulk-delete-nodes") {
    const ok = await performBulkDeleteNodes(config.nodes || config.nodeIds || []);
    if (ok) closeDangerDialog();
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

async function openDeletedNodesDialog() {
  const response = await fetchJSON("/api/nodes/deleted");
  if (!response?.ok) {
    showToast(readAPIError(response) || "加载已删除节点失败。", true);
    return;
  }

  const nodes = Array.isArray(response.nodes) ? response.nodes : [];
  const missingIDs = Array.isArray(response.missing_ids)
    ? response.missing_ids
    : [];
  if (!nodes.length && !missingIDs.length) {
    showToast("当前没有已删除的订阅节点。");
    return;
  }

  renderDeletedNodesDialog(nodes, missingIDs);
  const dialog = document.getElementById("restore-deleted-dialog");
  if (!dialog.open) dialog.showModal();
}

function renderDeletedNodesDialog(nodes, missingIDs) {
  const body = document.getElementById("restore-deleted-dialog-body");
  const meta = document.getElementById("restore-deleted-dialog-meta");
  const restoreAllButton = document.getElementById("restore-all-deleted-nodes-btn");
  const ids = nodes.map((node) => node.id).filter(Boolean);
  restoreAllButton.dataset.restoreDeletedAll = ids.join(",");
  restoreAllButton.disabled = ids.length === 0;
  meta.textContent = ids.length
    ? `当前可恢复 ${ids.length} 个已删除订阅节点。`
    : "当前订阅中没有可恢复节点。";

  const nodeItems = nodes
    .map(
      (node) => `
        <li class="deleted-node-list-item">
          <span>
            <strong title="${escapeHtml(node.name || shortNodeID(node.id))}">${escapeHtml(node.name || shortNodeID(node.id))}</strong>
            <small>${escapeHtml(node.type || "-")} · ${escapeHtml(node.source?.name || "-")}</small>
          </span>
          <button class="secondary-button tiny-button" type="button" data-restore-deleted-node="${escapeHtml(node.id)}">恢复</button>
        </li>
      `,
    )
    .join("");
  const missingNote = missingIDs.length
    ? `<div class="restore-deleted-note">另有 ${missingIDs.length} 条删除记录当前订阅里找不到对应节点，订阅恢复后可再次尝试。</div>`
    : "";
  body.innerHTML = `
    <div class="restore-deleted-note">恢复会取消本地删除标记，节点会重新显示并参与生成。</div>
    ${nodes.length ? `<ul class="deleted-node-list">${nodeItems}</ul>` : ""}
    ${missingNote}
  `;

  body.querySelectorAll("[data-restore-deleted-node]").forEach((button) => {
    button.addEventListener("click", async () => {
      const ok = await restoreDeletedNodes([button.dataset.restoreDeletedNode]);
      if (ok) closeDialog("restore-deleted-dialog");
    });
  });
}

async function restoreAllDeletedNodesFromDialog() {
  const raw = document.getElementById("restore-all-deleted-nodes-btn")?.dataset
    .restoreDeletedAll;
  const ids = String(raw || "")
    .split(",")
    .map((id) => id.trim())
    .filter(Boolean);
  const ok = await restoreDeletedNodes(ids);
  if (ok) closeDialog("restore-deleted-dialog");
}

async function restoreDeletedNodes(nodeIds) {
  const ids = Array.from(new Set((nodeIds || []).filter(Boolean)));
  if (!ids.length) {
    showToast("没有可恢复的节点。", true);
    return false;
  }

  const response = await fetchJSON("/api/nodes/enable", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ ids }),
  });
  if (!response?.ok) {
    showToast(readAPIError(response) || "恢复已删除节点失败。", true);
    return false;
  }

  ids.forEach((id) => {
    state.selectedNodeIds.delete(id);
    state.selectedNodeCache.delete(id);
  });
  await refreshNodesAfterGenerate();
  showToast(ids.length === 1 ? "节点已恢复。" : `已恢复 ${ids.length} 个节点。`);
  return true;
}

async function performDeleteNode(nodeId) {
  const node =
    state.nodeById.get(nodeId) || state.editingNode;
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

function getBulkDeleteTargets(scope) {
  const source =
    scope === "current_filtered"
      ? state.filteredNodes
      : Array.from(state.selectedNodeIds)
          .map((id) => state.selectedNodeCache.get(id) || state.nodeById.get(id))
          .filter(Boolean);
  const seen = new Set();
  const out = [];
  source.forEach((node) => {
    if (!node?.id || seen.has(node.id)) return;
    seen.add(node.id);
    out.push(node);
  });
  return out;
}

async function getBulkDeleteTargetsForScope(scope) {
  if (scope === "current_filtered") {
    return uniqueNodes(await loadAllFilteredNodes());
  }
  return getBulkDeleteTargets(scope);
}

function uniqueNodes(nodes) {
  const seen = new Set();
  const out = [];
  (nodes || []).forEach((node) => {
    if (!node?.id || seen.has(node.id)) return;
    seen.add(node.id);
    out.push(node);
  });
  return out;
}

async function openBulkDeleteDialog(scope) {
  const targets = await getBulkDeleteTargetsForScope(scope);
  if (!targets.length) {
    showToast(scope === "current_filtered" ? "当前筛选没有可删除的节点。" : "请先选择要删除的节点。", true);
    return;
  }
  const manualCount = targets.filter((node) => node.source?.kind === "custom").length;
  const subscriptionCount = targets.length - manualCount;
  const previewNames = targets
    .slice(0, 6)
    .map((node) => `<li>${escapeHtml(node.name || shortNodeID(node.id))}</li>`)
    .join("");
  const extraCount = targets.length > 6 ? targets.length - 6 : 0;
  openDangerDialog({
    action: "bulk-delete-nodes",
    nodeIds: targets.map((node) => node.id),
    nodes: targets,
    title: "批量删除节点",
    subtitle:
      scope === "current_filtered"
        ? `确定删除当前筛选的 ${targets.length} 个节点吗？`
        : `确定删除已选的 ${targets.length} 个节点吗？`,
    confirmText: "批量删除",
    detailsHtml: `
      <div class="danger-dialog-info">
        <div class="danger-dialog-row"><span>订阅节点</span><strong>${subscriptionCount}</strong></div>
        <div class="danger-dialog-row"><span>手动节点</span><strong>${manualCount}</strong></div>
      </div>
      <div class="danger-dialog-note">订阅节点会标记为本地删除，后续订阅更新不会恢复显示；手动节点会从当前配置中移除。</div>
      <ul class="danger-dialog-list">${previewNames}${extraCount ? `<li>另有 ${extraCount} 个节点...</li>` : ""}</ul>
    `,
  });
}

async function performBulkDeleteNodes(targets) {
  const nodes = normalizeBulkDeleteTargets(targets);
  const ids = nodes.map((node) => node.id);
  if (!ids.length) {
    showToast("没有可删除的节点。", true);
    return false;
  }
  state.activeNodeDeletePopoverId = "";
  const customIDs = [];
  const subscriptionIDs = [];
  nodes.forEach((node) => {
    const id = node.id;
    if (node?.source?.kind === "custom") customIDs.push(id);
    else subscriptionIDs.push(id);
  });

  if (subscriptionIDs.length) {
    const response = await fetchJSON("/api/nodes/delete", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ ids: subscriptionIDs }),
    });
    if (!response?.ok) {
      showToast(readAPIError(response) || "批量删除订阅节点失败。", true);
      return false;
    }
  }

  for (const nodeId of customIDs) {
    const response = await fetchJSON(
      `/api/nodes/custom/${encodeURIComponent(nodeId)}`,
      {
        method: "DELETE",
      },
    );
    if (!response?.ok) {
      showToast(readAPIError(response) || "批量删除手动节点失败。", true);
      return false;
    }
  }

  ids.forEach((id) => {
    state.selectedNodeIds.delete(id);
    state.selectedNodeCache.delete(id);
  });
  await refreshNodesAfterGenerate();
  showToast(`批量删除已完成，共处理 ${ids.length} 个节点。`);
  return true;
}

function normalizeBulkDeleteTargets(targets) {
  const seen = new Set();
  const out = [];
  (targets || []).forEach((target) => {
    const node =
      typeof target === "string"
        ? state.selectedNodeCache.get(target) || state.nodeById.get(target) || { id: target }
        : target;
    if (!node?.id || seen.has(node.id)) return;
    seen.add(node.id);
    out.push(node);
  });
  return out;
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

function loadYamlPreview() {
  if (!state.workspaceId) return Promise.resolve();
  const url = withWorkspace("/api/preview-yaml");
  if (state.yamlRequest?.url === url) return state.yamlRequest.promise;
  state.yamlRequest?.controller.abort();
  const request = { url, controller: new AbortController() };
  state.yamlRequest = request;
  state.yamlError = "";
  renderYamlViewer();
  request.promise = fetchYamlPreview(request).finally(() => {
    if (state.yamlRequest !== request) return;
    state.yamlRequest = null;
    renderYamlViewer();
  });
  return request.promise;
}

async function fetchYamlPreview(request) {
  try {
    const response = await fetch(request.url, {
      credentials: "same-origin",
      cache: "no-store",
      signal: request.controller.signal,
    });
    if (state.yamlRequest !== request) return;
    if (response.status === 401) {
      redirectToLogin();
      return;
    }
    if (!response.ok) {
      throw new Error(`HTTP ${response.status}`);
    }
    const yaml = await response.text();
    if (state.yamlRequest !== request) return;
    state.yamlPreview = yaml;
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
    renderResult();
  } catch (error) {
    if (state.yamlRequest !== request || error.name === "AbortError") return;
    state.yamlError = `YAML 预览加载失败：${error.message}`;
  }
}

function renderYamlViewer() {
  const viewer = document.getElementById("yaml-viewer");
  const gutter = document.getElementById("yaml-line-numbers");
  const code = document.getElementById("yaml-code-content");
  const searchInput = document.getElementById("yaml-search");
  const wrapToggle = document.getElementById("yaml-wrap-toggle");
  const copyButton = document.getElementById("copy-yaml-btn");
  const loading = Boolean(state.yamlRequest);
  const hasYaml = Boolean(String(state.yamlPreview || "").trim());
  const visibleText = limitLines(
    state.yamlPreview || "",
    YAML_PREVIEW_LINE_LIMIT,
  );
  const lines = hasYaml ? String(visibleText).split("\n") : [];
  viewer.classList.toggle("is-empty", !hasYaml);
  viewer.classList.toggle("wrapped", state.yamlWrap);
  viewer.setAttribute("aria-busy", String(loading));
  document.getElementById("refresh-yaml-btn").disabled = loading;

  if (searchInput) {
    searchInput.disabled = !hasYaml;
    searchInput.placeholder = hasYaml ? "搜索 YAML 内容" : "生成后可搜索 YAML";
    if (!hasYaml) {
      state.yamlSearch = "";
      searchInput.value = "";
    }
  }
  if (wrapToggle) wrapToggle.disabled = !hasYaml;
  if (copyButton) copyButton.disabled = !hasYaml;

  if (!hasYaml) {
    gutter.textContent = "";
    code.textContent = state.yamlError || (loading ? "正在加载 YAML..." : "生成配置后显示 YAML");
    renderYamlPreviewMeta(0, 0);
    return;
  }

  const query = state.yamlSearch.trim();
  let matchCount = 0;
  const html = lines
    .map((line, index) => {
      const highlighted = renderYAMLLine(line, query);
      matchCount += highlighted.count;
      return `<span class="yaml-source-line" data-line="${index + 1}">${highlighted.html || "&nbsp;"}</span>`;
    })
    .join("");

  code.innerHTML = html;
  syncYamlGutter();
  renderYamlPreviewMeta(lines.length, countYAMLLines(state.yamlPreview), {
    query,
    matchCount,
  });
}

function syncYamlGutter() {
  const viewer = document.getElementById("yaml-viewer");
  const gutter = document.getElementById("yaml-line-numbers");
  const code = document.getElementById("yaml-code-content");
  if (!viewer || !gutter || !code || viewer.classList.contains("is-empty")) {
    return;
  }

  const sourceLines = Array.from(code.querySelectorAll(".yaml-source-line"));
  if (!sourceLines.length) {
    gutter.textContent = "";
    return;
  }

  const lineHeight = Number.parseFloat(getComputedStyle(code).lineHeight) || 1;
  gutter.innerHTML = sourceLines
    .map((line, index) => {
      const height = Math.max(lineHeight, line.getBoundingClientRect().height);
      return `<span class="yaml-line-number" style="height:${height}px">${index + 1}</span>`;
    })
    .join("");
}

function renderYamlPreviewMeta(previewLines, totalLines, options = {}) {
  const meta = document.getElementById("yaml-search-meta");
  if (!meta) return;
  if (state.yamlError || state.yamlRequest) {
    meta.textContent = state.yamlError || "正在加载 YAML...";
    return;
  }
  const hasYaml = Boolean(String(state.yamlPreview || "").trim());
  if (!hasYaml) {
    meta.textContent = "生成配置后可在此查看、搜索和复制 YAML。";
    return;
  }
  const visibleLines = Math.min(previewLines || 0, YAML_PREVIEW_LINE_LIMIT);
  const lineText =
    totalLines > visibleLines
      ? `预览前 ${visibleLines} 行 / 共 ${totalLines} 行`
      : `共 ${visibleLines} 行`;
  const nodeCount = Number.parseInt(state.resultNodeCount, 10) || 0;
  const ruleCount = Number.parseInt(state.resultRuleCount, 10) || 0;
  const parts = [lineText, `${nodeCount} 个节点`, `${ruleCount} 条规则`];
  if (options.query) {
    parts.push(`匹配 ${options.matchCount || 0} 处`);
  }
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

async function loadBuildVersion() {
  const response = await fetchJSON("/healthz");
  if (!response?.ok || !response.version) return;
  setVersionBadge(response.version);
}

async function refreshVersionStatus() {
  await loadBuildVersion();
  await checkForUpdates();
}

async function checkForUpdates() {
  const badge = document.getElementById("version-badge");
  if (!badge || state.updateCheck.checking) return;

  state.updateCheck.checking = true;
  badge.classList.add("checking");
  badge.title = "正在检查版本更新";

  const response = await fetchJSON("/api/update-check");

  state.updateCheck.checking = false;
  badge.classList.remove("checking");

  if (!response?.ok) {
    badge.title = readAPIError(response) || "检查版本更新失败";
    return;
  }

  const current = formatVersionLabel(response.current_version || "");
  const latest = formatVersionLabel(response.latest_version || "");
  if (current) setVersionBadge(current);

  state.updateCheck.releaseUrl = response.release_url || "";
  state.updateCheck.updateAvailable = Boolean(response.update_available);
  badge.classList.toggle("update-available", state.updateCheck.updateAvailable);

  if (!response.comparable) {
    badge.title = latest
      ? `当前版本无法比较，最新发布 ${latest}`
      : "当前版本无法比较";
    return;
  }

  if (response.update_available) {
    badge.title = `发现新版本 ${latest}，点击打开发布页`;
    return;
  }

  badge.title = `当前已是最新版本 ${current || latest}`;
}

function setVersionBadge(value) {
  const label = document.getElementById("version-label");
  if (label) label.textContent = formatVersionLabel(value);
}

function formatVersionLabel(value) {
  const version = String(value || "").trim();
  if (!version) return "";
  return /^\d/.test(version) ? `v${version}` : version;
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
  badge.innerHTML = `${icon("server", `status-icon ${state.backendOnline ? "success" : "danger"}`)}<span class="dot"></span><span>${state.backendOnline ? "服务在线" : "服务离线"}</span>`;
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
    if (
      lower.includes("no allowed ips") &&
      !diagnostics.items.some((item) => item.title === "订阅地址受限")
    ) {
      diagnostics.items.push({
        title: "订阅地址受限",
        detail: "订阅域名解析到了受限网络地址，请检查 DNS 或代理的 Fake-IP 设置。",
      });
    } else if (
      (lower.includes("fetch_failed") || lower.includes("refresh failed")) &&
      !diagnostics.items.some((item) => item.title === "订阅抓取失败")
    ) {
      diagnostics.items.push({
        title: "订阅抓取失败",
        detail: "未能读取订阅内容，请检查订阅地址或稍后重试。",
      });
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
  markNodeCacheStale();
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

function buildBrowserViewURL(url) {
  if (!url) return "";
  try {
    const parsed = new URL(url, window.location.href);
    parsed.searchParams.set("view", "1");
    return parsed.toString();
  } catch {
    return `${url}${url.includes("?") ? "&" : "?"}view=1`;
  }
}

function openUrl(url) {
  if (!url) return;
  try {
    const parsed = new URL(url, window.location.href);
    if (parsed.protocol !== "http:" && parsed.protocol !== "https:") return;
    window.open(parsed.toString(), "_blank", "noopener,noreferrer");
  } catch {
    return;
  }
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
    const requestOptions = { credentials: "same-origin", ...(options || {}) };
    const headers = new Headers(options?.headers || {});
    const method = String(requestOptions.method || "GET").toUpperCase();
    if (["POST", "PUT", "PATCH", "DELETE"].includes(method) && state.csrfToken) {
      headers.set("X-SubConv-CSRF", state.csrfToken);
    }
    requestOptions.headers = headers;
    const response = await fetch(withWorkspace(url), requestOptions);
    if (response.status === 401 && !String(url).startsWith("/api/auth/")) {
      redirectToLogin();
      return {
        ok: false,
        _httpStatus: response.status,
        error: { code: "UNAUTHORIZED", message: "登录已过期" },
      };
    }
    const payload = await response.json();
    return {
      ...(payload || {}),
      _httpStatus: response.status,
      _retryAfter: Number.parseInt(response.headers.get("Retry-After"), 10) || 0,
    };
  } catch (error) {
    return {
      ok: false,
      _httpStatus: 0,
      _networkError: true,
      error: { code: "NETWORK_ERROR", message: error.message || "request failed" },
    };
  }
}

async function loadAuthSession() {
  try {
    const response = await fetch("/api/auth/session", {
      credentials: "same-origin",
      headers: { Accept: "application/json" },
    });
    const session = await response.json();
    state.authRequired = Boolean(session.required);
    state.publicConverter = Boolean(session.public_converter);
    state.csrfToken = session.csrf_token || "";
    document.getElementById("logout-btn")?.classList.toggle("hidden", !state.authRequired);
    if (state.authRequired && !session.authenticated) {
      redirectToLogin();
      return false;
    }
    return true;
  } catch (_error) {
    return true;
  }
}

async function logoutManagementSession() {
  const button = document.getElementById("logout-btn");
  if (button) button.disabled = true;
  try {
    await fetch("/api/auth/logout", {
      method: "POST",
      credentials: "same-origin",
      headers: state.csrfToken ? { "X-SubConv-CSRF": state.csrfToken } : {},
    });
  } finally {
    window.location.replace("/login");
  }
}

function redirectToLogin() {
  const next = `${window.location.pathname}${window.location.search}${window.location.hash}`;
  window.location.replace(`/login?next=${encodeURIComponent(next)}`);
}

function withWorkspace(rawUrl) {
  const url = String(rawUrl || "");
  if (!state.workspaceId) return url;
  if (!url.startsWith("/api/")) return url;
  if (
    url.startsWith("/api/workspaces") ||
    url.startsWith("/api/auth/") ||
    url.startsWith("/api/site-logo") ||
    url.startsWith("/api/parse") ||
    url.startsWith("/api/update-check")
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
  const openDialog = document.querySelector("dialog[open]");
  if (openDialog && toast.parentElement !== openDialog) {
    openDialog.appendChild(toast);
  } else if (!openDialog && toast.parentElement !== document.body) {
    document.body.appendChild(toast);
  }
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
