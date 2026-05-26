import Link from 'next/link'
import { getAllPosts } from '@/lib/blog'
import Nav from '@/components/Nav'
import Footer from '@/components/Footer'

export const metadata = {
  title: 'Blog · Daco',
  description: 'Releases, deep dives on OpenDPI, and lessons from teams shipping data products with Daco.',
}

export default function BlogPage() {
  const posts = getAllPosts()

  return (
    <>
      <Nav />
      <main style={{ background: 'var(--paper)', minHeight: '100vh' }}>
        <div className="container" style={{ paddingTop: '80px', paddingBottom: '120px' }}>
          <div style={{ marginBottom: '64px' }}>
            <div className="eyebrow"><span className="dot" />From the blog</div>
            <h1 style={{ fontSize: 'clamp(36px, 5vw, 64px)', fontWeight: 700, letterSpacing: '-0.03em', lineHeight: 1, margin: '0 0 16px' }}>
              Notes from the team.
            </h1>
            <p style={{ fontSize: '19px', color: 'var(--muted)', maxWidth: '520px', margin: 0 }}>
              Releases, deep dives on the OpenDPI standard, and lessons from teams shipping data products with Daco.
            </p>
          </div>

          <div className="blog-grid">
            {posts.map(post => (
              <Link key={post.slug} href={`/blog/${post.slug}`} className="blog-card">
                <div className={`blog-card-cover ${post.cover}`}>
                  <span className="glyph">{post.glyph}</span>
                </div>
                <div className="blog-card-body">
                  <div className="blog-meta">
                    <span className="tag">{post.tag}</span>
                    <span>{new Date(post.date).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })}</span>
                  </div>
                  <h3>{post.title}</h3>
                  <p>{post.excerpt}</p>
                  <span className="blog-card-read">
                    Read post
                    <svg width="11" height="11" viewBox="0 0 11 11" fill="none">
                      <path d="M2 5.5h7m0 0L5.5 2M9 5.5L5.5 9" stroke="currentColor" strokeWidth="1.6" strokeLinecap="round" strokeLinejoin="round"/>
                    </svg>
                  </span>
                </div>
              </Link>
            ))}
          </div>
        </div>
      </main>
      <Footer />
    </>
  )
}
