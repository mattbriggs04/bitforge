import { ProblemCatalogList } from "@/components/problem-catalog-list";
import { ProblemCatalogFilters } from "@/components/problem-catalog-filters";
import { SiteHeader } from "@/components/site-header";
import { getProblemsServer } from "@/lib/server-api";
import { ProblemSummary } from "@/lib/types";

type PageProps = {
  searchParams: Promise<Record<string, string | string[] | undefined>>;
};

function readParam(
  params: Record<string, string | string[] | undefined>,
  key: string,
): string {
  const value = params[key];
  if (Array.isArray(value)) {
    return value[0] ?? "";
  }
  return value ?? "";
}

function unique(items: string[]): string[] {
  return [...new Set(items)].sort((a, b) => a.localeCompare(b));
}

export default async function ProblemsPage({ searchParams }: PageProps) {
  const params = await searchParams;
  const q = readParam(params, "q");
  const difficulty = readParam(params, "difficulty");
  const category = readParam(params, "category");
  const tag = readParam(params, "tag");

  let problems: ProblemSummary[] = [];
  let allProblems: ProblemSummary[] = [];
  let backendUnavailable = false;

  try {
    [problems, allProblems] = await Promise.all([
      getProblemsServer({ q, difficulty, category, tag }),
      getProblemsServer({}),
    ]);
  } catch {
    backendUnavailable = true;
  }
  const categories = unique(allProblems.map((item) => item.category));
  const tags = unique(allProblems.flatMap((item) => item.tags));

  return (
    <div className="site-root">
      <SiteHeader />
      <main className="container catalog-page">
        <section className="catalog-header">
          <div>
            <h1>Problems<span className="accent-period">.</span></h1>
          </div>

        </section>

        <ProblemCatalogFilters
          key={`filters:${q}|${difficulty}|${category}|${tag}`}
          q={q}
          difficulty={difficulty}
          category={category}
          tag={tag}
          categories={categories}
          tags={tags}
        />

        {backendUnavailable ? (
          <p className="catalog-warning">
            Backend unavailable. Try again shortly.
          </p>
        ) : null}

        <ProblemCatalogList
          initialProblems={problems}
          q={q}
          difficulty={difficulty}
          category={category}
          tag={tag}
        />
      </main>
    </div>
  );
}
