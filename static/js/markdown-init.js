(() => {
    const escapeHtml = (text) => String(text)
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/"/g, '&quot;')
        .replace(/'/g, '&#39;');

    if (typeof window.markdownit === 'function') {
        window.mdParser = window.markdownit({
            html: false,
            breaks: true,
            linkify: true
        });
        return;
    }

    console.warn('markdown-it 未能加载，使用安全纯文本渲染');
    window.mdParser = {
        render(text) {
            return '<pre style="white-space: pre-wrap; font-family: inherit;">' +
                escapeHtml(text) +
                '</pre>';
        }
    };
})();
