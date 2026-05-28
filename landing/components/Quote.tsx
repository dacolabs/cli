export default function Quote() {
  return (
    <section className="quote-section">
      <div className="container">
        <blockquote className="quote">
          <span className="quotemark">&ldquo;</span>
          We were tired of having to leave our local development environment to test our pipelines.
          <span className="quotemark">&rdquo;</span>
        </blockquote>
        <div className="quote-attr">
          <div className="quote-avatar" />
          <div>
            <div className="quote-attr-name">Orri Pálsson &amp; Giuseppe Grieco</div>
            <div className="quote-attr-title">Founders, Daco</div>
          </div>
        </div>
      </div>
    </section>
  )
}
