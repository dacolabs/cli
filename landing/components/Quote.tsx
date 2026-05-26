export default function Quote() {
  return (
    <section className="quote-section">
      <div className="container">
        <blockquote className="quote">
          <span className="quotemark">&ldquo;</span>
          We were tired of our development being tied to a platform or database technology.
          Having to leave your local development environment to test your pipelines.
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
