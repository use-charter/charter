import { existsSync, readFileSync } from 'node:fs';
import { relative, resolve } from 'node:path';
import { resolveRepoRoot } from '../lib/process.ts';

type NavigationNode = string | { pages?: NavigationNode[]; groups?: NavigationNode[]; tabs?: NavigationNode[]; anchors?: NavigationNode[] };

const repoRoot = resolveRepoRoot();
const productDir = resolve(repoRoot, 'docs', 'product');
const docsJsonPath = resolve(productDir, 'docs.json');

const errors: string[] = [];

if (!existsSync(docsJsonPath)) {
  fail(['Missing docs/product/docs.json']);
}

const raw = readFileSync(docsJsonPath, 'utf8');
const parsed = JSON.parse(raw) as { navigation?: NavigationNode };
const pages = collectPages(parsed.navigation);

for (const page of pages) {
  const mdxPath = resolve(productDir, `${page}.mdx`);
  if (!existsSync(mdxPath)) {
    errors.push(`Navigation entry ${JSON.stringify(page)} is missing ${relative(repoRoot, mdxPath)}`);
  }
}

if (errors.length > 0) {
  fail(errors);
}

console.log(`validate-product-docs: PASS — docs.json found and ${pages.length} navigation entries resolve to MDX files.`);

function collectPages(node: NavigationNode | NavigationNode[] | undefined): string[] {
  if (!node) {
    return [];
  }

  if (typeof node === 'string') {
    return [node];
  }

  if (Array.isArray(node)) {
    return node.flatMap((entry) => collectPages(entry));
  }

  return [
    ...collectPages(node.pages),
    ...collectPages(node.groups),
    ...collectPages(node.tabs),
    ...collectPages(node.anchors),
  ];
}

function fail(lines: string[]): never {
  console.error('validate-product-docs: FAILED');
  for (const line of lines) {
    console.error(`- ${line}`);
  }
  process.exit(1);
}
