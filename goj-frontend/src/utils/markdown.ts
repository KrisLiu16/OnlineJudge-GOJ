import { marked } from 'marked'
import markedKatex from 'marked-katex-extension'

// 配置 marked
marked.use(
  markedKatex({
    throwOnError: false,
    output: 'html',
    displayMode: false,
    katex: {
      throwOnError: false,
      output: 'html',
    },
  }),
)

// 导出配置好的 marked
export const renderMarkdown = async (content: string): Promise<string> => {
  try {
    return await marked(content || '')
  } catch (error) {
    console.error('Markdown渲染错误:', error)
    return '渲染错误'
  }
}

// 添加全局样式
export const markdownStyles = `
.markdown-body {
  color: var(--text-light);
  line-height: 1.8;
  font-size: 1rem;
}

.markdown-body h1,
.markdown-body h2,
.markdown-body h3,
.markdown-body h4,
.markdown-body h5,
.markdown-body h6 {
  margin-top: 1.5em;
  margin-bottom: 1em;
  font-weight: 600;
  line-height: 1.25;
}

.markdown-body h1 {
  font-size: 2em;
  padding-bottom: 0.3em;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.markdown-body h2 {
  font-size: 1.5em;
  padding-bottom: 0.3em;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.markdown-body h3 {
  font-size: 1.25em;
}

.markdown-body p {
  margin: 1em 0;
}

.markdown-body blockquote {
  margin: 1em 0;
  padding: 0.5em 1em;
  color: var(--text-gray);
  border-left: 0.25em solid var(--primary-color);
  background: rgba(255, 255, 255, 0.05);
  border-radius: 0 4px 4px 0;
}

.markdown-body pre {
  margin: 1em 0;
  padding: 1rem 1.5rem;
  background: rgba(0, 0, 0, 0.3);
  border-radius: 8px;
  overflow-x: auto;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', 'Consolas', monospace;
  font-size: 0.9em;
  line-height: 1.5;
}

.markdown-body code {
  padding: 0.2em 0.4em;
  margin: 0 0.2em;
  background: rgba(0, 0, 0, 0.3);
  border-radius: 4px;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', 'Consolas', monospace;
  font-size: 0.9em;
}

.markdown-body pre code {
  padding: 0;
  margin: 0;
  background: transparent;
  font-size: 1em;
}

.markdown-body ul,
.markdown-body ol {
  margin: 1em 0;
  padding-left: 2em;
}

.markdown-body li {
  margin: 0.5em 0;
}

.markdown-body table {
  margin: 1em 0;
  border-collapse: collapse;
  width: 100%;
  overflow: auto;
}

.markdown-body table th,
.markdown-body table td {
  padding: 0.75rem 1rem;
  border: 1px solid rgba(255, 255, 255, 0.1);
}

.markdown-body table th {
  font-weight: 600;
  background: rgba(255, 255, 255, 0.05);
}

.markdown-body table tr:nth-child(2n) {
  background: rgba(255, 255, 255, 0.02);
}

.markdown-body hr {
  height: 1px;
  margin: 2em 0;
  background: rgba(255, 255, 255, 0.1);
  border: none;
}

.markdown-body img {
  max-width: 100%;
  border-radius: 8px;
  margin: 1em 0;
}

.markdown-body a {
  color: var(--primary-color);
  text-decoration: none;
  transition: all 0.2s;
}

.markdown-body a:hover {
  text-decoration: underline;
}

.katex {
  font-size: 1.1em;
}

.katex-display {
  margin: 1em 0;
  padding: 1em;
  overflow-x: auto;
  background: rgba(0, 0, 0, 0.2);
  border-radius: 8px;
}

/* 浅色主题适配 */
:global(.light-theme) .markdown-body {
  color: var(--text-dark);
}

:global(.light-theme) .markdown-body blockquote {
  background: rgba(0, 0, 0, 0.03);
}

:global(.light-theme) .markdown-body pre,
:global(.light-theme) .markdown-body code {
  background: rgba(0, 0, 0, 0.05);
}

:global(.light-theme) .markdown-body table th {
  background: rgba(0, 0, 0, 0.03);
}

:global(.light-theme) .markdown-body table tr:nth-child(2n) {
  background: rgba(0, 0, 0, 0.01);
}
`
