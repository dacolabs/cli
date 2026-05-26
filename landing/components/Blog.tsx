import Link from 'next/link'

const posts = [
  {
    href: '/blog/daco-cli-v0.2.0',
    cover: 'c1',
    glyph: 'v0.2',
    tag: 'Changelog',
    date: 'Feb 3, 2026',
    title: 'Daco CLI Changelog: v0.2.0',
    excerpt: 'The first public release brings schema translation to 12+ formats, connection management, and project scaffolding.',
  },
  {
    href: '/blog/opendpi',
    cover: 'c2',
    glyph: 'opendpi:\n  ports:\n    metrics:\n      schema:\n        ...',
    tag: 'Standard',
    date: 'Feb 3, 2026',
    title: 'OpenDPI: A Standard for Data Product Interfaces',
    excerpt: 'An open standard for describing and documenting data product interfaces, and why your team needs one.',
  },
  {
    href: '/blog/introducing-daco',
    cover: 'c3',
    glyph: 'daco',
    tag: 'Announcement',
    date: 'Jan 30, 2026',
    title: 'Welcome to Daco',
    excerpt: 'A new standard for data teams to align on definitions and ship data products faster, without rewriting their stack.',
  },
]

const ARROW = (
  <svg width="11" height="11" viewBox="0 0 11 11" fill="none">
    <path d="M2 5.5h7m0 0L5.5 2M9 5.5L5.5 9" stroke="currentColor" strokeWidth="1.6" strokeLinecap="round" strokeLinejoin="round"/>
  </svg>
)

export default function Blog() {
  return (
    <section id="blog">
      <div className="container">
        <div className="blog-head">
          <div className="blog-head-text">
            <div className="eyebrow"><span className="dot" />From the blog</div>
            <h2 className="section-title">Notes from the team.</h2>
            <p className="section-lede">
              Releases, deep dives on the OpenDPI standard, and lessons from teams shipping data products with Daco.
            </p>
          </div>
          <Link href="/blog" className="blog-all">
            All posts {ARROW}
          </Link>
        </div>

        <div className="blog-grid">
          {posts.map((post) => (
            <Link key={post.href} href={post.href} className="blog-card">
              <div className={`blog-card-cover ${post.cover}`}>
                <span className="glyph">{post.glyph}</span>
              </div>
              <div className="blog-card-body">
                <div className="blog-meta">
                  <span className="tag">{post.tag}</span>
                  <span>{post.date}</span>
                </div>
                <h3>{post.title}</h3>
                <p>{post.excerpt}</p>
                <span className="blog-card-read">Read post {ARROW}</span>
              </div>
            </Link>
          ))}
        </div>
      </div>
    </section>
  )
}
