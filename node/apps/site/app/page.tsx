import { fetchBusinessPing } from '@/lib/gateway';

export default async function HomePage() {
  const ping = await fetchBusinessPing();

  return (
    <main>
      <h1>Ting Boundless</h1>
      <p className="lead">
        Public SSR site behind the Gateway. This page calls <code>/v1/business/ping</code> on the
        server.
      </p>
      <section className="card" aria-label="Business service status">
        <dl>
          <dt>API status</dt>
          <dd className={ping ? 'ok' : 'err'}>{ping ? 'reachable' : 'unavailable'}</dd>
          {ping ? (
            <>
              <dt>Service</dt>
              <dd>{ping.service}</dd>
            </>
          ) : (
            <>
              <dt>Hint</dt>
              <dd>Start Gateway (:8080) and business-service (:3005), then refresh.</dd>
            </>
          )}
        </dl>
      </section>
    </main>
  );
}
