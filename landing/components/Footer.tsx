import Link from 'next/link'
import Logo from './Logo'

export default function Footer() {
  return (
    <footer className="foot">
      <div className="container">
        <div className="foot-grid">
          <div className="foot-brand">
            <Link href="/" className="logo" style={{ color: 'var(--ink)' }}>
              <Logo size={26} />
              daco
            </Link>
            <p>The standard way to define and manage your data products. Built in Amsterdam.</p>
          </div>
          <div>
            <h4>Product</h4>
            <ul>
              <li><a href="/studio">Daco Studio</a></li>
              <li><a href="/docs">CLI</a></li>
              <li><a href="https://github.com/dacolabs/cli" target="_blank" rel="noopener noreferrer">Roadmap</a></li>
            </ul>
          </div>
          <div>
            <h4>Resources</h4>
            <ul>
              <li><a href="/docs">Docs</a></li>
              <li><a href="/blog">Blog</a></li>
              <li><a href="/blog/daco-cli-v0.2.1">Changelog</a></li>
              <li><a href="https://github.com/dacolabs/cli" target="_blank" rel="noopener noreferrer">GitHub</a></li>
            </ul>
          </div>
          <div>
            <h4>Community</h4>
            <ul>
<li><a href="mailto:daco@dacolabs.com">daco@dacolabs.com</a></li>
              <li>Keizersgracht 555<br/>1017DB Amsterdam</li>
            </ul>
          </div>
        </div>
        <div className="foot-bottom">
          <div>© 2026 Daco. All rights reserved.</div>
          <div className="foot-social">
            <a href="https://github.com/dacolabs/cli" target="_blank" rel="noopener noreferrer" aria-label="GitHub">
              <svg width="14" height="14" viewBox="0 0 16 16" fill="currentColor">
                <path d="M8 0C3.58 0 0 3.58 0 8a8 8 0 0 0 5.47 7.59c.4.07.55-.17.55-.38v-1.33c-2.22.48-2.69-1.07-2.69-1.07-.36-.92-.89-1.17-.89-1.17-.73-.5.05-.49.05-.49.81.06 1.23.83 1.23.83.72 1.23 1.88.88 2.34.67.07-.52.28-.88.51-1.08-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.13 0 0 .67-.21 2.2.82a7.65 7.65 0 0 1 4 0c1.53-1.04 2.2-.82 2.2-.82.44 1.11.16 1.93.08 2.13.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48v2.2c0 .21.15.46.55.38A8 8 0 0 0 16 8c0-4.42-3.58-8-8-8z"/>
              </svg>
            </a>
          </div>
        </div>
      </div>
    </footer>
  )
}
