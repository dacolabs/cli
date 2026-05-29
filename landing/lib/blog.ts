import fs from 'fs'
import path from 'path'
import matter from 'gray-matter'
import { remark } from 'remark'
import remarkGfm from 'remark-gfm'
import remarkHtml from 'remark-html'

const POSTS_DIR = path.join(process.cwd(), 'content/blog')

export interface PostMeta {
  slug: string
  title: string
  date: string
  tag: string
  excerpt: string
  cover: string
  glyph: string
}

export interface Post extends PostMeta {
  content: string
}

export function getAllPosts(): PostMeta[] {
  const files = fs.readdirSync(POSTS_DIR).filter(f => f.endsWith('.md'))
  return files
    .map(filename => {
      const slug = filename.replace(/\.md$/, '')
      const raw = fs.readFileSync(path.join(POSTS_DIR, filename), 'utf8')
      const { data } = matter(raw)
      return {
        slug,
        title: data.title ?? '',
        date: data.date ?? '',
        tag: data.tag ?? '',
        excerpt: data.excerpt ?? '',
        cover: data.cover ?? 'c1',
        glyph: data.glyph ?? '',
      } satisfies PostMeta
    })
    .sort((a, b) => (a.date < b.date ? 1 : -1))
}

export async function getPostBySlug(slug: string): Promise<Post | null> {
  const filePath = path.join(POSTS_DIR, `${slug}.md`)
  if (!fs.existsSync(filePath)) return null
  const raw = fs.readFileSync(filePath, 'utf8')
  const { data, content: markdown } = matter(raw)
  const processed = await remark()
    .use(remarkGfm)
    .use(remarkHtml, { sanitize: false })
    .process(markdown)
  return {
    slug,
    title: data.title ?? '',
    date: data.date ?? '',
    tag: data.tag ?? '',
    excerpt: data.excerpt ?? '',
    cover: data.cover ?? 'c1',
    glyph: data.glyph ?? '',
    content: processed.toString(),
  }
}

export function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  })
}
