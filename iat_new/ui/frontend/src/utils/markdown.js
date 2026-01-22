import MarkdownIt from 'markdown-it'
import hljs from 'highlight.js'
import 'highlight.js/styles/atom-one-dark.css' // Choose your preferred style

const md = new MarkdownIt({
  html: false,
  linkify: true,
  typographer: true,
  breaks: true,
  highlight: function (str, lang) {
    if (lang && hljs.getLanguage(lang)) {
      try {
        const highlighted = hljs.highlight(str, { language: lang, ignoreIllegals: true }).value;
        return `<div class="code-block-wrapper">
                  <div class="code-block-header">
                    <span class="code-lang">${lang}</span>
                  </div>
                  <pre class="hljs"><code>${highlighted}</code></pre>
                </div>`;
      } catch (__) {}
    }

    return '<pre class="hljs"><code>' + md.utils.escapeHtml(str) + '</code></pre>';
  }
})

export function renderMarkdown(content) {
  if (!content) return ''
  return md.render(content)
}
