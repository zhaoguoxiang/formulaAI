import { Pipe, PipeTransform } from '@angular/core';

@Pipe({
  name: 'markdown',
  standalone: true,
})
export class MarkdownPipe implements PipeTransform {
  transform(value: string): string {
    if (!value) return '';
    return value
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/^### (.+)$/gm, '<h3>$1</h3>')
      .replace(/^## (.+)$/gm, '<h2>$1</h2>')
      .replace(/^# (.+)$/gm, '<h1>$1</h1>')
      .replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>')
      .replace(/\*(.+?)\*/g, '<em>$1</em>')
      .replace(/`([^`]+)`/g, '<code class="inline-code">$1</code>')
      .replace(/```(\w*)\n([\s\S]*?)```/g, '<pre><code class="language-$1">$2</code></pre>')
      .replace(/^\- (.+)$/gm, '<li>$1</li>')
      .replace(/(<li>.*<\/li>\n?)+/g, '<ul>$&</ul>')
      .replace(/^(\d+)\. (.+)$/gm, '<li>$2</li>')
      .replace(/^(<li>.*<\/li>\n?)+/gm, (match) => {
        if (match.includes('<ul>')) return match;
        return `<ol>${match}</ol>`;
      })
      .replace(/\n\n/g, '</p><p>')
      .replace(/\n/g, '<br>')
      .replace(/^(.+)$/gm, (match) => {
        if (/^<(\/?(?:h[1-3]|ul|ol|li|pre|p|strong|em|code|br))/.test(match)) return match;
        if (!match.startsWith('<p>')) return `<p>${match}</p>`;
        return match;
      })
      .replace(/<p><\/p>/g, '')
      .replace(/<\/p><p>/g, '</p>\n<p>')
      .replace(/<p>((?:<ul>|<ol>).*?(?:<\/ul>|<\/ol>))<\/p>/g, '$1')
      .replace(/<p>(<(?:h[1-3]|pre).*?(?:\/(?:h[1-3]|pre)>))<\/p>/g, '$1');
  }
}
