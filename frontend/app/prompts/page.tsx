import { SiteHeader } from "@/components/site-header";
import { PromptBuilder } from "@/components/prompt-builder";
import { getProblemServer, getProblemsServer, getPromptCatalogServer } from "@/lib/server-api";

export const metadata = { title: "Prompts · BitForge" };
export default async function PromptsPage({ searchParams }: { searchParams: Promise<Record<string, string | string[] | undefined>> }) {
  const query = await searchParams;
  const slug = typeof query.problem === "string" ? query.problem : "";
  const [catalog, problems, problem] = await Promise.all([
    getPromptCatalogServer().catch(() => null), getProblemsServer({}).catch(() => []),
    slug ? getProblemServer(slug).catch(() => null) : Promise.resolve(null),
  ]);
  return <div className="site-root"><SiteHeader /><main className="container catalog-page prompts-page">
    <section className="catalog-header"><h1>Prompts<span className="accent-period">.</span></h1><p>Set it up. Hand it to your agent.</p></section>
    {catalog ? <PromptBuilder catalog={catalog} problems={problems} initialProblem={problem} initialMode={query.mode === "help" || slug ? "help" : "generate"} /> : <p className="catalog-warning">Backend unavailable. Try again shortly.</p>}
  </main></div>;
}
