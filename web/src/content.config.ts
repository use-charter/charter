import { glob } from 'astro/loaders';
import { defineCollection } from 'astro:content';
import { z } from 'astro:schema';

// Blog content collection (Astro 6 content layer). Posts are Markdown in
// src/content/blog/*.md; files prefixed with `_` are ignored (drafts/wip).
const blog = defineCollection({
  loader: glob({ pattern: '**/[^_]*.md', base: './src/content/blog' }),
  schema: z.object({
    title: z.string(),
    description: z.string(),
    date: z.coerce.date(),
    updated: z.coerce.date().optional(),
    author: z.string().default('Charter'),
    tags: z.array(z.string()).default([]),
    cover: z.string().optional(),
    draft: z.boolean().default(false),
  }),
});

export const collections = { blog };
