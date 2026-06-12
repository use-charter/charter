// charter-go-vanity — serves the go-import meta tag for the vanity module path
// go.use-charter.dev/charter, so `go install go.use-charter.dev/charter/...`
// resolves to the GitHub repo. Browsers are redirected to pkg.go.dev.
//
// The go tool fetches https://go.use-charter.dev/charter?go-get=1 over HTTPS and
// reads the <meta name="go-import" content="<module-prefix> <vcs> <repo-url>">
// tag. Spec: https://pkg.go.dev/cmd/go#hdr-Remote_import_paths
//                https://go.dev/ref/mod#vcs-find

const MODULE_ROOT = 'go.use-charter.dev/charter';
const REPO = 'https://github.com/use-charter/charter';
const VCS = 'git';

const META = [
  `<meta name="go-import" content="${MODULE_ROOT} ${VCS} ${REPO}">`,
  `<meta name="go-source" content="${MODULE_ROOT} ${REPO} ${REPO}/tree/main{/dir} ${REPO}/blob/main{/dir}/{file}#L{line}">`,
].join('\n  ');

const humanPage = (target: string): string => `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  ${META}
  <meta http-equiv="refresh" content="0; url=${target}">
</head>
<body>Redirecting to <a href="${target}">${target}</a>…</body>
</html>`;

const html = (body: string): Response =>
  new Response(body, { headers: { 'Content-Type': 'text/html; charset=utf-8' } });

export default {
  async fetch(request: Request): Promise<Response> {
    const url = new URL(request.url);

    // The go tool always appends ?go-get=1; answer it with the meta tags only.
    if (url.searchParams.get('go-get') === '1') {
      return html(`<!DOCTYPE html><html><head>\n  ${META}\n</head><body></body></html>`);
    }

    // Humans: redirect to the package docs on pkg.go.dev, preserving subpath.
    const sub = url.pathname.replace(/\/+$/, '');
    const pkgPath = sub && sub !== '/' ? `go.use-charter.dev${sub}` : MODULE_ROOT;
    return html(humanPage(`https://pkg.go.dev/${pkgPath}`));
  },
} satisfies ExportedHandler;
