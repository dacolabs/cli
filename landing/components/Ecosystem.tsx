export default function Ecosystem() {
  return (
    <section id="ecosystem" className="ecosystem">
      <div className="container">
        <div className="eyebrow"><span className="dot" />The Daco ecosystem</div>
        <h2 className="section-title">Three layers. One standard.</h2>
        <p className="section-lede">A toolchain built around OpenDPI, from the engineer&apos;s terminal to the business control room.</p>

        <div className="eco-grid">
          {/* CLI */}
          <article className="eco-card">
            <div className="eco-card-glyph">
              <span className="eco-mono">$ daco</span>
            </div>
            <div className="eco-card-body">
              <div className="eco-tag">Open source</div>
              <h3>Daco CLI</h3>
              <p>The engineer&apos;s tool. Import schemas and metadata from any database into OpenDPI. Generate mock data, run tests, and translate to PySpark, dbt, or SQL. All on your machine.</p>
              <ul className="eco-feats">
                <li>Platform-agnostic schema import</li>
                <li>Mock data + local test runners</li>
                <li>Translate to PySpark, dbt, SQL, Pydantic</li>
              </ul>
            </div>
          </article>

          {/* Service & SDKs */}
          <article className="eco-card">
            <div className="eco-card-glyph dark">
              <svg viewBox="0 0 80 80" width="72" height="72" fill="none" stroke="currentColor" strokeWidth="1.6">
                <rect x="6" y="10" width="68" height="14" rx="3"/>
                <rect x="6" y="33" width="68" height="14" rx="3"/>
                <rect x="6" y="56" width="68" height="14" rx="3"/>
                <circle cx="14" cy="17" r="2" fill="currentColor"/>
                <circle cx="14" cy="40" r="2" fill="currentColor"/>
                <circle cx="14" cy="63" r="2" fill="currentColor"/>
              </svg>
            </div>
            <div className="eco-card-body">
              <div className="eco-tag">Open source</div>
              <h3>Daco Service &amp; SDKs</h3>
              <p>A version-controlled registry for OpenDPI specs. Push and pull data product definitions through native SDKs in Python, TypeScript, or Go.</p>
              <ul className="eco-feats">
                <li>Git-style versioning for specs</li>
                <li>Multi-language SDKs</li>
                <li>Self-host or run managed</li>
              </ul>
            </div>
          </article>

          {/* Studio */}
          <article className="eco-card highlighted">
            <div className="eco-card-glyph yellow">
              <svg viewBox="0 0 80 80" width="72" height="72" fill="none" stroke="currentColor" strokeWidth="1.8">
                <rect x="6" y="6" width="32" height="32" rx="3"/>
                <rect x="42" y="6" width="32" height="32" rx="3"/>
                <rect x="6" y="42" width="32" height="32" rx="3"/>
                <rect x="42" y="42" width="32" height="32" rx="3" fill="currentColor"/>
              </svg>
            </div>
            <div className="eco-card-body">
              <div className="eco-tag premium">Premium</div>
              <h3>Daco Studio</h3>
              <p>The control room for the business. Search the data product marketplace, verify compliance, monitor quality rules, and find the owner of every dataset.</p>
              <ul className="eco-feats">
                <li>Data product marketplace</li>
                <li>Compliance &amp; quality monitoring</li>
                <li>Ownership &amp; lineage at a glance</li>
              </ul>
            </div>
          </article>
        </div>
      </div>
    </section>
  )
}
