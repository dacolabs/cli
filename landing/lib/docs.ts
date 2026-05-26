import fs from 'fs'
import path from 'path'
import matter from 'gray-matter'
import { remark } from 'remark'
import remarkHtml from 'remark-html'

const DOCS_DIR = path.join(process.cwd(), 'content/docs')

export interface DocMeta {
  slug: string
  title: string
  description: string
  order: number
  section: string
}

export interface Doc extends DocMeta {
  content: string
}

export function getAllDocs(): DocMeta[] {
  const files = fs.readdirSync(DOCS_DIR).filter(f => f.endsWith('.md'))
  return files
    .map(filename => {
      const slug = filename.replace(/\.md$/, '')
      const raw = fs.readFileSync(path.join(DOCS_DIR, filename), 'utf8')
      const { data } = matter(raw)
      return {
        slug,
        title: data.title ?? '',
        description: data.description ?? '',
        order: data.order ?? 99,
        section: data.section ?? 'Other',
      } satisfies DocMeta
    })
    .sort((a, b) => a.order - b.order)
}

export async function getDocBySlug(slug: string): Promise<Doc | null> {
  const filePath = path.join(DOCS_DIR, `${slug}.md`)
  if (!fs.existsSync(filePath)) return null
  const raw = fs.readFileSync(filePath, 'utf8')
  const { data, content: markdown } = matter(raw)
  const processed = await remark().use(remarkHtml, { sanitize: false }).process(markdown)
  return {
    slug,
    title: data.title ?? '',
    description: data.description ?? '',
    order: data.order ?? 99,
    section: data.section ?? 'Other',
    content: processed.toString(),
  }
}
