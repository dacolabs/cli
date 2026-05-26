import Link from 'next/link'

export default function Footer() {
  return (
    <footer className="foot">
      <div className="container">
        <div className="foot-grid">
          <div className="foot-brand">
            <Link href="/" className="logo" style={{ color: 'var(--ink)' }}>
              <svg width="26" height="26" viewBox="0 0 26 26" fill="none">
                <rect width="26" height="26" rx="5" fill="#f4cf4a"/>
                <rect x="7" y="7" width="5" height="5" rx="1" fill="#0c0c0c"/>
                <rect x="14" y="7" width="5" height="5" rx="1" fill="#0c0c0c"/>
                <rect x="7" y="14" width="5" height="5" rx="1" fill="#0c0c0c"/>
                <rect x="14" y="14" width="5" height="5" rx="1" fill="#0c0c0c" opacity="0.3"/>
              </svg>
              daco
            </Link>
            <p>The standard way to define and manage your data products. Built in Amsterdam.</p>
          </div>
          <div>
            <h4>Product</h4>
            <ul>
              <li><a href="#">OpenDPI</a></li>
              <li><a href="#">CLI</a></li>
              <li><a href="#">Translators</a></li>
              <li><a href="#">Roadmap</a></li>
            </ul>
          </div>
          <div>
            <h4>Resources</h4>
            <ul>
              <li><a href="#">Docs</a></li>
              <li><a href="#">Blog</a></li>
              <li><a href="#">Changelog</a></li>
              <li><a href="#">GitHub</a></li>
            </ul>
          </div>
          <div>
            <h4>Community</h4>
            <ul>
              <li><a href="#">Discord</a></li>
              <li><a href="#">Twitter</a></li>
              <li><a href="mailto:community@dacolabs.com">community@dacolabs.com</a></li>
              <li>Keizersgracht 555<br/>1017DB Amsterdam</li>
            </ul>
          </div>
        </div>
        <div className="foot-bottom">
          <div>© 2026 Daco. All rights reserved.</div>
          <div className="foot-social">
            <a href="https://github.com/dacolabs" target="_blank" rel="noopener noreferrer" aria-label="GitHub">
              <svg width="14" height="14" viewBox="0 0 16 16" fill="currentColor">
                <path d="M8 0C3.58 0 0 3.58 0 8a8 8 0 0 0 5.47 7.59c.4.07.55-.17.55-.38v-1.33c-2.22.48-2.69-1.07-2.69-1.07-.36-.92-.89-1.17-.89-1.17-.73-.5.05-.49.05-.49.81.06 1.23.83 1.23.83.72 1.23 1.88.88 2.34.67.07-.52.28-.88.51-1.08-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.13 0 0 .67-.21 2.2.82a7.65 7.65 0 0 1 4 0c1.53-1.04 2.2-.82 2.2-.82.44 1.11.16 1.93.08 2.13.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48v2.2c0 .21.15.46.55.38A8 8 0 0 0 16 8c0-4.42-3.58-8-8-8z"/>
              </svg>
            </a>
            <a href="https://discord.gg/dacolabs" target="_blank" rel="noopener noreferrer" aria-label="Discord">
              <svg width="14" height="14" viewBox="0 0 16 16" fill="currentColor">
                <path d="M13.55 2.95A12.7 12.7 0 0 0 10.4 2c-.14.25-.3.59-.42.86a11.7 11.7 0 0 0-3.95 0A8.5 8.5 0 0 0 5.6 2a12.6 12.6 0 0 0-3.16.95C.43 5.96-.12 8.9.16 11.79a12.8 12.8 0 0 0 3.87 1.96c.31-.43.59-.88.83-1.36-.46-.17-.9-.39-1.32-.64.11-.08.22-.17.32-.25a9.13 9.13 0 0 0 7.88 0c.1.09.21.17.32.25-.42.25-.86.47-1.32.64.24.48.52.93.83 1.36a12.8 12.8 0 0 0 3.87-1.96c.34-3.36-.55-6.27-2.4-8.84zM5.34 9.96c-.76 0-1.39-.7-1.39-1.56s.62-1.56 1.39-1.56 1.4.7 1.39 1.56c0 .86-.62 1.56-1.39 1.56zm5.32 0c-.76 0-1.39-.7-1.39-1.56s.62-1.56 1.39-1.56 1.4.7 1.39 1.56c0 .86-.62 1.56-1.39 1.56z"/>
              </svg>
            </a>
          </div>
        </div>
      </div>
    </footer>
  )
}
