import { notFound } from 'next/navigation'
import Nav from '@/components/Nav'
import Footer from '@/components/Footer'
import DocsSidebar from '@/components/DocsSidebar'
import { getAllDocs, getDocBySlug } from '@/lib/docs'

export async function generateStaticParams() {
  return getAllDocs().map(d => ({ slug: d.slug }))
}

export async function generateMetadata({ params }: { params: Promise<{ slug: string }> }) {
  const { slug } = await params
  const doc = await getDocBySlug(slug)
  if (!doc) return {}
  return { title: `${doc.title} · Daco Docs`, description: doc.description }
}

export default async function DocPage({ params }: { params: Promise<{ slug: string }> }) {
  const { slug } = await params
  const [doc, allDocs] = await Promise.all([getDocBySlug(slug), Promise.resolve(getAllDocs())])
  if (!doc) notFound()

  const currentIndex = allDocs.findIndex(d => d.slug === slug)
  const prev = currentIndex > 0 ? allDocs[currentIndex - 1] : null
  const next = currentIndex < allDocs.length - 1 ? allDocs[currentIndex + 1] : null

  return (
    <>
      <Nav />
      <div style={{ background: 'var(--paper)', minHeight: '100vh' }}>
        <div className="container">
          <div style={{
            display: 'flex',
            gap: '64px',
            padding: '64px 0 120px',
            alignItems: 'flex-start',
          }}>
            <div className="docs-sidebar" style={{ position: 'sticky', top: '88px' }}>
              <DocsSidebar docs={allDocs} />
            </div>

            <main style={{ flex: 1, minWidth: 0 }}>
              <div style={{ marginBottom: '8px' }}>
                <span style={{
                  fontFamily: 'var(--mono)',
                  fontSize: '11px',
                  fontWeight: 600,
                  textTransform: 'uppercase',
                  letterSpacing: '0.08em',
                  color: 'var(--muted-2)',
                }}>
                  {doc.section}
                </span>
              </div>

              <h1 style={{
                fontSize: 'clamp(28px, 3.5vw, 40px)',
                fontWeight: 700,
                letterSpacing: '-0.03em',
                lineHeight: 1.15,
                margin: '0 0 12px',
                color: 'var(--ink)',
              }}>
                {doc.title}
              </h1>

              <p style={{
                fontSize: '17px',
                color: 'var(--muted)',
                margin: '0 0 48px',
                lineHeight: 1.5,
              }}>
                {doc.description}
              </p>

              <div
                className="prose"
                dangerouslySetInnerHTML={{ __html: doc.content }}
              />

              {(prev || next) && (
                <div style={{
                  display: 'flex',
                  justifyContent: 'space-between',
                  marginTop: '64px',
                  paddingTop: '32px',
                  borderTop: '1px solid var(--line)',
                  gap: '16px',
                }}>
                  {prev ? (
                    <a href={`/docs/${prev.slug}`} style={{
                      display: 'flex',
                      flexDirection: 'column',
                      gap: '4px',
                      textDecoration: 'none',
                      padding: '16px 20px',
                      border: '1px solid var(--line)',
                      borderRadius: 'var(--radius)',
                      minWidth: 0,
                      flex: 1,
                    }}>
                      <span style={{ fontSize: '12px', color: 'var(--muted-2)', fontFamily: 'var(--mono)' }}>← Previous</span>
                      <span style={{ fontSize: '14px', fontWeight: 600, color: 'var(--ink)' }}>{prev.title}</span>
                    </a>
                  ) : <div />}
                  {next && (
                    <a href={`/docs/${next.slug}`} style={{
                      display: 'flex',
                      flexDirection: 'column',
                      alignItems: 'flex-end',
                      gap: '4px',
                      textDecoration: 'none',
                      padding: '16px 20px',
                      border: '1px solid var(--line)',
                      borderRadius: 'var(--radius)',
                      minWidth: 0,
                      flex: 1,
                    }}>
                      <span style={{ fontSize: '12px', color: 'var(--muted-2)', fontFamily: 'var(--mono)' }}>Next →</span>
                      <span style={{ fontSize: '14px', fontWeight: 600, color: 'var(--ink)' }}>{next.title}</span>
                    </a>
                  )}
                </div>
              )}
            </main>
          </div>
        </div>
      </div>
      <Footer />
    </>
  )
}
