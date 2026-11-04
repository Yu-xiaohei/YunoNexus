# Agent 文件索引

## README 重制 / 2026-11-05

正式交付为 `README.md`、`assets/readme/hero.svg` 和 `assets/readme/source-map.svg`。本次不修改业务代码、部署配置或历史文档。用户复核后已明确授权提交并推送；临时文件不进入提交。

| 相对路径 | 类型 | 用途与读取方法 | 保留条件 |
| --- | --- | --- | --- |
| `.agent-tmp/readme/tmp_readme_preview_20261105_120000.py` | 临时预览生成脚本 | Python + markdown-it-py 渲染 README，检查本地文件及锚点链接，并导出预览和 diff | 交付复核期间保留；经明确授权可清理 |
| `.agent-tmp/readme/tmp_readme_preview_20261105_120000.html` | 临时静态预览 | 浏览器直接打开；近似 GitHub Markdown 布局，支持亮色和暗色系统主题 | 同上 |
| `.agent-tmp/readme/tmp_readme_diff_20261105_120000.patch` | 临时差异文件 | 相对克隆时 HEAD 的 README 差异；新增 SVG 单独交付 | 同上 |
| `.agent-tmp/readme/tmp_readme_checks_20261105_120000.json` | 临时结构检查结果 | UTF-8 JSON，本地链接、锚点、图像大小及预览生成结果 | 同上 |
| `.agent-tmp/readme/tmp_readme_browser_20261105_120000.json` | 临时浏览器配置 | Playwright CLI 的截图输出位置与浏览器启动参数 | 同上 |
| `.agent-tmp/readme/tmp_readme_desktop_20261105_120000.png` | 临时桌面校样 | 900px 内容宽度的完整页面截图 | 同上 |
| `.agent-tmp/readme/tmp_readme_mobile_20261105_120000.png` | 临时手机校样 | 360px 页面宽度截图 | 同上 |
| `.agent-tmp/readme/tmp_readme_dark_20261105_120000.png` | 临时暗色校样 | 暗色背景下的完整页面截图 | 同上 |

`.agent-tmp/readme/output/playwright/` 为 Playwright 自动管理的会话快照、日志位置。其固定命名由工具生成，集中在临时目录；不进入提交。
