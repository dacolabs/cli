const CHECK_ICON = (
  <svg viewBox="0 0 10 10" fill="none">
    <path d="M2 5l2 2 4-4" stroke="currentColor" strokeWidth="1.6" strokeLinecap="round" strokeLinejoin="round"/>
  </svg>
)

const points = [
  { strong: 'Ports & connections.', text: 'Typed, with explicit dependencies.' },
  { strong: 'JSON Schema.', text: 'Validate at every boundary.' },
  { strong: 'SLAs & quality rules.', text: 'Enforced contracts, not docs.' },
  { strong: 'Versioned in git.', text: 'Review schemas like code.' },
]

export default function SpecShowcase() {
  return (
    <section className="spec" id="opendpi">
      <div className="container">
        <div className="spec-grid">
          <div>
            <div className="eyebrow"><span className="dot" />The OpenDPI standard</div>
            <h2 className="section-title">One spec. Open source. Yours to extend.</h2>
            <p className="section-lede">
              OpenDPI is a public, MIT-licensed format for describing data products: ports, connections, schemas, SLAs.
              The format powers everything Daco does. Use it with or without us.
            </p>
            <ul className="spec-points">
              {points.map((p) => (
                <li key={p.strong}>
                  <span className="check">{CHECK_ICON}</span>
                  <div><strong>{p.strong}</strong><span>{p.text}</span></div>
                </li>
              ))}
            </ul>
          </div>
          <div className="code-card">
            <div className="code-card-header">
              <span><span className="yellow">●</span> dataproduct.yaml</span>
              <span>OpenDPI 1.0</span>
            </div>
            <pre className="code" dangerouslySetInnerHTML={{ __html: `<span class="k">opendpi</span><span class="pun">:</span> <span class="s">"1.0.0"</span>

<span class="k">info</span><span class="pun">:</span>
  <span class="k">title</span><span class="pun">:</span> <span class="s">"Customer Analytics"</span>
  <span class="k">version</span><span class="pun">:</span> <span class="s">"2.1.0"</span>
  <span class="k">owner</span><span class="pun">:</span> <span class="s">"data-platform@acme.io"</span>

<span class="k">connections</span><span class="pun">:</span>
  <span class="k">analytics_db</span><span class="pun">:</span>
    <span class="k">type</span><span class="pun">:</span> postgresql
    <span class="k">host</span><span class="pun">:</span> analytics.db.acme.io
    <span class="k">database</span><span class="pun">:</span> analytics

<span class="k">ports</span><span class="pun">:</span>
  <span class="a">daily_metrics</span><span class="pun">:</span>
    <span class="k">description</span><span class="pun">:</span> <span class="s">"Daily customer metrics"</span>
    <span class="k">connection</span><span class="pun">:</span> analytics_db
    <span class="k">location</span><span class="pun">:</span> public.customer_daily
    <span class="k">schema</span><span class="pun">:</span>
      <span class="k">type</span><span class="pun">:</span> object
      <span class="k">required</span><span class="pun">:</span> [customer_id, date]
      <span class="k">properties</span><span class="pun">:</span>
        <span class="k">customer_id</span><span class="pun">:</span> { type: string }
        <span class="k">date</span><span class="pun">:</span> { type: string, format: date }
        <span class="k">total_orders</span><span class="pun">:</span> { type: integer }
        <span class="k">revenue</span><span class="pun">:</span> { type: number }` }} />
          </div>
        </div>
      </div>
    </section>
  )
}
