import Link from 'next/link'

const posts = [
  {
    href: '/blog/code-first-data-catalog',
    cover: 'c1',
    glyph: 'cat',
    tag: 'Article',
    date: 'May 20, 2026',
    title: 'Why Your Data Catalog Should Live Next to Your Code',
    excerpt: "Daco Studio is a code-first data catalog that connects directly to your repository. Here's why that matters.",
  },
  {
    href: '/blog/getting-started-with-daco-cli',
    cover: 'c2',
    glyph: '$ daco init\n$ daco validate\n$ daco translate\n  --format pyspark',
    tag: 'Guide',
    date: 'May 12, 2026',
    title: 'Getting Started with the Daco CLI',
    excerpt: 'Step-by-step: install the CLI, initialize a data product, and translate your schema to PySpark, dbt, and more.',
  },
  {
    href: '/blog/standardize-your-data-definitions',
    cover: 'c3',
    glyph: 'opendpi',
    tag: 'Standard',
    date: 'Apr 28, 2026',
    title: 'Standardize Your Data Definitions with OpenDPI',
    excerpt: 'How standardizing data descriptions enables migrations, automation, and team collaboration at scale.',
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
