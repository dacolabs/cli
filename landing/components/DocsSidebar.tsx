'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'
import type { DocMeta } from '@/lib/docs'

const SECTIONS = ['Getting Started', 'OpenDPI', 'CLI Reference', 'Guides']

export default function DocsSidebar({ docs }: { docs: DocMeta[] }) {
  const pathname = usePathname()

  return (
    <nav style={{ width: '220px', flexShrink: 0 }}>
      {SECTIONS.map(section => {
        const items = docs.filter(d => d.section === section)
        if (!items.length) return null
        return (
          <div key={section} style={{ marginBottom: '28px' }}>
            <div style={{
              fontFamily: 'var(--mono)',
              fontSize: '11px',
              fontWeight: 600,
              textTransform: 'uppercase',
              letterSpacing: '0.08em',
              color: 'var(--muted-2)',
              marginBottom: '8px',
              paddingLeft: '12px',
            }}>
              {section}
            </div>
            {items.map(doc => {
              const active = pathname === `/docs/${doc.slug}`
              return (
                <Link
                  key={doc.slug}
                  href={`/docs/${doc.slug}`}
                  style={{
                    display: 'block',
                    padding: '7px 12px',
                    borderRadius: '6px',
                    fontSize: '14px',
                    fontWeight: active ? 600 : 400,
                    color: active ? 'var(--ink)' : 'var(--muted)',
                    background: active ? 'var(--paper-3)' : 'transparent',
                    textDecoration: 'none',
                    transition: 'color 0.15s, background 0.15s',
                  }}
                >
                  {doc.title}
                </Link>
              )
            })}
          </div>
        )
      })}
    </nav>
  )
}
