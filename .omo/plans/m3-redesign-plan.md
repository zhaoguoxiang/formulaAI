# M3 严格设计重构计划

## 项目现状分析

| 维度 | 现状 | 问题 |
|------|------|------|
| **技术栈** | Angular 21 + Angular Material 21 (已启用 M3) | 仅基础 violet palette，未利用 M3 完整 token 体系 |
| **主题** | `mat.theme()` violet palette + Roboto | 无自定义色板、无 token 级定制、无 dark mode 明确定义 |
| **CSS 架构** | 大量内联 `style="..."`，组件 CSS 仅有 `:host` | 违反 M3 design token 原则，样式不可维护 |
| **排版** | index.html 加载 DM Sans / Roboto / JetBrains Mono，但未定义 typescale | 无 M3 五级排版体系 |
| **间距** | 任意 px 值（16px gap, 280px min-width） | 需迁移到 M3 8dp 栅格 |
| **图表** | chart.js 硬编码 teal 色系 | 需使用 M3 color tokens |
| **导航** | 顶部 toolbar + tabs（单一模式） | 缺少 M3 navigation rail / responsive 导航 |

## 重构范围

### Phase 1: Design Token Foundation（基础）

**1.1 色彩体系**
- 使用 Material Theme Builder 生成完整 seed-based 配色方案
- 在 `styles.scss` 中通过 `mat.theme()` 注入自定义 palettes：
  - Primary, Secondary, Tertiary, Neutral, Neutral Variant, Error
- 定义 `--mat-sys-color-*` 全部 surface hierarchy tokens
- 明确定义 light / dark 两套方案（`color-scheme: light dark` 已就绪）

**1.2 排版体系**
- 定义完整 15 级 M3 typescale（Display L/M/S, Headline L/M/S, Title L/M/S, Body L/M/S, Label L/M/S）
- 使用 DM Sans 作为 brand typeface，系统字体为 plain typeface
- 确保中文排版的行高适配

**1.3 Shape 体系**
- 定义 7 级 corner radius：none(0) → extra-small(4) → small(8) → medium(12) → large(16) → extra-large(28) → full(9999)
- 为各组件类型分配正确的 shape token

**1.4 Elevation & State**
- 定义 6 级 elevation（0-5）
- 确保 state layer 使用 `--mat-sys-state-*` tokens（hover 8% / focus 10% / pressed 10% / dragged 16%）

**1.5 Motion 体系**
- 定义 duration tokens（short 50-200ms → medium 250-400ms → long 450-1000ms）
- 定义 easing tokens（emphasized / standard / legacy）

### Phase 2: CSS Architecture Refactor（架构重构）

**2.1 移除全部内联样式**
- `app.html` 中的 `style="display:flex; ..."` → CSS class
- 所有 `px` 值 → M3 spacing tokens（8dp 倍数）

**2.2 建立 CSS 自定义属性体系**
- 在 `styles.scss` 中定义全局 CSS 变量：
  ```css
  :root {
    --app-spacing-xs: 4px;   /* 0.5 × 8dp */
    --app-spacing-sm: 8px;   /* 1 × 8dp */
    --app-spacing-md: 16px;  /* 2 × 8dp */
    --app-spacing-lg: 24px;  /* 3 × 8dp */
    --app-spacing-xl: 32px;  /* 4 × 8dp */
    --app-spacing-2xl: 48px; /* 6 × 8dp */
  }
  ```

**2.3 重构布局**
- 主布局使用 CSS Grid / Flexbox + M3 spacing tokens
- 使用 `--mat-sys-shape-corner-*` tokens 统一圆角
- 使用 `--mat-sys-elevation-*` 统一阴影层级

### Phase 3: Component Redesign（组件重构）

**3.1 Top App Bar（顶部导航栏）**
- 确保使用 `mat-toolbar` 的 M3 变体
- 添加 navigation icon（可选 menu 图标）
- 考虑使用 `mat-toolbar` color="primary" 变体

**3.2 Navigation（导航体系）**
- 引入 M3 Navigation Rail 作为主要导航（替换/补充 tabs）
- Tabs 保留作为二级导航（在 main content 内）
- 响应式考虑：移动端 → bottom nav bar

**3.3 Dashboard Layout（分析看板）**
- Formula Analysis 中的图表面板遵循 M3 card 规范
- 统一使用 `--mat-sys-shape-corner-medium` (12px)
- 图表颜色从 M3 color tokens 生成

**3.4 Formula Matrix（配方矩阵）**
- 表格使用 M3 data table 规范
- Filter toolbar 使用 M3 segmented buttons 或 chips
- 展开行使用 M3 motion easing tokens
- 状态 chips 使用 M3 color tokens（draft/success/neutral）

**3.5 Formula Editor（配方编辑器）**
- 表单字段 `appearance="outline"` (已是 M3 默认)
- Expansion panels 样式对齐 M3 specs
- 按钮 hierarchy: primary action (filled) / secondary (outlined) / destructive (text)
- Card sections 使用 M3 spacing

**3.6 Chat Panel（AI 助手面板）**
- 聊天气泡对齐 M3 message/chip 组件规范
- 用户消息 vs AI 消息的 surface color 区分
- 输入框使用 M3 search bar / text field 样式
- 连接状态指示器使用 M3 badge

**3.7 Charts（图表）**
- Chart.js 颜色映射到 M3 primary/secondary/tertiary tokens
- 确保 dark mode 下图表的可读性
- 图表 tooltip 样式对齐 M3

### Phase 4: Polish & QA

**4.1 Dark Mode**
- 验证所有组件在 dark mode 下的表现
- 确保 chart colors 在 dark mode 下可读

**4.2 Accessibility**
- Color contrast ratio ≥ 4.5:1
- Focus indicators 符合 M3 规范
- Touch targets ≥ 48dp

**4.3 Motion & Micro-interactions**
- 所有 transition 使用 M3 easing tokens
- Tab 切换 / 展开折叠 / button hover 应用 motion tokens

**4.4 Visual QA**
- 逐组件对比 M3 规范截图
- 验证间距、圆角、阴影一致性

## 涉及文件清单

| 文件 | 操作 | 优先级 |
|------|------|--------|
| `frontend/src/styles.scss` | 重写：完整 M3 token 体系 | P0 |
| `frontend/src/index.html` | 审查：typography 字体加载 | P1 |
| `frontend/src/app/app.html` | 重写：移除内联样式，M3 布局 | P0 |
| `frontend/src/app/app.css` | 创建：布局样式 | P0 |
| `frontend/src/app/components/formula-matrix/formula-matrix.component.css` | 重写：M3 table/filter/chip 样式 | P1 |
| `frontend/src/app/components/formula-analysis/formula-analysis.component.css` | 重写：M3 card/chart 样式 | P1 |
| `frontend/src/app/components/formula-analysis/formula-analysis.component.ts` | 修改：图表颜色使用 M3 tokens | P1 |
| `frontend/src/app/components/formula-editor/formula-editor.component.css` | 重写：M3 form/expansion 样式 | P1 |
| `frontend/src/app/components/chat-panel/chat-panel.component.css` | 重写：M3 chat/message 样式 | P1 |
| `frontend/src/app/components/chat-panel/chat-panel.component.html` | 修改：对齐 M3 消息组件模式 | P2 |

## 实施策略

- **Phase 1 → Phase 2** 必须先完成（token 体系是一切的基础）
- **Phase 3** 各组件可并行推进
- **Phase 4** 在所有视觉变更后进行

## 预期成果

- 全部 UI 严格遵循 Material Design 3 规范
- 设计 token 驱动的 CSS 架构（单一真实来源）
- 完整的 light/dark mode 支持
- 组件间视觉一致性（间距、圆角、阴影、动效）
- CSS 可维护性大幅提升（零内联样式）
