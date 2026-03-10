import { CompetitionsLobby } from "@/components/competitions-lobby";
import { SiteHeader } from "@/components/site-header";

export default function CompetitionsPage() {
  return (
    <div className="site-root">
      <SiteHeader />
      <main className="container catalog-page">
        <section className="catalog-header">
          <div>
            <p className="eyebrow">Competitions</p>
            <h1>Compete With Friends</h1>
            <p>
              Open a room, share the code, and run the same challenge format together.
            </p>
          </div>
        </section>

        <CompetitionsLobby />
      </main>
    </div>
  );
}

