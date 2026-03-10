import { CompetitionRoomView } from "@/components/competition-room-view";
import { SiteHeader } from "@/components/site-header";

type PageProps = {
  params: Promise<{ code: string }>;
};

export default async function CompetitionRoomPage({ params }: PageProps) {
  const { code } = await params;

  return (
    <div className="site-root">
      <SiteHeader />
      <main className="container catalog-page">
        <CompetitionRoomView roomCode={code} />
      </main>
    </div>
  );
}

