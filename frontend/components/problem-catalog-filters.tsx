"use client";

import { problemLabel } from "@/lib/problem-labels";
import { useCallback, useEffect, useState } from "react";
import { usePathname, useRouter, useSearchParams } from "next/navigation";

type Props = {
  q: string;
  difficulty: string;
  category: string;
  tag: string;
  categories: string[];
  tags: string[];
};

type Filters = {
  q: string;
  difficulty: string;
  category: string;
  tag: string;
};

function buildQuery(filters: Filters): string {
  const params = new URLSearchParams();
  if (filters.q.trim()) {
    params.set("q", filters.q.trim());
  }
  if (filters.difficulty) {
    params.set("difficulty", filters.difficulty);
  }
  if (filters.category) {
    params.set("category", filters.category);
  }
  if (filters.tag) {
    params.set("tag", filters.tag);
  }
  return params.toString();
}

export function ProblemCatalogFilters({ q, difficulty, category, tag, categories, tags }: Props) {
  const router = useRouter();
  const pathname = usePathname();
  const searchParams = useSearchParams();

  const [searchText, setSearchText] = useState(q);
  const [selectedDifficulty, setSelectedDifficulty] = useState(difficulty);
  const [selectedCategory, setSelectedCategory] = useState(category);
  const [selectedTag, setSelectedTag] = useState(tag);

  const navigateWith = useCallback(
    (filters: Filters): void => {
      const query = buildQuery(filters);
      if (query === searchParams.toString()) {
        return;
      }
      const destination = query ? `${pathname}?${query}` : pathname;
      router.replace(destination, { scroll: false });
    },
    [pathname, router, searchParams],
  );

  useEffect(() => {
    const timer = setTimeout(() => {
      navigateWith({
        q: searchText,
        difficulty: selectedDifficulty,
        category: selectedCategory,
        tag: selectedTag,
      });
    }, 260);

    return () => clearTimeout(timer);
  }, [searchText, selectedDifficulty, selectedCategory, selectedTag, navigateWith]);

  return (
    <section className="catalog-filters">
      <label>
        Search
        <input
          name="q"
          value={searchText}
          onChange={(event) => setSearchText(event.target.value)}
          placeholder="Find a problem…"
        />
      </label>

      <label>
        Difficulty
        <select
          name="difficulty"
          value={selectedDifficulty}
          onChange={(event) => {
            const nextValue = event.target.value;
            setSelectedDifficulty(nextValue);
            navigateWith({
              q: searchText,
              difficulty: nextValue,
              category: selectedCategory,
              tag: selectedTag,
            });
          }}
        >
          <option value="">All</option>
          <option value="easy">Easy</option>
          <option value="medium">Medium</option>
          <option value="hard">Hard</option>
        </select>
      </label>

      <label>
        Category
        <select
          name="category"
          value={selectedCategory}
          onChange={(event) => {
            const nextValue = event.target.value;
            setSelectedCategory(nextValue);
            navigateWith({
              q: searchText,
              difficulty: selectedDifficulty,
              category: nextValue,
              tag: selectedTag,
            });
          }}
        >
          <option value="">All</option>
          {categories.map((item) => (
            <option key={item} value={item}>
              {problemLabel(item)}
            </option>
          ))}
        </select>
      </label>

      <label>
        Tag
        <select
          name="tag"
          value={selectedTag}
          onChange={(event) => {
            const nextValue = event.target.value;
            setSelectedTag(nextValue);
            navigateWith({
              q: searchText,
              difficulty: selectedDifficulty,
              category: selectedCategory,
              tag: nextValue,
            });
          }}
        >
          <option value="">All</option>
          {tags.map((item) => (
            <option key={item} value={item}>
              {problemLabel(item)}
            </option>
          ))}
        </select>
      </label>
    </section>
  );
}
