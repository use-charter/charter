/* Typed content model for the legal preview pages (privacy / terms /
   license). Each page is a hero plus an ordered list of numbered
   sections; every section is a list of typed blocks the renderer knows
   how to draw. HTML strings are trusted static copy authored here. */

export type Tone = 'blue' | 'green' | 'amber';

export interface ParagraphBlock {
  t: 'p';
  html: string;
  legalese?: boolean;
}

export interface ListBlock {
  t: 'list';
  items: string[];
}

export interface TldrBlock {
  t: 'tldr';
  tone: Tone;
  icon: string; // phosphor icon name without the `ph:` prefix
  html: string;
}

export interface PermsBlock {
  t: 'perms';
  permissions: string[];
  conditions: string[];
  limitations: string[];
}

export interface CodeBlock {
  t: 'code';
  title: string;
  body: string;
}

export type Block = ParagraphBlock | ListBlock | TldrBlock | PermsBlock | CodeBlock;

export interface Section {
  id: string;
  title: string;
  blocks: Block[];
}

export interface Chip {
  icon: string; // phosphor icon name without the `ph:` prefix
  t: string;
}

export interface LegalPage {
  file: string;
  cmd: string;
  title: string;
  updated: string;
  read: string;
  words: string;
  lede: string;
  chips: Chip[];
  sections: Section[];
}
