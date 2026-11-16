import Link from "next/link";
import { notFound } from "next/navigation";
import { SiteHeader } from "@/components/site-header";
import { InterviewOverview } from "@/components/interview-overview";
import { getInterviewServer } from "@/lib/server-api";

export default async function InterviewPage({ params }: { params: Promise<{ slug: string }> }) {
  const { slug } = await params;
  const interview = await getInterviewServer(slug);
  if (!interview) notFound();
  return (
    <div className="site-root">
      <SiteHeader />
      <main className="container interview-page">
        <Link className="text-link interview-back" href="/interviews">← Interviews</Link>
        <header className="interview-heading"><h1>{interview.title}<span className="accent-period">.</span></h1><p>{interview.description}</p></header>
        <InterviewOverview initial={interview} />
      </main>
    </div>
  );
}
